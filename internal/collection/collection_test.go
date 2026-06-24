package collection_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/internal/collection"
	"github.com/mitch-jensen/mymovies/internal/testdb"
)

func TestMain(m *testing.M) {
	os.Exit(testdb.Run(m))
}

func newCollection(ctx context.Context, t *testing.T) *collection.Collection {
	t.Helper()

	return collection.New(testdb.Setup(ctx, t))
}

func TestGetMovieReturnsErrNotFoundForMissingID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	_, err := col.GetMovie(ctx, uuid.New())
	if !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("GetMovie(missing) error = %v, want collection.ErrNotFound", err)
	}
}

func TestLocateReleaseReturnsErrNotFoundWhenUnplaced(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	_, err := col.LocateRelease(ctx, uuid.New())
	if !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("LocateRelease(missing) error = %v, want collection.ErrNotFound", err)
	}
}

func TestGetMovieReturnsStoredMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	created, err := col.CreateMovie(ctx, db.CreateMovieParams{Title: "Solaris", ReleaseYear: 1972})
	if err != nil {
		t.Fatalf("CreateMovie: %v", err)
	}

	got, err := col.GetMovie(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetMovie: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("GetMovie ID = %v, want %v", got.ID, created.ID)
	}
}
