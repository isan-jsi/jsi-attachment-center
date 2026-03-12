package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jsi/ibs-doc-engine/internal/domain"
)

// SearchRepo provides full-text and suggestion search over documents.
type SearchRepo struct {
	pool *pgxpool.Pool
}

// NewSearchRepo creates a new SearchRepo.
func NewSearchRepo(pool *pgxpool.Pool) *SearchRepo {
	return &SearchRepo{pool: pool}
}

// FullTextSearchParams defines parameters for a full-text search query.
type FullTextSearchParams struct {
	Query       string
	OwnerID     string
	ContentType string
	Limit       int
	Offset      int
}

// FullTextSearchResult holds matched documents, total count, and per-document highlights.
type FullTextSearchResult struct {
	Documents  []domain.Document `json:"documents"`
	Total      int               `json:"total"`
	Highlights map[string]string `json:"highlights,omitempty"`
}

// FullTextSearch performs a ranked full-text search using the search_vector column.
func (r *SearchRepo) FullTextSearch(ctx context.Context, params FullTextSearchParams) (*FullTextSearchResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}

	// Build WHERE clause with websearch_to_tsquery for user-friendly input.
	baseWhere := "d.search_vector @@ websearch_to_tsquery('english', $1) AND d.deleted_at IS NULL"
	args := []interface{}{params.Query}
	argIdx := 2

	if params.OwnerID != "" {
		baseWhere += fmt.Sprintf(" AND d.owner_id = $%d", argIdx)
		args = append(args, params.OwnerID)
		argIdx++
	}
	if params.ContentType != "" {
		baseWhere += fmt.Sprintf(" AND d.content_type = $%d", argIdx)
		args = append(args, params.ContentType)
		argIdx++
	}

	// Count total matches.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM documents d WHERE %s", baseWhere)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("search repo: count: %w", err)
	}

	// Fetch results with ts_rank and ts_headline.
	selectQuery := fmt.Sprintf(`
		SELECT d.id, d.minio_bucket, d.minio_key, d.filename, d.content_type, d.file_size, d.sha256_hash,
		       d.owner_id, d.owner_class_library, d.owner_class_name,
		       d.attachment_type_id, d.attachment_type, d.is_external, d.legacy_file_id,
		       d.current_version, d.created_by, d.created_at, d.updated_at, d.deleted_at,
		       ts_rank(d.search_vector, websearch_to_tsquery('english', $1)) AS rank,
		       ts_headline('english', d.filename, websearch_to_tsquery('english', $1),
		           'StartSel=<mark>, StopSel=</mark>, MaxWords=35, MinWords=15') AS headline
		FROM documents d
		WHERE %s
		ORDER BY rank DESC
		LIMIT $%d OFFSET $%d`, baseWhere, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search repo: query: %w", err)
	}
	defer rows.Close()

	result := &FullTextSearchResult{Total: total, Highlights: make(map[string]string)}
	for rows.Next() {
		var doc domain.Document
		var rank float64
		var headline string
		if err := rows.Scan(
			&doc.ID, &doc.MinioBucket, &doc.MinioKey, &doc.Filename, &doc.ContentType, &doc.FileSize, &doc.SHA256Hash,
			&doc.OwnerID, &doc.OwnerClassLibrary, &doc.OwnerClassName,
			&doc.AttachmentTypeID, &doc.AttachmentType, &doc.IsExternal, &doc.LegacyFileID,
			&doc.CurrentVersion, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
			&rank, &headline,
		); err != nil {
			return nil, fmt.Errorf("search repo: scan: %w", err)
		}
		result.Documents = append(result.Documents, doc)
		result.Highlights[doc.ID.String()] = headline
	}
	return result, rows.Err()
}

// Suggest returns filename autocomplete suggestions matching the given prefix.
func (r *SearchRepo) Suggest(ctx context.Context, prefix string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT DISTINCT filename FROM documents
		WHERE filename ILIKE $1 || '%' AND deleted_at IS NULL
		ORDER BY filename LIMIT $2`

	rows, err := r.pool.Query(ctx, query, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("search repo: suggest: %w", err)
	}
	defer rows.Close()

	var suggestions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("search repo: suggest scan: %w", err)
		}
		suggestions = append(suggestions, name)
	}
	return suggestions, rows.Err()
}
