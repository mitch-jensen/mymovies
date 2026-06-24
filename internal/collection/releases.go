package collection

import (
	"context"

	"github.com/google/uuid"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

// ListReleasesByMovie returns the home video releases of a movie.
func (c *Collection) ListReleasesByMovie(ctx context.Context, movieID uuid.UUID) ([]db.HomeVideoRelease, error) {
	releases, err := c.q.ListHomeVideoReleasesByMovie(ctx, movieID)
	if err != nil {
		return nil, wrap("list releases", err)
	}

	return releases, nil
}

// CreateRelease inserts a home video release and returns it.
func (c *Collection) CreateRelease(
	ctx context.Context, arg db.CreateHomeVideoReleaseParams,
) (db.HomeVideoRelease, error) {
	release, err := c.q.CreateHomeVideoRelease(ctx, arg)
	if err != nil {
		return db.HomeVideoRelease{}, wrap("create release", err)
	}

	return release, nil
}

// GetRelease returns the release with the given ID, or ErrNotFound if none exists.
func (c *Collection) GetRelease(ctx context.Context, id uuid.UUID) (db.HomeVideoRelease, error) {
	release, err := c.q.GetHomeVideoRelease(ctx, id)
	if err != nil {
		return db.HomeVideoRelease{}, notFound("get release", err)
	}

	return release, nil
}

// UpdateRelease updates a release and returns it, or ErrNotFound if none exists.
func (c *Collection) UpdateRelease(
	ctx context.Context, arg db.UpdateHomeVideoReleaseParams,
) (db.HomeVideoRelease, error) {
	release, err := c.q.UpdateHomeVideoRelease(ctx, arg)
	if err != nil {
		return db.HomeVideoRelease{}, notFound("update release", err)
	}

	return release, nil
}

// DeleteRelease removes a release by ID.
func (c *Collection) DeleteRelease(ctx context.Context, id uuid.UUID) error {
	err := c.q.DeleteHomeVideoRelease(ctx, id)
	if err != nil {
		return wrap("delete release", err)
	}

	return nil
}
