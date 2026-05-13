# Build stage — installiert die Workspace-Deps (Dev + Prod) für
# stream-analyzer und analyzer-service, baut beide Pakete und
# konsolidiert die produktiven Artefakte über `pnpm deploy --prod`
# in `/deploy`. Damit braucht die Runtime-Stage keinen weiteren
# pnpm-install-Schritt mehr (plan-0.3.0 §9 Tranche 7.5/3).
FROM node:22-trixie-slim AS build
WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@10.18.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/analyzer-service/package.json apps/analyzer-service/
COPY packages/stream-analyzer/package.json packages/stream-analyzer/

RUN pnpm install --frozen-lockfile --ignore-scripts \
    --filter @pt9912/stream-analyzer... \
    --filter @pt9912/analyzer-service...

COPY packages/stream-analyzer packages/stream-analyzer
COPY apps/analyzer-service apps/analyzer-service

RUN pnpm --filter @pt9912/stream-analyzer run build \
 && pnpm --filter @pt9912/analyzer-service run build

# `pnpm deploy --prod` flacht die Workspace-Symlinks in eine
# selbst-tragende node_modules-Hierarchie ab und lässt die Dev-Deps
# weg. `--legacy` ist nötig, weil pnpm v10 sonst inject-workspace-
# packages erwartet — was wir nicht globalisiert aktiv haben (siehe
# Plan §7.5/3 Begründung). Das Ergebnis-Verzeichnis kann von der
# Runtime-Stage als read-only-Bundle übernommen werden.
RUN pnpm --filter @pt9912/analyzer-service deploy --prod --legacy /deploy

# Runtime stage — startet direkt aus dem deploy-Bundle, ohne pnpm
# binär oder einen weiteren install-Schritt.
FROM node:22-trixie-slim AS runtime
WORKDIR /app

# `node:*-trixie-slim`-Images bringen `npm` unter
# `/usr/local/lib/node_modules/npm/` mit. NPM tragt eine eigene
# `picomatch@4.0.3`-Kopie in seinen internen `node_modules`, die
# Trivy als HIGH meldet (CVE-2026-33671). Im Runtime-Image starten
# wir mit `node dist/main.cjs` und brauchen npm nicht — entfernen
# schliesst diese CVE-Quelle ohne Workarounds (analog Dashboard-
# Dockerfile, plan-0.8.5 Tranche 1 Closeout).
RUN rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

ENV NODE_ENV=production \
    ANALYZER_LISTEN_HOST=0.0.0.0 \
    ANALYZER_LISTEN_PORT=7000

COPY --from=build /deploy ./

USER node
EXPOSE 7000

CMD ["node", "dist/main.cjs"]
