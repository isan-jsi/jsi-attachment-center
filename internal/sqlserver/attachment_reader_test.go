package sqlserver_test

import (
	"context"
	"testing"

	"github.com/jsi/ibs-doc-engine/internal/domain"
	"github.com/jsi/ibs-doc-engine/internal/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachmentReader_FetchAfterCheckpoint(t *testing.T) {
	reader := sqlserver.NewAttachmentReader(nil) // nil db = should fail gracefully
	_, err := reader.FetchAfterCheckpoint(context.Background(), nil, 10)
	require.Error(t, err)
}

func TestAttachmentReader_BuildMinioKey(t *testing.T) {
	tests := []struct {
		name     string
		ownerLib string
		ownerCls string
		ownerID  string
		attType  string
		fileID   string
		filename string
		want     string
	}{
		{
			name: "standard path",
			ownerLib: "Procurement", ownerCls: "PurchaseOrder",
			ownerID: "PO-2024-001", attType: "SupportingDocuments",
			fileID: "FILE001", filename: "invoice.pdf",
			want: "Procurement/PurchaseOrder/PO-2024-001/SupportingDocuments/FILE001/invoice.pdf",
		},
		{
			name: "special characters in ownerID",
			ownerLib: "Finance", ownerCls: "Invoice",
			ownerID: "INV/2024/001", attType: "Receipts",
			fileID: "FILE002", filename: "receipt scan.pdf",
			want: "Finance/Invoice/INV_2024_001/Receipts/FILE002/receipt_scan.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			att := domain.LegacyAttachment{
				OwnerID:           tt.ownerID,
				FileID:            tt.fileID,
				DocAttachmentType: tt.attType,
				FileName:          tt.filename,
			}
			owner := domain.LegacyOwner{
				OwnerClassLibrary: tt.ownerLib,
				OwnerClassName:    tt.ownerCls,
			}
			got := sqlserver.BuildMinioKey(owner, att)
			assert.Equal(t, tt.want, got)
		})
	}
}
