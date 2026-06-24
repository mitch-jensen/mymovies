set dotenv-load := true
set shell := ["bash", "-uc"]

app := "mymovies"
golangci_image := "golangci/golangci-lint:v2.12.1"
sqlc_image := "sqlc/sqlc:1.31.1"
goose := env_var_or_default("GOOSE_BIN", "goose")
postgres_address := env_var_or_default("POSTGRES_ADDRESS", "127.0.0.1")
postgres_port := env_var_or_default("POSTGRES_PORT", "5432")
postgres_db := env_var("POSTGRES_DB")
postgres_user := env_var("POSTGRES_USER")
postgres_password := env_var("POSTGRES_PASSWORD")
database_url := "postgres://" + postgres_user + ":" + postgres_password + "@" + postgres_address + ":" + postgres_port + "/" + postgres_db + "?sslmode=disable"

# Show available recipes.
default:
  @just --list

# Format Go files.
fmt:
  gofmt -w $(find . -name '*.go' -not -path './vendor/*')

# Update Go module dependencies.
tidy:
  go mod tidy

# 
vendor:
  go mod vendor

# Run unit tests.
test:
  go test ./...

# Run tests with cross-package coverage and print a per-package + total summary.
cover:
  go test -coverpkg=./... -coverprofile=coverage.out ./...
  @echo "--- coverage by package ---"
  @go tool cover -func=coverage.out | tail -1

# Run golangci-lint in Docker.
lint:
  docker run --rm -t \
    -v "{{justfile_directory()}}:/app:z" \
    -w /app \
    --user "$(id -u):$(id -g)" \
    -v "$(go env GOCACHE):/tmp/go-build:z" \
    -e GOCACHE=/tmp/go-build \
    -v "$(go env GOMODCACHE):/tmp/mod:z" \
    -e GOMODCACHE=/tmp/mod \
    -v "$HOME/.cache/golangci-lint:/tmp/golangci-lint:z" \
    -e GOLANGCI_LINT_CACHE=/tmp/golangci-lint \
    {{golangci_image}} \
    golangci-lint run

# Generate dbstore code from sqlc.
sqlc:
  docker run --rm \
    -v "{{justfile_directory()}}:/src:z" \
    -w /src \
    {{sqlc_image}} generate

# Create a new goose migration: just migration add_movies
migration name:
  {{goose}} -dir migrations create {{name}} sql

# Apply all pending migrations.
migrate-up:
  {{goose}} -dir migrations postgres '{{database_url}}' up

# Roll back the latest migration.
migrate-down:
  {{goose}} -dir migrations postgres '{{database_url}}' down

# Print migration status.
migrate-status:
  {{goose}} -dir migrations postgres '{{database_url}}' status

# Start the Postgres container.
db-up:
  docker compose up -d db

# Stop compose services.
db-down:
  docker compose down

# Run the sample app.
run:
  go run ./main.go

# Export the OpenAPI spec to frontend/openapi.yaml (drives client generation).
openapi:
  go run ./cmd/openapi > frontend/openapi.yaml

# Start Postgres, apply migrations, and run the sample app.
dev: db-up migrate-up run

# Build the app binary.
build:
  go build main.go

# Run the normal local verification suite.
check: fmt sqlc test lint

# --- Docker lifecycle --------------------------------------------------------

# Build and start all services (backend, frontend, db, adminer) in the background.
up:
  docker compose up -d --build

# Stop and remove all compose services.
down:
  docker compose down

# Follow logs for all services (or one: `just logs frontend`).
logs *args:
  docker compose logs -f {{args}}

# --- Frontend (codegen/lint via pnpm; RUN the app via `just up`, never a dev server) ---

# Install frontend dependencies.
fe-install:
  pnpm --dir frontend install

# Regenerate the typed client from frontend/openapi.yaml (run `just openapi` first).
fe-gen:
  pnpm --dir frontend gen

# Refresh the OpenAPI spec from the backend and regenerate the typed client.
fe-sync: openapi fe-gen

# Build the frontend (regenerates routeTree.gen.ts via the router plugin, then tsc).
fe-build:
  pnpm --dir frontend build

# Lint the frontend with oxlint.
fe-lint:
  pnpm --dir frontend lint

# Format the frontend with Prettier.
fe-format:
  pnpm --dir frontend fmt

fe-typecheck:
  pnpm --dir frontend typecheck

# Run the frontend unit/component tests once (Vitest).
fe-test:
  pnpm --dir frontend test

# Frontend verification suite: lint + format + type check + tests.
fe-check:
  pnpm --dir frontend lint
  pnpm --dir frontend fmt:check
  pnpm --dir frontend typecheck
  pnpm --dir frontend test
