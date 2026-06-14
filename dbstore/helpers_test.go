package db_test

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	testDatabaseURLEnv = "MYMOVIES_TEST_DATABASE_URL"
	testTemplateDBEnv  = "MYMOVIES_TEST_TEMPLATE_DATABASE"

	testTemplateDB = "app_template"
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
	exitCode := runTestMain(ctx, m)

	os.Exit(exitCode)
}

func runTestMain(ctx context.Context, m *testing.M) int {
	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(testTemplateDB),
		postgres.WithUsername("app"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		log.Printf("error starting postgres container: %v", err)

		return 1
	}

	databaseURL, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("error getting postgres connection string: %v", err)

		return terminateContainer(ctx, ctr, 1)
	}

	err = runMigrations(ctx, databaseURL)
	if err != nil {
		log.Printf("error migrating template database: %v", err)

		return terminateContainer(ctx, ctr, 1)
	}

	err = os.Setenv(testDatabaseURLEnv, databaseURL)
	if err != nil {
		log.Printf("error setting %s: %v", testDatabaseURLEnv, err)

		return terminateContainer(ctx, ctr, 1)
	}

	err = os.Setenv(testTemplateDBEnv, testTemplateDB)
	if err != nil {
		log.Printf("error setting %s: %v", testTemplateDBEnv, err)

		return terminateContainer(ctx, ctr, 1)
	}

	return terminateContainer(ctx, ctr, m.Run())
}

func terminateContainer(ctx context.Context, ctr *postgres.PostgresContainer, exitCode int) int {
	err := ctr.Terminate(ctx)
	if err != nil {
		log.Printf("error terminating postgres container: %v", err)

		return 1
	}

	return exitCode
}

func setupTestDB(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()

	templateURL := os.Getenv(testDatabaseURLEnv)
	if templateURL == "" {
		t.Fatalf("%s is not set", testDatabaseURLEnv)
	}

	templateDB := os.Getenv(testTemplateDBEnv)
	if templateDB == "" {
		t.Fatalf("%s is not set", testTemplateDBEnv)
	}

	dbName := testDatabaseName(t)
	adminURL := databaseURL(t, templateURL, "postgres")

	adminPool := connectPool(ctx, t, adminURL)
	defer adminPool.Close()

	createDatabase(ctx, t, adminPool, dbName, templateDB)
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		dropDatabase(cleanupCtx, t, adminURL, dbName)
	})

	pool := connectPool(ctx, t, databaseURL(t, templateURL, dbName))
	t.Cleanup(pool.Close)

	return pool
}

func runMigrations(ctx context.Context, databaseURL string) error {
	pgxConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse pgx config: %w", err)
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	sqlDB := stdlib.OpenDB(*pgxConfig)

	err = goose.UpContext(ctx, sqlDB, MigrationsDir())
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("close migration database: %w", err)
	}

	return nil
}

func testDatabaseName(t *testing.T) string {
	t.Helper()

	cleanName := strings.Map(func(char rune) rune {
		if char >= 'a' && char <= 'z' {
			return char
		}

		if char >= '0' && char <= '9' {
			return char
		}

		return '_'
	}, strings.ToLower(t.Name()))

	cleanName = strings.Trim(cleanName, "_")
	if len(cleanName) > 32 {
		cleanName = cleanName[:32]
	}

	return fmt.Sprintf("test_%s_%d", cleanName, time.Now().UnixNano())
}

func connectPool(ctx context.Context, t *testing.T, databaseURL string) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	return pool
}

func createDatabase(ctx context.Context, t *testing.T, pool *pgxpool.Pool, dbName string, templateDB string) {
	t.Helper()

	sql := fmt.Sprintf(
		"CREATE DATABASE %s TEMPLATE %s",
		pgx.Identifier{dbName}.Sanitize(),
		pgx.Identifier{templateDB}.Sanitize(),
	)

	_, err := pool.Exec(ctx, sql)
	if err != nil {
		t.Fatalf("create test database: %v", err)
	}
}

func dropDatabase(ctx context.Context, t *testing.T, adminURL string, dbName string) {
	t.Helper()

	adminPool := connectPool(ctx, t, adminURL)
	defer adminPool.Close()

	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pgx.Identifier{dbName}.Sanitize())

	_, err := adminPool.Exec(ctx, sql)
	if err != nil {
		t.Errorf("drop test database: %v", err)
	}
}

func databaseURL(t *testing.T, adminURL string, dbName string) string {
	t.Helper()

	parsedURL, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("parse database URL: %v", err)
	}

	parsedURL.Path = "/" + dbName

	return parsedURL.String()
}
