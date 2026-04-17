// Package scenario_test provides comprehensive scenario-based tests for Terminalog backend.
// These tests verify all requirements from requirements.md and api-spec.md.
package scenario_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"terminalog/internal/config"
	"terminalog/internal/handler"
	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

// =============================================================================
// Scenario Test Environment
// =============================================================================

// ScenarioEnv represents the complete test environment for scenario tests.
type ScenarioEnv struct {
	Repo       *testutil.TestRepo
	Server     *httptest.Server
	Router     *chi.Mux
	FileSvc    *service.FileService
	GitSvc     *service.GitService
	ArticleSvc *service.ArticleService
	AuthSvc    *service.AuthService
	AssetSvc   *service.AssetService
	Config     *config.Config
	Cleanup    func()
}

// SetupScenario creates a complete test environment.
func SetupScenario(t *testing.T, setup func(repo *testutil.TestRepo) error) *ScenarioEnv {
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

	// Create test auth config with multiple users
	hashedPass1, _ := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.DefaultCost)
	hashedPass2, _ := bcrypt.GenerateFromPassword([]byte("password2"), bcrypt.DefaultCost)

	cfg := &config.Config{
		Blog: config.BlogConfig{
			ContentDir: repo.Path,
		},
		Auth: config.AuthConfig{
			Users: []config.UserConfig{
				{Username: "admin", Password: string(hashedPass1)},
				{Username: "developer", Password: string(hashedPass2)},
			},
		},
	}
	authSvc := service.NewAuthService(cfg)

	// Create handlers
	articleHandler := handler.NewArticleHandler(articleSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)
	searchHandler := handler.NewSearchHandler(articleSvc)
	treeHandler := handler.NewTreeHandler(articleSvc)
	gitHandler := handler.NewGitHandler(gitSvc, authSvc)

	// Create router with all endpoints
	router := chi.NewRouter()
	router.Get("/api/articles", articleHandler.List)
	router.Get("/api/articles/*", articleHandler.HandleArticleRequest)
	router.Get("/api/tree", treeHandler.Get)
	router.Get("/api/search", searchHandler.Search)
	router.Get("/api/assets/*", assetHandler.Get)

	// Git Smart HTTP endpoints
	router.Get("/info/refs", gitHandler.InfoRefs)
	router.Post("/git-upload-pack", gitHandler.UploadPack)
	router.Post("/git-receive-pack", gitHandler.ReceivePack)

	server := httptest.NewServer(router)

	return &ScenarioEnv{
		Repo:       repo,
		Server:     server,
		Router:     router,
		FileSvc:    fileSvc,
		GitSvc:     gitSvc,
		ArticleSvc: articleSvc,
		AuthSvc:    authSvc,
		AssetSvc:   assetSvc,
		Config:     cfg,
		Cleanup: func() {
			server.Close()
			repo.Cleanup()
		},
	}
}

// =============================================================================
// Scenario 1: 文章列表获取（已提交Markdown文件）
// Requirements: 3.1.1, 3.2.4
// =============================================================================

func TestScenario01_ArticleList(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Setup: Create multiple markdown files with different timestamps
		now := time.Now()

		// Create oldest article
		if err := repo.CreateMarkdownFileWithTime("old.md", "# Old Article", "Add old", "author1", now.Add(-72*time.Hour)); err != nil {
			return err
		}

		// Create middle article
		if err := repo.CreateMarkdownFileWithTime("middle.md", "# Middle Article", "Add middle", "author2", now.Add(-48*time.Hour)); err != nil {
			return err
		}

		// Create newest article
		return repo.CreateMarkdownFileWithTime("new.md", "# New Article", "Add new", "author3", now.Add(-24*time.Hour))
	})
	defer env.Cleanup()

	// Test: Get article list with default sort (edited desc)
	resp, err := http.Get(env.Server.URL + "/api/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Response status and structure
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify: All committed articles are visible
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Articles, 3)

	// Verify: Each article has required metadata
	for _, article := range result.Articles {
		assert.NotEmpty(t, article.Path)
		assert.NotEmpty(t, article.Title)
		assert.NotEmpty(t, article.CreatedAt)
		assert.NotEmpty(t, article.CreatedBy)
		assert.NotEmpty(t, article.EditedAt)
		assert.NotEmpty(t, article.EditedBy)
		assert.NotEmpty(t, article.Contributors)
	}

	// Verify: Newest edited first (edited desc)
	// Note: created and edited times are the same since each file has only one commit
	assert.Equal(t, "new.md", result.Articles[0].Path, "Newest article should be first")
}

// =============================================================================
// Scenario 2: 文章内容获取（含元数据）
// Requirements: 3.1.2, 3.1.4
// =============================================================================

func TestScenario02_ArticleContent(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		return repo.CreateMarkdownFile("test-article.md", "# Test Article\n\n## Content\n\nThis is test content.\n\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```", "Initial commit", "creator")
	})
	defer env.Cleanup()

	// Test: Get article content
	resp, err := http.Get(env.Server.URL + "/api/articles/test-article.md")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Response status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var article model.ArticleResponse
	err = json.NewDecoder(resp.Body).Decode(&article)
	require.NoError(t, err)

	// Verify: Article metadata from Git history
	assert.Equal(t, "test-article.md", article.Path)
	assert.Equal(t, "test-article", article.Title)
	assert.Equal(t, "creator", article.CreatedBy)
	assert.Equal(t, "creator", article.EditedBy)
	assert.Contains(t, article.Contributors, "creator")

	// Verify: Content matches
	assert.Contains(t, article.Content, "# Test Article")
	assert.Contains(t, article.Content, "```go")
}

// =============================================================================
// Scenario 3: 文章时间线（Git提交历史）
// Requirements: 3.1.4, 3.3.3
// =============================================================================

func TestScenario03_ArticleTimeline(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		return repo.SetupMultiAuthorArticle("multi-author.md", "# Multi Author")
	})
	defer env.Cleanup()

	// Test: Get article timeline
	resp, err := http.Get(env.Server.URL + "/api/articles/multi-author.md/timeline")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Response status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var timeline model.TimelineResponse
	err = json.NewDecoder(resp.Body).Decode(&timeline)
	require.NoError(t, err)

	// Verify: Timeline has multiple commits
	assert.Len(t, timeline.Commits, 3)

	// Verify: Each commit has required fields
	for _, commit := range timeline.Commits {
		assert.Len(t, commit.Hash, 7) // Short hash format
		assert.NotEmpty(t, commit.Author)
		assert.NotEmpty(t, commit.Timestamp)
	}

	// Verify: Commits are sorted by time desc (most recent first)
	assert.Equal(t, "editor2", timeline.Commits[0].Author)
	assert.Equal(t, "editor1", timeline.Commits[1].Author)
	assert.Equal(t, "creator", timeline.Commits[2].Author)
}

// =============================================================================
// Scenario 4: 目录树获取
// Requirements: 3.1.1
// =============================================================================

func TestScenario04_DirectoryTree(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		return repo.SetupNestedDirectory()
	})
	defer env.Cleanup()

	// Test: Get directory tree
	resp, err := http.Get(env.Server.URL + "/api/tree")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Response status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.TreeResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify: Root node structure
	assert.Equal(t, model.NodeTypeDir, result.Root.Type)
	assert.NotEmpty(t, result.Root.Children)

	// Verify: Contains expected directories
	dirNames := make(map[string]bool)
	for _, child := range result.Root.Children {
		dirNames[child.Name] = true
	}
	assert.True(t, dirNames["tech"])
	assert.True(t, dirNames["life"])

	// Verify: tech directory has files
	for _, child := range result.Root.Children {
		if child.Name == "tech" {
			assert.Equal(t, model.NodeTypeDir, child.Type)
			assert.NotEmpty(t, child.Children)

			// Check files in tech directory
			fileNames := make(map[string]bool)
			for _, file := range child.Children {
				fileNames[file.Name] = true
			}
			assert.True(t, fileNames["golang.md"])
			assert.True(t, fileNames["rust.md"])
		}
	}
}

// =============================================================================
// Scenario 5: 标题搜索
// Requirements: 3.4
// =============================================================================

func TestScenario05_TitleSearch(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create articles with different titles
		if err := repo.CreateMarkdownFile("golang-intro.md", "# Golang Intro", "Add", "author"); err != nil {
			return err
		}
		if err := repo.CreateMarkdownFile("golang-advanced.md", "# Golang Advanced", "Add", "author"); err != nil {
			return err
		}
		if err := repo.CreateMarkdownFile("rust-guide.md", "# Rust Guide", "Add", "author"); err != nil {
			return err
		}
		return repo.CreateMarkdownFile("python-basics.md", "# Python Basics", "Add", "author")
	})
	defer env.Cleanup()

	// Test: Search for "golang"
	resp, err := http.Get(env.Server.URL + "/api/search?q=golang")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Response status
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result model.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify: Results contain matching articles
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Results, 2)

	// Verify: Results match the query
	for _, r := range result.Results {
		assert.Contains(t, r.Title, "golang")
		assert.Contains(t, r.Path, "golang")
	}

	// Test: Search for non-existent keyword
	resp2, err := http.Get(env.Server.URL + "/api/search?q=nonexistent")
	require.NoError(t, err)
	defer resp2.Body.Close()

	var result2 model.SearchResponse
	err = json.NewDecoder(resp2.Body).Decode(&result2)
	require.NoError(t, err)

	// Verify: No results for non-existent keyword
	assert.Equal(t, 0, result2.Total)
	assert.Len(t, result2.Results, 0)

	// Test: Search without query parameter
	resp3, err := http.Get(env.Server.URL + "/api/search")
	require.NoError(t, err)
	defer resp3.Body.Close()

	// Verify: Returns 400 for missing query
	assert.Equal(t, http.StatusBadRequest, resp3.StatusCode)
}

// =============================================================================
// Scenario 6: 未提交文件不显示
// Requirements: 3.2.4
// =============================================================================

func TestScenario06_UncommittedFilesNotVisible(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create committed file first
		if err := repo.CreateMarkdownFile("committed.md", "# Committed Article", "Add committed", "author"); err != nil {
			return err
		}

		// Create uncommitted files AFTER the commit (without committing them)
		// Note: These files are created in the working directory but not staged/committed
		if err := repo.CreateUncommittedFile("uncommitted.md", "# Uncommitted Article"); err != nil {
			return err
		}
		return repo.CreateUncommittedFile("draft.md", "# Draft Article")
	})
	defer env.Cleanup()

	// Test: Get article list - should only show committed files
	resp, err := http.Get(env.Server.URL + "/api/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify: Only committed file is visible
	assert.Equal(t, 1, result.Total)
	assert.Len(t, result.Articles, 1)
	assert.Equal(t, "committed.md", result.Articles[0].Path)

	// Test: Direct access to uncommitted file
	resp2, err := http.Get(env.Server.URL + "/api/articles/uncommitted.md")
	require.NoError(t, err)
	defer resp2.Body.Close()

	// Verify: Returns 400 for uncommitted file
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	// Test: Direct access to draft file
	resp3, err := http.Get(env.Server.URL + "/api/articles/draft.md")
	require.NoError(t, err)
	defer resp3.Body.Close()

	// Verify: Returns 400 for draft file
	assert.Equal(t, http.StatusBadRequest, resp3.StatusCode)

	// Test: Uncommitted file not in tree
	resp4, err := http.Get(env.Server.URL + "/api/tree")
	require.NoError(t, err)
	defer resp4.Body.Close()

	var treeResult model.TreeResponse
	err = json.NewDecoder(resp4.Body).Decode(&treeResult)
	require.NoError(t, err)

	// Verify: Tree only shows committed file
	fileCount := 0
	for _, child := range treeResult.Root.Children {
		if child.Type == model.NodeTypeFile {
			fileCount++
			assert.Equal(t, "committed.md", child.Name)
		}
	}
	assert.Equal(t, 1, fileCount)

	// Test: Search should not find uncommitted files
	resp5, err := http.Get(env.Server.URL + "/api/search?q=Uncommitted")
	require.NoError(t, err)
	defer resp5.Body.Close()

	var searchResult model.SearchResponse
	err = json.NewDecoder(resp5.Body).Decode(&searchResult)
	require.NoError(t, err)

	// Verify: Search doesn't find uncommitted files
	assert.Equal(t, 0, searchResult.Total)
}

// =============================================================================
// Scenario 7: 排序选项
// Requirements: 3.3.2, 3.3.3
// =============================================================================

func TestScenario07_SortOptions(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		now := time.Now()

		// Article 1: Created oldest, edited recently
		if err := repo.CreateMarkdownFileWithTime("alpha.md", "# Alpha", "Create alpha", "author1", now.Add(-72*time.Hour)); err != nil {
			return err
		}

		// Article 2: Created middle
		if err := repo.CreateMarkdownFileWithTime("beta.md", "# Beta", "Create beta", "author2", now.Add(-48*time.Hour)); err != nil {
			return err
		}

		// Article 3: Created newest
		return repo.CreateMarkdownFileWithTime("gamma.md", "# Gamma", "Create gamma", "author3", now.Add(-24*time.Hour))
	})
	defer env.Cleanup()

	tests := []struct {
		name     string
		params   string
		expected []string // Expected order of paths
	}{
		{
			name:     "Default sort (edited desc)",
			params:   "",
			expected: []string{"gamma.md", "beta.md", "alpha.md"}, // Newest edited first
		},
		{
			name:     "Sort by created asc",
			params:   "?sort=created&order=asc",
			expected: []string{"alpha.md", "beta.md", "gamma.md"}, // Oldest created first
		},
		{
			name:     "Sort by created desc",
			params:   "?sort=created&order=desc",
			expected: []string{"gamma.md", "beta.md", "alpha.md"}, // Newest created first
		},
		{
			name:     "Sort by edited asc",
			params:   "?sort=edited&order=asc",
			expected: []string{"alpha.md", "beta.md", "gamma.md"}, // Oldest edited first
		},
		{
			name:     "Sort by edited desc",
			params:   "?sort=edited&order=desc",
			expected: []string{"gamma.md", "beta.md", "alpha.md"}, // Newest edited first
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

			paths := make([]string, len(result.Articles))
			for i, a := range result.Articles {
				paths[i] = a.Path
			}

			assert.Equal(t, tt.expected, paths, "Sort order mismatch for %s", tt.name)
		})
	}
}

// =============================================================================
// Scenario 8: 多作者贡献者
// Requirements: 3.1.4
// =============================================================================

func TestScenario08_MultipleContributors(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		return repo.SetupMultiAuthorArticle("collaborative.md", "# Collaborative Article")
	})
	defer env.Cleanup()

	// Test: Get article content
	resp, err := http.Get(env.Server.URL + "/api/articles/collaborative.md")
	require.NoError(t, err)
	defer resp.Body.Close()

	var article model.ArticleResponse
	err = json.NewDecoder(resp.Body).Decode(&article)
	require.NoError(t, err)

	// Verify: Creator is the first commit author
	assert.Equal(t, "creator", article.CreatedBy)

	// Verify: Last editor is the most recent commit author
	assert.Equal(t, "editor2", article.EditedBy)

	// Verify: All contributors are listed
	assert.Len(t, article.Contributors, 3)
	assert.Contains(t, article.Contributors, "creator")
	assert.Contains(t, article.Contributors, "editor1")
	assert.Contains(t, article.Contributors, "editor2")

	// Test: Get timeline to verify commit order
	resp2, err := http.Get(env.Server.URL + "/api/articles/collaborative.md/timeline")
	require.NoError(t, err)
	defer resp2.Body.Close()

	var timeline model.TimelineResponse
	err = json.NewDecoder(resp2.Body).Decode(&timeline)
	require.NoError(t, err)

	// Verify: Timeline shows all authors
	authors := make(map[string]bool)
	for _, commit := range timeline.Commits {
		authors[commit.Author] = true
	}
	assert.Len(t, authors, 3)
	assert.True(t, authors["creator"])
	assert.True(t, authors["editor1"])
	assert.True(t, authors["editor2"])
}

// =============================================================================
// Scenario 9: Git Clone (HTTP无认证)
// Requirements: 3.2.1, 3.2.2
// =============================================================================

func TestScenario09_GitClone(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create some content to clone
		if err := repo.CreateMarkdownFile("article1.md", "# Article 1", "Add article1", "author"); err != nil {
			return err
		}
		return repo.CreateMarkdownFile("article2.md", "# Article 2", "Add article2", "author")
	})
	defer env.Cleanup()

	// Test: Get upload-pack refs (Clone info)
	resp, err := http.Get(env.Server.URL + "/info/refs?service=git-upload-pack")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Response status (Clone is public, no auth required)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify: Content-Type header
	assert.Contains(t, resp.Header.Get("Content-Type"), "git-upload-pack-advertisement")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Verify: Contains service announcement
	assert.Contains(t, string(body), "# service=git-upload-pack")

	// Verify: Contains HEAD reference
	assert.Contains(t, string(body), "HEAD")

	// Test: Clone without authentication should work
	// Note: Full clone test requires git CLI, we verify the refs endpoint here
}

// =============================================================================
// Scenario 10: Git Push认证
// Requirements: 3.2.2, 3.2.3
// =============================================================================

func TestScenario10_GitPushAuthentication(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Initialize with some content
		return repo.CreateMarkdownFile("initial.md", "# Initial", "Initial commit", "admin")
	})
	defer env.Cleanup()

	// Test: Get receive-pack refs without auth
	resp, err := http.Get(env.Server.URL + "/info/refs?service=git-receive-pack")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify: Requires authentication (401)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Verify: WWW-Authenticate header
	assert.Contains(t, resp.Header.Get("WWW-Authenticate"), "Basic")

	// Test: Get receive-pack refs with valid auth
	req, err := http.NewRequest("GET", env.Server.URL+"/info/refs?service=git-receive-pack", nil)
	require.NoError(t, err)

	// Add Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte("admin:password1"))
	req.Header.Set("Authorization", "Basic "+auth)

	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp2.Body.Close()

	// Verify: Returns 200 with valid auth
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Verify: Content-Type header
	assert.Contains(t, resp2.Header.Get("Content-Type"), "git-receive-pack-advertisement")

	// Test: Get receive-pack refs with invalid auth
	req2, err := http.NewRequest("GET", env.Server.URL+"/info/refs?service=git-receive-pack", nil)
	require.NoError(t, err)

	auth2 := base64.StdEncoding.EncodeToString([]byte("admin:wrongpassword"))
	req2.Header.Set("Authorization", "Basic "+auth2)

	resp3, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp3.Body.Close()

	// Verify: Returns 401 with invalid auth
	assert.Equal(t, http.StatusUnauthorized, resp3.StatusCode)

	// Test: Unknown user cannot push
	req3, err := http.NewRequest("GET", env.Server.URL+"/info/refs?service=git-receive-pack", nil)
	require.NoError(t, err)

	auth3 := base64.StdEncoding.EncodeToString([]byte("unknown:anything"))
	req3.Header.Set("Authorization", "Basic "+auth3)

	resp4, err := http.DefaultClient.Do(req3)
	require.NoError(t, err)
	defer resp4.Body.Close()

	// Verify: Returns 401 for unknown user
	assert.Equal(t, http.StatusUnauthorized, resp4.StatusCode)
}

// =============================================================================
// Scenario 11: 安全测试（路径遍历/.git目录）
// Requirements: 3.2.4, Security
// =============================================================================

func TestScenario11_SecurityTests(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create normal content
		if err := repo.CreateMarkdownFile("normal.md", "# Normal Article", "Add", "author"); err != nil {
			return err
		}
		// Create image in images directory
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
		return repo.CreateImageFileAndCommit("images/photo.png", pngData, "Add image", "author")
	})
	defer env.Cleanup()

	tests := []struct {
		name        string
		path        string
		wantStatus  int
		description string
	}{
		{
			name:        "Path traversal with ../",
			path:        "/api/articles/../normal.md",
			wantStatus:  http.StatusBadRequest, // Blocked by ValidatePath for invalid characters
			description: "Path traversal is blocked by path validation",
		},
		{
			name:        "Path traversal to parent directory",
			path:        "/api/articles/../../normal.md",
			wantStatus:  http.StatusBadRequest, // Blocked by ValidatePath
			description: "Multiple parent directory traversal is blocked",
		},
		{
			name:        "Access .git directory",
			path:        "/api/articles/.git/config",
			wantStatus:  http.StatusBadRequest, // Explicit .git access blocked by ValidatePath
			description: ".git directory access should be blocked",
		},
		{
			name:        "Access .git via asset endpoint",
			path:        "/api/assets/.git/config",
			wantStatus:  http.StatusBadRequest, // .git access blocked by ValidatePath
			description: ".git directory via asset endpoint should be blocked",
		},
		{
			name:        "Asset path traversal",
			path:        "/api/assets/../images/photo.png",
			wantStatus:  http.StatusBadRequest, // Blocked by ValidatePath
			description: "Asset path traversal is blocked by path validation",
		},
		{
			name:        "Absolute path attempt",
			path:        "/api/articles/etc/passwd",
			wantStatus:  http.StatusBadRequest, // Invalid path (not in repo)
			description: "Absolute path access should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(env.Server.URL + tt.path)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode, tt.description)
		})
	}

	// Additional test: Verify legitimate nested path works
	resp, err := http.Get(env.Server.URL + "/api/assets/images/photo.png")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Legitimate nested path should work")
	assert.Contains(t, resp.Header.Get("Content-Type"), "image/png")

	// Additional test: Verify legitimate .git look-alike is blocked
	resp2, err := http.Get(env.Server.URL + "/api/assets/images/.git-backup/config")
	require.NoError(t, err)
	defer resp2.Body.Close()

	// Path containing ".git" substring should be blocked
	assert.NotEqual(t, http.StatusOK, resp2.StatusCode, "Paths containing .git should be blocked")
}

// =============================================================================
// Scenario 12: 图片资源访问
// Requirements: 3.1.3, 3.1.3 (图片处理规则)
// =============================================================================

func TestScenario12_ImageAssets(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create PNG image
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
		if err := repo.CreateImageFileAndCommit("images/test.png", pngData, "Add PNG", "author"); err != nil {
			return err
		}

		// Create JPEG image (minimal valid JPEG header)
		jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		if err := repo.CreateImageFileAndCommit("images/test.jpg", jpegData, "Add JPEG", "author"); err != nil {
			return err
		}

		// Create SVG
		svgData := []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"><circle/></svg>")
		return repo.CreateImageFileAndCommit("images/test.svg", svgData, "Add SVG", "author")
	})
	defer env.Cleanup()

	tests := []struct {
		name            string
		path            string
		wantStatus      int
		wantContentType string
	}{
		{
			name:            "PNG image",
			path:            "/api/assets/images/test.png",
			wantStatus:      http.StatusOK,
			wantContentType: "image/png",
		},
		{
			name:            "JPEG image",
			path:            "/api/assets/images/test.jpg",
			wantStatus:      http.StatusOK,
			wantContentType: "image/jpeg",
		},
		{
			name:            "SVG image",
			path:            "/api/assets/images/test.svg",
			wantStatus:      http.StatusOK,
			wantContentType: "image/svg+xml",
		},
		{
			name:            "Non-existent image",
			path:            "/api/assets/images/not-exist.png",
			wantStatus:      http.StatusNotFound,
			wantContentType: "",
		},
		{
			name:            "Directory as image",
			path:            "/api/assets/images",
			wantStatus:      http.StatusNotFound,
			wantContentType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(env.Server.URL + tt.path)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				assert.Contains(t, contentType, tt.wantContentType)

				// Verify body is not empty
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.NotEmpty(t, body)
			}
		})
	}
}

// =============================================================================
// Scenario 13: 嵌套目录文章
// Requirements: 3.1.1
// =============================================================================

func TestScenario13_NestedDirectoryArticles(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create nested directory structure
		if err := repo.CreateFile("tech/golang/intro.md", "# Golang Intro"); err != nil {
			return err
		}
		if err := repo.CreateFile("tech/golang/advanced.md", "# Golang Advanced"); err != nil {
			return err
		}
		if err := repo.CreateFile("tech/rust/basics.md", "# Rust Basics"); err != nil {
			return err
		}
		if err := repo.CreateFile("life/travel/japan.md", "# Japan Travel"); err != nil {
			return err
		}
		return repo.Commit("Add nested articles", "author", "author@example.com")
	})
	defer env.Cleanup()

	// Test: Get root articles list
	resp, err := http.Get(env.Server.URL + "/api/articles")
	require.NoError(t, err)
	defer resp.Body.Close()

	var result model.ArticleListResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify: All nested articles are visible
	assert.Equal(t, 4, result.Total)

	// Verify: Paths include nested directories
	paths := make([]string, len(result.Articles))
	for i, a := range result.Articles {
		paths[i] = a.Path
	}
	assert.Contains(t, paths, "tech/golang/intro.md")
	assert.Contains(t, paths, "tech/golang/advanced.md")
	assert.Contains(t, paths, "tech/rust/basics.md")
	assert.Contains(t, paths, "life/travel/japan.md")

	// Test: Get specific nested article
	resp2, err := http.Get(env.Server.URL + "/api/articles/tech/golang/intro.md")
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Test: Get directory-specific articles
	resp3, err := http.Get(env.Server.URL + "/api/articles?dir=tech/golang")
	require.NoError(t, err)
	defer resp3.Body.Close()

	var dirResult model.ArticleListResponse
	err = json.NewDecoder(resp3.Body).Decode(&dirResult)
	require.NoError(t, err)

	// Verify: Only articles in tech/golang directory
	assert.Equal(t, 2, dirResult.Total)
	for _, a := range dirResult.Articles {
		assert.Contains(t, a.Path, "tech/golang/")
	}
}

// =============================================================================
// Scenario 14: 文件删除后重新添加
// Requirements: 3.2.4
// =============================================================================

func TestScenario14_DeleteAndRecreate(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		now := time.Now()

		// Create initial article
		if err := repo.CreateMarkdownFileWithTime("test.md", "# Original Content", "Create original", "author1", now.Add(-48*time.Hour)); err != nil {
			return err
		}

		// Get the repo to delete the file
		wt, err := repo.Repo.Worktree()
		if err != nil {
			return err
		}

		// Remove the file
		_, err = wt.Remove("test.md")
		if err != nil {
			return err
		}

		// Commit deletion
		_, err = wt.Commit("Delete test.md", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "author1",
				Email: "author1@example.com",
				When:  now.Add(-24 * time.Hour),
			},
		})
		if err != nil {
			return err
		}

		// Recreate the file with same name
		if err := repo.CreateMarkdownFileWithTime("test.md", "# New Content", "Recreate", "author2", now); err != nil {
			return err
		}

		return nil
	})
	defer env.Cleanup()

	// Test: Get article content
	resp, err := http.Get(env.Server.URL + "/api/articles/test.md")
	require.NoError(t, err)
	defer resp.Body.Close()

	var article model.ArticleResponse
	err = json.NewDecoder(resp.Body).Decode(&article)
	require.NoError(t, err)

	// Verify: Git tracks complete history including deletion and recreation
	// The first commit creates the file, so CreatedBy is the original author
	assert.Equal(t, "author1", article.CreatedBy)
	assert.Equal(t, "author2", article.EditedBy)

	// Verify: Contributors includes both authors (Git tracks full history)
	assert.Len(t, article.Contributors, 2)
	assert.Contains(t, article.Contributors, "author1")
	assert.Contains(t, article.Contributors, "author2")

	// Test: Get timeline - should have all commits (create + delete + recreate)
	resp2, err := http.Get(env.Server.URL + "/api/articles/test.md/timeline")
	require.NoError(t, err)
	defer resp2.Body.Close()

	var timeline model.TimelineResponse
	err = json.NewDecoder(resp2.Body).Decode(&timeline)
	require.NoError(t, err)

	// Verify: Git tracks all file lifecycle events
	assert.Len(t, timeline.Commits, 3)
	// Newest commit is the recreation
	assert.Equal(t, "author2", timeline.Commits[0].Author)
	assert.Equal(t, "Recreate", timeline.Commits[0].Message)
	// Middle commit is the deletion
	assert.Equal(t, "author1", timeline.Commits[1].Author)
	assert.Equal(t, "Delete test.md", timeline.Commits[1].Message)
	// Oldest commit is the original creation
	assert.Equal(t, "author1", timeline.Commits[2].Author)
	assert.Equal(t, "Create original", timeline.Commits[2].Message)
}

// =============================================================================
// Scenario 15: 错误响应处理
// Requirements: API Spec 2.2
// =============================================================================

func TestScenario15_ErrorHandling(t *testing.T) {
	env := SetupScenario(t, func(repo *testutil.TestRepo) error {
		// Create one article to initialize repo
		return repo.CreateMarkdownFile("exists.md", "# Exists", "Add", "author")
	})
	defer env.Cleanup()

	tests := []struct {
		name       string
		endpoint   string
		wantStatus int
		wantError  string
	}{
		{
			name:       "Article not found",
			endpoint:   "/api/articles/not-exist.md",
			wantStatus: http.StatusBadRequest,
			wantError:  "",
		},
		{
			name:       "Asset not found",
			endpoint:   "/api/assets/not-exist.png",
			wantStatus: http.StatusNotFound,
			wantError:  "",
		},
		{
			name:       "Search without query",
			endpoint:   "/api/search",
			wantStatus: http.StatusBadRequest,
			wantError:  "",
		},
		{
			name:       "Invalid sort parameter",
			endpoint:   "/api/articles?sort=invalid",
			wantStatus: http.StatusOK, // Invalid sort params are ignored, returns default
			wantError:  "",
		},
		{
			name:       "Git push without auth",
			endpoint:   "/info/refs?service=git-receive-pack",
			wantStatus: http.StatusUnauthorized,
			wantError:  "Authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(env.Server.URL + tt.endpoint)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode, tt.name)

			// Read body for error messages
			body, _ := io.ReadAll(resp.Body)

			if tt.wantError != "" {
				var errorResp model.ErrorResponse
				json.Unmarshal(body, &errorResp)
				assert.Contains(t, errorResp.Error, tt.wantError)
			}
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// MakeRequest creates an HTTP request with optional headers
func (env *ScenarioEnv) MakeRequest(method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, env.Server.URL+path, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return http.DefaultClient.Do(req)
}

// MakeAuthRequest creates an HTTP request with Basic Auth
func (env *ScenarioEnv) MakeAuthRequest(method, path string, body io.Reader, username, password string) (*http.Response, error) {
	req, err := http.NewRequest(method, env.Server.URL+path, body)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	req.Header.Set("Authorization", "Basic "+auth)

	return http.DefaultClient.Do(req)
}
