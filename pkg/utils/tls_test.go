package utils

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "test.crt")
	keyPath := filepath.Join(tmpDir, "test.key")

	err := GenerateSelfSignedCert(certPath, keyPath, "localhost")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert() error: %v", err)
	}

	// Verify cert file exists and is valid PEM
	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("failed to read cert file: %v", err)
	}

	certBlock, _ := pem.Decode(certData)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		t.Fatalf("cert file is not a valid PEM certificate")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	if cert.Subject.CommonName != "localhost" {
		t.Errorf("certificate CN = %q, want %q", cert.Subject.CommonName, "localhost")
	}

	// Verify key file exists and is valid PEM
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}

	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil || keyBlock.Type != "EC PRIVATE KEY" {
		t.Fatalf("key file is not a valid EC private key PEM")
	}
}

func TestGenerateSelfSignedCertDefaultHost(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	err := GenerateSelfSignedCert(certPath, keyPath, "")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert() with empty host error: %v", err)
	}

	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("failed to read cert file: %v", err)
	}

	certBlock, _ := pem.Decode(certData)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	if cert.Subject.CommonName != "localhost" {
		t.Errorf("certificate CN = %q, want %q (default)", cert.Subject.CommonName, "localhost")
	}
}

func TestGenerateSelfSignedCertCreatesDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "subdir", "deep", "cert.pem")
	keyPath := filepath.Join(tmpDir, "subdir", "deep", "key.pem")

	err := GenerateSelfSignedCert(certPath, keyPath, "example.com")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert() error: %v", err)
	}

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Error("cert file was not created in nested directory")
	}
}

func TestGenerateSelfSignedCertRefusesOverwrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Create first cert
	if err := GenerateSelfSignedCert(certPath, keyPath, "first.local"); err != nil {
		t.Fatalf("first GenerateSelfSignedCert() error: %v", err)
	}

	// Attempt to overwrite should fail (O_EXCL)
	err := GenerateSelfSignedCert(certPath, keyPath, "second.local")
	if err == nil {
		t.Error("expected error when overwriting existing cert, got nil")
	}
}
