// Package utils provides utility functions for the application.
package utils

import (
	"path"
	"path/filepath"
	"strings"

	"terminalog/internal/model"
)

// ValidateContentPath returns a canonical slash-separated path relative to a
// Git content root. Absolute paths, traversal, and .git access are rejected.
func ValidateContentPath(requestedPath string) (string, error) {
	if filepath.IsAbs(requestedPath) || filepath.VolumeName(requestedPath) != "" {
		return "", model.ErrInvalidPath
	}

	normalized := strings.ReplaceAll(requestedPath, "\\", "/")
	if strings.HasPrefix(normalized, "/") {
		return "", model.ErrInvalidPath
	}

	cleaned := path.Clean(normalized)
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", model.ErrInvalidPath
	}

	for _, segment := range strings.Split(cleaned, "/") {
		if segment == ".git" {
			return "", model.ErrInvalidPath
		}
	}

	return cleaned, nil
}

// ExtractTitle extracts the article title from a file path.
// It returns the filename without the .md extension.
func ExtractTitle(path string) string {
	// Get the filename
	filename := filepath.Base(path)

	// Remove .md extension
	return strings.TrimSuffix(filename, ".md")
}

// IsMarkdownFile checks if a file is a Markdown file based on its extension.
func IsMarkdownFile(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".md")
}

// NormalizePath normalizes a path for consistent representation.
// It ensures consistent forward slashes and removes leading/trailing slashes.
func NormalizePath(path string) string {
	// Replace backslashes with forward slashes (Windows compatibility)
	path = strings.ReplaceAll(path, "\\", "/")

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	return path
}
