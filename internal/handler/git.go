// Package handler provides HTTP request handlers for the application.
package handler

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"terminalog/internal/service"
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

// InfoRefs handles GET /info/refs?service=git-{upload-pack|receive-pack}.
// This is the reference advertisement endpoint for the Git Smart HTTP protocol.
// It runs `git {service} --stateless-rpc --advertise-refs .` as a subprocess.
func (h *GitHandler) InfoRefs(w http.ResponseWriter, r *http.Request) {
	serviceParam := r.URL.Query().Get("service")

	// Determine the git service type
	var gitService string
	switch serviceParam {
	case "git-upload-pack":
		gitService = service.ServiceTypeUploadPack
	case "git-receive-pack":
		gitService = service.ServiceTypeReceivePack
	default:
		// No service parameter or unknown service
		http.Error(w, "Invalid service", http.StatusBadRequest)
		return
	}

	// Require authentication for receive-pack (push)
	if gitService == service.ServiceTypeReceivePack {
		auth := h.extractAuth(r)
		if auth == nil {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		valid, err := h.authSvc.Validate(auth.Username, auth.Password)
		if err != nil || !valid {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	// Get refs advertisement from git subprocess
	refs, err := h.gitSvc.GetInfoRefs(gitService)
	if err != nil {
		log.Printf("InfoRefs: failed to get refs for service %s: %v", gitService, err)
		http.Error(w, "Failed to get refs", http.StatusInternalServerError)
		return
	}

	// Set response headers
	contentType := fmt.Sprintf("application/x-git-%s-advertisement", gitService)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")

	// Write pkt-line service announcement header
	// Format: pkt-line("# service=git-{service}\n") + flush + refs data
	pktLine(w, fmt.Sprintf("# service=git-%s\n", gitService))
	pktFlush(w)

	// Write refs data from git subprocess
	w.Write(refs)
}

// UploadPack handles POST /git-upload-pack (Clone/Fetch).
// This pipes the HTTP request body to `git upload-pack --stateless-rpc .`
// and streams the response back directly.
func (h *GitHandler) UploadPack(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type
	if r.Header.Get("Content-Type") != "application/x-git-upload-pack-request" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// Get request body, handling gzip if needed
	reqBody, err := h.getRequestBody(r)
	if err != nil {
		log.Printf("UploadPack: failed to decompress request body: %v", err)
		http.Error(w, "Failed to decompress request", http.StatusInternalServerError)
		return
	}
	defer reqBody.Close()

	// Stream response from git subprocess directly to HTTP response writer
	if err := h.gitSvc.ServiceRPC(service.ServiceTypeUploadPack, reqBody, w); err != nil {
		log.Printf("UploadPack: service RPC failed: %v", err)
		// Error already written to response by the subprocess or after it finished
		// Don't write another error header as data may have already been sent
	}
}

// ReceivePack handles POST /git-receive-pack (Push).
// This pipes the HTTP request body to `git receive-pack --stateless-rpc .`
// and streams the response back directly.
// Authentication is required for push operations.
func (h *GitHandler) ReceivePack(w http.ResponseWriter, r *http.Request) {
	// Extract and validate authentication
	auth := h.extractAuth(r)
	if auth == nil {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	valid, err := h.authSvc.Validate(auth.Username, auth.Password)
	if err != nil || !valid {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Validate Content-Type
	if r.Header.Get("Content-Type") != "application/x-git-receive-pack-request" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// Get request body, handling gzip if needed
	reqBody, err := h.getRequestBody(r)
	if err != nil {
		log.Printf("ReceivePack: failed to decompress request body: %v", err)
		http.Error(w, "Failed to decompress request", http.StatusInternalServerError)
		return
	}
	defer reqBody.Close()

	// Stream response from git subprocess directly to HTTP response writer
	if err := h.gitSvc.ServiceRPC(service.ServiceTypeReceivePack, reqBody, w); err != nil {
		log.Printf("ReceivePack: service RPC failed: %v", err)
	}

	// After push completes:
	// 1. Checkout working directory to reflect pushed content
	//    (git receive-pack only updates refs/objects, not working tree)
	// 2. Reload go-git repo to refresh cached state for read operations
	if checkoutErr := h.gitSvc.CheckoutWorkingTree(); checkoutErr != nil {
		log.Printf("ReceivePack: checkout failed: %v", checkoutErr)
	}
	if reloadErr := h.gitSvc.ReloadRepo(); reloadErr != nil {
		log.Printf("ReceivePack: failed to reload repo: %v", reloadErr)
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

// getRequestBody returns the request body, decompressing gzip if needed.
// Git clients may send gzip-compressed data for large pushes.
func (h *GitHandler) getRequestBody(r *http.Request) (io.ReadCloser, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		// Wrap gzip reader with a closer that also closes the underlying body
		return &gzipReadCloser{reader: gzReader, underlying: r.Body}, nil
	}
	return r.Body, nil
}

// gzipReadCloser wraps a gzip.Reader and closes both the reader and underlying body.
type gzipReadCloser struct {
	reader     *gzip.Reader
	underlying io.ReadCloser
}

func (g *gzipReadCloser) Read(p []byte) (int, error) {
	return g.reader.Read(p)
}

func (g *gzipReadCloser) Close() error {
	err := g.reader.Close()
	if closeErr := g.underlying.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}

// pktLine writes a pkt-line formatted data.
func pktLine(w io.Writer, data string) {
	size := len(data) + 4
	if size > 65524 {
		return
	}
	w.Write([]byte(fmt.Sprintf("%04x%s", size, data)))
}

// pktFlush writes a pkt-line flush packet.
func pktFlush(w io.Writer) {
	w.Write([]byte("0000"))
}
