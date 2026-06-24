# Domain glossary — mymovies

The **ubiquitous language** for this project: the words the code, tests, and
conversation should all use for the same concepts. When you name a new concept
(a type, a module, an endpoint), reach for a term here first; when you coin a new
one, add it here. For the *plan* see [ROADMAP.md](ROADMAP.md); for *how to work*
see [AGENTS.md](AGENTS.md).

The domain is a person's physical movie collection — the discs they own and where
those discs physically sit in the room — made searchable so they can walk to the
shelf and grab one.

## Core domain (built)

- **movie** — a film, independent of any physical copy: `title`, `release_year`,
  `runtime_min`. A movie has many releases.
- **home video release** (**release**) — a specific physical edition of a movie:
  studio, barcodes (`upc`/`ean`/`asin`), `release_date`, `casing`, disc counts,
  price, `watched`, etc. Belongs to exactly one movie. This — not the movie — is
  the thing that physically occupies space on a shelf.
- **bookcase** — a physical storage unit standing in the room: `name` and a
  `position` capturing its left-to-right order among bookcases. Has many shelves.
- **shelf** — one row within a bookcase: a `position` capturing its order
  (top-to-bottom) within that bookcase. Has many placements.
- **placement** — a release sitting on a shelf at an ordered `position`. A release
  is in **at most one** place (`placements.release_id` is `UNIQUE`); deleting a
  release, shelf, or bookcase cascades. The placement is what makes "where is it?"
  answerable.
- **location** — the *resolved* physical spot of a placed release: its bookcase +
  shelf + placement together. Derived by joining, not a table of its own. A
  release that has no placement has no location.
- **collection** — the whole modelled set: every movie, its releases, and their
  physical layout. Also the name of the domain module (`internal/collection`)
  that owns operations over it.

## Operations (built)

- **place** — put a release on a shelf at a position, or move it (upsert on
  `release_id`). Fails as *not found* if the release or shelf doesn't exist.
- **remove** — take a release off its shelf (delete its placement).
- **locate** — given a release, return its location; *not found* if unplaced.
- **search** — fuzzy, case-insensitive title match (`pg_trgm`) returning each
  matching movie together with the location of every placed copy, inline.
- **shelf contents** — list everything placed on a shelf, in slot order, each
  with its release and movie. The feed for rendering a shelf's spines.

## Module map

- **dbstore** (sqlc-generated) — the storage layer; speaks SQL and `pgx`.
- **collection** (`internal/collection`) — the domain module / **seam** over
  storage. Owns multi-step workflows (search assembly, place/add-shelf/
  create-release existence checks) and translates `pgx.ErrNoRows` into the domain
  error `ErrNotFound`. Nothing above it depends on `pgx`.
- **api** (`internal/api`) — the HTTP **adapter** (huma + chi). Thin handlers map
  requests to `collection` calls and `collection` results to wire **DTOs**, and
  `mapErr` maps domain errors to status codes.

## Packing engine — reserved seam (not built)

Phase 5 (see ROADMAP) adds an engine that decides *which shelf each release lands
on* from physical dimensions, instead of the user placing each by hand. These
terms are reserved now so the engine has names to land on; the **`placements`
table is its commit target** — nothing built so far is wasted.

- **casing** — a standard case type with dimensions in mm (`height`, `width`,
  `spine`). A catalog that replaces today's free-text `release.casing`.
- **dimensions** — a release's effective `height`/`width`/`spine` in mm
  (per-release override, falling back to its casing's). `spine` is what's consumed
  along a shelf when standing spine-out.
- **fill direction** — the order shelves are filled within a bookcase (e.g.
  top-to-bottom vs bottom-to-top).
- **orientation** — how a release sits: **standing** (spine-out) or **laid flat**
  (for an oversized release that only fits a short shelf horizontally).
- **layout** — a computed assignment of releases to shelves and positions; the
  in-memory candidate the engine produces.
- **packing** — computing a layout: pack by `spine` along each shelf in fill
  order, sending a release to the next shelf that fits it (tall enough to stand,
  else laid flat). Inserts/removes may cascade across shelves.
- **preview / commit** — **preview** computes a candidate layout without
  persisting; **commit** writes it to `placements`. The current location domain is
  the commit target, so the manual `place`/`remove` operations and the engine
  share one materialised representation.
- **heuristic** — a user-selectable layout strategy, in preference order:
  **preserve current order** (shift to keep existing order on add/remove) or
  **alphabetise** (reflow the whole collection by title).
