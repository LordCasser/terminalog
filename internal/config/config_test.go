package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
