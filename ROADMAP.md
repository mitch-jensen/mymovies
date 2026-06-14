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
- [ ] Pick the stack (deferred; see Open decisions).
- [ ] Generate a fully-typed TypeScript client from the OpenAPI spec.
- [ ] Render bookcases/shelves; spines show the title as plain text (MVP).
- [ ] Search → highlight a movie's location.
- [ ] (Post-MVP) Realistic spine appearance.

### Cross-cutting backlog
- [ ] **API DTOs vs. db types:** handlers currently return sqlc-generated structs
      (`db.Movie`) directly, leaking DB types into the OpenAPI contract / TS
      client. Define explicit request/response types so the public schema is
      decoupled from the database. Matters because of the schema-first approach.
- [ ] API-level (httptest) tests for handlers once the shared harness exists.
- [ ] Pagination on list endpoints.
- [ ] Request validation via huma input tags.
- [ ] CI (GitHub Actions) running `just check`.
- [ ] Serve the auto-generated OpenAPI / docs UI from huma.

## Decisions made
- **Single-user, local-only, no auth.** Runs locally for one user; no
  authentication/authorization layer.
- **Placement:** dedicated `placements` table (release ↔ shelf + ordered
  position), not columns on `home_video_releases`.
- **Schema-first API:** huma's generated OpenAPI spec is the contract; the
  frontend consumes a fully-typed TypeScript client generated from it.
- **Spine rendering (MVP):** plain-text movie title on a spine shape. No spine
  appearance data (dimensions/colour/art) in the MVP.
- **Frontend stack:** deferred until the backend is more feature-complete.

## Open decisions
- **Frontend stack:** chosen later, once the backend stabilises (will consume the
  generated TypeScript client).
- **Spine appearance data (post-MVP):** where to store dimensions/colour/art.

## How to resume
1. Read this file and [AGENTS.md](AGENTS.md).
2. `just test` to confirm green.
3. Pick the first unchecked item in the current phase.
