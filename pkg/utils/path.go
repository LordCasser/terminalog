// Package utils provides utility functions for the application.
package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"terminalog/internal/model"
)

// ValidatePath checks if the requested path is safe (no directory traversal).
// It returns the absolute validated path or an error.
func ValidatePath(baseDir, requestedPath string) (string, error) {
	// Clean the requested path to resolve .. and redundant separators.
	cleanedPath := filepath.Clean(requestedPath)

	// Reject explicitly absolute paths — filepath.Join would strip the
	// leading separator, making them appear to be inside baseDir.
	if filepath.IsAbs(cleanedPath) {
		return "", model.ErrInvalidPath
	}

	// Construct the full path relative to base directory.
	fullPath := filepath.Join(baseDir, cleanedPath)

	// Resolve both paths to absolute form.
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute base path: %w", err)
	}

	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute full path: %w", err)
	}

	// Use filepath.Rel to verify the resolved path does not escape baseDir.
	// If absFull is outside absBase, Rel returns a path starting with "..".
	relPath, err := filepath.Rel(absBase, absFull)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}
	if strings.HasPrefix(relPath, "..") {
		return "", model.ErrInvalidPath
	}

	// Protect .git directory from access.
	// Check each path segment individually — avoids false positives from
	// filenames containing ".git" as a substring.
	segments := strings.Split(filepath.ToSlash(relPath), "/")
	for _, seg := range segments {
		if seg == ".git" {
			return "", model.ErrInvalidPath
		}
	}

	return absFull, nil
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
