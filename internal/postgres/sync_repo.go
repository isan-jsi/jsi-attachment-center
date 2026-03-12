package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jsi/ibs-doc-engine/internal/domain"
)

type SyncRepo struct {
	pool *pgxpool.Pool
}

func NewSyncRepo(pool *pgxpool.Pool) *SyncRepo {
	return &SyncRepo{pool: pool}
}

func (r *SyncRepo) GetCheckpoint(ctx context.Context, tableName string) (*domain.SyncCheckpoint, error) {
	query := `SELECT table_name, last_dc_check, last_sync_at, records_processed, status
		FROM sync_checkpoints WHERE table_name = $1`

	var cp domain.SyncCheckpoint
	err := r.pool.QueryRow(ctx, query, tableName).Scan(
		&cp.TableName, &cp.LastDCCheck, &cp.LastSyncAt, &cp.RecordsProcessed, &cp.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sync repo: get checkpoint: %w", err)
	}
	return &cp, nil
}

func (r *SyncRepo) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	query := `INSERT INTO sync_checkpoints (table_name, last_dc_check, last_sync_at, records_processed, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (table_name) DO UPDATE SET
			last_dc_check = EXCLUDED.last_dc_check,
			last_sync_at = EXCLUDED.last_sync_at,
			records_processed = sync_checkpoints.records_processed + EXCLUDED.records_processed,
			status = EXCLUDED.status`

	_, err := r.pool.Exec(ctx, query, cp.TableName, cp.LastDCCheck, cp.LastSyncAt, cp.RecordsProcessed, cp.Status)
	if err != nil {
		return fmt.Errorf("sync repo: save checkpoint: %w", err)
	}
	return nil
}

func (r *SyncRepo) InsertLog(ctx context.Context, entry *domain.SyncLogEntry) error {
	query := `INSERT INTO sync_log (id, document_id, legacy_owner_id, legacy_file_id, action, status, error_message, duration_ms, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.DocumentID, entry.LegacyOwnerID, entry.LegacyFileID,
		entry.Action, entry.Status, entry.ErrorMessage, entry.DurationMs, entry.SyncedAt,
	)
	if err != nil {
		return fmt.Errorf("sync repo: insert log: %w", err)
	}
	return nil
}

func (r *SyncRepo) UpsertDLQ(ctx context.Context, entry *domain.DLQEntry) error {
	query := `INSERT INTO sync_dlq (id, legacy_owner_id, legacy_file_id, table_name, retry_count, max_retries, next_retry_at, last_error, payload_json, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (legacy_owner_id, legacy_file_id, table_name) DO UPDATE SET
			retry_count = EXCLUDED.retry_count,
			next_retry_at = EXCLUDED.next_retry_at,
			last_error = EXCLUDED.last_error,
			status = EXCLUDED.status,
			updated_at = NOW()`

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	payloadBytes, err := json.Marshal(entry.PayloadJSON)
	if err != nil {
		return fmt.Errorf("sync repo: marshal dlq payload: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		entry.ID, entry.LegacyOwnerID, entry.LegacyFileID, entry.TableName,
		entry.RetryCount, entry.MaxRetries, entry.NextRetryAt, entry.LastError,
		payloadBytes, entry.Status,
	)
	if err != nil {
		return fmt.Errorf("sync repo: upsert dlq: %w", err)
	}
	return nil
}

func (r *SyncRepo) FetchRetryableDLQ(ctx context.Context, limit int) ([]domain.DLQEntry, error) {
	query := `SELECT id, legacy_owner_id, legacy_file_id, table_name, retry_count, max_retries, next_retry_at, last_error, payload_json, status, created_at, updated_at
		FROM sync_dlq
		WHERE status IN ('pending', 'retrying') AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY next_retry_at ASC NULLS FIRST
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("sync repo: fetch retryable dlq: %w", err)
	}
	defer rows.Close()

	var entries []domain.DLQEntry
	for rows.Next() {
		var e domain.DLQEntry
		err := rows.Scan(
			&e.ID, &e.LegacyOwnerID, &e.LegacyFileID, &e.TableName,
			&e.RetryCount, &e.MaxRetries, &e.NextRetryAt, &e.LastError,
			&e.PayloadJSON, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("sync repo: scan dlq: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// CalculateBackoff returns exponential backoff duration: 2^retryCount * 30 seconds (max 30 min).
func CalculateBackoff(retryCount int) time.Duration {
	base := 30 * time.Second
	backoff := base * (1 << retryCount)
	max := 30 * time.Minute
	if backoff > max {
		return max
	}
	return backoff
}
