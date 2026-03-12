package sync

import (
	"context"

	"github.com/jsi/ibs-doc-engine/internal/domain"
	miniosvc "github.com/jsi/ibs-doc-engine/internal/minio"
	"github.com/jsi/ibs-doc-engine/internal/sqlserver"
)

// Transformer processes a SyncRecord during the pipeline.
type Transformer interface {
	Transform(ctx context.Context, record *domain.SyncRecord) (*domain.SyncRecord, error)
}

// ContentTypeDetector detects MIME type when ContentType is null/empty (SE-008).
type ContentTypeDetector struct{}

func NewContentTypeDetector() *ContentTypeDetector {
	return &ContentTypeDetector{}
}

func (t *ContentTypeDetector) Transform(ctx context.Context, record *domain.SyncRecord) (*domain.SyncRecord, error) {
	if record.LegacyAttachment.ContentType == "" {
		record.LegacyAttachment.ContentType = miniosvc.DetectContentType(
			record.LegacyAttachment.FileContent,
			record.LegacyAttachment.FileName,
		)
	}
	return record, nil
}

// PathGenerator constructs the deterministic MinIO key.
type PathGenerator struct{}

func NewPathGenerator() *PathGenerator {
	return &PathGenerator{}
}

func (t *PathGenerator) Transform(ctx context.Context, record *domain.SyncRecord) (*domain.SyncRecord, error) {
	owner := domain.LegacyOwner{
		OwnerClassLibrary: record.OwnerClassLibrary,
		OwnerClassName:    record.OwnerClassName,
	}
	record.MinioKey = sqlserver.BuildMinioKey(owner, record.LegacyAttachment)
	return record, nil
}

// HashComputer computes SHA-256 of the file content.
type HashComputer struct{}

func NewHashComputer() *HashComputer {
	return &HashComputer{}
}

func (t *HashComputer) Transform(ctx context.Context, record *domain.SyncRecord) (*domain.SyncRecord, error) {
	record.SHA256Hash = miniosvc.ComputeSHA256(record.LegacyAttachment.FileContent)
	return record, nil
}

// TransformerChain runs multiple transformers in sequence.
type TransformerChain struct {
	transformers []Transformer
}

func NewTransformerChain(transformers ...Transformer) *TransformerChain {
	return &TransformerChain{transformers: transformers}
}

func (c *TransformerChain) Transform(ctx context.Context, record *domain.SyncRecord) (*domain.SyncRecord, error) {
	var err error
	for _, t := range c.transformers {
		record, err = t.Transform(ctx, record)
		if err != nil {
			return nil, err
		}
	}
	return record, nil
}
