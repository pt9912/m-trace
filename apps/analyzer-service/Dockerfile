# Build stage — installiert die Workspace-Deps (Dev + Prod) für
# stream-analyzer und analyzer-service, baut beide Pakete und
# konsolidiert die produktiven Artefakte über `pnpm deploy --prod`
# in `/deploy`. Damit braucht die Runtime-Stage keinen weiteren
# pnpm-install-Schritt mehr (plan-0.3.0 §9 Tranche 7.5/3).
FROM node:22-alpine AS build
WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@10.18.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/analyzer-service/package.json apps/analyzer-service/
COPY packages/stream-analyzer/package.json packages/stream-analyzer/

RUN pnpm install --frozen-lockfile --ignore-scripts \
    --filter @npm9912/stream-analyzer... \
    --filter @npm9912/analyzer-service...

COPY packages/stream-analyzer packages/stream-analyzer
COPY apps/analyzer-service apps/analyzer-service

RUN pnpm --filter @npm9912/stream-analyzer run build \
 && pnpm --filter @npm9912/analyzer-service run build

# `pnpm deploy --prod` flacht die Workspace-Symlinks in eine
# selbst-tragende node_modules-Hierarchie ab und lässt die Dev-Deps
# weg. `--legacy` ist nötig, weil pnpm v10 sonst inject-workspace-
# packages erwartet — was wir nicht globalisiert aktiv haben (siehe
# Plan §7.5/3 Begründung). Das Ergebnis-Verzeichnis kann von der
# Runtime-Stage als read-only-Bundle übernommen werden.
RUN pnpm --filter @npm9912/analyzer-service deploy --prod --legacy /deploy

# Runtime stage — startet direkt aus dem deploy-Bundle, ohne pnpm
# binär oder einen weiteren install-Schritt.
FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    ANALYZER_LISTEN_HOST=0.0.0.0 \
    ANALYZER_LISTEN_PORT=7000

COPY --from=build /deploy ./

USER node
EXPOSE 7000

CMD ["node", "dist/main.cjs"]
