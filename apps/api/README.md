# m-trace API (Go)

Backend-Service für m-trace. Hexagon-Architektur mit Go 1.22 +
stdlib `net/http` + Prometheus + OpenTelemetry, Distroless-Runtime.

Stack-Entscheidung: `docs/adr/0001-backend-stack.md` (Accepted).
Roadmap & Release-Plan: `docs/roadmap.md`, `docs/lastenheft.md`.
Hexagon-Struktur: `docs/plan-spike.md` §12.1.

## Workflow (Docker-only)

```bash
make test     # docker build --target test
make lint     # docker build --target lint  (Soll, golangci-lint)
make build    # docker build --target runtime
make run      # docker run -p 8080:8080
```

Lokales Go ist **nicht erforderlich** — das Build-Image
`golang:1.22` bringt die Toolchain mit.

## Verzeichnisstruktur

```text
apps/api/
├── cmd/api/main.go
├── hexagon/
│   ├── domain/
│   ├── port/{driving,driven}/
│   └── application/
└── adapters/
    ├── driving/http/
    └── driven/{auth,metrics,persistence,ratelimit,telemetry}/
```

## Status

`0.2.0`-Stand, integriert aus dem Backend-Spike-Sieger und der
`0.1.x`/`0.2.0`-Implementierung. Vorhanden: Domain, Use Case
`RegisterPlaybackEventBatch`, Pflicht-Endpunkte
(`POST /api/playback-events`, `GET /api/health`, `GET /api/metrics`),
Stream-Sessions-Endpoints (`GET /api/stream-sessions`,
`GET /api/stream-sessions/{id}`), CORS/Origin-Validierung, Rate-Limits,
Telemetry-Port, OTel-Span-Export und Prometheus-Mindestmetriken für
Sessions, API-Requests, Playback-Fehler, Rebuffer-Events und Startup-Zeit
sowie Unit- und HTTP-Integrationstests.

API-Kontrakt (frozen): `docs/spike/backend-api-contract.md`.
