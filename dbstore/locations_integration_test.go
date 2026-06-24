package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	db "github.com/mitch-jensen/mymovies/dbstore"
	"github.com/mitch-jensen/mymovies/internal/testdb"
)

const (
	bookcaseLounge = "Lounge"
	bookcaseStudy  = "Study"
)

// seedReleaseOf creates a bare release of the given movie, returning the release.
func seedReleaseOf(ctx context.Context, t *testing.T, queries *db.Queries, movieID uuid.UUID) db.HomeVideoRelease {
	t.Helper()

	release, err := queries.CreateHomeVideoRelease(ctx, db.CreateHomeVideoReleaseParams{
		MovieID: movieID, DigitalCopy: false, Watched: false,
	})
	if err != nil {
		t.Fatalf("CreateHomeVideoRelease() error = %v", err)
	}

	return release
}

// seedRelease creates a movie and a bare release of it, returning the release.
func seedRelease(ctx context.Context, t *testing.T, queries *db.Queries) db.HomeVideoRelease {
	t.Helper()

	movie := seedMovie(ctx, t, queries)

	return seedReleaseOf(ctx, t, queries, movie.ID)
}

// seedShelf creates a bookcase and one shelf on it, returning the shelf.
func seedShelf(ctx context.Context, t *testing.T, queries *db.Queries, bookcaseName string) db.Shelf {
	t.Helper()

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseName, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	shelf, err := queries.CreateShelf(ctx, db.CreateShelfParams{BookcaseID: bookcase.ID, Position: 0})
	if err != nil {
		t.Fatalf("CreateShelf() error = %v", err)
	}

	return shelf
}

// mustPlace places a release on a shelf at a position, failing the test on error.
func mustPlace(ctx context.Context, t *testing.T, queries *db.Queries, releaseID, shelfID uuid.UUID, position int32) {
	t.Helper()

	_, err := queries.PlaceRelease(ctx, db.PlaceReleaseParams{ReleaseID: releaseID, ShelfID: shelfID, Position: position})
	if err != nil {
		t.Fatalf("PlaceRelease() error = %v", err)
	}
}

func TestQueries_ListPlacementsByShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	movie := seedMovie(ctx, t, queries)
	shelf := seedShelf(ctx, t, queries, bookcaseLounge)

	// Two releases of the movie, placed out of slot order to prove ORDER BY.
	first := seedReleaseOf(ctx, t, queries, movie.ID)
	second := seedReleaseOf(ctx, t, queries, movie.ID)
	mustPlace(ctx, t, queries, first.ID, shelf.ID, 1)
	mustPlace(ctx, t, queries, second.ID, shelf.ID, 0)

	rows, err := queries.ListPlacementsByShelf(ctx, shelf.ID)
	if err != nil {
		t.Fatalf("ListPlacementsByShelf() error = %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("got %d placements, want 2", len(rows))
	}

	// Ordered by position: the release placed at slot 0 comes first.
	if rows[0].HomeVideoRelease.ID != second.ID {
		t.Errorf("rows[0] release = %v, want %v (slot 0 first)", rows[0].HomeVideoRelease.ID, second.ID)
	}

	if rows[0].Movie.Title != movie.Title {
		t.Errorf("rows[0] movie title = %q, want %q", rows[0].Movie.Title, movie.Title)
	}
}

func TestQueries_ListShelvesByBookcaseOrdered(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseStudy, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	// Insert out of order to prove ListShelvesByBookcase orders by position.
	for _, pos := range []int32{2, 0, 1} {
		_, err = queries.CreateShelf(ctx, db.CreateShelfParams{BookcaseID: bookcase.ID, Position: pos})
		if err != nil {
			t.Fatalf("CreateShelf() error = %v", err)
		}
	}

	shelves, err := queries.ListShelvesByBookcase(ctx, bookcase.ID)
	if err != nil {
		t.Fatalf("ListShelvesByBookcase() error = %v", err)
	}

	if len(shelves) != 3 {
		t.Fatalf("got %d shelves, want 3", len(shelves))
	}

	for i, want := range []int32{0, 1, 2} {
		if shelves[i].Position != want {
			t.Errorf("shelves[%d].Position = %d, want %d", i, shelves[i].Position, want)
		}
	}
}

func TestQueries_UpdateBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseLounge, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	updated, err := queries.UpdateBookcase(ctx, db.UpdateBookcaseParams{ID: bookcase.ID, Name: bookcaseStudy, Position: 5})
	if err != nil {
		t.Fatalf("UpdateBookcase() error = %v", err)
	}

	if updated.Name != bookcaseStudy || updated.Position != 5 {
		t.Errorf("UpdateBookcase() = {%q, %d}, want {%q, 5}", updated.Name, updated.Position, bookcaseStudy)
	}
}

func TestQueries_DeleteBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	bookcase, err := queries.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseLounge, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase() error = %v", err)
	}

	err = queries.DeleteBookcase(ctx, bookcase.ID)
	if err != nil {
		t.Fatalf("DeleteBookcase() error = %v", err)
	}

	_, err = queries.GetBookcase(ctx, bookcase.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetBookcase() after delete error = %v, want %v", err, pgx.ErrNoRows)
	}
}

func TestQueries_UpdateShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	shelf := seedShelf(ctx, t, queries, bookcaseLounge)

	updated, err := queries.UpdateShelf(ctx, db.UpdateShelfParams{ID: shelf.ID, Position: 9})
	if err != nil {
		t.Fatalf("UpdateShelf() error = %v", err)
	}

	if updated.Position != 9 {
		t.Errorf("UpdateShelf().Position = %d, want 9", updated.Position)
	}
}

func TestQueries_DeleteShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	queries := db.New(testdb.Setup(ctx, t))

	shelf := seedShelf(ctx, t, queries, bookcaseLounge)

	err := queries.DeleteShelf(ctx, shelf.ID)
	if err != nil {
		t.Fatalf("DeleteShelf() error = %v", err)
	}

	_, err = queries.GetShelf(ctx, shelf.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetShelf() after delete error = %v, want %v", err, pgx.ErrNoRows)
	}
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
