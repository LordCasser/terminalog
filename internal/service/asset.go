// Package service provides business logic services for the application.
package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// Asset represents a static asset (image, etc.).
type Asset struct {
	// Data is the file content.
	Data []byte

	// ContentType is the MIME type.
	ContentType string

	// Size is the file size in bytes.
	Size int64
}

// AssetService provides asset operations.
type AssetService struct {
	fileSvc *FileService
}

// NewAssetService creates a new AssetService instance.
func NewAssetService(fileSvc *FileService) *AssetService {
	return &AssetService{
		fileSvc: fileSvc,
	}
}

// GetAsset returns an asset by path.
// The path should NOT include .assets directory - it's automatically resolved.
// For example, requesting "guides/images/photo.jpg" will look for:
// 1. guides/.assets/images/photo.jpg (assets in article directory)
// 2. .assets/guides/images/photo.jpg (assets in root directory)
func (s *AssetService) GetAsset(ctx context.Context, path string) (*Asset, error) {
	// Normalize path
	path = utils.NormalizePath(path)

	// Validate path for security (reject .., .git, etc.)
	baseDir := s.fileSvc.GetBaseDir()
	if _, err := utils.ValidatePath(baseDir, path); err != nil {
		return nil, err
	}

	// Try to find the asset in .assets directories
	// Strategy: try multiple locations where .assets might exist
	actualPath, err := s.resolveAssetPath(path)
	if err != nil {
		return nil, err
	}

	// Read file content
	content, err := s.fileSvc.ReadFile(ctx, actualPath)
	if err != nil {
		return nil, err
	}

	// Get MIME type
	contentType := utils.GetMimeType(path)

	return &Asset{
		Data:        content,
		ContentType: contentType,
		Size:        int64(len(content)),
	}, nil
}

// resolveAssetPath resolves the actual file path by searching .assets directories.
// Input: "guides/images/photo.jpg" (without .assets)
// Output: "guides/.assets/images/photo.jpg" (if found)
// Fallback: ".assets/guides/images/photo.jpg" (if found)
func (s *AssetService) resolveAssetPath(requestPath string) (string, error) {
	baseDir := s.fileSvc.GetBaseDir()

	// Strategy 1: Check if .assets exists at each directory level
	// For "guides/images/photo.jpg", try:
	// - "guides/.assets/images/photo.jpg"
	// - ".assets/guides/images/photo.jpg"

	// Split the path into components
	parts := strings.Split(requestPath, "/")
	if len(parts) < 1 {
		return "", model.ErrNotFound
	}

	// Try inserting .assets at each level (from deepest to shallowest)
	// e.g., for "guides/images/photo.jpg":
	// Level 2: "guides/.assets/images/photo.jpg" (assets in guides directory)
	// Level 0: ".assets/guides/images/photo.jpg" (assets in root directory)
	for i := len(parts) - 1; i >= 0; i-- {
		// Build path with .assets inserted at position i
		var pathParts []string
		for j := 0; j < i; j++ {
			pathParts = append(pathParts, parts[j])
		}
		pathParts = append(pathParts, ".assets")
		for j := i; j < len(parts); j++ {
			pathParts = append(pathParts, parts[j])
		}

		testPath := strings.Join(pathParts, "/")

		// Check if file exists
		absPath := filepath.Join(baseDir, testPath)
		if info, err := os.Stat(absPath); err == nil && !info.IsDir() {
			return testPath, nil
		}
	}

	return "", model.ErrNotFound
}
