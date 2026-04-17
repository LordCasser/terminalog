// Package handler provides HTTP request handlers for the application.
package handler

import (
	"errors"
	"net/http"
	"net/url"
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
}

// NewArticleHandler creates a new ArticleHandler instance.
func NewArticleHandler(svc *service.ArticleService, versionSvc *service.VersionService) *ArticleHandler {
	return &ArticleHandler{svc: svc, versionSvc: versionSvc}
}

// List handles GET /api/articles.
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	dir := r.URL.Query().Get("dir")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	// Default values
	if sort == "" {
		sort = "edited"
	}
	if order == "" {
		order = "desc"
	}

	// Build options
	opts := service.ListOptions{
		Dir:   dir,
		Sort:  model.ParseSortField(sort),
		Order: model.ParseSortOrder(order),
	}

	// Get articles
	articles, err := h.svc.ListArticles(ctx, opts)
	if err != nil {
		utils.RespondInternalServerError(w, err.Error())
		return
	}

	// Respond
	utils.RespondJSON(w, http.StatusOK, model.ArticleListResponse{
		Articles:   articles,
		CurrentDir: dir,
		Total:      len(articles),
	})
}

// Get handles GET /api/articles/*.
func (h *ArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path from URL (using wildcard pattern)
	path := chi.URLParam(r, "*")

	// Get article
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

	// Respond
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

// HandleArticleRequest handles GET /api/articles/*, routing to Get, Timeline, or Version.
func (h *ArticleHandler) HandleArticleRequest(w http.ResponseWriter, r *http.Request) {
	// Get path from URL (using wildcard pattern)
	path := chi.URLParam(r, "*")

	// Decode URL-encoded path (e.g., tech%2Ffrontend%2Fvue-components.md -> tech/frontend/vue-components.md)
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		utils.RespondBadRequest(w, "Invalid path encoding")
		return
	}
	path = decodedPath

	// Check if this is a timeline request
	if strings.HasSuffix(path, "/timeline") {
		// Remove /timeline suffix to get the article path
		articlePath := strings.TrimSuffix(path, "/timeline")
		h.handleTimeline(w, r, articlePath)
		return
	}

	// Check if this is a version request
	if strings.HasSuffix(path, "/version") {
		// Remove /version suffix to get the article path
		articlePath := strings.TrimSuffix(path, "/version")
		h.handleVersion(w, r, articlePath)
		return
	}

	// Regular article request
	h.handleGet(w, r, path)
}

// handleGet handles the article get request.
func (h *ArticleHandler) handleGet(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	// Get article
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

	// Respond
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

// handleTimeline handles the timeline request.
func (h *ArticleHandler) handleTimeline(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	// Get timeline
	commits, err := h.svc.GetTimeline(ctx, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrNotCommitted) {
			utils.RespondNotFound(w, "Article not found")
		} else {
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	// Respond
	utils.RespondJSON(w, http.StatusOK, model.TimelineResponse{
		Commits: commits,
	})
}

// handleVersion handles the version request.
func (h *ArticleHandler) handleVersion(w http.ResponseWriter, r *http.Request, path string) {
	ctx := r.Context()

	// Get version info
	versionInfo, err := h.versionSvc.GetVersion(ctx, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrNotCommitted) {
			utils.RespondNotFound(w, "Article not found")
		} else {
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	// Respond
	utils.RespondJSON(w, http.StatusOK, versionInfo)
}

// Timeline handles GET /api/articles/{path}/timeline.
func (h *ArticleHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path from URL
	path := chi.URLParam(r, "path")

	// Get timeline
	commits, err := h.svc.GetTimeline(ctx, path)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrNotCommitted) {
			utils.RespondNotFound(w, "Article not found")
		} else {
			utils.RespondInternalServerError(w, err.Error())
		}
		return
	}

	// Respond
	utils.RespondJSON(w, http.StatusOK, model.TimelineResponse{
		Commits: commits,
	})
}
