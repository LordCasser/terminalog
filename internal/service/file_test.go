package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestFileService_ScanMarkdownFiles(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(repo *testutil.TestRepo) error
		dir     string
		wantLen int
		wantErr bool
	}{
		{
			name:    "root directory with files",
			setup:   func(repo *testutil.TestRepo) error { return repo.CreateMarkdownFile("a.md", "# A", "Add a", "author") },
			dir:     "",
			wantLen: 1,
		},
		{
			name:    "nested directory",
			setup:   func(repo *testutil.TestRepo) error { return repo.SetupNestedDirectory() },
			dir:     "",
			wantLen: 4, // tech/golang.md, tech/rust.md, life/travel.md, welcome.md
		},
		{
			name:    "empty directory",
			setup:   func(repo *testutil.TestRepo) error { return nil },
			dir:     "",
			wantLen: 0,
		},
		{
			name:    "non-existent directory",
			setup:   func(repo *testutil.TestRepo) error { return nil },
			dir:     "nonexistent",
			wantErr: true,
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

			files, err := fileSvc.ScanMarkdownFiles(context.Background(), tt.dir)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, files, tt.wantLen)
		})
	}
}

func TestFileService_ReadFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(repo *testutil.TestRepo) error
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "read markdown file",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test\nContent.", "Add", "author")
			},
			path: "test.md",
			want: "# Test\nContent.",
		},
		{
			name:  "read nested file",
			setup: func(repo *testutil.TestRepo) error { return repo.SetupNestedDirectory() },
			path:  "tech/golang.md",
			want:  "# Golang\nGolang content.",
		},
		{
			name:    "read non-existent file",
			setup:   func(repo *testutil.TestRepo) error { return nil },
			path:    "not-exist.md",
			wantErr: true,
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

			content, err := fileSvc.ReadFile(context.Background(), tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(content))
		})
	}
}

func TestFileService_PathTraversal(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	// Create a sensitive file outside the repo
	sensitivePath := filepath.Join(repo.Path, "..", "sensitive.txt")
	os.WriteFile(sensitivePath, []byte("secret"), 0644)

	require.NoError(t, repo.CreateMarkdownFile("test.md", "# Test", "Add", "author"))

	fileSvc, err := service.NewFileService(repo.Path)
	require.NoError(t, err)

	// Try to access file outside the repo
	_, err = fileSvc.ReadFile(context.Background(), "../sensitive.txt")
	assert.ErrorIs(t, err, model.ErrInvalidPath)
}
