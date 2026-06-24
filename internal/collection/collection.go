// Package collection is the domain module for the movie collection. It owns data
// access and multi-step workflows, translating storage errors (notably
// pgx.ErrNoRows) into domain errors so callers never depend on the database
// driver. It is the seam the HTTP layer talks to instead of *dbstore.Queries.
package collection

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

// ErrNotFound is returned when a requested entity does not exist. It replaces
// pgx.ErrNoRows leaking out of the data layer; the HTTP layer maps it to 404.
var ErrNotFound = errors.New("not found")

// Collection provides access to the movie collection and its physical layout.
type Collection struct {
	q *db.Queries
}

// New builds a Collection backed by the given database pool.
func New(pool *pgxpool.Pool) *Collection {
	return &Collection{q: db.New(pool)}
}

// notFound translates pgx.ErrNoRows into ErrNotFound and wraps any other error
// with the given operation name for context.
func notFound(op string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return fmt.Errorf("%s: %w", op, err)
}

// wrap annotates a non-nil error with the operation name.
func wrap(op string, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
