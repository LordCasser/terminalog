// Package utils provides utility functions for the application.
package utils

import (
	"path/filepath"
	"strings"
)

// MIME type mappings for common file types.
var mimeTypes = map[string]string{
	// Images
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",
	".bmp":  "image/bmp",

	// Documents
	".pdf": "application/pdf",

	// Text
	".txt":  "text/plain",
	".md":   "text/markdown",
	".html": "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",

	// Audio
	".mp3": "audio/mpeg",
	".wav": "audio/wav",
	".ogg": "audio/ogg",

	// Video
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
}

// GetMimeType returns the MIME type for a file based on its extension.
// If the extension is not recognized, it returns "application/octet-stream".
func GetMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	return "application/octet-stream"
}

// IsImageFile checks if a file is an image based on its MIME type.
func IsImageFile(path string) bool {
	mimeType := GetMimeType(path)
	return strings.HasPrefix(mimeType, "image/")
}
