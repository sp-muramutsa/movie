package metadata

import (
	"context"
	"errors"

	"movieexample.com/metadata/internal/repository"
	model "movieexample.com/metadata/pkg/model"
)

// ErrNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
}

// Controller defines a metadata service controller.
type Controller struct {
	repo metadataRepository
}

// New createds a metadata service controller.
func New(repo metadataRepository) *Controller {
	return &Controller{repo}
}

// Get returns a movie metadata by id
func (c *Controller) Get(ctx context.Context, id string) (*model.Metadata, error) {
	movie, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return movie, err
}