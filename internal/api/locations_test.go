package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/mitch-jensen/mymovies/internal/api"
)

const (
	bookcasesBase     = "/bookcases"
	testBookcaseName  = "Lounge"
	testBookcaseAltNm = "Study"
)

func bookcasePath(id uuid.UUID) string {
	return bookcasesBase + "/" + id.String()
}

func bookcaseShelvesPath(id uuid.UUID) string {
	return bookcasePath(id) + "/shelves"
}

func shelfPath(id uuid.UUID) string {
	return "/shelves/" + id.String()
}

func shelfPlacementsPath(id uuid.UUID) string {
	return shelfPath(id) + "/placements"
}

func placementPath(releaseID uuid.UUID) string {
	return releasePath(releaseID) + "/placement"
}

func locationPath(releaseID uuid.UUID) string {
	return releasePath(releaseID) + "/location"
}

func createBookcase(ctx context.Context, t *testing.T, handler http.Handler, fields api.BookcaseFields) api.Bookcase {
	t.Helper()

	recorder := doRequest(ctx, t, handler, http.MethodPost, bookcasesBase, fields)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST bookcase status = %d, want %d (body: %s)", recorder.Code, http.StatusCreated, recorder.Body)
	}

	var bookcase api.Bookcase

	err := json.Unmarshal(recorder.Body.Bytes(), &bookcase)
	if err != nil {
		t.Fatalf("decode bookcase: %v", err)
	}

	return bookcase
}

func createShelf(
	ctx context.Context, t *testing.T, handler http.Handler, bookcaseID uuid.UUID, position int32,
) api.Shelf {
	t.Helper()

	recorder := doRequest(
		ctx, t, handler, http.MethodPost, bookcaseShelvesPath(bookcaseID), api.ShelfFields{Position: position},
	)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST shelf status = %d, want %d (body: %s)", recorder.Code, http.StatusCreated, recorder.Body)
	}

	var shelf api.Shelf

	err := json.Unmarshal(recorder.Body.Bytes(), &shelf)
	if err != nil {
		t.Fatalf("decode shelf: %v", err)
	}

	return shelf
}

// seedPlacedRelease creates a movie, a release, a bookcase, a shelf, places the
// release on the shelf, and returns the release and shelf.
func seedPlacedRelease(ctx context.Context, t *testing.T, handler http.Handler) (api.Release, api.Shelf) {
	t.Helper()

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Akira", ReleaseYear: 1988})
	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})
	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 2)

	recorder := doRequest(ctx, t, handler, http.MethodPut, placementPath(release.ID), api.PlacementFields{
		ShelfID:  shelf.ID,
		Position: 5,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("PUT placement status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	return release, shelf
}

func TestCreateBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 1})

	if bookcase.ID == uuid.Nil {
		t.Error("created bookcase has nil ID")
	}

	if bookcase.Name != testBookcaseName {
		t.Errorf("Name = %q, want %q", bookcase.Name, testBookcaseName)
	}
}

func TestCreateShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 3)

	if shelf.BookcaseID != bookcase.ID {
		t.Errorf("BookcaseID = %v, want %v", shelf.BookcaseID, bookcase.ID)
	}

	if shelf.Position != 3 {
		t.Errorf("Position = %d, want 3", shelf.Position)
	}
}

func TestCreateShelfBookcaseNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodPost, bookcaseShelvesPath(uuid.New()), api.ShelfFields{Position: 0})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("POST shelf status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestPlaceAndLocateRelease(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	release, shelf := seedPlacedRelease(ctx, t, handler)

	recorder := doRequest(ctx, t, handler, http.MethodGet, locationPath(release.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET location status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var location api.Location

	err := json.Unmarshal(recorder.Body.Bytes(), &location)
	if err != nil {
		t.Fatalf("decode location: %v", err)
	}

	if location.Shelf.ID != shelf.ID {
		t.Errorf("Shelf.ID = %v, want %v", location.Shelf.ID, shelf.ID)
	}

	if location.Bookcase.Name != testBookcaseName {
		t.Errorf("Bookcase.Name = %q, want %q", location.Bookcase.Name, testBookcaseName)
	}

	if location.Placement.Position != 5 {
		t.Errorf("Placement.Position = %d, want 5", location.Placement.Position)
	}
}

func TestLocateReleaseNotPlaced(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Solaris", ReleaseYear: 1972})
	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})

	recorder := doRequest(ctx, t, handler, http.MethodGet, locationPath(release.ID), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET location status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestPlaceReleaseReleaseNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 0)

	recorder := doRequest(ctx, t, handler, http.MethodPut, placementPath(uuid.New()), api.PlacementFields{
		ShelfID:  shelf.ID,
		Position: 0,
	})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("PUT placement status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestPlaceReleaseShelfNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	movie := createMovie(ctx, t, handler, api.MovieFields{Title: "Videodrome", ReleaseYear: 1983})
	release := createRelease(ctx, t, handler, movie.ID, api.ReleaseFields{DigitalCopy: false, Watched: false})

	recorder := doRequest(ctx, t, handler, http.MethodPut, placementPath(release.ID), api.PlacementFields{
		ShelfID:  uuid.New(),
		Position: 0,
	})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("PUT placement status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestRemovePlacement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	release, _ := seedPlacedRelease(ctx, t, handler)

	recorder := doRequest(ctx, t, handler, http.MethodDelete, placementPath(release.ID), nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE placement status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	recorder = doRequest(ctx, t, handler, http.MethodGet, locationPath(release.ID), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET location after remove status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestListShelfPlacements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	release, shelf := seedPlacedRelease(ctx, t, handler)

	recorder := doRequest(ctx, t, handler, http.MethodGet, shelfPlacementsPath(shelf.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET shelf placements status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	var contents []api.ShelfPlacement

	err := json.Unmarshal(recorder.Body.Bytes(), &contents)
	if err != nil {
		t.Fatalf("decode shelf placements: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("len(contents) = %d, want 1", len(contents))
	}

	if contents[0].Release.ID != release.ID {
		t.Errorf("Release.ID = %v, want %v", contents[0].Release.ID, release.ID)
	}

	if contents[0].Movie.Title != "Akira" {
		t.Errorf("Movie.Title = %q, want %q", contents[0].Movie.Title, "Akira")
	}

	if contents[0].Placement.Position != 5 {
		t.Errorf("Placement.Position = %d, want 5", contents[0].Placement.Position)
	}
}

func TestListShelfPlacementsEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 0)

	recorder := doRequest(ctx, t, handler, http.MethodGet, shelfPlacementsPath(shelf.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET shelf placements status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var contents []api.ShelfPlacement

	err := json.Unmarshal(recorder.Body.Bytes(), &contents)
	if err != nil {
		t.Fatalf("decode shelf placements: %v", err)
	}

	if len(contents) != 0 {
		t.Errorf("len(contents) = %d, want 0", len(contents))
	}
}

func TestListBookcases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseAltNm, Position: 1})
	createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})

	recorder := doRequest(ctx, t, handler, http.MethodGet, bookcasesBase, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET bookcases status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var bookcases []api.Bookcase

	err := json.Unmarshal(recorder.Body.Bytes(), &bookcases)
	if err != nil {
		t.Fatalf("decode bookcases: %v", err)
	}

	if len(bookcases) != 2 {
		t.Fatalf("len(bookcases) = %d, want 2", len(bookcases))
	}

	// Ordered by position: the position-0 bookcase comes first.
	if bookcases[0].Name != testBookcaseName {
		t.Errorf("bookcases[0].Name = %q, want %q", bookcases[0].Name, testBookcaseName)
	}
}

func TestGetBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	created := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})

	recorder := doRequest(ctx, t, handler, http.MethodGet, bookcasePath(created.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET bookcase status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var got api.Bookcase

	err := json.Unmarshal(recorder.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("decode bookcase: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %v, want %v", got.ID, created.ID)
	}
}

func TestGetBookcaseNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodGet, bookcasePath(uuid.New()), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET bookcase status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestUpdateBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	created := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})

	recorder := doRequest(ctx, t, handler, http.MethodPut, bookcasePath(created.ID), api.BookcaseFields{
		Name:     testBookcaseAltNm,
		Position: 4,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("PUT bookcase status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	var got api.Bookcase

	err := json.Unmarshal(recorder.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("decode bookcase: %v", err)
	}

	if got.Name != testBookcaseAltNm || got.Position != 4 {
		t.Errorf("updated bookcase = {%q, %d}, want {%q, 4}", got.Name, got.Position, testBookcaseAltNm)
	}
}

func TestUpdateBookcaseNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodPut, bookcasePath(uuid.New()), api.BookcaseFields{
		Name:     "Ghost",
		Position: 0,
	})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("PUT bookcase status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeleteBookcase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	created := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})

	recorder := doRequest(ctx, t, handler, http.MethodDelete, bookcasePath(created.ID), nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE bookcase status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	recorder = doRequest(ctx, t, handler, http.MethodGet, bookcasePath(created.ID), nil)
	if recorder.Code != http.StatusNotFound {
		t.Errorf("GET after delete status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestListBookcaseShelves(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	createShelf(ctx, t, handler, bookcase.ID, 1)
	createShelf(ctx, t, handler, bookcase.ID, 0)

	recorder := doRequest(ctx, t, handler, http.MethodGet, bookcaseShelvesPath(bookcase.ID), nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET shelves status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var shelves []api.Shelf

	err := json.Unmarshal(recorder.Body.Bytes(), &shelves)
	if err != nil {
		t.Fatalf("decode shelves: %v", err)
	}

	if len(shelves) != 2 {
		t.Fatalf("len(shelves) = %d, want 2", len(shelves))
	}

	// Ordered by position.
	if shelves[0].Position != 0 || shelves[1].Position != 1 {
		t.Errorf("positions = [%d, %d], want [0, 1]", shelves[0].Position, shelves[1].Position)
	}
}

func TestUpdateShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 0)

	recorder := doRequest(ctx, t, handler, http.MethodPut, shelfPath(shelf.ID), api.ShelfFields{Position: 7})
	if recorder.Code != http.StatusOK {
		t.Fatalf("PUT shelf status = %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body)
	}

	var got api.Shelf

	err := json.Unmarshal(recorder.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("decode shelf: %v", err)
	}

	if got.Position != 7 {
		t.Errorf("Position = %d, want 7", got.Position)
	}
}

func TestUpdateShelfNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	recorder := doRequest(ctx, t, handler, http.MethodPut, shelfPath(uuid.New()), api.ShelfFields{Position: 0})
	if recorder.Code != http.StatusNotFound {
		t.Errorf("PUT shelf status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDeleteShelf(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	handler := newHandler(ctx, t)

	bookcase := createBookcase(ctx, t, handler, api.BookcaseFields{Name: testBookcaseName, Position: 0})
	shelf := createShelf(ctx, t, handler, bookcase.ID, 0)

	recorder := doRequest(ctx, t, handler, http.MethodDelete, shelfPath(shelf.ID), nil)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("DELETE shelf status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	// The bookcase should now have no shelves.
	recorder = doRequest(ctx, t, handler, http.MethodGet, bookcaseShelvesPath(bookcase.ID), nil)

	var shelves []api.Shelf

	err := json.Unmarshal(recorder.Body.Bytes(), &shelves)
	if err != nil {
		t.Fatalf("decode shelves: %v", err)
	}

	if len(shelves) != 0 {
		t.Errorf("len(shelves) after delete = %d, want 0", len(shelves))
	}
}
