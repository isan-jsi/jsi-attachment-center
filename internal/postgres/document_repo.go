package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jsi/ibs-doc-engine/internal/domain"
)

type DocumentRepo struct {
	pool *pgxpool.Pool
}

func NewDocumentRepo(pool *pgxpool.Pool) *DocumentRepo {
	return &DocumentRepo{pool: pool}
}

func (r *DocumentRepo) Upsert(ctx context.Context, doc *domain.Document) error {
	query := `
		INSERT INTO documents (
			id, minio_bucket, minio_key, filename, content_type, file_size, sha256_hash,
			owner_id, owner_class_library, owner_class_name,
			attachment_type_id, attachment_type, is_external, legacy_file_id,
			current_version, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, $18
		)
		ON CONFLICT (minio_bucket, minio_key) DO UPDATE SET
			sha256_hash = EXCLUDED.sha256_hash,
			file_size = EXCLUDED.file_size,
			content_type = EXCLUDED.content_type,
			updated_at = NOW()`

	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}

	_, err := r.pool.Exec(ctx, query,
		doc.ID, doc.MinioBucket, doc.MinioKey, doc.Filename, doc.ContentType, doc.FileSize, doc.SHA256Hash,
		doc.OwnerID, doc.OwnerClassLibrary, doc.OwnerClassName,
		doc.AttachmentTypeID, doc.AttachmentType, doc.IsExternal, doc.LegacyFileID,
		doc.CurrentVersion, doc.CreatedBy, doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("document repo: upsert: %w", err)
	}
	return nil
}

func (r *DocumentRepo) GetByLegacyID(ctx context.Context, ownerID, fileID string) (*domain.Document, error) {
	query := `SELECT
		id, minio_bucket, minio_key, filename, content_type, file_size, sha256_hash,
		owner_id, owner_class_library, owner_class_name,
		attachment_type_id, attachment_type, is_external, legacy_file_id,
		current_version, created_by, created_at, updated_at, deleted_at
	FROM documents
	WHERE owner_id = $1 AND legacy_file_id = $2 AND deleted_at IS NULL`

	var doc domain.Document
	err := r.pool.QueryRow(ctx, query, ownerID, fileID).Scan(
		&doc.ID, &doc.MinioBucket, &doc.MinioKey, &doc.Filename, &doc.ContentType, &doc.FileSize, &doc.SHA256Hash,
		&doc.OwnerID, &doc.OwnerClassLibrary, &doc.OwnerClassName,
		&doc.AttachmentTypeID, &doc.AttachmentType, &doc.IsExternal, &doc.LegacyFileID,
		&doc.CurrentVersion, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("document repo: get by legacy id: %w", err)
	}
	return &doc, nil
}
