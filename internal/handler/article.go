// Package handler provides HTTP request handlers for the application.
package handler

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// ArticleHandler handles article-related HTTP requests.
type ArticleHandler struct {
	svc        *service.ArticleService
	versionSvc *service.VersionService
	fileSvc    *service.FileService
}

// NewArticleHandler creates a new ArticleHandler instance.
func NewArticleHandler(svc *service.ArticleService, versionSvc *service.VersionService, fileSvc *service.FileService) *ArticleHandler {
	return &ArticleHandler{svc: svc, versionSvc: versionSvc, fileSvc: fileSvc}
}

// parseSortParams extracts sort and order from query parameters.
// Default: sort=name (alphabetical), order=asc. Dirs first, then files.
func parseSortParams(r *http.Request) (model.SortField, model.SortOrder) {
	sortParam := r.URL.Query().Get("sort")
	orderParam := r.URL.Query().Get("order")

	// Default: alphabetical ascending for directory listing
	if sortParam == "" {
		return model.SortName, model.OrderAsc
	}

	return model.ParseSortField(sortParam), model.ParseSortOrder(orderParam)
}

// ListRoot handles GET /api/v1/articles - returns root directory listing.
func (h *ArticleHandler) ListRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sortField, sortOrder := parseSortParams(r)

	articles, err := h.svc.ListDirectory(ctx, "", sortField, sortOrder)
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.ArticleListResponse{
		Articles:   articles,
		CurrentDir: "",
		Total:      len(articles),
	})
}

// HandleRequest handles GET /api/v1/articles/* - smart routing based on path type.
// If the path is a directory, returns its listing.
// If the path is a file, returns its content.
// Supports /timeline and /version suffixes for files.
func (h *ArticleHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {

	// Get path from URL (wildcard pattern)
	path := chi.URLParam(r, "*")

	// Decode URL-encoded path
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid path encoding")
		return
	}
	path = decodedPath

	// Check for /timeline suffix
	if strings.HasSuffix(path, "/timeline") {
		articlePath := strings.TrimSuffix(path, "/timeline")
		h.handleTimeline(w, r, articlePath)
		return
	}

	// Check for /version suffix
	if strings.HasSuffix(path, "/version") {
		articlePath := strings.TrimSuffix(path, "/version")
		h.handleVersion(w, r, articlePath)
		return
	}

	// Determine if this is a directory or a file
	absPath, err := h.fileSvc.ValidatePath(path)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid path")
		return
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			utils.RespondNotFound(w, "Not found")
			return
		}
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	if info.IsDir() {
		// Path is a directory -> return listing
		h.handleDirectoryListing(w, r, path)
		return
	}

	// Path is a file -> return article content
	h.handleArticleContent(w, r, path)
}

// handleDirectoryListing returns the listing of a directory.
func (h *ArticleHandler) handleDirectoryListing(w http.ResponseWriter, r *http.Request, dirPath string) {
	ctx := r.Context()
	sortField, sortOrder := parseSortParams(r)

	articles, err := h.svc.ListDirectory(ctx, dirPath, sortField, sortOrder)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			utils.RespondNotFound(w, "Directory not found")
			return
		}
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.ArticleListResponse{
		Articles:   articles,
		CurrentDir: dirPath,
		Total:      len(articles),
	})
}

// handleArticleContent returns the content of a specific article.
func (h *ArticleHandler) handleArticleContent(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	article, err := h.svc.GetArticle(ctx, path)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrNotFound):
			utils.RespondNotFound(w, "Article not found")
		case errors.Is(err, model.ErrNotCommitted):
			utils.RespondBadRequest(w, "File not committed")
		case errors.Is(err, model.ErrInvalidPath):
			utils.RespondBadRequest(w, "Invalid path")
		default:
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.ArticleResponse{
		Path:         article.Path,
		Title:        article.Title,
		Content:      article.Content,
		CreatedAt:    article.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:    article.CreatedBy,
		EditedAt:     article.EditedAt.Format("2006-01-02 15:04:05"),
		EditedBy:     article.EditedBy,
		Contributors: article.Contributors,
	})
}

// handleTimeline returns the git timeline for an article.
func (h *ArticleHandler) handleTimeline(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	commits, err := h.svc.GetTimeline(ctx, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrNotCommitted) {
			utils.RespondNotFound(w, "Article not found")
		} else {
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	utils.RespondJSON(w, http.StatusOK, model.TimelineResponse{
		Commits: commits,
	})
}

// handleVersion returns the version info for an article.
func (h *ArticleHandler) handleVersion(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	versionInfo, err := h.versionSvc.GetVersion(ctx, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrNotCommitted) {
			utils.RespondNotFound(w, "Article not found")
		} else {
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	utils.RespondJSON(w, http.StatusOK, versionInfo)
}
