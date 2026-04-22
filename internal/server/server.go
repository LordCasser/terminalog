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

// TLSConfig holds TLS configuration for HTTPS.
type TLSConfig struct {
	// Enabled indicates whether TLS is enabled.
	Enabled bool

	// CertFile is the path to the TLS certificate file.
	CertFile string

	// KeyFile is the path to the TLS private key file.
	KeyFile string
}

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

	// debug enables debug mode.
	// When true, static files are not embedded and CORS is enabled.
	debug bool

	// tls holds TLS configuration. When Enabled is true, the server serves HTTPS.
	tls TLSConfig

	// redirectServer is an optional HTTP server that redirects to HTTPS.
	// Only used when TLS is enabled.
	redirectServer *http.Server
}

// Handlers contains all HTTP handlers for the server.
type Handlers struct {
	Article   *handler.ArticleHandler
	Asset     *handler.AssetHandler
	Git       *handler.GitHandler
	Search    *handler.SearchHandler
	Tree      *handler.TreeHandler
	Static    *handler.StaticHandler
	Health    *handler.HealthHandler
	AboutMe   *handler.AboutMeHandler
	WebSocket *WebSocketHandler
	Config    *handler.ConfigHandler
}

// NewServer creates a new Server instance.
func NewServer(addr string, handlers *Handlers, logger *slog.Logger, embedFS embed.FS, debug bool, tls TLSConfig) *Server {
	// Initialize static handler with embedded files (if not in debug mode)
	if !debug {
		handlers.Static = handler.NewStaticHandler(embedFS)
	}

	s := &Server{
		addr:     addr,
		router:   chi.NewRouter(),
		logger:   logger,
		Handlers: handlers,
		debug:    debug,
		tls:      tls,
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
// When TLS is enabled, it starts the HTTPS server and optionally an HTTP redirect server.
func (s *Server) Start() error {
	if s.tls.Enabled {
		s.logger.Info("Server starting (HTTPS)", "addr", s.addr)
		return s.server.ListenAndServeTLS(s.tls.CertFile, s.tls.KeyFile)
	}
	s.logger.Info("Server starting (HTTP)", "addr", s.addr)
	return s.server.ListenAndServe()
}

// StartRedirect starts an HTTP redirect server that redirects all requests to HTTPS.
// This should be called in a goroutine when TLS is enabled.
// The redirect server listens on port 80.
func (s *Server) StartRedirect() error {
	if !s.tls.Enabled {
		return nil
	}

	redirectAddr := ":80"
	s.logger.Info("Starting HTTP redirect server", "addr", redirectAddr)

	s.redirectServer = &http.Server{
		Addr: redirectAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	return s.redirectServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server and the redirect server if running.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Server shutting down")

	// Shutdown redirect server if it exists
	if s.redirectServer != nil {
		if err := s.redirectServer.Shutdown(ctx); err != nil {
			s.logger.Warn("Redirect server shutdown error", "error", err)
		}
	}

	return s.server.Shutdown(ctx)
}

// setupRoutes configures all routes for the server.
func (s *Server) setupRoutes() {
	r := s.router

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(s.loggingMiddleware)
	r.Use(Gzip)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Enable CORS in debug mode
	if s.debug {
		r.Use(s.corsMiddleware)
	}

	// API routes (RESTful v1)
	// See docs/api-spec.md for complete API specification
	r.Route("/api/v1", func(r chi.Router) {
		// Health check routes (no auth required)
		// GET /api/v1/healthz, /api/v1/readyz, /api/v1/livez, /api/v1/status
		if s.Handlers.Health != nil {
			r.Get("/healthz", s.Handlers.Health.Healthz)
			r.Get("/readyz", s.Handlers.Health.Readyz)
			r.Get("/livez", s.Handlers.Health.Livez)
			r.Get("/status", s.Handlers.Health.Status)
		}

		// Articles API (RESTful path-based routing)
		// GET /api/v1/articles           - root directory listing (dirs + files)
		// GET /api/v1/articles/tech      - directory listing for tech/
		// GET /api/v1/articles/tech/go.md - article content
		// GET /api/v1/articles/tech/go.md/timeline - article timeline
		// GET /api/v1/articles/tech/go.md/version  - article version
		r.Get("/articles", s.Handlers.Article.ListRoot)
		r.Get("/articles/*", s.Handlers.Article.HandleRequest)

		// Search API (independent resource, not nested under articles)
		// GET /api/v1/search?q=xxx - search articles
		r.Get("/search", s.Handlers.Search.Search)

		// Tree API
		r.Get("/tree", s.Handlers.Tree.Get)

		// Assets API (images from Git repository)
		r.Get("/assets/*", s.Handlers.Asset.Get)

		// Special Pages API
		// GET /api/v1/special/aboutme - About Me page content
		if s.Handlers.AboutMe != nil {
			r.Get("/special/aboutme", s.Handlers.AboutMe.Get)
		}

		// Settings API (frontend configuration)
		// GET /api/v1/settings - returns frontend settings
		if s.Handlers.Config != nil {
			r.Get("/settings", s.Handlers.Config.Get)
		}

		// Resources API (frontend static resources like _next/static)
		// In production, these are embedded in the binary
		// In debug mode, frontend runs separately
		if !s.debug && s.Handlers.Static != nil {
			r.Get("/resources/*", s.Handlers.Static.ServeResources)
		}

		// Git Smart HTTP routes
		// Git clone URL: http://xxx/api/v1/git/
		// GET /api/v1/git/info/refs?service=git-upload-pack - refs advertisement
		// GET /api/v1/git/info/refs?service=git-receive-pack - refs advertisement (auth required)
		// POST /api/v1/git/git-upload-pack - packfile transfer for clone/fetch
		// POST /api/v1/git/git-receive-pack - packfile transfer for push (auth required)
		r.Get("/git/info/refs", s.Handlers.Git.InfoRefs)
		r.Post("/git/git-upload-pack", s.Handlers.Git.UploadPack)
		r.Post("/git/git-receive-pack", s.Handlers.Git.ReceivePack)
	})

	// WebSocket routes (v1.4) - keep at root level for simplicity
	if s.Handlers.WebSocket != nil {
		r.Get("/ws/terminal", s.Handlers.WebSocket.HandleTerminal)
	}

	// Git Smart HTTP routes at root level for standard Git clone URL
	// Git clone URL: http://xxx/ (standard Git HTTP)
	// GET /info/refs?service=git-upload-pack - refs advertisement
	// GET /info/refs?service=git-receive-pack - refs advertisement (auth required)
	// POST /git-upload-pack - packfile transfer for clone/fetch
	// POST /git-receive-pack - packfile transfer for push (auth required)
	r.Get("/info/refs", s.Handlers.Git.InfoRefs)
	r.Post("/git-upload-pack", s.Handlers.Git.UploadPack)
	r.Post("/git-receive-pack", s.Handlers.Git.ReceivePack)

	// Static files (frontend) - catch-all at the end
	// In debug mode, frontend runs separately, so we don't serve static files
	if !s.debug && s.Handlers.Static != nil {
		r.Handle("/*", s.Handlers.Static)
	}
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

// corsMiddleware enables CORS for development mode.
// This allows frontend dev server (e.g., Next.js on port 3000) to access backend API.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins in debug mode
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Process request
		next.ServeHTTP(w, r)
	})
}
