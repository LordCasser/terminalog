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

func TestGitService_GetFileHistory(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(repo *testutil.TestRepo) error
		filePath    string
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, history *model.FileHistory)
	}{
		{
			name: "single commit file",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test\nTest content.", "Add test.md", "author1")
			},
			filePath: "test.md",
			checkResult: func(t *testing.T, history *model.FileHistory) {
				assert.Len(t, history.AllCommits, 1)
				assert.Equal(t, "author1", history.FirstCommit.Author)
				assert.Equal(t, "author1", history.LastCommit.Author)
				assert.Equal(t, []string{"author1"}, history.Contributors)
			},
		},
		{
			name: "multi-author file",
			setup: func(repo *testutil.TestRepo) error {
				return repo.SetupMultiAuthorArticle("article.md", "# Article\nOriginal content.")
			},
			filePath: "article.md",
			checkResult: func(t *testing.T, history *model.FileHistory) {
				assert.Len(t, history.AllCommits, 3)
				assert.Equal(t, "creator", history.FirstCommit.Author)
				assert.Equal(t, "editor2", history.LastCommit.Author)
				assert.ElementsMatch(t, []string{"creator", "editor1", "editor2"}, history.Contributors)
				// Verify order (most recent first)
				assert.True(t, history.AllCommits[0].Timestamp.After(history.AllCommits[1].Timestamp))
			},
		},
		{
			name: "uncommitted file",
			setup: func(repo *testutil.TestRepo) error {
				// First create a committed file to initialize the repo
				if err := repo.CreateMarkdownFile("dummy.md", "# Dummy", "Init", "author"); err != nil {
					return err
				}
				// Then create an uncommitted file
				return repo.CreateUncommittedFile("uncommitted.md", "# Uncommitted")
			},
			filePath:    "uncommitted.md",
			wantErr:     true,
			errContains: "not committed",
		},
		{
			name: "non-existent file",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("exists.md", "# Exists", "Add", "author")
			},
			filePath:    "not-exist.md",
			wantErr:     true,
			errContains: "not committed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := testutil.NewTestRepo()
			require.NoError(t, err)
			defer repo.Cleanup()

			require.NoError(t, tt.setup(repo))

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			history, err := gitSvc.GetFileHistory(context.Background(), tt.filePath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, history)
			}
		})
	}
}

func TestGitService_IsFileCommitted(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(repo *testutil.TestRepo) error
		filePath string
		want     bool
		wantErr  bool
	}{
		{
			name: "committed file",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test", "Add", "author")
			},
			filePath: "test.md",
			want:     true,
		},
		{
			name: "uncommitted file in initialized repo",
			setup: func(repo *testutil.TestRepo) error {
				// First create a committed file to initialize the repo
				if err := repo.CreateMarkdownFile("dummy.md", "# Dummy", "Init", "author"); err != nil {
					return err
				}
				// Then create an uncommitted file
				return repo.CreateUncommittedFile("uncommitted.md", "# Uncommitted")
			},
			filePath: "uncommitted.md",
			want:     false,
		},
		{
			name: "non-existent file in initialized repo",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("exists.md", "# Exists", "Add", "author")
			},
			filePath: "not-exist.md",
			want:     false,
		},
		{
			name: "empty repo returns error",
			setup: func(repo *testutil.TestRepo) error {
				// Do nothing - leave repo empty
				return nil
			},
			filePath: "any.md",
			wantErr:  true, // Empty repo has no HEAD reference
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

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			committed, err := gitSvc.IsFileCommitted(context.Background(), tt.filePath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, committed)
		})
	}
}

func TestGitService_GetRepo(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	require.NoError(t, repo.CreateMarkdownFile("test.md", "# Test", "Add", "author"))

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	assert.NotNil(t, gitSvc.GetRepo())
}

func TestGitService_GetRepoPath(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	require.NoError(t, repo.CreateMarkdownFile("test.md", "# Test", "Add", "author"))

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	assert.Equal(t, repo.Path, gitSvc.GetRepoPath())
}
