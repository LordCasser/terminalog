// Terminalog - Terminal-style Blog System
// Main entry point for the HTTP server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"terminalog/internal/config"
	"terminalog/internal/handler"
	"terminalog/internal/server"
	"terminalog/internal/service"
	"terminalog/pkg/embed"
	"terminalog/pkg/utils"
)

// Version information (can be set at build time).
var (
	version   = "dev"
	buildDate = "unknown"
)

func main() {
	// Parse command-line flags
	flags := parseFlags()

	// Setup logger
	logger := setupLogger(flags.logLevel)
	slog.SetDefault(logger)

	// Print version info
	logger.Info("Terminalog starting",
		"version", version,
		"buildDate", buildDate,
	)

	// Load or create configuration
	cfg, created, err := config.LoadOrCreate(flags.configPath)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	if created {
		logger.Info("Created default configuration file", "path", flags.configPath)
		logger.Info("Please edit config.toml and set your content directory")
	}

	// Apply command-line overrides (host, content dir, debug)
	if flags.host != "" {
		cfg.Server.Host = flags.host
		logger.Info("Overriding server host from command line", "host", flags.host)
	}
	// Note: port override is handled after ResolveDefaultPort() in TLS section below
	if flags.contentDir != "" {
		cfg.Blog.ContentDir = flags.contentDir
		logger.Info("Overriding content directory from command line", "contentDir", flags.contentDir)
	}
	if flags.debug {
		cfg.Server.Debug = true
		logger.Info("Enabling debug mode from command line")
	}

	contentDir := cfg.Blog.ContentDir
	if flags.contentDir == "" {
		contentDir, err = config.ResolveContentDir(cfg.Blog.ContentDir, flags.configPath)
		if err != nil {
			logger.Error("Failed to resolve content directory from config", "error", err)
			os.Exit(1)
		}
	} else {
		contentDir, err = config.ResolveContentDir(flags.contentDir, "")
		if err != nil {
			logger.Error("Failed to resolve content directory from command line", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("Using content directory", "contentDir", contentDir)

	// Ensure Git repository exists (auto-initialize if needed)
	gitInitSvc := service.NewGitInitService()
	if err := gitInitSvc.EnsureGitRepo(contentDir, true); err != nil {
		logger.Error("Failed to initialize Git repository", "error", err)
		os.Exit(1)
	}

	// Get repository status
	repoStatus, err := gitInitSvc.GetRepoStatus(contentDir)
	if err == nil {
		logger.Info("Git repository status",
			"branch", repoStatus.CurrentBranch,
			"commits", repoStatus.CommitCount,
		)
	}

	// Initialize services
	fileSvc, err := service.NewFileService(contentDir)
	if err != nil {
		logger.Error("Failed to initialize file service", "error", err)
		os.Exit(1)
	}

	gitSvc, err := service.NewGitService(contentDir)
	if err != nil {
		logger.Error("Failed to initialize git service", "error", err)
		os.Exit(1)
	}

	logger.Info("Git service initialized", "repoPath", contentDir)

	articleSvc := service.NewArticleService(fileSvc, gitSvc)
	authSvc := service.NewAuthService(cfg)
	assetSvc := service.NewAssetService(fileSvc)
	versionSvc := service.NewVersionService(articleSvc, gitSvc, fileSvc)       // v1.2
	completionSvc := service.NewCompletionService(articleSvc, fileSvc, gitSvc) // v1.4 - for WebSocket path completion

	// Handle default user generation if no users configured
	if !cfg.HasUsers() {
		defaultUser, password, err := authSvc.GenerateDefaultUser()
		if err != nil {
			logger.Error("Failed to generate default user", "error", err)
			os.Exit(1)
		}

		cfg.AddUser(*defaultUser)
		if err := cfg.Save(flags.configPath); err != nil {
			logger.Error("Failed to save configuration", "error", err)
			os.Exit(1)
		}

		logger.Info("Generated default user for Git push authentication",
			"username", defaultUser.Username,
			"password", password,
		)
		logger.Warn("Please change the default password in config.toml")
	}

	// Create health handler
	healthHandler := handler.NewHealthHandler(gitSvc, articleSvc.GetCacheStats)

	// Create handlers
	handlers := &server.Handlers{
		Article:   handler.NewArticleHandler(articleSvc, versionSvc, fileSvc),
		Asset:     handler.NewAssetHandler(assetSvc),
		Git:       handler.NewGitHandler(gitSvc, authSvc, articleSvc.InvalidateCache),
		Search:    handler.NewSearchHandler(articleSvc),
		Tree:      handler.NewTreeHandler(articleSvc),
		Health:    healthHandler,
		AboutMe:   handler.NewAboutMeHandler(fileSvc),                                  // v1.2
		Config: handler.NewConfigHandler(cfg.GetOwner(), handler.FilingInfo{
			ICPFiling:       cfg.Site.ICPFiling,
			ICPFilingURL:    cfg.Site.ICPFilingURL,
			PoliceFiling:    cfg.Site.PoliceFiling,
			PoliceFilingURL: cfg.Site.PoliceFilingURL,
		}),
		WebSocket: server.NewWebSocketHandler(completionSvc, logger, cfg.Server.Debug), // v1.4
	}

	// Resolve TLS configuration with auto-detection and smart defaults
	// 1. Resolve default port (443 for TLS, 8080 for HTTP) when port is 0
	cfg.ResolveDefaultPort()

	// Apply command-line port override after default resolution
	if flags.port != 0 {
		cfg.Server.Port = flags.port
		logger.Info("Overriding server port from command line", "port", flags.port)
	}

	// 2. Auto-detect cert/key from default paths when not explicitly configured
	if cfg.Server.TLSEnabled {
		certFile, keyFile, err := cfg.ResolveTLSSettings()
		if err != nil {
			// If auto-detection failed and AutoCert is enabled, generate self-signed cert
			if cfg.Server.AutoCert {
				logger.Info("Auto-generating self-signed TLS certificate for development")
				certPath := filepath.Join(config.DefaultAutoCertDir, config.DefaultAutoCertName)
				keyPath := filepath.Join(config.DefaultAutoCertDir, config.DefaultAutoKeyName)
				if err := utils.GenerateSelfSignedCert(certPath, keyPath, "localhost"); err != nil {
					logger.Error("Failed to auto-generate TLS certificate", "error", err)
					os.Exit(1)
				}
				cfg.Server.CertFile = certPath
				cfg.Server.KeyFile = keyPath
				certFile = certPath
				keyFile = keyPath
				logger.Info("Self-signed certificate generated", "cert", certPath, "key", keyPath)
			} else {
				logger.Error("TLS configuration error", "error", err)
				logger.Info("Tip: Set auto_cert = true in [server] section to auto-generate a development certificate")
				logger.Info("Tip: Or place your certificate files in the resources/ directory:")
				logger.Info("       cert_file → resources/https.crt  (or .pem)")
				logger.Info("       key_file  → resources/https.key")
				logger.Info("     For cloud provider certificates (Tencent/Aliyun Nginx type):")
				logger.Info("       cp your_domain_bundle.crt resources/https.crt")
				logger.Info("       cp your_domain.key         resources/https.key")
				os.Exit(1)
			}
		} else {
			logger.Info("TLS certificate auto-detected", "cert", certFile, "key", keyFile)
		}
	}

	// 3. Resolve HTTP redirect address (auto-enable :80 redirect on standard HTTPS port)
	cfg.ResolveHTTPRedirectAddr()

	// Build TLS configuration
	autoCert := cfg.Server.AutoCert && cfg.Server.TLSEnabled
	tlsConfig := server.TLSConfig{
		Enabled:          cfg.Server.TLSEnabled,
		CertFile:         cfg.Server.CertFile,
		KeyFile:          cfg.Server.KeyFile,
		HTTPRedirectAddr: cfg.Server.HTTPRedirectAddr,
		HSTS:             cfg.Server.TLSEnabled, // Enable HSTS when TLS is on
		AutoCert:         autoCert,
	}

	// Create HTTP server
	srv := server.NewServer(cfg.GetAddr(), handlers, logger, embed.StaticFS, cfg.Server.Debug, tlsConfig)

	// Mark server as ready
	healthHandler.SetReady()

	// Setup graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-shutdownCh
		logger.Info("Received shutdown signal", "signal", sig.String())

		// Mark server as not ready
		healthHandler.SetNotReady()

		// Give server time to finish pending requests
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
	}()

	// Start HTTP redirect server if TLS is enabled
	if tlsConfig.Enabled {
		go func() {
			if err := srv.StartRedirect(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					logger.Warn("Redirect server error", "error", err)
				}
			}
		}()
	}

	// Start server
	protocol := "http"
	if tlsConfig.Enabled {
		protocol = "https"
	}

	logger.Info("Server started", "addr", cfg.GetAddr(), "tls", tlsConfig.Enabled)
	logger.Info("Access the blog at "+protocol+"://"+cfg.GetAddr())
	logger.Info("Git clone URL: "+protocol+"://"+cfg.GetAddr()+"/api/v1/git/")
	logger.Info("Health check: "+protocol+"://"+cfg.GetAddr()+"/api/v1/healthz")

	// Console reminder for TLS certificate
	if tlsConfig.Enabled {
		if tlsConfig.AutoCert {
			logger.Warn("Using AUTO-GENERATED self-signed certificate — for development only!")
			logger.Warn("Browsers will show security warnings. Do NOT use in production.")
		} else {
			logger.Info("TLS enabled with certificate", "cert", tlsConfig.CertFile)
			logger.Info("HSTS header enabled (Strict-Transport-Security)")
		}
		if cfg.Server.Port == 443 && cfg.Server.HTTPRedirectAddr != "-" {
			logger.Info("HTTP→HTTPS redirect enabled", "redirect_addr", cfg.Server.HTTPRedirectAddr, "status", "307 Temporary Redirect")
		}
	}

	if err := srv.Start(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("Server stopped")
}

// Flags holds command-line flags.
type Flags struct {
	host        string
	port        int
	configPath  string
	logLevel    string
	contentDir  string
	showVersion bool
	debug       bool
}

// parseFlags parses command-line flags.
func parseFlags() *Flags {
	f := &Flags{}

	flag.StringVar(&f.host, "host", "", "Server host (overrides config)")
	flag.IntVar(&f.port, "port", 0, "Server port (overrides config)")
	flag.StringVar(&f.configPath, "config", "config.toml", "Configuration file path")
	flag.StringVar(&f.logLevel, "log", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&f.contentDir, "content", "", "Content directory path (overrides config)")
	flag.BoolVar(&f.showVersion, "version", false, "Show version information")
	flag.BoolVar(&f.debug, "debug", false, "Enable debug mode (frontend dev server, CORS enabled)")

	flag.Parse()

	// Handle version flag
	if f.showVersion {
		fmt.Printf("Terminalog %s (built %s)\n", version, buildDate)
		os.Exit(0)
	}

	return f
}

// setupLogger creates a structured logger with the given level.
func setupLogger(level string) *slog.Logger {
	var lvl slog.Level

	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	// Use text handler for better readability
	handler := slog.NewTextHandler(os.Stdout, opts)

	return slog.New(handler)
}
