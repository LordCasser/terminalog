package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()
	// Create nested subdir for testing
	subDir := filepath.Join(tmpDir, "sub", "nested")
	os.MkdirAll(subDir, 0755)

	tests := []struct {
		name      string
		baseDir   string
		requested string
		wantErr   error
		desc      string
	}{
		// ---- Valid paths ----
		{
			name:      "simple file",
			baseDir:   tmpDir,
			requested: "file.md",
			wantErr:   nil,
			desc:      "Direct child file should be valid",
		},
		{
			name:      "nested file",
			baseDir:   tmpDir,
			requested: "sub/nested/file.md",
			wantErr:   nil,
			desc:      "Deeply nested file within base should be valid",
		},
		{
			name:      "normalized path with ..",
			baseDir:   tmpDir,
			requested: "sub/../sub/nested/file.md",
			wantErr:   nil,
			desc:      "Path with .. that resolves inside base should be valid",
		},
		{
			name:      "filename containing dotdot",
			baseDir:   tmpDir,
			requested: "sub/nested/v1..2-changelog.md",
			wantErr:   nil,
			desc:      "Legitimate filename containing '..' should NOT be rejected",
		},
		{
			name:      "filename containing .git as substring",
			baseDir:   tmpDir,
			requested: "sub/nested/my.git.config.md",
			wantErr:   nil,
			desc:      "Filename with .git substring should NOT be rejected (only exact '.git' segment)",
		},
		{
			name:      "empty path",
			baseDir:   tmpDir,
			requested: "",
			wantErr:   nil,
			desc:      "Empty path should resolve to base directory",
		},

		// ---- Invalid paths ----
		{
			name:      "escape via ..",
			baseDir:   tmpDir,
			requested: "../../etc/passwd",
			wantErr:   model.ErrInvalidPath,
			desc:      "Directory traversal with .. should be rejected",
		},
		{
			name:      "absolute path outside base",
			baseDir:   tmpDir,
			requested: "/etc/passwd",
			wantErr:   model.ErrInvalidPath,
			desc:      "Absolute path outside base should be rejected",
		},
		{
			name:      ".git directory",
			baseDir:   tmpDir,
			requested: ".git/config",
			wantErr:   model.ErrInvalidPath,
			desc:      "Access to .git directory should be rejected",
		},
		{
			name:      "nested .git",
			baseDir:   tmpDir,
			requested: "sub/nested/.git/HEAD",
			wantErr:   model.ErrInvalidPath,
			desc:      "Access to nested .git should be rejected",
		},
		{
			name:      "only .git",
			baseDir:   tmpDir,
			requested: ".git",
			wantErr:   model.ErrInvalidPath,
			desc:      ".git directory itself should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ValidatePath(tt.baseDir, tt.requested)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, got nil (result=%q). %s", result, tt.desc)
				} else if err != tt.wantErr {
					t.Errorf("expected error %v, got %v. %s", tt.wantErr, err, tt.desc)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v. %s", err, tt.desc)
				}
				if !filepath.IsAbs(result) {
					t.Errorf("expected absolute path, got %q. %s", result, tt.desc)
				}
			}
		})
	}
}
