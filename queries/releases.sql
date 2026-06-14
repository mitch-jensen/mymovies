-- name: GetHomeVideoRelease :one
SELECT * FROM home_video_releases
WHERE id = $1 LIMIT 1;

-- name: ListHomeVideoReleasesByMovie :many
SELECT * FROM home_video_releases
WHERE movie_id = $1
ORDER BY created_at;

-- name: CreateHomeVideoRelease :one
INSERT INTO home_video_releases (
    movie_id,
    studio,
    country_code,
    upc,
    ean,
    asin,
    release_date,
    casing,
    slipcover,
    blu_ray_discs,
    dvd_discs,
    digital_copy,
    watched,
    comment,
    retailer,
    price,
    price_comment
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: UpdateHomeVideoRelease :one
UPDATE home_video_releases
SET
    studio = $2,
    country_code = $3,
    upc = $4,
    ean = $5,
    asin = $6,
    release_date = $7,
    casing = $8,
    slipcover = $9,
    blu_ray_discs = $10,
    dvd_discs = $11,
    digital_copy = $12,
    watched = $13,
    comment = $14,
    retailer = $15,
    price = $16,
    price_comment = $17
WHERE id = $1
RETURNING *;

-- name: DeleteHomeVideoRelease :exec
DELETE FROM home_video_releases
WHERE id = $1;
