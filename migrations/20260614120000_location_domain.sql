-- +goose Up
-- Physical location domain: a movie's release lives on a shelf within a
-- bookcase. Ordering columns ("position") capture left-to-right / top-to-bottom
-- arrangement so the collection can be rendered to match the real-world setup.

CREATE TABLE
bookcases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE
shelves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bookcase_id UUID NOT NULL REFERENCES bookcases (id) ON DELETE CASCADE,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A release is in at most one place, hence release_id is UNIQUE.
CREATE TABLE
placements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id UUID NOT NULL UNIQUE REFERENCES home_video_releases (id) ON DELETE CASCADE,
    shelf_id UUID NOT NULL REFERENCES shelves (id) ON DELETE CASCADE,
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX shelves_bookcase_id_idx ON shelves (bookcase_id);

CREATE INDEX placements_shelf_id_idx ON placements (shelf_id);

-- +goose Down
DROP TABLE placements;

DROP TABLE shelves;

DROP TABLE bookcases;
