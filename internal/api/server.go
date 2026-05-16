package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

type Server struct {
	queries *db.Queries
	router  *chi.Mux
	api     huma.API
}

func NewServer(pool *pgxpool.Pool) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	humaConfig := huma.DefaultConfig("My Movies API", "1.0.0")
	api := humachi.New(r, humaConfig)

	s := &Server{
		queries: db.New(pool),
		router:  r,
		api:     api,
	}
	s.registerMovieRoutes()
	return s
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
