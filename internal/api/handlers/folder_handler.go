package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/domain"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

// FolderDetail contains a folder with its children and documents.
type FolderDetail struct {
	Folder    *domain.Folder    `json:"folder"`
	Children  []domain.Folder   `json:"children"`
	Documents []domain.Document `json:"documents"`
}

// FolderHandler handles folder CRUD operations.
type FolderHandler struct {
	repo *postgres.FolderRepo
}

// NewFolderHandler creates a new FolderHandler.
func NewFolderHandler(repo *postgres.FolderRepo) *FolderHandler {
	return &FolderHandler{repo: repo}
}

// Routes returns a chi.Router with folder routes.
func (h *FolderHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(mw.RequirePermission("folders:read")).Get("/", h.ListRoots)
	r.With(mw.RequirePermission("folders:read")).Get("/{id}", h.GetByID)
	r.With(mw.RequirePermission("folders:write")).Post("/", h.Create)
	r.With(mw.RequirePermission("folders:write")).Put("/{id}", h.Update)
	r.With(mw.RequirePermission("folders:write")).Delete("/{id}", h.Delete)

	return r
}

// ListRoots returns all root folders (no parent).
func (h *FolderHandler) ListRoots(w http.ResponseWriter, r *http.Request) {
	folders, err := h.repo.ListRoots(r.Context())
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list root folders")
		return
	}

	api.JSON(w, http.StatusOK, folders, &api.Meta{
		RequestID: mw.GetRequestID(r.Context()),
	})
}

// GetByID returns a folder with its children and documents.
func (h *FolderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid folder ID")
		return
	}

	folder, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get folder")
		return
	}
	if folder == nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "folder not found")
		return
	}

	children, err := h.repo.GetChildren(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get children")
		return
	}

	docs, err := h.repo.GetDocuments(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get documents")
		return
	}

	detail := FolderDetail{
		Folder:    folder,
		Children:  children,
		Documents: docs,
	}

	api.JSON(w, http.StatusOK, detail, nil)
}

// Create creates a new folder.
func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name              string     `json:"name"`
		ParentID          *uuid.UUID `json:"parent_id"`
		Path              string     `json:"path"`
		OwnerClassLibrary string     `json:"owner_class_library"`
		OwnerClassName    string     `json:"owner_class_name"`
		OwnerID           string     `json:"owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if body.Name == "" {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "name is required")
		return
	}

	folder := &domain.Folder{
		Name:              body.Name,
		ParentID:          body.ParentID,
		Path:              body.Path,
		OwnerClassLibrary: body.OwnerClassLibrary,
		OwnerClassName:    body.OwnerClassName,
		OwnerID:           body.OwnerID,
	}

	if err := h.repo.Create(r.Context(), folder); err != nil {
		api.JSONError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create folder")
		return
	}

	api.JSON(w, http.StatusCreated, folder, nil)
}

// Update updates folder metadata.
func (h *FolderHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid folder ID")
		return
	}

	folder, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get folder")
		return
	}
	if folder == nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "folder not found")
		return
	}

	var body struct {
		Name     *string    `json:"name"`
		ParentID *uuid.UUID `json:"parent_id"`
		Path     *string    `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if body.Name != nil {
		folder.Name = *body.Name
	}
	if body.ParentID != nil {
		folder.ParentID = body.ParentID
	}
	if body.Path != nil {
		folder.Path = *body.Path
	}

	if err := h.repo.Update(r.Context(), folder); err != nil {
		api.JSONError(w, http.StatusInternalServerError, "DB_ERROR", "failed to update folder")
		return
	}

	api.JSON(w, http.StatusOK, folder, nil)
}

// Delete deletes a folder, optionally cascading to children and linked documents.
func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid folder ID")
		return
	}

	cascade := r.URL.Query().Get("cascade") == "true"

	if err := h.repo.Delete(r.Context(), id, cascade); err != nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "folder not found")
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{"status": "deleted"}, nil)
}
