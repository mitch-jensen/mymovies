// Package api exposes the HTTP API for the movie service.
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// Server serves the movie API over HTTP.
type Server struct {
	queries *db.Queries
	router  *chi.Mux
	api     huma.API
}

// NewServer builds a movie API server backed by the given database pool.
func NewServer(pool *pgxpool.Pool) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	humaConfig := huma.DefaultConfig("My Movies API", "1.0.0")
	humaAPI := humachi.New(router, humaConfig)

	server := &Server{
		queries: db.New(pool),
		router:  router,
		api:     humaAPI,
	}
	server.registerMovieRoutes()

	return server
}

// Run serves HTTP requests on addr until ctx is cancelled, then shuts down
// gracefully within shutdownTimeout.
func (s *Server) Run(ctx context.Context, addr string) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}

		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}

		return nil
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	return nil
}
