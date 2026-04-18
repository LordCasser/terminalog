// Package handler provides HTTP request handlers for the application.
package handler

import (
	"net/http"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// TreeHandler handles directory tree-related HTTP requests.
type TreeHandler struct {
	svc *service.ArticleService
}

// NewTreeHandler creates a new TreeHandler instance.
func NewTreeHandler(svc *service.ArticleService) *TreeHandler {
	return &TreeHandler{svc: svc}
}

// Get handles GET /api/v1/tree.
func (h *TreeHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get directory parameter (optional)
	dir := r.URL.Query().Get("dir")

	// Get tree
	tree, err := h.svc.GetTree(ctx, dir)
	if err != nil {
		// If the directory doesn't exist, return root tree
		if dir != "" {
			tree, err = h.svc.GetTree(ctx, "")
			if err != nil {
				utils.RespondInternalServerError(w, err.Error())
				return
			}
		} else {
			utils.RespondInternalServerError(w, err.Error())
			return
		}
	}

	// Respond
	utils.RespondJSON(w, http.StatusOK, model.TreeResponse{
		Root:       tree,
		CurrentDir: dir,
	})
}
