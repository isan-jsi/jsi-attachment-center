package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	SQLServer  SQLServerConfig
	PostgreSQL PostgreSQLConfig
	MinIO      MinIOConfig
	Sync       SyncConfig
	Log        LogConfig
}

type SQLServerConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func (c SQLServerConfig) ConnectionString() string {
	return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=true&TrustServerCertificate=true",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

type PostgreSQLConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

func (c PostgreSQLConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type SyncConfig struct {
	PollInterval time.Duration
	BatchSize    int
	Workers      int
}

type LogConfig struct {
	Level string // debug, info, warn, error
}

func Load() (*Config, error) {
	cfg := &Config{
		SQLServer: SQLServerConfig{
			Host:     getEnv("SQLSERVER_HOST", "localhost"),
			Port:     getEnvInt("SQLSERVER_PORT", 1433),
			Database: getEnv("SQLSERVER_DATABASE", "IBS"),
			User:     getEnv("SQLSERVER_USER", "sa"),
			Password: getEnv("SQLSERVER_PASSWORD", ""),
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			Database: getEnv("POSTGRES_DATABASE", "ibs_doc_engine"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "ibs-documents"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
		},
		Sync: SyncConfig{
			PollInterval: getEnvDuration("SYNC_POLL_INTERVAL", 30*time.Second),
			BatchSize:    getEnvInt("SYNC_BATCH_SIZE", 100),
			Workers:      getEnvInt("SYNC_WORKERS", 10),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	if cfg.SQLServer.Password == "" {
		return nil, fmt.Errorf("SQLSERVER_PASSWORD is required")
	}
	if cfg.PostgreSQL.Password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
