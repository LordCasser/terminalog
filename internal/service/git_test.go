package service_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func TestGitService_PublishedTreeIgnoresWorktreeChanges(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()
	require.NoError(t, repo.CreateMarkdownFile("published.md", "# Published", "Add", "author"))

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(repo.Path, "published.md"), []byte("# Draft"), 0644))
	content, err := gitSvc.ReadFileAtHead("published.md")
	require.NoError(t, err)
	assert.Equal(t, "# Published", string(content))

	require.NoError(t, os.Remove(filepath.Join(repo.Path, "published.md")))
	files, err := gitSvc.ListMarkdownFilesAtHead("")
	require.NoError(t, err)
	assert.Equal(t, []string{"published.md"}, files)
}

func TestGitService_CurrentHead(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	require.NoError(t, repo.CreateMarkdownFile("test.md", "# Test", "Add", "author"))

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	head, err := gitSvc.CurrentHead()
	require.NoError(t, err)
	require.Len(t, head, 40)
	ref, err := repo.Repo.Head()
	require.NoError(t, err)
	assert.Equal(t, ref.Hash().String(), head)
}

func TestGitService_PushUpdatesCheckedOutWorktree(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	require.NoError(t, repo.CreateMarkdownFile("existing.md", "# Existing", "Add existing", "author"))

	_, err = service.NewGitService(repo.Path)
	require.NoError(t, err)

	clientDir := filepath.Join(t.TempDir(), "client")
	runGit(t, "", "clone", repo.Path, clientDir)
	runGit(t, clientDir, "config", "user.email", "push@example.com")
	runGit(t, clientDir, "config", "user.name", "Pusher")

	require.NoError(t, os.WriteFile(filepath.Join(clientDir, "new-article.md"), []byte("# New Article\n"), 0644))
	runGit(t, clientDir, "add", "new-article.md")
	runGit(t, clientDir, "commit", "-m", "Add new article")

	branch := currentBranch(t, repo.Path)
	runGit(t, clientDir, "push", "origin", "HEAD:"+branch)

	// receive.denyCurrentBranch=updateInstead makes worktree publication part
	// of receive-pack; no post-push reset or garbage collection is required.
	content, err := os.ReadFile(filepath.Join(repo.Path, "new-article.md"))
	require.NoError(t, err)
	assert.Equal(t, "# New Article\n", string(content))
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, output)
}

func currentBranch(t *testing.T, dir string) string {
	t.Helper()

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	output, err := cmd.Output()
	require.NoError(t, err)
	return strings.TrimSpace(string(output))
}

func TestGitService_ConcurrentReload(t *testing.T) {
	// This test verifies that concurrent reads during ReloadRepo do not panic
	// or produce incorrect results.
	// Run with: go test -race ./internal/service/ -run TestGitService_ConcurrentReload

	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	require.NoError(t, repo.CreateMarkdownFile("test.md", "# Test", "Add test.md", "author"))

	svc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	var wg sync.WaitGroup
	ctx := context.Background()

	// Launch 10 concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = svc.GetFileHistory(ctx, "test.md")
		}()
	}

	// Launch 3 concurrent reloads
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = svc.ReloadRepo()
		}()
	}

	wg.Wait()
}
