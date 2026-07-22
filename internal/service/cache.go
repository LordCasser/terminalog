// Package service provides business logic services for the application.
package service

import (
	"sync"
	"time"

	"terminalog/internal/model"
)

type cacheEntry[T any] struct {
	value     T
	expiresAt time.Time
}

// ArticleCache provides caching for article metadata and timelines.
type ArticleCache struct {
	// articles stores cached article metadata.
	articles map[string]cacheEntry[*model.Article]

	// timelines stores cached commit timelines.
	timelines map[string]cacheEntry[[]model.CommitInfo]

	// treeCache stores cached directory trees.
	treeCache map[string]cacheEntry[*model.TreeNode]

	// mutex protects concurrent access.
	mutex sync.RWMutex

	// ttl is the cache time-to-live.
	ttl time.Duration

	// lastUpdate tracks when the cache was last updated.
	lastUpdate time.Time

	// generation prevents a request that started before invalidation from
	// writing stale data back after a repository update.
	generation uint64

	// articleListCache stores cached article lists.
	articleListCache map[string]cacheEntry[[]model.Article]

	// directoryListCache stores cached directory listings.
	directoryListCache map[string]cacheEntry[[]model.Article]
}

// NewArticleCache creates a new ArticleCache with the given TTL.
func NewArticleCache(ttl time.Duration) *ArticleCache {
	return &ArticleCache{
		articles:           make(map[string]cacheEntry[*model.Article]),
		timelines:          make(map[string]cacheEntry[[]model.CommitInfo]),
		treeCache:          make(map[string]cacheEntry[*model.TreeNode]),
		articleListCache:   make(map[string]cacheEntry[[]model.Article]),
		directoryListCache: make(map[string]cacheEntry[[]model.Article]),
		ttl:                ttl,
		lastUpdate:         time.Now(),
		generation:         1,
	}
}

// Generation returns the current repository-cache generation. Callers capture
// it before loading data and pass it back when storing the result.
func (c *ArticleCache) Generation() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.generation
}

// GetArticle retrieves a cached article by path.
func (c *ArticleCache) GetArticle(path string) (*model.Article, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.articles[path]
	return entry.value, ok && time.Now().Before(entry.expiresAt)
}

// SetArticle stores an article in the cache.
func (c *ArticleCache) SetArticle(generation uint64, path string, article *model.Article) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if generation != c.generation {
		return false
	}
	c.articles[path] = cacheEntry[*model.Article]{value: article, expiresAt: time.Now().Add(c.ttl)}
	c.lastUpdate = time.Now()
	return true
}

// GetTimeline retrieves a cached timeline by path.
func (c *ArticleCache) GetTimeline(path string) ([]model.CommitInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.timelines[path]
	return entry.value, ok && time.Now().Before(entry.expiresAt)
}

// SetTimeline stores a timeline in the cache.
func (c *ArticleCache) SetTimeline(generation uint64, path string, timeline []model.CommitInfo) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if generation != c.generation {
		return false
	}
	c.timelines[path] = cacheEntry[[]model.CommitInfo]{value: timeline, expiresAt: time.Now().Add(c.ttl)}
	c.lastUpdate = time.Now()
	return true
}

// GetTree retrieves a cached directory tree.
func (c *ArticleCache) GetTree(dir string) (*model.TreeNode, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.treeCache[dir]
	return entry.value, ok && time.Now().Before(entry.expiresAt)
}

// SetTree stores a directory tree in the cache.
func (c *ArticleCache) SetTree(generation uint64, dir string, tree *model.TreeNode) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if generation != c.generation {
		return false
	}
	c.treeCache[dir] = cacheEntry[*model.TreeNode]{value: tree, expiresAt: time.Now().Add(c.ttl)}
	c.lastUpdate = time.Now()
	return true
}

// GetArticleList retrieves a cached article list for a directory.
func (c *ArticleCache) GetArticleList(dir string) ([]model.Article, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.articleListCache[dir]
	return entry.value, ok && time.Now().Before(entry.expiresAt)
}

// SetArticleList stores an article list in the cache.
func (c *ArticleCache) SetArticleList(generation uint64, dir string, list []model.Article) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if generation != c.generation {
		return false
	}
	c.articleListCache[dir] = cacheEntry[[]model.Article]{value: list, expiresAt: time.Now().Add(c.ttl)}
	c.lastUpdate = time.Now()
	return true
}

// GetDirectoryList retrieves a cached directory listing.
func (c *ArticleCache) GetDirectoryList(key string) ([]model.Article, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.directoryListCache[key]
	return entry.value, ok && time.Now().Before(entry.expiresAt)
}

// SetDirectoryList stores a directory listing in the cache.
func (c *ArticleCache) SetDirectoryList(generation uint64, key string, list []model.Article) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if generation != c.generation {
		return false
	}
	c.directoryListCache[key] = cacheEntry[[]model.Article]{value: list, expiresAt: time.Now().Add(c.ttl)}
	c.lastUpdate = time.Now()
	return true
}

// Invalidate advances the generation and removes every entry atomically.
func (c *ArticleCache) Invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.clearLocked()
	c.generation++
	c.lastUpdate = time.Now()
}

func (c *ArticleCache) clearLocked() {
	c.articles = make(map[string]cacheEntry[*model.Article])
	c.timelines = make(map[string]cacheEntry[[]model.CommitInfo])
	c.treeCache = make(map[string]cacheEntry[*model.TreeNode])
	c.articleListCache = make(map[string]cacheEntry[[]model.Article])
	c.directoryListCache = make(map[string]cacheEntry[[]model.Article])
}

// Stats returns cache statistics.
func (c *ArticleCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		Articles:       len(c.articles),
		Timelines:      len(c.timelines),
		Trees:          len(c.treeCache),
		ArticleLists:   len(c.articleListCache),
		DirectoryLists: len(c.directoryListCache),
		LastUpdate:     c.lastUpdate,
		TTL:            c.ttl,
		Generation:     c.generation,
	}
}

// CacheStats represents cache statistics.
type CacheStats struct {
	Articles       int           `json:"articles"`
	Timelines      int           `json:"timelines"`
	Trees          int           `json:"trees"`
	ArticleLists   int           `json:"articleLists"`
	DirectoryLists int           `json:"directoryLists"`
	LastUpdate     time.Time     `json:"lastUpdate"`
	TTL            time.Duration `json:"ttl"`
	Generation     uint64        `json:"generation"`
}

// DefaultCacheTTL is the default cache TTL (5 minutes).
const DefaultCacheTTL = 5 * time.Minute
