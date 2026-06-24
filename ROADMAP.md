# Roadmap

Living plan for **mymovies**. Update this as work progresses so any new session
(human or agent) can resume from here. For *how to work* in this repo (commands,
skills, style) see [AGENTS.md](AGENTS.md).

## Vision

A full-stack app to organise a physical movie collection. The user models their
real bookcases and shelves (rendered realistically, spines showing), then
searches for a movie and is shown where it physically is, so they can go get it.

## Domain model (target)

```
bookcase 1───* shelf 1───* placement *───1 home_video_release *───1 movie
```

- **movie** — the film (title, year, runtime). Exists.
- **home_video_release** — a physical disc/edition of a movie (studio, barcodes,
  discs, price, casing…). Table exists; no queries/API yet.
- **bookcase** — a physical unit in the room (name/label, ordering, dimensions
  for rendering). Not built.
- **shelf** — a row within a bookcase (position, dimensions). Not built.
- **placement** — a release sitting on a shelf at an ordered position. This is
  what makes "where is it?" answerable. Not built.

## Status (updated 2026-06-14)

- Go backend only; no frontend.
- `movies` + `home_video_releases` tables exist (sqlc + goose + pgx + huma/chi).
- Movie queries exist: Create/Get/List/Update/Delete.
- Exposed over HTTP: `GET /movies`, `POST /movies` only.
- Integration-test harness (testcontainers, template-DB clone per test). Green.
- Done this session: graceful shutdown, accurate startup logging, config
  simplification + first config test.

## Plan

### Phase 1 — Complete the movie & release API ✓ done
- [x] Expose remaining movie routes: `GET/PUT/DELETE /movies/{id}` (404 on
      not-found; `UpdateMovie` now `RETURNING *` so PUT returns the resource).
- [x] dbstore tests for `ListMovies`, `UpdateMovie`, `DeleteMovie`.
- [x] Shared test-DB harness extracted to `internal/testdb`; `internal/api` now
      has real httptest handler tests.
- [x] API DTOs (`Movie`, `MovieFields`) decouple the schema from db types — done
      early because of the schema-first / typed-client goal.
- [x] `home_video_releases` CRUD: queries + `Release`/`ReleaseFields` DTOs +
      routes (`POST`/`GET /movies/{movieId}/releases`, `GET`/`PUT`/`DELETE
      /releases/{id}`), with creation 404ing on a missing movie. dbstore + api
      tests cover it.

### Phase 2 — Physical location domain ✓ done
- [x] Migration `20260614120000_location_domain`: `bookcases`, `shelves`,
      `placements` (release↔shelf, `release_id` UNIQUE, ordered `position`,
      `ON DELETE CASCADE` throughout).
- [x] Queries: bookcase & shelf CRUD; `PlaceRelease` (upsert/move via
      `ON CONFLICT (release_id)`); `RemovePlacement`; `LocateRelease` (join via
      `sqlc.embed` → `{Bookcase, Shelf, Placement}`). dbstore tests cover them.
- [x] API routes + DTOs + httptest: bookcase CRUD (`/bookcases`), shelf CRUD
      (`/bookcases/{id}/shelves`, `/shelves/{id}`), `PUT`/`DELETE
      /releases/{id}/placement`, and `GET /releases/{id}/location` (the
      **locate** endpoint; 404 when unplaced).

### Phase 3 — Search ✓ done
- [x] `GET /search?q=&limit=` — fuzzy, case-insensitive title search via
      `pg_trgm` (GIN trigram index; `ILIKE` substring + `similarity` ranking).
      Returns `[]SearchResult{movie, locatedReleases:[{release, location}]}` with
      each placed copy's physical location inline. No Elasticsearch needed at
      single-user scale; "instant feel" is a client concern (debounce + cap).
      dbstore + api tests cover case-insensitivity, typo tolerance, limit, and
      inline location.

### Phase 5 — Dimensions & packing engine (post-MVP, deferred)
The big one. Deferred entirely for now (data model included) — recorded here so
the MVP location API doesn't block on it. The existing `placements(shelf_id,
position)` table is the engine's **commit target**: preview computes a candidate
layout in memory; commit writes it to `placements`. So nothing built so far is
wasted.

Data model to add (when we start this phase):
- [ ] **Casing catalog**: `casings` table with standard case dimensions in mm —
      `height_mm` (vertical / standing-fit), `width_mm` (into shelf depth),
      `spine_mm` (consumed along the shelf when standing spine-out). Replace the
      free-text `home_video_releases.casing` with a `casing_id` FK.
- [ ] **Per-release dimension override**: nullable `height_mm` / `width_mm` /
      `spine_mm` on `home_video_releases`; effective dim = override ?? casing's.
- [ ] **Shelf dimensions**: interior `height_mm` / `width_mm` / `depth_mm` on
      `shelves`. No special "short shelf" flag — a shelf is short iff a release
      won't fit standing; the engine derives that from `height_mm`.
- [ ] **Bookcase fill direction**: e.g. `top_to_bottom` / `bottom_to_top`.
- [ ] **Placement orientation**: standing vs. laid flat (for oversized releases
      that only fit horizontally).

Engine behaviour:
- [ ] Pack by `spine_mm` along each shelf (1-D wrap); a release goes on the next
      shelf in fill order that fits it (tall enough to stand, else laid flat on a
      short shelf). **No explicit pinning** — a release's shelf is simply its
      current materialised assignment; adds/removes may move it, and the engine
      finds the next/previous fitting shelf.
- [ ] Insert/remove cascades across shelves (may shift by more than one shelf to
      make things fit).
- [ ] **Preview → commit**: compute candidate layout without persisting; commit
      writes `placements`.
- [ ] Heuristics (user-selectable, in preference order):
      1. **Preserve current order** — on add/remove, shift releases forward/back
         to keep the existing order.
      2. **Alphabetise** — reflow the whole collection in title order (for
         re-organising an out-of-order collection).

### Phase 4 — Frontend (full-stack) ← current
- [x] Stack chosen (see Decisions made).
- [x] Scaffolded `frontend/`: Vite + React + TS, TanStack Router (file-based) +
      Query, `@hey-api/openapi-ts` client gen wired to `../openapi.yaml`, dev
      proxy `/api` → backend `:8081`. First screen = search with inline location.
      *Pending the user's first `pnpm install` / `pnpm gen` / `pnpm dev` — may
      need version/import tweaks (can't verify Node tooling here).*
- [ ] Generate the typed client (`pnpm gen`) and get the search screen running.
- [ ] Render bookcases/shelves; spines show the title as plain text (MVP).
- [ ] Manage collection from the UI (add movies/releases, place on shelves).
- [ ] (Post-MVP) Realistic spine appearance.

### Cross-cutting backlog
- [x] **API DTOs vs. db types:** handlers now use explicit `Movie` / `MovieFields`
      types instead of leaking sqlc structs into the schema. Apply the same
      pattern to future resources (releases, bookcases…).
- [x] API-level (httptest) tests for handlers (via `internal/testdb`).
- [x] OpenAPI: huma serves `/openapi.json`, `/openapi.yaml`, and a `/docs` UI
      out of the box (test-locked). `just openapi` exports the spec to
      `frontend/openapi.yaml` (no DB needed) as the contract for client
      generation. It lives under `frontend/` (gitignored) so the frontend Docker
      build context is self-contained.
- [x] Coverage measured via `just cover` (cross-package, currently ~77%).
      Remaining gaps are entrypoints (`main`, `cmd/openapi`, `server.Run`),
      generated `WithTx`, and infra error/teardown paths — intentionally untested.
- [x] Bookcase & shelf CRUD now have full api + dbstore test coverage (previously
      only create was tested).
- [x] Shelf contents endpoint: `GET /shelves/{id}/placements` returns each placed
      release on a shelf with its movie, in slot order (the spine-rendering feed;
      replaced the unused `ListPlacementsByShelf` query).
- [ ] Pagination on list endpoints.
- [ ] Request validation via huma input tags.
- [ ] `GET /shelves/{id}` (single shelf) — not yet needed; revisit if the UI wants it.
- [ ] CI (GitHub Actions) running `just check`. *(Owner: Mitch.)*

## Decisions made
- **Single-user, local-only, no auth.** Runs locally for one user; no
  authentication/authorization layer.
- **Domain module (`internal/collection`).** A seam between the HTTP layer and
  sqlc: it owns multi-step workflows (`Search` assembly, `PlaceRelease` /
  `AddShelf` / `CreateRelease` existence checks) and translates `pgx.ErrNoRows`
  into `collection.ErrNotFound`. `internal/api` no longer imports `pgx`; handlers
  are thin (decode → call `collection` → `mapErr` → DTO) and a single `mapErr`
  maps domain errors to HTTP status. The module returns `db.*` row types (and
  `SearchResult`/`LocatedRelease` composites); `api` keeps owning DTO mapping —
  no third type system. This is the intended home for the Phase 5 packing engine.
- **Placement:** dedicated `placements` table (release ↔ shelf + ordered
  position), not columns on `home_video_releases`.
- **Schema-first API:** huma's generated OpenAPI spec is the contract; the
  frontend consumes a fully-typed TypeScript client generated from it.
- **Spine rendering (MVP):** plain-text movie title on a spine shape. No spine
  appearance data (dimensions/colour/art) in the MVP.
- **Frontend stack:** **Vite + React + TypeScript** SPA (no SSR — local,
  single-user), with **TanStack Router** + **TanStack Query**. Package manager
  **pnpm**. Lives in `frontend/` with Go staying at repo root.
- **Client generation:** **`@hey-api/openapi-ts`** with its `@tanstack/react-query`
  plugin, run against `openapi.yaml` (huma emits OpenAPI 3.1 — hey-api supports
  it). Use the generated `queryOptions`/`mutationOptions`; hand-write only thin
  wrappers for debounced search, cross-entity cache invalidation, and optimistic
  updates.

## Open decisions
- **Spine appearance data (post-MVP):** where to store dimensions/colour/art.

## How to resume
1. Read this file, [CONTEXT.md](CONTEXT.md) (domain language), and
   [AGENTS.md](AGENTS.md).
2. `just test` to confirm green.
3. Pick the first unchecked item in the current phase.
