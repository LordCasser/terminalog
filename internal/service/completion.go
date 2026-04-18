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

// CompletionService provides path completion functionality.
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
// When dir is empty (global search), it matches against all levels of path names.
// When dir is specified, it only matches against names relative to that directory.
func (s *CompletionService) GetMatchingItems(ctx context.Context, dir, prefix string) ([]string, error) {
	// Normalize inputs
	dir = utils.NormalizePath(dir)
	prefix = strings.ToLower(prefix)

	// Get list of committed articles (recursive from dir)
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

	// Track items we've already added (to avoid duplicates)
	itemSet := make(map[string]bool)

	// If dir is empty (global search), match against all levels of path names
	if dir == "" || dir == "/" {
		for _, article := range articles {
			// Split the path into components
			parts := strings.Split(article.Path, "/")

			// Check each directory component (excluding the last which is filename)
			for i := 0; i < len(parts)-1; i++ {
				if parts[i] == "" {
					continue
				}
				// Check if this directory name matches prefix
				if strings.HasPrefix(strings.ToLower(parts[i]), prefix) {
					// Build the relative path up to this directory
					dirPath := strings.Join(parts[:i+1], "/") + "/"
					if !itemSet[dirPath] {
						items = append(items, dirPath)
						itemSet[dirPath] = true
					}
				}
			}

			// Check the filename
			filename := filepath.Base(article.Path)
			if strings.HasPrefix(strings.ToLower(filename), prefix) {
				// Return full path for file
				filePath := article.Path
				if !itemSet[filePath] {
					items = append(items, filePath)
					itemSet[filePath] = true
				}
			}
		}
	} else {
		// When dir is specified, match against names relative to that directory
		for _, article := range articles {
			// Get the name (basename)
			name := filepath.Base(article.Path)

			// Check if name matches prefix
			if strings.HasPrefix(strings.ToLower(name), prefix) {
				// Add file without trailing slash
				if !itemSet[name] {
					items = append(items, name)
					itemSet[name] = true
				}
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
					if strings.HasPrefix(strings.ToLower(subDirName), prefix) && !itemSet[subDirName+"/"] {
						// Add directory with trailing slash
						items = append(items, subDirName+"/")
						itemSet[subDirName+"/"] = true
					}
				}
			}
		}
	}

	return items, nil
}
