package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mitch-jensen/mymovies/internal/api"
	"github.com/mitch-jensen/mymovies/internal/testdb"
	"github.com/mitch-jensen/mymovies/ptr"
)

const moviesPath = "/movies"

func TestMain(m *testing.M) {
	os.Exit(testdb.Run(m))
}

func newHandler(ctx context.Context, t *testing.T) http.Handler {
	t.Helper()

	return api.NewServer(testdb.Setup(ctx, t)).Handler()
}

func moviePath(id uuid.UUID) string {
	return moviesPath + "/" + id.String()
}

func doRequest(
	ctx context.Context, t *testing.T, handler http.Handler, method, path string, body any,
) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader

	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}

		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequestWithContext(ctx, method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	return recorder
}

// createMovie creates a movie through the API and returns it.
func createMovie(ctx context.Context, t *testing.T, handler http.Handler, fields api.MovieFields) api.Movie {
	t.Helper()

	recorder := doRequest(ctx, t, handler, http.MethodPost, moviesPath, fields)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST %s status = %d, want %d (body: %s)", moviesPath, recorder.Code, http.StatusCreated, recorder.Body)
	}

	var movie api.Movie

	err := json.Unmarshal(recorder.Body.Bytes(), &movie)
	if err != nil {
		t.Fatalf("decode created movie: %v", err)
	}

	return movie
}

func TestCreateMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{
		Title:       "Stalker",
		ReleaseYear: 1979,
		RuntimeMin:  ptr.To(int32(162)),
	})

	if movie.ID == uuid.Nil {
		t.Error("created movie has nil ID")
	}

	if movie.Title != "Stalker" {
		t.Errorf("Title = %q, want %q", movie.Title, "Stalker")
	}
}

func TestGetMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	created := createMovie(ctx, t, handler, api.MovieFields{Title: "Tampopo", ReleaseYear: 1985})

	recorder := doRequest(ctx, t, handler, http.MethodGet, moviePath(created.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var got api.Movie

	err := json.Unmarshal(recorder.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("decode movie: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %v, want %v", got.ID, created.ID)
	}
}

func TestGetMovieNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodGet, moviePath(uuid.New()), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestUpdateMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	created := createMovie(ctx, t, handler, api.MovieFields{Title: "Brazil", ReleaseYear: 1985})

	recorder := doRequest(ctx, t, handler, http.MethodPut, moviePath(created.ID), api.MovieFields{
		Title:       "Brazil (Director's Cut)",
		ReleaseYear: 1985,
		RuntimeMin:  ptr.To(int32(143)),
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	var got api.Movie

	err := json.Unmarshal(recorder.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("decode movie: %v", err)
	}

	if got.Title != "Brazil (Director's Cut)" {
		t.Errorf("Title = %q, want %q", got.Title, "Brazil (Director's Cut)")
	}
}

func TestUpdateMovieNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodPut, moviePath(uuid.New()), api.MovieFields{
		Title:       "Ghost",
		ReleaseYear: 2000,
	})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("PUT status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeleteMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	created := createMovie(ctx, t, handler, api.MovieFields{Title: "Paprika", ReleaseYear: 2006})

	recorder := doRequest(ctx, t, handler, http.MethodDelete, moviePath(created.ID), nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	recorder = doRequest(ctx, t, handler, http.MethodGet, moviePath(created.ID), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET after delete status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestListMovies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	createMovie(ctx, t, handler, api.MovieFields{Title: "Dune", ReleaseYear: 2021})
	createMovie(ctx, t, handler, api.MovieFields{Title: "Arrival", ReleaseYear: 2016})

	recorder := doRequest(ctx, t, handler, http.MethodGet, moviesPath, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", moviesPath, recorder.Code, http.StatusOK)
	}

	var movies []api.Movie

	err := json.Unmarshal(recorder.Body.Bytes(), &movies)
	if err != nil {
		t.Fatalf("decode movies: %v", err)
	}

	if len(movies) != 2 {
		t.Fatalf("len(movies) = %d, want 2", len(movies))
	}

	// ListMovies orders by title.
	if movies[0].Title != "Arrival" || movies[1].Title != "Dune" {
		t.Errorf("titles = [%q, %q], want [Arrival, Dune]", movies[0].Title, movies[1].Title)
	}
}
