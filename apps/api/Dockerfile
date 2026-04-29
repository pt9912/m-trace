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
