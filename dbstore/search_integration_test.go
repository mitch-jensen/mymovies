package db_test

import (
	"context"
	"testing"

	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/internal/testdb"
)

func TestQueries_SearchMovies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	for _, title := range []string{"Blade Runner", "Blade Runner 2049", "Ran"} {
		_, err := queries.CreateMovie(ctx, db.CreateMovieParams{Title: title, ReleaseYear: 2000})
		if err != nil {
			t.Fatalf("CreateMovie(%q) error = %v", title, err)
		}
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
	}{
		{name: "case-insensitive substring", query: "blade", wantCount: 2},
		{name: "fuzzy / typo tolerant", query: "Blaed Runner", wantCount: 2},
		{name: "exact title", query: "Ran", wantCount: 1},
		{name: "no match", query: "Nosferatu", wantCount: 0},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			movies, err := queries.SearchMovies(ctx, db.SearchMoviesParams{
				Query:       testCase.query,
				ResultLimit: 20,
			})
			if err != nil {
				t.Fatalf("SearchMovies(%q) error = %v", testCase.query, err)
			}

			if len(movies) != testCase.wantCount {
				t.Errorf("SearchMovies(%q) returned %d movies, want %d", testCase.query, len(movies), testCase.wantCount)
			}
		})
	}
}

func TestQueries_SearchMoviesRespectsLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	for _, title := range []string{"Alien", "Aliens", "Alien 3", "Alien Resurrection"} {
		_, err := queries.CreateMovie(ctx, db.CreateMovieParams{Title: title, ReleaseYear: 2000})
		if err != nil {
			t.Fatalf("CreateMovie(%q) error = %v", title, err)
		}
	}

	movies, err := queries.SearchMovies(ctx, db.SearchMoviesParams{Query: "alien", ResultLimit: 2})
	if err != nil {
		t.Fatalf("SearchMovies() error = %v", err)
	}

	if len(movies) != 2 {
		t.Errorf("SearchMovies() with limit 2 returned %d movies, want 2", len(movies))
	}
}
