# syntax=docker/dockerfile:1.7

# ---------------------------------------------------------------------------
# m-trace API spike — Go variant
#
# Docker-only workflow per docs/plan/planning/done/plan-spike.md §14.11. The repository ships
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
# Digest-gepinnt (Modul 14): Update = bewusster Commit, der die Digest-Zeile
# anhebt. Ermitteln via `docker buildx imagetools inspect golang:1.26.6`.
FROM golang:1.26.6@sha256:0d1d3a794be25f809dd2cb3160d8c73276c4056a9f8242a138e908ddeee7b6b6 AS deps

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
# Digest ermitteln via `docker buildx imagetools inspect golangci/golangci-lint:v2.12.1-alpine`.
FROM golangci/golangci-lint:v2.12.1-alpine@sha256:25cd0933d1a4ced61e907aabac044cfc0182ab067c09c25d568e15750e9695aa AS lint

WORKDIR /src
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY . .
RUN golangci-lint run ./...

# ---- test ------------------------------------------------------------------
FROM deps AS test

COPY . .
RUN CGO_ENABLED=0 go test ./...

# ---- race ------------------------------------------------------------------
# Race-Detector-Stage: `go test -race ./...` mit CGO_ENABLED=1. Opt-in
# (nicht in `make gates`), ~5–10× langsamer als der `test`-Stage.
# Runtime-Image bleibt CGO-frei `distroless-static` — diese Stage
# erbt nur `deps` und nutzt den GCC, der im Debian-`golang:1.26`-
# Base-Image bereits installiert ist. Aufruf via `make race` ruft
# mit `--no-cache-filter race`, damit ein einmal grüner Race-Lauf
# keinen späteren flaky Race über den Layer-Cache verdeckt.
FROM deps AS race

COPY . .
RUN CGO_ENABLED=1 go test -race ./...

# ---- benchmark ---------------------------------------------------------
# Hermetische Bench-Basis (Modul 14 — kein Mount): Quellen per COPY
# gebacken, kein RUN des Bench selbst. Der Lauf passiert bei `docker run`
# gegen dieses Image (`make benchmark-smoke`) — würde der Bench stattdessen
# hier als RUN-Schritt laufen, käme sein Ergebnis im BuildKit-Log
# präfixiert an und wäre für den verankerten `^Benchmark`-Regex in
# `scripts/check-bench-budgets.mjs` nicht mehr parsebar.
FROM deps AS benchmark

COPY . .

# ---- fuzz ----------------------------------------------------------------
# Hermetische Basis für `make fuzz-check` **und** `make mutation-report`
# (beide brauchen dieselbe Quellbasis, kein RUN hier). Crash-Reproduktionen
# (`testdata/fuzz/`) bzw. Mutation-Reports (`.tmp/mutation/`) entstehen erst
# beim `docker run`/`docker create`-Lauf gegen dieses Image und kommen per
# `docker cp` zurück — nicht per Bind-Mount, sonst gehörten sie `root`
# (Modul 14 §Besitz der Belege eines containerisierten Gates).
FROM deps AS fuzz

COPY . .

# ---- vuln ----------------------------------------------------------------
# Hermetische Basis für `make vuln-check`: `govulncheck` als Build-Schritt
# installiert (legitimer Build-Bestandteil), der eigentliche Scan läuft
# bei `docker run` gegen dieses Image — derselbe Grund wie bei `benchmark`.
FROM deps AS vuln

ARG GOVULNCHECK_VERSION=latest
RUN go install golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}
COPY . .

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
# Digest ermitteln via `docker buildx imagetools inspect gcr.io/distroless/static-debian12:nonroot`.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab AS runtime

COPY --from=build /out/api /usr/local/bin/api
COPY --from=build --chown=65532:65532 /staging/var/lib/mtrace /var/lib/mtrace

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
