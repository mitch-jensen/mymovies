package db_test

import (
	"context"

	"testing"

	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/ptr"
)

func TestQueries_GetMovie(t *testing.T) {
	ctx := context.Background()

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := setupTestDB(t)
			queries := db.New(conn)

			created, err := queries.CreateMovie(ctx, tt.params)
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
			if got.Title != tt.params.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.params.Title)
			}
			if got.ReleaseYear != tt.params.ReleaseYear {
				t.Errorf("ReleaseYear = %d, want %d", got.ReleaseYear, tt.params.ReleaseYear)
			}
			if got.RuntimeMin != nil && tt.params.RuntimeMin != nil {
				if *got.RuntimeMin != *tt.params.RuntimeMin {
					t.Errorf("RuntimeMin: got %d, want %d", *got.RuntimeMin, *tt.params.RuntimeMin)
				}
			} else if got.RuntimeMin != tt.params.RuntimeMin {
				t.Errorf("RuntimeMin: got %v, want %v", got.RuntimeMin, tt.params.RuntimeMin)
			}
		})
	}
}
