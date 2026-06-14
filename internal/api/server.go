// Package api exposes the HTTP API for the movie service.
package api

import (
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

// Start listens for HTTP requests on addr.
func (s *Server) Start(addr string) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
