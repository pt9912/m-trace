# m-trace API spike — Go variant
# Docker-only workflow per docs/plan-spike.md §14.11.

IMAGE := m-trace-api-spike

.PHONY: test lint build deps compile run arch-check clean

# Run unit + integration tests inside Docker.
test:
	docker build --target test -t $(IMAGE):go-test .

# Run golangci-lint (Soll, plan §14.9 / §14.11).
lint:
	docker build --target lint -t $(IMAGE):go-lint .

# Build the runtime image (distroless static).
build:
	docker build --target runtime -t $(IMAGE):go .

# Resolve dependencies only (used to size the deps cache layer, plan §7.2).
deps:
	docker build --no-cache --target deps -t $(IMAGE):go-deps .

# Compile-only feedback loop (no tests, no lint).
compile:
	docker build --target compile -t $(IMAGE):go-compile .

# Run the runtime container with port 8080 exposed.
run: build
	docker run --rm -p 8080:8080 --name m-trace-api-go $(IMAGE):go

# Architecture-Boundary-Test (siehe docs/architecture.md §3.4).
# Prüft, dass hexagon/ keine direkten Imports auf Adapter,
# OTel, Prometheus, database/sql oder net/http enthält und die
# Schichtengrenzen domain → application → port respektiert sind.
arch-check:
	docker run --rm -v "$(CURDIR):/src" -w /src golang:1.22 \
		bash scripts/check-architecture.sh

# Remove all spike-related images.
clean:
	-docker rmi $(IMAGE):go $(IMAGE):go-test $(IMAGE):go-lint $(IMAGE):go-compile $(IMAGE):go-deps
