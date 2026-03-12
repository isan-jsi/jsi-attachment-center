package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jsi/ibs-doc-engine/internal/domain"
	miniosvc "github.com/jsi/ibs-doc-engine/internal/minio"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
	"github.com/jsi/ibs-doc-engine/internal/sqlserver"
)

type Pipeline struct {
	attReader   *sqlserver.AttachmentReader
	minioClient *miniosvc.Client
	docRepo     *postgres.DocumentRepo
	syncRepo    *postgres.SyncRepo
	checkpoint  *CheckpointManager
	chain       *TransformerChain
	batchSize   int
	workers     int
}

type PipelineConfig struct {
	AttachmentReader *sqlserver.AttachmentReader
	MinIOClient      *miniosvc.Client
	DocumentRepo     *postgres.DocumentRepo
	SyncRepo         *postgres.SyncRepo
	TableName        string
	BatchSize        int
	Workers          int
}

func NewPipeline(cfg PipelineConfig) *Pipeline {
	chain := NewTransformerChain(
		NewContentTypeDetector(),
		NewPathGenerator(),
		NewHashComputer(),
	)

	return &Pipeline{
		attReader:   cfg.AttachmentReader,
		minioClient: cfg.MinIOClient,
		docRepo:     cfg.DocumentRepo,
		syncRepo:    cfg.SyncRepo,
		checkpoint:  NewCheckpointManager(cfg.SyncRepo, cfg.TableName),
		chain:       chain,
		batchSize:   cfg.BatchSize,
		workers:     cfg.Workers,
	}
}

// RunOnce executes a single sync cycle: discover → extract → transform → upload → register → verify.
func (p *Pipeline) RunOnce(ctx context.Context) (int, error) {
	// 1. Load checkpoint
	lastDCCheck, err := p.checkpoint.Load(ctx)
	if err != nil {
		return 0, fmt.Errorf("pipeline: load checkpoint: %w", err)
	}

	// 2. Discover new/updated records
	attachments, err := p.attReader.FetchAfterCheckpoint(ctx, lastDCCheck, p.batchSize)
	if err != nil {
		return 0, fmt.Errorf("pipeline: fetch attachments: %w", err)
	}

	if len(attachments) == 0 {
		slog.Debug("pipeline: no new records")
		return 0, nil
	}

	slog.Info("pipeline: discovered records", "count", len(attachments))

	// 3. Process each attachment
	processed := 0
	var lastProcessedDCCheck []byte

	for _, att := range attachments {
		select {
		case <-ctx.Done():
			return processed, ctx.Err()
		default:
		}

		if err := p.processAttachment(ctx, att); err != nil {
			slog.Error("pipeline: process failed", "owner_id", att.OwnerID, "file_id", att.FileID, "error", err)
			p.handleFailure(ctx, att, err)
			continue
		}

		processed++
		lastProcessedDCCheck = att.DCCheck
	}

	// 4. Save checkpoint
	if lastProcessedDCCheck != nil {
		if err := p.checkpoint.Save(ctx, lastProcessedDCCheck, int64(processed)); err != nil {
			return processed, fmt.Errorf("pipeline: save checkpoint: %w", err)
		}
	}

	slog.Info("pipeline: cycle complete", "processed", processed, "total_discovered", len(attachments))
	return processed, nil
}

func (p *Pipeline) processAttachment(ctx context.Context, att domain.LegacyAttachment) error {
	start := time.Now()

	// Look up owner info
	owner, err := p.attReader.FetchOwnerForAttachment(ctx, att.DocAttachmentTypeID)
	if err != nil {
		return fmt.Errorf("fetch owner: %w", err)
	}

	// Build sync record
	record := &domain.SyncRecord{
		LegacyAttachment:  att,
		OwnerClassLibrary: owner.OwnerClassLibrary,
		OwnerClassName:    owner.OwnerClassName,
	}

	// Run transformer chain
	record, err = p.chain.Transform(ctx, record)
	if err != nil {
		return fmt.Errorf("transform: %w", err)
	}

	// Upload to MinIO
	uploadResult, err := p.minioClient.Upload(ctx, record.MinioKey, att.FileContent, att.ContentType, map[string]string{
		"legacy-owner-id": att.OwnerID,
		"legacy-file-id":  att.FileID,
		"synced-at":       time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	// Verify integrity
	if err := Verify(record.SHA256Hash, att.FileSize, uploadResult.SHA256Hash, uploadResult.Size); err != nil {
		return fmt.Errorf("verify: %w", err)
	}

	// Register in PostgreSQL
	doc := &domain.Document{
		MinioBucket:       uploadResult.Bucket,
		MinioKey:          uploadResult.Key,
		Filename:          att.FileName,
		ContentType:       att.ContentType,
		FileSize:          att.FileSize,
		SHA256Hash:        uploadResult.SHA256Hash,
		OwnerID:           att.OwnerID,
		OwnerClassLibrary: owner.OwnerClassLibrary,
		OwnerClassName:    owner.OwnerClassName,
		AttachmentTypeID:  att.DocAttachmentTypeID,
		AttachmentType:    att.DocAttachmentType,
		IsExternal:        att.IsExternal,
		LegacyFileID:      att.FileID,
		CurrentVersion:    1,
		CreatedBy:         att.CreatedBy,
		CreatedAt:         att.CreatedOn,
		UpdatedAt:         time.Now(),
	}

	if err := p.docRepo.Upsert(ctx, doc); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// Log success
	duration := time.Since(start)
	p.syncRepo.InsertLog(ctx, &domain.SyncLogEntry{
		DocumentID:    &doc.ID,
		LegacyOwnerID: att.OwnerID,
		LegacyFileID:  att.FileID,
		Action:        "create",
		Status:        "success",
		DurationMs:    int(duration.Milliseconds()),
		SyncedAt:      time.Now(),
	})

	return nil
}

func (p *Pipeline) handleFailure(ctx context.Context, att domain.LegacyAttachment, syncErr error) {
	// Log the failure
	p.syncRepo.InsertLog(ctx, &domain.SyncLogEntry{
		LegacyOwnerID: att.OwnerID,
		LegacyFileID:  att.FileID,
		Action:        "error",
		Status:        "failed",
		ErrorMessage:  syncErr.Error(),
		SyncedAt:      time.Now(),
	})

	// Add to DLQ
	payload, _ := json.Marshal(map[string]interface{}{
		"owner_id": att.OwnerID,
		"file_id":  att.FileID,
		"type_id":  att.DocAttachmentTypeID,
	})

	retryAt := time.Now().Add(postgres.CalculateBackoff(0))
	p.syncRepo.UpsertDLQ(ctx, &domain.DLQEntry{
		ID:            uuid.New(),
		LegacyOwnerID: att.OwnerID,
		LegacyFileID:  att.FileID,
		TableName:     "IBSDocAttachments",
		RetryCount:    0,
		MaxRetries:    5,
		NextRetryAt:   &retryAt,
		LastError:     syncErr.Error(),
		PayloadJSON:   payload,
		Status:        "pending",
	})
}

// RunLoop starts the continuous sync loop with the configured poll interval.
func (p *Pipeline) RunLoop(ctx context.Context, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	slog.Info("pipeline: starting sync loop", "interval", pollInterval)

	// Run immediately on start
	if _, err := p.RunOnce(ctx); err != nil {
		slog.Error("pipeline: initial sync failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("pipeline: sync loop stopped")
			return
		case <-ticker.C:
			if _, err := p.RunOnce(ctx); err != nil {
				slog.Error("pipeline: sync cycle failed", "error", err)
			}
		}
	}
}
