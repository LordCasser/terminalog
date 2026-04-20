package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestCompletionService_ChineseMatching(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(repo *testutil.TestRepo) error
		dir       string
		prefix    string
		wantItems []string // expected items (order may vary)
		wantMin   int      // minimum number of matches (if wantItems is nil)
	}{
		// --- Global search (dir=""): Uses Contains (substring) matching ---
		{
			name: "global: substring match Chinese directory name",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("硬件安全/微体系结构攻击.md", "# 微体系结构攻击\nContent.", "Add", "author")
			},
			dir:    "",
			prefix: "体系",
			// "体系" should match inside "微体系结构攻击" (substring of title/filename)
			wantMin: 1,
		},
		{
			name: "global: substring match Chinese filename",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("硬件安全/微体系结构攻击.md", "# 微体系结构攻击\nContent.", "Add", "author")
			},
			dir:    "",
			prefix: "攻击",
			// "攻击" is a substring of title "微体系结构攻击"
			wantMin: 1,
		},
		{
			name: "global: prefix match English directory name",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("hardware/usb-pd.md", "# USB PD\nContent.", "Add", "author")
			},
			dir:       "",
			prefix:    "hard",
			wantItems: []string{"hardware/"},
		},
		{
			name: "global: substring match English filename",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("linux/qemu-setup.md", "# QEMU Setup\nContent.", "Add", "author")
			},
			dir:      "",
			prefix:   "qemu",
			wantMin:  1,
		},
		{
			name: "global: no match for unrelated Chinese query",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("hardware/usb-pd.md", "# USB PD\nContent.", "Add", "author")
			},
			dir:       "",
			prefix:    "不存在",
			wantItems: []string{},
		},
		// --- Directory-scoped completion (dir != ""): Uses HasPrefix for filenames/dirs, Contains for titles ---
		{
			name: "dir-scoped: title substring match still works",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("硬件安全/微体系结构攻击.md", "# 微体系结构攻击\nContent.", "Add", "author")
			},
			dir:      "硬件安全",
			prefix:   "攻击",
			wantMin:  1, // title Contains "攻击" → match
		},
		{
			name: "dir-scoped: filename prefix match",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("tech/go-best-practices.md", "# Go Best Practices\nContent.", "Add", "author")
			},
			dir:       "tech",
			prefix:    "go",
			wantItems: []string{"go-best-practices.md"},
		},
		{
			name: "dir-scoped: filename HasPrefix fails but title Contains match still works",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("tech/go-best-practices.md", "# Go Best Practices\nContent.", "Add", "author")
			},
			dir:      "tech",
			prefix:   "best",
			wantMin:  1, // title "Go Best Practices" Contains "best" → match despite filename HasPrefix failure
		},
		{
			name: "dir-scoped: subdir prefix match",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("hardware/cpu/intel.md", "# Intel CPU\nContent.", "Add", "author")
			},
			dir:       "hardware",
			prefix:    "cp",
			wantItems: []string{"cpu/"},
		},
		{
			name: "dir-scoped: subdir substring does NOT match (HasPrefix only)",
			setup: func(repo *testutil.TestRepo) error {
				return repo.CreateMarkdownFile("hardware/cpu/intel.md", "# Intel CPU\nContent.", "Add", "author")
			},
			dir:       "hardware",
			prefix:    "pu",
			wantItems: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := testutil.NewTestRepo()
			require.NoError(t, err)
			defer repo.Cleanup()

			require.NoError(t, tt.setup(repo))

			fileSvc, err := service.NewFileService(repo.Path)
			require.NoError(t, err)

			gitSvc, err := service.NewGitService(repo.Path)
			require.NoError(t, err)

			articleSvc := service.NewArticleService(fileSvc, gitSvc)
			completionSvc := service.NewCompletionService(articleSvc, fileSvc, gitSvc)

			items, err := completionSvc.GetMatchingItems(context.Background(), tt.dir, tt.prefix)
			require.NoError(t, err)

			if tt.wantItems != nil {
				assert.ElementsMatch(t, tt.wantItems, items)
			} else {
				assert.GreaterOrEqual(t, len(items), tt.wantMin,
					"expected at least %d matches for prefix %q in dir %q, got %d: %v", tt.wantMin, tt.prefix, tt.dir, len(items), items)
			}
		})
	}
}