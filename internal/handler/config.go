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

	// filing holds regulatory filing information.
	filing FilingInfo
}

// FilingInfo holds Chinese regulatory filing information.
type FilingInfo struct {
	// ICPFiling is the ICP filing number (ICP备案号).
	ICPFiling string

	// ICPFilingURL is the URL for ICP filing verification.
	ICPFilingURL string

	// PoliceFiling is the public security filing number (公安备案号).
	PoliceFiling string

	// PoliceFilingURL is the URL for the public security filing record.
	PoliceFilingURL string
}

// NewConfigHandler creates a new ConfigHandler.
func NewConfigHandler(owner string, filing FilingInfo) *ConfigHandler {
	return &ConfigHandler{
		owner:  owner,
		filing: filing,
	}
}

// FrontendConfig represents the configuration response for frontend.
type FrontendConfig struct {
	// Owner is the blog owner name displayed in navbar.
	Owner string `json:"owner"`

	// ICPFiling is the ICP filing number for regulatory compliance in mainland China.
	ICPFiling string `json:"icp_filing,omitempty"`

	// ICPFilingURL is the verification URL for the ICP filing.
	ICPFilingURL string `json:"icp_filing_url,omitempty"`

	// PoliceFiling is the public security filing number.
	PoliceFiling string `json:"police_filing,omitempty"`

	// PoliceFilingURL is the verification URL for the public security filing.
	PoliceFilingURL string `json:"police_filing_url,omitempty"`
}

// Get handles GET /api/v1/settings - returns frontend configuration.
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := FrontendConfig{
		Owner:           h.owner,
		ICPFiling:       h.filing.ICPFiling,
		ICPFilingURL:    h.filing.ICPFilingURL,
		PoliceFiling:    h.filing.PoliceFiling,
		PoliceFilingURL: h.filing.PoliceFilingURL,
	}

	json.NewEncoder(w).Encode(response)
}
