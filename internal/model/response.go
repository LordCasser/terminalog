// Package model defines the data structures used throughout the application.
package model

// API response structures for JSON serialization.

// ArticleListResponse is the response for GET /api/articles.
type ArticleListResponse struct {
	// Articles is the list of articles.
	Articles []Article `json:"articles"`

	// CurrentDir is the current directory being browsed.
	CurrentDir string `json:"currentDir"`

	// Total is the total number of articles in the list.
	Total int `json:"total"`
}

// ArticleResponse is the response for GET /api/articles/{path}.
type ArticleResponse struct {
	// Path is the file path relative to the repository root.
	Path string `json:"path"`

	// Title is the article title.
	Title string `json:"title"`

	// Content is the raw Markdown content.
	Content string `json:"content"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"createdAt"`

	// CreatedBy is the creator name.
	CreatedBy string `json:"createdBy"`

	// EditedAt is the last edit timestamp.
	EditedAt string `json:"editedAt"`

	// EditedBy is the last editor name.
	EditedBy string `json:"editedBy"`

	// Contributors is the list of all contributors.
	Contributors []string `json:"contributors"`
}

// TimelineResponse is the response for GET /api/articles/{path}/timeline.
type TimelineResponse struct {
	// Commits is the list of commits for the article.
	Commits []CommitInfo `json:"commits"`
}

// TreeResponse is the response for GET /api/tree.
type TreeResponse struct {
	// Root is the root node of the directory tree.
	Root *TreeNode `json:"root"`

	// CurrentDir is the current directory.
	CurrentDir string `json:"currentDir"`
}

// SearchResponse is the response for GET /api/search.
type SearchResponse struct {
	// Results is the list of search results.
	Results []SearchResult `json:"results"`

	// Query is the search query.
	Query string `json:"query"`

	// Total is the total number of results.
	Total int `json:"total"`
}

// ErrorResponse is the response for error cases.
type ErrorResponse struct {
	// Error is the error message.
	Error string `json:"error"`

	// Code is the error code (optional).
	Code string `json:"code,omitempty"`
}
