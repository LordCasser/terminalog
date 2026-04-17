// Package utils provides utility functions for the application.
package utils

import (
	"encoding/json"
	"net/http"

	"terminalog/internal/model"
)

// RespondJSON writes a JSON response with the given status code and data.
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, we can't do much more than log
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// RespondError writes a JSON error response with the given status code and message.
func RespondError(w http.ResponseWriter, statusCode int, message string) {
	RespondJSON(w, statusCode, model.ErrorResponse{
		Error: message,
	})
}

// RespondNotFound writes a 404 Not Found response.
func RespondNotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, message)
}

// RespondBadRequest writes a 400 Bad Request response.
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, message)
}

// RespondUnauthorized writes a 401 Unauthorized response.
func RespondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Git\"")
	RespondError(w, http.StatusUnauthorized, message)
}

// RespondInternalServerError writes a 500 Internal Server Error response.
func RespondInternalServerError(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusInternalServerError, message)
}
