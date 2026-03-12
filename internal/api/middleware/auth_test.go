package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeyPair(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func signTestJWT(t *testing.T, key *rsa.PrivateKey, sub string, perms []string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":         sub,
		"permissions": perms,
		"exp":         time.Now().Add(time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func marshalRSAPublicKeyPEM(t *testing.T, pub *rsa.PublicKey) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

func TestAuth_MissingCredentials_Returns401(t *testing.T) {
	handler := mw.Auth(mw.AuthConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidBearerToken_Returns401(t *testing.T) {
	key := generateTestKeyPair(t)
	pubPEM := marshalRSAPublicKeyPEM(t, &key.PublicKey)

	handler := mw.Auth(mw.AuthConfig{JWTPublicKeyPEM: string(pubPEM)})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ValidJWT_SetsAuthUser(t *testing.T) {
	key := generateTestKeyPair(t)
	pubPEM := marshalRSAPublicKeyPEM(t, &key.PublicKey)
	tokenStr := signTestJWT(t, key, "user123", []string{"documents:read"})

	var gotUser *mw.AuthUser
	handler := mw.Auth(mw.AuthConfig{JWTPublicKeyPEM: string(pubPEM)})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUser = mw.GetAuthUser(r.Context())
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, gotUser)
	assert.Equal(t, "user123", gotUser.Subject)
	assert.Equal(t, "jwt", gotUser.AuthMethod)
	assert.Contains(t, gotUser.Permissions, "documents:read")
}
