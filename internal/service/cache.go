// Package service provides business logic services for the application.
package service

import (
	"sync"
	"time"

	"terminalog/internal/model"
)

// ArticleCache provides caching for article metadata and timelines.
type ArticleCache struct {
	// articles stores cached article metadata.
	articles map[string]*model.Article

	// timelines stores cached commit timelines.
	timelines map[string][]model.CommitInfo

	// treeCache stores cached directory trees.
	treeCache map[string]*model.TreeNode

	// mutex protects concurrent access.
	mutex sync.RWMutex

	// ttl is the cache time-to-live.
	ttl time.Duration

	// lastUpdate tracks when the cache was last updated.
	lastUpdate time.Time

	// articleListCache stores cached article lists.
	articleListCache map[string][]model.Article
}

// NewArticleCache creates a new ArticleCache with the given TTL.
func NewArticleCache(ttl time.Duration) *ArticleCache {
	return &ArticleCache{
		articles:         make(map[string]*model.Article),
		timelines:        make(map[string][]model.CommitInfo),
		treeCache:        make(map[string]*model.TreeNode),
		articleListCache: make(map[string][]model.Article),
		ttl:              ttl,
		lastUpdate:       time.Now(),
	}
}

// GetArticle retrieves a cached article by path.
func (c *ArticleCache) GetArticle(path string) (*model.Article, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check TTL
	if time.Since(c.lastUpdate) > c.ttl {
		return nil, false
	}

	article, ok := c.articles[path]
	return article, ok
}

// SetArticle stores an article in the cache.
func (c *ArticleCache) SetArticle(path string, article *model.Article) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.articles[path] = article
	c.lastUpdate = time.Now()
}

// GetTimeline retrieves a cached timeline by path.
func (c *ArticleCache) GetTimeline(path string) ([]model.CommitInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check TTL
	if time.Since(c.lastUpdate) > c.ttl {
		return nil, false
	}

	timeline, ok := c.timelines[path]
	return timeline, ok
}

// SetTimeline stores a timeline in the cache.
func (c *ArticleCache) SetTimeline(path string, timeline []model.CommitInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.timelines[path] = timeline
	c.lastUpdate = time.Now()
}

// GetTree retrieves a cached directory tree.
func (c *ArticleCache) GetTree(dir string) (*model.TreeNode, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check TTL
	if time.Since(c.lastUpdate) > c.ttl {
		return nil, false
	}

	tree, ok := c.treeCache[dir]
	return tree, ok
}

// SetTree stores a directory tree in the cache.
func (c *ArticleCache) SetTree(dir string, tree *model.TreeNode) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.treeCache[dir] = tree
	c.lastUpdate = time.Now()
}

// GetArticleList retrieves a cached article list for a directory.
func (c *ArticleCache) GetArticleList(dir string) ([]model.Article, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check TTL
	if time.Since(c.lastUpdate) > c.ttl {
		return nil, false
	}

	list, ok := c.articleListCache[dir]
	return list, ok
}

// SetArticleList stores an article list in the cache.
func (c *ArticleCache) SetArticleList(dir string, list []model.Article) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.articleListCache[dir] = list
	c.lastUpdate = time.Now()
}

// Clear clears all cache entries.
func (c *ArticleCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.articles = make(map[string]*model.Article)
	c.timelines = make(map[string][]model.CommitInfo)
	c.treeCache = make(map[string]*model.TreeNode)
	c.articleListCache = make(map[string][]model.Article)
	c.lastUpdate = time.Now()
}

// Invalidate invalidates the cache (marks it as expired).
func (c *ArticleCache) Invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastUpdate = time.Time{} // Reset to zero time
}

// Size returns the number of cached articles.
func (c *ArticleCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.articles)
}

// Stats returns cache statistics.
func (c *ArticleCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		Articles:     len(c.articles),
		Timelines:    len(c.timelines),
		Trees:        len(c.treeCache),
		ArticleLists: len(c.articleListCache),
		LastUpdate:   c.lastUpdate,
		TTL:          c.ttl,
	}
}

// CacheStats represents cache statistics.
type CacheStats struct {
	Articles     int           `json:"articles"`
	Timelines    int           `json:"timelines"`
	Trees        int           `json:"trees"`
	ArticleLists int           `json:"articleLists"`
	LastUpdate   time.Time     `json:"lastUpdate"`
	TTL          time.Duration `json:"ttl"`
}

// DefaultCacheTTL is the default cache TTL (5 minutes).
const DefaultCacheTTL = 5 * time.Minute
