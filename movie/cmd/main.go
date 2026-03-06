package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"movieexample.com/gen"
	controller "movieexample.com/movie/internal/controller/movie"
	metadatagateway "movieexample.com/movie/internal/gateway/metadata/grpc"
	ratinggateway "movieexample.com/movie/internal/gateway/rating/grpc"
	grpchandler "movieexample.com/movie/internal/handler/grpc"
	memoryregistry "movieexample.com/pkg/discovery/memorypackage"
)

func main() {
	log.Println("Starting the movie service")

	registry := memoryregistry.NewRegistry()

	ctx := context.Background()
	if err := registry.Register(ctx, "movie-1", "movie", "localhost:8083"); err != nil {
		panic(err)
	}
	defer registry.Deregister(ctx, "movie-1", "movie")

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	svc := controller.New(ratingGateway, metadataGateway)
	h := grpchandler.New(svc)

	lis, err := net.Listen("tcp", "localhost:8083")
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
