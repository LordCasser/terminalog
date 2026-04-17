// Package service provides business logic services for the application.
package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// ArticleService provides article-related operations.
type ArticleService struct {
	fileSvc *FileService
	gitSvc  *GitService
	cache   *ArticleCache
}

// NewArticleService creates a new ArticleService instance.
func NewArticleService(fileSvc *FileService, gitSvc *GitService) *ArticleService {
	return &ArticleService{
		fileSvc: fileSvc,
		gitSvc:  gitSvc,
		cache:   NewArticleCache(DefaultCacheTTL),
	}
}

// NewArticleServiceWithCache creates a new ArticleService with custom cache.
func NewArticleServiceWithCache(fileSvc *FileService, gitSvc *GitService, cache *ArticleCache) *ArticleService {
	return &ArticleService{
		fileSvc: fileSvc,
		gitSvc:  gitSvc,
		cache:   cache,
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

	// UseCache indicates whether to use cached results.
	UseCache bool

	// Parallel indicates whether to use parallel processing.
	Parallel bool
}

// ListArticles returns a list of articles in the given directory.
// Only committed Markdown files are included.
func (s *ArticleService) ListArticles(ctx context.Context, opts ListOptions) ([]model.Article, error) {
	// Check cache first
	cacheKey := opts.Dir + ":" + string(opts.Sort) + ":" + string(opts.Order)
	if opts.UseCache {
		if cached, ok := s.cache.GetArticleList(cacheKey); ok {
			return cached, nil
		}
	}

	// 1. Scan Markdown files
	files, err := s.fileSvc.ScanMarkdownFiles(ctx, opts.Dir)
	if err != nil {
		return nil, err
	}

	// 2. Process files (parallel or sequential)
	articles := make([]model.Article, 0, len(files))

	if opts.Parallel && len(files) > 10 {
		// Use parallel processing for large file lists
		articles = s.listArticlesParallel(ctx, files)
	} else {
		// Use sequential processing for small file lists
		articles = s.listArticlesSequential(ctx, files)
	}

	// 3. Sort
	sortArticles(articles, opts.Sort, opts.Order)

	// 4. Cache result
	s.cache.SetArticleList(cacheKey, articles)

	return articles, nil
}

// listArticlesSequential processes files sequentially.
func (s *ArticleService) listArticlesSequential(ctx context.Context, files []string) []model.Article {
	articles := make([]model.Article, 0, len(files))

	for _, file := range files {
		article, err := s.processFile(ctx, file)
		if err != nil {
			continue // Skip on error
		}
		articles = append(articles, article)
	}

	return articles
}

// listArticlesParallel processes files in parallel using goroutines.
func (s *ArticleService) listArticlesParallel(ctx context.Context, files []string) []model.Article {
	// Result channel
	resultCh := make(chan articleResult, len(files))

	// Worker pool (limit concurrent goroutines)
	const maxWorkers = 10
	semaphore := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)

		go func(f string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Process file
			article, err := s.processFile(ctx, f)
			resultCh <- articleResult{article: article, err: err}
		}(file)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	articles := make([]model.Article, 0, len(files))
	for result := range resultCh {
		if result.err == nil {
			articles = append(articles, result.article)
		}
	}

	return articles
}

// articleResult represents the result of processing a file.
type articleResult struct {
	article model.Article
	err     error
}

// processFile processes a single file and returns its article metadata.
func (s *ArticleService) processFile(ctx context.Context, file string) (model.Article, error) {
	// Check cache for this file
	if cached, ok := s.cache.GetArticle(file); ok {
		return *cached, nil
	}

	// Check if committed
	committed, err := s.gitSvc.IsFileCommitted(ctx, file)
	if err != nil {
		return model.Article{}, err
	}
	if !committed {
		return model.Article{}, model.ErrNotCommitted
	}

	// Get history
	history, err := s.gitSvc.GetFileHistory(ctx, file)
	if err != nil {
		return model.Article{}, err
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

	// Cache article
	s.cache.SetArticle(file, &article)

	return article, nil
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

	// Check cache for metadata
	var article model.Article
	if cached, ok := s.cache.GetArticle(path); ok {
		article = *cached
	} else {
		// Get history
		history, err := s.gitSvc.GetFileHistory(ctx, path)
		if err != nil {
			return nil, err
		}

		article = model.Article{
			Path:         path,
			Title:        utils.ExtractTitle(path),
			CreatedAt:    history.FirstCommit.Timestamp,
			CreatedBy:    history.FirstCommit.Author,
			EditedAt:     history.LastCommit.Timestamp,
			EditedBy:     history.LastCommit.Author,
			Contributors: history.Contributors,
		}

		// Cache
		s.cache.SetArticle(path, &article)
	}

	return &model.ArticleDetail{
		Article: article,
		Content: string(content),
	}, nil
}

// GetTimeline returns the commit timeline for an article.
func (s *ArticleService) GetTimeline(ctx context.Context, path string) ([]model.CommitInfo, error) {
	path = utils.NormalizePath(path)

	// Check cache
	if cached, ok := s.cache.GetTimeline(path); ok {
		return cached, nil
	}

	// Get history
	history, err := s.gitSvc.GetFileHistory(ctx, path)
	if err != nil {
		return nil, err
	}

	// Cache timeline
	s.cache.SetTimeline(path, history.AllCommits)

	return history.AllCommits, nil
}

// GetTree returns the directory tree structure.
func (s *ArticleService) GetTree(ctx context.Context, dir string) (*model.TreeNode, error) {
	dir = utils.NormalizePath(dir)

	// Check cache
	if cached, ok := s.cache.GetTree(dir); ok {
		return cached, nil
	}

	// Get tree
	tree, err := s.fileSvc.GetDirectoryTree(ctx, dir)
	if err != nil {
		return nil, err
	}

	// Cache tree
	s.cache.SetTree(dir, tree)

	return tree, nil
}

// Search searches articles by title.
func (s *ArticleService) Search(ctx context.Context, query string, dir string) ([]model.SearchResult, error) {
	// Normalize inputs
	query = strings.ToLower(query)
	dir = utils.NormalizePath(dir)

	// Get article list (use cache)
	articles, err := s.ListArticles(ctx, ListOptions{
		Dir:      dir,
		Sort:     model.SortEdited,
		Order:    model.OrderDesc,
		UseCache: true,
		Parallel: true,
	})
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

// InvalidateCache invalidates the article cache.
func (s *ArticleService) InvalidateCache() {
	s.cache.Invalidate()
}

// ClearCache clears the article cache.
func (s *ArticleService) ClearCache() {
	s.cache.Clear()
}

// GetCacheStats returns cache statistics.
func (s *ArticleService) GetCacheStats() CacheStats {
	return s.cache.Stats()
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
