// Package service provides business logic services for the application.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitInitService provides Git initialization and validation services.
type GitInitService struct{}

// NewGitInitService creates a new GitInitService.
func NewGitInitService() *GitInitService {
	return &GitInitService{}
}

// EnsureGitRepo ensures the directory is a Git repository.
// If it's not, it initializes a new repository.
func (s *GitInitService) EnsureGitRepo(dir string, autoInit bool) error {
	// Get absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, create it
			if err := os.MkdirAll(absPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to stat directory: %w", err)
		}
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Check if .git exists
	gitDir := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Already a Git repository, verify it's valid
		_, err := git.PlainOpen(absPath)
		if err != nil {
			return fmt.Errorf("invalid git repository: %w", err)
		}
		return nil
	}

	// Not a Git repository
	if !autoInit {
		return fmt.Errorf("directory is not a git repository: %s (set --init to auto-initialize)", absPath)
	}

	// Initialize Git repository
	return s.InitGitRepo(absPath)
}

// InitGitRepo initializes a new Git repository in the given directory.
func (s *GitInitService) InitGitRepo(dir string) error {
	// Use go-git to initialize
	_, err := git.PlainInit(dir, false)
	if err != nil {
		// Fallback to system git command
		cmd := exec.Command("git", "init")
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to initialize git repository: %w, output: %s", err, output)
		}
	}

	// Create initial commit (required for some operations)
	return s.createInitialCommit(dir)
}

// createInitialCommit creates an initial empty commit.
func (s *GitInitService) createInitialCommit(dir string) error {
	// Create a README.md file
	readmePath := filepath.Join(dir, "README.md")
	readmeContent := "# Terminalog Blog\n\nThis is a Terminalog blog repository.\n\nAdd your articles as Markdown files here.\n"

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	// Use system git to add and commit
	cmdAdd := exec.Command("git", "add", "README.md")
	cmdAdd.Dir = dir
	if output, err := cmdAdd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to git add: %w, output: %s", err, output)
	}

	cmdCommit := exec.Command("git", "commit", "-m", "Initial commit: Terminalog blog")
	cmdCommit.Dir = dir
	if _, err := cmdCommit.CombinedOutput(); err != nil {
		// If commit fails, it might be because no user name/email is configured
		// Try with default values
		cmdConfigName := exec.Command("git", "config", "user.name", "Terminalog")
		cmdConfigName.Dir = dir
		cmdConfigName.Run()

		cmdConfigEmail := exec.Command("git", "config", "user.email", "terminalog@localhost")
		cmdConfigEmail.Dir = dir
		cmdConfigEmail.Run()

		// Retry commit
		cmdCommit2 := exec.Command("git", "commit", "-m", "Initial commit: Terminalog blog")
		cmdCommit2.Dir = dir
		if output, err := cmdCommit2.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to git commit: %w, output: %s", err, output)
		}
	}

	return nil
}

// IsGitRepo checks if a directory is a valid Git repository.
func (s *GitInitService) IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return false
	}

	_, err := git.PlainOpen(dir)
	return err == nil
}

// GetRepoStatus returns the status of a Git repository.
func (s *GitInitService) GetRepoStatus(dir string) (RepoStatus, error) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return RepoStatus{}, err
	}

	head, err := repo.Head()
	if err != nil {
		return RepoStatus{IsRepo: true, HasCommits: false}, nil
	}

	// Count commits
	commitIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return RepoStatus{IsRepo: true, HasCommits: false}, nil
	}

	commitCount := 0
	commitIter.ForEach(func(c *object.Commit) error {
		commitCount++
		return nil
	})

	return RepoStatus{
		IsRepo:        true,
		HasCommits:    commitCount > 0,
		CommitCount:   commitCount,
		CurrentBranch: head.Name().Short(),
	}, nil
}

// RepoStatus represents the status of a Git repository.
type RepoStatus struct {
	IsRepo        bool   `json:"isRepo"`
	HasCommits    bool   `json:"hasCommits"`
	CommitCount   int    `json:"commitCount"`
	CurrentBranch string `json:"currentBranch"`
}
