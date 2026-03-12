package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jsi/ibs-doc-engine/internal/config"
)

func NewPool(ctx context.Context, cfg config.PostgreSQLConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("postgres: new pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
}
