package testutil

import (
	"github.com/uber-go/tally/v4"
	"movieexample.com/gen"
	"movieexample.com/metadata/internal/controller/metadata"
	grpchandler "movieexample.com/metadata/internal/handler/grpc"
	"movieexample.com/metadata/internal/repository/memory"
)

// NewTestMetadataGRPCServer creates a new metadata gRPC server to be used in tests.
func NewTestMetadataGRPCServer() gen.MetadataServiceServer {
	repo := memory.New()
	cache := memory.New()
	ctrl := metadata.New(repo, cache)
	return grpchandler.New(ctrl, tally.NoopScope)
}
