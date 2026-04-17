// Package handler provides HTTP request handlers for the application.
package handler

import (
	"embed"
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
		path = "/index.html"
	}

	// Clean path
	path = strings.TrimPrefix(path, "/")

	// Try to open the file directly
	file, err := h.fs.Open(path)
	if err == nil {
		file.Close()
		// File exists, serve it
		http.FileServerFS(h.fs).ServeHTTP(w, r)
		return
	}

	// File doesn't exist, try with .html extension (for Next.js routes)
	if !strings.HasSuffix(path, ".html") && !strings.Contains(path, ".") {
		htmlPath := path + ".html"
		file, err = h.fs.Open(htmlPath)
		if err == nil {
			file.Close()
			// Serve the HTML file
			r.URL.Path = "/" + htmlPath
			http.FileServerFS(h.fs).ServeHTTP(w, r)
			return
		}
	}

	// Fallback to index.html for SPA routing
	indexFile, err := h.fs.Open("index.html")
	if err == nil {
		indexFile.Close()
		r.URL.Path = "/index.html"
		http.FileServerFS(h.fs).ServeHTTP(w, r)
		return
	}

	// No index.html found, return 404
	http.NotFound(w, r)
}

// SetFS sets the filesystem for the handler (used for testing).
func (h *StaticHandler) SetFS(newFS fs.FS) {
	h.fs = newFS
}
