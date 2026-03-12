package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsi/ibs-doc-engine/internal/api"
	"github.com/jsi/ibs-doc-engine/internal/config"
	miniosvc "github.com/jsi/ibs-doc-engine/internal/minio"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
	"github.com/jsi/ibs-doc-engine/internal/sqlserver"
	syncsvc "github.com/jsi/ibs-doc-engine/internal/sync"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Setup logger
	var logLevel slog.Level
	switch cfg.Log.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("sync-engine starting")

	// Connect to SQL Server (legacy, read-only)
	sqlDB, err := sqlserver.NewConnection(cfg.SQLServer)
	if err != nil {
		return fmt.Errorf("sqlserver connect: %w", err)
	}
	defer sqlDB.Close()
	slog.Info("connected to SQL Server")

	// Connect to PostgreSQL
	pgPool, err := postgres.NewPool(ctx, cfg.PostgreSQL)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	defer pgPool.Close()
	slog.Info("connected to PostgreSQL")

	// Connect to MinIO
	minioClient, err := miniosvc.NewClient(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("minio connect: %w", err)
	}
	if err := minioClient.EnsureBucket(ctx); err != nil {
		return fmt.Errorf("minio ensure bucket: %w", err)
	}
	slog.Info("connected to MinIO", "bucket", cfg.MinIO.Bucket)

	// Create repositories
	docRepo := postgres.NewDocumentRepo(pgPool)
	syncRepo := postgres.NewSyncRepo(pgPool)

	// Start health endpoint in background
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", api.HealthHandler("sync-engine"))
	healthServer := &http.Server{Addr: ":8081", Handler: healthMux}
	go func() {
		slog.Info("health endpoint listening", "addr", ":8081")
		if err := healthServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("health server error", "error", err)
		}
	}()
	defer healthServer.Shutdown(context.Background())

	// Create and start pipeline
	pipeline := syncsvc.NewPipeline(syncsvc.PipelineConfig{
		AttachmentReader: sqlserver.NewAttachmentReader(sqlDB),
		MinIOClient:      minioClient,
		DocumentRepo:     docRepo,
		SyncRepo:         syncRepo,
		TableName:        "IBSDocAttachments",
		BatchSize:        cfg.Sync.BatchSize,
		Workers:          cfg.Sync.Workers,
	})

	slog.Info("sync-engine running",
		"poll_interval", cfg.Sync.PollInterval,
		"workers", cfg.Sync.Workers,
		"batch_size", cfg.Sync.BatchSize,
	)

	pipeline.RunLoop(ctx, cfg.Sync.PollInterval)

	slog.Info("sync-engine stopped")
	return nil
}
