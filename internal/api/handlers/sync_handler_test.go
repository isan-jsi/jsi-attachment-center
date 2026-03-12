package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jsi/ibs-doc-engine/internal/api"
	"github.com/jsi/ibs-doc-engine/internal/api/handlers"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncHandler_RetryDLQ_InvalidUUID(t *testing.T) {
	r := chi.NewRouter()
	h := handlers.NewSyncHandler(nil)
	r.Post("/{id}/retry", h.RetryDLQ)

	req := httptest.NewRequest("POST", "/bad-uuid/retry", nil)
	ctx := context.WithValue(req.Context(), mw.AuthUserKey, &mw.AuthUser{Subject: "test"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var env api.Envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.Equal(t, "INVALID_ID", env.Error.Code)
}
