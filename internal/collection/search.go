package collection

import (
	"context"

	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

// SearchMovies returns movies whose titles fuzzily match the query.
func (c *Collection) SearchMovies(ctx context.Context, arg db.SearchMoviesParams) ([]db.Movie, error) {
	movies, err := c.q.SearchMovies(ctx, arg)
	if err != nil {
		return nil, wrap("search movies", err)
	}

	return movies, nil
}

// ListLocatedReleasesByMovies returns the placed releases of the given movies,
// each joined with its physical location.
func (c *Collection) ListLocatedReleasesByMovies(
	ctx context.Context, movieIDs []uuid.UUID,
) ([]db.ListLocatedReleasesByMoviesRow, error) {
	rows, err := c.q.ListLocatedReleasesByMovies(ctx, movieIDs)
	if err != nil {
		return nil, wrap("list located releases", err)
	}

	return rows, nil
}
