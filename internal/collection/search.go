package collection

import (
	"context"
	"strings"

	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

// LocatedRelease is a placed release together with the physical location it sits
// in: which shelf of which bookcase, at which position.
type LocatedRelease struct {
	Release   db.HomeVideoRelease
	Bookcase  db.Bookcase
	Shelf     db.Shelf
	Placement db.Placement
}

// SearchResult is a matching movie together with each of its placed copies and
// their locations. Releases is empty when the movie matches but no copy is
// shelved.
type SearchResult struct {
	Movie    db.Movie
	Releases []LocatedRelease
}

// Search fuzzily matches movies by title and returns each match with the
// physical location of every placed copy. A blank query yields no results.
func (c *Collection) Search(ctx context.Context, query string, limit int32) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchResult{}, nil
	}

	movies, err := c.q.SearchMovies(ctx, db.SearchMoviesParams{Query: query, ResultLimit: limit})
	if err != nil {
		return nil, wrap("search movies", err)
	}

	if len(movies) == 0 {
		return []SearchResult{}, nil
	}

	located, err := c.locatedReleasesByMovie(ctx, movies)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(movies))
	for index, movie := range movies {
		results[index] = SearchResult{Movie: movie, Releases: located[movie.ID]}
	}

	return results, nil
}

// locatedReleasesByMovie loads the placed releases of the given movies and groups
// them by movie ID.
func (c *Collection) locatedReleasesByMovie(
	ctx context.Context, movies []db.Movie,
) (map[uuid.UUID][]LocatedRelease, error) {
	movieIDs := make([]uuid.UUID, len(movies))
	for index, movie := range movies {
		movieIDs[index] = movie.ID
	}

	rows, err := c.q.ListLocatedReleasesByMovies(ctx, movieIDs)
	if err != nil {
		return nil, wrap("list located releases", err)
	}

	byMovie := make(map[uuid.UUID][]LocatedRelease)
	for _, row := range rows {
		byMovie[row.HomeVideoRelease.MovieID] = append(byMovie[row.HomeVideoRelease.MovieID], LocatedRelease{
			Release:   row.HomeVideoRelease,
			Bookcase:  row.Bookcase,
			Shelf:     row.Shelf,
			Placement: row.Placement,
		})
	}

	return byMovie, nil
}
