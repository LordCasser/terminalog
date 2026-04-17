// Package handler provides HTTP request handlers for the application.
package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"terminalog/internal/service"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	// ready indicates if the server is ready.
	ready bool

	// mutex protects the ready flag.
	mutex sync.RWMutex

	// startTime is when the server started.
	startTime time.Time

	// gitSvc is the Git service for checking Git availability.
	gitSvc *service.GitService

	// cacheStats returns cache statistics.
	cacheStats func() service.CacheStats
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(gitSvc *service.GitService, cacheStats func() service.CacheStats) *HealthHandler {
	return &HealthHandler{
		ready:      false,
		startTime:  time.Now(),
		gitSvc:     gitSvc,
		cacheStats: cacheStats,
	}
}

// SetReady marks the server as ready.
func (h *HealthHandler) SetReady() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.ready = true
}

// SetNotReady marks the server as not ready.
func (h *HealthHandler) SetNotReady() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.ready = false
}

// IsReady checks if the server is ready.
func (h *HealthHandler) IsReady() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.ready
}

// Healthz handles GET /healthz - basic health check.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).Seconds(),
	}

	json.NewEncoder(w).Encode(response)
}

// Readyz handles GET /readyz - readiness check.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	h.mutex.RLock()
	ready := h.ready
	h.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if ready {
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		response := map[string]interface{}{
			"status":    "not_ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}
}

// Livez handles GET /livez - liveness check.
func (h *HealthHandler) Livez(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// Status handles GET /status - detailed status check.
func (h *HealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	gitAvailable := h.gitSvc != nil && h.gitSvc.CheckGitAvailable()

	cacheStats := service.CacheStats{}
	if h.cacheStats != nil {
		cacheStats = h.cacheStats()
	}

	response := map[string]interface{}{
		"status":       "ok",
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"uptime":       time.Since(h.startTime).Seconds(),
		"ready":        h.IsReady(),
		"gitAvailable": gitAvailable,
		"cacheStats":   cacheStats,
		"startTime":    h.startTime.UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}
