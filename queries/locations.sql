-- Bookcases -----------------------------------------------------------------

-- name: CreateBookcase :one
INSERT INTO bookcases (name, position)
VALUES ($1, $2)
RETURNING *;

-- name: GetBookcase :one
SELECT * FROM bookcases
WHERE id = $1 LIMIT 1;

-- name: ListBookcases :many
SELECT * FROM bookcases
ORDER BY position, created_at;

-- name: UpdateBookcase :one
UPDATE bookcases
SET name = $2, position = $3
WHERE id = $1
RETURNING *;

-- name: DeleteBookcase :exec
DELETE FROM bookcases
WHERE id = $1;

-- Shelves -------------------------------------------------------------------

-- name: CreateShelf :one
INSERT INTO shelves (bookcase_id, position)
VALUES ($1, $2)
RETURNING *;

-- name: GetShelf :one
SELECT * FROM shelves
WHERE id = $1 LIMIT 1;

-- name: ListShelvesByBookcase :many
SELECT * FROM shelves
WHERE bookcase_id = $1
ORDER BY position, created_at;

-- name: UpdateShelf :one
UPDATE shelves
SET position = $2
WHERE id = $1
RETURNING *;

-- name: DeleteShelf :exec
DELETE FROM shelves
WHERE id = $1;

-- Placements ----------------------------------------------------------------

-- name: PlaceRelease :one
INSERT INTO placements (release_id, shelf_id, position)
VALUES ($1, $2, $3)
ON CONFLICT (release_id)
DO UPDATE SET shelf_id = EXCLUDED.shelf_id, position = EXCLUDED.position
RETURNING *;

-- name: ListPlacementsByShelf :many
SELECT * FROM placements
WHERE shelf_id = $1
ORDER BY position, created_at;

-- name: RemovePlacement :exec
DELETE FROM placements
WHERE release_id = $1;

-- name: LocateRelease :one
SELECT
    sqlc.embed(b),
    sqlc.embed(s),
    sqlc.embed(p)
FROM placements p
JOIN shelves s ON s.id = p.shelf_id
JOIN bookcases b ON b.id = s.bookcase_id
WHERE p.release_id = $1;
