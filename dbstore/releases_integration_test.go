package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/internal/testdb"
	"github.com/mitch-jensen/mymovies/ptr"
)

func seedMovie(ctx context.Context, t *testing.T, queries *db.Queries) db.Movie {
	t.Helper()

	movie, err := queries.CreateMovie(ctx, db.CreateMovieParams{Title: "Tetsuo", ReleaseYear: 1989})
	if err != nil {
		t.Fatalf("CreateMovie() error = %v", err)
	}

	return movie
}

func TestQueries_CreateAndGetHomeVideoRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))
	movie := seedMovie(ctx, t, queries)

	price := decimal.RequireFromString("24.99")

	created, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
		MovieID:     movie.ID,
		Studio:      ptr.To("Bandai"),
		BluRayDiscs: ptr.To(int32(2)),
		DigitalCopy: true,
		Watched:     false,
		Price:       &price,
	})
	if err != nil {
		t.Fatalf("CreateHomeVideoRelease() error = %v", err)
	}

	got, err := queries.GetHomeVideoRelease(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetHomeVideoRelease() error = %v", err)
	}

	if got.MovieID != movie.ID {
		t.Errorf("MovieID = %v, want %v", got.MovieID, movie.ID)
	}

	if got.Studio == nil || *got.Studio != "Bandai" {
		t.Errorf("Studio = %v, want %q", got.Studio, "Bandai")
	}

	if got.Price == nil || !got.Price.Equal(price) {
		t.Errorf("Price = %v, want %v", got.Price, price)
	}
}

func TestQueries_ListHomeVideoReleasesByMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))
	movie := seedMovie(ctx, t, queries)
	other := seedMovie(ctx, t, queries)

	for range 2 {
		_, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
			MovieID: movie.ID, DigitalCopy: false, Watched: false,
		})
		if err != nil {
			t.Fatalf("CreateHomeVideoRelease() error = %v", err)
		}
	}

	_, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
		MovieID: other.ID, DigitalCopy: false, Watched: false,
	})
	if err != nil {
		t.Fatalf("CreateHomeVideoRelease() error = %v", err)
	}

	releases, err := queries.ListHomeVideoReleasesByMovie(ctx, movie.ID)
	if err != nil {
		t.Fatalf("ListHomeVideoReleasesByMovie() error = %v", err)
	}

	if len(releases) != 2 {
		t.Errorf("len(releases) = %d, want 2", len(releases))
	}
}

func TestQueries_UpdateHomeVideoRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))
	movie := seedMovie(ctx, t, queries)

	created, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
		MovieID: movie.ID, DigitalCopy: false, Watched: false,
	})
	if err != nil {
		t.Fatalf("CreateHomeVideoRelease() error = %v", err)
	}

	updated, err := queries.UpdateHomeVideoRelease(ctx, db.UpdateHomeVideoReleaseParams{
		ID:          created.ID,
		Studio:      ptr.To("Criterion"),
		DigitalCopy: true,
		Watched:     true,
	})
	if err != nil {
		t.Fatalf("UpdateHomeVideoRelease() error = %v", err)
	}

	if updated.Studio == nil || *updated.Studio != "Criterion" {
		t.Errorf("Studio = %v, want %q", updated.Studio, "Criterion")
	}

	if !updated.Watched {
		t.Error("Watched = false, want true")
	}
}

func TestQueries_UpdateHomeVideoRelease_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	_, err := queries.UpdateHomeVideoRelease(ctx, db.UpdateHomeVideoReleaseParams{
		ID: uuid.New(), DigitalCopy: false, Watched: false,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("UpdateHomeVideoRelease() error = %v, want %v", err, pgx.ErrNoRows)
	}
}

func TestQueries_DeleteHomeVideoRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))
	movie := seedMovie(ctx, t, queries)

	created, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
		MovieID: movie.ID, DigitalCopy: false, Watched: false,
	})
	if err != nil {
		t.Fatalf("CreateHomeVideoRelease() error = %v", err)
	}

	err = queries.DeleteHomeVideoRelease(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteHomeVideoRelease() error = %v", err)
	}

	_, err = queries.GetHomeVideoRelease(ctx, created.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetHomeVideoRelease() after delete error = %v, want %v", err, pgx.ErrNoRows)
	}
}
