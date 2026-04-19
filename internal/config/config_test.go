package config

import (
	"path/filepath"
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
