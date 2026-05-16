package db_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var ( //nolint:gochecknoglobals // TestMain shares the container snapshot with integration tests.
	ctr   *postgres.PostgresContainer
	dbURL string
)

func RepoRoot() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("testhelper: could not determine caller path")
	}

	dir := filepath.Dir(file)
	for {
		_, err := os.Stat(filepath.Join(dir, "go.mod"))
		if err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			panic("testhelper: could not find go.mod walking up from " + file)
		}

		dir = parent
	}
}

// MigrationsDir returns the absolute path to the top-level migrations directory.
func MigrationsDir() string {
	return filepath.Join(RepoRoot(), "migrations")
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error

	ctr, err = postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("app_test"),
		postgres.WithUsername("app"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)

	if err != nil {
		log.Printf("error starting postgres container: %v", err)
		os.Exit(1)
	}

	dbURL, err = ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("error getting postgres connection string: %v", err)
		os.Exit(1)
	}

	pgxConfig, err := pgx.ParseConfig(dbURL)
	if err != nil {
		log.Printf("could not parse pgx config: %v", err)
		os.Exit(1)
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		log.Printf("set goose dialect: %v", err)
		os.Exit(1)
	}

	sqlDB := stdlib.OpenDB(*pgxConfig)

	err = goose.UpContext(ctx, sqlDB, MigrationsDir())
	if err != nil {
		log.Printf("apply migrations: %v", err)
		os.Exit(1)
	}

	// Close the database connection before taking the snapshot, otherwise the snapshot
	// will try to open a new connection to the database and fail.
	err = sqlDB.Close()
	if err != nil {
		log.Printf("close database: %v", err)
	}

	err = ctr.Snapshot(ctx)
	if err != nil {
		log.Printf("snapshot database: %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func setupTestDB(ctx context.Context, t *testing.T) *pgx.Conn {
	t.Helper()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		err := conn.Close(cleanupCtx)
		if err != nil {
			t.Errorf("close database connection: %v", err)
		}

		err = ctr.Restore(cleanupCtx)
		if err != nil {
			t.Errorf("restore database snapshot: %v", err)
		}
	})

	return conn
}
