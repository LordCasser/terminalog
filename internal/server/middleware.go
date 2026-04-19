// Package server provides HTTP server functionality.
package server

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

// HealthzHandler returns a simple health check handler.
func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// ReadyHandler returns a readiness check handler.
func ReadyHandler(check func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check() {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	}
}

// CORS middleware for development (if needed).
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimit middleware (simple implementation).
func RateLimit(requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Simple in-memory rate limiting
		// For production, use a more robust solution
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For MVP, we skip rate limiting
			next.ServeHTTP(w, r)
		})
	}
}

// Cache middleware for static assets.
func Cache(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age="+duration.String())
			next.ServeHTTP(w, r)
		})
	}
}

// Gzip compresses compressible HTTP responses when the client supports it.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldAttemptGzip(r) {
			next.ServeHTTP(w, r)
			return
		}

		gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
		defer gzipWriterPool.Put(gzipWriter)

		grw := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gzipWriter,
			statusCode:     http.StatusOK,
		}
		defer grw.finish()

		next.ServeHTTP(grw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter       *gzip.Writer
	statusCode       int
	wroteHeader      bool
	compressionReady bool
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.statusCode = statusCode
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.prepareHeaders(data)
	}

	if w.compressionReady {
		return w.gzipWriter.Write(data)
	}

	return w.ResponseWriter.Write(data)
}

func (w *gzipResponseWriter) Flush() {
	if !w.wroteHeader {
		w.prepareHeaders(nil)
	}

	if w.compressionReady {
		_ = w.gzipWriter.Flush()
	}

	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (w *gzipResponseWriter) finish() {
	if !w.wroteHeader {
		w.prepareHeaders(nil)
	}

	if w.compressionReady {
		_ = w.gzipWriter.Close()
		w.gzipWriter.Reset(io.Discard)
	}
}

func (w *gzipResponseWriter) prepareHeaders(peek []byte) {
	if w.wroteHeader {
		return
	}

	headers := w.Header()
	if headers.Get("Content-Type") == "" && len(peek) > 0 {
		headers.Set("Content-Type", http.DetectContentType(peek))
	}

	if shouldCompressResponse(w.statusCode, headers.Get("Content-Type"), headers.Get("Content-Encoding")) {
		headers.Set("Content-Encoding", "gzip")
		headers.Add("Vary", "Accept-Encoding")
		headers.Del("Content-Length")
		w.ResponseWriter.WriteHeader(w.statusCode)
		w.gzipWriter.Reset(w.ResponseWriter)
		w.compressionReady = true
		w.wroteHeader = true
		return
	}

	w.ResponseWriter.WriteHeader(w.statusCode)
	w.wroteHeader = true
}

func shouldAttemptGzip(r *http.Request) bool {
	if r.Method == http.MethodHead {
		return false
	}
	if r.Header.Get("Range") != "" {
		return false
	}
	if !strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip") {
		return false
	}
	connectionHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	if strings.Contains(connectionHeader, "upgrade") || upgradeHeader == "websocket" {
		return false
	}
	return true
}

func shouldCompressResponse(statusCode int, contentType, contentEncoding string) bool {
	if contentEncoding != "" {
		return false
	}
	if statusCode < 200 || statusCode == http.StatusNoContent || statusCode == http.StatusNotModified {
		return false
	}

	contentType = strings.ToLower(contentType)
	compressiblePrefixes := []string{
		"text/",
		"application/javascript",
		"application/json",
		"application/ld+json",
		"application/problem+json",
		"application/xml",
		"application/xhtml+xml",
		"application/rss+xml",
		"application/atom+xml",
		"image/svg+xml",
	}

	for _, prefix := range compressiblePrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}

	return false
}
