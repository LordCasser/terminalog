package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/handler"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestHealthHandler_ReportsPublishedHead(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()
	require.NoError(t, repo.CreateMarkdownFile("article.md", "# Article", "Add", "author"))
	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)

	health := handler.NewHealthHandler(gitSvc, nil)
	health.SetReady()

	ready := httptest.NewRecorder()
	health.Readyz(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	assert.Equal(t, http.StatusOK, ready.Code)
	assert.Equal(t, "no-store", ready.Header().Get("Cache-Control"))

	status := httptest.NewRecorder()
	health.Status(status, httptest.NewRequest(http.MethodGet, "/status", nil))
	require.Equal(t, http.StatusOK, status.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(status.Body.Bytes(), &payload))
	assert.Equal(t, true, payload["gitAvailable"])
	assert.Len(t, payload["gitHead"], 40)
}
