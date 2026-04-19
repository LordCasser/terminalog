package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"terminalog/internal/handler"
	"testing/fstest"
)

func TestGzipCompressesJSONResponses(t *testing.T) {
	testHandler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	req.Header.Set("Accept-Encoding", "gzip, br")
	recorder := httptest.NewRecorder()

	testHandler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	if got := resp.Header.Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip content encoding, got %q", got)
	}
	if got := resp.Header.Get("Vary"); !strings.Contains(got, "Accept-Encoding") {
		t.Fatalf("expected Vary to include Accept-Encoding, got %q", got)
	}

	reader, err := gzip.NewReader(bytes.NewReader(recorder.Body.Bytes()))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read gzipped body: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected response body: %q", string(body))
	}
}

func TestGzipSkipsAlreadyEncodedStaticAssets(t *testing.T) {
	staticHandler := &handler.StaticHandler{}
	staticHandler.SetFS(fstest.MapFS{
		"index.html":    {Data: []byte("<html>plain</html>")},
		"index.html.br": {Data: []byte("brotli-content")},
	})

	testHandler := Gzip(staticHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, br")
	recorder := httptest.NewRecorder()

	testHandler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	if got := resp.Header.Get("Content-Encoding"); got != "br" {
		t.Fatalf("expected existing brotli encoding to win, got %q", got)
	}
	if body := recorder.Body.String(); body != "brotli-content" {
		t.Fatalf("unexpected body after static response: %q", body)
	}
}

func TestGzipSkipsWebSocketUpgradeRequests(t *testing.T) {
	testHandler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("upgrade"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/ws/terminal", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	recorder := httptest.NewRecorder()

	testHandler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Fatalf("expected websocket upgrade request to skip gzip, got %q", got)
	}
	if body := recorder.Body.String(); body != "upgrade" {
		t.Fatalf("unexpected plain response body: %q", body)
	}
}
