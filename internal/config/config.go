package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	SQLServer      SQLServerConfig
	PostgreSQL     PostgreSQLConfig
	MinIO          MinIOConfig
	Sync           SyncConfig
	DLQRetry       DLQRetryConfig
	Reconciliation ReconciliationConfig
	API            APIConfig
	Log            LogConfig
	NATS           NATSConfig
	OIDC           OIDCConfig
	RateLimit      RateLimitConfig
}

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

// OIDCConfig holds OpenID Connect provider configuration.
type OIDCConfig struct {
	Enabled   bool
	IssuerURL string
	ClientID  string
}

type NATSConfig struct {
	URL        string
	StreamName string
	Subjects   []string
}

type APIConfig struct {
	Port               int
	JWTPublicKeyPEM    string
	CORSAllowedOrigins string // comma-separated
}

type DLQRetryConfig struct {
	Interval  time.Duration
	BatchSize int
}

type ReconciliationConfig struct {
	Schedule time.Duration
	Enabled  bool
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
	Host              string
	Port              int
	Database          string
	User              string
	Password          string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func (c PostgreSQLConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

type MinIOConfig struct {
	Endpoint           string
	AccessKey          string
	SecretKey          string
	Bucket             string
	UseSSL             bool
	MultipartThreshold int64  // default 5MB
	PartSize           uint64 // default 16MB
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
			Host:              getEnv("POSTGRES_HOST", "localhost"),
			Port:              getEnvInt("POSTGRES_PORT", 5432),
			Database:          getEnv("POSTGRES_DATABASE", "ibs_doc_engine"),
			User:              getEnv("POSTGRES_USER", "postgres"),
			Password:          getEnv("POSTGRES_PASSWORD", ""),
			SSLMode:           getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns:          int32(getEnvInt("POSTGRES_MAX_CONNS", 20)),
			MinConns:          int32(getEnvInt("POSTGRES_MIN_CONNS", 5)),
			MaxConnLifetime:   getEnvDuration("POSTGRES_MAX_CONN_LIFETIME", 1*time.Hour),
			MaxConnIdleTime:   getEnvDuration("POSTGRES_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getEnvDuration("POSTGRES_HEALTH_CHECK_PERIOD", 1*time.Minute),
		},
		MinIO: MinIOConfig{
			Endpoint:           getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey:          getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey:          getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:             getEnv("MINIO_BUCKET", "ibs-documents"),
			UseSSL:             getEnvBool("MINIO_USE_SSL", false),
			MultipartThreshold: int64(getEnvInt("MINIO_MULTIPART_THRESHOLD", 5*1024*1024)),
			PartSize:           uint64(getEnvInt("MINIO_PART_SIZE", 16*1024*1024)),
		},
		Sync: SyncConfig{
			PollInterval: getEnvDuration("SYNC_POLL_INTERVAL", 30*time.Second),
			BatchSize:    getEnvInt("SYNC_BATCH_SIZE", 100),
			Workers:      getEnvInt("SYNC_WORKERS", 10),
		},
		DLQRetry: DLQRetryConfig{
			Interval:  getEnvDuration("DLQ_RETRY_INTERVAL", 5*time.Minute),
			BatchSize: getEnvInt("DLQ_RETRY_BATCH_SIZE", 20),
		},
		Reconciliation: ReconciliationConfig{
			Schedule: getEnvDuration("RECONCILIATION_SCHEDULE", 24*time.Hour),
			Enabled:  getEnvBool("RECONCILIATION_ENABLED", true),
		},
		API: APIConfig{
			Port:               getEnvInt("API_PORT", 8080),
			JWTPublicKeyPEM:    getEnv("JWT_PUBLIC_KEY_PEM", ""),
			CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		NATS: NATSConfig{
			URL:        getEnv("NATS_URL", "nats://localhost:4222"),
			StreamName: getEnv("NATS_STREAM", "IBSDOCS"),
			Subjects:   []string{"document.>", "sync.>"},
		},
		OIDC: OIDCConfig{
			Enabled:   getEnvBool("OIDC_ENABLED", false),
			IssuerURL: getEnv("OIDC_ISSUER_URL", ""),
			ClientID:  getEnv("OIDC_CLIENT_ID", ""),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getEnvFloat("RATE_LIMIT_RPS", 10.0),
			Burst:             getEnvInt("RATE_LIMIT_BURST", 20),
		},
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

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
