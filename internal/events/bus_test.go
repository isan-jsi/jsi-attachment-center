package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	data := DocumentEventData{
		DocumentID: uuid.New(),
		OwnerID:    "owner-123",
		Filename:   "test.pdf",
		Action:     "created",
	}

	evt := NewEvent(SubjectDocumentCreated, data)

	assert.NotEmpty(t, evt.ID)
	assert.Equal(t, SubjectDocumentCreated, evt.Subject)
	assert.WithinDuration(t, time.Now().UTC(), evt.Timestamp, 2*time.Second)
	assert.Equal(t, data, evt.Data)
}

func TestEventJSONRoundTrip(t *testing.T) {
	docData := DocumentEventData{
		DocumentID: uuid.New(),
		OwnerID:    "owner-456",
		Filename:   "report.xlsx",
		Action:     "updated",
		ActorID:    "user-789",
	}

	evt := NewEvent(SubjectDocumentUpdated, docData)
	b, err := json.Marshal(evt)
	require.NoError(t, err)

	var decoded Event
	err = json.Unmarshal(b, &decoded)
	require.NoError(t, err)

	assert.Equal(t, evt.ID, decoded.ID)
	assert.Equal(t, evt.Subject, decoded.Subject)
}

func TestSyncEventDataJSON(t *testing.T) {
	docID := uuid.New()
	syncData := SyncEventData{
		DocumentID:    &docID,
		LegacyOwnerID: "legacy-owner",
		LegacyFileID:  "legacy-file",
		Action:        "completed",
		DurationMs:    1234,
	}

	b, err := json.Marshal(syncData)
	require.NoError(t, err)

	var decoded SyncEventData
	err = json.Unmarshal(b, &decoded)
	require.NoError(t, err)
	assert.Equal(t, syncData.LegacyOwnerID, decoded.LegacyOwnerID)
	assert.Equal(t, syncData.DurationMs, decoded.DurationMs)
}

func TestNoopEventBus_ImplementsInterface(t *testing.T) {
	var bus EventBus = NewNoopEventBus()

	err := bus.Publish(context.Background(), NewEvent(SubjectDocumentCreated, nil))
	assert.NoError(t, err)

	err = bus.Close()
	assert.NoError(t, err)
}
