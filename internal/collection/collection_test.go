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

func TestCreateReleaseReturnsErrNotFoundForMissingMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	_, err := col.CreateRelease(ctx, db.CreateHomeVideoReleaseParams{MovieID: uuid.New()})
	if !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("CreateRelease(missing movie) error = %v, want collection.ErrNotFound", err)
	}
}

func TestAddShelfReturnsErrNotFoundForMissingBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	_, err := col.AddShelf(ctx, uuid.New(), 0)
	if !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("AddShelf(missing bookcase) error = %v, want collection.ErrNotFound", err)
	}
}

func TestPlaceReleaseReturnsErrNotFoundForMissingRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	_, err := col.PlaceRelease(ctx, uuid.New(), uuid.New(), 0)
	if !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("PlaceRelease(missing release) error = %v, want collection.ErrNotFound", err)
	}
}

func TestPlaceReleaseReturnsErrNotFoundForMissingShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	movie, err := col.CreateMovie(ctx, db.CreateMovieParams{Title: "Ran", ReleaseYear: 1985})
	if err != nil {
		t.Fatalf("CreateMovie: %v", err)
	}

	release, err := col.CreateRelease(ctx, db.CreateHomeVideoReleaseParams{MovieID: movie.ID})
	if err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	_, err = col.PlaceRelease(ctx, release.ID, uuid.New(), 0)
	if !errors.Is(err, collection.ErrNotFound) {
		t.Errorf("PlaceRelease(missing shelf) error = %v, want collection.ErrNotFound", err)
	}
}

// placeReleaseInBookcase creates a movie with one release and places it on a new
// shelf of a freshly created bookcase, returning the movie and bookcase name.
func placeReleaseInBookcase(
	ctx context.Context, t *testing.T, col *collection.Collection, title, bookcaseName string,
) (db.Movie, string) {
	t.Helper()

	movie, err := col.CreateMovie(ctx, db.CreateMovieParams{Title: title, ReleaseYear: 1988})
	if err != nil {
		t.Fatalf("CreateMovie: %v", err)
	}

	release, err := col.CreateRelease(ctx, db.CreateHomeVideoReleaseParams{MovieID: movie.ID})
	if err != nil {
		t.Fatalf("CreateRelease: %v", err)
	}

	bookcase, err := col.CreateBookcase(ctx, db.CreateBookcaseParams{Name: bookcaseName, Position: 0})
	if err != nil {
		t.Fatalf("CreateBookcase: %v", err)
	}

	shelf, err := col.AddShelf(ctx, bookcase.ID, 0)
	if err != nil {
		t.Fatalf("AddShelf: %v", err)
	}

	_, err = col.PlaceRelease(ctx, release.ID, shelf.ID, 0)
	if err != nil {
		t.Fatalf("PlaceRelease: %v", err)
	}

	return movie, bookcaseName
}

func TestSearchGroupsPlacedReleasesByMovie(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	movie, bookcaseName := placeReleaseInBookcase(ctx, t, col, "Akira", "Hallway")

	results, err := col.Search(ctx, "akira", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	if results[0].Movie.ID != movie.ID {
		t.Errorf("result movie ID = %v, want %v", results[0].Movie.ID, movie.ID)
	}

	if len(results[0].Releases) != 1 {
		t.Fatalf("len(located releases) = %d, want 1", len(results[0].Releases))
	}

	if results[0].Releases[0].Bookcase.Name != bookcaseName {
		t.Errorf("located bookcase = %q, want %q", results[0].Releases[0].Bookcase.Name, bookcaseName)
	}
}

func TestSearchBlankQueryReturnsNoResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	col := newCollection(ctx, t)

	results, err := col.Search(ctx, "   ", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}
