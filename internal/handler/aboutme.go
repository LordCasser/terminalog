// Package handler provides HTTP request handlers for the application.
package handler

import (
	"net/http"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// AboutMeHandler handles About Me related HTTP requests.
type AboutMeHandler struct {
	svc *service.AboutMeService
}

// NewAboutMeHandler creates a new AboutMeHandler instance.
func NewAboutMeHandler(svc *service.AboutMeService) *AboutMeHandler {
	return &AboutMeHandler{svc: svc}
}

// Get handles GET /api/v1/special/aboutme.
// It reads and returns the content of _ABOUTME.md.
func (h *AboutMeHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	content, exists, err := h.svc.Get(ctx)
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.AboutMeResponse{
		Path:    service.AboutMeFilename,
		Title:   "About Me",
		Content: string(content),
		Exists:  exists,
	})
}
