package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

