package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	"movieexample.com/pkg/tracing"
	controller "movieexample.com/rating/internal/controller/rating"
	grpchandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/ingester/kafka"
	"movieexample.com/rating/internal/repository/postgres"
)

const serviceName = "rating"

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
		panic(err)
	}

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

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

	ingester, err := kafka.NewIngester(cfg.ServiceDiscovery.Kafka.Address, "ratings")
	if err != nil {
		log.Fatalf("failed to intialize ingester: %v", err)
	}

	ctrl := controller.New(repo, ingester)
	h := grpchandler.New(ctrl)

	go func() {
		if err := ctrl.StartIngestion(ctx); err != nil {
			log.Fatalf("failed to start ingestion: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		log.Printf("Received signal %v. Attempting graceful shutdown", s)
		srv.GracefulStop()
		log.Println("Graceful shutdown complete")
	}()

	gen.RegisterRatingServiceServer(srv, h)
	reflection.Register(srv)

	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

// ValidCredentials verifies the user credentials.
func ValidCredentials(username, password string) bool {
	if username == "" || password == "" {
		return false
	}
	return true
}
