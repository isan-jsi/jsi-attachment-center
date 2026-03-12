package minio

import (
	"context"
	"fmt"
	"log/slog"

	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/jsi/ibs-doc-engine/internal/config"
)

type Client struct {
	mc                 *miniogo.Client
	bucket             string
	multipartThreshold int64
	partSize           uint64
}

func NewClient(cfg config.MinIOConfig) (*Client, error) {
	mc, err := miniogo.New(cfg.Endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: new client: %w", err)
	}

	multipartThreshold := cfg.MultipartThreshold
	if multipartThreshold <= 0 {
		multipartThreshold = 5 * 1024 * 1024
	}
	partSize := cfg.PartSize
	if partSize <= 0 {
		partSize = 16 * 1024 * 1024
	}

	return &Client{
		mc:                 mc,
		bucket:             cfg.Bucket,
		multipartThreshold: multipartThreshold,
		partSize:           partSize,
	}, nil
}

// EnsureBucket creates the bucket if it doesn't exist.
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("minio: bucket exists check: %w", err)
	}
	if !exists {
		err = c.mc.MakeBucket(ctx, c.bucket, miniogo.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("minio: make bucket: %w", err)
		}
		slog.Info("minio: bucket created", "bucket", c.bucket)
	}
	return nil
}

func (c *Client) Bucket() string {
	return c.bucket
}

func (c *Client) Inner() *miniogo.Client {
	return c.mc
}
