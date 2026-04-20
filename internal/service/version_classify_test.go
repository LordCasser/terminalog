// Package service tests for internal version classification logic.
// This file uses package service (not service_test) to directly access
// internal functions like classifyChange and formatVersion without
// duplicating their implementation.
package service

import (
	"testing"

	"terminalog/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestClassifyChange(t *testing.T) {
	tests := []struct {
		name          string
		linesChanged  int
		prevFileLines int
		wantType      model.ChangeType
	}{
		{
			name:          "patch - less than 10 lines",
			linesChanged:  5,
			prevFileLines: 100,
			wantType:      model.ChangeTypePatch,
		},
		{
			name:          "patch - exactly 9 lines",
			linesChanged:  9,
			prevFileLines: 100,
			wantType:      model.ChangeTypePatch,
		},
		{
			name:          "minor - 10 lines changed (10% of 100)",
			linesChanged:  10,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "minor - 40% changed",
			linesChanged:  40,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "minor - 50% exactly (not > 50%)",
			linesChanged:  50,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "major - 51% changed",
			linesChanged:  51,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMajor,
		},
		{
			name:          "major - 100% changed",
			linesChanged:  100,
			prevFileLines: 100,
			wantType:      model.ChangeTypeMajor,
		},
		{
			name:          "zero previous lines - small change is patch",
			linesChanged:  5,
			prevFileLines: 0,
			wantType:      model.ChangeTypePatch,
		},
		{
			name:          "zero previous lines - moderate change is minor",
			linesChanged:  50,
			prevFileLines: 0,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "zero previous lines - large change is major",
			linesChanged:  100,
			prevFileLines: 0,
			wantType:      model.ChangeTypeMajor,
		},
		{
			name:          "small file - 10 lines changed in 20-line file (50%)",
			linesChanged:  10,
			prevFileLines: 20,
			wantType:      model.ChangeTypeMinor,
		},
		{
			name:          "small file - 11 lines changed in 20-line file (55%)",
			linesChanged:  11,
			prevFileLines: 20,
			wantType:      model.ChangeTypeMajor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyChange(tt.linesChanged, tt.prevFileLines)
			assert.Equal(t, tt.wantType, result)
		})
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name  string
		major int
		minor int
		patch int
		want  string
	}{
		{
			name:  "initial version",
			major: 1,
			minor: 0,
			patch: 0,
			want:  "v1.0.0",
		},
		{
			name:  "patch bump",
			major: 1,
			minor: 0,
			patch: 1,
			want:  "v1.0.1",
		},
		{
			name:  "minor bump",
			major: 1,
			minor: 1,
			patch: 0,
			want:  "v1.1.0",
		},
		{
			name:  "major bump",
			major: 2,
			minor: 0,
			patch: 0,
			want:  "v2.0.0",
		},
		{
			name:  "complex version",
			major: 2,
			minor: 3,
			patch: 48,
			want:  "v2.3.48",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVersion(tt.major, tt.minor, tt.patch)
			assert.Equal(t, tt.want, result)
		})
	}
}