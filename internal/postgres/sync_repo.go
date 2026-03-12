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
	query := `SELECT table_name, owner_class_key, last_dc_check, last_sync_at, records_processed, status
		FROM sync_checkpoints WHERE table_name = $1`

	var cp domain.SyncCheckpoint
	var ownerClassKey *string
	err := r.pool.QueryRow(ctx, query, tableName).Scan(
		&cp.TableName, &ownerClassKey, &cp.LastDCCheck, &cp.LastSyncAt, &cp.RecordsProcessed, &cp.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sync repo: get checkpoint: %w", err)
	}
	if ownerClassKey != nil {
		cp.OwnerClassKey = *ownerClassKey
	}
	return &cp, nil
}

func (r *SyncRepo) SaveCheckpoint(ctx context.Context, cp *domain.SyncCheckpoint) error {
	query := `INSERT INTO sync_checkpoints (table_name, owner_class_key, last_dc_check, last_sync_at, records_processed, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (table_name) DO UPDATE SET
			owner_class_key = EXCLUDED.owner_class_key,
			last_dc_check = EXCLUDED.last_dc_check,
			last_sync_at = EXCLUDED.last_sync_at,
			records_processed = sync_checkpoints.records_processed + EXCLUDED.records_processed,
			status = EXCLUDED.status`

	_, err := r.pool.Exec(ctx, query, cp.TableName, cp.OwnerClassKey, cp.LastDCCheck, cp.LastSyncAt, cp.RecordsProcessed, cp.Status)
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

// MarkDLQStatus updates the status and retry metadata for a DLQ entry.
func (r *SyncRepo) MarkDLQStatus(ctx context.Context, id uuid.UUID, status string, retryCount int, nextRetryAt *time.Time, lastError string) error {
	query := `UPDATE sync_dlq
		SET status = $2, retry_count = $3, next_retry_at = $4, last_error = $5, updated_at = NOW()
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id, status, retryCount, nextRetryAt, lastError)
	if err != nil {
		return fmt.Errorf("sync repo: mark dlq status: %w", err)
	}
	return nil
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

// SyncStatus is an aggregated view of sync health.
type SyncStatus struct {
	Checkpoints  []domain.SyncCheckpoint `json:"checkpoints"`
	DLQPending   int64                   `json:"dlq_pending"`
	DLQExhausted int64                   `json:"dlq_exhausted"`
	DLQTotal     int64                   `json:"dlq_total"`
}

func (r *SyncRepo) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	cpQuery := `SELECT table_name, last_dc_check, last_sync_at, records_processed, status
		FROM sync_checkpoints ORDER BY table_name`
	rows, err := r.pool.Query(ctx, cpQuery)
	if err != nil {
		return nil, fmt.Errorf("sync repo: get status checkpoints: %w", err)
	}
	defer rows.Close()

	var checkpoints []domain.SyncCheckpoint
	for rows.Next() {
		var cp domain.SyncCheckpoint
		if err := rows.Scan(&cp.TableName, &cp.LastDCCheck, &cp.LastSyncAt, &cp.RecordsProcessed, &cp.Status); err != nil {
			return nil, fmt.Errorf("sync repo: scan checkpoint: %w", err)
		}
		checkpoints = append(checkpoints, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var pending, exhausted, total int64
	dlqQuery := `SELECT
		COUNT(*) FILTER (WHERE status IN ('pending','retrying')) AS pending,
		COUNT(*) FILTER (WHERE status = 'exhausted') AS exhausted,
		COUNT(*) AS total
	FROM sync_dlq`
	if err := r.pool.QueryRow(ctx, dlqQuery).Scan(&pending, &exhausted, &total); err != nil {
		return nil, fmt.Errorf("sync repo: dlq counts: %w", err)
	}

	return &SyncStatus{
		Checkpoints:  checkpoints,
		DLQPending:   pending,
		DLQExhausted: exhausted,
		DLQTotal:     total,
	}, nil
}

// ListDLQ returns paginated DLQ entries.
func (r *SyncRepo) ListDLQ(ctx context.Context, page, perPage int) ([]domain.DLQEntry, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM sync_dlq").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("sync repo: count dlq: %w", err)
	}

	if perPage <= 0 {
		perPage = 20
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	query := `SELECT id, legacy_owner_id, legacy_file_id, table_name, retry_count, max_retries,
		next_retry_at, last_error, payload_json, status, created_at, updated_at
	FROM sync_dlq ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("sync repo: list dlq: %w", err)
	}
	defer rows.Close()

	var entries []domain.DLQEntry
	for rows.Next() {
		var e domain.DLQEntry
		if err := rows.Scan(
			&e.ID, &e.LegacyOwnerID, &e.LegacyFileID, &e.TableName,
			&e.RetryCount, &e.MaxRetries, &e.NextRetryAt, &e.LastError,
			&e.PayloadJSON, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("sync repo: scan dlq list: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, total, rows.Err()
}

// RetryDLQEntry resets a DLQ entry for immediate retry.
func (r *SyncRepo) RetryDLQEntry(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sync_dlq SET status = 'pending', next_retry_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'retrying', 'exhausted')`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("sync repo: retry dlq: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("sync repo: retry dlq: not found or already resolved")
	}
	return nil
}
