FROM node:24-slim AS base

ENV PNPM_HOME="/pnpm" \
    PATH="$PNPM_HOME/bin:$PATH"

RUN corepack enable

FROM base AS prod-deps

RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=pnpm-lock.yaml,target=pnpm-lock.yaml \
    pnpm install --prod --frozen-lockfile

FROM base AS build

RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=pnpm-lock.yaml,target=pnpm-lock.yaml \
    pnpm install --frozen-lockfile

RUN pnpm run build

FROM base AS runtime

ENV NODE_ENV=production
COPY --from=prod-deps /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
COPY --from=build /app/vite.config.ts ./vite.config.ts
COPY --from=build /app/package.json ./package.json

USER nonroot

EXPOSE 4173

CMD ["pnpm", "exec", "vite", "preview", "--host", "--port", "4173"]
