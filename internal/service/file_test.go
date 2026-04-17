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

func TestFileService_SpecialFileFilter(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(repo *testutil.TestRepo) error
		dir          string
		wantFiles    []string
		wantNotFiles []string
		wantErr      bool
	}{
		{
			name: "skip files starting with underscore",
			setup: func(repo *testutil.TestRepo) error {
				// Create regular files
				if err := repo.CreateMarkdownFile("article1.md", "# Article 1", "Add article1", "author"); err != nil {
					return err
				}
				// Create special file that should be filtered
				if err := repo.CreateFile("_ABOUTME.md", "# About Me\nThis is about me."); err != nil {
					return err
				}
				// Commit all files
				return repo.Commit("Initial commit", "author", "author@example.com")
			},
			dir:          "",
			wantFiles:    []string{"article1.md"},
			wantNotFiles: []string{"_ABOUTME.md"},
		},
		{
			name: "skip special directory",
			setup: func(repo *testutil.TestRepo) error {
				// Create regular files
				if err := repo.CreateMarkdownFile("article1.md", "# Article 1", "Add article1", "author"); err != nil {
					return err
				}
				// Create special directory with files (should be skipped entirely)
				if err := repo.CreateFile("_special/hidden.md", "# Hidden"); err != nil {
					return err
				}
				// Commit all files
				return repo.Commit("Initial commit", "author", "author@example.com")
			},
			dir:          "",
			wantFiles:    []string{"article1.md"},
			wantNotFiles: []string{"_special/hidden.md"},
		},
		{
			name: "include regular files only",
			setup: func(repo *testutil.TestRepo) error {
				return repo.SetupNestedDirectory()
			},
			dir:       "",
			wantFiles: []string{"tech/golang.md", "tech/rust.md", "life/travel.md", "welcome.md"},
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

			// Check expected files are present
			for _, wantFile := range tt.wantFiles {
				assert.Contains(t, files, wantFile, "expected file %s to be in list", wantFile)
			}

			// Check unwanted files are NOT present
			for _, notWantFile := range tt.wantNotFiles {
				assert.NotContains(t, files, notWantFile, "expected special file %s to be filtered out", notWantFile)
			}
		})
	}
}

func TestFileService_ReadSpecialFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(repo *testutil.TestRepo) error
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "read _ABOUTME.md",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateFile("_ABOUTME.md", "# About Me\nContent."); err != nil {
					return err
				}
				return repo.Commit("Add about me", "author", "author@example.com")
			},
			path: "_ABOUTME.md",
			want: "# About Me\nContent.",
		},
		{
			name: "read non-special file should fail",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("article.md", "# Article", "Add", "author")
			},
			path:    "article.md",
			wantErr: true, // Should fail because it doesn't start with "_"
		},
		{
			name: "read non-existent special file",
			setup: func(repo *testutil.TestRepo) error {
				return nil
			},
			path:    "_NOTEXIST.md",
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

			content, err := fileSvc.ReadSpecialFile(context.Background(), tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, string(content))
		})
	}
}
