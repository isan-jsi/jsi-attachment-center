# IBS Document Engine

Document attachment sync engine and API gateway for the IBS ecosystem. Migrates attachments from SQL Server `varbinary(MAX)` to MinIO object storage, exposes them via REST API, and provides a web-based virtual folder UI.

## Architecture

| Component | Tech | Port |
|-----------|------|------|
| **Sync Engine** | Go 1.22 | 8081 (health + metrics) |
| **API Gateway** | Go 1.22 + Chi v5 | 8080 |
| **Web UI** | React 18 + Vite + Tailwind | 3000 (dev) / 3001 (Docker) |
| **PostgreSQL** | v16 | 5432 |
| **MinIO** | S3-compatible storage | 9000 (API) / 9001 (console) |
| **NATS** | JetStream event bus | 4222 (client) / 8222 (monitor) |
| **SQL Server** | 2019 (legacy source) | 1433 |

---

## Prerequisites

### Required for all setups

- [Git](https://git-scm.com/)

### Without Docker

- [Go 1.22.3+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) and npm
- [PostgreSQL 16](https://www.postgresql.org/download/)
- [MinIO](https://min.io/download) (or any S3-compatible storage)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for database migrations)
- (Optional) [NATS Server](https://nats.io/download/) with JetStream
- (Optional) SQL Server 2019+ (legacy source — only needed for real sync)

### With Docker

- [Docker](https://www.docker.com/) and Docker Compose v2

---

## Quick Start with Docker

### 1. Clone and configure

```bash
git clone https://github.com/isan-jsi/jsi-attachment-center.git
cd jsi-attachment-center
cp .env.example .env
```

### 2. Start all infrastructure services

```bash
docker compose -f deployments/docker-compose.yml up -d
```

This starts: PostgreSQL, MinIO, NATS, SQL Server, Prometheus, and Grafana.

Wait for all services to be healthy:

```bash
docker compose -f deployments/docker-compose.yml ps
```

### 3. Run database migrations

```bash
export POSTGRES_URL="postgres://postgres:postgres@localhost:5432/ibs_doc_engine?sslmode=disable"
make migrate-up
```

### 4. Build and run the Go services

**Option A — Run natively (recommended for development):**

```bash
# Terminal 1: API Gateway
make run-api

# Terminal 2: Sync Engine (requires SQL Server with IBS data)
make run-sync
```

**Option B — Build Docker images and run everything:**

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

This also builds and starts `api-gateway` (port 8080) and `web` (port 3001).

### 5. Start the Web UI (dev mode)

```bash
cd web
npm install
npm run dev
```

Web UI is available at http://localhost:3000. The Vite dev server proxies `/api/*` to the API Gateway at `localhost:8080`.

### 6. Verify

```bash
# API Gateway health
curl http://localhost:8080/health

# Sync Engine health (if running)
curl http://localhost:8081/health

# Swagger UI
open http://localhost:8080/api/docs
```

---

## Quick Start without Docker

### 1. Install and start PostgreSQL

```bash
# macOS
brew install postgresql@16 && brew services start postgresql@16

# Ubuntu/Debian
sudo apt install postgresql-16 && sudo systemctl start postgresql

# Windows — use the installer from https://www.postgresql.org/download/windows/
```

Create the database:

```bash
psql -U postgres -c "CREATE DATABASE ibs_doc_engine;"
```

### 2. Install and start MinIO

```bash
# macOS
brew install minio/stable/minio
minio server ~/minio-data --console-address ":9001"

# Windows — download from https://min.io/download
# Then run:
minio.exe server C:\minio-data --console-address ":9001"
```

MinIO Console: http://localhost:9001 (login: `minioadmin` / `minioadmin`)

### 3. (Optional) Install and start NATS

```bash
# macOS
brew install nats-server
nats-server --jetstream --store_dir /tmp/nats-data

# Windows — download from https://nats.io/download/
nats-server.exe --jetstream --store_dir C:\nats-data
```

> If NATS is not available, the API Gateway falls back to a no-op event bus automatically.

### 4. Clone and configure

```bash
git clone https://github.com/isan-jsi/jsi-attachment-center.git
cd jsi-attachment-center
cp .env.example .env
```

Edit `.env` with your local settings if they differ from the defaults.

### 5. Run database migrations

```bash
# Install golang-migrate if you don't have it
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
export POSTGRES_URL="postgres://postgres:postgres@localhost:5432/ibs_doc_engine?sslmode=disable"
make migrate-up
```

### 6. Build and run

```bash
# Build both services
make build

# Terminal 1: API Gateway
./bin/api-gateway

# Terminal 2: Sync Engine (requires SQL Server with IBS data)
./bin/sync-engine
```

Or run directly without building:

```bash
make run-api    # Terminal 1
make run-sync   # Terminal 2
```

### 7. Start the Web UI

```bash
cd web
npm install
npm run dev
```

Web UI: http://localhost:3000

### 8. Verify

```bash
curl http://localhost:8080/health
# {"status":"ok","service":"api-gateway","timestamp":"..."}

curl http://localhost:8081/health
# {"status":"ok","service":"sync-engine","timestamp":"..."}
```

---

## Manual API Testing

### Authentication

All `/api/v1/*` endpoints require authentication. You have three options:

#### Option 1: API Key (simplest for testing)

First, insert a test API key directly into PostgreSQL:

```sql
-- Connect to PostgreSQL
psql -U postgres -d ibs_doc_engine

-- Insert a test API key (raw key: "test-api-key-12345")
INSERT INTO api_keys (name, key_hash, permissions, created_at, expires_at)
VALUES (
  'test-key',
  encode(sha256('test-api-key-12345'::bytea), 'hex'),
  '["documents:read","documents:write","folders:read","folders:write","search:read"]',
  NOW(),
  NOW() + INTERVAL '1 year'
);
```

Then use the key in requests:

```bash
curl -H "X-API-Key: test-api-key-12345" http://localhost:8080/api/v1/documents
```

#### Option 2: JWT Bearer Token

Set `JWT_PUBLIC_KEY_PEM` env var to the path of your RSA public key PEM file. Then pass a signed JWT:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:8080/api/v1/documents
```

#### Option 3: OIDC

Set `OIDC_ENABLED=true`, `OIDC_ISSUER_URL`, and `OIDC_CLIENT_ID` in your `.env`, then use a token from your OIDC provider.

### Example API Calls

```bash
API_KEY="test-api-key-12345"
BASE="http://localhost:8080/api/v1"

# List documents
curl -s -H "X-API-Key: $API_KEY" "$BASE/documents" | jq

# Get a specific document
curl -s -H "X-API-Key: $API_KEY" "$BASE/documents/{id}" | jq

# Download a document
curl -H "X-API-Key: $API_KEY" "$BASE/documents/{id}/download" -o output.pdf

# List folders
curl -s -H "X-API-Key: $API_KEY" "$BASE/folders" | jq

# Search (ILIKE)
curl -s -H "X-API-Key: $API_KEY" "$BASE/search?q=invoice" | jq

# Full-text search
curl -s -H "X-API-Key: $API_KEY" "$BASE/search?q=invoice&fts=true" | jq

# Search suggestions
curl -s -H "X-API-Key: $API_KEY" "$BASE/search/suggest?q=inv" | jq

# List owner classes
curl -s -H "X-API-Key: $API_KEY" "$BASE/owners" | jq

# Sync status
curl -s -H "X-API-Key: $API_KEY" "$BASE/sync/status" | jq

# Create an API key (via API)
curl -s -X POST -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-app-key","permissions":["documents:read"]}' \
  "$BASE/api-keys" | jq

# List API keys
curl -s -H "X-API-Key: $API_KEY" "$BASE/api-keys" | jq
```

### Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/api/docs
```

Raw OpenAPI spec:

```
http://localhost:8080/api/docs/openapi.yaml
```

---

## Running Tests

```bash
# All tests (with race detector)
make test

# Unit tests only
make test-unit

# TypeScript type check
cd web && npx tsc --noEmit

# Go linting (requires golangci-lint)
make lint
```

---

## Load Testing

k6 load test scripts are in `scripts/k6/`:

```bash
# Install k6: https://k6.io/docs/get-started/installation/

# API load test (100 VUs)
k6 run scripts/k6/load_test_api.js

# Search load test (50 VUs)
k6 run scripts/k6/load_test_search.js

# Upload load test (10 VUs)
k6 run scripts/k6/load_test_upload.js
```

---

## Monitoring

When running with Docker Compose:

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | — |
| MinIO Console | http://localhost:9001 | minioadmin / minioadmin |
| NATS Monitor | http://localhost:8222 | — |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_DATABASE` | `ibs_doc_engine` | Database name |
| `POSTGRES_USER` | `postgres` | Database user |
| `POSTGRES_PASSWORD` | `postgres` | Database password |
| `POSTGRES_SSLMODE` | `disable` | SSL mode |
| `POSTGRES_MAX_CONNS` | `20` | Max pool connections |
| `POSTGRES_MIN_CONNS` | `5` | Min pool connections |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO access key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO secret key |
| `MINIO_BUCKET` | `ibs-documents` | Storage bucket name |
| `MINIO_USE_SSL` | `false` | Use TLS for MinIO |
| `SQLSERVER_HOST` | `localhost` | Legacy SQL Server host |
| `SQLSERVER_PORT` | `1433` | Legacy SQL Server port |
| `SQLSERVER_DATABASE` | `IBS` | Legacy database name |
| `SQLSERVER_USER` | `sa` | Legacy database user |
| `SQLSERVER_PASSWORD` | — | Legacy database password |
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `API_PORT` | `8080` | API Gateway listen port |
| `JWT_PUBLIC_KEY_PEM` | — | Path to RSA public key PEM |
| `OIDC_ENABLED` | `false` | Enable OIDC authentication |
| `OIDC_ISSUER_URL` | — | OIDC provider URL |
| `OIDC_CLIENT_ID` | — | OIDC audience / client ID |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed origins |
| `RATE_LIMIT_RPS` | `10` | Rate limit (requests/sec) |
| `RATE_LIMIT_BURST` | `20` | Rate limit burst size |
| `SYNC_POLL_INTERVAL` | `30s` | Sync polling interval |
| `SYNC_BATCH_SIZE` | `100` | Documents per sync batch |
| `SYNC_WORKERS` | `10` | Concurrent sync workers |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

---

## Project Structure

```
ibs-doc-engine/
├── cmd/
│   ├── api-gateway/        # REST API server entrypoint
│   └── sync-engine/        # Sync engine entrypoint
├── internal/
│   ├── api/                # HTTP router, Swagger UI
│   │   ├── handlers/       # Request handlers (documents, folders, search, etc.)
│   │   └── middleware/     # Auth (JWT/OIDC/API key), RBAC, rate limiting
│   ├── config/             # Environment-based configuration
│   ├── domain/             # Domain models (document, folder, API key, sync)
│   ├── events/             # NATS JetStream event bus
│   ├── metrics/            # Prometheus metrics collectors
│   ├── minio/              # MinIO client (upload, SHA-256, multipart)
│   ├── postgres/           # PostgreSQL repositories
│   ├── sqlserver/          # Legacy SQL Server attachment reader
│   └── sync/               # Sync pipeline, transformers, worker pool
├── migrations/             # PostgreSQL migration files (001-009)
├── api/                    # OpenAPI 3.0 spec (embedded)
├── web/                    # React frontend (Vite + Tailwind + shadcn/ui)
├── deployments/
│   ├── docker-compose.yml  # Full dev environment
│   └── docker/             # Dockerfiles for api-gateway and sync-engine
├── scripts/k6/             # k6 load test scripts
├── docs/                   # Onboarding guide, specs, plans
├── Makefile                # Build, test, run, migrate commands
└── .env.example            # Environment variable template
```
