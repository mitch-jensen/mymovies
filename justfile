lint:
  docker run --rm -t \
    -v "$(pwd):/app:z" \
    -w /app \
    --user "$(id -u):$(id -g)" \
    -v "$(go env GOCACHE):/tmp/go-build:z" \
    -e GOCACHE=/tmp/go-build \
    -v "$(go env GOMODCACHE):/tmp/mod:z" \
    -e GOMODCACHE=/tmp/mod \
    -v "$HOME/.cache/golangci-lint:/tmp/golangci-lint:z" \
    -e GOLANGCI_LINT_CACHE=/tmp/golangci-lint \
    golangci/golangci-lint:v2.12.1 \
    golangci-lint run

sqlc-generate:
  docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate

migrate:
  ~/go/bin/goose up

run-app:
  go run main.go

run-db:
  docker compose up -d db

run:
  just run-db
  just run-app

build:
  go build -o mymovies main.go
