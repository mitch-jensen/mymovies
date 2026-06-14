package db_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/internal/testdb"
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
			pool := testdb.Setup(ctx, t)
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

func TestQueries_ListMovies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	// ListMovies orders by title, so seed out of order to prove the sort.
	titles := []string{"Possession", "Akira", "Brazil"}
	for _, title := range titles {
		_, err := queries.CreateMovie(ctx, db.CreateMovieParams{Title: title, ReleaseYear: 1985})
		if err != nil {
			t.Fatalf("CreateMovie(%q) error = %v", title, err)
		}
	}

	movies, err := queries.ListMovies(ctx)
	if err != nil {
		t.Fatalf("ListMovies() error = %v", err)
	}

	got := make([]string, len(movies))
	for i, movie := range movies {
		got[i] = movie.Title
	}

	want := []string{"Akira", "Brazil", "Possession"}
	if !slices.Equal(got, want) {
		t.Errorf("ListMovies() titles = %v, want %v", got, want)
	}
}

func TestQueries_UpdateMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	created, err := queries.CreateMovie(ctx, db.CreateMovieParams{
		Title:       "Blade Runner",
		ReleaseYear: 1982,
		RuntimeMin:  ptr.To(int32(117)),
	})
	if err != nil {
		t.Fatalf("CreateMovie() error = %v", err)
	}

	updated, err := queries.UpdateMovie(ctx, db.UpdateMovieParams{
		ID:          created.ID,
		Title:       "Blade Runner: The Final Cut",
		ReleaseYear: 2007,
		RuntimeMin:  ptr.To(int32(118)),
	})
	if err != nil {
		t.Fatalf("UpdateMovie() error = %v", err)
	}

	if updated.Title != "Blade Runner: The Final Cut" {
		t.Errorf("Title = %q, want %q", updated.Title, "Blade Runner: The Final Cut")
	}

	if updated.ReleaseYear != 2007 {
		t.Errorf("ReleaseYear = %d, want %d", updated.ReleaseYear, 2007)
	}
}

func TestQueries_UpdateMovie_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	_, err := queries.UpdateMovie(ctx, db.UpdateMovieParams{
		ID:          uuid.New(),
		Title:       "Nonexistent",
		ReleaseYear: 2000,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("UpdateMovie() error = %v, want %v", err, pgx.ErrNoRows)
	}
}

func TestQueries_DeleteMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	created, err := queries.CreateMovie(ctx, db.CreateMovieParams{Title: "Solaris", ReleaseYear: 1972})
	if err != nil {
		t.Fatalf("CreateMovie() error = %v", err)
	}

	err = queries.DeleteMovie(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteMovie() error = %v", err)
	}

	_, err = queries.GetMovie(ctx, created.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetMovie() after delete error = %v, want %v", err, pgx.ErrNoRows)
	}
}
