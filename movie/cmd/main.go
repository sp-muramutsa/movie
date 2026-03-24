package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	controller "movieexample.com/movie/internal/controller/movie"
	metadatagateway "movieexample.com/movie/internal/gateway/metadata/grpc"
	ratinggateway "movieexample.com/movie/internal/gateway/rating/grpc"
	grpchandler "movieexample.com/movie/internal/handler/grpc"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/tracing"
)

const serviceName = "movie"

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

	// Tracing with Jaeger.
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

	// startup test span - temporary debug to verify Jaeger exporter connectivity
	{
		ctx := context.Background()
		tr := otel.Tracer(serviceName)
		_, span := tr.Start(ctx, "startup-test-span")
		span.End()
		// give the batcher a short moment to export the span
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
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", metricsPort), metricsMux); err != nil {
			logger.Fatal("Failed to start the metrics handler", zap.Error(err))
		}
	}()

	counter := scope.Tagged(map[string]string{
		"service": serviceName,
	}).Counter("service_started")
	counter.Inc(1)

	ctx := context.Background()
	port := cfg.API.Port

	// Service Discovery: Consul or Localhost
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)
	host := os.Getenv("POD_IP")
	if host == "" {
		host = "localhost"
	}
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("%s:%d", host, port)); err != nil {
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

	// Setup TLS.
	certBytes, err := os.ReadFile("../configs/server.cert")
	if err != nil {
		log.Fatalf("Failed to read server certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certBytes) {
		log.Fatalf("Failed to append certificate to pool: %v", err)
	}

	cert, err := tls.LoadX509KeyPair("../configs/server.cert", "../configs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	})

	metadataGateway := metadatagateway.New(registry, creds)
	ratingGateway := ratinggateway.New(registry)
	svc := controller.New(ratingGateway, metadataGateway)
	h := grpchandler.New(svc)

	// Starting gRPC server.
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		panic(err)
	}

	// Rate Limiting Interceptor
	l := newLimiter(100, 10)
	srv := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			ratelimit.UnaryServerInterceptor(l),
		),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}

type limiter struct {
	l *rate.Limiter
}

func newLimiter(limit, burst int) *limiter {
	return &limiter{rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *limiter) Limit() bool {
	return l.l.Allow()
}
