// Package model defines the data structures used throughout the application.
package model

import "time"

// Article represents a blog article metadata.
type Article struct {
	// Path is the file path relative to the repository root (e.g., "tech/go-best-practices.md").
	Path string `json:"path"`

	// Name is the file or directory name (extracted from path).
	Name string `json:"name"`

	// Title is the article title extracted from the filename (without .md extension).
	Title string `json:"title"`

	// Type indicates whether this is a file or directory.
	Type NodeType `json:"type"`

	// CreatedAt is the timestamp of the first commit that created this file.
	CreatedAt time.Time `json:"createdAt"`

	// CreatedBy is the author name of the first commit.
	CreatedBy string `json:"createdBy"`

	// EditedAt is the timestamp of the last commit that modified this file.
	EditedAt time.Time `json:"editedAt"`

	// EditedBy is the author name of the last commit.
	EditedBy string `json:"editedBy"`

	// Contributors is a list of all author names who have committed changes to this file.
	Contributors []string `json:"contributors"`

	// LatestCommit is the latest commit message.
	LatestCommit string `json:"latestCommit"`
}

// ArticleDetail represents an article with its content.
type ArticleDetail struct {
	Article

	// Content is the raw Markdown content of the article.
	Content string `json:"content"`
}

// CommitInfo represents a single Git commit information.
type CommitInfo struct {
	// Hash is the short commit hash (7 characters).
	Hash string `json:"hash"`

	// Author is the commit author name.
	Author string `json:"author"`

	// Timestamp is the commit timestamp.
	Timestamp time.Time `json:"timestamp"`

	// Message is the commit message.
	Message string `json:"message"`
}

// NodeType represents the type of a tree node.
type NodeType string

const (
	// NodeTypeDir indicates a directory.
	NodeTypeDir NodeType = "dir"

	// NodeTypeFile indicates a file.
	NodeTypeFile NodeType = "file"
)

// TreeNode represents a node in the directory tree structure.
type TreeNode struct {
	// Name is the directory or file name.
	Name string `json:"name"`

	// Path is the full path relative to the repository root.
	Path string `json:"path"`

	// Type indicates whether this is a directory or file.
	Type NodeType `json:"type"`

	// Children contains child nodes (only for directories).
	Children []*TreeNode `json:"children,omitempty"`
}

// SearchResult represents a search result.
type SearchResult struct {
	// Path is the file path relative to the repository root.
	Path string `json:"path"`

	// Title is the article title.
	Title string `json:"title"`

	// MatchedTitle is the title with matched portion highlighted (for future use).
	MatchedTitle string `json:"matchedTitle"`
}

// FileHistory represents the complete Git history of a file.
type FileHistory struct {
	// FirstCommit is the commit that first created this file.
	FirstCommit CommitInfo `json:"firstCommit"`

	// LastCommit is the most recent commit that modified this file.
	LastCommit CommitInfo `json:"lastCommit"`

	// AllCommits is the list of all commits involving this file, ordered by time descending.
	AllCommits []CommitInfo `json:"allCommits"`

	// Contributors is the list of all unique author names.
	Contributors []string `json:"contributors"`
}

// User represents a user for authentication.
type User struct {
	// Username is the user's login name.
	Username string `json:"username"`

	// PasswordHash is the bcrypt hashed password.
	PasswordHash string `json:"-"`
}

// SortField represents the field to sort by.
type SortField string

const (
	// SortCreated sorts by creation time (first commit).
	SortCreated SortField = "created"

	// SortEdited sorts by last edit time (last commit).
	SortEdited SortField = "edited"

	// SortName sorts alphabetically by file/directory name.
	SortName SortField = "name"
)

// SortOrder represents the sort direction.
type SortOrder string

const (
	// OrderAsc sorts in ascending order (oldest first).
	OrderAsc SortOrder = "asc"

	// OrderDesc sorts in descending order (newest first).
	OrderDesc SortOrder = "desc"
)

// ParseSortField parses a string into SortField, defaulting to SortEdited.
func ParseSortField(s string) SortField {
	switch s {
	case "created":
		return SortCreated
	case "edited":
		return SortEdited
	case "name":
		return SortName
	default:
		return SortEdited
	}
}

// ParseSortOrder parses a string into SortOrder, defaulting to OrderDesc.
func ParseSortOrder(s string) SortOrder {
	switch s {
	case "asc":
		return OrderAsc
	case "desc":
		return OrderDesc
	default:
		return OrderDesc
	}
}

// ChangeType represents the type of change for version calculation.
type ChangeType string

const (
	// ChangeTypePatch indicates a small change (<10 lines).
	ChangeTypePatch ChangeType = "patch"

	// ChangeTypeMinor indicates a moderate change (10-50% of total lines).
	ChangeTypeMinor ChangeType = "minor"

	// ChangeTypeMajor indicates a large change (>50% of total lines).
	ChangeTypeMajor ChangeType = "major"
)

// VersionInfo represents version information for an article.
type VersionInfo struct {
	// CurrentVersion is the current semantic version (e.g., "v2.0.48").
	CurrentVersion string `json:"currentVersion"`

	// History contains historical version information.
	History []VersionHistoryEntry `json:"history"`
}

// VersionHistoryEntry represents a single version history entry.
type VersionHistoryEntry struct {
	// Version is the semantic version.
	Version string `json:"version"`

	// Hash is the short commit hash (7 characters).
	Hash string `json:"hash"`

	// Author is the commit author name.
	Author string `json:"author"`

	// Timestamp is the commit timestamp.
	Timestamp time.Time `json:"timestamp"`

	// LinesChanged is the number of lines changed in this commit.
	LinesChanged int `json:"linesChanged"`

	// ChangeType indicates whether this was a patch, minor, or major change.
	ChangeType ChangeType `json:"changeType"`
}

// AboutMeResponse represents the response for the About Me API.
type AboutMeResponse struct {
	// Path is the file path (_ABOUTME.md).
	Path string `json:"path"`

	// Title is the title extracted from the filename.
	Title string `json:"title"`

	// Content is the raw Markdown content.
	Content string `json:"content"`

	// Exists indicates whether the _ABOUTME.md file exists.
	Exists bool `json:"exists"`
}
