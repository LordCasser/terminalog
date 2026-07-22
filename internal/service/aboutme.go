// Package service provides business logic services for the application.
package service

import (
	"context"
	"errors"

	"terminalog/internal/model"
)

// AboutMeFilename is the committed special file exposed by the About Me API.
const AboutMeFilename = "_ABOUTME.md"

// AboutMeService keeps special-page visibility aligned with the repository's
// current HEAD rather than the mutable worktree alone.
type AboutMeService struct {
	gitSvc *GitService
}

func NewAboutMeService(gitSvc *GitService) *AboutMeService {
	return &AboutMeService{gitSvc: gitSvc}
}

func (s *AboutMeService) Get(_ context.Context) ([]byte, bool, error) {
	content, err := s.gitSvc.ReadFileAtHead(AboutMeFilename)
	if errors.Is(err, model.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return content, true, nil
}
