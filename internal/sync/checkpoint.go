package sync

import (
	"context"
	"log/slog"
	"time"

	"github.com/jsi/ibs-doc-engine/internal/domain"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
)

type CheckpointManager struct {
	repo      *postgres.SyncRepo
	tableName string
}

func NewCheckpointManager(repo *postgres.SyncRepo, tableName string) *CheckpointManager {
	return &CheckpointManager{repo: repo, tableName: tableName}
}

func (m *CheckpointManager) Load(ctx context.Context) ([]byte, error) {
	cp, err := m.repo.GetCheckpoint(ctx, m.tableName)
	if err != nil {
		return nil, err
	}
	if cp == nil {
		slog.Info("checkpoint: no existing checkpoint, starting from beginning", "table", m.tableName)
		return nil, nil
	}
	slog.Info("checkpoint: loaded",
		"table", m.tableName,
		"last_sync_at", cp.LastSyncAt,
		"records_processed", cp.RecordsProcessed,
	)
	return cp.LastDCCheck, nil
}

func (m *CheckpointManager) Save(ctx context.Context, lastDCCheck []byte, batchCount int64) error {
	cp := &domain.SyncCheckpoint{
		TableName:        m.tableName,
		LastDCCheck:      lastDCCheck,
		LastSyncAt:       time.Now(),
		RecordsProcessed: batchCount,
		Status:           "active",
	}
	return m.repo.SaveCheckpoint(ctx, cp)
}
