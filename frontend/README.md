# Frontend

Vite + React + TypeScript SPA. TanStack Router (file-based routing) + TanStack
Query. The API client is generated from the backend's OpenAPI spec with
[`@hey-api/openapi-ts`](https://heyapi.dev) — no hand-written client.

## Setup

All commands run from this `frontend/` directory unless noted.

1. **Export the API spec** (from the repo root): `just openapi` → writes
   `../openapi.yaml`.
2. `pnpm install`
3. `pnpm gen` — generates the typed client into `src/client/` from
   `../openapi.yaml`. Re-run after any API change (and after `just openapi`).
4. **Start the backend** (repo root): `just dev` (or `just run`) — listens on
   `:8081`.
5. `pnpm dev` — Vite dev server. Requests to `/api/*` are proxied to the backend,
   so there's no CORS to configure.

## How it fits together

- `src/routes/` — file-based routes (TanStack Router). `index.tsx` is the search
  screen. The route tree is generated into `src/routeTree.gen.ts` on dev/build.
- `src/client/` — generated SDK + TanStack Query `*Options` helpers (gitignored).
- `src/hooks/` — small app-specific hooks (e.g. debounce) layered on top of the
  generated client.
- Dev proxy: `/api/*` → `http://localhost:8081` (override with `VITE_API_TARGET`).

## Notes for first run

A few things can't be verified without Node tooling, so adjust if `pnpm`
complains:

- **Dependency versions** in `package.json` are recent-but-approximate; bump with
  `pnpm up --latest` if needed.
- **Generated import paths** in `src/main.tsx` (`./client/client.gen`) and
  `src/routes/index.tsx` (`searchMoviesOptions`) assume hey-api's current output
  shape. If `pnpm gen` produces different names/paths, tweak those two imports.
- The **router Vite plugin** export is `TanStackRouterVite`; newer versions may
  call it `tanstackRouter` (see `vite.config.ts`).
