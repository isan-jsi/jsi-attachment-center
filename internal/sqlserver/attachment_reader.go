package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jsi/ibs-doc-engine/internal/domain"
)

type AttachmentReader struct {
	db *sql.DB
}

func NewAttachmentReader(db *sql.DB) *AttachmentReader {
	return &AttachmentReader{db: db}
}

// FetchAfterCheckpoint reads attachments with DCCheck > lastCheckpoint.
// Uses NOLOCK hint to avoid blocking the legacy application.
// Pass nil for lastCheckpoint to fetch from the beginning.
func (r *AttachmentReader) FetchAfterCheckpoint(ctx context.Context, lastCheckpoint []byte, batchSize int) ([]domain.LegacyAttachment, error) {
	if r.db == nil {
		return nil, fmt.Errorf("attachment reader: database connection is nil")
	}

	var query string
	var args []interface{}

	if lastCheckpoint == nil {
		query = `SELECT TOP(@batch)
			a.OwnerID, a.FileID, a.DocAttachmentTypeID, a.DocAttachmentType,
			a.FileName, a.ContentType, a.FileContent, a.FileSize,
			a.IsExternal, a.IsPostLockUpdate,
			a.CreatedBy, a.CreatedOn, a.LastUpdateBy, a.LastUpdateOn, a.DCCheck
		FROM IBSDocAttachments a WITH (NOLOCK)
		ORDER BY a.DCCheck ASC`
		args = []interface{}{sql.Named("batch", batchSize)}
	} else {
		query = `SELECT TOP(@batch)
			a.OwnerID, a.FileID, a.DocAttachmentTypeID, a.DocAttachmentType,
			a.FileName, a.ContentType, a.FileContent, a.FileSize,
			a.IsExternal, a.IsPostLockUpdate,
			a.CreatedBy, a.CreatedOn, a.LastUpdateBy, a.LastUpdateOn, a.DCCheck
		FROM IBSDocAttachments a WITH (NOLOCK)
		WHERE a.DCCheck > @checkpoint
		ORDER BY a.DCCheck ASC`
		args = []interface{}{sql.Named("batch", batchSize), sql.Named("checkpoint", lastCheckpoint)}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("attachment reader: query: %w", err)
	}
	defer rows.Close()

	var results []domain.LegacyAttachment
	for rows.Next() {
		var att domain.LegacyAttachment
		var contentType, createdBy, lastUpdateBy sql.NullString
		var fileContent []byte

		err := rows.Scan(
			&att.OwnerID, &att.FileID, &att.DocAttachmentTypeID, &att.DocAttachmentType,
			&att.FileName, &contentType, &fileContent, &att.FileSize,
			&att.IsExternal, &att.IsPostLockUpdate,
			&createdBy, &att.CreatedOn, &lastUpdateBy, &att.LastUpdateOn, &att.DCCheck,
		)
		if err != nil {
			return nil, fmt.Errorf("attachment reader: scan: %w", err)
		}

		att.ContentType = contentType.String
		att.CreatedBy = createdBy.String
		att.LastUpdateBy = lastUpdateBy.String
		att.FileContent = fileContent
		results = append(results, att)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("attachment reader: rows: %w", err)
	}

	return results, nil
}

// FetchOwnerForAttachment looks up the owner class info for a given attachment.
func (r *AttachmentReader) FetchOwnerForAttachment(ctx context.Context, attachmentTypeID int) (*domain.LegacyOwner, error) {
	if r.db == nil {
		return nil, fmt.Errorf("attachment reader: database connection is nil")
	}

	query := `SELECT o.OwnerClassLibrary, o.OwnerClassName, o.OwnerClassDescription
		FROM IBSDocAttachmentTypes t WITH (NOLOCK)
		JOIN IBSDocAttachmentOwners o WITH (NOLOCK)
			ON t.OwnerClassLibrary = o.OwnerClassLibrary
			AND t.OwnerClassName = o.OwnerClassName
		WHERE t.DocAttachmentTypeID = @typeID`

	var owner domain.LegacyOwner
	err := r.db.QueryRowContext(ctx, query, sql.Named("typeID", attachmentTypeID)).Scan(
		&owner.OwnerClassLibrary, &owner.OwnerClassName, &owner.OwnerClassDescription,
	)
	if err != nil {
		return nil, fmt.Errorf("attachment reader: fetch owner: %w", err)
	}
	return &owner, nil
}

// BuildMinioKey constructs a deterministic, URL-safe MinIO object key.
// Format: {ownerLib}/{ownerClass}/{ownerID}/{attachType}/{fileID}/{filename}
func BuildMinioKey(owner domain.LegacyOwner, att domain.LegacyAttachment) string {
	sanitize := func(s string) string {
		s = strings.ReplaceAll(s, "/", "_")
		s = strings.ReplaceAll(s, "\\", "_")
		s = strings.ReplaceAll(s, " ", "_")
		s = strings.ReplaceAll(s, "\x00", "")
		return s
	}

	return fmt.Sprintf("%s/%s/%s/%s/%s/%s",
		sanitize(owner.OwnerClassLibrary),
		sanitize(owner.OwnerClassName),
		sanitize(att.OwnerID),
		sanitize(att.DocAttachmentType),
		sanitize(att.FileID),
		sanitize(att.FileName),
	)
}
