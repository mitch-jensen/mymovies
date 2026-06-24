package collection

import (
	"context"

	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

// ListMovies returns all movies, ordered by title.
func (c *Collection) ListMovies(ctx context.Context) ([]db.Movie, error) {
	movies, err := c.q.ListMovies(ctx)
	if err != nil {
		return nil, wrap("list movies", err)
	}

	return movies, nil
}

// CreateMovie inserts a movie and returns it.
func (c *Collection) CreateMovie(ctx context.Context, arg db.CreateMovieParams) (db.Movie, error) {
	movie, err := c.q.CreateMovie(ctx, arg)
	if err != nil {
		return db.Movie{}, wrap("create movie", err)
	}

	return movie, nil
}

// GetMovie returns the movie with the given ID, or ErrNotFound if none exists.
func (c *Collection) GetMovie(ctx context.Context, id uuid.UUID) (db.Movie, error) {
	movie, err := c.q.GetMovie(ctx, id)
	if err != nil {
		return db.Movie{}, notFound("get movie", err)
	}

	return movie, nil
}

// UpdateMovie updates a movie and returns it, or ErrNotFound if none exists.
func (c *Collection) UpdateMovie(ctx context.Context, arg db.UpdateMovieParams) (db.Movie, error) {
	movie, err := c.q.UpdateMovie(ctx, arg)
	if err != nil {
		return db.Movie{}, notFound("update movie", err)
	}

	return movie, nil
}

// DeleteMovie removes a movie by ID.
func (c *Collection) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	err := c.q.DeleteMovie(ctx, id)
	if err != nil {
		return wrap("delete movie", err)
	}

	return nil
}
