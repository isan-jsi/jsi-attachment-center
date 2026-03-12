package sqlserver

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jsi/ibs-doc-engine/internal/config"
)

func NewConnection(cfg config.SQLServerConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("sqlserver: open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("sqlserver: ping: %w", err)
	}
	return db, nil
}
