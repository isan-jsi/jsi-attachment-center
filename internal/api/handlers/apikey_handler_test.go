package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jsi/ibs-doc-engine/internal/api"
	"github.com/jsi/ibs-doc-engine/internal/api/handlers"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeyHandler_Create_MissingName verifies that a 400 is returned when name is empty.
func TestAPIKeyHandler_Create_MissingName(t *testing.T) {
	h := handlers.NewAPIKeyHandler(nil)

	r := chi.NewRouter()
	r.Post("/", h.Create)

	body := `{"name":"","permissions":[]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// inject an authenticated user so we get past the auth check
	ctx := context.WithValue(req.Context(), mw.AuthUserKey, &mw.AuthUser{Subject: "user-123"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var env api.Envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, "VALIDATION_ERROR", env.Error.Code)
}

// TestAPIKeyHandler_Create_NoAuth verifies that 401 is returned when no auth context is present.
func TestAPIKeyHandler_Create_NoAuth(t *testing.T) {
	h := handlers.NewAPIKeyHandler(nil)

	r := chi.NewRouter()
	r.Post("/", h.Create)

	body := `{"name":"my-key","permissions":[]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var env api.Envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, "UNAUTHORIZED", env.Error.Code)
}

// TestAPIKeyHandler_Revoke_InvalidID verifies that a 400 is returned for a non-UUID id.
func TestAPIKeyHandler_Revoke_InvalidID(t *testing.T) {
	h := handlers.NewAPIKeyHandler(nil)

	r := chi.NewRouter()
	r.Delete("/{id}", h.Revoke)

	req := httptest.NewRequest(http.MethodDelete, "/not-a-uuid", nil)
	ctx := context.WithValue(req.Context(), mw.AuthUserKey, &mw.AuthUser{Subject: "user-123"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var env api.Envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, "INVALID_ID", env.Error.Code)
}
