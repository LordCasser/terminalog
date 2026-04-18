// Package handler provides HTTP request handlers for the application.
package handler

import (
	"net/http"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// SearchHandler handles search-related HTTP requests.
type SearchHandler struct {
	svc *service.ArticleService
}

// NewSearchHandler creates a new SearchHandler instance.
func NewSearchHandler(svc *service.ArticleService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Search handles GET /api/v1/search.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.RespondBadRequest(w, "Query parameter 'q' is required")
		return
	}

	// Get directory parameter (optional)
	dir := r.URL.Query().Get("dir")

	// Search
	results, err := h.svc.Search(ctx, query, dir)
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	// Respond
	utils.RespondJSON(w, http.StatusOK, model.SearchResponse{
		Results: results,
		Query:   query,
		Total:   len(results),
	})
}
