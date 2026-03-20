package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"time"

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

	f, err := os.Open("../configs/default.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

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

	// Service Discovery
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
