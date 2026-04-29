# syntax=docker/dockerfile:1.7

# ---------------------------------------------------------------------------
# m-trace API spike — Go variant
#
# Docker-only workflow per docs/plan-spike.md §14.11. The repository ships
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
FROM golang:1.22 AS deps

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
# golangci-lint with default linters per docs/plan-spike.md §14.9.
FROM golangci/golangci-lint:v1.62-alpine AS lint

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
FROM deps AS coverage

ARG COVERAGE_THRESHOLD=90
ENV COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD}

COPY . .
RUN mkdir -p /out && \
    CGO_ENABLED=0 go test \
        -coverpkg=./hexagon/...,./adapters/... \
        -coverprofile=/out/coverage.out \
        -covermode=atomic \
        ./... && \
    go tool cover -func=/out/coverage.out | tee /out/coverage-func.txt && \
    go tool cover -html=/out/coverage.out -o /out/coverage.html && \
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

# ---- runtime ---------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=build /out/api /usr/local/bin/api

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
