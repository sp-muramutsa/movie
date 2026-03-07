package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"movieexample.com/gen"
	controller "movieexample.com/rating/internal/controller/rating"
	grpchandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/ingester/kafka"
	"movieexample.com/rating/internal/repository/postgres"
)

func main() {
	log.Println("Starting the rating service")
	repo, err := postgres.New()
	if err != nil {
		panic(err)
	}

	ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	if err != nil {
		log.Fatalf("failed to intialize ingester: %v", err)
	}

	ctrl := controller.New(repo, ingester)
	h := grpchandler.New(ctrl)

	ctx := context.Background()

	go func() {
		if err := ctrl.StartIngestion(ctx); err != nil {
			log.Fatalf("failed to start ingestion: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", "localhost:8082")
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
