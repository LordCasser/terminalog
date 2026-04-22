package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveContentDirRelativeToConfig(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join("configs", "config.toml")

	got, err := ResolveContentDir("./articles", configPath)
	if err != nil {
		t.Fatalf("ResolveContentDir returned error: %v", err)
	}

	want, err := filepath.Abs(filepath.Join("configs", "articles"))
	if err != nil {
		t.Fatalf("failed to build expected path: %v", err)
	}

	if got != want {
		t.Fatalf("ResolveContentDir() = %q, want %q", got, want)
	}
}

func TestResolveContentDirWithoutConfigPathUsesWorkingDirectory(t *testing.T) {
	t.Parallel()

	got, err := ResolveContentDir("./articles", "")
	if err != nil {
		t.Fatalf("ResolveContentDir returned error: %v", err)
	}

	want, err := filepath.Abs("./articles")
	if err != nil {
		t.Fatalf("failed to build expected path: %v", err)
	}

	if got != want {
		t.Fatalf("ResolveContentDir() = %q, want %q", got, want)
	}
}

func TestResolveContentDirPreservesAbsolutePath(t *testing.T) {
	t.Parallel()

	absolute, err := filepath.Abs(filepath.Join("fixtures", "blog"))
	if err != nil {
		t.Fatalf("failed to build absolute path: %v", err)
	}

	got, err := ResolveContentDir(absolute, filepath.Join("configs", "config.toml"))
	if err != nil {
		t.Fatalf("ResolveContentDir returned error: %v", err)
	}

	if got != absolute {
		t.Fatalf("ResolveContentDir() = %q, want %q", got, absolute)
	}
}

func TestValidateTLSSettings(t *testing.T) {
	t.Parallel()

	// Use the project root as a valid git repository for testing
	projectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("failed to get project root: %v", err)
	}

	// Create temporary cert files for testing
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")

	if err := os.WriteFile(certFile, []byte("test cert"), 0644); err != nil {
		t.Fatalf("failed to create test cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("test key"), 0600); err != nil {
		t.Fatalf("failed to create test key file: %v", err)
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "TLS disabled - no cert required",
			config: Config{
				Blog: BlogConfig{ContentDir: projectRoot},
				Server: ServerConfig{
					Host:       "0.0.0.0",
					Port:       8080,
					TLSEnabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "TLS enabled with valid cert files",
			config: Config{
				Blog: BlogConfig{ContentDir: projectRoot},
				Server: ServerConfig{
					Host:       "0.0.0.0",
					Port:       8443,
					TLSEnabled: true,
					CertFile:   certFile,
					KeyFile:    keyFile,
				},
			},
			wantErr: false,
		},
		{
			name: "TLS enabled without cert_file",
			config: Config{
				Blog: BlogConfig{ContentDir: projectRoot},
				Server: ServerConfig{
					Host:       "0.0.0.0",
					Port:       8443,
					TLSEnabled: true,
					CertFile:   "",
					KeyFile:    keyFile,
				},
			},
			wantErr: true,
			errMsg:  "cert_file is required",
		},
		{
			name: "TLS enabled without key_file",
			config: Config{
				Blog: BlogConfig{ContentDir: projectRoot},
				Server: ServerConfig{
					Host:       "0.0.0.0",
					Port:       8443,
					TLSEnabled: true,
					CertFile:   certFile,
					KeyFile:    "",
				},
			},
			wantErr: true,
			errMsg:  "key_file is required",
		},
		{
			name: "TLS enabled with non-existent cert_file",
			config: Config{
				Blog: BlogConfig{ContentDir: projectRoot},
				Server: ServerConfig{
					Host:       "0.0.0.0",
					Port:       8443,
					TLSEnabled: true,
					CertFile:   "/nonexistent/cert.pem",
					KeyFile:    keyFile,
				},
			},
			wantErr: true,
			errMsg:  "cert_file not found",
		},
		{
			name: "TLS enabled with non-existent key_file",
			config: Config{
				Blog: BlogConfig{ContentDir: projectRoot},
				Server: ServerConfig{
					Host:       "0.0.0.0",
					Port:       8443,
					TLSEnabled: true,
					CertFile:   certFile,
					KeyFile:    "/nonexistent/key.pem",
				},
			},
			wantErr: true,
			errMsg:  "key_file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
				} else if tt.errMsg != "" {
					if !strings.Contains(err.Error(), tt.errMsg) {
						t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestResolveDefaultPort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tlsEnabled bool
		port       int
		wantPort   int
	}{
		{"TLS disabled, port 0 → 8080", false, 0, 8080},
		{"TLS enabled, port 0 → 443", true, 0, 443},
		{"TLS disabled, explicit port → unchanged", false, 3000, 3000},
		{"TLS enabled, explicit port → unchanged", true, 8443, 8443},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{
				Server: ServerConfig{
					Port:       tt.port,
					TLSEnabled: tt.tlsEnabled,
				},
			}
			cfg.ResolveDefaultPort()
			assert.Equal(t, tt.wantPort, cfg.Server.Port)
		})
	}
}

func TestResolveHTTPRedirectAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		tlsEnabled     bool
		port           int
		explicitAddr   string
		wantRedirectAddr string
	}{
		{"TLS disabled → no redirect", false, 443, "", ""},
		{"TLS + port 443, no explicit → :80", true, 443, "", ":80"},
		{"TLS + port 8443, no explicit → empty", true, 8443, "", ""},
		{"TLS + port 443, explicit - → -", true, 443, "-", "-"},
		{"TLS + port 8443, explicit :8080 → :8080", true, 8443, ":8080", ":8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{
				Server: ServerConfig{
					Port:             tt.port,
					TLSEnabled:       tt.tlsEnabled,
					HTTPRedirectAddr: tt.explicitAddr,
				},
			}
			cfg.ResolveHTTPRedirectAddr()
			assert.Equal(t, tt.wantRedirectAddr, cfg.Server.HTTPRedirectAddr)
		})
	}
}

func TestResolveTLSSettingsAutoDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cert/key in a "resources" subdirectory
	resDir := filepath.Join(tmpDir, "resources")
	if err := os.MkdirAll(resDir, 0755); err != nil {
		t.Fatal(err)
	}

	certFile := filepath.Join(resDir, "https.crt")
	keyFile := filepath.Join(resDir, "https.key")
	if err := os.WriteFile(certFile, []byte("cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, []byte("key"), 0600); err != nil {
		t.Fatal(err)
	}

	// Override default paths to point to tmpDir
	origCertPaths := DefaultCertPaths
	origKeyPaths := DefaultKeyPaths
	defer func() {
		DefaultCertPaths = origCertPaths
		DefaultKeyPaths = origKeyPaths
	}()

	DefaultCertPaths = []string{certFile}
	DefaultKeyPaths = []string{keyFile}

	cfg := &Config{
		Server: ServerConfig{
			TLSEnabled: true,
			CertFile:   "", // Not explicitly set → auto-detect
			KeyFile:    "", // Not explicitly set → auto-detect
		},
	}

	resolvedCert, resolvedKey, err := cfg.ResolveTLSSettings()
	assert.NoError(t, err)
	assert.Equal(t, certFile, resolvedCert)
	assert.Equal(t, keyFile, resolvedKey)
}

func TestResolveTLSSettingsNoAutoDetectionWhenExplicit(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Server: ServerConfig{
			TLSEnabled: true,
			CertFile:   "/explicit/cert.pem",
			KeyFile:    "/explicit/key.pem",
		},
	}

	cert, key, err := cfg.ResolveTLSSettings()
	assert.NoError(t, err)
	assert.Equal(t, "/explicit/cert.pem", cert)
	assert.Equal(t, "/explicit/key.pem", key)
}

func TestResolveTLSSettingsFailsWhenNoCertFound(t *testing.T) {
	t.Parallel()

	origCertPaths := DefaultCertPaths
	defer func() { DefaultCertPaths = origCertPaths }()

	DefaultCertPaths = []string{"/nonexistent/path/cert.pem"}

	cfg := &Config{
		Server: ServerConfig{
			TLSEnabled: true,
			CertFile:   "",
			KeyFile:    "",
		},
	}

	_, _, err := cfg.ResolveTLSSettings()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no cert_file configured")
}

func TestFindFirstExisting(t *testing.T) {
	t.Parallel()

	tmpFile := filepath.Join(t.TempDir(), "found.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		paths  []string
		wantOK bool
	}{
		{"existing file", []string{tmpFile}, true},
		{"non-existing files", []string{"/no/such/file/a", "/no/such/file/b"}, false},
		{"empty list", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := findFirstExisting(tt.paths)
			if tt.wantOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
