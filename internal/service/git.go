// Package service provides business logic services for the application.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"terminalog/internal/model"
)

// GitService provides Git operations.
type GitService struct {
	// repoPath is the absolute path to the Git repository.
	repoPath string

	// repo is the opened Git repository.
	repo *git.Repository

	// storer is the storage backend.
	storer *filesystem.Storage
}

// NewGitService creates a new GitService instance.
func NewGitService(repoPath string) (*GitService, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GitService{
		repoPath: repoPath,
		repo:     repo,
	}, nil
}

// GetFileHistory returns the complete Git history of a file.
func (s *GitService) GetFileHistory(ctx context.Context, filePath string) (*model.FileHistory, error) {
	if s.repo == nil {
		return nil, model.ErrRepoNotFound
	}

	// Normalize path
	filePath = strings.TrimPrefix(filePath, "/")

	// Get all commits
	commitIter, err := s.repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	// Filter commits that involve this file
	fileCommits := make([]model.CommitInfo, 0)
	contributors := make(map[string]bool)

	err = commitIter.ForEach(func(c *object.Commit) error {
		// Check if file exists in this commit
		file, err := c.File(filePath)
		if err != nil {
			// File doesn't exist in this commit, skip
			if errors.Is(err, object.ErrFileNotFound) {
				return nil
			}
			return err
		}
		_ = file // We just need to know it exists

		// Record commit info
		commitInfo := model.CommitInfo{
			Hash:      shortHash(c.Hash.String()),
			Author:    c.Author.Name,
			Timestamp: c.Author.When,
			Message:   strings.Split(c.Message, "\n")[0], // First line only
		}

		fileCommits = append(fileCommits, commitInfo)
		contributors[c.Author.Name] = true

		return nil
	})

	if err != nil {
		return nil, err
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

	return history, nil
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

// GetUploadPackRefs returns the refs advertisement for git-upload-pack (Clone).
func (s *GitService) GetUploadPackRefs(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer

	// Write service announcement
	pktLine(&buf, "# service=git-upload-pack\n")
	pktFlush(&buf)

	// Get references
	refs, err := s.repo.References()
	if err != nil {
		return nil, err
	}

	// Write HEAD first
	head, err := s.repo.Head()
	if err == nil {
		pktLine(&buf, fmt.Sprintf("%s HEAD\n", head.Hash().String()))
	}

	// Write all branches
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			pktLine(&buf, fmt.Sprintf("%s %s\n", ref.Hash().String(), ref.Name().String()))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	pktFlush(&buf)

	return buf.Bytes(), nil
}

// HandleUploadPack handles the git-upload-pack request (Clone/Fetch).
// For MVP, this uses a simplified implementation.
func (s *GitService) HandleUploadPack(ctx context.Context, body io.Reader) ([]byte, error) {
	// Read and parse the wants/haves from the request body
	// For MVP, we use a simplified approach
	// The full implementation would require proper packfile generation

	// This is a placeholder - full implementation would use go-git's
	// internal packfile generation mechanisms
	// For now, return an error indicating this needs system git
	return nil, fmt.Errorf("git-upload-pack requires system git binary for full functionality")
}

// GetReceivePackRefs returns the refs advertisement for git-receive-pack (Push).
func (s *GitService) GetReceivePackRefs(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer

	// Write service announcement
	pktLine(&buf, "# service=git-receive-pack\n")
	pktFlush(&buf)

	// Get references
	refs, err := s.repo.References()
	if err != nil {
		return nil, err
	}

	// Write HEAD first
	head, err := s.repo.Head()
	if err == nil {
		pktLine(&buf, fmt.Sprintf("%s HEAD\n", head.Hash().String()))
	}

	// Write all branches
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			pktLine(&buf, fmt.Sprintf("%s %s\n", ref.Hash().String(), ref.Name().String()))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	pktFlush(&buf)

	return buf.Bytes(), nil
}

// HandleReceivePack handles the git-receive-pack request (Push).
// For MVP, this uses a simplified implementation.
func (s *GitService) HandleReceivePack(ctx context.Context, body io.Reader) ([]byte, error) {
	// For MVP, we use a simplified approach
	// The full implementation would require proper packfile parsing

	// This is a placeholder - full implementation would use go-git's
	// packfile parsing and object storage mechanisms
	// For now, return an error indicating this needs system git
	return nil, fmt.Errorf("git-receive-pack requires system git binary for full functionality")
}

// GetRepo returns the underlying git repository.
func (s *GitService) GetRepo() *git.Repository {
	return s.repo
}

// Helper functions

// shortHash returns a short commit hash (7 characters).
func shortHash(hash string) string {
	if len(hash) >= 7 {
		return hash[:7]
	}
	return hash
}

// pktLine writes a pkt-line formatted data.
func pktLine(w io.Writer, data string) {
	size := len(data) + 4
	if size > 65524 {
		// pkt-line max size
		return
	}
	w.Write([]byte(fmt.Sprintf("%04x%s", size, data)))
}

// pktFlush writes a pkt-line flush packet.
func pktFlush(w io.Writer) {
	w.Write([]byte("0000"))
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
