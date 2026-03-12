package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

type SyncHandler struct {
	repo *postgres.SyncRepo
}

func NewSyncHandler(repo *postgres.SyncRepo) *SyncHandler {
	return &SyncHandler{repo: repo}
}

func (h *SyncHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(mw.RequirePermission("sync:read")).Get("/status", h.Status)
	r.With(mw.RequirePermission("sync:read")).Get("/dlq", h.ListDLQ)
	r.With(mw.RequirePermission("sync:admin")).Post("/dlq/{id}/retry", h.RetryDLQ)

	return r
}

func (h *SyncHandler) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.repo.GetSyncStatus(r.Context())
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "failed to get sync status")
		return
	}
	api.JSON(w, http.StatusOK, status, nil)
}

func (h *SyncHandler) ListDLQ(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	perPage := queryInt(r, "per_page", 20)

	entries, total, err := h.repo.ListDLQ(r.Context(), page, perPage)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL", "failed to list DLQ entries")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	api.JSON(w, http.StatusOK, entries, &api.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		RequestID:  mw.GetRequestID(r.Context()),
	})
}

func (h *SyncHandler) RetryDLQ(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid DLQ entry ID format")
		return
	}

	if err := h.repo.RetryDLQEntry(r.Context(), id); err != nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "DLQ entry not found or already resolved")
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{"status": "retry_scheduled"}, nil)
}
