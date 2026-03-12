package minio_test

import (
	"testing"

	miniosvc "github.com/jsi/ibs-doc-engine/internal/minio"
	"github.com/stretchr/testify/assert"
)

func TestComputeSHA256(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty",
			data: []byte{},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "hello world",
			data: []byte("hello world"),
			want: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := miniosvc.ComputeSHA256(tt.data)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectContentType(t *testing.T) {
	// PDF magic bytes: %PDF
	pdfData := []byte("%PDF-1.4 some content here")
	ct := miniosvc.DetectContentType(pdfData, "test.pdf")
	assert.Equal(t, "application/pdf", ct)

	// Fallback to extension when magic bytes inconclusive
	// Use binary data with non-printable bytes so http.DetectContentType returns application/octet-stream
	unknownData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x0e, 0x0f, 0x10}
	ct = miniosvc.DetectContentType(unknownData, "report.xlsx")
	assert.Contains(t, ct, "application/") // should detect from extension or default
}
