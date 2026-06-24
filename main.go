// Package main runs a small sample program against the movie database.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mitch-jensen/mymovies/internal/api"
	"github.com/mitch-jensen/mymovies/internal/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbCfg, srvCfg, err := config.Load(".")
	if err != nil {
		slog.Error("failed to load configuration", "error", err)

		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, dbCfg.ConnectionString())
	if err != nil {
		slog.Error("failed to create database pool", "error", err)

		return 1
	}
	defer pool.Close()

	addr := net.JoinHostPort(srvCfg.Address, srvCfg.Port)
	slog.Info("starting server", "address", addr)

	srv := api.NewServer(pool)

	err = srv.Run(ctx, addr)
	if err != nil {
		slog.Error("server stopped", "error", err)

		return 1
	}

	return 0
}
