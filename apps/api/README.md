# m-trace API — Go-Spike-Prototyp

Backend-Spike gemäß `docs/plan-spike.md` §6.2 und
`docs/spike/0001-backend-stack.md` §12.2.

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

`hexagon/` und `adapters/` liegen direkt unter `apps/api/`,
flach und sichtbar. Details in `docs/plan-spike.md` §12.1.

```text
apps/api/
├── cmd/api/main.go
├── hexagon/
│   ├── domain/
│   ├── port/{driving,driven}/
│   └── application/
└── adapters/
    ├── driving/http/
    └── driven/{persistence,telemetry,metrics}/
```

## Status

Bootstrap. `/api/health` liefert `200 OK`. Domain, Use Case, Tests
und Pflicht-Endpunkte folgen in nachfolgenden Commits.
