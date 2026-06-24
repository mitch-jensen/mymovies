package api

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/mitch-jensen/mymovies/internal/collection"
)

// mapErr translates a collection error into the appropriate HTTP status. It is
// the single place the API decides how domain errors surface to clients, so no
// handler needs to know about pgx or repeat not-found branching.
func mapErr(err error) error {
	if errors.Is(err, collection.ErrNotFound) {
		return huma.Error404NotFound("not found")
	}

	return huma.Error500InternalServerError("internal server error", err)
}
