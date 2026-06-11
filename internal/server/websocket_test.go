// Package server_test provides external tests for the server package.
package server_test

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"

	"terminalog/internal/server"
	"terminalog/internal/service"
)

// startTestServer creates a real HTTP server on a random port for WebSocket upgrade testing.
// httptest.NewServer does NOT support http.Hijacker, which gorilla/websocket requires
// for successful upgrades. Using a real net/http.Server ensures hijack support.
func startTestServer(t *testing.T, h http.HandlerFunc) (*http.Server, string) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	srv := &http.Server{Handler: h}
	go srv.Serve(l)

	addr := "http://" + l.Addr().String()
	return srv, addr
}

// makeUpgradeRequest sends a WebSocket upgrade request with the given Origin header
// to the test server and returns the HTTP status code.
func makeUpgradeRequest(t *testing.T, url, origin string) int {
	t.Helper()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	// Use a client that does NOT follow redirects and does NOT handle 101 specially
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

// TestIsLocalhost tests the IsLocalhost helper function directly.
func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name string
		host string
		want bool
	}{
		{"plain localhost", "localhost", true},
		{"localhost with port", "localhost:8080", true},
		{"localhost with standard port", "localhost:80", true},
		{"ipv4 loopback", "127.0.0.1", true},
		{"ipv4 loopback with port", "127.0.0.1:3000", true},
		{"ipv6 loopback", "::1", true},
		{"ipv6 loopback with port", "[::1]:443", true},
		{"external hostname", "myblog.com", false},
		{"external hostname with port", "myblog.com:443", false},
		{"external IP", "192.168.1.1", false},
		{"external IP with port", "192.168.1.1:8080", false},
		{"empty", "", false},
		{"malformed but not localhost", "not-a-host", false},
		{"localhost upper case", "LOCALHOST", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := server.IsLocalhost(tt.host)
			if got != tt.want {
				t.Errorf("IsLocalhost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

// TestWebSocketOriginValidation_Production tests the CheckOrigin policy in production (non-debug) mode.
func TestWebSocketOriginValidation_Production(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name       string
		serverHost string
		origin     string // empty means no Origin header
		wantCode   int
	}{
		{
			name:       "no Origin header -> allowed (same-origin)",
			serverHost: "myblog.com",
			origin:     "",
			wantCode:   http.StatusSwitchingProtocols, // 101 - upgrade proceeds
		},
		{
			name:       "localhost origin -> allowed",
			serverHost: "myblog.com",
			origin:     "http://localhost:8080",
			wantCode:   http.StatusSwitchingProtocols, // 101
		},
		{
			name:       "127.0.0.1 origin -> allowed",
			serverHost: "myblog.com",
			origin:     "http://127.0.0.1:3000",
			wantCode:   http.StatusSwitchingProtocols, // 101
		},
		{
			name:       "matching host -> allowed",
			serverHost: "myblog.com",
			origin:     "https://myblog.com",
			wantCode:   http.StatusSwitchingProtocols, // 101
		},
		{
			name:       "matching host with port -> allowed",
			serverHost: "myblog.com:443",
			origin:     "https://myblog.com:443",
			wantCode:   http.StatusSwitchingProtocols, // 101
		},
		{
			name:       "mismatched origin -> rejected",
			serverHost: "myblog.com",
			origin:     "https://evil.com",
			wantCode:   http.StatusForbidden, // 403
		},
		{
			name:       "invalid Origin URL -> rejected",
			serverHost: "myblog.com",
			origin:     "not-a-valid-url",
			wantCode:   http.StatusForbidden, // 403
		},
		{
			name:       "subdomain mismatch -> rejected",
			serverHost: "myblog.com",
			origin:     "https://evil.myblog.com",
			wantCode:   http.StatusForbidden, // 403
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with debug=false (production mode)
			handler := server.NewWebSocketHandler(nil, logger, false, tt.serverHost)

			srv, baseURL := startTestServer(t, handler.HandleTerminal)
			defer srv.Close()

			code := makeUpgradeRequest(t, baseURL, tt.origin)
			if code != tt.wantCode {
				t.Errorf("got status %d, want %d (origin=%q, serverHost=%q)",
					code, tt.wantCode, tt.origin, tt.serverHost)
			}
		})
	}
}

// TestWebSocketOriginValidation_Debug tests that debug mode allows all origins.
func TestWebSocketOriginValidation_Debug(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name   string
		origin string
	}{
		{"no Origin", ""},
		{"localhost", "http://localhost:8080"},
		{"matching", "https://myblog.com"},
		{"evil origin", "https://evil.com"},
		{"invalid URL", "not-a-valid-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := server.NewWebSocketHandler(nil, logger, true, "myblog.com")

			srv, baseURL := startTestServer(t, handler.HandleTerminal)
			defer srv.Close()

			code := makeUpgradeRequest(t, baseURL, tt.origin)
			// In debug mode, all origins are allowed -> upgrade proceeds -> 101
			if code != http.StatusSwitchingProtocols {
				t.Errorf("debug mode: got status %d, want %d (origin=%q)",
					code, http.StatusSwitchingProtocols, tt.origin)
			}
		})
	}
}

// TestWebSocketHandler_NilCompletionService ensures handler works with nil completion service.
// The handler only uses completionSvc when processing messages; the upgrade itself is independent.
func TestWebSocketHandler_NilCompletionService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := server.NewWebSocketHandler(nil, logger, true, "localhost")

	if handler == nil {
		t.Fatal("NewWebSocketHandler returned nil")
	}
	// nil completionSvc is fine — actual message processing would panic, but upgrade is safe
	_ = service.CompletionService{} // ensure import is used
}
