package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

// SearchHandler handles document search operations.
// It supports both ILIKE-based search (via DocumentRepo) and full-text search
// with highlights and ranking (via SearchRepo).
type SearchHandler struct {
	repo       *postgres.DocumentRepo
	searchRepo *postgres.SearchRepo
}

// NewSearchHandler creates a SearchHandler that uses both the document repo
// (for ILIKE/date-range search) and the search repo (for full-text search).
func NewSearchHandler(repo *postgres.DocumentRepo, searchRepo *postgres.SearchRepo) *SearchHandler {
	return &SearchHandler{repo: repo, searchRepo: searchRepo}
}

func (h *SearchHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(mw.RequirePermission("documents:read")).Get("/", h.Search)
	r.With(mw.RequirePermission("documents:read")).Get("/suggest", h.Suggest)
	return r
}

// Search handles GET /api/v1/search
//
// When the "fts=true" query parameter is set (and a non-empty "q" is provided),
// it delegates to the full-text search engine (tsvector + websearch_to_tsquery).
// Otherwise it falls back to the existing ILIKE-based search.
//
// Query parameters:
//
//	q            - search query string
//	fts          - set to "true" to use full-text search
//	owner_id     - filter by owner_id (fts mode only)
//	owner_class  - filter by owner_class_name (ILIKE mode only)
//	content_type - filter by content_type
//	date_from    - RFC3339 lower bound on created_at (ILIKE mode only)
//	date_to      - RFC3339 upper bound on created_at (ILIKE mode only)
//	page         - page number (ILIKE mode, default 1)
//	per_page     - page size  (ILIKE mode, default 20)
//	limit        - result limit (fts mode, default 20)
//	offset       - result offset (fts mode, default 0)
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Full-text search path.
	if q.Get("fts") == "true" && q.Get("q") != "" {
		params := postgres.FullTextSearchParams{
			Query:       q.Get("q"),
			OwnerID:     q.Get("owner_id"),
			ContentType: q.Get("content_type"),
			Limit:       queryInt(r, "limit", 20),
			Offset:      queryInt(r, "offset", 0),
		}

		result, err := h.searchRepo.FullTextSearch(r.Context(), params)
		if err != nil {
			api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "full-text search failed")
			return
		}

		api.JSON(w, http.StatusOK, result, &api.Meta{
			Page:       params.Offset/params.Limit + 1,
			PerPage:    params.Limit,
			Total:      int64(result.Total),
			TotalPages: totalPages(result.Total, params.Limit),
			RequestID:  mw.GetRequestID(r.Context()),
		})
		return
	}

	// ILIKE / date-range fallback path.
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

	api.JSON(w, http.StatusOK, docs, &api.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages(int(total), params.PerPage),
		RequestID:  mw.GetRequestID(r.Context()),
	})
}

// Suggest handles GET /api/v1/search/suggest
//
// Returns up to `limit` (default 10) filename completions matching the
// given `q` prefix. Requires documents:read permission.
//
// Query parameters:
//
//	q     - filename prefix to autocomplete
//	limit - maximum number of suggestions to return (default 10)
func (h *SearchHandler) Suggest(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("q")
	limit := queryInt(r, "limit", 10)

	suggestions, err := h.searchRepo.Suggest(r.Context(), prefix, limit)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "suggest failed")
		return
	}

	if suggestions == nil {
		suggestions = []string{}
	}

	api.JSON(w, http.StatusOK, suggestions, &api.Meta{
		RequestID: mw.GetRequestID(r.Context()),
	})
}

// totalPages computes the number of pages given a total count and page size.
func totalPages(total, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := total / perPage
	if total%perPage > 0 {
		pages++
	}
	return pages
}
