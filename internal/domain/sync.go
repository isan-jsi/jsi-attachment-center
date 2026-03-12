package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SyncRecord struct {
	LegacyAttachment  LegacyAttachment
	OwnerClassLibrary string
	OwnerClassName    string
	MinioKey          string
	SHA256Hash        string
}

type SyncCheckpoint struct {
	TableName        string
	LastDCCheck      []byte
	LastSyncAt       time.Time
	RecordsProcessed int64
	Status           string
}

type SyncLogEntry struct {
	ID            uuid.UUID
	DocumentID    *uuid.UUID
	LegacyOwnerID string
	LegacyFileID  string
	Action        string // create, update, delete, error
	Status        string // success, failed, retrying, dlq
	ErrorMessage  string
	DurationMs    int
	SyncedAt      time.Time
}

type DLQEntry struct {
	ID            uuid.UUID
	LegacyOwnerID string
	LegacyFileID  string
	TableName     string
	RetryCount    int
	MaxRetries    int
	NextRetryAt   *time.Time
	LastError     string
	PayloadJSON   json.RawMessage
	Status        string // pending, retrying, exhausted, resolved
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
