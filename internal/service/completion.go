// Package service provides business logic services for the application.
package service

import (
	"context"
	"path/filepath"
	"strings"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// CompletionRequest represents a path completion request.
type CompletionRequest struct {
	// Type is the message type ("completion_request").
	Type string `json:"type"`

	// Dir is the current directory path (e.g., "/" or "/tech").
	Dir string `json:"dir"`

	// Prefix is the path prefix to match (e.g., "RE" or "tec").
	Prefix string `json:"prefix"`
}

// CompletionResponse represents a path completion response.
type CompletionResponse struct {
	// Type is the message type ("completion_response").
	Type string `json:"type"`

	// Items is the list of completion results.
	Items []string `json:"items"`
}

// SearchRequest represents a search request via WebSocket.
type SearchRequest struct {
	// Type is the message type ("search_request").
	Type string `json:"type"`

	// Keyword is the search keyword.
	Keyword string `json:"keyword"`
}

// SearchResponse represents a search response via WebSocket.
type SearchResponse struct {
	// Type is the message type ("search_response").
	Type string `json:"type"`

	// Results is the list of search results.
	Results []WebSocketSearchResult `json:"results"`
}

// WebSocketSearchResult represents a search result for WebSocket.
type WebSocketSearchResult struct {
	// Path is the result path (e.g., "tech/golang" for dirs, "tech/golang/go-guide.md" for files).
	Path string `json:"path"`

	// Title is the display title (article title for files, directory name for dirs).
	Title string `json:"title"`

	// Type is the result type: "file" for articles, "dir" for directories.
	Type string `json:"type"`
}

// CompletionService provides path completion and search functionality.
type CompletionService struct {
	articleSvc *ArticleService
	fileSvc    *FileService
	gitSvc     *GitService
}

// NewCompletionService creates a new CompletionService instance.
func NewCompletionService(articleSvc *ArticleService, fileSvc *FileService, gitSvc *GitService) *CompletionService {
	return &CompletionService{
		articleSvc: articleSvc,
		fileSvc:    fileSvc,
		gitSvc:     gitSvc,
	}
}

// HandleCompletion handles a path completion request.
// It returns matching files and directories in the given directory.
// Files are returned without a trailing slash, directories with a trailing slash.
// Special files (starting with "_") are filtered out.
func (s *CompletionService) HandleCompletion(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Normalize directory path
	dir := utils.NormalizePath(req.Dir)
	if dir == "/" {
		dir = ""
	}

	// Get matching items
	items, err := s.GetMatchingItems(ctx, dir, req.Prefix)
	if err != nil {
		return nil, err
	}

	return &CompletionResponse{
		Type:  "completion_response",
		Items: items,
	}, nil
}

// GetMatchingItems returns matching files and directories based on prefix.
// Files are returned without slash, directories with trailing slash.
// Only committed Markdown files are included.
// Special files (starting with "_") are filtered out.
func (s *CompletionService) GetMatchingItems(ctx context.Context, dir, prefix string) ([]string, error) {
	// Normalize inputs
	dir = utils.NormalizePath(dir)
	prefix = strings.ToLower(prefix)

	// Get list of committed articles
	articles, err := s.articleSvc.ListArticles(ctx, ListOptions{
		Dir:      dir,
		Sort:     model.SortEdited,
		Order:    model.OrderDesc,
		UseCache: true,
		Parallel: false,
	})
	if err != nil {
		return nil, err
	}

	// Collect matching items
	items := make([]string, 0)

	// Track directories we've already added (to avoid duplicates)
	dirSet := make(map[string]bool)

	// Process articles
	for _, article := range articles {
		// Get the name (basename)
		name := filepath.Base(article.Path)

		// Check if name matches prefix
		if strings.HasPrefix(strings.ToLower(name), prefix) {
			// Add file without trailing slash
			items = append(items, name)
		}

		// Handle directories - extract parent directories from paths
		articleDir := filepath.Dir(article.Path)
		articleDir = utils.NormalizePath(articleDir)

		// Get relative directory from the requested dir
		relDir := strings.TrimPrefix(articleDir, dir)
		relDir = strings.TrimPrefix(relDir, "/")

		// If there's a subdirectory that matches the prefix
		if relDir != "" && relDir != "." {
			// Get the first component of the relative path
			parts := strings.Split(relDir, "/")
			if len(parts) > 0 && parts[0] != "" {
				subDirName := parts[0]
				if strings.HasPrefix(strings.ToLower(subDirName), prefix) && !dirSet[subDirName] {
					// Add directory with trailing slash
					items = append(items, subDirName+"/")
					dirSet[subDirName] = true
				}
			}
		}
	}

	// If dir is empty (root), also check for top-level directories
	if dir == "" || dir == "/" {
		// Get all top-level directories from article paths
		for _, article := range articles {
			parts := strings.Split(article.Path, "/")
			if len(parts) > 1 && parts[0] != "" {
				topDir := parts[0]
				if strings.HasPrefix(strings.ToLower(topDir), prefix) && !dirSet[topDir] {
					items = append(items, topDir+"/")
					dirSet[topDir] = true
				}
			}
		}
	}

	return items, nil
}

// HandleSearch handles a search request via WebSocket.
// It searches article titles and returns matching results.
// Special files (starting with "_") are filtered out.
// Maximum of 10 results are returned.
func (s *CompletionService) HandleSearch(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	// Search using the article service
	results, err := s.articleSvc.Search(ctx, req.Keyword, "")
	if err != nil {
		return nil, err
	}

	// Convert to WebSocket search result format
	wsResults := make([]WebSocketSearchResult, 0, len(results))

	// Limit to 10 results
	maxResults := 10
	if len(results) < maxResults {
		maxResults = len(results)
	}

	for i := 0; i < maxResults; i++ {
		wsResults = append(wsResults, WebSocketSearchResult{
			Path:  results[i].Path,
			Title: results[i].Title,
			Type:  string(results[i].Type),
		})
	}

	return &SearchResponse{
		Type:    "search_response",
		Results: wsResults,
	}, nil
}
