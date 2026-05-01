# Build stage — installiert Dev-Deps inkl. tsup + typescript und baut
# das ESM/CJS-Bundle zusammen mit dem Workspace-Paket
# @npm9912/stream-analyzer.
FROM node:22-alpine AS build
WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@10.18.0 --activate

# Kopiere die für den Workspace-Hoist benötigten Manifeste, damit
# pnpm fetch ohne Voll-Source-Copy starten kann.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/analyzer-service/package.json apps/analyzer-service/
COPY packages/stream-analyzer/package.json packages/stream-analyzer/

RUN pnpm install --frozen-lockfile --ignore-scripts \
    --filter @npm9912/stream-analyzer... \
    --filter @npm9912/analyzer-service...

# Quellen für stream-analyzer und analyzer-service
COPY packages/stream-analyzer packages/stream-analyzer
COPY apps/analyzer-service apps/analyzer-service

RUN pnpm --filter @npm9912/stream-analyzer run build \
 && pnpm --filter @npm9912/analyzer-service run build

# Runtime stage — schlankes Node-Image mit production-only Deps.
FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    ANALYZER_LISTEN_HOST=0.0.0.0 \
    ANALYZER_LISTEN_PORT=7000

RUN corepack enable && corepack prepare pnpm@10.18.0 --activate

COPY --from=build /workspace/package.json /workspace/pnpm-lock.yaml /workspace/pnpm-workspace.yaml ./
COPY --from=build /workspace/apps/analyzer-service/package.json apps/analyzer-service/
COPY --from=build /workspace/packages/stream-analyzer/package.json packages/stream-analyzer/

RUN pnpm install --frozen-lockfile --prod \
    --filter @npm9912/analyzer-service...

COPY --from=build /workspace/apps/analyzer-service/dist apps/analyzer-service/dist
COPY --from=build /workspace/packages/stream-analyzer/dist packages/stream-analyzer/dist

USER node
EXPOSE 7000

CMD ["node", "apps/analyzer-service/dist/main.cjs"]
