package integration_test

import (
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
	fileSvc, err := service.NewFileService(repo.Path)
	require.NoError(t, err)

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	articleSvc := service.NewArticleService(fileSvc, gitSvc)
	assetSvc := service.NewAssetService(fileSvc)
	versionSvc := service.NewVersionService(articleSvc, gitSvc, fileSvc) // v1.2

	// Create test auth config
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	require.NoError(t, err)

	cfg := &config.Config{
		Auth: config.AuthConfig{
			Users: []config.UserConfig{
				{Username: "testuser", Password: string(hashedPass)},
			},
		},
	}
	authSvc := service.NewAuthService(cfg)
	_ = authSvc // unused for now

	// Create handlers
	articleHandler := handler.NewArticleHandler(articleSvc, versionSvc, fileSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)
	searchHandler := handler.NewSearchHandler(articleSvc)
	treeHandler := handler.NewTreeHandler(articleSvc)

	// Create router (RESTful v1)
	router := chi.NewRouter()
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/articles", articleHandler.ListRoot)
		r.Get("/articles/*", articleHandler.HandleRequest)
		r.Get("/tree", treeHandler.Get)
		r.Get("/search", searchHandler.Search)
		r.Get("/assets/*", assetHandler.Get)
	})

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

	// Test root directory listing (alphabetical order: dirs first, then files)
	resp, err := http.Get(env.Server.URL + "/api/v1/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Articles, 2)
	// Files are sorted alphabetically: first.md before second.md
	assert.Equal(t, "first.md", result.Articles[0].Path)
	assert.Equal(t, "second.md", result.Articles[1].Path)
}

func TestAPI_Articles_Get(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		return repo.CreateMarkdownFile("test.md", "# Test\n\nContent.", "Add", "author")
	})
	defer env.Cleanup()

	resp, err := http.Get(env.Server.URL + "/api/v1/articles/test.md")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var article model.ArticleResponse
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

	resp, err := http.Get(env.Server.URL + "/api/v1/articles/article.md/timeline")
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

	resp, err := http.Get(env.Server.URL + "/api/v1/search?q=golang")
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

	resp, err := http.Get(env.Server.URL + "/api/v1/tree")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.TreeResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, model.NodeTypeDir, result.Root.Type)
	assert.Len(t, result.Root.Children, 3)
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
	resp, err := http.Get(env.Server.URL + "/api/v1/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Articles, 1)
	assert.Equal(t, "committed.md", result.Articles[0].Path)

	// Direct access to uncommitted should fail
	resp2, err := http.Get(env.Server.URL + "/api/v1/articles/uncommitted.md")
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

func TestAPI_DirectoryListingOrder(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		// Create a directory with markdown files and subdirectories
		if err := repo.CreateMarkdownFile("alpha.md", "# Alpha", "Add alpha", "author"); err != nil {
			return err
		}
		if err := repo.CreateMarkdownFile("beta.md", "# Beta", "Add beta", "author"); err != nil {
			return err
		}
		if err := repo.CreateFile("tech/golang.md", "# Golang"); err != nil {
			return err
		}
		return repo.Commit("Add all", "author", "author@example.com")
	})
	defer env.Cleanup()

	// Test: Root listing returns dirs first alphabetically, then files alphabetically (default sort=name asc)
	resp, err := http.Get(env.Server.URL + "/api/v1/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify: 3 items (tech dir, alpha.md file, beta.md file)
	assert.Len(t, result.Articles, 3)

	// Verify: Directories come first, then files, both sorted alphabetically
	assert.Equal(t, model.NodeTypeDir, result.Articles[0].Type)
	assert.Equal(t, "tech", result.Articles[0].Path)

	assert.Equal(t, model.NodeTypeFile, result.Articles[1].Type)
	assert.Equal(t, "alpha.md", result.Articles[1].Path)

	assert.Equal(t, model.NodeTypeFile, result.Articles[2].Type)
	assert.Equal(t, "beta.md", result.Articles[2].Path)

	// Test: Sort by edited time descending (most recently edited first)
	resp2, err := http.Get(env.Server.URL + "/api/v1/articles?sort=edited&order=desc")
	require.NoError(t, err)
	defer resp2.Body.Close()

	var editedDescResult model.ArticleListResponse
	err = json.NewDecoder(resp2.Body).Decode(&editedDescResult)
	require.NoError(t, err)

	assert.Len(t, editedDescResult.Articles, 3)
	assert.Equal(t, model.NodeTypeDir, editedDescResult.Articles[0].Type)
	// beta.md committed after alpha.md (edited desc: beta first)
	assert.Equal(t, "beta.md", editedDescResult.Articles[1].Path)
	assert.Equal(t, "alpha.md", editedDescResult.Articles[2].Path)

	// Test: Sort by created time ascending (oldest first)
	resp3, err := http.Get(env.Server.URL + "/api/v1/articles?sort=created&order=asc")
	require.NoError(t, err)
	defer resp3.Body.Close()

	var createdAscResult model.ArticleListResponse
	err = json.NewDecoder(resp3.Body).Decode(&createdAscResult)
	require.NoError(t, err)

	assert.Len(t, createdAscResult.Articles, 3)
	assert.Equal(t, model.NodeTypeDir, createdAscResult.Articles[0].Type)
	// alpha.md created first (asc: alpha before beta)
	assert.Equal(t, "alpha.md", createdAscResult.Articles[1].Path)
	assert.Equal(t, "beta.md", createdAscResult.Articles[2].Path)
}

func TestAPI_ErrorResponses(t *testing.T) {
	env := SetupIntegrationTest(t, func(repo *testutil.TestRepo) error {
		// Create a file to initialize the repo
		return repo.CreateMarkdownFile("dummy.md", "# Dummy", "Add", "author")
	})
	defer env.Cleanup()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "article not found",
			path:       "/api/v1/articles/not-exist.md",
			wantStatus: http.StatusNotFound, // Nonexistent path returns 404
		},
		{
			name:       "asset not found",
			path:       "/api/v1/assets/not-exist.png",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "search without query",
			path:       "/api/v1/search",
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
