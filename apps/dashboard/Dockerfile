FROM node:22-bookworm-slim AS build

WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
COPY apps/dashboard/package.json apps/dashboard/package.json
COPY packages/player-sdk/package.json packages/player-sdk/package.json

RUN pnpm install --frozen-lockfile --filter @m-trace/dashboard...

COPY apps/dashboard apps/dashboard
COPY packages/player-sdk packages/player-sdk

RUN pnpm --filter @m-trace/player-sdk run build && pnpm --filter @m-trace/dashboard run build

FROM node:22-bookworm-slim AS runtime

WORKDIR /app/apps/dashboard

ENV HOST=0.0.0.0
ENV PORT=5173
ENV NODE_ENV=production

COPY --from=build /app /app

EXPOSE 5173

CMD ["node", "build"]
