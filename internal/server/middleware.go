// Package server provides HTTP server functionality.
package server

import (
	"net/http"
	"time"
)

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
