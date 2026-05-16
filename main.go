// Package main runs a small sample program against the movie database.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mitch-jensen/mymovies/internal/api"
	"github.com/mitch-jensen/mymovies/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbCfg, srvCfg, err := config.Load(".")
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbCfg.ConnectionString())
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("starting server")
	srv := api.NewServer(pool)
	srv.Start(net.JoinHostPort(srvCfg.Address, srvCfg.Port))
}
