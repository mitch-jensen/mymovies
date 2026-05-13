package db_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	db "github.com/mitch-jensen/mymovies/dbstore"
)

func TestQueries_GetMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	conn := testDB(ctx, t)
	queries := db.New(conn)

	created, err := queries.CreateMovie(ctx, db.CreateMovieParams{
		Title:       "The Abominable Dr. Phibes",
		ReleaseYear: 1971,
		RuntimeMin:  pgtype.Int4{Int32: 0, Valid: false},
	})
	if err != nil {
		t.Fatalf("CreateMovie() error = %v", err)
	}

	got, err := queries.GetMovie(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetMovie() error = %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("GetMovie() ID = %v, want %v", got.ID, created.ID)
	}

	if got.Title != created.Title {
		t.Errorf("GetMovie() Title = %q, want %q", got.Title, created.Title)
	}

	if got.ReleaseYear != created.ReleaseYear {
		t.Errorf("GetMovie() ReleaseYear = %d, want %d", got.ReleaseYear, created.ReleaseYear)
	}

	if got.RuntimeMin != created.RuntimeMin {
		t.Errorf("GetMovie() RuntimeMin = %v, want %v", got.RuntimeMin, created.RuntimeMin)
	}
}

func testDB(ctx context.Context, t *testing.T) *pgx.Conn {
	t.Helper()

	const (
		dbName     = "mymovies_test"
		dbUser     = "postgres"
		dbPassword = "postgres"
	)

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	t.Cleanup(func() {
		err := testcontainers.TerminateContainer(container)
		if err != nil {
			t.Errorf("terminate postgres container: %v", err)
		}
	})

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("postgres connection string: %v", err)
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}

	t.Cleanup(func() {
		err := conn.Close(ctx)
		if err != nil {
			t.Errorf("close postgres connection: %v", err)
		}
	})

	applyMigrations(ctx, t, conn)

	return conn
}

func applyMigrations(ctx context.Context, t *testing.T, conn *pgx.Conn) {
	t.Helper()

	const migrationPath = "../migrations/20260227093145_initial.sql"

	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", migrationPath, err)
	}

	upSQL, ok := gooseUpSQL(string(migration))
	if !ok {
		t.Fatalf("migration %s does not contain a goose Up section", migrationPath)
	}

	_, err = conn.Exec(ctx, upSQL)
	if err != nil {
		t.Fatalf("apply migration %s: %v", migrationPath, err)
	}
}

func gooseUpSQL(migration string) (string, bool) {
	const (
		upMarker   = "-- +goose Up"
		downMarker = "-- +goose Down"
	)

	upStart := strings.Index(migration, upMarker)
	if upStart == -1 {
		return "", false
	}

	upStart += len(upMarker)

	upSQL := migration[upStart:]
	if downStart := strings.Index(upSQL, downMarker); downStart != -1 {
		upSQL = upSQL[:downStart]
	}

	upSQL = strings.TrimSpace(upSQL)

	return upSQL, upSQL != ""
}
