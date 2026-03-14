package metadata

import (
	"context"
	"errors"
	"fmt"

	"movieexample.com/metadata/internal/repository"
	model "movieexample.com/metadata/pkg/model"
)

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, id string, metadata *model.Metadata) error
}

// Controller defines a metadata service controller.
type Controller struct {
	repo  metadataRepository
	cache metadataRepository
}

// New creates a metadata service controller.
func New(repo metadataRepository, cache metadataRepository) *Controller {
	return &Controller{repo, cache}
}

// Get returns a movie metadata by id
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {

	cacheData, err := c.cache.Get(ctx, id)
	if err == nil {
		fmt.Println("Returning metadata from cache for" + id)
		return cacheData, nil
	}

	repoData, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err := c.cache.Put(ctx, id, repoData); err != nil {
		fmt.Println("Error updatring cache: " + err.Error())
	}

	return repoData, err

}

// Put writes movie metadata to repository.
func (c *Controller) Put(ctx context.Context, m *model.Metadata) error {
	return c.repo.Put(ctx, m.ID, m)
}
