// Package api exposes the HTTP API for the movie service.
package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

func (s *Server) registerMovieRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "list-movies",
		Method:      http.MethodGet,
		Path:        "/movies",
		Summary:     "List all movies",
	}, s.listMovies)

	huma.Register(s.api, huma.Operation{
		OperationID: "create-movie",
		Method:      http.MethodPost,
		Path:        "/movies",
		Summary:     "Create a movie",
	}, s.createMovie)
}

// ListMoviesOutput is the response body for the list movies endpoint.
type ListMoviesOutput struct {
	Body []db.Movie
}

func (s *Server) listMovies(ctx context.Context, _ *struct{}) (*ListMoviesOutput, error) {
	movies, err := s.queries.ListMovies(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list movies", err)
	}

	return &ListMoviesOutput{Body: movies}, nil
}

// CreateMovieOutput is the response body for the create movie endpoint.
type CreateMovieOutput struct {
	Body db.Movie
}

func (s *Server) createMovie(ctx context.Context, input *struct {
	Body db.CreateMovieParams
}) (*CreateMovieOutput, error) {
	movie, err := s.queries.CreateMovie(ctx, input.Body)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create movie", err)
	}

	return &CreateMovieOutput{Body: movie}, nil
}
