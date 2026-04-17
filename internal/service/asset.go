// Package service provides business logic services for the application.
package service

import (
	"context"

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
func (s *AssetService) GetAsset(ctx context.Context, path string) (*Asset, error) {
	// Normalize path
	path = utils.NormalizePath(path)

	// Validate path
	if _, err := s.fileSvc.ValidatePath(path); err != nil {
		return nil, model.ErrInvalidPath
	}

	// Read file content
	content, err := s.fileSvc.ReadFile(ctx, path)
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
