.PHONY: build test lint run-sync run-api migrate-up migrate-down docker-up docker-down

# Build
build:
	go build -o bin/sync-engine ./cmd/sync-engine
	go build -o bin/api-gateway ./cmd/api-gateway

build-sync:
	go build -o bin/sync-engine ./cmd/sync-engine

build-api:
	go build -o bin/api-gateway ./cmd/api-gateway

# Test
test:
	go test ./... -v -race -count=1

test-unit:
	go test ./internal/... -v -race -count=1 -short

test-integration:
	go test ./internal/... -v -race -count=1 -run Integration

# Lint
lint:
	golangci-lint run ./...

# Run
run-sync:
	go run ./cmd/sync-engine

run-api:
	go run ./cmd/api-gateway

# Migrations
migrate-up:
	migrate -path migrations -database "$${POSTGRES_URL}" up

migrate-down:
	migrate -path migrations -database "$${POSTGRES_URL}" down 1

# Docker
docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down
