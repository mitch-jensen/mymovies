package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/mitch-jensen/mymovies/internal/api"
	"github.com/mitch-jensen/mymovies/ptr"
)

const releasesBase = "/releases"

func movieReleasesPath(movieID uuid.UUID) string {
	return moviePath(movieID) + "/releases"
}

func releasePath(id uuid.UUID) string {
	return releasesBase + "/" + id.String()
}

// createRelease adds a release to a movie through the API and returns it.
func createRelease(
	ctx context.Context, t *testing.T, handler http.Handler, movieID uuid.UUID, fields api.ReleaseFields,
) api.Release {
	t.Helper()

	recorder := doRequest(ctx, t, handler, http.MethodPost, movieReleasesPath(movieID), fields)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST release status = %d, want %d (body: %s)", recorder.Code, http.StatusCreated, recorder.Body)
	}

	var release api.Release

	err := json.Unmarshal(recorder.Body.Bytes(), &release)
	if err != nil {
		t.Fatalf("decode created release: %v", err)
	}

	return release
}

func TestCreateRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Perfect Blue", ReleaseYear: 1997})

	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{
		Studio:      ptr.To("Manga Entertainment"),
		DigitalCopy: false,
		Watched:     true,
	})

	if release.ID == uuid.Nil {
		t.Error("created release has nil ID")
	}

	if release.MovieID != movie.ID {
		t.Errorf("MovieID = %v, want %v", release.MovieID, movie.ID)
	}

	if release.Studio == nil || *release.Studio != "Manga Entertainment" {
		t.Errorf("Studio = %v, want %q", release.Studio, "Manga Entertainment")
	}
}

func TestCreateReleaseMovieNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodPost, movieReleasesPath(uuid.New()), api.ReleaseFields{
		DigitalCopy: false,
		Watched:     false,
	})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("POST status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestListMovieReleases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Ghost in the Shell", ReleaseYear: 1995})
	createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})
	createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: true, Watched: true})

	recorder := doRequest(ctx, t, handler, http.MethodGet, movieReleasesPath(movie.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET releases status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var releases []api.Release

	err := json.Unmarshal(recorder.Body.Bytes(), &releases)
	if err != nil {
		t.Fatalf("decode releases: %v", err)
	}

	if len(releases) != 2 {
		t.Errorf("len(releases) = %d, want 2", len(releases))
	}
}

func TestGetReleaseNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodGet, releasePath(uuid.New()), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestUpdateRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Paprika", ReleaseYear: 2006})
	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})

	recorder := doRequest(ctx, t, handler, http.MethodPut, releasePath(release.ID), api.ReleaseFields{
		Studio:      ptr.To("Sony"),
		DigitalCopy: true,
		Watched:     true,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	var got api.Release

	err := json.Unmarshal(recorder.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("decode release: %v", err)
	}

	if !got.Watched {
		t.Error("Watched = false, want true")
	}

	if got.Studio == nil || *got.Studio != "Sony" {
		t.Errorf("Studio = %v, want %q", got.Studio, "Sony")
	}
}

func TestDeleteRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Millennium Actress", ReleaseYear: 2001})
	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})

	recorder := doRequest(ctx, t, handler, http.MethodDelete, releasePath(release.ID), nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	recorder = doRequest(ctx, t, handler, http.MethodGet, releasePath(release.ID), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET after delete status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
