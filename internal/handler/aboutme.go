// Package handler provides HTTP request handlers for the application.
package handler

import (
	"errors"
	"net/http"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// AboutMeFilename is the filename for the About Me page.
const AboutMeFilename = "_ABOUTME.md"

// AboutMeHandler handles About Me related HTTP requests.
type AboutMeHandler struct {
	fileSvc *service.FileService
}

// NewAboutMeHandler creates a new AboutMeHandler instance.
func NewAboutMeHandler(fileSvc *service.FileService) *AboutMeHandler {
	return &AboutMeHandler{fileSvc: fileSvc}
}

// Get handles GET /api/aboutme.
// It reads and returns the content of _ABOUTME.md.
func (h *AboutMeHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read the special file
	content, err := h.fileSvc.ReadSpecialFile(ctx, AboutMeFilename)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			utils.RespondNotFound(w, "About Me not found")
		default:
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	// Extract title from filename (remove .md extension and _ prefix)
	title := "About Me"

	// Respond
	utils.RespondJSON(w, http.StatusOK, model.AboutMeResponse{
		Path:    AboutMeFilename,
		Title:   title,
		Content: string(content),
	})
}
