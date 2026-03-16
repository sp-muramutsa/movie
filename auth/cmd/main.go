package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	grpchandler "movieexample.com/auth/internal/handler/grpc"
	"movieexample.com/gen"
)

func main() {
	port := 8084
	log.Printf("Starting grpc server on port %d", port)

	cert, err := tls.LoadX509KeyPair("../configs/server.cert", "../configs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	h := grpchandler.New(func() []byte {
		return []byte("test-secret")
	})

	srv := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(srv)
	gen.RegisterAuthServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
