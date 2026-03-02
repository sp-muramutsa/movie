package memory

import (
	"context"
	"sync"

	"movieapp.com/metadata/internal/repository"
	"movieapp.com/metadata/pkg/model"
	model "movieexample.com/metadata/pkg"
)

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
func (*r Repository) Get(_ context.Context, id string) (*model.Metadata, error) {
	r.RLock()
	defer r.Unlock()
	m, ok := data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return m, nil
}

// Puts a movie metadata in the repository
func (*r Repository) Put(_ context.Context, id string, metadata *model.Metadata) error {
	r.Lock()
	defer r.Unlock()
	r.data[id] = metadata
	return nil
}

