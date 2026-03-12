package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID          uuid.UUID
	KeyHash     string
	Name        string
	Permissions json.RawMessage // ["documents:read","documents:write",...]
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	RevokedAt   *time.Time
}

// IsValid returns true if the key is not expired and not revoked.
func (k *APIKey) IsValid() bool {
	if k.RevokedAt != nil {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}
