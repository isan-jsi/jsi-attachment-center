package sync_test

import (
	"context"
	"testing"

	"github.com/jsi/ibs-doc-engine/internal/domain"
	syncsvc "github.com/jsi/ibs-doc-engine/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeDetector_NullContentType(t *testing.T) {
	detector := syncsvc.NewContentTypeDetector()
	record := &domain.SyncRecord{
		LegacyAttachment: domain.LegacyAttachment{
			ContentType: "",
			FileContent: []byte("%PDF-1.4 fake pdf content"),
			FileName:    "report.pdf",
		},
	}

	result, err := detector.Transform(context.Background(), record)
	require.NoError(t, err)
	assert.Equal(t, "application/pdf", result.LegacyAttachment.ContentType)
}

func TestContentTypeDetector_ExistingContentType(t *testing.T) {
	detector := syncsvc.NewContentTypeDetector()
	record := &domain.SyncRecord{
		LegacyAttachment: domain.LegacyAttachment{
			ContentType: "image/png",
			FileContent: []byte("doesn't matter"),
			FileName:    "photo.png",
		},
	}

	result, err := detector.Transform(context.Background(), record)
	require.NoError(t, err)
	assert.Equal(t, "image/png", result.LegacyAttachment.ContentType, "should not modify existing content type")
}

func TestPathGenerator(t *testing.T) {
	gen := syncsvc.NewPathGenerator()
	record := &domain.SyncRecord{
		OwnerClassLibrary: "Procurement",
		OwnerClassName:    "PurchaseOrder",
		LegacyAttachment: domain.LegacyAttachment{
			OwnerID:           "PO-2024-001",
			FileID:            "FILE001",
			DocAttachmentType: "SupportingDocuments",
			FileName:          "invoice.pdf",
		},
	}

	result, err := gen.Transform(context.Background(), record)
	require.NoError(t, err)
	assert.Equal(t, "Procurement/PurchaseOrder/PO-2024-001/SupportingDocuments/FILE001/invoice.pdf", result.MinioKey)
}
