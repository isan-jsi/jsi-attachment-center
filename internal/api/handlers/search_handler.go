package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

type SearchHandler struct {
	repo *postgres.DocumentRepo
}

func NewSearchHandler(repo *postgres.DocumentRepo) *SearchHandler {
	return &SearchHandler{repo: repo}
}

func (h *SearchHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(mw.RequirePermission("documents:read")).Get("/", h.Search)
	return r
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := postgres.SearchParams{
		Query:       q.Get("q"),
		OwnerClass:  q.Get("owner_class"),
		ContentType: q.Get("content_type"),
		Page:        queryInt(r, "page", 1),
		PerPage:     queryInt(r, "per_page", 20),
	}

	if df := q.Get("date_from"); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			params.DateFrom = &t
		}
	}
	if dt := q.Get("date_to"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			params.DateTo = &t
		}
	}

	docs, total, err := h.repo.Search(r.Context(), params)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "search failed")
		return
	}

	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage > 0 {
		totalPages++
	}

	api.JSON(w, http.StatusOK, docs, &api.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
		RequestID:  mw.GetRequestID(r.Context()),
	})
}
