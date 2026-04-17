// Package server provides HTTP server functionality.
package server

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"terminalog/internal/handler"
)

// Server represents the HTTP server.
type Server struct {
	// addr is the server address (host:port).
	addr string

	// router is the chi router.
	router *chi.Mux

	// server is the underlying HTTP server.
	server *http.Server

	// logger is the structured logger.
	logger *slog.Logger

	// Handlers contains all HTTP handlers.
	Handlers *Handlers
}

// Handlers contains all HTTP handlers for the server.
type Handlers struct {
	Article *handler.ArticleHandler
	Asset   *handler.AssetHandler
	Git     *handler.GitHandler
	Search  *handler.SearchHandler
	Tree    *handler.TreeHandler
	Static  *handler.StaticHandler
	Health  *handler.HealthHandler
}

// NewServer creates a new Server instance.
func NewServer(addr string, handlers *Handlers, logger *slog.Logger, embedFS embed.FS) *Server {
	// Initialize static handler with embedded files
	handlers.Static = handler.NewStaticHandler(embedFS)

	s := &Server{
		addr:     addr,
		router:   chi.NewRouter(),
		logger:   logger,
		Handlers: handlers,
	}

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return s
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.logger.Info("Server starting", "addr", s.addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Server shutting down")
	return s.server.Shutdown(ctx)
}

// setupRoutes configures all routes for the server.
func (s *Server) setupRoutes() {
	r := s.router

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(s.loggingMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check routes (no auth required)
	if s.Handlers.Health != nil {
		r.Get("/healthz", s.Handlers.Health.Healthz)
		r.Get("/readyz", s.Handlers.Health.Readyz)
		r.Get("/livez", s.Handlers.Health.Livez)
		r.Get("/status", s.Handlers.Health.Status)
	}

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Articles API
		r.Get("/articles", s.Handlers.Article.List)
		r.Get("/articles/{path}", s.Handlers.Article.Get)
		r.Get("/articles/{path}/timeline", s.Handlers.Article.Timeline)

		// Tree API
		r.Get("/tree", s.Handlers.Tree.Get)

		// Search API
		r.Get("/search", s.Handlers.Search.Search)

		// Assets API
		r.Get("/assets/{path}", s.Handlers.Asset.Get)
	})

	// Git Smart HTTP routes
	r.Get("/info/refs", s.Handlers.Git.InfoRefs)
	r.Post("/git-upload-pack", s.Handlers.Git.UploadPack)
	r.Post("/git-receive-pack", s.Handlers.Git.ReceivePack)

	// Static files (frontend) - catch-all at the end
	r.Handle("/*", s.Handlers.Static)
}

// loggingMiddleware logs HTTP requests.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create wrapped response writer to capture status code
		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(wrw, r)

		// Log request
		s.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrw.Status(),
			"duration", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}
