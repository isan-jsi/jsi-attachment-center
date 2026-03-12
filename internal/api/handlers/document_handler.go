package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jsi/ibs-doc-engine/internal/api"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/domain"
	minioclient "github.com/jsi/ibs-doc-engine/internal/minio"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

// DocumentHandler handles document CRUD operations.
type DocumentHandler struct {
	repo  *postgres.DocumentRepo
	minio *minioclient.Client
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(repo *postgres.DocumentRepo, mc *minioclient.Client) *DocumentHandler {
	return &DocumentHandler{repo: repo, minio: mc}
}

// Routes returns a chi.Router with document routes.
func (h *DocumentHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.With(mw.RequirePermission("documents:read")).Get("/", h.List)
	r.With(mw.RequirePermission("documents:read")).Get("/{id}", h.GetByID)
	r.With(mw.RequirePermission("documents:read")).Get("/{id}/download", h.Download)
	r.With(mw.RequirePermission("documents:write")).Post("/", h.Create)
	r.With(mw.RequirePermission("documents:write")).Put("/{id}", h.Update)
	r.With(mw.RequirePermission("documents:write")).Delete("/{id}", h.Delete)

	return r
}

// List returns a paginated list of documents.
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	params := postgres.ListParams{
		OwnerClassLibrary: r.URL.Query().Get("owner_class_library"),
		OwnerClassName:    r.URL.Query().Get("owner_class_name"),
		OwnerID:           r.URL.Query().Get("owner_id"),
		Page:              queryInt(r, "page", 1),
		PerPage:           queryInt(r, "per_page", 20),
	}

	docs, total, err := h.repo.List(r.Context(), params)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list documents")
		return
	}

	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage != 0 {
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

// GetByID returns a single document by UUID.
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid document ID")
		return
	}

	doc, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get document")
		return
	}
	if doc == nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	api.JSON(w, http.StatusOK, doc, nil)
}

// Download generates a presigned GET URL and redirects to it.
func (h *DocumentHandler) Download(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid document ID")
		return
	}

	doc, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get document")
		return
	}
	if doc == nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	presigned, err := h.minio.Inner().PresignedGetObject(r.Context(), doc.MinioBucket, doc.MinioKey, 15*time.Minute, nil)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "PRESIGN_ERROR", "failed to generate download URL")
		return
	}

	http.Redirect(w, r, presigned.String(), http.StatusTemporaryRedirect)
}

// Create handles multipart form upload.
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 100 << 20 // 100 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "missing file field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read uploaded file")
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		filename = header.Filename
	}

	ownerID := r.FormValue("owner_id")
	ownerClassLibrary := r.FormValue("owner_class_library")
	ownerClassName := r.FormValue("owner_class_name")
	attachmentType := r.FormValue("attachment_type")
	attachmentTypeID := 0
	if v := r.FormValue("attachment_type_id"); v != "" {
		attachmentTypeID, _ = strconv.Atoi(v)
	}

	contentType := minioclient.DetectContentType(data, filename)
	docID := uuid.New()
	key := fmt.Sprintf("%s/%s/%s/%s", ownerClassLibrary, ownerClassName, ownerID, docID.String())

	result, err := h.minio.Upload(r.Context(), key, data, contentType, nil)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to upload file to storage")
		return
	}

	user := mw.GetAuthUser(r.Context())
	createdBy := ""
	if user != nil {
		createdBy = user.Subject
	}

	doc := &domain.Document{
		ID:                docID,
		MinioBucket:       result.Bucket,
		MinioKey:          result.Key,
		Filename:          filename,
		ContentType:       contentType,
		FileSize:          result.Size,
		SHA256Hash:        result.SHA256Hash,
		OwnerID:           ownerID,
		OwnerClassLibrary: ownerClassLibrary,
		OwnerClassName:    ownerClassName,
		AttachmentTypeID:  attachmentTypeID,
		AttachmentType:    attachmentType,
		CurrentVersion:    1,
		CreatedBy:         createdBy,
	}

	if err := h.repo.Upsert(r.Context(), doc); err != nil {
		api.JSONError(w, http.StatusInternalServerError, "DB_ERROR", "failed to save document record")
		return
	}

	api.JSON(w, http.StatusCreated, doc, nil)
}

// Update handles partial metadata update via JSON body.
func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid document ID")
		return
	}

	doc, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		api.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get document")
		return
	}
	if doc == nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	var body struct {
		Filename          *string `json:"filename"`
		ContentType       *string `json:"content_type"`
		OwnerID           *string `json:"owner_id"`
		OwnerClassLibrary *string `json:"owner_class_library"`
		OwnerClassName    *string `json:"owner_class_name"`
		AttachmentTypeID  *int    `json:"attachment_type_id"`
		AttachmentType    *string `json:"attachment_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		api.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	if body.Filename != nil {
		doc.Filename = *body.Filename
	}
	if body.ContentType != nil {
		doc.ContentType = *body.ContentType
	}
	if body.OwnerID != nil {
		doc.OwnerID = *body.OwnerID
	}
	if body.OwnerClassLibrary != nil {
		doc.OwnerClassLibrary = *body.OwnerClassLibrary
	}
	if body.OwnerClassName != nil {
		doc.OwnerClassName = *body.OwnerClassName
	}
	if body.AttachmentTypeID != nil {
		doc.AttachmentTypeID = *body.AttachmentTypeID
	}
	if body.AttachmentType != nil {
		doc.AttachmentType = *body.AttachmentType
	}

	if err := h.repo.Update(r.Context(), doc); err != nil {
		api.JSONError(w, http.StatusInternalServerError, "DB_ERROR", "failed to update document")
		return
	}

	api.JSON(w, http.StatusOK, doc, nil)
}

// Delete soft-deletes a document.
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		api.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid document ID")
		return
	}

	if err := h.repo.SoftDelete(r.Context(), id); err != nil {
		api.JSONError(w, http.StatusNotFound, "NOT_FOUND", "document not found")
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{"status": "deleted"}, nil)
}

// queryInt parses an integer query parameter with a default value.
func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}
