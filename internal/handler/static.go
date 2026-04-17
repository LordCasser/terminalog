// Package handler provides HTTP request handlers for the application.
package handler

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

// StaticHandler handles static file serving for the frontend.
type StaticHandler struct {
	fs fs.FS
}

// NewStaticHandler creates a new StaticHandler with the given embedded filesystem.
func NewStaticHandler(embedFS embed.FS) *StaticHandler {
	// Get the static subdirectory
	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		// If static directory doesn't exist, use empty FS
		staticFS = embedFS
	}

	return &StaticHandler{
		fs: staticFS,
	}
}

// ServeHTTP handles static file requests.
// It implements the http.Handler interface.
func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the request path
	path := r.URL.Path

	// Handle root path
	if path == "/" || path == "" {
		h.serveFile(w, r, "index.html")
		return
	}

	// Clean path (remove leading slash)
	path = strings.TrimPrefix(path, "/")

	// Try to serve the path directly
	if h.tryServePath(w, r, path) {
		return
	}

	// Fallback to index.html for SPA routing
	h.serveFile(w, r, "index.html")
}

// tryServePath attempts to serve a file at the given path.
// Returns true if file was served, false otherwise.
func (h *StaticHandler) tryServePath(w http.ResponseWriter, r *http.Request, path string) bool {
	// Try direct file
	if h.serveFile(w, r, path) {
		return true
	}

	// Try with index.html for directories
	indexPath := strings.TrimSuffix(path, "/") + "/index.html"
	indexPath = strings.TrimPrefix(indexPath, "/")
	if h.serveFile(w, r, indexPath) {
		return true
	}

	// Try with .html extension (for Next.js routes like /article -> article.html)
	if !strings.HasSuffix(path, ".html") && !strings.Contains(path, ".") && !strings.HasSuffix(path, "/") {
		htmlPath := path + ".html"
		if h.serveFile(w, r, htmlPath) {
			return true
		}
	}

	return false
}

// serveFile serves a file from the embedded filesystem.
// Returns true if file was served, false if not found.
func (h *StaticHandler) serveFile(w http.ResponseWriter, r *http.Request, path string) bool {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	// Open file
	file, err := h.fs.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Get file info
	stat, err := file.Stat()
	if err != nil {
		return false
	}

	// Don't serve directories
	if stat.IsDir() {
		return false
	}

	// Set content type based on extension
	contentType := getContentType(path)
	w.Header().Set("Content-Type", contentType)

	// Serve file content
	io.Copy(w, file)
	return true
}

// getContentType returns the MIME type for a file path.
func getContentType(path string) string {
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".html") {
		return "text/html; charset=utf-8"
	}
	if strings.HasSuffix(ext, ".css") {
		return "text/css; charset=utf-8"
	}
	if strings.HasSuffix(ext, ".js") {
		return "application/javascript"
	}
	if strings.HasSuffix(ext, ".json") {
		return "application/json"
	}
	if strings.HasSuffix(ext, ".svg") {
		return "image/svg+xml"
	}
	if strings.HasSuffix(ext, ".png") {
		return "image/png"
	}
	if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		return "image/jpeg"
	}
	if strings.HasSuffix(ext, ".ico") {
		return "image/x-icon"
	}
	if strings.HasSuffix(ext, ".woff") || strings.HasSuffix(ext, ".woff2") {
		return "font/woff2"
	}
	return "application/octet-stream"
}

// SetFS sets the filesystem for the handler (used for testing).
func (h *StaticHandler) SetFS(newFS fs.FS) {
	h.fs = newFS
}
