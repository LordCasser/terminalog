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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		logger.Info("Make sure blog.content_dir points to an existing Git repository")
		os.Exit(1)
	}

	logger.Info("Configuration loaded",
		"contentDir", cfg.Blog.ContentDir,
		"serverAddr", cfg.GetAddr(),
	)

	// Initialize services
	fileSvc, err := service.NewFileService(cfg.Blog.ContentDir)
	if err != nil {
		logger.Error("Failed to initialize file service", "error", err)
		os.Exit(1)
	}

	gitSvc, err := service.NewGitService(cfg.Blog.ContentDir)
	if err != nil {
		logger.Error("Failed to initialize git service", "error", err)
		os.Exit(1)
	}

	articleSvc := service.NewArticleService(fileSvc, gitSvc)
	authSvc := service.NewAuthService(cfg)
	assetSvc := service.NewAssetService(fileSvc)

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

	// Create handlers
	handlers := &server.Handlers{
		Article: handler.NewArticleHandler(articleSvc),
		Asset:   handler.NewAssetHandler(assetSvc),
		Git:     handler.NewGitHandler(gitSvc, authSvc),
		Search:  handler.NewSearchHandler(articleSvc),
		Tree:    handler.NewTreeHandler(articleSvc),
	}

	// Create HTTP server
	srv := server.NewServer(cfg.GetAddr(), handlers, logger, embed.StaticFS)

	// Setup graceful shutdown
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-shutdownCh
		logger.Info("Received shutdown signal", "signal", sig.String())

		// Give server time to finish pending requests
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Stop(shutdownCtx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}
	}()

	// Start server
	logger.Info("Server started", "addr", cfg.GetAddr())
	logger.Info("Access the blog at http://" + cfg.GetAddr())
	logger.Info("Git clone URL: http://" + cfg.GetAddr() + "/.git")

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
	host       string
	port       int
	configPath string
	logLevel   string
}

// parseFlags parses command-line flags.
func parseFlags() *Flags {
	f := &Flags{}

	flag.StringVar(&f.host, "host", "", "Server host (overrides config)")
	flag.IntVar(&f.port, "port", 0, "Server port (overrides config)")
	flag.StringVar(&f.configPath, "config", "config.toml", "Configuration file path")
	flag.StringVar(&f.logLevel, "log", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

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
	fmt.Println()
	fmt.Println("First run:")
	fmt.Println("  1. terminalog will create a default config.toml")
	fmt.Println("  2. Edit config.toml to set your content directory (Git repository)")
	fmt.Println("  3. Restart terminalog")
}
