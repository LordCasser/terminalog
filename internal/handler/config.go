// Package handler provides HTTP request handlers for the application.
package handler

import (
	"encoding/json"
	"net/http"
)

// ConfigHandler handles configuration requests for frontend.
type ConfigHandler struct {
	// owner is the blog owner name.
	owner string
}

// NewConfigHandler creates a new ConfigHandler.
func NewConfigHandler(owner string) *ConfigHandler {
	return &ConfigHandler{
		owner: owner,
	}
}

// FrontendConfig represents the configuration response for frontend.
type FrontendConfig struct {
	// Owner is the blog owner name displayed in navbar.
	Owner string `json:"owner"`
}

// Get handles GET /api/v1/settings - returns frontend configuration.
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := FrontendConfig{
		Owner: h.owner,
	}

	json.NewEncoder(w).Encode(response)
}
