APP_NAME=payments-service

DB_HOST=localhost
DB_PORT=5432
DB_NAME=payments_service
DB_USER=payments_service
DB_PASSWORD=payments_service
DB_SSLMODE=disable
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: run test fmt build
.PHONY: db-up db-down db-logs db-psql db-schema perf-baseline

run:
	go run ./cmd/api

test:
	go test ./...

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

build:
	go build -o bin/$(APP_NAME) ./cmd/api

db-up:
	docker compose -f deployments/docker-compose.yml up -d

db-down:
	docker compose -f deployments/docker-compose.yml down

db-logs:
	docker compose -f deployments/docker-compose.yml logs -f postgres

db-psql:
	psql "$(DATABASE_URL)"

db-schema:
	psql "$(DATABASE_URL)" -f db/schema.sql

perf-baseline:
	k6 run k6/baseline.js
