package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/domain"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

// APIKeyHandler handles API key CRUD operations.
type APIKeyHandler struct {
	repo *postgres.APIKeyRepo
}

// NewAPIKeyHandler creates a new APIKeyHandler.
func NewAPIKeyHandler(repo *postgres.APIKeyRepo) *APIKeyHandler {
	return &APIKeyHandler{repo: repo}
}

// Routes returns a chi.Router with API key routes.
func (h *APIKeyHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Delete("/{id}", h.Revoke)
	r.Put("/{id}", h.Update)

	return r
}

// createRequest is the body for POST /.
type createRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	ExpiresIn   *int     `json:"expires_in"` // seconds; nil = no expiry
}

// updateRequest is the body for PUT /{id}.
type updateRequest struct {
	Name        *string  `json:"name"`
	Permissions *[]string `json:"permissions"`
	ExpiresIn   *int     `json:"expires_in"` // seconds; nil = leave unchanged
}

// Create handles POST / — creates a new API key and returns it with the one-time plaintext.
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := mw.GetAuthUser(r.Context())
	if user == nil {
		api.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var body createRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if body.Name == "" {
		api.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "name is required")
		return
	}

	if body.Permissions == nil {
		body.Permissions = []string{}
	}

	var expiresAt *time.Time
	if body.ExpiresIn != nil && *body.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*body.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	result, err := h.repo.Create(r.Context(), body.Name, user.Subject, body.Permissions, expiresAt)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create API key")
		return
	}

	api.JSON(w, http.StatusCreated, result, nil)
}

// List handles GET / — returns all active keys for the authenticated user.
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	user := mw.GetAuthUser(r.Context())
	if user == nil {
		api.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	keys, err := h.repo.ListByOwner(r.Context(), user.Subject)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list API keys")
		return
	}

	if keys == nil {
		keys = []domain.APIKey{}
	}

	api.JSON(w, http.StatusOK, keys, nil)
}

// Revoke handles DELETE /{id} — revokes the key (204).
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	user := mw.GetAuthUser(r.Context())
	if user == nil {
		api.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid API key ID")
		return
	}

	if err := h.repo.Revoke(r.Context(), id, user.Subject); err != nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "API key not found or already revoked")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Update handles PUT /{id} — updates the key's name, permissions, or expiry (204).
func (h *APIKeyHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := mw.GetAuthUser(r.Context())
	if user == nil {
		api.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid API key ID")
		return
	}

	var body updateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	var expiresAt *time.Time
	if body.ExpiresIn != nil && *body.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(*body.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	if err := h.repo.Update(r.Context(), id, user.Subject, body.Name, body.Permissions, expiresAt); err != nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "API key not found or revoked")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
