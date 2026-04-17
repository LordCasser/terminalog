// Package handler provides HTTP request handlers for the application.
package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"terminalog/internal/model"
	"terminalog/internal/service"
	"terminalog/pkg/utils"
)

// ArticleHandler handles article-related HTTP requests.
type ArticleHandler struct {
	svc *service.ArticleService
}

// NewArticleHandler creates a new ArticleHandler instance.
func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
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

// Get handles GET /api/articles/{path}.
func (h *ArticleHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get path from URL
	path := chi.URLParam(r, "path")

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
