FROM node:26-slim AS base

ENV PNPM_HOME="/pnpm" \
    PATH="$PNPM_HOME/bin:$PATH"

WORKDIR /app

RUN npm install -g pnpm@11

FROM base AS build

RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    --mount=type=bind,source=package.json,target=/app/package.json \
    --mount=type=bind,source=pnpm-lock.yaml,target=/app/pnpm-lock.yaml \
    pnpm install --frozen-lockfile

COPY . /app

RUN pnpm gen && pnpm run build

FROM base AS runtime

ENV NODE_ENV=production
COPY --from=build --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/dist ./dist
COPY --from=build --chown=node:node /app/package.json ./package.json
COPY --from=build --chown=node:node /app/vite.config.ts ./vite.config.ts

USER node

EXPOSE 4173

CMD ["pnpm", "exec", "vite", "preview", "--host", "--port", "4173"]
