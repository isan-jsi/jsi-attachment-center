# IBS Document Engine — External App Integration Guide

This guide explains how to integrate an external application with the IBS Document Sync Engine API.

---

## 1. Obtaining an API Key

Send a `POST` request to `/api/v1/api-keys`. You must be authenticated (Bearer token) to create keys.

```http
POST /api/v1/api-keys
Authorization: Bearer <your-jwt-token>
Content-Type: application/json

{
  "name": "my-service",
  "permissions": ["documents:read", "documents:write"],
  "expires_at": "2027-01-01T00:00:00Z"
}
```

**Response:**

```json
{
  "data": {
    "id": "01HX...",
    "name": "my-service",
    "key": "ibskey_xxxxxxxxxxxxxxxxxxxx",
    "permissions": ["documents:read", "documents:write"],
    "expires_at": "2027-01-01T00:00:00Z",
    "created_at": "2026-03-12T09:00:00Z"
  }
}
```

> **Important:** The `key` value is only returned once. Store it securely — it cannot be retrieved again.

---

## 2. Authenticating Requests

The API supports two authentication methods.

### X-API-Key Header (recommended for service-to-service)

```http
GET /api/v1/documents
X-API-Key: ibskey_xxxxxxxxxxxxxxxxxxxx
```

### Bearer Token (JWT or OIDC)

```http
GET /api/v1/documents
Authorization: Bearer <token>
```

Bearer tokens can be:
- **RS256-signed JWTs** issued by your identity provider and configured via `JWT_PUBLIC_KEY_PEM`.
- **OIDC tokens** when OIDC is enabled (`OIDC_ENABLED=true`). The token is verified against the configured issuer.

---

## 3. Subscribing to Document Events (NATS JetStream)

The engine publishes domain events to a NATS JetStream stream. Consumers can subscribe using the standard NATS JetStream API.

### Go Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/nats-io/nats.go"
    "github.com/nats-io/nats.go/jetstream"
)

func main() {
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Drain()

    js, err := jetstream.New(nc)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a durable consumer on the IBSDOCS stream
    consumer, err := js.CreateOrUpdateConsumer(ctx, "IBSDOCS", jetstream.ConsumerConfig{
        Name:          "my-service-consumer",
        Durable:       "my-service-consumer",
        FilterSubject: "document.>",
        AckPolicy:     jetstream.AckExplicitPolicy,
    })
    if err != nil {
        log.Fatal(err)
    }

    msgs, err := consumer.Messages()
    if err != nil {
        log.Fatal(err)
    }

    for msg := range msgs {
        var event map[string]interface{}
        if err := json.Unmarshal(msg.Data(), &event); err != nil {
            log.Printf("unmarshal error: %v", err)
            msg.Nak()
            continue
        }
        fmt.Printf("received event: %s → %v\n", msg.Subject(), event)
        msg.Ack()
    }
}
```

---

## 4. Available Event Subjects

| Subject                  | Triggered When                          |
|--------------------------|-----------------------------------------|
| `document.created`       | A new document record is created        |
| `document.updated`       | Document metadata or content is updated |
| `document.deleted`       | A document is deleted                   |
| `document.synced`        | A document is successfully synced       |
| `document.sync_failed`   | A document sync attempt has failed      |
| `sync.started`           | A sync cycle has begun                  |
| `sync.completed`         | A sync cycle has finished               |

---

## 5. Event Payload Example

All events share a common envelope:

```json
{
  "id": "01HX9ABCDE1234567890ABCDEF",
  "subject": "document.synced",
  "payload": {
    "document_id": "01HX...",
    "file_name": "invoice-2026-03.pdf",
    "owner_id": "dept-finance",
    "size_bytes": 204800,
    "synced_at": "2026-03-12T09:15:00Z"
  },
  "occurred_at": "2026-03-12T09:15:00Z"
}
```

Fields:

| Field         | Type      | Description                                     |
|---------------|-----------|-------------------------------------------------|
| `id`          | string    | ULID — unique event identifier                  |
| `subject`     | string    | NATS subject this event was published on        |
| `payload`     | object    | Event-specific data (varies by subject)         |
| `occurred_at` | timestamp | RFC3339 UTC timestamp of when the event occurred|

---

## 6. Rate Limits

All API endpoints are subject to per-client rate limiting.

| Parameter                | Default | Environment Variable  |
|--------------------------|---------|-----------------------|
| Requests per second      | 10      | `RATE_LIMIT_RPS`      |
| Burst size               | 20      | `RATE_LIMIT_BURST`    |

When a request exceeds the limit, the server returns:

```
HTTP 429 Too Many Requests
Retry-After: 1
Content-Type: text/plain

{"error":"rate limit exceeded"}
```

### Recommended Backoff Strategy

Use **exponential backoff with jitter** when you receive a 429:

```go
func retryWithBackoff(maxAttempts int, call func() error) error {
    for attempt := 0; attempt < maxAttempts; attempt++ {
        if err := call(); err == nil {
            return nil
        }
        // Exponential backoff: 100ms, 200ms, 400ms, 800ms ...
        wait := time.Duration(100<<attempt) * time.Millisecond
        // Add jitter ±25%
        jitter := time.Duration(rand.Int63n(int64(wait / 2)))
        time.Sleep(wait + jitter)
    }
    return fmt.Errorf("max retries exceeded")
}
```

The rate-limit key is determined in priority order:
1. `X-API-Key` header value
2. `X-Forwarded-For` header (first IP)
3. TCP `RemoteAddr`

---

## 7. API Documentation Links

| Resource         | URL                              |
|------------------|----------------------------------|
| Swagger UI       | `http://<host>:<port>/api/docs`  |
| OpenAPI 3.0 spec | `http://<host>:<port>/api/docs/openapi.yaml` |

The Swagger UI lets you explore all endpoints interactively, including request/response schemas and authentication.

---

## 8. Key Management Operations

| Operation        | Method   | Path                          | Description                            |
|------------------|----------|-------------------------------|----------------------------------------|
| Create key       | `POST`   | `/api/v1/api-keys`            | Issue a new API key                    |
| List keys        | `GET`    | `/api/v1/api-keys`            | List all keys for the authenticated user |
| Get key          | `GET`    | `/api/v1/api-keys/{id}`       | Retrieve metadata for a specific key   |
| Revoke key       | `DELETE` | `/api/v1/api-keys/{id}`       | Permanently revoke a key               |
| Rotate key       | `POST`   | `/api/v1/api-keys/{id}/rotate`| Issue a replacement key and revoke old |

> Keys cannot be listed with their plaintext value after creation. If a key is lost, rotate or revoke it and issue a new one.

---

## 9. Environment Variables

The following environment variables configure the API gateway.

| Variable                | Default                    | Description                                      |
|-------------------------|----------------------------|--------------------------------------------------|
| `API_PORT`              | `8080`                     | Port the HTTP server listens on                  |
| `JWT_PUBLIC_KEY_PEM`    | _(empty)_                  | Path to or raw PEM of the RSA public key for JWT |
| `CORS_ALLOWED_ORIGINS`  | `*`                        | Comma-separated list of allowed CORS origins     |
| `RATE_LIMIT_RPS`        | `10.0`                     | Sustained requests per second per client         |
| `RATE_LIMIT_BURST`      | `20`                       | Maximum burst of requests per client             |
| `OIDC_ENABLED`          | `false`                    | Enable OIDC token verification                   |
| `OIDC_ISSUER_URL`       | _(empty)_                  | OIDC provider issuer URL                         |
| `OIDC_CLIENT_ID`        | _(empty)_                  | OIDC client ID for audience validation           |
| `NATS_URL`              | `nats://localhost:4222`    | NATS server URL                                  |
| `NATS_STREAM`           | `IBSDOCS`                  | JetStream stream name                            |
| `POSTGRES_HOST`         | `localhost`                | PostgreSQL host                                  |
| `POSTGRES_PORT`         | `5432`                     | PostgreSQL port                                  |
| `POSTGRES_DATABASE`     | `ibs_doc_engine`           | PostgreSQL database name                         |
| `POSTGRES_USER`         | `postgres`                 | PostgreSQL user                                  |
| `POSTGRES_PASSWORD`     | _(required)_               | PostgreSQL password                              |
| `POSTGRES_SSLMODE`      | `disable`                  | PostgreSQL SSL mode                              |
| `MINIO_ENDPOINT`        | `localhost:9000`           | MinIO endpoint                                   |
| `MINIO_ACCESS_KEY`      | `minioadmin`               | MinIO access key                                 |
| `MINIO_SECRET_KEY`      | `minioadmin`               | MinIO secret key                                 |
| `MINIO_BUCKET`          | `ibs-documents`            | MinIO bucket name                                |
| `LOG_LEVEL`             | `info`                     | Log level: `debug`, `info`, `warn`, `error`      |
