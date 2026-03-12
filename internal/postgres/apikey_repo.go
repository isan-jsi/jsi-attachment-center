package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jsi/ibs-doc-engine/internal/domain"
)

type APIKeyRepo struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{pool: pool}
}

// HashKey returns the SHA-256 hex digest of a raw API key.
func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}

// GetByHash looks up an active (non-revoked) API key by its hash.
func (r *APIKeyRepo) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	query := `SELECT id, key_hash, name, permissions, created_at, expires_at, revoked_at
		FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL`

	var k domain.APIKey
	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&k.ID, &k.KeyHash, &k.Name, &k.Permissions,
		&k.CreatedAt, &k.ExpiresAt, &k.RevokedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("apikey repo: get by hash: %w", err)
	}
	return &k, nil
}
