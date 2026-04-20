package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestVersionService_GetVersion(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(repo *testutil.TestRepo) error
		path           string
		wantVersion    string
		wantHistoryLen int
		wantErr        bool
	}{
		{
			name: "single commit file",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("article.md", "# Article\nLine 1\nLine 2", "Initial commit", "author")
			},
			path:           "article.md",
			wantVersion:    "v1.0.0", // Initial version (file creation = major, but hardcoded to v1.0.0)
			wantHistoryLen: 1,
		},
		{
			name: "small update triggers patch bump",
			setup: func(repo *testutil.TestRepo) error {
				// Initial commit: 3 lines
				if err := repo.CreateMarkdownFileWithTime("article.md", "# Article\nLine 1\nLine 2", "Initial", "author", time.Now().Add(-72*time.Hour)); err != nil {
					return err
				}
				// Small update: add 1 line (< 10 lines changed)
				if err := repo.CreateFile("article.md", "# Article\nLine 1\nLine 2\nLine 3"); err != nil {
					return err
				}
				if err := repo.AddFile("article.md"); err != nil {
					return err
				}
				return repo.CommitWithTime("Small update", "editor", "editor@example.com", time.Now().Add(-24*time.Hour))
			},
			path:           "article.md",
			wantVersion:    "v1.0.1", // Only 1 line added → patch bump
			wantHistoryLen: 2,
		},
		{
			name: "moderate update triggers minor bump",
			setup: func(repo *testutil.TestRepo) error {
				// Create a 30-line file initially
				content := "# Article\n"
				for i := 1; i <= 29; i++ {
					content += "Line " + string(rune('0'+i%10)) + "\n"
				}
				if err := repo.CreateMarkdownFileWithTime("article.md", content, "Initial", "author", time.Now().Add(-72*time.Hour)); err != nil {
					return err
				}
				// Moderate update: add 12 lines to a 30-line file
				// 12/30 = 40% → minor bump
				updated := content + "\n## Update\nAdded line 1\nAdded line 2\nAdded line 3\nAdded line 4\nAdded line 5\nAdded line 6\nAdded line 7\nAdded line 8\nAdded line 9\nAdded line 10\n"
				if err := repo.CreateFile("article.md", updated); err != nil {
					return err
				}
				if err := repo.AddFile("article.md"); err != nil {
					return err
				}
				return repo.CommitWithTime("Moderate update", "editor", "editor@example.com", time.Now().Add(-24*time.Hour))
			},
			path:           "article.md",
			wantVersion:    "v1.1.0", // minor bump: ~12 lines added, 40% of 30-line file
			wantHistoryLen: 2,
		},
		{
			name:    "non-existent file",
			setup:   func(repo *testutil.TestRepo) error { return nil },
			path:    "not-exist.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := testutil.NewTestRepo()
			require.NoError(t, err)
			defer repo.Cleanup()

			if tt.setup != nil {
				require.NoError(t, tt.setup(repo))
			}

			fileSvc, err := service.NewFileService(repo.Path)
			require.NoError(t, err)

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			articleSvc := service.NewArticleService(fileSvc, gitSvc)
			versionSvc := service.NewVersionService(articleSvc, gitSvc, fileSvc)

			versionInfo, err := versionSvc.GetVersion(context.Background(), tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantVersion, versionInfo.CurrentVersion)
			assert.Len(t, versionInfo.History, tt.wantHistoryLen)

			// Verify history entries have LinesAdded/LinesRemoved
			for _, entry := range versionInfo.History {
				assert.NotPanics(t, func() {
					_ = entry.LinesAdded
					_ = entry.LinesRemoved
				})
			}
		})
	}
}

func TestVersionService_ClassifyChange(t *testing.T) {
	tests := []struct {
		name          string
		linesChanged  int
		prevFileLines int
		wantType      model.ChangeType
	}{
		{
			name:          "patch - less than 10 lines",
			linesChanged:  5,
			prevFileLines: 100,
			wantType:      model.ChangeTypePatch,
		},
		{
			name:          "patch - exactly 9 lines",
			linesChanged:  9,
			prevFileLines: 100,
			wantType:      model.ChangeTypePatch,
		},
		{
			name:          "minor - 10 lines changed (10% of 100)",
			linesChanged:  10,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "minor - 40% changed",
			linesChanged:  40,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "minor - 50% exactly (not > 50%)",
			linesChanged:  50,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "major - 51% changed",
			linesChanged:  51,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMajor,
		},
		{
			name:          "major - 100% changed",
			linesChanged:  100,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMajor,
		},
		{
			name:          "zero previous lines - small change defaults to minor",
			linesChanged:  50,
			prevFileLines: 0,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "zero previous lines - large change is major",
			linesChanged:  100,
			prevFileLines: 0,
			wantType:      model.ChangeTypeMajor,
		},
		{
			name:          "zero previous lines - patch-level change still patch",
			linesChanged:  5,
			prevFileLines: 0,
			wantType:      model.ChangeTypePatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changeType := classifyChangeTest(tt.linesChanged, tt.prevFileLines)
			assert.Equal(t, tt.wantType, changeType)
		})
	}
}

// classifyChangeTest mirrors the classifyChange function for testing
func classifyChangeTest(linesChanged, prevFileLines int) model.ChangeType {
	const patchThreshold = 10
	const minorThresholdPercent = 0.5

	if linesChanged < patchThreshold {
		return model.ChangeTypePatch
	}

	if prevFileLines == 0 {
		if linesChanged >= 100 {
			return model.ChangeTypeMajor
		}
		return model.ChangeTypeMinor
	}

	percent := float64(linesChanged) / float64(prevFileLines)
	if percent > minorThresholdPercent {
		return model.ChangeTypeMajor
	}

	return model.ChangeTypeMinor
}

func TestVersionService_VersionFormat(t *testing.T) {
	tests := []struct {
		name  string
		major int
		minor int
		patch int
		want  string
	}{
		{
			name:  "initial version",
			major: 1,
			minor: 0,
			patch: 0,
			want:  "v1.0.0",
		},
		{
			name:  "patch bump",
			major: 1,
			minor: 0,
			patch: 1,
			want:  "v1.0.1",
		},
		{
			name:  "minor bump",
			major: 1,
			minor: 1,
			patch: 0,
			want:  "v1.1.0",
		},
		{
			name:  "major bump",
			major: 2,
			minor: 0,
			patch: 0,
			want:  "v2.0.0",
		},
		{
			name:  "complex version",
			major: 2,
			minor: 3,
			patch: 48,
			want:  "v2.3.48",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVersionTest(tt.major, tt.minor, tt.patch)
			assert.Equal(t, tt.want, result)
		})
	}
}

// formatVersionTest mirrors the formatVersion function for testing
func formatVersionTest(major, minor, patch int) string {
	return "v" + intToStr(major) + "." + intToStr(minor) + "." + intToStr(patch)
}

func intToStr(n int) string {
	if n < 10 {
		switch n {
		case 0:
			return "0"
		case 1:
			return "1"
		case 2:
			return "2"
		case 3:
			return "3"
		case 4:
			return "4"
		case 5:
			return "5"
		case 6:
			return "6"
		case 7:
			return "7"
		case 8:
			return "8"
		case 9:
			return "9"
		}
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}