package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOIDCVerifier_Verify_NilVerifier ensures Verify returns an error when the
// underlying oidc.IDTokenVerifier has not been initialized.
func TestOIDCVerifier_Verify_NilVerifier(t *testing.T) {
	v := &OIDCVerifier{verifier: nil}
	_, err := v.Verify(context.Background(), "some-token")
	assert.Error(t, err)
}

// TestAuth_BearerToken_NoJWT_NoOIDC_Returns401 confirms that a Bearer token
// presented to a middleware with no JWT key and no OIDC verifier results in
// a 401 response and the downstream handler is never called.
func TestAuth_BearerToken_NoJWT_NoOIDC_Returns401(t *testing.T) {
	mw := Auth(AuthConfig{})

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer some-invalid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestAuth_BearerToken_OIDCVerifier_Nil_Returns401 verifies that when
// OIDCVerifier is explicitly set to nil (not configured), a bad Bearer token
// still results in 401.
func TestAuth_BearerToken_OIDCVerifier_Nil_Returns401(t *testing.T) {
	mw := Auth(AuthConfig{OIDCVerifier: nil})

	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
