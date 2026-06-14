package db_test

import (
	"context"
	"testing"

	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/ptr"
)

func TestQueries_GetMovie(t *testing.T) { //nolint:cyclop,funlen // Keep table cases and checks inline.
	t.Parallel()

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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			pool := setupTestDB(ctx, t)
			queries := db.New(pool)

			created, err := queries.CreateMovie(ctx, testCase.params)
			if err != nil {
				t.Fatalf("CreateMovie() error = %v", err)
			}

			got, err := queries.GetMovie(ctx, created.ID)
			if err != nil {
				t.Fatalf("GetMovie() error = %v", err)
			}

			if got.ID != created.ID {
				t.Errorf("ID = %v, want %v", got.ID, created.ID)
			}

			if got.Title != testCase.params.Title {
				t.Errorf("Title = %q, want %q", got.Title, testCase.params.Title)
			}

			if got.ReleaseYear != testCase.params.ReleaseYear {
				t.Errorf("ReleaseYear = %d, want %d", got.ReleaseYear, testCase.params.ReleaseYear)
			}

			switch {
			case got.RuntimeMin == nil && testCase.params.RuntimeMin == nil:
			case got.RuntimeMin != nil && testCase.params.RuntimeMin != nil &&
				*got.RuntimeMin == *testCase.params.RuntimeMin:
			default:
				t.Errorf("RuntimeMin = %v, want %v", got.RuntimeMin, testCase.params.RuntimeMin)
			}
		})
	}
}
