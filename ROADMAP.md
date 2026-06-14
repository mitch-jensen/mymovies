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

### Phase 1 — Complete the movie & release API ← current
- [ ] Expose remaining movie routes: `GET/PUT/DELETE /movies/{id}` (map
      not-found to 404). Queries already exist.
- [ ] dbstore tests for `ListMovies`, `UpdateMovie`, `DeleteMovie`.
- [ ] Queries + routes for `home_video_releases` (CRUD, linked to a movie).
- [ ] Decide & build a shared test-DB harness so `internal/api` can have
      handler tests (currently the harness lives only in `dbstore`).

### Phase 2 — Physical location domain
- [ ] Migration: `bookcases`, `shelves`, and placement of releases onto shelves
      (FK + ordered position).
- [ ] Queries + routes: create/list bookcases & shelves; place / move / remove a
      release; **locate** a release (→ bookcase + shelf + position).

### Phase 3 — Search
- [ ] Search movies by title (consider `pg_trgm` / full-text), returning the
      physical location. `GET /search?q=`.

### Phase 4 — Frontend (full-stack)
- [ ] Pick the stack (see Open decisions).
- [ ] Render bookcases/shelves with realistic spines.
- [ ] Search → highlight a movie's location.

### Cross-cutting backlog
- [ ] API-level (httptest) tests for handlers once the shared harness exists.
- [ ] Pagination on list endpoints.
- [ ] Request validation via huma input tags.
- [ ] CI (GitHub Actions) running `just check`.
- [ ] Serve the auto-generated OpenAPI / docs UI from huma.

## Open decisions
- **Placement modelling:** dedicated `placements` table (allows history/empty
  slots) vs. `shelf_id` + `position` columns on `home_video_releases` (simpler).
  Leaning: dedicated table.
- **Spine appearance data:** where to store dimensions/colour/art for rendering.
- **Frontend stack:** framework + how it talks to the API (the OpenAPI spec huma
  generates could drive a typed client).
- **Single-user vs. auth:** assume single-user for now?

## How to resume
1. Read this file and [AGENTS.md](AGENTS.md).
2. `just test` to confirm green.
3. Pick the first unchecked item in the current phase.
