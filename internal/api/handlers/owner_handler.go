package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

type OwnerHandler struct {
	repo *postgres.DocumentRepo
}

func NewOwnerHandler(repo *postgres.DocumentRepo) *OwnerHandler {
	return &OwnerHandler{repo: repo}
}

func (h *OwnerHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(mw.RequirePermission("documents:read")).Get("/", h.ListOwnerClasses)
	r.With(mw.RequirePermission("documents:read")).Get("/{class}/documents", h.ListByOwnerClass)
	return r
}

func (h *OwnerHandler) ListOwnerClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := h.repo.ListOwnerClasses(r.Context())
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "failed to list owner classes")
		return
	}
	api.JSON(w, http.StatusOK, classes, nil)
}

func (h *OwnerHandler) ListByOwnerClass(w http.ResponseWriter, r *http.Request) {
	className := chi.URLParam(r, "class")
	params := postgres.ListParams{
		OwnerClassName: className,
		Page:           queryInt(r, "page", 1),
		PerPage:        queryInt(r, "per_page", 20),
	}

	docs, total, err := h.repo.List(r.Context(), params)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "failed to list documents")
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
