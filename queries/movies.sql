-- name: GetMovie :one
SELECT * FROM movies
WHERE id = $1 LIMIT 1;

-- name: ListMovies :many
SELECT * FROM movies
ORDER BY title;

-- name: CreateMovie :one
INSERT INTO movies (
    title, release_year, runtime_min
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateMovie :exec
UPDATE movies
SET
    title = $2,
    release_year = $3,
    runtime_min = $4
WHERE id = $1;

-- name: DeleteMovie :exec
DELETE FROM movies
WHERE id = $1;
