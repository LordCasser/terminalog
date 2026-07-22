// Package handler provides HTTP request handlers for the application.
package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// AssetHandler handles asset-related HTTP requests.
type AssetHandler struct {
	svc *service.AssetService
}

// NewAssetHandler creates a new AssetHandler instance.
func NewAssetHandler(svc *service.AssetService) *AssetHandler {
	return &AssetHandler{svc: svc}
}

// Get handles GET /api/assets/*.
func (h *AssetHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path from URL (using wildcard pattern)
	path := chi.URLParam(r, "*")

	// Get asset
	asset, err := h.svc.GetAsset(ctx, path)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			utils.RespondNotFound(w, "Asset not found")
		case errors.Is(err, model.ErrInvalidPath):
			utils.RespondBadRequest(w, "Invalid path")
		default:
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	// Set headers
	w.Header().Set("Content-Type", asset.ContentType)
	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("ETag", asset.ETag)
	if r.Header.Get("If-None-Match") == asset.ETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", asset.Size))

	// Write content
	w.Write(asset.Data)
}
