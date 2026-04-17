// Package service provides business logic services for the application.
package service

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"terminalog/internal/model"
	"terminalog/pkg/utils"
)

// SpecialFilePrefix is the prefix for special files that should be excluded from article lists.
// Files starting with this prefix are reserved for special purposes (e.g., _ABOUTME.md).
const SpecialFilePrefix = "_"

// AssetsDirName is the name of the assets directory that should be excluded from article lists.
// Assets directories are hidden directories for storing images and other static resources.
const AssetsDirName = ".assets"

// FileService provides file system operations.
type FileService struct {
	// baseDir is the absolute path to the content directory.
	baseDir string
}

// NewFileService creates a new FileService instance.
func NewFileService(baseDir string) (*FileService, error) {
	absPath, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, err
	}

	return &FileService{baseDir: absPath}, nil
}

// ScanMarkdownFiles scans the directory for all Markdown files.
// It returns a list of relative paths to Markdown files.
// Files starting with "_" (SpecialFilePrefix) are excluded.
func (s *FileService) ScanMarkdownFiles(ctx context.Context, dir string) ([]string, error) {
	// Validate and get absolute path
	absDir, err := utils.ValidatePath(s.baseDir, dir)
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	if !info.IsDir() {
		return nil, model.ErrNotFound
	}

	// Walk the directory
	files := make([]string, 0)

	err = filepath.WalkDir(absDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if strings.Contains(path, "/.git/") || d.Name() == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip .assets directory (hidden assets storage)
		if d.Name() == AssetsDirName {
			if d.IsDir() {
				return filepath.SkipDir // Skip entire .assets directory
			}
			return nil
		}

		// Skip special files (starting with "_")
		if strings.HasPrefix(d.Name(), SpecialFilePrefix) {
			if d.IsDir() {
				return filepath.SkipDir // Skip entire special directory
			}
			return nil // Skip special file
		}

		// Only include Markdown files
		if !d.IsDir() && utils.IsMarkdownFile(path) {
			// Convert to relative path
			relPath, err := filepath.Rel(s.baseDir, path)
			if err != nil {
				return err
			}
			files = append(files, utils.NormalizePath(relPath))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files alphabetically
	sort.Strings(files)

	return files, nil
}

// ReadSpecialFile reads a special file (e.g., _ABOUTME.md).
// This method allows reading files that start with "_" prefix.
func (s *FileService) ReadSpecialFile(ctx context.Context, filename string) ([]byte, error) {
	// Validate that the filename starts with special prefix
	if !strings.HasPrefix(filename, SpecialFilePrefix) {
		return nil, model.ErrNotFound
	}

	// Validate path (but allow special files)
	absPath, err := utils.ValidatePath(s.baseDir, filename)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	if info.IsDir() {
		return nil, model.ErrNotFound
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// ReadFile reads the content of a file.
func (s *FileService) ReadFile(ctx context.Context, path string) ([]byte, error) {
	// Validate path
	absPath, err := utils.ValidatePath(s.baseDir, path)
	if err != nil {
		return nil, err
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	if info.IsDir() {
		return nil, model.ErrNotFound
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// GetDirectoryTree returns the directory tree structure.
func (s *FileService) GetDirectoryTree(ctx context.Context, dir string) (*model.TreeNode, error) {
	// Validate and get absolute path
	absDir, err := utils.ValidatePath(s.baseDir, dir)
	if err != nil {
		return nil, err
	}

	// Handle empty dir (root)
	if dir == "" || dir == "/" {
		absDir = s.baseDir
	}

	// Check if directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	if !info.IsDir() {
		return nil, model.ErrNotFound
	}

	// Build tree
	root := &model.TreeNode{
		Name:     filepath.Base(absDir),
		Path:     utils.NormalizePath(dir),
		Type:     model.NodeTypeDir,
		Children: make([]*model.TreeNode, 0),
	}

	// Read directory entries
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	// Sort entries (directories first, then files, alphabetically)
	sort.Slice(entries, func(i, j int) bool {
		// Directories first
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		// Alphabetically
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		// Skip .git directory
		if entry.Name() == ".git" {
			continue
		}

		// Skip .assets directory (hidden assets storage)
		if entry.Name() == AssetsDirName {
			continue
		}

		childPath := utils.NormalizePath(dir + "/" + entry.Name())

		if entry.IsDir() {
			// Recursively get subtree
			child, err := s.GetDirectoryTree(ctx, childPath)
			if err != nil {
				continue // Skip inaccessible directories
			}
			root.Children = append(root.Children, child)
		} else {
			// Add file node (only Markdown files)
			if utils.IsMarkdownFile(entry.Name()) {
				root.Children = append(root.Children, &model.TreeNode{
					Name: entry.Name(),
					Path: childPath,
					Type: model.NodeTypeFile,
				})
			}
		}
	}

	return root, nil
}

// FileExists checks if a file exists.
func (s *FileService) FileExists(ctx context.Context, path string) (bool, error) {
	absPath, err := utils.ValidatePath(s.baseDir, path)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

// ValidatePath validates and returns the absolute path for a requested path.
func (s *FileService) ValidatePath(requestedPath string) (string, error) {
	return utils.ValidatePath(s.baseDir, requestedPath)
}

// GetBaseDir returns the base directory path.
func (s *FileService) GetBaseDir() string {
	return s.baseDir
}
