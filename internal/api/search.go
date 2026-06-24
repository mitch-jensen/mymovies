package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/mitch-jensen/mymovies/internal/collection"
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
	limit := int32(input.Limit) //nolint:gosec // bounded to [1,100] by huma validation.

	results, err := s.collection.Search(ctx, input.Query, limit)
	if err != nil {
		return nil, mapErr(err)
	}

	body := make([]SearchResult, len(results))
	for index, result := range results {
		body[index] = searchResultFromDB(result)
	}

	return &SearchOutput{Body: body}, nil
}

func searchResultFromDB(result collection.SearchResult) SearchResult {
	located := make([]LocatedRelease, len(result.Releases))
	for index, release := range result.Releases {
		located[index] = LocatedRelease{
			Release: releaseFromDB(release.Release),
			Location: Location{
				Bookcase:  bookcaseFromDB(release.Bookcase),
				Shelf:     shelfFromDB(release.Shelf),
				Placement: placementFromDB(release.Placement),
			},
		}
	}

	return SearchResult{
		Movie:           movieFromDB(result.Movie),
		LocatedReleases: located,
	}
}
