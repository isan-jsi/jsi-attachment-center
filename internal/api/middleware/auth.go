package middleware

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jsi/ibs-doc-engine/internal/domain"
)

// jsonError writes a JSON error response without importing the api package (to avoid cycles).
func jsonError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// AuthUser represents the authenticated identity stored in context.
type AuthUser struct {
	Subject     string   // JWT "sub" or API key name
	Permissions []string // from JWT claims or API key permissions
	AuthMethod  string   // "jwt" or "apikey"
}

const AuthUserKey contextKey = "auth_user"

// GetAuthUser extracts the authenticated user from context.
func GetAuthUser(ctx context.Context) *AuthUser {
	if u, ok := ctx.Value(AuthUserKey).(*AuthUser); ok {
		return u
	}
	return nil
}

// APIKeyQuerier abstracts API key lookups for testability.
type APIKeyQuerier interface {
	GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
}

// AuthConfig configures the dual auth middleware.
type AuthConfig struct {
	JWTPublicKeyPEM string // path to PEM file or raw PEM string
	APIKeyRepo      APIKeyQuerier
}

// Auth returns middleware that validates JWT Bearer tokens or X-API-Key headers.
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	var pubKey *rsa.PublicKey
	if cfg.JWTPublicKeyPEM != "" {
		var err error
		pubKey, err = loadRSAPublicKey(cfg.JWTPublicKeyPEM)
		if err != nil {
			panic(fmt.Sprintf("auth middleware: load JWT public key: %v", err))
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try JWT first
			if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				user, err := validateJWT(tokenStr, pubKey)
				if err != nil {
					jsonError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired JWT")
					return
				}
				ctx := context.WithValue(r.Context(), AuthUserKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Try API key
			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" && cfg.APIKeyRepo != nil {
				user, err := validateAPIKey(r.Context(), apiKey, cfg.APIKeyRepo)
				if err != nil {
					jsonError(w, http.StatusUnauthorized, "INVALID_API_KEY", "invalid or expired API key")
					return
				}
				ctx := context.WithValue(r.Context(), AuthUserKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			jsonError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing Authorization header or X-API-Key")
		})
	}
}

func validateJWT(tokenStr string, pubKey *rsa.PublicKey) (*AuthUser, error) {
	if pubKey == nil {
		return nil, fmt.Errorf("JWT validation not configured")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid claims")
	}

	sub, _ := claims.GetSubject()
	var perms []string
	if p, ok := claims["permissions"].([]interface{}); ok {
		for _, v := range p {
			if s, ok := v.(string); ok {
				perms = append(perms, s)
			}
		}
	}

	return &AuthUser{
		Subject:     sub,
		Permissions: perms,
		AuthMethod:  "jwt",
	}, nil
}

// hashKey computes the SHA-256 hex digest of a raw API key.
func hashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

func validateAPIKey(ctx context.Context, raw string, repo APIKeyQuerier) (*AuthUser, error) {
	hash := hashKey(raw)
	key, err := repo.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if key == nil || !key.IsValid() {
		return nil, fmt.Errorf("api key not found or invalid")
	}

	var perms []string
	if len(key.Permissions) > 0 {
		if err := json.Unmarshal(key.Permissions, &perms); err != nil {
			slog.Warn("auth: failed to parse API key permissions", "key_name", key.Name, "error", err)
		}
	}

	return &AuthUser{
		Subject:     key.Name,
		Permissions: perms,
		AuthMethod:  "apikey",
	}, nil
}

func loadRSAPublicKey(pemPathOrRaw string) (*rsa.PublicKey, error) {
	var pemData []byte
	if _, err := os.Stat(pemPathOrRaw); err == nil {
		pemData, err = os.ReadFile(pemPathOrRaw)
		if err != nil {
			return nil, fmt.Errorf("read PEM file: %w", err)
		}
	} else {
		pemData = []byte(pemPathOrRaw)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA")
	}
	return rsaPub, nil
}
