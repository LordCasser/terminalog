package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestAssetService_GetAsset(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(repo *testutil.TestRepo) error
		path     string
		wantErr  error
		wantMime string
	}{
		{
			name: "png image",
			setup: func(repo *testutil.TestRepo) error {
				// Create a minimal PNG (1x1 pixel) in .assets directory
				pngData := []byte{
					0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
				}
				return repo.CreateImageFileAndCommit(".assets/images/test.png", pngData, "Add image", "author")
			},
			path:     "images/test.png", // Request without .assets prefix
			wantMime: "image/png",
		},
		{
			name: "jpeg image",
			setup: func(repo *testutil.TestRepo) error {
				jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
				return repo.CreateImageFileAndCommit(".assets/images/test.jpg", jpegData, "Add image", "author")
			},
			path:     "images/test.jpg", // Request without .assets prefix
			wantMime: "image/jpeg",
		},
		{
			name: "svg image",
			setup: func(repo *testutil.TestRepo) error {
				svgData := []byte("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"1\" height=\"1\"></svg>")
				return repo.CreateImageFileAndCommit(".assets/images/test.svg", svgData, "Add image", "author")
			},
			path:     "images/test.svg", // Request without .assets prefix
			wantMime: "image/svg+xml",
		},
		{
			name:    "non-existent asset",
			setup:   func(repo *testutil.TestRepo) error { return nil },
			path:    "images/not-exist.png",
			wantErr: model.ErrNotFound,
		},
		{
			name: "path traversal attack",
			setup: func(repo *testutil.TestRepo) error {
				pngData := []byte{0x89, 0x50, 0x4E, 0x47}
				return repo.CreateImageFileAndCommit("images/test.png", pngData, "Add image", "author")
			},
			path:    "../sensitive.png",
			wantErr: model.ErrInvalidPath,
		},
		{
			name: ".git directory access blocked",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test", "Add", "author")
			},
			path:    ".git/config",
			wantErr: model.ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := testutil.NewTestRepo()
			require.NoError(t, err)
			defer repo.Cleanup()

			if tt.setup != nil {
				require.NoError(t, tt.setup(repo))
			}

			fileSvc, err := service.NewFileService(repo.Path)
			require.NoError(t, err)

			assetSvc := service.NewAssetService(fileSvc)

			asset, err := assetSvc.GetAsset(context.Background(), tt.path)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMime, asset.ContentType)
			assert.Greater(t, asset.Size, int64(0))
		})
	}
}
