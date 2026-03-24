package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	"movieexample.com/metadata/internal/repository/memory"
	"movieexample.com/metadata/internal/repository/postgres"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/tracing"
)

const serviceName = "metadata"

func main() {

	logger, _ := zap.NewProduction()
	logger.Info("Started the service", zap.String("serviceName", serviceName))

	var f *os.File
	var err error
	for _, path := range []string{"../configs/default.yaml", "/configs/default.yaml", "configs/default.yaml"} {
		f, err = os.Open(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}

	// Tracing with Jaeger .
	tp, err := tracing.NewJaegerProvider(cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Fatal("Failed to shutdown jaeger provider", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// startup test span - temporary debug to verify Jaeger exporter connectivity.
	{
		ctx := context.Background()
		tr := otel.Tracer(serviceName)
		_, span := tr.Start(ctx, "startup-test-span")
		span.End()
		time.Sleep(2 * time.Second)
	}

	// Alerting with Prometheus.
	reporter := prometheus.NewReporter(prometheus.Options{})
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Tags:            map[string]string{"service": serviceName},
		CachedReporter:  reporter,
		SanitizeOptions: &prometheus.DefaultSanitizerOpts,
	}, 10*time.Second)
	defer closer.Close()

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", reporter.HTTPHandler())
	metricsPort := cfg.Prometheus.MetricsPort
	if metricsPort == 0 {
		metricsPort = 8091 // Fallback if missing from YAML
	}
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", metricsPort), metricsMux)
		if err != nil {
			log.Printf("[METRICS] Failed to start metrics handler on port %d: %v", metricsPort, err)
		} else {
			log.Printf("[METRICS] Metrics handler exited cleanly (unexpected)")
		}
	}()

	counter := scope.Tagged(map[string]string{
		"service": serviceName,
	}).Counter("service_started")
	counter.Inc(1)

	// Repository configuration.

	repo, err := postgres.New()
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	cache := memory.New()
	svc := metadata.New(repo, cache)
	h := grpchandler.New(svc)

	cert, err := tls.LoadX509KeyPair("../configs/server.cert", "../configs/server.key")
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}
	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

	// Service Discovery: Consul or Localhost
	ctx := context.Background()
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}
	instanceID := discovery.GenerateInstanceID(serviceName)
	host := os.Getenv("POD_IP")
	if host == "" {
		host = "localhost"
	}
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("%s:%d", host, cfg.API.Port)); err != nil {
		panic(err)
	}
	defer registry.Deregister(ctx, instanceID, serviceName)

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Starting gRPC server.
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.Creds(creds))
	gen.RegisterMetadataServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
