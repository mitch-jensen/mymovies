-- name: SearchMovies :many
-- Case-insensitive fuzzy title search: substring match (ILIKE) OR trigram
-- similarity (typo tolerance), ranked by similarity. Backed by the GIN trigram
-- index on movies.title.
SELECT * FROM movies
WHERE title ILIKE '%' || @query::text || '%'
   OR title % @query::text
ORDER BY similarity(title, @query::text) DESC, title
LIMIT @result_limit::int;

-- name: ListLocatedReleasesByMovies :many
-- For the given movies, every release that is physically placed, joined to its
-- bookcase/shelf/placement. Unplaced releases are omitted (no location).
SELECT
    sqlc.embed(r),
    sqlc.embed(b),
    sqlc.embed(s),
    sqlc.embed(p)
FROM home_video_releases r
JOIN placements p ON p.release_id = r.id
JOIN shelves s ON s.id = p.shelf_id
JOIN bookcases b ON b.id = s.bookcase_id
WHERE r.movie_id = ANY(@movie_ids::uuid[])
ORDER BY b.position, s.position, p.position;
