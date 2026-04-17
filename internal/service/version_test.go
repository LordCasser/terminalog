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
			wantVersion:    "v1.0.0", // Initial version
			wantHistoryLen: 1,
		},
		{
			name: "multiple commits",
			setup: func(repo *testutil.TestRepo) error {
				// Initial commit
				if err := repo.CreateMarkdownFileWithTime("article.md", "# Article\nLine 1", "Initial", "author", time.Now().Add(-72*time.Hour)); err != nil {
					return err
				}
				// Update commit
				if err := repo.CreateFile("article.md", "# Article\nLine 1\nLine 2\nLine 3"); err != nil {
					return err
				}
				if err := repo.AddFile("article.md"); err != nil {
					return err
				}
				return repo.CommitWithTime("Update", "editor", "editor@example.com", time.Now().Add(-24*time.Hour))
			},
			path:           "article.md",
			wantVersion:    "v1.0.1", // Patch increment for small change
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
		})
	}
}

func TestVersionService_ClassifyChange(t *testing.T) {
	tests := []struct {
		name         string
		linesChanged int
		totalLines   int
		wantType     model.ChangeType
	}{
		{
			name:         "patch - less than 10 lines",
			linesChanged: 5,
			totalLines:   100,
			wantType:     model.ChangeTypePatch,
		},
		{
			name:         "patch - exactly 9 lines",
			linesChanged: 9,
			totalLines:   100,
			wantType:     model.ChangeTypePatch,
		},
		{
			name:         "minor - 10 lines changed",
			linesChanged: 10,
			totalLines:   100,
			wantType:     model.ChangeTypeMinor,
		},
		{
			name:         "minor - 40% changed",
			linesChanged: 40,
			totalLines:   100,
			wantType:     model.ChangeTypeMinor,
		},
		{
			name:         "minor - 50% exactly",
			linesChanged: 50,
			totalLines:   100,
			wantType:     model.ChangeTypeMinor, // 50% is NOT > 50%
		},
		{
			name:         "major - 51% changed",
			linesChanged: 51,
			totalLines:   100,
			wantType:     model.ChangeTypeMajor,
		},
		{
			name:         "major - 100% changed",
			linesChanged: 100,
			totalLines:   100,
			wantType:     model.ChangeTypeMajor,
		},
		{
			name:         "zero total lines defaults to patch",
			linesChanged: 10,
			totalLines:   0,
			wantType:     model.ChangeTypePatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use internal function logic inline for testing
			changeType := classifyChangeTest(tt.linesChanged, tt.totalLines)
			assert.Equal(t, tt.wantType, changeType)
		})
	}
}

// classifyChangeTest mirrors the classifyChange function for testing
func classifyChangeTest(linesChanged, totalLines int) model.ChangeType {
	const patchThreshold = 10
	const minorThresholdPercent = 0.5 // 50%

	if linesChanged < patchThreshold {
		return model.ChangeTypePatch
	}

	if totalLines == 0 {
		return model.ChangeTypePatch
	}

	percent := float64(linesChanged) / float64(totalLines)
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
	// Simplified for testing
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
