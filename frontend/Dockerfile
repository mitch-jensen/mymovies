# syntax=docker/dockerfile:1

# Multi-stage build for the Vite + React SPA, mirroring the aggressive
# cache/bind-mount approach in the root Go Dockerfile.
#
# DHI assumptions to confirm against your DHI catalog entitlement:
#   - `dhi.io/node:24-alpine-dev` is the build/dev variant (has a shell + npm so
#     corepack/pnpm work). Adjust the tag if your catalog differs.
#   - The runtime variant runs as the `nonroot` user by default.
FROM dhi.io/node:24-alpine-dev AS base
USER root
ENV PNPM_HOME=/pnpm \
    PATH=/pnpm:$PATH \
    CI=true
RUN corepack enable
WORKDIR /app

# --- deps: install node_modules from the lockfile only -----------------------
# Bind-mount the manifests instead of COPYing so the layer is keyed purely on
# their contents; cache the pnpm content-addressable store across builds.
FROM base AS deps
RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=pnpm-lock.yaml,target=pnpm-lock.yaml \
    --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm install --frozen-lockfile

# --- build: generate the client + route tree, then bundle --------------------
# `pnpm gen` (openapi-ts) and the TanStack Router plugin both emit files that are
# gitignored, so generate before bundling. `vite build` runs the router plugin
# which writes routeTree.gen.ts, so we skip the standalone `tsc` step here.
FROM deps AS build
COPY . .
RUN pnpm gen && pnpm exec vite build

# --- runtime: serve the built SPA via `vite preview` -------------------------
# vite preview reuses the proxy config (preview.proxy) so /api still reaches the
# backend in the compose network. node_modules comes from deps; only dist and
# the config the preview server needs are copied from build.
FROM base AS runtime
ENV NODE_ENV=production
COPY --from=deps /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
COPY --from=build /app/vite.config.ts ./vite.config.ts
COPY --from=build /app/package.json ./package.json
USER nonroot
EXPOSE 4173
CMD ["pnpm", "exec", "vite", "preview", "--host", "--port", "4173"]
