-- +goose Up
-- Trigram index powers fuzzy, case-insensitive title search (ILIKE + similarity)
-- without a separate search service.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX movies_title_trgm_idx ON movies USING gin (title gin_trgm_ops);

-- +goose Down
DROP INDEX movies_title_trgm_idx;

DROP EXTENSION pg_trgm;
