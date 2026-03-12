package events

import (
	"time"

	"github.com/google/uuid"
)

// Event subjects
const (
	SubjectDocumentCreated = "document.created"
	SubjectDocumentUpdated = "document.updated"
	SubjectDocumentDeleted = "document.deleted"
	SubjectSyncCompleted   = "sync.completed"
	SubjectSyncFailed      = "sync.failed"
)

// Event is the envelope published to NATS JetStream.
type Event struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// DocumentEventData carries document lifecycle event payloads.
type DocumentEventData struct {
	DocumentID uuid.UUID `json:"document_id"`
	OwnerID    string    `json:"owner_id"`
	Filename   string    `json:"filename"`
	Action     string    `json:"action"` // created, updated, deleted
	ActorID    string    `json:"actor_id,omitempty"`
}

// SyncEventData carries sync lifecycle event payloads.
type SyncEventData struct {
	DocumentID    *uuid.UUID `json:"document_id,omitempty"`
	LegacyOwnerID string     `json:"legacy_owner_id"`
	LegacyFileID  string     `json:"legacy_file_id"`
	Action        string     `json:"action"` // completed, failed
	ErrorMessage  string     `json:"error_message,omitempty"`
	DurationMs    int        `json:"duration_ms"`
}

// NewEvent creates an Event with a generated ID and current timestamp.
func NewEvent(subject string, data any) Event {
	return Event{
		ID:        uuid.New().String(),
		Subject:   subject,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}
