// Package config handles application configuration loading and management.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"terminalog/internal/model"
)

// Config represents the application configuration.
type Config struct {
	// Blog configuration section.
	Blog BlogConfig `toml:"blog"`

	// Server configuration section.
	Server ServerConfig `toml:"server"`

	// Auth configuration section.
	Auth AuthConfig `toml:"auth"`
}

// BlogConfig contains blog-related settings.
type BlogConfig struct {
	// ContentDir is the path to the Git repository containing blog content.
	ContentDir string `toml:"content_dir"`

	// Owner is the blog owner name displayed in navbar (e.g., ~/lordcasser).
	// Default value is "lordcasser".
	Owner string `toml:"owner"`
}

// ServerConfig contains server-related settings.
type ServerConfig struct {
	// Host is the server host address.
	Host string `toml:"host"`

	// Port is the server port.
	Port int `toml:"port"`

	// Debug enables debug mode for development.
	// When true, frontend static files are not embedded, allowing separate frontend dev server.
	// CORS is enabled to allow cross-origin requests from frontend dev server.
	Debug bool `toml:"debug"`

	// TLSEnabled enables HTTPS/TLS.
	// When true, the server serves HTTPS and redirects HTTP to HTTPS.
	TLSEnabled bool `toml:"tls_enabled"`

	// CertFile is the path to the TLS certificate file (PEM format).
	// Required when TLSEnabled is true.
	CertFile string `toml:"cert_file"`

	// KeyFile is the path to the TLS private key file (PEM format).
	// Required when TLSEnabled is true.
	KeyFile string `toml:"key_file"`
}

// AuthConfig contains authentication settings.
type AuthConfig struct {
	// Users is the list of authenticated users.
	Users []UserConfig `toml:"users"`
}

// UserConfig represents a user in the configuration file.
type UserConfig struct {
	// Username is the user's login name.
	Username string `toml:"username"`

	// Password can be either a plain text password (will be hashed on first run)
	// or a bcrypt hashed password (identified by starting with $2a$ or $2b$).
	Password string `toml:"password"`
}

// Load reads and parses the configuration file from the given path.
// If the file does not exist, it returns an error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// ResolveContentDir resolves a content directory path to an absolute path.
// Relative paths from the config file are resolved against the config file directory.
func ResolveContentDir(contentDir, configPath string) (string, error) {
	if contentDir == "" {
		return "", fmt.Errorf("blog.content_dir is required")
	}

	if filepath.IsAbs(contentDir) {
		return filepath.Clean(contentDir), nil
	}

	if configPath == "" {
		return filepath.Abs(contentDir)
	}

	configAbsPath, err := filepath.Abs(configPath)
	if err != nil {
		return "", fmt.Errorf("invalid config path: %w", err)
	}

	baseDir := filepath.Dir(configAbsPath)
	return filepath.Abs(filepath.Join(baseDir, contentDir))
}

// LoadOrCreate reads the configuration file, or creates a default one if it doesn't exist.
func LoadOrCreate(path string) (*Config, bool, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, false, nil
	}

	// File doesn't exist, create default
	if errors.Is(err, os.ErrNotExist) {
		cfg = Default()
		if saveErr := cfg.Save(path); saveErr != nil {
			return nil, false, fmt.Errorf("failed to create default config: %w", saveErr)
		}
		return cfg, true, nil
	}

	return nil, false, err
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Blog: BlogConfig{
			ContentDir: "./content",
			Owner:      "lordcasser",
		},
		Server: ServerConfig{
			Host:       "0.0.0.0",
			Port:       8080,
			Debug:      false, // Debug mode disabled by default
			TLSEnabled: false, // TLS disabled by default
		},
		Auth: AuthConfig{
			Users: []UserConfig{},
		},
	}
}

// Save writes the configuration to the given path.
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Check content directory
	if c.Blog.ContentDir == "" {
		return fmt.Errorf("blog.content_dir is required")
	}

	// Validate content directory exists and is a Git repository
	absPath, err := filepath.Abs(c.Blog.ContentDir)
	if err != nil {
		return fmt.Errorf("invalid content_dir path: %w", err)
	}

	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("content_dir is not a git repository: %s", absPath)
	}

	// Validate server settings
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Validate TLS settings
	if c.Server.TLSEnabled {
		if c.Server.CertFile == "" {
			return fmt.Errorf("server.cert_file is required when tls_enabled is true")
		}
		if c.Server.KeyFile == "" {
			return fmt.Errorf("server.key_file is required when tls_enabled is true")
		}
		if _, err := os.Stat(c.Server.CertFile); err != nil {
			return fmt.Errorf("server.cert_file not found: %w", err)
		}
		if _, err := os.Stat(c.Server.KeyFile); err != nil {
			return fmt.Errorf("server.key_file not found: %w", err)
		}
	}

	return nil
}

// GetContentDir returns the absolute path to the content directory.
func (c *Config) GetContentDir() (string, error) {
	return filepath.Abs(c.Blog.ContentDir)
}

// GetOwner returns the blog owner name.
// If not configured, returns default value "lordcasser".
func (c *Config) GetOwner() string {
	if c.Blog.Owner == "" {
		return "lordcasser"
	}
	return c.Blog.Owner
}

// GetAddr returns the server address in host:port format.
func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetUsers returns the list of configured users.
func (c *Config) GetUsers() []UserConfig {
	return c.Auth.Users
}

// AddUser adds a new user to the configuration.
func (c *Config) AddUser(user UserConfig) {
	c.Auth.Users = append(c.Auth.Users, user)
}

// HasUsers returns true if there are configured users.
func (c *Config) HasUsers() bool {
	return len(c.Auth.Users) > 0
}

// ToModelUser converts UserConfig to model.User.
func (u *UserConfig) ToModelUser() model.User {
	return model.User{
		Username:     u.Username,
		PasswordHash: u.Password,
	}
}
