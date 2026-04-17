// Package testutil provides utilities for testing.
package testutil

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TestRepo represents a test Git repository.
type TestRepo struct {
	// Path is the absolute path to the repository.
	Path string

	// Repo is the git repository instance.
	Repo *git.Repository
}

// NewTestRepo creates a new test Git repository in a temporary directory.
func NewTestRepo() (*TestRepo, error) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "terminalog-test-")
	if err != nil {
		return nil, err
	}

	// Initialize git repo
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	return &TestRepo{
		Path: tmpDir,
		Repo: repo,
	}, nil
}

// Cleanup removes the test repository directory.
func (r *TestRepo) Cleanup() {
	os.RemoveAll(r.Path)
}

// CreateFile creates a file in the repository.
func (r *TestRepo) CreateFile(path, content string) error {
	fullPath := filepath.Join(r.Path, path)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// Commit creates a commit with the given message and author.
func (r *TestRepo) Commit(message, authorName, authorEmail string) error {
	wt, err := r.Repo.Worktree()
	if err != nil {
		return err
	}

	// Add all files
	_, err = wt.Add(".")
	if err != nil {
		return err
	}

	// Commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	return err
}

// CommitWithTime creates a commit with a specific time.
func (r *TestRepo) CommitWithTime(message, authorName, authorEmail string, when time.Time) error {
	wt, err := r.Repo.Worktree()
	if err != nil {
		return err
	}

	// Add all files
	_, err = wt.Add(".")
	if err != nil {
		return err
	}

	// Commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  when,
		},
	})
	return err
}

// CreateMarkdownFile creates a Markdown file and commits it.
func (r *TestRepo) CreateMarkdownFile(path, content, message, author string) error {
	// Create file
	if err := r.CreateFile(path, content); err != nil {
		return err
	}

	// Commit
	return r.Commit(message, author, author+"@example.com")
}

// CreateMarkdownFileWithTime creates a Markdown file with a specific commit time.
func (r *TestRepo) CreateMarkdownFileWithTime(path, content, message, author string, when time.Time) error {
	// Create file
	if err := r.CreateFile(path, content); err != nil {
		return err
	}

	// Commit
	return r.CommitWithTime(message, author, author+"@example.com", when)
}

// CreateImageFile creates an image file.
func (r *TestRepo) CreateImageFile(path string, data []byte) error {
	fullPath := filepath.Join(r.Path, path)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file
	return os.WriteFile(fullPath, data, 0644)
}

// CreateImageFileAndCommit creates an image file and commits it.
func (r *TestRepo) CreateImageFileAndCommit(path string, data []byte, message, author string) error {
	if err := r.CreateImageFile(path, data); err != nil {
		return err
	}
	return r.Commit(message, author, author+"@example.com")
}

// SetupMultiAuthorArticle creates an article with multiple commits from different authors.
func (r *TestRepo) SetupMultiAuthorArticle(path, content string) error {
	// Create initial file
	if err := r.CreateFile(path, content); err != nil {
		return err
	}

	// First commit (creator)
	if err := r.CommitWithTime("Initial commit", "creator", "creator@example.com", time.Now().Add(-48*time.Hour)); err != nil {
		return err
	}

	// Update file
	if err := r.CreateFile(path, content+"\n\n## Update 1\nAdded by editor1."); err != nil {
		return err
	}

	// Second commit (editor1)
	if err := r.CommitWithTime("Update by editor1", "editor1", "editor1@example.com", time.Now().Add(-24*time.Hour)); err != nil {
		return err
	}

	// Update file again
	if err := r.CreateFile(path, content+"\n\n## Update 1\nAdded by editor1.\n\n## Update 2\nAdded by editor2."); err != nil {
		return err
	}

	// Third commit (editor2)
	return r.Commit("Update by editor2", "editor2", "editor2@example.com")
}

// SetupNestedDirectory creates nested directories with files.
func (r *TestRepo) SetupNestedDirectory() error {
	// Create tech directory
	if err := r.CreateFile("tech/golang.md", "# Golang\nGolang content."); err != nil {
		return err
	}
	if err := r.CreateFile("tech/rust.md", "# Rust\nRust content."); err != nil {
		return err
	}

	// Create life directory
	if err := r.CreateFile("life/travel.md", "# Travel\nTravel content."); err != nil {
		return err
	}

	// Create root files
	if err := r.CreateFile("welcome.md", "# Welcome\nWelcome content."); err != nil {
		return err
	}

	return r.Commit("Initial commit", "admin", "admin@example.com")
}

// CreateUncommittedFile creates a file without committing it.
func (r *TestRepo) CreateUncommittedFile(path, content string) error {
	return r.CreateFile(path, content)
}
