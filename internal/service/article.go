// Package service provides business logic services for the application.
package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// ArticleService provides article-related operations.
type ArticleService struct {
	fileSvc *FileService
	gitSvc  *GitService
}

// NewArticleService creates a new ArticleService instance.
func NewArticleService(fileSvc *FileService, gitSvc *GitService) *ArticleService {
	return &ArticleService{
		fileSvc: fileSvc,
		gitSvc:  gitSvc,
	}
}

// ListOptions contains options for listing articles.
type ListOptions struct {
	// Dir is the directory path to scan.
	Dir string

	// Sort is the field to sort by.
	Sort model.SortField

	// Order is the sort direction.
	Order model.SortOrder
}

// ListArticles returns a list of articles in the given directory.
// Only committed Markdown files are included.
func (s *ArticleService) ListArticles(ctx context.Context, opts ListOptions) ([]model.Article, error) {
	// 1. Scan Markdown files
	files, err := s.fileSvc.ScanMarkdownFiles(ctx, opts.Dir)
	if err != nil {
		return nil, err
	}

	// 2. Filter and get history
	articles := make([]model.Article, 0, len(files))

	for _, file := range files {
		// Check if committed
		committed, err := s.gitSvc.IsFileCommitted(ctx, file)
		if err != nil {
			continue // Skip on error
		}
		if !committed {
			continue // Skip uncommitted files
		}

		// Get history
		history, err := s.gitSvc.GetFileHistory(ctx, file)
		if err != nil {
			continue // Skip on error
		}

		// Build Article
		article := model.Article{
			Path:         file,
			Title:        utils.ExtractTitle(file),
			CreatedAt:    history.FirstCommit.Timestamp,
			CreatedBy:    history.FirstCommit.Author,
			EditedAt:     history.LastCommit.Timestamp,
			EditedBy:     history.LastCommit.Author,
			Contributors: history.Contributors,
		}

		articles = append(articles, article)
	}

	// 3. Sort
	sortArticles(articles, opts.Sort, opts.Order)

	return articles, nil
}

// GetArticle returns the content and metadata of a specific article.
func (s *ArticleService) GetArticle(ctx context.Context, path string) (*model.ArticleDetail, error) {
	// Normalize path
	path = utils.NormalizePath(path)

	// Validate path
	if _, err := s.fileSvc.ValidatePath(path); err != nil {
		return nil, model.ErrInvalidPath
	}

	// Check if committed
	committed, err := s.gitSvc.IsFileCommitted(ctx, path)
	if err != nil {
		return nil, err
	}
	if !committed {
		return nil, model.ErrNotCommitted
	}

	// Read content
	content, err := s.fileSvc.ReadFile(ctx, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	// Get history
	history, err := s.gitSvc.GetFileHistory(ctx, path)
	if err != nil {
		return nil, err
	}

	return &model.ArticleDetail{
		Article: model.Article{
			Path:         path,
			Title:        utils.ExtractTitle(path),
			CreatedAt:    history.FirstCommit.Timestamp,
			CreatedBy:    history.FirstCommit.Author,
			EditedAt:     history.LastCommit.Timestamp,
			EditedBy:     history.LastCommit.Author,
			Contributors: history.Contributors,
		},
		Content: string(content),
	}, nil
}

// GetTimeline returns the commit timeline for an article.
func (s *ArticleService) GetTimeline(ctx context.Context, path string) ([]model.CommitInfo, error) {
	path = utils.NormalizePath(path)

	history, err := s.gitSvc.GetFileHistory(ctx, path)
	if err != nil {
		return nil, err
	}

	return history.AllCommits, nil
}

// GetTree returns the directory tree structure.
func (s *ArticleService) GetTree(ctx context.Context, dir string) (*model.TreeNode, error) {
	dir = utils.NormalizePath(dir)
	return s.fileSvc.GetDirectoryTree(ctx, dir)
}

// Search searches articles by title.
func (s *ArticleService) Search(ctx context.Context, query string, dir string) ([]model.SearchResult, error) {
	// Normalize inputs
	query = strings.ToLower(query)
	dir = utils.NormalizePath(dir)

	// Get article list
	articles, err := s.ListArticles(ctx, ListOptions{Dir: dir})
	if err != nil {
		return nil, err
	}

	// Search titles
	results := make([]model.SearchResult, 0)

	for _, article := range articles {
		titleLower := strings.ToLower(article.Title)
		if strings.Contains(titleLower, query) {
			results = append(results, model.SearchResult{
				Path:         article.Path,
				Title:        article.Title,
				MatchedTitle: article.Title,
			})
		}
	}

	return results, nil
}

// Helper function: sortArticles sorts articles by the given field and order.
func sortArticles(articles []model.Article, sortField model.SortField, order model.SortOrder) {
	sort.Slice(articles, func(i, j int) bool {
		var less bool

		switch sortField {
		case model.SortCreated:
			less = articles[i].CreatedAt.Before(articles[j].CreatedAt)
		case model.SortEdited:
			less = articles[i].EditedAt.Before(articles[j].EditedAt)
		default:
			less = articles[i].EditedAt.Before(articles[j].EditedAt)
		}

		if order == model.OrderDesc {
			return !less
		}
		return less
	})
}
