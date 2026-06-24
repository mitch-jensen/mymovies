package collection

import (
	"context"

	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

// ListBookcases returns all bookcases.
func (c *Collection) ListBookcases(ctx context.Context) ([]db.Bookcase, error) {
	bookcases, err := c.q.ListBookcases(ctx)
	if err != nil {
		return nil, wrap("list bookcases", err)
	}

	return bookcases, nil
}

// CreateBookcase inserts a bookcase and returns it.
func (c *Collection) CreateBookcase(ctx context.Context, arg db.CreateBookcaseParams) (db.Bookcase, error) {
	bookcase, err := c.q.CreateBookcase(ctx, arg)
	if err != nil {
		return db.Bookcase{}, wrap("create bookcase", err)
	}

	return bookcase, nil
}

// GetBookcase returns the bookcase with the given ID, or ErrNotFound if none exists.
func (c *Collection) GetBookcase(ctx context.Context, id uuid.UUID) (db.Bookcase, error) {
	bookcase, err := c.q.GetBookcase(ctx, id)
	if err != nil {
		return db.Bookcase{}, notFound("get bookcase", err)
	}

	return bookcase, nil
}

// UpdateBookcase updates a bookcase and returns it, or ErrNotFound if none exists.
func (c *Collection) UpdateBookcase(ctx context.Context, arg db.UpdateBookcaseParams) (db.Bookcase, error) {
	bookcase, err := c.q.UpdateBookcase(ctx, arg)
	if err != nil {
		return db.Bookcase{}, notFound("update bookcase", err)
	}

	return bookcase, nil
}

// DeleteBookcase removes a bookcase by ID.
func (c *Collection) DeleteBookcase(ctx context.Context, id uuid.UUID) error {
	err := c.q.DeleteBookcase(ctx, id)
	if err != nil {
		return wrap("delete bookcase", err)
	}

	return nil
}

// ListShelvesByBookcase returns the shelves of a bookcase.
func (c *Collection) ListShelvesByBookcase(ctx context.Context, bookcaseID uuid.UUID) ([]db.Shelf, error) {
	shelves, err := c.q.ListShelvesByBookcase(ctx, bookcaseID)
	if err != nil {
		return nil, wrap("list shelves", err)
	}

	return shelves, nil
}

// CreateShelf inserts a shelf and returns it.
func (c *Collection) CreateShelf(ctx context.Context, arg db.CreateShelfParams) (db.Shelf, error) {
	shelf, err := c.q.CreateShelf(ctx, arg)
	if err != nil {
		return db.Shelf{}, wrap("create shelf", err)
	}

	return shelf, nil
}

// GetShelf returns the shelf with the given ID, or ErrNotFound if none exists.
func (c *Collection) GetShelf(ctx context.Context, id uuid.UUID) (db.Shelf, error) {
	shelf, err := c.q.GetShelf(ctx, id)
	if err != nil {
		return db.Shelf{}, notFound("get shelf", err)
	}

	return shelf, nil
}

// UpdateShelf updates a shelf and returns it, or ErrNotFound if none exists.
func (c *Collection) UpdateShelf(ctx context.Context, arg db.UpdateShelfParams) (db.Shelf, error) {
	shelf, err := c.q.UpdateShelf(ctx, arg)
	if err != nil {
		return db.Shelf{}, notFound("update shelf", err)
	}

	return shelf, nil
}

// DeleteShelf removes a shelf by ID.
func (c *Collection) DeleteShelf(ctx context.Context, id uuid.UUID) error {
	err := c.q.DeleteShelf(ctx, id)
	if err != nil {
		return wrap("delete shelf", err)
	}

	return nil
}

// PlaceRelease places or moves a release onto a shelf and returns the placement.
func (c *Collection) PlaceRelease(ctx context.Context, arg db.PlaceReleaseParams) (db.Placement, error) {
	placement, err := c.q.PlaceRelease(ctx, arg)
	if err != nil {
		return db.Placement{}, wrap("place release", err)
	}

	return placement, nil
}

// RemovePlacement removes a release from its shelf.
func (c *Collection) RemovePlacement(ctx context.Context, releaseID uuid.UUID) error {
	err := c.q.RemovePlacement(ctx, releaseID)
	if err != nil {
		return wrap("remove placement", err)
	}

	return nil
}

// LocateRelease returns where a release physically lives, or ErrNotFound if it is
// not placed anywhere.
func (c *Collection) LocateRelease(ctx context.Context, releaseID uuid.UUID) (db.LocateReleaseRow, error) {
	row, err := c.q.LocateRelease(ctx, releaseID)
	if err != nil {
		return db.LocateReleaseRow{}, notFound("locate release", err)
	}

	return row, nil
}
