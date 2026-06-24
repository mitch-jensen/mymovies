package config_test

import (
	"testing"

	"github.com/mitch-jensen/mymovies/internal/config"
)

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("POSTGRES_ADDRESS", "db.example.com")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "app")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "movies")
	t.Setenv("SERVER_ADDRESS", "0.0.0.0")
	t.Setenv("SERVER_PORT", "8000")

	// An empty directory has no .env, so configuration comes purely from the environment.
	dbCfg, srvCfg, err := config.Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	//nolint:gosec // Test fixture, not a real credential.
	wantConn := "postgresql://app:secret@db.example.com:5432/movies"
	if got := dbCfg.ConnectionString(); got != wantConn {
		t.Errorf("ConnectionString() = %q, want %q", got, wantConn)
	}

	if srvCfg.Address != "0.0.0.0" {
		t.Errorf("server Address = %q, want %q", srvCfg.Address, "0.0.0.0")
	}

	if srvCfg.Port != "8000" {
		t.Errorf("server Port = %q, want %q", srvCfg.Port, "8000")
	}
}
