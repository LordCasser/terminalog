package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestServeFileUsesBrotliWhenAvailable(t *testing.T) {
	handler := &StaticHandler{}
	handler.SetFS(fstest.MapFS{
		"index.html":    {Data: []byte("<html>plain</html>")},
		"index.html.br": {Data: []byte("brotli-content")},
		"index.html.gz": {Data: []byte("gzip-content")},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, br")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Encoding"); got != "br" {
		t.Fatalf("expected brotli encoding, got %q", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("expected html content type, got %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=0, must-revalidate" {
		t.Fatalf("expected html cache control, got %q", got)
	}
}

func TestServeFileFallsBackToOriginalWhenCompressionUnavailable(t *testing.T) {
	handler := &StaticHandler{}
	handler.SetFS(fstest.MapFS{
		"_next/static/chunks/app.js": {Data: []byte("console.log('plain');")},
	})

	req := httptest.NewRequest(http.MethodGet, "/_next/static/chunks/app.js", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Encoding"); got != "" {
		t.Fatalf("expected no content encoding, got %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("expected immutable asset cache control, got %q", got)
	}
}
