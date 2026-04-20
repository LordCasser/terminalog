// Package service provides business logic services for the application.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"terminalog/internal/model"
)

// GitService provides Git operations.
// Smart HTTP protocol (clone/push) uses system git subprocesses (--stateless-rpc).
// Read-only operations (file history, etc.) use go-git/v5.
type GitService struct {
	// repoPath is the absolute path to the Git repository.
	repoPath string

	// repo is the opened Git repository (for read-only operations).
	repo *git.Repository

	// historyCache avoids rescanning the full commit history for the same file.
	historyCache sync.Map // map[string]*model.FileHistory

	// diffCache avoids recomputing commit diff statistics for the same file.
	diffCache sync.Map // map[string][]model.CommitDiffInfo
}

// NewGitService creates a new GitService instance.
// It opens the repository and configures it for push operations:
// - Sets receive.denyCurrentBranch=ignore to allow pushing to the checked-out branch
func NewGitService(repoPath string) (*GitService, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Configure repository to allow pushing to the checked-out branch.
	// By default, git rejects pushes to the current branch in non-bare repos.
	// We handle working directory updates ourselves via checkout after push.
	cmd := exec.Command("git", "config", "receive.denyCurrentBranch", "ignore")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		log.Printf("NewGitService: failed to set receive.denyCurrentBranch: %v", err)
		// Non-critical - push may fail later if this isn't set
	}

	return &GitService{
		repoPath: repoPath,
		repo:     repo,
	}, nil
}

// ----- Smart HTTP Protocol (git subprocess) -----

// ServiceType constants for Git Smart HTTP protocol.
const (
	ServiceTypeUploadPack  = "upload-pack"
	ServiceTypeReceivePack = "receive-pack"
)

// GetInfoRefs runs `git {upload-pack|receive-pack} --stateless-rpc --advertise-refs .`
// to produce the reference advertisement for the Smart HTTP protocol.
func (s *GitService) GetInfoRefs(service string) ([]byte, error) {
	if service != ServiceTypeUploadPack && service != ServiceTypeReceivePack {
		return nil, fmt.Errorf("invalid service: %s", service)
	}

	cmd := exec.Command("git", service, "--stateless-rpc", "--advertise-refs", ".")
	cmd.Dir = s.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s --advertise-refs failed: %w", service, err)
	}

	return output, nil
}

// ServiceRPC pipes the HTTP request body to a git subprocess and streams
// the response back. This handles both upload-pack (clone/fetch) and
// receive-pack (push) using `git {service} --stateless-rpc .`.
func (s *GitService) ServiceRPC(service string, reqBody io.Reader, respWriter io.Writer) error {
	if service != ServiceTypeUploadPack && service != ServiceTypeReceivePack {
		return fmt.Errorf("invalid service: %s", service)
	}

	cmd := exec.Command("git", service, "--stateless-rpc", ".")
	cmd.Dir = s.repoPath
	cmd.Stdin = reqBody
	cmd.Stdout = respWriter

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("ServiceRPC: git %s --stateless-rpc failed: %v, stderr: %s", service, err, stderr.String())
		return fmt.Errorf("git %s --stateless-rpc failed: %w", service, err)
	}

	return nil
}

// ----- Read-only operations (go-git) -----

// GetFileHistory returns the complete Git history of a file.
// Only commits where the file was actually modified are included.
func (s *GitService) GetFileHistory(ctx context.Context, filePath string) (*model.FileHistory, error) {
	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	// Normalize path
	filePath = strings.TrimPrefix(filePath, "/")

	if cached, ok := s.historyCache.Load(filePath); ok {
		return cached.(*model.FileHistory), nil
	}

	// Get all commits
	commitIter, err := s.repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	// Collect all commits (newest first from Log)
	commits := make([]*object.Commit, 0)
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reverse to get oldest first
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	// Process commits to find actual file modifications
	fileCommits := make([]model.CommitInfo, 0)
	contributors := make(map[string]bool)

	// Track the file hash to detect changes
	prevFileHash := plumbing.ZeroHash
	fileExists := false

	// Iterate in chronological order (oldest first)
	for _, c := range commits {
		// Check if file exists in this commit
		file, err := c.File(filePath)
		if err != nil {
			// File doesn't exist in this commit
			if errors.Is(err, object.ErrFileNotFound) {
				// If file existed before, this is a deletion
				if fileExists {
					commitInfo := model.CommitInfo{
						Hash:      shortHash(c.Hash.String()),
						Author:    c.Author.Name,
						Timestamp: c.Author.When,
						Message:   strings.Split(c.Message, "\n")[0],
					}
					fileCommits = append(fileCommits, commitInfo)
					contributors[c.Author.Name] = true
				}
				prevFileHash = plumbing.ZeroHash
				fileExists = false
				continue
			}
			return nil, err
		}

		// File exists in this commit
		currentFileHash := file.Hash

		// First appearance (creation) or hash changed (modification)
		if !fileExists || currentFileHash != prevFileHash {
			commitInfo := model.CommitInfo{
				Hash:      shortHash(c.Hash.String()),
				Author:    c.Author.Name,
				Timestamp: c.Author.When,
				Message:   strings.Split(c.Message, "\n")[0],
			}
			fileCommits = append(fileCommits, commitInfo)
			contributors[c.Author.Name] = true
		}

		prevFileHash = currentFileHash
		fileExists = true
	}

	// Check if we have any history
	if len(fileCommits) == 0 {
		return nil, model.ErrNotCommitted
	}

	// Sort by time descending (most recent first)
	sort.Slice(fileCommits, func(i, j int) bool {
		return fileCommits[i].Timestamp.After(fileCommits[j].Timestamp)
	})

	// Build result
	history := &model.FileHistory{
		FirstCommit:  fileCommits[len(fileCommits)-1], // Oldest
		LastCommit:   fileCommits[0],                  // Newest
		AllCommits:   fileCommits,
		Contributors: mapKeys(contributors),
	}

	s.historyCache.Store(filePath, history)

	return history, nil
}

// GetFileHistories returns Git history for multiple files using a single commit walk.
func (s *GitService) GetFileHistories(ctx context.Context, filePaths []string) (map[string]*model.FileHistory, error) {
	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	targets := make(map[string]struct{}, len(filePaths))
	results := make(map[string]*model.FileHistory, len(filePaths))
	pending := make([]string, 0, len(filePaths))

	for _, filePath := range filePaths {
		normalized := strings.TrimPrefix(filePath, "/")
		if normalized == "" {
			continue
		}
		if _, seen := targets[normalized]; seen {
			continue
		}
		targets[normalized] = struct{}{}

		if cached, ok := s.historyCache.Load(normalized); ok {
			results[normalized] = cached.(*model.FileHistory)
			continue
		}

		pending = append(pending, normalized)
	}

	if len(pending) == 0 {
		return results, nil
	}

	commitIter, err := s.repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	commits := make([]*object.Commit, 0)
	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	type historyBuilder struct {
		commits      []model.CommitInfo
		contributors map[string]bool
		prevHash     plumbing.Hash
		fileExists   bool
	}

	builders := make(map[string]*historyBuilder, len(pending))
	for _, filePath := range pending {
		builders[filePath] = &historyBuilder{
			commits:      make([]model.CommitInfo, 0),
			contributors: make(map[string]bool),
		}
	}

	for _, c := range commits {
		for filePath, builder := range builders {
			file, err := c.File(filePath)
			if err != nil {
				if errors.Is(err, object.ErrFileNotFound) {
					if builder.fileExists {
						commitInfo := model.CommitInfo{
							Hash:      shortHash(c.Hash.String()),
							Author:    c.Author.Name,
							Timestamp: c.Author.When,
							Message:   strings.Split(c.Message, "\n")[0],
						}
						builder.commits = append(builder.commits, commitInfo)
						builder.contributors[c.Author.Name] = true
					}
					builder.prevHash = plumbing.ZeroHash
					builder.fileExists = false
					continue
				}
				return nil, err
			}

			currentHash := file.Hash
			if !builder.fileExists || currentHash != builder.prevHash {
				commitInfo := model.CommitInfo{
					Hash:      shortHash(c.Hash.String()),
					Author:    c.Author.Name,
					Timestamp: c.Author.When,
					Message:   strings.Split(c.Message, "\n")[0],
				}
				builder.commits = append(builder.commits, commitInfo)
				builder.contributors[c.Author.Name] = true
			}

			builder.prevHash = currentHash
			builder.fileExists = true
		}
	}

	for filePath, builder := range builders {
		if len(builder.commits) == 0 {
			continue
		}

		sort.Slice(builder.commits, func(i, j int) bool {
			return builder.commits[i].Timestamp.After(builder.commits[j].Timestamp)
		})

		history := &model.FileHistory{
			FirstCommit:  builder.commits[len(builder.commits)-1],
			LastCommit:   builder.commits[0],
			AllCommits:   builder.commits,
			Contributors: mapKeys(builder.contributors),
		}

		s.historyCache.Store(filePath, history)
		results[filePath] = history
	}

	return results, nil
}

// GetFileCommitDiffs returns real diff statistics (lines added/removed) for each
// commit that touched a file. The results are ordered oldest-first.
// It uses go-git's Patch API to compute accurate add/remove counts per commit.
// Results are cached and invalidated on git push (via ReloadRepo).
func (s *GitService) GetFileCommitDiffs(ctx context.Context, filePath string) ([]model.CommitDiffInfo, error) {
	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	filePath = strings.TrimPrefix(filePath, "/")

	// Check cache first
	if cached, ok := s.diffCache.Load(filePath); ok {
		return cached.([]model.CommitDiffInfo), nil
	}

	// Get file history to obtain ordered commit list
	history, err := s.GetFileHistory(ctx, filePath)
	if err != nil {
		return nil, err
	}

	if len(history.AllCommits) == 0 {
		return nil, model.ErrNotCommitted
	}

	// We need commits in chronological order (oldest first)
	commits := reverseCommits(history.AllCommits)

	// Resolve each commit hash back to a go-git commit object
	var gitCommits []*object.Commit
	for _, ci := range commits {
		hash := plumbing.NewHash(ci.Hash)
		// ci.Hash is a short hash (7 chars); resolve it via repo
		commitObj, err := s.repo.CommitObject(hash)
		if err != nil {
			// Try to find by prefix
			found := false
			iter, iterErr := s.repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
			if iterErr != nil {
				return nil, iterErr
			}
			iterErr = iter.ForEach(func(c *object.Commit) error {
				if strings.HasPrefix(c.Hash.String(), ci.Hash) {
					gitCommits = append(gitCommits, c)
					found = true
				}
				return nil
			})
			if iterErr != nil || !found {
				return nil, fmt.Errorf("commit not found: %s", ci.Hash)
			}
		} else {
			gitCommits = append(gitCommits, commitObj)
		}
	}

	diffs := make([]model.CommitDiffInfo, 0, len(gitCommits))

	for i, commit := range gitCommits {
		var added, removed, fileLinesAfter int

		if i == 0 {
			// First commit: file creation. The "diff" is the entire file content.
			file, err := commit.File(filePath)
			if err != nil {
				// File might not exist in this commit (e.g., it was deleted)
				diffs = append(diffs, model.CommitDiffInfo{
					Hash:           shortHash(commit.Hash.String()),
					LinesAdded:     0,
					LinesRemoved:   0,
					FileLinesAfter: 0,
				})
				continue
			}
			contents, err := file.Contents()
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s at commit %s: %w", filePath, shortHash(commit.Hash.String()), err)
			}
			fileLinesAfter = countLines(contents)
			added = fileLinesAfter
			removed = 0
		} else {
			// Subsequent commit: compute patch from parent
			parent := gitCommits[i-1]

			patch, err := parent.Patch(commit)
			if err != nil {
				// If patch fails (e.g., parent didn't have the file), fall back to diff stat
				diffs = append(diffs, model.CommitDiffInfo{
					Hash:           shortHash(commit.Hash.String()),
					LinesAdded:     0,
					LinesRemoved:   0,
					FileLinesAfter: 0,
				})
				continue
			}

			// Find our file in the patch stats
			stats := patch.Stats()
			for _, stat := range stats {
				if stat.Name == filePath {
					added = stat.Addition
					removed = stat.Deletion
					break
				}
			}

			// Get the file content at this commit for line count
			file, err := commit.File(filePath)
			if err != nil {
				// File was deleted in this commit
				diffs = append(diffs, model.CommitDiffInfo{
					Hash:           shortHash(commit.Hash.String()),
					LinesAdded:     added,
					LinesRemoved:   removed,
					FileLinesAfter: 0,
				})
				continue
			}
			contents, err := file.Contents()
			if err != nil {
				contents = ""
			}
			fileLinesAfter = countLines(contents)
		}

		diffs = append(diffs, model.CommitDiffInfo{
			Hash:           shortHash(commit.Hash.String()),
			LinesAdded:     added,
			LinesRemoved:   removed,
			FileLinesAfter: fileLinesAfter,
		})
	}

	// Store in cache before returning
	s.diffCache.Store(filePath, diffs)
	return diffs, nil
}

// countLines counts the number of lines in a string.
func countLines(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Split(content, "\n"))
}

// IsFileCommitted checks if a file has been committed to Git.
func (s *GitService) IsFileCommitted(ctx context.Context, filePath string) (bool, error) {
	history, err := s.GetFileHistory(ctx, filePath)
	if err != nil {
		if errors.Is(err, model.ErrNotCommitted) {
			return false, nil
		}
		return false, err
	}
	return len(history.AllCommits) > 0, nil
}

// GetRepo returns the underlying git repository.
func (s *GitService) GetRepo() *git.Repository {
	return s.repo
}

// GetRepoPath returns the repository path.
func (s *GitService) GetRepoPath() string {
	return s.repoPath
}

// ReloadRepo re-opens the git repository to refresh cached state.
// Called after push operations to ensure caches reflect the latest commits.
func (s *GitService) ReloadRepo() error {
	repo, err := git.PlainOpen(s.repoPath)
	if err != nil {
		return fmt.Errorf("failed to re-open git repository: %w", err)
	}
	s.repo = repo
	s.historyCache = sync.Map{}
	s.diffCache = sync.Map{}
	return nil
}

// CheckoutWorkingTree runs `git checkout --force` in the repository to
// synchronize the working directory with the current HEAD.
// This is needed after push operations because `git receive-pack` only
// updates refs and objects, not the working tree.
func (s *GitService) CheckoutWorkingTree() error {
	cmd := exec.Command("git", "checkout", "--force")
	cmd.Dir = s.repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("CheckoutWorkingTree: git checkout failed: %v, stderr: %s", err, stderr.String())
		return fmt.Errorf("git checkout failed: %w", err)
	}

	log.Printf("CheckoutWorkingTree: working directory updated to match HEAD")
	return nil
}

// shortHash returns a short commit hash (7 characters).
func shortHash(hash string) string {
	if len(hash) >= 7 {
		return hash[:7]
	}
	return hash
}

// mapKeys returns the keys of a map as a slice.
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
