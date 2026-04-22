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

	// Site configuration section (metadata, filing, etc.).
	Site SiteConfig `toml:"site"`

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

// SiteConfig contains site-level metadata and regulatory information.
type SiteConfig struct {
	// ICPFiling is the ICP filing number (ICP备案号) required for websites in mainland China.
	// Example: "京ICP备12345678号-1"
	// When set, the filing info is embedded in the page for crawler/program detection.
	ICPFiling string `toml:"icp_filing"`

	// ICPFilingURL is the URL for the ICP filing record verification.
	// Defaults to "https://beian.miit.gov.cn/" if not set when icp_filing is configured.
	// Example: "https://beian.miit.gov.cn/#/Integrated/recordQuery/京ICP备12345678号-1"
	ICPFilingURL string `toml:"icp_filing_url"`

	// PoliceFiling is the public security filing number (公安备案号).
	// Example: "京公网安备11010502012345号"
	PoliceFiling string `toml:"police_filing"`

	// PoliceFilingURL is the URL for the public security filing record.
	// Typically points to the MPS filing verification page.
	// Example: "https://beian.mps.gov.cn/query/verifyQuery/11010502012345"
	PoliceFilingURL string `toml:"police_filing_url"`
}

// ServerConfig contains server-related settings.
type ServerConfig struct {
	// Host is the server host address.
	Host string `toml:"host"`

	// Port is the server port.
	// When TLS is enabled and Port is 0, defaults to 443.
	// When TLS is disabled and Port is 0, defaults to 8080.
	Port int `toml:"port"`

	// Debug enables debug mode for development.
	// When true, frontend static files are not embedded, allowing separate frontend dev server.
	// CORS is enabled to allow cross-origin requests from frontend dev server.
	Debug bool `toml:"debug"`

	// TLSEnabled enables HTTPS/TLS.
	// When true, the server serves HTTPS and redirects HTTP to HTTPS.
	// If CertFile/KeyFile are not set, auto-detection searches default paths.
	TLSEnabled bool `toml:"tls_enabled"`

	// CertFile is the path to the TLS certificate file (PEM format).
	// When empty and TLSEnabled is true, auto-detected from default paths:
	//   - resources/https.crt
	//   - resources/cert.pem
	//   - cert.pem
	CertFile string `toml:"cert_file"`

	// KeyFile is the path to the TLS private key file (PEM format).
	// When empty and TLSEnabled is true, auto-detected from default paths:
	//   - resources/https.key
	//   - resources/key.pem
	//   - key.pem
	KeyFile string `toml:"key_file"`

	// HTTPRedirectAddr is the address for HTTP-to-HTTPS redirect server.
	// When empty and TLS is enabled on standard port 443, defaults to ":80".
	// When empty and TLS is on non-standard port, redirect server is not started.
	// Set to "-" to explicitly disable the redirect server.
	HTTPRedirectAddr string `toml:"http_redirect_addr"`

	// AutoCert enables auto-generation of self-signed certificates for development.
	// When true and TLSEnabled is true but no cert files are found,
	// a self-signed certificate will be generated to the default cert path.
	// This should ONLY be used for development/testing, never in production.
	AutoCert bool `toml:"auto_cert"`
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

// DefaultCertPaths lists default certificate file paths for auto-detection.
// Searched in order; first existing file is used.
// Covers common naming conventions from cloud providers (Tencent, Aliyun, etc.)
// and generic names (https, server, cert).
var DefaultCertPaths = []string{
	"resources/https.crt",
	"resources/https.pem",
	"resources/server.crt",
	"resources/server.pem",
	"resources/cert.pem",
	"cert.pem",
}

// DefaultKeyPaths lists default private key file paths for auto-detection.
// Searched in order; first existing file is used.
var DefaultKeyPaths = []string{
	"resources/https.key",
	"resources/server.key",
	"resources/key.pem",
	"key.pem",
}

// DefaultAutoCertDir is the directory where auto-generated certificates are saved.
const DefaultAutoCertDir = "resources"

// DefaultAutoCertName is the filename for auto-generated certificate.
const DefaultAutoCertName = "https.crt"

// DefaultAutoKeyName is the filename for auto-generated private key.
const DefaultAutoKeyName = "https.key"

// ResolveTLSSettings resolves TLS configuration with auto-detection and smart defaults.
// It performs the following:
//  1. Auto-detects cert/key from default paths when not explicitly configured.
//  2. Adjusts the default port (443 for TLS, 8080 for HTTP) when port is 0.
//  3. Determines the HTTP redirect address based on the resolved port.
//
// Returns the resolved TLS config and any resolution errors.
func (c *Config) ResolveTLSSettings() (certFile, keyFile string, err error) {
	if !c.Server.TLSEnabled {
		return "", "", nil
	}

	// Auto-detect cert file if not explicitly set
	certFile = c.Server.CertFile
	if certFile == "" {
		certFile, err = findFirstExisting(DefaultCertPaths)
		if err != nil {
			return "", "", fmt.Errorf("tls_enabled is true but no cert_file configured and none found in default paths: %w", err)
		}
		c.Server.CertFile = certFile
	}

	// Auto-detect key file if not explicitly set
	keyFile = c.Server.KeyFile
	if keyFile == "" {
		keyFile, err = findFirstExisting(DefaultKeyPaths)
		if err != nil {
			return "", "", fmt.Errorf("tls_enabled is true but no key_file configured and none found in default paths: %w", err)
		}
		c.Server.KeyFile = keyFile
	}

	return certFile, keyFile, nil
}

// ResolveDefaultPort adjusts the default server port based on TLS status.
// If the current port is 0, it sets 443 for TLS or 8080 for HTTP.
func (c *Config) ResolveDefaultPort() {
	if c.Server.Port == 0 {
		if c.Server.TLSEnabled {
			c.Server.Port = 443
		} else {
			c.Server.Port = 8080
		}
	}
}

// ResolveHTTPRedirectAddr determines the HTTP redirect address.
// When TLS is enabled on standard port 443 and no explicit redirect addr is set,
// it defaults to ":80". For non-standard TLS ports, the redirect server
// is disabled unless explicitly configured.
func (c *Config) ResolveHTTPRedirectAddr() {
	if !c.Server.TLSEnabled {
		return
	}

	if c.Server.HTTPRedirectAddr != "" {
		return // Already explicitly configured
	}

	// Standard HTTPS port → auto-enable redirect on :80
	if c.Server.Port == 443 {
		c.Server.HTTPRedirectAddr = ":80"
	}
	// Non-standard port → no redirect by default (user must opt-in)
}

// GetTLSCertAndKey returns the resolved TLS certificate and key file paths.
// This should be called after ResolveTLSSettings.
func (c *Config) GetTLSCertAndKey() (certFile, keyFile string) {
	return c.Server.CertFile, c.Server.KeyFile
}

// findFirstExisting returns the first path that exists from the given list.
func findFirstExisting(paths []string) (string, error) {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, err := filepath.Abs(p)
			if err != nil {
				return p, nil
			}
			return abs, nil
		}
	}
	return "", fmt.Errorf("no file found in paths: %v", paths)
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
		Site: SiteConfig{
			ICPFiling:       "",
			ICPFilingURL:    "",
			PoliceFiling:    "",
			PoliceFilingURL: "",
		},
		Server: ServerConfig{
			Host:       "0.0.0.0",
			Port:       0, // Will be resolved: 443 for TLS, 8080 for HTTP
			Debug:      false,
			TLSEnabled: false,
			AutoCert:   false,
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

	// Validate TLS settings (after auto-detection, cert/key should be resolved)
	if c.Server.TLSEnabled {
		if c.Server.CertFile == "" {
			return fmt.Errorf("server.cert_file is required when tls_enabled is true (auto-detection found no default cert)")
		}
		if c.Server.KeyFile == "" {
			return fmt.Errorf("server.key_file is required when tls_enabled is true (auto-detection found no default key)")
		}
		if _, err := os.Stat(c.Server.CertFile); err != nil {
			return fmt.Errorf("server.cert_file not found: %s: %w", c.Server.CertFile, err)
		}
		if _, err := os.Stat(c.Server.KeyFile); err != nil {
			return fmt.Errorf("server.key_file not found: %s: %w", c.Server.KeyFile, err)
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

// GetSiteFiling returns the ICP filing number.
func (c *Config) GetSiteFiling() string {
	return c.Site.ICPFiling
}

// GetPoliceFiling returns the public security filing number and URL.
func (c *Config) GetPoliceFiling() (number, url string) {
	return c.Site.PoliceFiling, c.Site.PoliceFilingURL
}

// HasFilingInfo returns true if any filing information is configured.
func (c *Config) HasFilingInfo() bool {
	return c.Site.ICPFiling != "" || c.Site.PoliceFiling != ""
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
