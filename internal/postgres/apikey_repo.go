package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"github.com/jsi/ibs-doc-engine/internal/domain"
)

// APIKeyCreateResult holds the newly-created key plus the one-time plaintext.
type APIKeyCreateResult struct {
	domain.APIKey
	PlaintextKey string `json:"key"`
}

type APIKeyRepo struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{pool: pool}
}

// HashKey returns the SHA-256 hex digest of a raw API key.
func HashKey(raw string) string {
	return domain.HashAPIKey(raw)
}

// generateAPIKey creates a random 32-byte key and returns (plaintext, sha256hex, prefix).
func generateAPIKey() (string, string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("generate key: %w", err)
	}
	plaintext := "ibsde_" + hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(plaintext))
	hashHex := hex.EncodeToString(hash[:])
	prefix := plaintext[:8]
	return plaintext, hashHex, prefix, nil
}

// GetByHash looks up an active (non-revoked) API key by its hash.
func (r *APIKeyRepo) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	query := `SELECT id, key_hash, key_prefix, name, owner_id, permissions,
		created_at, updated_at, expires_at, last_used_at, revoked_at
		FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL`

	var k domain.APIKey
	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.OwnerID, &k.Permissions,
		&k.CreatedAt, &k.UpdatedAt, &k.ExpiresAt, &k.LastUsedAt, &k.RevokedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("apikey repo: get by hash: %w", err)
	}
	return &k, nil
}

// Create generates a new API key, stores its hash, and returns the result including plaintext.
func (r *APIKeyRepo) Create(ctx context.Context, name, ownerID string, permissions []string, expiresAt *time.Time) (*APIKeyCreateResult, error) {
	plaintext, hashHex, prefix, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("apikey repo: create: %w", err)
	}

	permJSON, err := json.Marshal(permissions)
	if err != nil {
		return nil, fmt.Errorf("apikey repo: create: marshal permissions: %w", err)
	}

	query := `INSERT INTO api_keys (key_hash, key_prefix, name, owner_id, permissions, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, key_hash, key_prefix, name, owner_id, permissions,
		          created_at, updated_at, expires_at, last_used_at, revoked_at`

	var k domain.APIKey
	err = r.pool.QueryRow(ctx, query, hashHex, prefix, name, ownerID, permJSON, expiresAt).Scan(
		&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.OwnerID, &k.Permissions,
		&k.CreatedAt, &k.UpdatedAt, &k.ExpiresAt, &k.LastUsedAt, &k.RevokedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("apikey repo: create: insert: %w", err)
	}

	return &APIKeyCreateResult{APIKey: k, PlaintextKey: plaintext}, nil
}

// ListByOwner returns all non-revoked API keys for the given owner.
func (r *APIKeyRepo) ListByOwner(ctx context.Context, ownerID string) ([]domain.APIKey, error) {
	query := `SELECT id, key_hash, key_prefix, name, owner_id, permissions,
		created_at, updated_at, expires_at, last_used_at, revoked_at
		FROM api_keys WHERE owner_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("apikey repo: list by owner: %w", err)
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		if err := rows.Scan(
			&k.ID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.OwnerID, &k.Permissions,
			&k.CreatedAt, &k.UpdatedAt, &k.ExpiresAt, &k.LastUsedAt, &k.RevokedAt,
		); err != nil {
			return nil, fmt.Errorf("apikey repo: list by owner: scan: %w", err)
		}
		keys = append(keys, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("apikey repo: list by owner: rows: %w", err)
	}
	return keys, nil
}

// Revoke sets revoked_at = NOW() for the given key, scoped to the owner.
func (r *APIKeyRepo) Revoke(ctx context.Context, id uuid.UUID, ownerID string) error {
	query := `UPDATE api_keys SET revoked_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND owner_id = $2 AND revoked_at IS NULL`

	tag, err := r.pool.Exec(ctx, query, id, ownerID)
	if err != nil {
		return fmt.Errorf("apikey repo: revoke: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("apikey repo: revoke: key not found or already revoked")
	}
	return nil
}

// Update patches name, permissions, and/or expires_at for a key scoped to the owner.
func (r *APIKeyRepo) Update(ctx context.Context, id uuid.UUID, ownerID string, name *string, permissions *[]string, expiresAt *time.Time) error {
	// Build dynamic SET clause
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *name)
		argIdx++
	}
	if permissions != nil {
		permJSON, err := json.Marshal(*permissions)
		if err != nil {
			return fmt.Errorf("apikey repo: update: marshal permissions: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("permissions = $%d", argIdx))
		args = append(args, permJSON)
		argIdx++
	}
	if expiresAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("expires_at = $%d", argIdx))
		args = append(args, expiresAt)
		argIdx++
	}

	// Append WHERE args
	args = append(args, id, ownerID)
	idIdx := argIdx
	ownerIdx := argIdx + 1

	setStr := ""
	for i, c := range setClauses {
		if i > 0 {
			setStr += ", "
		}
		setStr += c
	}

	query := fmt.Sprintf(
		"UPDATE api_keys SET %s WHERE id = $%d AND owner_id = $%d AND revoked_at IS NULL",
		setStr, idIdx, ownerIdx,
	)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("apikey repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("apikey repo: update: key not found or revoked")
	}
	return nil
}

// TouchLastUsed updates the last_used_at timestamp for a key.
func (r *APIKeyRepo) TouchLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("apikey repo: touch last used: %w", err)
	}
	return nil
}
