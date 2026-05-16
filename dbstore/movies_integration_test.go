package db_test

import (
	"context"
	"testing"

	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/ptr"
)

func TestQueries_GetMovie(t *testing.T) {
	tests := []struct {
		name   string
		params db.CreateMovieParams
	}{
		{
			name: "all fields set",
			params: db.CreateMovieParams{
				Title:       "The Abominable Dr. Phibes",
				ReleaseYear: 1971,
				RuntimeMin:  ptr.To(int32(94)),
			},
		},
		{
			name: "null runtime",
			params: db.CreateMovieParams{
				Title:       "Eraserhead",
				ReleaseYear: 1977,
				RuntimeMin:  nil,
			},
		},
		{
			name: "recent release",
			params: db.CreateMovieParams{
				Title:       "Everything Everywhere All at Once",
				ReleaseYear: 2022,
				RuntimeMin:  ptr.To(int32(139)),
			},
		},
		{
			name: "release year in the future",
			params: db.CreateMovieParams{
				Title:       "Future Movie",
				ReleaseYear: 2028,
				RuntimeMin:  ptr.To(int32(120)),
			},
		},
	}

	for _, tt := range tests { //nolint:paralleltest // Test cases share one container snapshot and restore it after each run.
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			conn := setupTestDB(ctx, t)
			queries := db.New(conn)

			created, err := queries.CreateMovie(ctx, tt.params)
			if err != nil {
				t.Fatalf("CreateMovie() error = %v", err)
			}

			got, err := queries.GetMovie(ctx, created.ID)
			if err != nil {
				t.Fatalf("GetMovie() error = %v", err)
			}

			assertMovie(t, got, created.ID, tt.params)
		})
	}
}

func assertMovie(t *testing.T, got db.Movie, wantID int32, want db.CreateMovieParams) {
	t.Helper()

	if got.ID != wantID {
		t.Errorf("ID = %v, want %v", got.ID, wantID)
	}

	if got.Title != want.Title {
		t.Errorf("Title = %q, want %q", got.Title, want.Title)
	}

	if got.ReleaseYear != want.ReleaseYear {
		t.Errorf("ReleaseYear = %d, want %d", got.ReleaseYear, want.ReleaseYear)
	}

	assertRuntimeMin(t, got.RuntimeMin, want.RuntimeMin)
}

func assertRuntimeMin(t *testing.T, got *int32, want *int32) {
	t.Helper()

	if got != nil && want != nil {
		if *got != *want {
			t.Errorf("RuntimeMin: got %d, want %d", *got, *want)
		}

		return
	}

	if got != want {
		t.Errorf("RuntimeMin: got %v, want %v", got, want)
	}
}
