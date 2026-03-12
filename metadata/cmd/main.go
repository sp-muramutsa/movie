package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
	"movieexample.com/gen"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	memory "movieexample.com/metadata/internal/repository/memory"
)

func main() {
	log.Println("Starting the movie metadata service")
	f, err := os.Open("../configs/default.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	repo := memory.New()
	cache := memory.New()
	svc := metadata.New(repo, cache)
	h := grpchandler.New(svc)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	gen.RegisterMetadataServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
