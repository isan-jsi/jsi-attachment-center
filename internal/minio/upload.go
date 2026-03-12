package minio

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"

	miniogo "github.com/minio/minio-go/v7"
)

// UploadResult contains the result of a successful upload.
type UploadResult struct {
	Bucket     string
	Key        string
	SHA256Hash string
	Size       int64
}

// Upload streams file data to MinIO with metadata tags.
func (c *Client) Upload(ctx context.Context, key string, data []byte, contentType string, metadata map[string]string) (*UploadResult, error) {
	hash := ComputeSHA256(data)

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["sha256"] = hash

	reader := bytes.NewReader(data)
	info, err := c.mc.PutObject(ctx, c.bucket, key, reader, int64(len(data)), miniogo.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: upload %s: %w", key, err)
	}

	slog.Debug("minio: uploaded", "key", key, "size", info.Size, "hash", hash)

	return &UploadResult{
		Bucket:     c.bucket,
		Key:        key,
		SHA256Hash: hash,
		Size:       info.Size,
	}, nil
}

// ComputeSHA256 returns the hex-encoded SHA-256 hash of data.
func ComputeSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// DetectContentType detects MIME type from file content (magic bytes) with filename fallback.
func DetectContentType(data []byte, filename string) string {
	// Try magic byte detection first
	ct := http.DetectContentType(data)
	if ct != "application/octet-stream" {
		return ct
	}

	// Fallback: detect from file extension
	ext := filepath.Ext(filename)
	if ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}

	return ct
}
