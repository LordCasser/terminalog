// Package model defines the data structures used throughout the application.
package model

import "errors"

// Common errors used throughout the application.

// ErrNotFound indicates the requested resource was not found.
var ErrNotFound = errors.New("resource not found")

// ErrNotCommitted indicates the file has not been committed to Git.
var ErrNotCommitted = errors.New("file not committed to git")

// ErrInvalidPath indicates the path is invalid or contains unsafe characters.
var ErrInvalidPath = errors.New("invalid path")

// ErrUnauthorized indicates authentication is required or failed.
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden indicates the user lacks permission for the operation.
var ErrForbidden = errors.New("forbidden")

// ErrInvalidConfig indicates the configuration is invalid.
var ErrInvalidConfig = errors.New("invalid configuration")

// ErrRepoNotFound indicates the Git repository was not found.
var ErrRepoNotFound = errors.New("git repository not found")

// ErrNotMarkdown indicates the file is not a Markdown file.
var ErrNotMarkdown = errors.New("not a markdown file")
