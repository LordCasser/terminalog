// Package service provides business logic services for the application.
package service

import (
	"context"
	"fmt"
	"strings"

	"terminalog/internal/model"
)

// VersionService provides version-related operations.
type VersionService struct {
	articleSvc *ArticleService
	gitSvc     *GitService
	fileSvc    *FileService
}

// NewVersionService creates a new VersionService instance.
func NewVersionService(articleSvc *ArticleService, gitSvc *GitService, fileSvc *FileService) *VersionService {
	return &VersionService{
		articleSvc: articleSvc,
		gitSvc:     gitSvc,
		fileSvc:    fileSvc,
	}
}

// VersionRule constants for version number calculation.
const (
	// PatchThreshold: changes < 10 lines trigger a patch version bump.
	PatchThreshold = 10

	// MinorThresholdPercent: changes 10-50% of total lines trigger a minor version bump.
	MinorThresholdPercent = 0.5 // 50%

	// InitialVersion is the starting version for new files.
	InitialVersion = "v1.0.0"
)

// GetVersion returns the version information for an article.
// Version numbers are calculated based on line count changes across commits.
func (s *VersionService) GetVersion(ctx context.Context, path string) (*model.VersionInfo, error) {
	// Normalize path
	path = strings.TrimPrefix(path, "/")

	// Get file history
	history, err := s.gitSvc.GetFileHistory(ctx, path)
	if err != nil {
		return nil, err
	}

	// Get current content line count
	currentContent, err := s.fileSvc.ReadFile(ctx, path)
	if err != nil {
		return nil, err
	}
	currentLines := countLines(string(currentContent))

	// Build version history
	versionHistory := make([]model.VersionHistoryEntry, 0)

	// Start from initial version
	major, minor, patch := 1, 0, 0

	// Process commits in chronological order (oldest to newest)
	// We need to reverse the AllCommits slice
	commits := reverseCommits(history.AllCommits)

	for i, commit := range commits {
		// Get the content at this commit (simplified: use current content for now)
		// In a real implementation, we would extract content from each commit
		// For MVP, we estimate lines changed based on commit position

		var linesChanged int
		var changeType model.ChangeType

		if i == 0 {
			// First commit: initial version
			linesChanged = currentLines
			changeType = model.ChangeTypeMajor
			major = 1
			minor = 0
			patch = 0
		} else {
			// Estimate lines changed (simplified for MVP)
			// In production, we would diff between commits
			linesChanged = estimateLinesChanged(i, len(commits), currentLines)
			changeType = classifyChange(linesChanged, currentLines)

			// Update version based on change type
			switch changeType {
			case model.ChangeTypePatch:
				patch++
			case model.ChangeTypeMinor:
				minor++
				patch = 0
			case model.ChangeTypeMajor:
				major++
				minor = 0
				patch = 0
			}
		}

		versionHistory = append(versionHistory, model.VersionHistoryEntry{
			Version:      formatVersion(major, minor, patch),
			Hash:         commit.Hash,
			Author:       commit.Author,
			Timestamp:    commit.Timestamp,
			LinesChanged: linesChanged,
			ChangeType:   changeType,
		})
	}

	// Reverse history to show newest first
	versionHistory = reverseVersionHistory(versionHistory)

	// Current version is the last calculated
	currentVersion := formatVersion(major, minor, patch)

	return &model.VersionInfo{
		CurrentVersion: currentVersion,
		History:        versionHistory,
	}, nil
}

// classifyChange determines the change type based on lines changed.
func classifyChange(linesChanged, totalLines int) model.ChangeType {
	if linesChanged < PatchThreshold {
		return model.ChangeTypePatch
	}

	// Calculate percentage of total lines
	if totalLines == 0 {
		return model.ChangeTypePatch
	}

	percent := float64(linesChanged) / float64(totalLines)
	if percent > MinorThresholdPercent {
		return model.ChangeTypeMajor
	}

	return model.ChangeTypeMinor
}

// formatVersion formats major, minor, patch into a version string.
func formatVersion(major, minor, patch int) string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

// countLines counts the number of lines in a string.
func countLines(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Split(content, "\n"))
}

// estimateLinesChanged estimates lines changed for a commit.
// This is a simplified estimation for MVP.
func estimateLinesChanged(commitIndex, totalCommits, totalLines int) int {
	// Simplified estimation: distribute changes evenly across commits
	if totalCommits <= 1 {
		return totalLines
	}

	// Estimate based on commit position
	avgChange := totalLines / (totalCommits - 1)
	if avgChange < 1 {
		avgChange = 1
	}

	return avgChange
}

// reverseCommits reverses a slice of CommitInfo.
func reverseCommits(commits []model.CommitInfo) []model.CommitInfo {
	result := make([]model.CommitInfo, len(commits))
	for i, j := 0, len(commits)-1; i < len(commits); i, j = i+1, j-1 {
		result[i] = commits[j]
	}
	return result
}

// reverseVersionHistory reverses a slice of VersionHistoryEntry.
func reverseVersionHistory(history []model.VersionHistoryEntry) []model.VersionHistoryEntry {
	result := make([]model.VersionHistoryEntry, len(history))
	for i, j := 0, len(history)-1; i < len(history); i, j = i+1, j-1 {
		result[i] = history[j]
	}
	return result
}
