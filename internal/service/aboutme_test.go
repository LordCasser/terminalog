package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestAboutMeService_OnlyExposesCommittedContent(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testutil.TestRepo) error
		exists  bool
		content string
	}{
		{
			name: "committed",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateFile(service.AboutMeFilename, "# About Me"); err != nil {
					return err
				}
				return repo.Commit("Add about me", "author", "author@example.com")
			},
			exists:  true,
			content: "# About Me",
		},
		{
			name: "uncommitted",
			setup: func(repo *testutil.TestRepo) error {
				if err := repo.CreateMarkdownFile("article.md", "# Article", "Add", "author"); err != nil {
					return err
				}
				return repo.CreateFile(service.AboutMeFilename, "# Draft")
			},
		},
		{name: "missing", setup: func(*testutil.TestRepo) error { return nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := testutil.NewTestRepo()
			require.NoError(t, err)
			defer repo.Cleanup()
			require.NoError(t, tt.setup(repo))

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)
			svc := service.NewAboutMeService(gitSvc)

			content, exists, err := svc.Get(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.exists, exists)
			assert.Equal(t, tt.content, string(content))
		})
	}
}
