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

# Start Postgres, apply migrations, and run the sample app.
dev: db-up migrate-up run

# Build the app binary.
build:
  go build main.go

# Run the normal local verification suite.
check: fmt sqlc test lint
