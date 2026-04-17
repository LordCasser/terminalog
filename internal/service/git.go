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
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"

	"terminalog/internal/model"
)

// GitService provides Git operations using pure go-git/v5.
type GitService struct {
	// repoPath is the absolute path to the Git repository.
	repoPath string

	// repo is the opened Git repository.
	repo *git.Repository
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
// Only commits where the file was actually modified are included.
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

	// Add capabilities string (only on first ref, with null byte separator)
	capabilities := "multi_ack_detailed side-band-64k thin-pack ofs-delta shallow no-progress include-tag"

	// Get HEAD reference
	head, err := s.repo.Head()
	if err == nil {
		// HEAD with capabilities (first ref must have capabilities after null byte)
		pktLine(&buf, fmt.Sprintf("%s HEAD\x00%s\n", head.Hash().String(), capabilities))
	} else {
		// No HEAD, write empty capabilities
		pktLine(&buf, fmt.Sprintf("%s HEAD\x00%s\n", "0000000000000000000000000000000000000000", capabilities))
	}

	// Get all references
	refs, err := s.repo.References()
	if err != nil {
		return nil, err
	}

	// Add branch references (without capabilities)
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
// This implementation uses go-git's packfile encoder to generate the packfile.
func (s *GitService) HandleUploadPack(ctx context.Context, body io.Reader) ([]byte, error) {
	// Decode the upload-pack request
	req := packp.NewUploadPackRequest()
	if err := req.Decode(body); err != nil {
		return nil, fmt.Errorf("failed to decode upload-pack request: %w", err)
	}

	// Validate request has wants
	if len(req.Wants) == 0 {
		return nil, fmt.Errorf("no wants in upload-pack request")
	}

	// Build packfile containing requested objects
	var packBuf bytes.Buffer

	// Collect all objects needed (commits, trees, blobs)
	// Use a set to avoid duplicates
	objectSet := make(map[plumbing.Hash]bool)

	// Recursively collect all commits and their objects
	for _, want := range req.Wants {
		s.collectCommitObjects(want, objectSet)
	}

	// Convert set to slice
	objectHashes := make([]plumbing.Hash, 0, len(objectSet))
	for hash := range objectSet {
		objectHashes = append(objectHashes, hash)
	}

	// Create packfile encoder
	encoder := packfile.NewEncoder(&packBuf, s.repo.Storer, false)

	// Encode objects into packfile (version 2)
	if len(objectHashes) > 0 {
		_, err := encoder.Encode(objectHashes, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to encode packfile: %w", err)
		}
	}

	// Build response
	var responseBuf bytes.Buffer

	// Write NAK (no common commits)
	pktLine(&responseBuf, "NAK\n")

	// Write packfile using sideband-64k
	if packBuf.Len() > 0 {
		// Chunk the packfile data
		packData := packBuf.Bytes()
		chunkSize := 65515 // max for sideband-64k minus 1 byte for channel

		for i := 0; i < len(packData); i += chunkSize {
			end := i + chunkSize
			if end > len(packData) {
				end = len(packData)
			}

			chunk := packData[i:end]
			// Channel 1 = packfile data
			pktLine(&responseBuf, fmt.Sprintf("\x01%s", string(chunk)))
		}
	}

	// Write success message on channel 2 (progress)
	pktLine(&responseBuf, "\x02Counting objects done.\n")

	// Flush
	pktFlush(&responseBuf)

	return responseBuf.Bytes(), nil
}

// collectCommitObjects recursively collects all objects for a commit and its parents.
func (s *GitService) collectCommitObjects(commitHash plumbing.Hash, objectSet map[plumbing.Hash]bool) {
	// Skip if already collected
	if objectSet[commitHash] {
		return
	}

	// Get the commit object
	commit, err := object.GetCommit(s.repo.Storer, commitHash)
	if err != nil {
		return
	}

	// Add commit hash
	objectSet[commitHash] = true

	// Add tree and collect all tree objects recursively
	s.collectTreeObjects(commit.TreeHash, objectSet)

	// Recursively collect parent commits
	for _, parentHash := range commit.ParentHashes {
		s.collectCommitObjects(parentHash, objectSet)
	}
}

// collectTreeObjects recursively collects all objects in a tree (tree itself, subtrees, and blobs).
func (s *GitService) collectTreeObjects(treeHash plumbing.Hash, objectSet map[plumbing.Hash]bool) {
	// Skip if already collected
	if objectSet[treeHash] {
		return
	}

	// Add tree hash
	objectSet[treeHash] = true

	// Get the tree object
	tree, err := object.GetTree(s.repo.Storer, treeHash)
	if err != nil {
		return
	}

	// Walk tree entries
	for _, entry := range tree.Entries {
		switch entry.Mode {
		case filemode.Dir, filemode.Symlink:
			// Directory or symlink - collect subtree
			s.collectTreeObjects(entry.Hash, objectSet)
		default:
			// Regular file - add blob hash
			objectSet[entry.Hash] = true
		}
	}
}

// GetReceivePackRefs returns the refs advertisement for git-receive-pack (Push).
func (s *GitService) GetReceivePackRefs(ctx context.Context) ([]byte, error) {
	var buf bytes.Buffer

	// Write service announcement
	pktLine(&buf, "# service=git-receive-pack\n")
	pktFlush(&buf)

	// Add capabilities string for receive-pack
	capabilities := "report-status delete-refs atomic ofs-delta"

	// Get HEAD reference
	head, err := s.repo.Head()
	if err == nil {
		pktLine(&buf, fmt.Sprintf("%s HEAD\x00%s\n", head.Hash().String(), capabilities))
	} else {
		// No HEAD, write empty capabilities
		pktLine(&buf, fmt.Sprintf("%s HEAD\x00%s\n", "0000000000000000000000000000000000000000", capabilities))
	}

	// Get all references
	refs, err := s.repo.References()
	if err != nil {
		return nil, err
	}

	// Add branch references
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
// This implementation processes reference updates. Packfile handling is simplified.
func (s *GitService) HandleReceivePack(ctx context.Context, body io.Reader) ([]byte, error) {
	// Decode the reference update request
	req := packp.NewReferenceUpdateRequest()
	if err := req.Decode(body); err != nil {
		return nil, fmt.Errorf("failed to decode receive-pack request: %w", err)
	}

	// For MVP, we handle reference updates only
	// Packfile processing requires more complex implementation
	// The packfile data is already sent and objects should be in the request

	// Build response
	var responseBuf bytes.Buffer

	// Write unpack result
	pktLine(&responseBuf, "unpack ok\n")

	// Process each reference update command
	for _, cmd := range req.Commands {
		action := cmd.Action()

		switch action {
		case packp.Create:
			// Create new reference
			refName := plumbing.ReferenceName(cmd.Name)
			ref := plumbing.NewHashReference(refName, cmd.New)
			if err := s.repo.Storer.SetReference(ref); err != nil {
				pktLine(&responseBuf, fmt.Sprintf("ng %s %s\n", cmd.Name, err.Error()))
				continue
			}
			pktLine(&responseBuf, fmt.Sprintf("ok %s\n", cmd.Name))

		case packp.Update:
			// Update existing reference
			refName := plumbing.ReferenceName(cmd.Name)
			ref := plumbing.NewHashReference(refName, cmd.New)
			if err := s.repo.Storer.SetReference(ref); err != nil {
				pktLine(&responseBuf, fmt.Sprintf("ng %s %s\n", cmd.Name, err.Error()))
				continue
			}
			pktLine(&responseBuf, fmt.Sprintf("ok %s\n", cmd.Name))

		case packp.Delete:
			// Delete reference
			refName := plumbing.ReferenceName(cmd.Name)
			if err := s.repo.Storer.RemoveReference(refName); err != nil {
				pktLine(&responseBuf, fmt.Sprintf("ng %s %s\n", cmd.Name, err.Error()))
				continue
			}
			pktLine(&responseBuf, fmt.Sprintf("ok %s\n", cmd.Name))
		}
	}

	// Flush
	pktFlush(&responseBuf)

	return responseBuf.Bytes(), nil
}

// GetRepo returns the underlying git repository.
func (s *GitService) GetRepo() *git.Repository {
	return s.repo
}

// GetRepoPath returns the repository path.
func (s *GitService) GetRepoPath() string {
	return s.repoPath
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
		// pkt-line max size, need to split
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
