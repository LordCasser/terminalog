// Package service provides business logic services for the application.
package service

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// ArticleService provides article-related operations.
type ArticleService struct {
	gitSvc *GitService
	cache  *ArticleCache
}

// NewArticleService creates a new ArticleService instance.
func NewArticleService(gitSvc *GitService) *ArticleService {
	return &ArticleService{
		gitSvc: gitSvc,
		cache:  NewArticleCache(DefaultCacheTTL),
	}
}

// NewArticleServiceWithCache creates a new ArticleService with custom cache.
func NewArticleServiceWithCache(gitSvc *GitService, cache *ArticleCache) *ArticleService {
	return &ArticleService{gitSvc: gitSvc, cache: cache}
}

// ListOptions contains options for listing articles.
type ListOptions struct {
	Dir      string
	Sort     model.SortField
	Order    model.SortOrder
	UseCache bool
}

// ResolveNodeType resolves a public content path against the current HEAD.
func (s *ArticleService) ResolveNodeType(ctx context.Context, contentPath string) (model.NodeType, error) {
	contentPath, err := utils.ValidateContentPath(contentPath)
	if err != nil {
		return "", err
	}
	nodeType, err := s.gitSvc.NodeTypeAtHead(contentPath)
	if err != nil {
		return "", err
	}
	if nodeType == model.NodeTypeFile {
		if !utils.IsMarkdownFile(contentPath) {
			return "", model.ErrNotFound
		}
		return nodeType, nil
	}

	articles, err := s.ListArticles(ctx, ListOptions{
		Dir: contentPath, Sort: model.SortName, Order: model.OrderAsc, UseCache: true,
	})
	if err != nil {
		return "", err
	}
	if len(articles) == 0 {
		return "", model.ErrNotFound
	}
	return nodeType, nil
}

// ListArticles returns every published Markdown article below a directory.
func (s *ArticleService) ListArticles(ctx context.Context, opts ListOptions) ([]model.Article, error) {
	dir, err := utils.ValidateContentPath(opts.Dir)
	if err != nil {
		return nil, err
	}
	generation := s.cache.Generation()
	cacheKey := dir + ":" + string(opts.Sort) + ":" + string(opts.Order)
	if opts.UseCache {
		if cached, ok := s.cache.GetArticleList(cacheKey); ok {
			return cached, nil
		}
	}

	files, err := s.gitSvc.ListMarkdownFilesAtHead(dir)
	if err != nil {
		return nil, err
	}
	histories, err := s.gitSvc.GetFileHistories(ctx, files)
	if err != nil {
		return nil, err
	}

	articles := make([]model.Article, 0, len(files))
	for _, file := range files {
		article, err := s.articleFromHistory(generation, file, histories[file])
		if err == nil {
			articles = append(articles, article)
		}
	}
	sortArticles(articles, opts.Sort, opts.Order)
	s.cache.SetArticleList(generation, cacheKey, articles)
	return articles, nil
}

// ListDirectory derives direct children from the same published article index
// used by search, tree, and completion.
func (s *ArticleService) ListDirectory(ctx context.Context, dir string, sortField model.SortField, sortOrder model.SortOrder) ([]model.Article, error) {
	dir, err := utils.ValidateContentPath(dir)
	if err != nil {
		return nil, err
	}
	generation := s.cache.Generation()
	cacheKey := dir + ":" + string(sortField) + ":" + string(sortOrder)
	if cached, ok := s.cache.GetDirectoryList(cacheKey); ok {
		return cached, nil
	}

	published, err := s.ListArticles(ctx, ListOptions{
		Dir: dir, Sort: model.SortEdited, Order: model.OrderDesc, UseCache: true,
	})
	if err != nil {
		return nil, err
	}

	articles := make([]model.Article, 0)
	directories := make(map[string]model.Article)
	prefix := ""
	if dir != "" {
		prefix = dir + "/"
	}

	for _, article := range published {
		relative := strings.TrimPrefix(article.Path, prefix)
		parts := strings.SplitN(relative, "/", 2)
		if len(parts) == 1 {
			articles = append(articles, article)
			continue
		}

		dirPath := parts[0]
		if dir != "" {
			dirPath = dir + "/" + dirPath
		}
		current, exists := directories[dirPath]
		if exists && !article.EditedAt.After(current.EditedAt) {
			continue
		}
		directories[dirPath] = model.Article{
			Path: dirPath, Name: filepath.Base(dirPath), Title: filepath.Base(dirPath),
			Type: model.NodeTypeDir, CreatedAt: article.CreatedAt, CreatedBy: article.CreatedBy,
			EditedAt: article.EditedAt, EditedBy: article.EditedBy,
			Contributors: article.Contributors, LatestCommit: article.LatestCommit,
		}
	}

	for _, directory := range directories {
		articles = append(articles, directory)
	}
	sortDirectoryListing(articles, sortField, sortOrder)
	s.cache.SetDirectoryList(generation, cacheKey, articles)
	return articles, nil
}

func (s *ArticleService) articleFromHistory(generation uint64, file string, history *model.FileHistory) (model.Article, error) {
	if history == nil || len(history.AllCommits) == 0 {
		return model.Article{}, model.ErrNotCommitted
	}
	article := model.Article{
		Path: file, Name: filepath.Base(file), Title: utils.ExtractTitle(file), Type: model.NodeTypeFile,
		CreatedAt:    history.FirstCommit.Timestamp,
		CreatedBy:    history.FirstCommit.Author,
		EditedAt:     history.LastCommit.Timestamp,
		EditedBy:     history.LastCommit.Author,
		Contributors: history.Contributors,
		LatestCommit: history.LastCommit.Message,
	}

	s.cache.SetArticle(generation, file, &article)

	return article, nil
}

// GetArticle returns the content and metadata of a specific article.
func (s *ArticleService) GetArticle(ctx context.Context, path string) (*model.ArticleDetail, error) {
	generation := s.cache.Generation()
	path, err := utils.ValidateContentPath(path)
	if err != nil {
		return nil, err
	}
	if !utils.IsMarkdownFile(path) {
		return nil, model.ErrNotFound
	}
	content, err := s.gitSvc.ReadFileAtHead(path)
	if err != nil {
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
		s.cache.SetArticle(generation, path, &article)
	}

	return &model.ArticleDetail{
		Article: article,
		Content: string(content),
	}, nil
}

// GetTimeline returns the commit timeline for an article.
func (s *ArticleService) GetTimeline(ctx context.Context, path string) ([]model.CommitInfo, error) {
	generation := s.cache.Generation()
	path, err := utils.ValidateContentPath(path)
	if err != nil {
		return nil, err
	}

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
	s.cache.SetTimeline(generation, path, history.AllCommits)

	return history.AllCommits, nil
}

// GetTree returns the directory tree structure.
func (s *ArticleService) GetTree(ctx context.Context, dir string) (*model.TreeNode, error) {
	generation := s.cache.Generation()
	dir, err := utils.ValidateContentPath(dir)
	if err != nil {
		return nil, err
	}

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
	s.cache.SetTree(generation, dir, tree)

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

// Search searches articles by title, file name, and directory name.
// It returns both matching articles (files) and matching directories.
// Directory results have Type=NodeTypeDir; article results have Type=NodeTypeFile.
func (s *ArticleService) Search(ctx context.Context, query string, dir string) ([]model.SearchResult, error) {
	// Normalize inputs
	query = strings.ToLower(query)
	dir = utils.NormalizePath(dir)

	articles, err := s.ListArticles(ctx, ListOptions{
		Dir:      dir,
		Sort:     model.SortEdited,
		Order:    model.OrderDesc,
		UseCache: true,
	})
	if err != nil {
		return nil, err
	}

	results := make([]model.SearchResult, 0)
	directories := make(map[string]struct{})
	for _, article := range articles {
		parent := utils.NormalizePath(filepath.Dir(article.Path))
		for parent != "" && parent != "." && parent != dir {
			if dir != "" && !strings.HasPrefix(parent, dir+"/") {
				break
			}
			if strings.Contains(strings.ToLower(filepath.Base(parent)), query) {
				directories[parent] = struct{}{}
			}
			next := utils.NormalizePath(filepath.Dir(parent))
			if next == parent {
				break
			}
			parent = next
		}
	}

	for path := range directories {
		name := filepath.Base(path)
		results = append(results, model.SearchResult{
			Path:         path,
			Title:        name,
			MatchedTitle: name,
			Type:         model.NodeTypeDir,
		})
	}

	for _, article := range articles {
		titleLower := strings.ToLower(article.Title)
		fileNameLower := strings.ToLower(filepath.Base(article.Path))

		// Match by title or file name
		if strings.Contains(titleLower, query) || strings.Contains(fileNameLower, query) {
			results = append(results, model.SearchResult{
				Path:         article.Path,
				Title:        article.Title,
				MatchedTitle: article.Title,
				Type:         model.NodeTypeFile,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Type != results[j].Type {
			return results[i].Type == model.NodeTypeDir
		}
		return results[i].Path < results[j].Path
	})

	return results, nil
}

// InvalidateCache invalidates the article cache.
func (s *ArticleService) InvalidateCache() {
	s.cache.Invalidate()
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
		var comparison int
		switch sortField {
		case model.SortCreated:
			comparison = articles[i].CreatedAt.Compare(articles[j].CreatedAt)
		case model.SortEdited:
			comparison = articles[i].EditedAt.Compare(articles[j].EditedAt)
		default:
			comparison = strings.Compare(articles[i].Name, articles[j].Name)
		}
		if comparison == 0 {
			comparison = strings.Compare(articles[i].Path, articles[j].Path)
		}
		if sortOrder == model.OrderDesc {
			return comparison > 0
		}
		return comparison < 0
	})
}

// Helper function: sortArticles sorts articles by the given field and order.
func sortArticles(articles []model.Article, sortField model.SortField, order model.SortOrder) {
	sort.Slice(articles, func(i, j int) bool {
		var comparison int

		switch sortField {
		case model.SortCreated:
			comparison = articles[i].CreatedAt.Compare(articles[j].CreatedAt)
		case model.SortEdited:
			comparison = articles[i].EditedAt.Compare(articles[j].EditedAt)
		default:
			comparison = articles[i].EditedAt.Compare(articles[j].EditedAt)
		}
		if comparison == 0 {
			comparison = strings.Compare(articles[i].Path, articles[j].Path)
		}
		if order == model.OrderDesc {
			return comparison > 0
		}
		return comparison < 0
	})
}
