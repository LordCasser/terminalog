// Package handler provides HTTP request handlers for the application.
package handler

import (
	"encoding/base64"
	"net/http"
	"strings"

	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// GitHandler handles Git Smart HTTP requests.
type GitHandler struct {
	gitSvc  *service.GitService
	authSvc *service.AuthService
}

// NewGitHandler creates a new GitHandler instance.
func NewGitHandler(gitSvc *service.GitService, authSvc *service.AuthService) *GitHandler {
	return &GitHandler{
		gitSvc:  gitSvc,
		authSvc: authSvc,
	}
}

// UploadPackRefs handles GET /info/refs?service=git-upload-pack (Clone refs).
func (h *GitHandler) UploadPackRefs(w http.ResponseWriter, r *http.Request) {
	// Verify service parameter
	service := r.URL.Query().Get("service")
	if service != "git-upload-pack" {
		utils.RespondBadRequest(w, "Invalid service")
		return
	}

	// Get refs
	refs, err := h.gitSvc.GetUploadPackRefs(r.Context())
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	// Set Git Smart HTTP headers
	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	w.Header().Set("Cache-Control", "no-cache")

	// Write response
	w.Write(refs)
}

// UploadPack handles POST /git-upload-pack (Clone packfile).
func (h *GitHandler) UploadPack(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")

	// Handle upload-pack
	result, err := h.gitSvc.HandleUploadPack(r.Context(), r.Body)
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	// Write result
	w.Write(result)
}

// ReceivePackRefs handles GET /info/refs?service=git-receive-pack (Push refs).
// This endpoint requires authentication.
func (h *GitHandler) ReceivePackRefs(w http.ResponseWriter, r *http.Request) {
	// Verify service parameter
	service := r.URL.Query().Get("service")
	if service != "git-receive-pack" {
		utils.RespondBadRequest(w, "Invalid service")
		return
	}

	// Extract and validate authentication
	auth := h.extractAuth(r)
	if auth == nil {
		utils.RespondUnauthorized(w, "Authentication required")
		return
	}

	valid, err := h.authSvc.Validate(auth.Username, auth.Password)
	if err != nil || !valid {
		utils.RespondUnauthorized(w, "Invalid credentials")
		return
	}

	// Get refs
	refs, err := h.gitSvc.GetReceivePackRefs(r.Context())
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	// Set Git Smart HTTP headers
	w.Header().Set("Content-Type", "application/x-git-receive-pack-advertisement")
	w.Header().Set("Cache-Control", "no-cache")

	// Write response
	w.Write(refs)
}

// ReceivePack handles POST /git-receive-pack (Push data).
// This endpoint requires authentication.
func (h *GitHandler) ReceivePack(w http.ResponseWriter, r *http.Request) {
	// Extract and validate authentication (must be provided in POST too)
	auth := h.extractAuth(r)
	if auth == nil {
		utils.RespondUnauthorized(w, "Authentication required")
		return
	}

	valid, err := h.authSvc.Validate(auth.Username, auth.Password)
	if err != nil || !valid {
		utils.RespondUnauthorized(w, "Invalid credentials")
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")

	// Handle receive-pack
	result, err := h.gitSvc.HandleReceivePack(r.Context(), r.Body)
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	// Write result
	w.Write(result)
}

// InfoRefs handles GET /info/refs (router determines which service).
func (h *GitHandler) InfoRefs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")

	switch service {
	case "git-upload-pack":
		h.UploadPackRefs(w, r)
	case "git-receive-pack":
		h.ReceivePackRefs(w, r)
	default:
		// Default to upload-pack (clone) for compatibility
		if service == "" {
			h.UploadPackRefs(w, r)
		} else {
			utils.RespondBadRequest(w, "Invalid service")
		}
	}
}

// Helper: extractAuth extracts Basic Auth credentials from the request.
type authInfo struct {
	Username string
	Password string
}

func (h *GitHandler) extractAuth(r *http.Request) *authInfo {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	// Check for Basic auth
	if !strings.HasPrefix(authHeader, "Basic ") {
		return nil
	}

	// Decode credentials
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}

	// Split username:password
	credentials := string(decoded)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	return &authInfo{
		Username: parts[0],
		Password: parts[1],
	}
}
