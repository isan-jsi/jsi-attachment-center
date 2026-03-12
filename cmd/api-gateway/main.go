package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jsi/ibs-doc-engine/internal/api"
	"github.com/jsi/ibs-doc-engine/internal/api/handlers"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
	"github.com/jsi/ibs-doc-engine/internal/config"
	"github.com/jsi/ibs-doc-engine/internal/events"
	minioclient "github.com/jsi/ibs-doc-engine/internal/minio"
	"github.com/jsi/ibs-doc-engine/internal/postgres"
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

	slog.Info("api-gateway starting")

	// Connect to PostgreSQL
	pool, err := postgres.NewPool(ctx, cfg.PostgreSQL)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	defer pool.Close()
	slog.Info("connected to PostgreSQL")

	// Connect to MinIO
	mc, err := minioclient.NewClient(cfg.MinIO)
	if err != nil {
		return fmt.Errorf("minio connect: %w", err)
	}
	slog.Info("connected to MinIO")

	// Wire NATS event bus with graceful fallback to NoopEventBus.
	var bus events.EventBus
	natsbus, natsErr := events.NewNATSEventBus(cfg.NATS.URL, cfg.NATS.StreamName, cfg.NATS.Subjects)
	if natsErr != nil {
		slog.Warn("NATS unavailable, using noop event bus", "error", natsErr)
		bus = events.NewNoopEventBus()
	} else {
		bus = natsbus
		defer bus.Close()
	}
	_ = bus // bus is available for handlers that accept an EventBus

	// Wire OIDC verifier if enabled.
	var oidcVerifier *mw.OIDCVerifier
	if cfg.OIDC.Enabled && cfg.OIDC.IssuerURL != "" {
		oidcVerifier, err = mw.NewOIDCVerifier(ctx, cfg.OIDC.IssuerURL, cfg.OIDC.ClientID)
		if err != nil {
			slog.Warn("OIDC initialization failed", "error", err)
		}
	}

	// Repos
	docRepo := postgres.NewDocumentRepo(pool)
	folderRepo := postgres.NewFolderRepo(pool)
	syncRepo := postgres.NewSyncRepo(pool)
	apiKeyRepo := postgres.NewAPIKeyRepo(pool)
	searchRepo := postgres.NewSearchRepo(pool)

	// Handlers
	docHandler := handlers.NewDocumentHandler(docRepo, mc)
	folderHandler := handlers.NewFolderHandler(folderRepo)
	searchHandler := handlers.NewSearchHandler(docRepo, searchRepo)
	ownerHandler := handlers.NewOwnerHandler(docRepo)
	syncHandler := handlers.NewSyncHandler(syncRepo)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyRepo)

	// Router
	var corsOrigins []string
	if cfg.API.CORSAllowedOrigins != "" {
		corsOrigins = strings.Split(cfg.API.CORSAllowedOrigins, ",")
	}
	r := api.NewRouter(api.RouterConfig{
		CORSAllowedOrigins: corsOrigins,
	})

	// Rate limiter — applied globally before auth.
	rl := mw.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	r.Use(rl.Middleware())

	// Auth middleware
	authMiddleware := mw.Auth(mw.AuthConfig{
		JWTPublicKeyPEM: cfg.API.JWTPublicKeyPEM,
		APIKeyRepo:      apiKeyRepo,
		OIDCVerifier:    oidcVerifier,
	})

	// Public API documentation routes (no auth required)
	r.Get("/api/docs", api.SwaggerUIHTML("/api/docs/openapi.yaml"))
	r.Get("/api/docs/openapi.yaml", api.ServeOpenAPISpec())

	// Mount API v1 routes
	r.Route("/api/v1", func(sub chi.Router) {
		sub.Use(authMiddleware)

		sub.Mount("/documents", docHandler.Routes())
		sub.Mount("/folders", folderHandler.Routes())
		sub.Mount("/search", searchHandler.Routes())
		sub.Mount("/owners", ownerHandler.Routes())
		sub.Mount("/sync", syncHandler.Routes())
		sub.Mount("/api-keys", apiKeyHandler.Routes())
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.API.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("api-gateway listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("api-gateway shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	slog.Info("api-gateway stopped")
	return nil
}
