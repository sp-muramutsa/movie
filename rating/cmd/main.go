package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"movieexample.com/gen"
	controller "movieexample.com/rating/internal/controller/rating"
	grpchandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/ingester/kafka"
	"movieexample.com/rating/internal/repository/memory"
)

func main() {
	log.Println("Starting the rating service")
	repo := memory.New()

	ingester, err := kafka.NewIngester("localhost", "rating", "ratings")
	if err != nil {
		log.Fatalf("failed to intialize ingester: %v", err)
	}

	ctrl := controller.New(repo, ingester)
	h := grpchandler.New(ctrl)
	
	ctx := context.Background()
	if err := ctrl.StartIngestion(ctx); err != nil {
		log.Fatalf("failed to start ingestion: %v", err)
	}

	lis, err := net.Listen("tcp", "localhost:8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
