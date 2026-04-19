// Package handler provides HTTP request handlers for the application.
package handler

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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

	// Try Next.js static export fallback for dynamic routes
	// Next.js generates _fallback directories for catch-all routes:
	//   /article/tech/go-guide.md -> article/_fallback/index.html
	//   /dir/tech/subdir -> dir/_fallback/index.html
	fallbackPath := h.getFallbackPath(path)
	if fallbackPath != "" {
		if h.serveFile(w, r, fallbackPath) {
			return true
		}
	}

	return false
}

// getFallbackPath returns the Next.js _fallback path for a dynamic route.
// For catch-all routes like /article/[...slug], Next.js static export generates
// article/_fallback/index.html. We need to map any /article/* path to this fallback.
func (h *StaticHandler) getFallbackPath(path string) string {
	// Split path into segments
	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return ""
	}

	// Check for known dynamic route prefixes
	dynamicRoutes := []string{"article", "dir"}
	for _, route := range dynamicRoutes {
		if segments[0] == route {
			return route + "/_fallback/index.html"
		}
	}

	return ""
}

// serveFile serves a file from the embedded filesystem.
// Returns true if file was served, false if not found.
func (h *StaticHandler) serveFile(w http.ResponseWriter, r *http.Request, path string) bool {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	servedPath, contentEncoding, ok := h.resolveCompressedPath(r, path)
	if !ok {
		return false
	}

	// Open file
	file, err := h.fs.Open(servedPath)
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
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	if contentEncoding != "" {
		w.Header().Set("Content-Encoding", contentEncoding)
		w.Header().Set("Vary", "Accept-Encoding")
	}
	if cacheControl := getCacheControl(path); cacheControl != "" {
		w.Header().Set("Cache-Control", cacheControl)
	}

	// Serve file content
	io.Copy(w, file)
	return true
}

func (h *StaticHandler) resolveCompressedPath(r *http.Request, path string) (string, string, bool) {
	if !isCompressibleAsset(path) || r.Header.Get("Range") != "" {
		if _, err := fs.Stat(h.fs, path); err != nil {
			return "", "", false
		}
		return path, "", true
	}

	acceptEncoding := strings.ToLower(r.Header.Get("Accept-Encoding"))
	candidates := []struct {
		suffix   string
		encoding string
	}{
		{suffix: ".br", encoding: "br"},
		{suffix: ".gz", encoding: "gzip"},
	}

	for _, candidate := range candidates {
		if !strings.Contains(acceptEncoding, candidate.encoding) {
			continue
		}
		compressedPath := path + candidate.suffix
		if _, err := fs.Stat(h.fs, compressedPath); err == nil {
			return compressedPath, candidate.encoding, true
		}
	}

	if _, err := fs.Stat(h.fs, path); err != nil {
		return "", "", false
	}
	return path, "", true
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

func isCompressibleAsset(path string) bool {
	lowerPath := strings.ToLower(path)
	compressibleExtensions := []string{
		".css",
		".html",
		".js",
		".json",
		".map",
		".md",
		".svg",
		".txt",
		".xml",
	}

	for _, ext := range compressibleExtensions {
		if strings.HasSuffix(lowerPath, ext) {
			return true
		}
	}

	return false
}

func getCacheControl(path string) string {
	lowerPath := strings.ToLower(path)
	if strings.Contains(lowerPath, "_next/static/") {
		return "public, max-age=31536000, immutable"
	}
	if strings.HasSuffix(lowerPath, ".html") {
		return "public, max-age=0, must-revalidate"
	}
	return "public, max-age=3600"
}

// ServeResources handles static resource requests for /api/v1/resources/*.
// This route is used for serving frontend compiled resources (like _next/static/chunks).
// It strips the /api/v1/resources prefix and serves from the embedded static directory.
func (h *StaticHandler) ServeResources(w http.ResponseWriter, r *http.Request) {
	// Get the path from chi wildcard parameter
	// Path will be like "_next/static/chunks/main.js"
	path := chi.URLParam(r, "*")

	// Clean path
	path = strings.TrimPrefix(path, "/")

	// Try to serve the file directly
	if h.serveFile(w, r, path) {
		return
	}

	// File not found
	http.NotFound(w, r)
}

// SetFS sets the filesystem for the handler (used for testing).
func (h *StaticHandler) SetFS(newFS fs.FS) {
	h.fs = newFS
}
