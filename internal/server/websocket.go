// Package server provides HTTP server functionality.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"terminalog/internal/service"
)

// WebSocketHandler handles WebSocket connections for terminal commands.
type WebSocketHandler struct {
	// completionSvc is the completion service.
	completionSvc *service.CompletionService

	// logger is the structured logger.
	logger *slog.Logger

	// upgrader is the WebSocket upgrader.
	upgrader websocket.Upgrader

	// connections stores active WebSocket connections.
	connections sync.Map // map[string]*websocket.Conn
}

// NewWebSocketHandler creates a new WebSocketHandler instance.
// serverHost is the configured server host (from config), used for origin validation
// in production mode. May be "localhost", "0.0.0.0", "", or a domain name.
func NewWebSocketHandler(completionSvc *service.CompletionService, logger *slog.Logger, debug bool, serverHost string) *WebSocketHandler {
	// Configure upgrader with origin validation
	// Policy:
	//   - Debug mode: allow all origins (convenience for development)
	//   - Production mode:
	//     1. Same-origin requests (no Origin header) are always allowed
	//     2. Localhost origins are allowed (local access convenience)
	//     3. Origin host must match the configured serverHost
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if debug {
				return true
			}
			origin := r.Header.Get("Origin")
			// Same-origin requests (no Origin header) are always allowed
			if origin == "" {
				return true
			}
			// Parse the origin URL to extract host
			originURL, err := url.Parse(origin)
			if err != nil {
				logger.Warn("WebSocket: invalid Origin header", "origin", origin)
				return false
			}
			originHost := originURL.Host
			// Allow localhost origins (common for development and local access)
			if IsLocalhost(originHost) {
				return true
			}
			// Allow if origin matches configured server host
			if originHost == serverHost {
				return true
			}
			logger.Warn("WebSocket: rejected Origin", "origin", origin, "serverHost", serverHost)
			return false
		},
	}

	return &WebSocketHandler{
		completionSvc: completionSvc,
		logger:        logger,
		upgrader:      upgrader,
	}
}

// HandleTerminal handles WebSocket connections for the terminal endpoint.
// It processes path completion requests from the frontend.
func (h *WebSocketHandler) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return
	}

	// Generate connection ID
	connID := generateConnectionID()
	h.connections.Store(connID, conn)
	h.logger.Info("WebSocket connection established", "connID", connID, "remote", r.RemoteAddr)

	// Ensure connection is closed when handler returns
	defer func() {
		h.connections.Delete(connID)
		conn.Close()
		h.logger.Info("WebSocket connection closed", "connID", connID)
	}()

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Handle ping/pong for keepalive
	conn.SetPingHandler(func(appData string) error {
		h.logger.Debug("Received ping", "connID", connID)
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
	})

	// Message handling loop
	for {
		// Read message
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error", "connID", connID, "error", err)
			} else {
				h.logger.Debug("WebSocket connection closed by client", "connID", connID)
			}
			break
		}

		// Reset read deadline after successful read
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Only handle text messages
		if messageType != websocket.TextMessage {
			h.logger.Debug("Ignoring non-text message", "connID", connID, "type", messageType)
			continue
		}

		// Parse message type
		var baseMsg struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(message, &baseMsg); err != nil {
			h.logger.Error("Failed to parse message type", "connID", connID, "error", err)
			h.sendError(conn, "Invalid message format")
			continue
		}

		// Handle message based on type
		switch baseMsg.Type {
		case "completion_request":
			h.handleCompletionRequest(conn, message)
		default:
			h.logger.Warn("Unknown message type", "connID", connID, "type", baseMsg.Type)
			h.sendError(conn, "Unknown message type: "+baseMsg.Type)
		}
	}
}

// handleCompletionRequest handles a path completion request.
func (h *WebSocketHandler) handleCompletionRequest(conn *websocket.Conn, message []byte) {
	// Parse request
	var req service.CompletionRequest
	if err := json.Unmarshal(message, &req); err != nil {
		h.logger.Error("Failed to parse completion request", "error", err)
		h.sendError(conn, "Invalid completion request format")
		return
	}

	h.logger.Debug("Received completion request", "dir", req.Dir, "prefix", req.Prefix)

	// Process request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := h.completionSvc.HandleCompletion(ctx, req)
	if err != nil {
		h.logger.Error("Failed to handle completion", "error", err)
		h.sendError(conn, "Completion error: "+err.Error())
		return
	}

	// Send response
	h.sendMessage(conn, response)
}

// sendMessage sends a JSON message to the WebSocket connection.
func (h *WebSocketHandler) sendMessage(conn *websocket.Conn, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal message", "error", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		h.logger.Error("Failed to send message", "error", err)
	}
}

// sendError sends an error message to the WebSocket connection.
func (h *WebSocketHandler) sendError(conn *websocket.Conn, errMsg string) {
	errorMsg := struct {
		Type  string `json:"type"`
		Error string `json:"error"`
	}{
		Type:  "error",
		Error: errMsg,
	}

	h.sendMessage(conn, errorMsg)
}

// IsLocalhost checks if a host string refers to localhost (including port variants).
// DNS names are case-insensitive per RFC 4343, so "LOCALHOST", "Localhost", etc. all match.
func IsLocalhost(host string) bool {
	hostOnly := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostOnly = h
	}
	// Normalize to lowercase for case-insensitive hostname comparison
	switch strings.ToLower(hostOnly) {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return false
}

// generateConnectionID generates a unique connection ID.
func generateConnectionID() string {
	return time.Now().Format("20060102-150405-") + randomString(6)
}

// randomString generates a random string of given length.
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().Nanosecond()%len(charset)]
		time.Sleep(1 * time.Nanosecond) // Ensure different nanoseconds
	}
	return string(b)
}
