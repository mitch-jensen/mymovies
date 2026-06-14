package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/mitch-jensen/mymovies/internal/api"
)

func searchPath(query string) string {
	return "/search?q=" + url.QueryEscape(query)
}

func doSearch(ctx context.Context, t *testing.T, handler http.Handler, query string) []api.SearchResult {
	t.Helper()

	recorder := doRequest(ctx, t, handler, http.MethodGet, searchPath(query), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET search status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	var results []api.SearchResult

	err := json.Unmarshal(recorder.Body.Bytes(), &results)
	if err != nil {
		t.Fatalf("decode search results: %v", err)
	}

	return results
}

func TestSearchReturnsInlineLocation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Blade Runner", ReleaseYear: 1982})
	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})
	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 1)

	placed := doRequest(ctx, t, handler, http.MethodPut, placementPath(release.ID), api.PlacementFields{
		ShelfID:  shelf.ID,
		Position: 4,
	})
	if placed.Code != http.StatusOK {
		t.Fatalf("PUT placement status = %d, want %d", placed.Code, http.StatusOK)
	}

	results := doSearch(ctx, t, handler, "blade")

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].Movie.ID != movie.ID {
		t.Errorf("Movie.ID = %v, want %v", results[0].Movie.ID, movie.ID)
	}

	if len(results[0].LocatedReleases) != 1 {
		t.Fatalf("got %d located releases, want 1", len(results[0].LocatedReleases))
	}

	located := results[0].LocatedReleases[0]
	if located.Release.ID != release.ID {
		t.Errorf("Release.ID = %v, want %v", located.Release.ID, release.ID)
	}

	if located.Location.Shelf.ID != shelf.ID {
		t.Errorf("Location.Shelf.ID = %v, want %v", located.Location.Shelf.ID, shelf.ID)
	}
}

func TestSearchMatchWithoutLocation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	createMovie(ctx, t, handler, api.MovieFields{Title: "Tetsuo", ReleaseYear: 1989})

	results := doSearch(ctx, t, handler, "tetsuo")

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	// A match with no placed copies must still return an (empty) array, not null.
	if results[0].LocatedReleases == nil {
		t.Error("LocatedReleases is nil, want empty slice")
	}

	if len(results[0].LocatedReleases) != 0 {
		t.Errorf("got %d located releases, want 0", len(results[0].LocatedReleases))
	}
}

func TestSearchNoMatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	createMovie(ctx, t, handler, api.MovieFields{Title: "Possession", ReleaseYear: 1981})

	results := doSearch(ctx, t, handler, "nosferatu")

	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
