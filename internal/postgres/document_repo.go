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

// CountByOwnerClass returns document counts grouped by owner class.
func (r *DocumentRepo) CountByOwnerClass(ctx context.Context) (map[string]int64, error) {
	query := `SELECT owner_class_library || '::' || owner_class_name AS owner_class, COUNT(*)
		FROM documents WHERE deleted_at IS NULL
		GROUP BY owner_class_library, owner_class_name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("document repo: count by owner class: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var ownerClass string
		var count int64
		if err := rows.Scan(&ownerClass, &count); err != nil {
			return nil, fmt.Errorf("document repo: scan count: %w", err)
		}
		counts[ownerClass] = count
	}
	return counts, rows.Err()
}

// ListParams defines filters for listing documents.
type ListParams struct {
	OwnerClassLibrary string
	OwnerClassName    string
	OwnerID           string
	Page              int
	PerPage           int
}

func (p ListParams) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PerPage
}

func (r *DocumentRepo) List(ctx context.Context, params ListParams) ([]domain.Document, int64, error) {
	where := "deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if params.OwnerClassLibrary != "" {
		where += fmt.Sprintf(" AND owner_class_library = $%d", argIdx)
		args = append(args, params.OwnerClassLibrary)
		argIdx++
	}
	if params.OwnerClassName != "" {
		where += fmt.Sprintf(" AND owner_class_name = $%d", argIdx)
		args = append(args, params.OwnerClassName)
		argIdx++
	}
	if params.OwnerID != "" {
		where += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, params.OwnerID)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM documents WHERE " + where
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("document repo: count: %w", err)
	}

	if params.PerPage <= 0 {
		params.PerPage = 20
	}
	listQuery := fmt.Sprintf(`SELECT
		id, minio_bucket, minio_key, filename, content_type, file_size, sha256_hash,
		owner_id, owner_class_library, owner_class_name,
		attachment_type_id, attachment_type, is_external, legacy_file_id,
		current_version, created_by, created_at, updated_at, deleted_at
	FROM documents WHERE %s
	ORDER BY created_at DESC
	LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, params.PerPage, params.Offset())

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("document repo: list: %w", err)
	}
	defer rows.Close()

	var docs []domain.Document
	for rows.Next() {
		var doc domain.Document
		if err := rows.Scan(
			&doc.ID, &doc.MinioBucket, &doc.MinioKey, &doc.Filename, &doc.ContentType, &doc.FileSize, &doc.SHA256Hash,
			&doc.OwnerID, &doc.OwnerClassLibrary, &doc.OwnerClassName,
			&doc.AttachmentTypeID, &doc.AttachmentType, &doc.IsExternal, &doc.LegacyFileID,
			&doc.CurrentVersion, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("document repo: scan: %w", err)
		}
		docs = append(docs, doc)
	}
	return docs, total, rows.Err()
}

func (r *DocumentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Document, error) {
	query := `SELECT
		id, minio_bucket, minio_key, filename, content_type, file_size, sha256_hash,
		owner_id, owner_class_library, owner_class_name,
		attachment_type_id, attachment_type, is_external, legacy_file_id,
		current_version, created_by, created_at, updated_at, deleted_at
	FROM documents WHERE id = $1 AND deleted_at IS NULL`

	var doc domain.Document
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.MinioBucket, &doc.MinioKey, &doc.Filename, &doc.ContentType, &doc.FileSize, &doc.SHA256Hash,
		&doc.OwnerID, &doc.OwnerClassLibrary, &doc.OwnerClassName,
		&doc.AttachmentTypeID, &doc.AttachmentType, &doc.IsExternal, &doc.LegacyFileID,
		&doc.CurrentVersion, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("document repo: get by id: %w", err)
	}
	return &doc, nil
}

func (r *DocumentRepo) Update(ctx context.Context, doc *domain.Document) error {
	query := `UPDATE documents SET
		filename = $2, content_type = $3, owner_id = $4,
		owner_class_library = $5, owner_class_name = $6,
		attachment_type_id = $7, attachment_type = $8,
		updated_at = NOW()
	WHERE id = $1 AND deleted_at IS NULL`

	ct, err := r.pool.Exec(ctx, query,
		doc.ID, doc.Filename, doc.ContentType, doc.OwnerID,
		doc.OwnerClassLibrary, doc.OwnerClassName,
		doc.AttachmentTypeID, doc.AttachmentType,
	)
	if err != nil {
		return fmt.Errorf("document repo: update: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("document repo: update: not found")
	}
	return nil
}

func (r *DocumentRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE documents SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("document repo: soft delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("document repo: soft delete: not found")
	}
	return nil
}
