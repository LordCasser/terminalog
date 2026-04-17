package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"terminalog/internal/config"
	"terminalog/internal/handler"
	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

// IntegrationTestEnv represents the test environment.
type IntegrationTestEnv struct {
	Repo    *testutil.TestRepo
	Server  *httptest.Server
	Router  *chi.Mux
	Cleanup func()
}

// SetupIntegrationTest creates a full test environment.
func SetupIntegrationTest(t *testing.T, setup func(repo *testutil.TestRepo) error) *IntegrationTestEnv {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)

	if setup != nil {
		require.NoError(t, setup(repo))
	}

	// Create services
	fileSvc := service.NewFileService(repo.Path)
	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)
	articleSvc := service.NewArticleService(fileSvc, gitSvc)
	assetSvc := service.NewAssetService(fileSvc)

	// Create test auth config
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	require.NoError(t, err)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			Users: []config.User{
				{Username: "testuser", Password: string(hashedPass)},
			},
		},
	}
	authSvc := service.NewAuthService(cfg)

	// Create handlers
	articleHandler := handler.NewArticleHandler(articleSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)
	searchHandler := handler.NewSearchHandler(articleSvc)
	treeHandler := handler.NewTreeHandler(articleSvc)

	// Create router
	router := chi.NewRouter()
	router.Get("/api/articles", articleHandler.List)
	router.Get("/api/articles/{path}", articleHandler.Get)
	router.Get("/api/articles/{path}/timeline", articleHandler.GetTimeline)
	router.Get("/api/tree", treeHandler.Get)
	router.Get("/api/search", searchHandler.Search)
	router.Get("/api/assets/{path}", assetHandler.Get)

	server := httptest.NewServer(router)

	return &IntegrationTestEnv{
		Repo:   repo,
		Server: server,
		Router: router,
		Cleanup: func() {
			server.Close()
			repo.Cleanup()
		},
	}
}

func TestAPI_Articles_List(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		now := time.Now()
		if err := repo.CreateMarkdownFileWithTime("first.md", "# First", "Add", "author", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		return repo.CreateMarkdownFileWithTime("second.md", "# Second", "Add", "author", now.Add(-1*time.Hour))
	})
	defer env.Cleanup()

	// Test default sort (edited desc)
	resp, err := http.Get(env.Server.URL + "/api/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Articles, 2)
	assert.Equal(t, "second.md", result.Articles[0].Path) // Most recent edited first
}

func TestAPI_Articles_Get(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		return repo.CreateMarkdownFile("test.md", "# Test\n\nContent.", "Add", "author")
	})
	defer env.Cleanup()

	resp, err := http.Get(env.Server.URL + "/api/articles/test.md")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var article model.ArticleDetail
	err = json.NewDecoder(resp.Body).Decode(&article)
	require.NoError(t, err)

	assert.Equal(t, "test.md", article.Path)
	assert.Equal(t, "test", article.Title)
	assert.Equal(t, "# Test\n\nContent.", article.Content)
}

func TestAPI_Articles_Timeline(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		return repo.SetupMultiAuthorArticle("article.md", "# Article\nOriginal content.")
	})
	defer env.Cleanup()

	resp, err := http.Get(env.Server.URL + "/api/articles/article.md/timeline")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var timeline model.TimelineResponse
	err = json.NewDecoder(resp.Body).Decode(&timeline)
	require.NoError(t, err)

	assert.Len(t, timeline.Commits, 3)
	assert.Len(t, timeline.Commits[0].Hash, 7) // Short hash
}

func TestAPI_Search(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		if err := repo.CreateMarkdownFile("golang-intro.md", "# Golang Intro", "Add", "author"); err != nil {
			return err
		}
		return repo.CreateMarkdownFile("rust-guide.md", "# Rust Guide", "Add", "author")
	})
	defer env.Cleanup()

	resp, err := http.Get(env.Server.URL + "/api/search?q=golang")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Results, 1)
	assert.Equal(t, "golang-intro.md", result.Results[0].Path)
}

func TestAPI_Tree(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		return repo.SetupNestedDirectory()
	})
	defer env.Cleanup()

	resp, err := http.Get(env.Server.URL + "/api/tree")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.TreeResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "dir", result.Tree.Type)
	assert.Len(t, result.Tree.Children, 3)
}

func TestAPI_Articles_UncommittedNotVisible(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		if err := repo.CreateMarkdownFile("committed.md", "# Committed", "Add", "author"); err != nil {
			return err
		}
		return repo.CreateUncommittedFile("uncommitted.md", "# Uncommitted")
	})
	defer env.Cleanup()

	// List should only show committed
	resp, err := http.Get(env.Server.URL + "/api/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Articles, 1)
	assert.Equal(t, "committed.md", result.Articles[0].Path)

	// Direct access to uncommitted should fail
	resp2, err := http.Get(env.Server.URL + "/api/articles/uncommitted.md")
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

func TestAPI_SortOptions(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		now := time.Now()
		if err := repo.CreateMarkdownFileWithTime("first.md", "# First", "Add", "author", now.Add(-2*time.Hour)); err != nil {
			return err
		}
		return repo.CreateMarkdownFileWithTime("second.md", "# Second", "Add", "author", now.Add(-1*time.Hour))
	})
	defer env.Cleanup()

	tests := []struct {
		name   string
		params string
		check  func(t *testing.T, articles []model.Article)
	}{
		{
			name:   "sort by created asc",
			params: "?sort=created&order=asc",
			check: func(t *testing.T, articles []model.Article) {
				assert.Equal(t, "first.md", articles[0].Path) // Oldest first
			},
		},
		{
			name:   "sort by edited desc (default)",
			params: "",
			check: func(t *testing.T, articles []model.Article) {
				assert.Equal(t, "second.md", articles[0].Path) // Newest first
			},
		},
		{
			name:   "sort by created desc",
			params: "?sort=created&order=desc",
			check: func(t *testing.T, articles []model.Article) {
				assert.Equal(t, "second.md", articles[0].Path) // Newest first
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(env.Server.URL + "/api/articles" + tt.params)
			require.NoError(t, err)
			defer resp.Body.Close()

			var result model.ArticleListResponse
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			tt.check(t, result.Articles)
		})
	}
}

func TestAPI_ErrorResponses(t *testing.T) {
	env := SetupIntegrationTest(t, nil)
	defer env.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "article not found",
			path:       "/api/articles/not-exist.md",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "asset not found",
			path:       "/api/assets/not-exist.png",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "search without query",
			path:       "/api/search",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(env.Server.URL + tt.path)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// Helper function
func (env *IntegrationTestEnv) DoRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, env.Server.URL+path, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return http.DefaultClient.Do(req)
}
