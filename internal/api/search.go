package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

func (s *Server) registerSearchRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "search-movies",
		Method:      http.MethodGet,
		Path:        "/search",
		Summary:     "Fuzzy-search movies by title, with each copy's physical location",
	}, s.search)
}

// SearchInput is the query for the search endpoint.
type SearchInput struct {
	Query string `doc:"Case-insensitive fuzzy title search" query:"q"`
	Limit int    `default:"20" doc:"Maximum number of results" maximum:"100" minimum:"1" query:"limit"`
}

// LocatedRelease pairs a release with where it physically lives.
type LocatedRelease struct {
	Release  Release  `json:"release"`
	Location Location `json:"location"`
}

// SearchResult is a matching movie together with the physical location of each
// of its placed releases.
type SearchResult struct {
	Movie           Movie            `json:"movie"`
	LocatedReleases []LocatedRelease `json:"locatedReleases"`
}

// SearchOutput is the response body for the search endpoint.
type SearchOutput struct {
	Body []SearchResult
}

func (s *Server) search(ctx context.Context, input *SearchInput) (*SearchOutput, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return &SearchOutput{Body: []SearchResult{}}, nil
	}

	movies, err := s.collection.SearchMovies(ctx, db.SearchMoviesParams{
		Query:       query,
		ResultLimit: int32(input.Limit), //nolint:gosec // bounded to [1,100] by huma validation.
	})
	if err != nil {
		return nil, mapErr(err)
	}

	if len(movies) == 0 {
		return &SearchOutput{Body: []SearchResult{}}, nil
	}

	located, err := s.locatedReleasesByMovie(ctx, movies)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(movies))
	for index, movie := range movies {
		releases := located[movie.ID]
		if releases == nil {
			releases = []LocatedRelease{}
		}

		results[index] = SearchResult{
			Movie:           movieFromDB(movie),
			LocatedReleases: releases,
		}
	}

	return &SearchOutput{Body: results}, nil
}

// locatedReleasesByMovie fetches the placed releases for the given movies and
// groups them by movie ID.
func (s *Server) locatedReleasesByMovie(
	ctx context.Context, movies []db.Movie,
) (map[uuid.UUID][]LocatedRelease, error) {
	movieIDs := make([]uuid.UUID, len(movies))
	for i, movie := range movies {
		movieIDs[i] = movie.ID
	}

	rows, err := s.collection.ListLocatedReleasesByMovies(ctx, movieIDs)
	if err != nil {
		return nil, mapErr(err)
	}

	byMovie := make(map[uuid.UUID][]LocatedRelease)
	for _, row := range rows {
		byMovie[row.HomeVideoRelease.MovieID] = append(byMovie[row.HomeVideoRelease.MovieID], LocatedRelease{
			Release: releaseFromDB(row.HomeVideoRelease),
			Location: Location{
				Bookcase:  bookcaseFromDB(row.Bookcase),
				Shelf:     shelfFromDB(row.Shelf),
				Placement: placementFromDB(row.Placement),
			},
		})
	}

	return byMovie, nil
}
