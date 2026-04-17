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
	// Clean the requested path to remove any .. or other dangerous components
	cleanedPath := filepath.Clean(requestedPath)

	// Check for path traversal attempts
	if strings.Contains(cleanedPath, "..") {
		return "", model.ErrInvalidPath
	}

	// Construct the full path
	fullPath := filepath.Join(baseDir, cleanedPath)

	// Get absolute paths for comparison
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute base path: %w", err)
	}

	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute full path: %w", err)
	}

	// Ensure the full path is within the base directory
	if !strings.HasPrefix(absFull, absBase) {
		return "", model.ErrInvalidPath
	}

	// Protect .git directory
	if strings.Contains(absFull, "/.git/") || strings.HasSuffix(absFull, "/.git") {
		return "", model.ErrInvalidPath
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
