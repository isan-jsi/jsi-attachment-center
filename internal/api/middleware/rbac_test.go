package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/stretchr/testify/assert"
)

func withAuthUser(r *http.Request, user *mw.AuthUser) *http.Request {
	ctx := context.WithValue(r.Context(), mw.AuthUserKey, user)
	return r.WithContext(ctx)
}

func TestRequirePermission_NoUser_Returns401(t *testing.T) {
	handler := mw.RequirePermission("documents:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_MissingPerm_Returns403(t *testing.T) {
	handler := mw.RequirePermission("documents:write")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	req = withAuthUser(req, &mw.AuthUser{Subject: "u1", Permissions: []string{"documents:read"}})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_ExactMatch_Passes(t *testing.T) {
	handler := mw.RequirePermission("documents:read")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	req = withAuthUser(req, &mw.AuthUser{Subject: "u1", Permissions: []string{"documents:read"}})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_WildcardResource_Passes(t *testing.T) {
	handler := mw.RequirePermission("documents:write")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	req = withAuthUser(req, &mw.AuthUser{Subject: "admin", Permissions: []string{"documents:*"}})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_GlobalWildcard_Passes(t *testing.T) {
	handler := mw.RequirePermission("sync:admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	req = withAuthUser(req, &mw.AuthUser{Subject: "superadmin", Permissions: []string{"*"}})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyPermission_OneMatches_Passes(t *testing.T) {
	handler := mw.RequireAnyPermission("documents:read", "documents:write")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/", nil)
	req = withAuthUser(req, &mw.AuthUser{Subject: "u1", Permissions: []string{"documents:read"}})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
