-- +goose Up
CREATE EXTENSION isn;

CREATE TABLE
movies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    release_year INT NOT NULL,
    runtime_min INT
);

CREATE TABLE
home_video_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    movie_id UUID NOT NULL REFERENCES movies (id) ON DELETE CASCADE,
    studio TEXT,
    country_code CHAR(2),
    upc UPC UNIQUE,
    ean EAN13 UNIQUE,
    asin TEXT UNIQUE,
    release_date DATE,
    casing TEXT,
    slipcover BOOLEAN,
    blu_ray_discs INT,
    dvd_discs INT,
    digital_copy BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    watched BOOLEAN NOT NULL,
    comment TEXT,
    retailer TEXT,
    price NUMERIC,
    price_comment TEXT
);

-- +goose Down
DROP TABLE home_video_releases;

DROP TABLE movies;

DROP EXTENSION isn;
