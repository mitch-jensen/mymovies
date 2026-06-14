package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

const (
	bookcasesPath       = "/bookcases"
	bookcaseByIDPath    = "/bookcases/{id}"
	bookcaseShelvesPath = "/bookcases/{bookcaseId}/shelves"
	shelfByIDPath       = "/shelves/{id}"
	placementPath       = "/releases/{id}/placement"
	locationPath        = "/releases/{id}/location"
)

// Bookcase is the API representation of a bookcase.
type Bookcase struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Position  int32     `json:"position"`
	CreatedAt time.Time `json:"createdAt"`
}

func bookcaseFromDB(bookcase db.Bookcase) Bookcase {
	return Bookcase{
		ID:        bookcase.ID,
		Name:      bookcase.Name,
		Position:  bookcase.Position,
		CreatedAt: bookcase.CreatedAt,
	}
}

// BookcaseFields holds the editable fields of a bookcase.
type BookcaseFields struct {
	Name     string `json:"name"`
	Position int32  `json:"position"`
}

// Shelf is the API representation of a shelf within a bookcase.
type Shelf struct {
	ID         uuid.UUID `json:"id"`
	BookcaseID uuid.UUID `json:"bookcaseId"`
	Position   int32     `json:"position"`
	CreatedAt  time.Time `json:"createdAt"`
}

func shelfFromDB(shelf db.Shelf) Shelf {
	return Shelf{
		ID:         shelf.ID,
		BookcaseID: shelf.BookcaseID,
		Position:   shelf.Position,
		CreatedAt:  shelf.CreatedAt,
	}
}

// ShelfFields holds the editable fields of a shelf.
type ShelfFields struct {
	Position int32 `json:"position"`
}

// Placement is the API representation of a release's spot on a shelf.
type Placement struct {
	ID        uuid.UUID `json:"id"`
	ReleaseID uuid.UUID `json:"releaseId"`
	ShelfID   uuid.UUID `json:"shelfId"`
	Position  int32     `json:"position"`
	CreatedAt time.Time `json:"createdAt"`
}

func placementFromDB(placement db.Placement) Placement {
	return Placement{
		ID:        placement.ID,
		ReleaseID: placement.ReleaseID,
		ShelfID:   placement.ShelfID,
		Position:  placement.Position,
		CreatedAt: placement.CreatedAt,
	}
}

// PlacementFields specifies where to place a release.
type PlacementFields struct {
	ShelfID  uuid.UUID `json:"shelfId"`
	Position int32     `json:"position"`
}

// Location is where a release physically lives: which shelf of which bookcase.
type Location struct {
	Bookcase  Bookcase  `json:"bookcase"`
	Shelf     Shelf     `json:"shelf"`
	Placement Placement `json:"placement"`
}

func (s *Server) registerBookcaseRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "list-bookcases",
		Method:      http.MethodGet,
		Path:        bookcasesPath,
		Summary:     "List all bookcases",
	}, s.listBookcases)

	huma.Register(s.api, huma.Operation{
		OperationID:   "create-bookcase",
		Method:        http.MethodPost,
		Path:          bookcasesPath,
		Summary:       "Create a bookcase",
		DefaultStatus: http.StatusCreated,
	}, s.createBookcase)

	huma.Register(s.api, huma.Operation{
		OperationID: "get-bookcase",
		Method:      http.MethodGet,
		Path:        bookcaseByIDPath,
		Summary:     "Get a bookcase by ID",
	}, s.getBookcase)

	huma.Register(s.api, huma.Operation{
		OperationID: "update-bookcase",
		Method:      http.MethodPut,
		Path:        bookcaseByIDPath,
		Summary:     "Update a bookcase",
	}, s.updateBookcase)

	huma.Register(s.api, huma.Operation{
		OperationID:   "delete-bookcase",
		Method:        http.MethodDelete,
		Path:          bookcaseByIDPath,
		Summary:       "Delete a bookcase",
		DefaultStatus: http.StatusNoContent,
	}, s.deleteBookcase)
}

func (s *Server) registerShelfRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "list-bookcase-shelves",
		Method:      http.MethodGet,
		Path:        bookcaseShelvesPath,
		Summary:     "List the shelves of a bookcase",
	}, s.listBookcaseShelves)

	huma.Register(s.api, huma.Operation{
		OperationID:   "create-shelf",
		Method:        http.MethodPost,
		Path:          bookcaseShelvesPath,
		Summary:       "Add a shelf to a bookcase",
		DefaultStatus: http.StatusCreated,
	}, s.createShelf)

	huma.Register(s.api, huma.Operation{
		OperationID: "update-shelf",
		Method:      http.MethodPut,
		Path:        shelfByIDPath,
		Summary:     "Update a shelf",
	}, s.updateShelf)

	huma.Register(s.api, huma.Operation{
		OperationID:   "delete-shelf",
		Method:        http.MethodDelete,
		Path:          shelfByIDPath,
		Summary:       "Delete a shelf",
		DefaultStatus: http.StatusNoContent,
	}, s.deleteShelf)
}

func (s *Server) registerPlacementRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "place-release",
		Method:      http.MethodPut,
		Path:        placementPath,
		Summary:     "Place or move a release onto a shelf",
	}, s.placeRelease)

	huma.Register(s.api, huma.Operation{
		OperationID:   "remove-placement",
		Method:        http.MethodDelete,
		Path:          placementPath,
		Summary:       "Remove a release from its shelf",
		DefaultStatus: http.StatusNoContent,
	}, s.removePlacement)

	huma.Register(s.api, huma.Operation{
		OperationID: "locate-release",
		Method:      http.MethodGet,
		Path:        locationPath,
		Summary:     "Find where a release physically lives",
	}, s.locateRelease)
}

// BookcaseInput is a request carrying bookcase field values in its body.
type BookcaseInput struct {
	Body BookcaseFields
}

// BookcaseIDInput identifies a bookcase by its path parameter.
type BookcaseIDInput struct {
	ID uuid.UUID `doc:"Bookcase ID" path:"id"`
}

// UpdateBookcaseInput carries the bookcase ID in the path and new values in the body.
type UpdateBookcaseInput struct {
	ID   uuid.UUID `doc:"Bookcase ID" path:"id"`
	Body BookcaseFields
}

// BookcaseOutput is a response carrying a single bookcase.
type BookcaseOutput struct {
	Body Bookcase
}

// ListBookcasesOutput is the response body for listing bookcases.
type ListBookcasesOutput struct {
	Body []Bookcase
}

func (s *Server) listBookcases(ctx context.Context, _ *struct{}) (*ListBookcasesOutput, error) {
	bookcases, err := s.queries.ListBookcases(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list bookcases", err)
	}

	body := make([]Bookcase, len(bookcases))
	for i, bookcase := range bookcases {
		body[i] = bookcaseFromDB(bookcase)
	}

	return &ListBookcasesOutput{Body: body}, nil
}

func (s *Server) createBookcase(ctx context.Context, input *BookcaseInput) (*BookcaseOutput, error) {
	bookcase, err := s.queries.CreateBookcase(ctx, db.CreateBookcaseParams{
		Name:     input.Body.Name,
		Position: input.Body.Position,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create bookcase", err)
	}

	return &BookcaseOutput{Body: bookcaseFromDB(bookcase)}, nil
}

func (s *Server) getBookcase(ctx context.Context, input *BookcaseIDInput) (*BookcaseOutput, error) {
	bookcase, err := s.queries.GetBookcase(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("bookcase not found")
		}

		return nil, huma.Error500InternalServerError("failed to get bookcase", err)
	}

	return &BookcaseOutput{Body: bookcaseFromDB(bookcase)}, nil
}

func (s *Server) updateBookcase(ctx context.Context, input *UpdateBookcaseInput) (*BookcaseOutput, error) {
	bookcase, err := s.queries.UpdateBookcase(ctx, db.UpdateBookcaseParams{
		ID:       input.ID,
		Name:     input.Body.Name,
		Position: input.Body.Position,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("bookcase not found")
		}

		return nil, huma.Error500InternalServerError("failed to update bookcase", err)
	}

	return &BookcaseOutput{Body: bookcaseFromDB(bookcase)}, nil
}

func (s *Server) deleteBookcase(ctx context.Context, input *BookcaseIDInput) (*struct{}, error) {
	err := s.queries.DeleteBookcase(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to delete bookcase", err)
	}

	return nil, nil //nolint:nilnil // 204 No Content: no body and no error.
}

// ShelvesInput identifies the bookcase whose shelves are addressed.
type ShelvesInput struct {
	BookcaseID uuid.UUID `doc:"Bookcase ID" path:"bookcaseId"`
}

// CreateShelfInput carries the owning bookcase ID in the path and shelf fields in the body.
type CreateShelfInput struct {
	BookcaseID uuid.UUID `doc:"Bookcase ID" path:"bookcaseId"`
	Body       ShelfFields
}

// ShelfIDInput identifies a shelf by its path parameter.
type ShelfIDInput struct {
	ID uuid.UUID `doc:"Shelf ID" path:"id"`
}

// UpdateShelfInput carries the shelf ID in the path and new values in the body.
type UpdateShelfInput struct {
	ID   uuid.UUID `doc:"Shelf ID" path:"id"`
	Body ShelfFields
}

// ShelfOutput is a response carrying a single shelf.
type ShelfOutput struct {
	Body Shelf
}

// ListShelvesOutput is the response body for listing a bookcase's shelves.
type ListShelvesOutput struct {
	Body []Shelf
}

func (s *Server) listBookcaseShelves(ctx context.Context, input *ShelvesInput) (*ListShelvesOutput, error) {
	shelves, err := s.queries.ListShelvesByBookcase(ctx, input.BookcaseID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list shelves", err)
	}

	body := make([]Shelf, len(shelves))
	for i, shelf := range shelves {
		body[i] = shelfFromDB(shelf)
	}

	return &ListShelvesOutput{Body: body}, nil
}

func (s *Server) createShelf(ctx context.Context, input *CreateShelfInput) (*ShelfOutput, error) {
	// Confirm the bookcase exists so a missing one yields 404, not a raw FK error.
	_, err := s.queries.GetBookcase(ctx, input.BookcaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("bookcase not found")
		}

		return nil, huma.Error500InternalServerError("failed to look up bookcase", err)
	}

	shelf, err := s.queries.CreateShelf(ctx, db.CreateShelfParams{
		BookcaseID: input.BookcaseID,
		Position:   input.Body.Position,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create shelf", err)
	}

	return &ShelfOutput{Body: shelfFromDB(shelf)}, nil
}

func (s *Server) updateShelf(ctx context.Context, input *UpdateShelfInput) (*ShelfOutput, error) {
	shelf, err := s.queries.UpdateShelf(ctx, db.UpdateShelfParams{
		ID:       input.ID,
		Position: input.Body.Position,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("shelf not found")
		}

		return nil, huma.Error500InternalServerError("failed to update shelf", err)
	}

	return &ShelfOutput{Body: shelfFromDB(shelf)}, nil
}

func (s *Server) deleteShelf(ctx context.Context, input *ShelfIDInput) (*struct{}, error) {
	err := s.queries.DeleteShelf(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to delete shelf", err)
	}

	return nil, nil //nolint:nilnil // 204 No Content: no body and no error.
}

// PlaceReleaseInput carries the release ID in the path and the target shelf and
// position in the body.
type PlaceReleaseInput struct {
	ID   uuid.UUID `doc:"Release ID" path:"id"`
	Body PlacementFields
}

// PlacementOutput is a response carrying a single placement.
type PlacementOutput struct {
	Body Placement
}

func (s *Server) placeRelease(ctx context.Context, input *PlaceReleaseInput) (*PlacementOutput, error) {
	_, err := s.queries.GetHomeVideoRelease(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("release not found")
		}

		return nil, huma.Error500InternalServerError("failed to look up release", err)
	}

	_, err = s.queries.GetShelf(ctx, input.Body.ShelfID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("shelf not found")
		}

		return nil, huma.Error500InternalServerError("failed to look up shelf", err)
	}

	placement, err := s.queries.PlaceRelease(ctx, db.PlaceReleaseParams{
		ReleaseID: input.ID,
		ShelfID:   input.Body.ShelfID,
		Position:  input.Body.Position,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to place release", err)
	}

	return &PlacementOutput{Body: placementFromDB(placement)}, nil
}

func (s *Server) removePlacement(ctx context.Context, input *ReleaseIDInput) (*struct{}, error) {
	err := s.queries.RemovePlacement(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to remove placement", err)
	}

	return nil, nil //nolint:nilnil // 204 No Content: no body and no error.
}

// LocationOutput is a response carrying a release's physical location.
type LocationOutput struct {
	Body Location
}

func (s *Server) locateRelease(ctx context.Context, input *ReleaseIDInput) (*LocationOutput, error) {
	row, err := s.queries.LocateRelease(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("release is not placed anywhere")
		}

		return nil, huma.Error500InternalServerError("failed to locate release", err)
	}

	return &LocationOutput{Body: Location{
		Bookcase:  bookcaseFromDB(row.Bookcase),
		Shelf:     shelfFromDB(row.Shelf),
		Placement: placementFromDB(row.Placement),
	}}, nil
}
