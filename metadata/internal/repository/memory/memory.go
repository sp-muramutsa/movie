package memory

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

const tracerID = "metadata-repository-memory"

// Repository defines a memory movie metadata Repository
type Repository struct {
	sync.RWMutex
	data map[string]*model.Metadata
}

// New creates a new memory Repository
func New() *Repository {
	return &Repository{data: map[string]*model.Metadata{}}
}

// Get returns a movie metadata by movie id
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	r.RLock()
	defer r.RUnlock()

	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Get")
	defer span.End()

	m, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// Puts a movie metadata in the repository.
func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	r.Lock()
	defer r.Unlock()

	_, span := otel.Tracer(tracerID).Start(ctx, "Repository/Put")
	defer span.End()

	r.data[id] = metadata
	return nil
}
