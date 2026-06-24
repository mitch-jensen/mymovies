package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

const movieByIDPath = "/movies/{id}"

func (s *Server) registerMovieRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "list-movies",
		Method:      http.MethodGet,
		Path:        "/movies",
		Summary:     "List all movies",
	}, s.listMovies)

	huma.Register(s.api, huma.Operation{
		OperationID:   "create-movie",
		Method:        http.MethodPost,
		Path:          "/movies",
		Summary:       "Create a movie",
		DefaultStatus: http.StatusCreated,
	}, s.createMovie)

	huma.Register(s.api, huma.Operation{
		OperationID: "get-movie",
		Method:      http.MethodGet,
		Path:        movieByIDPath,
		Summary:     "Get a movie by ID",
	}, s.getMovie)

	huma.Register(s.api, huma.Operation{
		OperationID: "update-movie",
		Method:      http.MethodPut,
		Path:        movieByIDPath,
		Summary:     "Update a movie",
	}, s.updateMovie)

	huma.Register(s.api, huma.Operation{
		OperationID:   "delete-movie",
		Method:        http.MethodDelete,
		Path:          movieByIDPath,
		Summary:       "Delete a movie",
		DefaultStatus: http.StatusNoContent,
	}, s.deleteMovie)
}

// Movie is the API representation of a movie. It is intentionally decoupled from
// the database model so the public schema (and the generated client) is stable.
type Movie struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	ReleaseYear int32     `json:"releaseYear"`
	RuntimeMin  *int32    `json:"runtimeMin,omitempty"`
}

func movieFromDB(movie db.Movie) Movie {
	return Movie{
		ID:          movie.ID,
		Title:       movie.Title,
		ReleaseYear: movie.ReleaseYear,
		RuntimeMin:  movie.RuntimeMin,
	}
}

// MovieFields holds the mutable fields of a movie, used in create and update
// request bodies.
type MovieFields struct {
	Title       string `json:"title"`
	ReleaseYear int32  `json:"releaseYear"`
	RuntimeMin  *int32 `json:"runtimeMin,omitempty"`
}

// MovieIDInput identifies a movie by its path parameter.
type MovieIDInput struct {
	ID uuid.UUID `doc:"Movie ID" path:"id"`
}

// ListMoviesOutput is the response body for the list movies endpoint.
type ListMoviesOutput struct {
	Body []Movie
}

func (s *Server) listMovies(ctx context.Context, _ *struct{}) (*ListMoviesOutput, error) {
	movies, err := s.collection.ListMovies(ctx)
	if err != nil {
		return nil, mapErr(err)
	}

	return &ListMoviesOutput{Body: mapSlice(movies, movieFromDB)}, nil
}

// MovieInput is a request carrying movie field values in its body.
type MovieInput struct {
	Body MovieFields
}

// MovieOutput is a response carrying a single movie.
type MovieOutput struct {
	Body Movie
}

func (s *Server) createMovie(ctx context.Context, input *MovieInput) (*MovieOutput, error) {
	movie, err := s.collection.CreateMovie(ctx, db.CreateMovieParams{
		Title:       input.Body.Title,
		ReleaseYear: input.Body.ReleaseYear,
		RuntimeMin:  input.Body.RuntimeMin,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	return &MovieOutput{Body: movieFromDB(movie)}, nil
}

func (s *Server) getMovie(ctx context.Context, input *MovieIDInput) (*MovieOutput, error) {
	movie, err := s.collection.GetMovie(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}

	return &MovieOutput{Body: movieFromDB(movie)}, nil
}

// UpdateMovieInput carries the movie ID in the path and new field values in the
// body.
type UpdateMovieInput struct {
	ID   uuid.UUID `doc:"Movie ID" path:"id"`
	Body MovieFields
}

func (s *Server) updateMovie(ctx context.Context, input *UpdateMovieInput) (*MovieOutput, error) {
	movie, err := s.collection.UpdateMovie(ctx, db.UpdateMovieParams{
		ID:          input.ID,
		Title:       input.Body.Title,
		ReleaseYear: input.Body.ReleaseYear,
		RuntimeMin:  input.Body.RuntimeMin,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	return &MovieOutput{Body: movieFromDB(movie)}, nil
}

func (s *Server) deleteMovie(ctx context.Context, input *MovieIDInput) (*struct{}, error) {
	err := s.collection.DeleteMovie(ctx, input.ID)
	if err != nil {
		return nil, mapErr(err)
	}

	return nil, nil //nolint:nilnil // 204 No Content: no body and no error.
}
