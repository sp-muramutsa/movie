package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/pkg/discovery"
	"movieexample.com/pkg/discovery/consul"
	controller "movieexample.com/rating/internal/controller/rating"
	grpchandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/ingester/kafka"
	"movieexample.com/rating/internal/repository/postgres"
)

const serviceName = "rating"

func main() {
	log.Println("Starting the rating service")

	f, err := os.Open("../configs/default.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	repo, err := postgres.New()
	if err != nil {
		panic(err)
	}

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
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
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)
	reflection.Register(srv)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
