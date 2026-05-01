# @npm9912/analyzer-service

Internal HTTP wrapper around `@npm9912/stream-analyzer` for the m-trace
API. Runs as a Node container next to `apps/api`, exposes
`POST /analyze` and `GET /health`. The Go backend speaks to this
service via HTTP so its distroless-static image stays free of any
Node/TypeScript runtime (plan-0.3.0 §7).

Status: **0.3.0 Tranche 6**.

## Endpoints

### `POST /analyze`

Request (`application/json`):

```json
{ "kind": "url", "url": "https://cdn.example.test/manifest.m3u8" }
```

or

```json
{ "kind": "text", "text": "#EXTM3U\n…", "baseUrl": "https://cdn.example.test/" }
```

Optional `fetch` block forwards loader limits:

```json
{ "kind": "url", "url": "...", "fetch": { "timeoutMs": 5000, "maxBytes": 1048576, "maxRedirects": 3 } }
```

Response: 200 with the full `AnalyzeOutput` JSON from
`@npm9912/stream-analyzer` (siehe `docs/user/stream-analyzer.md` §2).
Failure modes (415/400/413/500) carry a small problem-shape
`{ status: "error", code, message }`.

### `GET /health`

`200 { "status": "ok" }`. Used as a liveness probe.

## Running locally

```bash
pnpm install
pnpm --filter @npm9912/analyzer-service run build
pnpm --filter @npm9912/analyzer-service run start
```

`ANALYZER_LISTEN_HOST` (default `0.0.0.0`) and
`ANALYZER_LISTEN_PORT` (default `7000`) are the only env knobs.

## Container

Dockerfile builds a multi-stage image based on `node:22-alpine`:

```bash
docker build -f apps/analyzer-service/Dockerfile -t m-trace-analyzer:0.3.0 .
docker run --rm -p 7000:7000 m-trace-analyzer:0.3.0
```

The Compose stack starts it as `analyzer-service` and routes
`apps/api` to it via `ANALYZER_BASE_URL=http://analyzer-service:7000`.
