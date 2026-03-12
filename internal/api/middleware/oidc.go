package middleware

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OIDCVerifier validates OIDC ID tokens using provider discovery.
type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
}

// NewOIDCVerifier creates an OIDC verifier by performing provider discovery
// against the given issuerURL and configuring it for the given clientID.
func NewOIDCVerifier(ctx context.Context, issuerURL, clientID string) (*OIDCVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc provider discovery: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})
	return &OIDCVerifier{verifier: verifier}, nil
}

// OIDCClaims represents standard claims extracted from an OIDC ID token.
type OIDCClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

// Verify validates a raw ID token string and returns an AuthUser on success.
// OIDC users receive an empty permissions slice; permissions should be mapped
// from the subject via a separate authorization layer.
func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (*AuthUser, error) {
	if v.verifier == nil {
		return nil, fmt.Errorf("oidc verifier not initialized")
	}

	idToken, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("oidc verify: %w", err)
	}

	var claims OIDCClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc claims: %w", err)
	}

	return &AuthUser{
		Subject:     claims.Subject,
		Permissions: []string{}, // OIDC users get permissions from a separate mapping
		AuthMethod:  "oidc",
	}, nil
}
