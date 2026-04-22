// Terminalog - Terminal-style Blog System
// Main entry point for the HTTP server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terminalog/internal/config"
	"terminalog/internal/handler"
	"terminalog/internal/server"
	"terminalog/internal/service"
	"terminalog/pkg/embed"
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

	// Apply command-line overrides
	if flags.host != "" {
		cfg.Server.Host = flags.host
		logger.Info("Overriding server host from command line", "host", flags.host)
	}
	if flags.port != 0 {
		cfg.Server.Port = flags.port
		logger.Info("Overriding server port from command line", "port", flags.port)
	}
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
		Config:    handler.NewConfigHandler(cfg.GetOwner()),                            // v1.5
		WebSocket: server.NewWebSocketHandler(completionSvc, logger, cfg.Server.Debug), // v1.4
	}

	// Build TLS configuration
	tlsConfig := server.TLSConfig{
		Enabled:  cfg.Server.TLSEnabled,
		CertFile: cfg.Server.CertFile,
		KeyFile:  cfg.Server.KeyFile,
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
				if err.Error() != "http: Server closed" {
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

	// Console reminder for self-signed certificate
	if tlsConfig.Enabled {
		logger.Warn("TLS is enabled. If using a self-signed certificate, your browser will show a security warning.")
		logger.Warn("To generate a self-signed certificate: openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj \"/CN=localhost\"")
	}

	if err := srv.Start(); err != nil {
		if err.Error() != "http: Server closed" {
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

// printUsage prints usage information.
func printUsage() {
	fmt.Println("Terminalog - Terminal-style Blog System")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  terminalog [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  terminalog --config config.toml --log debug")
	fmt.Println("  terminalog --port 3000 --content ./my-blog")
	fmt.Println()
	fmt.Println("First run:")
	fmt.Println("  1. terminalog will create a default config.toml")
	fmt.Println("  2. Edit config.toml to set your content directory")
	fmt.Println("  3. Git repository will be auto-initialized if not exists")
	fmt.Println("  4. Restart terminalog")
}
