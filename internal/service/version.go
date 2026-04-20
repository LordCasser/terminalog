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
	// PatchThreshold: total changed lines (added + removed) < 10 trigger a patch version bump.
	PatchThreshold = 10

	// MinorThresholdPercent: changes exceeding this fraction of the file trigger a major bump.
	// Changes between PatchThreshold and MinorThresholdPercent trigger a minor bump.
	MinorThresholdPercent = 0.5 // 50%

	// InitialVersion is the starting version for new files.
	InitialVersion = "v1.0.0"
)

// GetVersion returns the version information for an article.
// Version numbers are calculated from real git diff statistics across commits.
func (s *VersionService) GetVersion(ctx context.Context, path string) (*model.VersionInfo, error) {
	// Normalize path
	path = strings.TrimPrefix(path, "/")

	// Get file history for commit metadata
	history, err := s.gitSvc.GetFileHistory(ctx, path)
	if err != nil {
		return nil, err
	}

	// Get real diff statistics for each commit
	diffs, err := s.gitSvc.GetFileCommitDiffs(ctx, path)
	if err != nil {
		return nil, err
	}

	// Build a map from commit hash to diff info for O(1) lookup
	diffMap := make(map[string]*model.CommitDiffInfo, len(diffs))
	for i := range diffs {
		diffMap[diffs[i].Hash] = &diffs[i]
	}

	// Commits from history are newest-first; reverse to chronological order
	commits := reverseCommits(history.AllCommits)

	// Build version history
	versionHistory := make([]model.VersionHistoryEntry, 0)

	// Start from initial version
	major, minor, patch := 1, 0, 0

	for i, commit := range commits {
		var linesAdded, linesRemoved, linesChanged int
		var fileLinesAfter int
		var changeType model.ChangeType

		diff, hasDiff := diffMap[commit.Hash]

		if i == 0 {
			// First commit: file creation
			if hasDiff {
				linesAdded = diff.LinesAdded
				linesRemoved = diff.LinesRemoved
				fileLinesAfter = diff.FileLinesAfter
			} else {
				// Fallback: treat as creation of current file size
				currentContent, err := s.fileSvc.ReadFile(ctx, path)
				if err != nil {
					return nil, err
				}
				fileLinesAfter = countLines(string(currentContent))
				linesAdded = fileLinesAfter
			}
			linesChanged = linesAdded + linesRemoved
			changeType = model.ChangeTypeMajor // File creation is always major
			major = 1
			minor = 0
			patch = 0
		} else {
			// Subsequent commits: use real diff data
			if hasDiff {
				linesAdded = diff.LinesAdded
				linesRemoved = diff.LinesRemoved
				fileLinesAfter = diff.FileLinesAfter
			}
			linesChanged = linesAdded + linesRemoved

			// Use the file's line count after the PREVIOUS commit as the denominator.
			// This represents "how big was the file before this change" — which is the
			// correct baseline for measuring the significance of this change.
			var prevFileLines int
			prevCommit := commits[i-1]
			if prevDiff, ok := diffMap[prevCommit.Hash]; ok {
				prevFileLines = prevDiff.FileLinesAfter
			}
			// Fallback: if we don't have the previous file size, use the current file size
			if prevFileLines == 0 {
				prevFileLines = fileLinesAfter
			}

			changeType = classifyChange(linesChanged, prevFileLines)

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
			Version:       formatVersion(major, minor, patch),
			Hash:          commit.Hash,
			Author:        commit.Author,
			Timestamp:     commit.Timestamp,
			LinesAdded:    linesAdded,
			LinesRemoved:  linesRemoved,
			LinesChanged:  linesChanged,
			ChangeType:    changeType,
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

// classifyChange determines the change type based on lines changed relative to file size.
// prevFileLines is the total line count of the file BEFORE this change, used as the denominator
// to calculate the percentage of change. If the file didn't exist before (prevFileLines == 0),
// the change is classified based solely on absolute lines changed.
func classifyChange(linesChanged, prevFileLines int) model.ChangeType {
	if linesChanged < PatchThreshold {
		return model.ChangeTypePatch
	}

	// If we don't have a previous file size (e.g., file was created from nothing),
	// fall back to absolute thresholds for classification.
	if prevFileLines == 0 {
		// File creation or unknown previous state: if linesChanged >= PatchThreshold,
		// it's at least minor. Large additions are major.
		if linesChanged >= 100 {
			return model.ChangeTypeMajor
		}
		return model.ChangeTypeMinor
	}

	// Calculate percentage relative to the file size before this change
	percent := float64(linesChanged) / float64(prevFileLines)
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
// (Defined in git.go; this is a duplicate for the version package's internal use.)
// If needed, this can be moved to a shared utility.

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