package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/internal/testdb"
)

const (
	bookcaseLounge = "Lounge"
	bookcaseStudy  = "Study"
)

// seedRelease creates a movie and a bare release of it, returning the release.
func seedRelease(ctx context.Context, t *testing.T, queries *db.Queries) db.HomeVideoRelease {
	t.Helper()

	movie := seedMovie(ctx, t, queries)

	release, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
		MovieID: movie.ID, DigitalCopy: false, Watched: false,
	})
	if err != nil {
		t.Fatalf("CreateHomeVideoRelease() error = %v", err)
	}

	return release
}

func TestQueries_ListBookcasesOrdered(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	// Insert out of order to prove ListBookcases orders by position.
	for _, bc := range []db.CreateBookcaseParams{
		{Name: "Hallway", Position: 2},
		{Name: bookcaseLounge, Position: 0},
		{Name: bookcaseStudy, Position: 1},
	} {
		_, err := queries.CreateBookcase(ctx, bc)
		if err != nil {
			t.Fatalf("CreateBookcase() error = %v", err)
		}
	}

	bookcases, err := queries.ListBookcases(ctx)
	if err != nil {
		t.Fatalf("ListBookcases() error = %v", err)
	}

	got := make([]string, len(bookcases))
	for i, bookcase := range bookcases {
		got[i] = bookcase.Name
	}

	want := []string{bookcaseLounge, bookcaseStudy, "Hallway"}
	if len(got) != len(want) {
		t.Fatalf("got %d bookcases, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("bookcase[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestQueries_PlaceAndLocateRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	release := seedRelease(ctx, t, queries)

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseLounge, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	shelf, err := queries.CreateShelf(ctx, db.CreateShelfParams{BookcaseID: bookcase.ID, Position: 3})
	if err != nil {
		t.Fatalf("CreateShelf() error = %v", err)
	}

	_, err = queries.PlaceRelease(ctx, db.PlaceReleaseParams{
		ReleaseID: release.ID, ShelfID: shelf.ID, Position: 7,
	})
	if err != nil {
		t.Fatalf("PlaceRelease() error = %v", err)
	}

	located, err := queries.LocateRelease(ctx, release.ID)
	if err != nil {
		t.Fatalf("LocateRelease() error = %v", err)
	}

	if located.Bookcase.Name != bookcaseLounge {
		t.Errorf("Bookcase.Name = %q, want %q", located.Bookcase.Name, bookcaseLounge)
	}

	if located.Shelf.ID != shelf.ID {
		t.Errorf("Shelf.ID = %v, want %v", located.Shelf.ID, shelf.ID)
	}

	if located.Placement.Position != 7 {
		t.Errorf("Placement.Position = %d, want 7", located.Placement.Position)
	}
}

func TestQueries_PlaceReleaseMovesExisting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	release := seedRelease(ctx, t, queries)

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseStudy, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	first, err := queries.CreateShelf(ctx, db.CreateShelfParams{BookcaseID: bookcase.ID, Position: 0})
	if err != nil {
		t.Fatalf("CreateShelf() error = %v", err)
	}

	second, err := queries.CreateShelf(ctx, db.CreateShelfParams{BookcaseID: bookcase.ID, Position: 1})
	if err != nil {
		t.Fatalf("CreateShelf() error = %v", err)
	}

	_, err = queries.PlaceRelease(ctx, db.PlaceReleaseParams{ReleaseID: release.ID, ShelfID: first.ID, Position: 0})
	if err != nil {
		t.Fatalf("first PlaceRelease() error = %v", err)
	}

	// Placing the same release again should move it, not create a duplicate.
	_, err = queries.PlaceRelease(ctx, db.PlaceReleaseParams{ReleaseID: release.ID, ShelfID: second.ID, Position: 5})
	if err != nil {
		t.Fatalf("second PlaceRelease() error = %v", err)
	}

	located, err := queries.LocateRelease(ctx, release.ID)
	if err != nil {
		t.Fatalf("LocateRelease() error = %v", err)
	}

	if located.Shelf.ID != second.ID {
		t.Errorf("Shelf.ID = %v, want %v (release should have moved)", located.Shelf.ID, second.ID)
	}

	onFirst, err := queries.ListPlacementsByShelf(ctx, first.ID)
	if err != nil {
		t.Fatalf("ListPlacementsByShelf() error = %v", err)
	}

	if len(onFirst) != 0 {
		t.Errorf("first shelf has %d placements, want 0", len(onFirst))
	}
}

func TestQueries_LocateReleaseNotPlaced(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	release := seedRelease(ctx, t, queries)

	_, err := queries.LocateRelease(ctx, release.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("LocateRelease() error = %v, want %v", err, pgx.ErrNoRows)
	}
}

func TestQueries_RemovePlacement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	release := seedRelease(ctx, t, queries)

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseLounge, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	shelf, err := queries.CreateShelf(ctx, db.CreateShelfParams{BookcaseID: bookcase.ID, Position: 0})
	if err != nil {
		t.Fatalf("CreateShelf() error = %v", err)
	}

	_, err = queries.PlaceRelease(ctx, db.PlaceReleaseParams{ReleaseID: release.ID, ShelfID: shelf.ID, Position: 0})
	if err != nil {
		t.Fatalf("PlaceRelease() error = %v", err)
	}

	err = queries.RemovePlacement(ctx, release.ID)
	if err != nil {
		t.Fatalf("RemovePlacement() error = %v", err)
	}

	_, err = queries.LocateRelease(ctx, release.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("LocateRelease() after remove error = %v, want %v", err, pgx.ErrNoRows)
	}
}
