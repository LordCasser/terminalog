package utils_test

import (
	"testing"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

func TestValidateContentPath(t *testing.T) {
	tests := []struct {
		name      string
		requested string
		want      string
		wantErr   error
		desc      string
	}{
		// ---- Valid paths ----
		{
			name:      "simple file",
			requested: "file.md",
			want:      "file.md",
			wantErr:   nil,
			desc:      "Direct child file should be valid",
		},
		{
			name:      "nested file",
			requested: "sub/nested/file.md",
			want:      "sub/nested/file.md",
			wantErr:   nil,
			desc:      "Deeply nested file within base should be valid",
		},
		{
			name:      "normalized path with ..",
			requested: "sub/../sub/nested/file.md",
			want:      "sub/nested/file.md",
			wantErr:   nil,
			desc:      "Path with .. that resolves inside base should be valid",
		},
		{
			name:      "filename containing dotdot",
			requested: "sub/nested/v1..2-changelog.md",
			want:      "sub/nested/v1..2-changelog.md",
			wantErr:   nil,
			desc:      "Legitimate filename containing '..' should NOT be rejected",
		},
		{
			name:      "filename containing .git as substring",
			requested: "sub/nested/my.git.config.md",
			want:      "sub/nested/my.git.config.md",
			wantErr:   nil,
			desc:      "Filename with .git substring should NOT be rejected (only exact '.git' segment)",
		},
		{
			name:      "empty path",
			requested: "",
			want:      "",
			wantErr:   nil,
			desc:      "Empty path should resolve to base directory",
		},

		// ---- Invalid paths ----
		{
			name:      "escape via ..",
			requested: "../../etc/passwd",
			wantErr:   model.ErrInvalidPath,
			desc:      "Directory traversal with .. should be rejected",
		},
		{
			name:      "absolute path outside base",
			requested: "/etc/passwd",
			wantErr:   model.ErrInvalidPath,
			desc:      "Absolute path outside base should be rejected",
		},
		{
			name:      ".git directory",
			requested: ".git/config",
			wantErr:   model.ErrInvalidPath,
			desc:      "Access to .git directory should be rejected",
		},
		{
			name:      "nested .git",
			requested: "sub/nested/.git/HEAD",
			wantErr:   model.ErrInvalidPath,
			desc:      "Access to nested .git should be rejected",
		},
		{
			name:      "only .git",
			requested: ".git",
			wantErr:   model.ErrInvalidPath,
			desc:      ".git directory itself should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.ValidateContentPath(tt.requested)
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
				if result != tt.want {
					t.Errorf("expected %q, got %q. %s", tt.want, result, tt.desc)
				}
			}
		})
	}
}
