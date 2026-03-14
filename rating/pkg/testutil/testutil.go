package testutil

import (
	"movieexample.com/gen"
	"movieexample.com/rating/internal/controller/rating"
	grpchandler "movieexample.com/rating/internal/handler/grpc"
	"movieexample.com/rating/internal/ingester/kafka"
	"movieexample.com/rating/internal/repository/memory"
)

// NewTestRatingGRPCServer creates a new rating gRPC server to be used in tests.
func NewTestRatingGRPCServer() gen.RatingServiceServer {
	repo := memory.New()
	ingester, err := kafka.NewIngester("localhost:9092", "ratings")
	if err != nil {
		panic(err)
	}
	ctrl := rating.New(repo, ingester)
	return grpchandler.New(ctrl)
}
