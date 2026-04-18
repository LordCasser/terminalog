// Package service provides business logic services for the application.
package service

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

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

// ListArticles returns a list of articles in the given directory (recursive).
// Only committed Markdown files are included.
// Deprecated: Use ListDirectory for hierarchical browsing.
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

// ListDirectory returns the direct children (subdirectories and files) of a directory.
// This implements hierarchical browsing: directories are shown as navigable items,
// files are shown as viewable items. Subdirectories are only included if they
// contain at least one markdown file recursively.
// Sort controls the order of items within each group (dirs and files separately).
func (s *ArticleService) ListDirectory(ctx context.Context, dir string, sortField model.SortField, sortOrder model.SortOrder) ([]model.Article, error) {
	// Scan the directory for direct children
	entries, err := s.fileSvc.ScanDirectory(ctx, dir)
	if err != nil {
		return nil, err
	}

	articles := make([]model.Article, 0, len(entries))

	for _, entry := range entries {
		if entry.Type == model.NodeTypeDir {
			dirArticle, err := s.getDirectoryArticle(ctx, entry.Path)
			if err != nil {
				dirArticle = model.Article{
					Path:  entry.Path,
					Name:  entry.Name,
					Title: entry.Name,
					Type:  model.NodeTypeDir,
				}
			}
			articles = append(articles, dirArticle)
		} else {
			article, err := s.processFile(ctx, entry.Path)
			if err != nil {
				continue
			}
			articles = append(articles, article)
		}
	}

	// Sort: directories first, then files; within each group, apply sort criteria
	sortDirectoryListing(articles, sortField, sortOrder)

	return articles, nil
}

// getDirectoryArticle creates an Article entry for a directory.
// It uses metadata from the most recently edited file in the directory as the
// directory's metadata, providing useful information without requiring git
// operations on the directory itself.
func (s *ArticleService) getDirectoryArticle(ctx context.Context, dirPath string) (model.Article, error) {
	// Find the most recently edited file in this directory (recursive)
	files, err := s.fileSvc.ScanMarkdownFiles(ctx, dirPath)
	if err != nil {
		return model.Article{}, err
	}

	if len(files) == 0 {
		return model.Article{
			Path:  dirPath,
			Name:  filepath.Base(dirPath),
			Title: filepath.Base(dirPath),
			Type:  model.NodeTypeDir,
		}, nil
	}

	// Process each file and find the one with the latest edit time
	var latestArticle *model.Article
	var latestTime time.Time

	for _, file := range files {
		article, err := s.processFile(ctx, file)
		if err != nil {
			continue
		}
		if article.EditedAt.After(latestTime) {
			latestTime = article.EditedAt
			latestArticle = &article
		}
	}

	dirName := filepath.Base(dirPath)
	result := model.Article{
		Path:  dirPath,
		Name:  dirName,
		Title: dirName,
		Type:  model.NodeTypeDir,
	}

	// Use metadata from the most recently edited file for the directory
	if latestArticle != nil {
		result.CreatedAt = latestArticle.CreatedAt
		result.CreatedBy = latestArticle.CreatedBy
		result.EditedAt = latestArticle.EditedAt
		result.EditedBy = latestArticle.EditedBy
		result.LatestCommit = latestArticle.LatestCommit
		result.Contributors = latestArticle.Contributors
	}

	return result, nil
}
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
		Name:         filepath.Base(file),
		Title:        utils.ExtractTitle(file),
		Type:         model.NodeTypeFile,
		CreatedAt:    history.FirstCommit.Timestamp,
		CreatedBy:    history.FirstCommit.Author,
		EditedAt:     history.LastCommit.Timestamp,
		EditedBy:     history.LastCommit.Author,
		Contributors: history.Contributors,
		LatestCommit: history.LastCommit.Message,
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
			Name:         filepath.Base(path),
			Title:        utils.ExtractTitle(path),
			Type:         model.NodeTypeFile,
			CreatedAt:    history.FirstCommit.Timestamp,
			CreatedBy:    history.FirstCommit.Author,
			EditedAt:     history.LastCommit.Timestamp,
			EditedBy:     history.LastCommit.Author,
			Contributors: history.Contributors,
			LatestCommit: history.LastCommit.Message,
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

	// Get list of committed articles first
	articles, err := s.ListArticles(ctx, ListOptions{
		Dir:      dir,
		Sort:     model.SortEdited,
		Order:    model.OrderDesc,
		UseCache: true,
	})
	if err != nil {
		return nil, err
	}

	// Build tree from committed articles only
	tree := buildTreeFromArticles(articles, dir)

	// Cache tree
	s.cache.SetTree(dir, tree)

	return tree, nil
}

// buildTreeFromArticles builds a tree structure from a list of committed articles.
func buildTreeFromArticles(articles []model.Article, rootDir string) *model.TreeNode {
	rootDir = utils.NormalizePath(rootDir)
	rootName := filepath.Base(rootDir)
	if rootDir == "" || rootDir == "/" {
		rootName = "root"
		rootDir = ""
	}

	root := &model.TreeNode{
		Name:     rootName,
		Path:     rootDir,
		Type:     model.NodeTypeDir,
		Children: make([]*model.TreeNode, 0),
	}

	// Map to track directory nodes
	dirNodes := make(map[string]*model.TreeNode)
	dirNodes[rootDir] = root

	// Sort articles by path for consistent ordering
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Path < articles[j].Path
	})

	// Add each article to the tree
	for _, article := range articles {
		// Normalize the article path
		articlePath := utils.NormalizePath(article.Path)

		// Get the directory part of the path
		articleDir := filepath.Dir(articlePath)
		articleDir = utils.NormalizePath(articleDir)

		// Handle root directory case
		if articleDir == "." || articleDir == "" {
			articleDir = ""
		}

		// Ensure parent directories exist
		ensureParentDirs(dirNodes, articleDir, rootDir)

		// Get the parent node (should exist now)
		parent, ok := dirNodes[articleDir]
		if !ok {
			// Skip if parent doesn't exist (shouldn't happen)
			continue
		}

		// Add file node
		fileName := filepath.Base(articlePath)
		parent.Children = append(parent.Children, &model.TreeNode{
			Name: fileName,
			Path: articlePath,
			Type: model.NodeTypeFile,
		})
	}

	// Sort children in each directory (directories first, then files, alphabetically)
	sortTreeChildren(dirNodes)

	return root
}

// ensureParentDirs ensures all parent directories exist in the tree.
func ensureParentDirs(dirNodes map[string]*model.TreeNode, targetDir, rootDir string) {
	// Already exists
	if _, ok := dirNodes[targetDir]; ok {
		return
	}

	// Handle root directory case
	if targetDir == "" || targetDir == rootDir {
		return
	}

	// Split the path into components
	parts := strings.Split(strings.Trim(targetDir, "/"), "/")
	if len(parts) == 0 {
		return
	}

	// Build path from root to target
	currentPath := rootDir
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Build child path
		var childPath string
		if currentPath == "" {
			childPath = part
		} else {
			childPath = currentPath + "/" + part
		}
		childPath = utils.NormalizePath(childPath)

		// Create directory node if it doesn't exist
		if _, ok := dirNodes[childPath]; !ok {
			dirNodes[childPath] = &model.TreeNode{
				Name:     part,
				Path:     childPath,
				Type:     model.NodeTypeDir,
				Children: make([]*model.TreeNode, 0),
			}

			// Add to parent
			parent, ok := dirNodes[currentPath]
			if ok {
				parent.Children = append(parent.Children, dirNodes[childPath])
			}
		}

		currentPath = childPath
	}
}

// sortTreeChildren sorts children in each directory node.
func sortTreeChildren(dirNodes map[string]*model.TreeNode) {
	for _, node := range dirNodes {
		if node.Type == model.NodeTypeDir && len(node.Children) > 0 {
			sort.Slice(node.Children, func(i, j int) bool {
				// Directories first
				if node.Children[i].Type != node.Children[j].Type {
					return node.Children[i].Type == model.NodeTypeDir
				}
				// Alphabetically
				return node.Children[i].Name < node.Children[j].Name
			})
		}
	}
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

// sortDirectoryListing sorts articles in a directory listing.
// Directories always come first, then files. Within each group,
// items are sorted by the specified field and order.
func sortDirectoryListing(articles []model.Article, sortField model.SortField, sortOrder model.SortOrder) {
	sort.Slice(articles, func(i, j int) bool {
		// Directories always first
		if articles[i].Type != articles[j].Type {
			return articles[i].Type == model.NodeTypeDir
		}

		// Within same type group, sort by field
		var less bool
		switch sortField {
		case model.SortCreated:
			less = articles[i].CreatedAt.Before(articles[j].CreatedAt)
		case model.SortEdited:
			less = articles[i].EditedAt.Before(articles[j].EditedAt)
		default:
			// Default: alphabetical by name
			less = articles[i].Name < articles[j].Name
		}

		if sortOrder == model.OrderDesc {
			return !less
		}
		return less
	})
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
