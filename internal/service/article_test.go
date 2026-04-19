package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestArticleService_ListArticles(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(repo *testutil.TestRepo) error
		opts     service.ListOptions
		wantLen  int
		wantErr  bool
		checkRes func(t *testing.T, articles []model.Article)
	}{
		{
			name: "default sort (edited desc)",
			setup: func(repo *testutil.TestRepo) error {
				now := time.Now()
				if err := repo.CreateMarkdownFileWithTime("first.md", "# First", "Add first", "author", now.Add(-2*time.Hour)); err != nil {
					return err
				}
				return repo.CreateMarkdownFileWithTime("second.md", "# Second", "Add second", "author", now.Add(-1*time.Hour))
			},
			opts:    service.ListOptions{Dir: "", Sort: model.SortEdited, Order: model.OrderDesc},
			wantLen: 2,
			checkRes: func(t *testing.T, articles []model.Article) {
				// Most recent edited first
				assert.Equal(t, "second.md", articles[0].Path)
				assert.Equal(t, "first.md", articles[1].Path)
			},
		},
		{
			name: "sort by created asc",
			setup: func(repo *testutil.TestRepo) error {
				now := time.Now()
				if err := repo.CreateMarkdownFileWithTime("first.md", "# First", "Add first", "author", now.Add(-2*time.Hour)); err != nil {
					return err
				}
				return repo.CreateMarkdownFileWithTime("second.md", "# Second", "Add second", "author", now.Add(-1*time.Hour))
			},
			opts:    service.ListOptions{Dir: "", Sort: model.SortCreated, Order: model.OrderAsc},
			wantLen: 2,
			checkRes: func(t *testing.T, articles []model.Article) {
				// Oldest created first
				assert.Equal(t, "first.md", articles[0].Path)
				assert.Equal(t, "second.md", articles[1].Path)
			},
		},
		{
			name: "uncommitted file not shown",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateMarkdownFile("committed.md", "# Committed", "Add", "author"); err != nil {
					return err
				}
				return repo.CreateUncommittedFile("uncommitted.md", "# Uncommitted")
			},
			opts:    service.ListOptions{Dir: ""},
			wantLen: 1,
			checkRes: func(t *testing.T, articles []model.Article) {
				assert.Equal(t, "committed.md", articles[0].Path)
			},
		},
		{
			name: "non-markdown file not shown",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateFile("test.md", "# Test"); err != nil {
					return err
				}
				if err := repo.CreateFile("test.txt", "Text file"); err != nil {
					return err
				}
				return repo.Commit("Add files", "author", "author@example.com")
			},
			opts:    service.ListOptions{Dir: ""},
			wantLen: 1,
			checkRes: func(t *testing.T, articles []model.Article) {
				assert.Equal(t, "test.md", articles[0].Path)
			},
		},
		{
			name:    "empty directory",
			setup:   func(repo *testutil.TestRepo) error { return nil },
			opts:    service.ListOptions{Dir: ""},
			wantLen: 0,
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

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			articleSvc := service.NewArticleService(fileSvc, gitSvc)

			articles, err := articleSvc.ListArticles(context.Background(), tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, articles, tt.wantLen)
			if tt.checkRes != nil {
				tt.checkRes(t, articles)
			}
		})
	}
}

func TestArticleService_GetArticle(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(repo *testutil.TestRepo) error
		path     string
		wantErr  error
		checkRes func(t *testing.T, article *model.ArticleDetail)
	}{
		{
			name: "valid article",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test\n\nContent here.", "Add test", "author1")
			},
			path: "test.md",
			checkRes: func(t *testing.T, article *model.ArticleDetail) {
				assert.Equal(t, "test.md", article.Path)
				assert.Equal(t, "test", article.Title)
				assert.Equal(t, "# Test\n\nContent here.", article.Content)
				assert.Equal(t, "author1", article.CreatedBy)
				assert.Equal(t, "author1", article.EditedBy)
			},
		},
		{
			name: "uncommitted article",
			setup: func(repo *testutil.TestRepo) error {
				// First create a committed file to initialize the repo
				if err := repo.CreateMarkdownFile("dummy.md", "# Dummy", "Init", "author"); err != nil {
					return err
				}
				// Then create an uncommitted file
				return repo.CreateUncommittedFile("uncommitted.md", "# Uncommitted")
			},
			path:    "uncommitted.md",
			wantErr: model.ErrNotCommitted,
		},
		{
			name: "non-existent article",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("exists.md", "# Exists", "Add", "author")
			},
			path:    "not-exist.md",
			wantErr: model.ErrNotCommitted,
		},
		{
			name: "multi-author article",
			setup: func(repo *testutil.TestRepo) error {
				return repo.SetupMultiAuthorArticle("article.md", "# Article\nOriginal content.")
			},
			path: "article.md",
			checkRes: func(t *testing.T, article *model.ArticleDetail) {
				assert.Equal(t, "creator", article.CreatedBy)
				assert.Equal(t, "editor2", article.EditedBy)
				assert.ElementsMatch(t, []string{"creator", "editor1", "editor2"}, article.Contributors)
			},
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

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			articleSvc := service.NewArticleService(fileSvc, gitSvc)

			article, err := articleSvc.GetArticle(context.Background(), tt.path)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			if tt.checkRes != nil {
				tt.checkRes(t, article)
			}
		})
	}
}

func TestArticleService_GetTimeline(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(repo *testutil.TestRepo) error
		path     string
		wantErr  bool
		checkRes func(t *testing.T, commits []model.CommitInfo)
	}{
		{
			name: "single commit",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test", "Add test", "author")
			},
			path: "test.md",
			checkRes: func(t *testing.T, commits []model.CommitInfo) {
				assert.Len(t, commits, 1)
				assert.Len(t, commits[0].Hash, 7) // short hash
				assert.Equal(t, "author", commits[0].Author)
			},
		},
		{
			name: "multiple commits",
			setup: func(repo *testutil.TestRepo) error {
				return repo.SetupMultiAuthorArticle("article.md", "# Article\nOriginal content.")
			},
			path: "article.md",
			checkRes: func(t *testing.T, commits []model.CommitInfo) {
				assert.Len(t, commits, 3)
				// Check order (most recent first)
				assert.True(t, commits[0].Timestamp.After(commits[1].Timestamp))
			},
		},
		{
			name:    "non-existent file",
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

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			articleSvc := service.NewArticleService(fileSvc, gitSvc)

			commits, err := articleSvc.GetTimeline(context.Background(), tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkRes != nil {
				tt.checkRes(t, commits)
			}
		})
	}
}

func TestArticleService_ListDirectory_AfterDeleteAndInvalidate(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()

	now := time.Now()
	require.NoError(t, repo.CreateMarkdownFileWithTime("hardware_security/FragAttack.md", "# FragAttack", "Add FragAttack", "author", now.Add(-2*time.Hour)))

	fileSvc, err := service.NewFileService(repo.Path)
	require.NoError(t, err)

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	articleSvc := service.NewArticleService(fileSvc, gitSvc)
	ctx := context.Background()

	articles, err := articleSvc.ListDirectory(ctx, "hardware_security", model.SortName, model.OrderAsc)
	require.NoError(t, err)
	require.Len(t, articles, 1)
	assert.Equal(t, "hardware_security/FragAttack.md", articles[0].Path)

	wt, err := repo.Repo.Worktree()
	require.NoError(t, err)

	_, err = wt.Remove("hardware_security/FragAttack.md")
	require.NoError(t, err)

	_, err = wt.Commit("Delete FragAttack", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "author",
			Email: "author@example.com",
			When:  now.Add(-1 * time.Hour),
		},
	})
	require.NoError(t, err)

	// Directory listing is cached until explicitly invalidated after a repo update.
	cachedArticles, err := articleSvc.ListDirectory(ctx, "hardware_security", model.SortName, model.OrderAsc)
	require.NoError(t, err)
	require.Len(t, cachedArticles, 1)

	articleSvc.InvalidateCache()

	updatedArticles, err := articleSvc.ListDirectory(ctx, "hardware_security", model.SortName, model.OrderAsc)
	require.NoError(t, err)
	assert.Empty(t, updatedArticles)
}

func TestArticleService_Search(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(repo *testutil.TestRepo) error
		query    string
		dir      string
		wantLen  int
		checkRes func(t *testing.T, results []model.SearchResult)
	}{
		{
			name: "search by title",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateMarkdownFile("golang-intro.md", "# Golang Intro", "Add", "author"); err != nil {
					return err
				}
				if err := repo.CreateMarkdownFile("rust-guide.md", "# Rust Guide", "Add", "author"); err != nil {
					return err
				}
				return repo.CreateMarkdownFile("python-tutorial.md", "# Python Tutorial", "Add", "author")
			},
			query:   "golang",
			dir:     "",
			wantLen: 1,
			checkRes: func(t *testing.T, results []model.SearchResult) {
				assert.Equal(t, "golang-intro.md", results[0].Path)
				assert.Contains(t, results[0].Title, "golang")
			},
		},
		{
			name: "search multiple matches",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateMarkdownFile("golang-intro.md", "# Golang Intro", "Add", "author"); err != nil {
					return err
				}
				if err := repo.CreateMarkdownFile("golang-advanced.md", "# Golang Advanced", "Add", "author"); err != nil {
					return err
				}
				return repo.CreateMarkdownFile("rust-guide.md", "# Rust Guide", "Add", "author")
			},
			query:   "golang",
			dir:     "",
			wantLen: 2,
		},
		{
			name: "case insensitive",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("GOLANG-intro.md", "# GOLANG Intro", "Add", "author")
			},
			query:   "golang",
			dir:     "",
			wantLen: 1,
		},
		{
			name: "no matches",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("test.md", "# Test", "Add", "author")
			},
			query:   "nonexistent",
			dir:     "",
			wantLen: 0,
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

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			articleSvc := service.NewArticleService(fileSvc, gitSvc)

			results, err := articleSvc.Search(context.Background(), tt.query, tt.dir)
			require.NoError(t, err)
			assert.Len(t, results, tt.wantLen)
			if tt.checkRes != nil {
				tt.checkRes(t, results)
			}
		})
	}
}
