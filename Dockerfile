# syntax=docker/dockerfile:1.7

# Root TypeScript workspace build.
#
# This Dockerfile keeps pnpm installs and generated node_modules inside Docker
# layers. It covers the pnpm workspace members under apps/* and packages/*.

FROM node:22-trixie-slim AS pnpm-base

WORKDIR /workspace

RUN corepack enable && corepack prepare pnpm@10.18.0 --activate

FROM pnpm-base AS lock-refresh-tool

ENV XDG_CACHE_HOME=/tmp/.cache

FROM pnpm-base AS deps

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/analyzer-service/package.json apps/analyzer-service/package.json
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY packages/player-sdk/package.json packages/player-sdk/package.json
COPY packages/stream-analyzer/package.json packages/stream-analyzer/package.json

RUN pnpm install --frozen-lockfile --ignore-scripts

FROM deps AS source

COPY eslint.config.shared.mjs ./
COPY scripts scripts
COPY contracts contracts
COPY spec spec
COPY apps/api/adapters/driven/streamanalyzer/testdata apps/api/adapters/driven/streamanalyzer/testdata
COPY apps/api/hexagon/application/register_playback_event_batch.go apps/api/hexagon/application/register_playback_event_batch.go
COPY apps/analyzer-service apps/analyzer-service
COPY apps/dashboard apps/dashboard
COPY packages/player-sdk packages/player-sdk
COPY packages/stream-analyzer packages/stream-analyzer

FROM source AS build

RUN pnpm run build
RUN pnpm install --frozen-lockfile --ignore-scripts --offline

FROM build AS test

RUN pnpm run test

FROM build AS lint

RUN pnpm run lint

FROM build AS coverage

RUN pnpm --filter @npm9912/player-sdk run test:coverage \
 && pnpm --filter @npm9912/m-trace-dashboard run test:coverage \
 && pnpm --filter @npm9912/stream-analyzer run test:coverage \
 && pnpm --filter @npm9912/analyzer-service run test:coverage

FROM deps AS audit

RUN pnpm audit --audit-level high

FROM build AS sdk-performance-smoke

RUN pnpm --filter @npm9912/player-sdk run performance:smoke

FROM build AS sdk-pack-smoke

RUN mkdir -p .tmp/player-sdk-pack \
 && pnpm --filter @npm9912/player-sdk run pack:smoke

FROM build AS public-api-check

RUN pnpm --filter @npm9912/player-sdk exec node scripts/check-public-api.mjs

FROM build AS cli-smoke

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
 && apt-get install -y --no-install-recommends jq python3 \
 && rm -rf /var/lib/apt/lists/*
RUN bash scripts/smoke-cli.sh
