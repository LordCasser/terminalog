// Package service provides business logic services for the application.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// GitService provides Git operations.
// Smart HTTP protocol (clone/push) uses system git subprocesses (--stateless-rpc).
// Read-only operations (file history, etc.) use go-git/v5.
type GitService struct {
	// repoPath is the absolute path to the Git repository.
	repoPath string

	// repo is the opened Git repository (for read-only operations).
	repo *git.Repository

	// mu protects repo, historyCache, and diffCache fields
	// from concurrent access during ReloadRepo.
	mu sync.RWMutex

	// updateMu serializes complete receive-pack publication transactions.
	updateMu sync.Mutex

	// historyCache avoids rescanning the full commit history for the same file.
	historyCache sync.Map // map[string]*model.FileHistory

	// diffCache avoids recomputing commit diff statistics for the same file.
	diffCache sync.Map // map[string][]model.CommitDiffInfo
}

const (
	specialFilePrefix = "_"
	assetsDirName     = ".assets"
)

// NewGitService creates a new GitService instance.
// It opens the repository and configures it for push operations:
//   - Sets receive.denyCurrentBranch=updateInstead so Git updates the worktree
//     before reporting a successful push.
func NewGitService(repoPath string) (*GitService, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Git is the authority for both the ref and worktree update. updateInstead
	// rejects pushes when the worktree is dirty and only reports success after
	// the checked-out files have been updated.
	cmd := exec.Command("git", "config", "receive.denyCurrentBranch", "updateInstead")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("configure receive.denyCurrentBranch: %w", err)
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
		return fmt.Errorf("git %s --stateless-rpc failed: %w (stderr: %s)", service, err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

// ReceivePack applies one push as a publication transaction. The protocol
// response is buffered so the client cannot observe success before the
// worktree, go-git reader, and dependent caches represent the same commit.
func (s *GitService) ReceivePack(reqBody io.Reader, onRepoUpdate func()) ([]byte, error) {
	s.updateMu.Lock()
	defer s.updateMu.Unlock()

	var response bytes.Buffer
	if err := s.ServiceRPC(ServiceTypeReceivePack, reqBody, &response); err != nil {
		return nil, err
	}
	if err := s.ReloadRepo(); err != nil {
		return nil, fmt.Errorf("refresh repository after receive-pack: %w", err)
	}
	if onRepoUpdate != nil {
		onRepoUpdate()
	}

	return response.Bytes(), nil
}

// ----- Read-only operations (go-git) -----

// GetFileHistory returns the complete Git history of a file.
// Only commits where the file was actually modified are included.
func (s *GitService) GetFileHistory(ctx context.Context, filePath string) (*model.FileHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getFileHistoryLocked(ctx, filePath)
}

// getFileHistoryLocked requires s.mu to be held for reading. Keeping the
// implementation separate avoids recursively acquiring an RWMutex when
// version calculation needs history and diffs from the same repository view.
func (s *GitService) getFileHistoryLocked(_ context.Context, filePath string) (*model.FileHistory, error) {

	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	filePath, err := utils.ValidateContentPath(filePath)
	if err != nil || filePath == "" {
		return nil, model.ErrInvalidPath
	}

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

	// Historical existence is not public visibility. A path deleted from the
	// current HEAD must remain invisible even if it is recreated in the raw
	// worktree without a commit.
	if len(fileCommits) == 0 || !fileExists {
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	targets := make(map[string]struct{}, len(filePaths))
	results := make(map[string]*model.FileHistory, len(filePaths))
	pending := make([]string, 0, len(filePaths))

	for _, filePath := range filePaths {
		normalized, err := utils.ValidateContentPath(filePath)
		if err != nil || normalized == "" {
			return nil, model.ErrInvalidPath
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
		if len(builder.commits) == 0 || !builder.fileExists {
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	filePath, err := utils.ValidateContentPath(filePath)
	if err != nil || filePath == "" {
		return nil, model.ErrInvalidPath
	}

	// Check cache first
	if cached, ok := s.diffCache.Load(filePath); ok {
		return cached.([]model.CommitDiffInfo), nil
	}

	// Get file history to obtain ordered commit list
	history, err := s.getFileHistoryLocked(ctx, filePath)
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

// NodeTypeAtHead resolves a path against the published commit tree.
func (s *GitService) NodeTypeAtHead(filePath string) (model.NodeType, error) {
	filePath, err := utils.ValidateContentPath(filePath)
	if err != nil {
		return "", err
	}
	if filePath == "" {
		return model.NodeTypeDir, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	commit, err := s.headCommitLocked()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", model.ErrNotFound
		}
		return "", err
	}
	tree, err := commit.Tree()
	if err != nil {
		return "", err
	}
	entry, err := tree.FindEntry(filePath)
	if err != nil {
		if errors.Is(err, object.ErrEntryNotFound) || errors.Is(err, object.ErrDirectoryNotFound) {
			return "", model.ErrNotFound
		}
		return "", err
	}
	if entry.Mode == filemode.Dir {
		return model.NodeTypeDir, nil
	}
	return model.NodeTypeFile, nil
}

// ReadFileAtHead reads the exact blob published by the current HEAD, never the
// mutable server worktree.
func (s *GitService) ReadFileAtHead(filePath string) ([]byte, error) {
	filePath, err := utils.ValidateContentPath(filePath)
	if err != nil || filePath == "" {
		return nil, model.ErrInvalidPath
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	commit, err := s.headCommitLocked()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	file, err := commit.File(filePath)
	if err != nil {
		if errors.Is(err, object.ErrFileNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	reader, err := file.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// ListMarkdownFilesAtHead returns recursively visible articles below dir.
func (s *GitService) ListMarkdownFilesAtHead(dir string) ([]string, error) {
	dir, err := utils.ValidateContentPath(dir)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	commit, err := s.headCommitLocked()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) && dir == "" {
			return []string{}, nil
		}
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	if dir != "" {
		tree, err = tree.Tree(dir)
		if err != nil {
			if errors.Is(err, object.ErrDirectoryNotFound) {
				return nil, model.ErrNotFound
			}
			return nil, err
		}
	}

	files := make([]string, 0)
	err = tree.Files().ForEach(func(file *object.File) error {
		filePath := utils.NormalizePath(file.Name)
		segments := strings.Split(filePath, "/")
		for _, segment := range segments {
			if segment == assetsDirName || strings.HasPrefix(segment, specialFilePrefix) {
				return nil
			}
		}
		if !utils.IsMarkdownFile(filePath) {
			return nil
		}
		if dir != "" {
			filePath = dir + "/" + filePath
		}
		files = append(files, filePath)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func (s *GitService) headCommitLocked() (*object.Commit, error) {
	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}
	ref, err := s.repo.Head()
	if err != nil {
		return nil, err
	}
	return s.repo.CommitObject(ref.Hash())
}

// CurrentHead returns the commit currently published by the read view. An
// unborn but valid repository returns an empty hash.
func (s *GitService) CurrentHead() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.repo == nil {
		return "", model.ErrRepoNotFound
	}
	ref, err := s.repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", nil
		}
		return "", err
	}
	return ref.Hash().String(), nil
}

// ReloadRepo re-opens the git repository to refresh cached read state after a
// push. The replacement is verified before it is published to readers.
func (s *GitService) ReloadRepo() error {
	repo, err := git.PlainOpen(s.repoPath)
	if err != nil {
		return fmt.Errorf("failed to re-open git repository: %w", err)
	}
	if _, err := repo.Head(); err != nil {
		return fmt.Errorf("failed to verify repository HEAD after reload: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.repo = repo
	s.historyCache = sync.Map{}
	s.diffCache = sync.Map{}
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
