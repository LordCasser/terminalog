package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terminalog/internal/handler"
	"terminalog/internal/service"
	"terminalog/pkg/testutil"
)

func TestAssetHandler_RevalidatesMutableContent(t *testing.T) {
	repo, err := testutil.NewTestRepo()
	require.NoError(t, err)
	defer repo.Cleanup()
	require.NoError(t, repo.CreateImageFileAndCommit(".assets/image.png", []byte("first"), "add image", "author"))

	gitSvc, err := service.NewGitService(repo.Path)
	require.NoError(t, err)
	assetHandler := handler.NewAssetHandler(service.NewAssetService(gitSvc))
	router := chi.NewRouter()
	router.Get("/api/v1/assets/*", assetHandler.Get)

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/v1/assets/image.png", nil))
	require.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, "public, no-cache", first.Header().Get("Cache-Control"))
	etag := first.Header().Get("ETag")
	require.NotEmpty(t, etag)

	revalidateRequest := httptest.NewRequest(http.MethodGet, "/api/v1/assets/image.png", nil)
	revalidateRequest.Header.Set("If-None-Match", etag)
	revalidated := httptest.NewRecorder()
	router.ServeHTTP(revalidated, revalidateRequest)

	assert.Equal(t, http.StatusNotModified, revalidated.Code)
	assert.Empty(t, revalidated.Body.String())
}
