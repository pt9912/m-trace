# syntax=docker/dockerfile:1.7

# ---------------------------------------------------------------------------
# m-trace API spike — Go variant
#
# Docker-only workflow per docs/planning/plan-spike.md §14.11. The repository ships
# no Go toolchain, no go.sum manipulation outside Docker, and no host-level
# build artefacts. All build/test/lint steps run via:
#   docker build --target <stage> -t m-trace-api-spike:go-<stage> .
#
# Stages:
#   deps     - resolves Go module dependencies (cache layer)
#   compile  - compiles the binary for fast feedback
#   lint     - runs golangci-lint (Soll target, plan §14.9)
#   test     - runs `go test ./...`
#   coverage - runs `go test -coverprofile` + `go tool cover -func`,
#              prüft Threshold (Default 90 %) und exportiert das
#              Profil + den HTML-Report in /out
#   build    - produces the final binary
#   runtime  - distroless runtime image
# ---------------------------------------------------------------------------

# ---- deps ------------------------------------------------------------------
FROM golang:1.26 AS deps

WORKDIR /src
ENV GOFLAGS="-mod=readonly" \
    GOMODCACHE=/go/pkg/mod \
    GOCACHE=/root/.cache/go-build

COPY go.mod ./
# go.sum is copied if present (initial bootstrap may not have it yet);
# `go mod download` will populate the cache from go.mod alone.
COPY go.su[m] ./

# Ensure GOMODCACHE exists even with zero external deps so later stages can
# COPY --from=deps reliably.
RUN mkdir -p "$GOMODCACHE" && go mod download

# ---- compile ---------------------------------------------------------------
FROM deps AS compile

COPY . .
RUN CGO_ENABLED=0 go build -o /tmp/api ./cmd/api

# ---- lint ------------------------------------------------------------------
# golangci-lint mit dem SOLID-nahen Pflicht-Profil aus
# spec/lastenheft.md §10.1 (5 Defaults + 24 SOLID-nahe Linter, ohne
# //nolint-Suppressions). Konfiguration in apps/api/.golangci.yml;
# Pro-Scope-Carveouts (Tests, scripts/, Test-Contract-Bibliothek) sind
# dort dokumentiert. Bezug: docs/user/quality.md §1.2.
FROM golangci/golangci-lint:v2.12.1-alpine AS lint

WORKDIR /src
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY . .
RUN golangci-lint run ./...

# ---- test ------------------------------------------------------------------
FROM deps AS test

COPY . .
RUN CGO_ENABLED=0 go test ./...

# ---- coverage --------------------------------------------------------------
# Misst Coverage über hexagon/ und adapters/ (cmd/api ist Wiring +
# Signal-Handling; bewusst aus dem Coverage-Range ausgenommen) und
# erzwingt eine Threshold von 90 % (überschreibbar via Build-Argument
# COVERAGE_THRESHOLD). Erzeugt zusätzlich einen HTML-Report — nutzbar
# via `docker build --target coverage --output type=local,dest=...`.
#
# Bewusst **nicht** im coverpkg-Profil:
#   - `internal/storage` (ADR-0002 §8.2 Apply-Runner): defensive
#     Error-Pfade (sql.Open-Failures, fs.FS-Read-Errors,
#     BeginTx/Commit-Failures) sind ohne FS-/SQL-Mocks unerreichbar.
#   - `adapters/driven/persistence/sqlite` (ADR-0002 §8.1 SQLite-
#     Adapter): dieselbe Pfad-Klasse — Tx-Begin-Failures, Time-Parse-
#     Errors, Scan-Errors brauchen Mocks.
#   - `adapters/driven/persistence/contract`: Test-Helper-Paket (kein
#     Produktionscode); würde Coverage-Profil künstlich beeinflussen.
#
# Plan-0.4.0 §2.6 trackt die Folge-Entscheidung, ob die ausgenommenen
# Pakete ein dediziertes Storage-Coverage-Setup mit niedrigerer
# Threshold bekommen.
FROM deps AS coverage

ARG COVERAGE_THRESHOLD=90
ENV COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD}

COPY . .
RUN mkdir -p /out && \
    COVERPKG=$(go list ./hexagon/... ./adapters/... | \
        grep -vE 'persistence/(sqlite|postgres|mysql|contract)$' | \
        tr '\n' ',' | sed 's/,$//') && \
    CGO_ENABLED=0 go test \
        -coverpkg="$COVERPKG" \
        -coverprofile=/out/coverage.out \
        -covermode=atomic \
        ./... && \
    go tool cover -func=/out/coverage.out | tee /out/coverage-func.txt && \
    go tool cover -html=/out/coverage.out -o /out/coverage.html && \
    go run ./scripts/coverage-overview \
        -profile /out/coverage.out \
        -threshold "$COVERAGE_THRESHOLD" \
        -out /out/index.html && \
    bash scripts/coverage-gate.sh /out/coverage-func.txt "$COVERAGE_THRESHOLD"

# ---- coverage-export -------------------------------------------------------
# `FROM scratch` mit ausschließlich /out drin — dadurch extrahiert
# `docker build --target coverage-export --output type=local,dest=…`
# nur die Coverage-Artefakte (Profil, Func-Text, HTML), nicht das ganze
# Toolchain-Rootfs.
FROM scratch AS coverage-export
COPY --from=coverage /out /

# ---- build -----------------------------------------------------------------
FROM deps AS build

COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api

# Leeres Stage-Verzeichnis für `/var/lib/mtrace` mit dem nonroot-User
# vorbereiten. Beim ersten Container-Start initialisiert Docker das
# benannte Volume mit den Permissions des Image-Mountpoints — sonst
# legt Docker das Volume als root:root an, und der nonroot-API-Prozess
# kann die SQLite-Datei nicht erzeugen.
RUN mkdir -p /staging/var/lib/mtrace

# ---- runtime ---------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=build /out/api /usr/local/bin/api
COPY --from=build --chown=65532:65532 /staging/var/lib/mtrace /var/lib/mtrace

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
