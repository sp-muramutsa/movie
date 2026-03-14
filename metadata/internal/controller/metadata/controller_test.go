package metadata

import (
	context "context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	gen "movieexample.com/gen/mock/metadata/repository"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

func TestController(t *testing.T) {
	tests := []struct {
		name       string
		expRepoRes *model.Metadata
		expRepoErr error
		wantRes    *model.Metadata
		wantErr    error
	}{
		{
			name:       "not found",
			expRepoErr: repository.ErrNotFound,
			wantErr:    ErrNotFound,
		},
		{
			name:       "unexpected error",
			expRepoErr: errors.New("unexpected error"),
			wantErr:    errors.New("unexpected error"),
		},
		{
			name:       "success",
			expRepoRes: &model.Metadata{},
			wantRes:    &model.Metadata{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repoMock := gen.NewMockmetadataRepository(ctrl)
			cache := gen.NewMockmetadataRepository(ctrl)
			c := New(repoMock, cache)
			ctx := context.Background()
			id := "id"
			// cache is checked first; have it miss so controller will call repo
			cache.EXPECT().Get(ctx, id).Return(nil, errors.New("cache miss"))
			repoMock.EXPECT().Get(ctx, id).Return(tt.expRepoRes, tt.expRepoErr)
			// After repo get, controller will try to update cache for non-NotFound paths
			if !errors.Is(tt.expRepoErr, repository.ErrNotFound) {
				cache.EXPECT().Put(ctx, id, tt.expRepoRes).Return(nil)
			}
			res, err := c.Get(ctx, id)
			// result
			assert.Equal(t, tt.wantRes, res, tt.name)
			// error assertions: handle sentinel ErrNotFound specially, otherwise compare messages
			if tt.wantErr == nil {
				assert.NoError(t, err, tt.name)
			} else if errors.Is(tt.wantErr, ErrNotFound) {
				assert.Error(t, err, tt.name)
				assert.True(t, errors.Is(err, ErrNotFound), "expected ErrNotFound")
			} else {
				assert.Error(t, err, tt.name)
				assert.Contains(t, err.Error(), tt.wantErr.Error(), tt.name)
			}
		})
	}
}
