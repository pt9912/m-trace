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

Pre-MVP `0.1.0`-Skelett, integriert aus dem Backend-Spike-Sieger.
Vorhanden: Domain, Use Case `RegisterPlaybackEventBatch`, alle drei
Pflicht-Endpunkte (`POST /api/playback-events`,
`GET /api/sessions/{id}`, `GET /metrics`), `/api/health`,
Unit- und HTTP-Integrationstests grün.

API-Kontrakt (frozen): `docs/spike/backend-api-contract.md`.
