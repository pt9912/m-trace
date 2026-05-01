COMPOSE ?= docker compose
PNPM ?= pnpm
API_MAKE ?= $(MAKE) -C apps/api

COVERAGE_THRESHOLD ?= 90
THRESHOLD ?= $(COVERAGE_THRESHOLD)

.DEFAULT_GOAL := help

.PHONY: help dev dev-observability stop smoke smoke-observability smoke-rak10-console smoke-analyzer smoke-cli seed-rak9 browser-e2e test api-test workspace-test lint api-lint workspace-lint build api-build workspace-build coverage-gate api-coverage-gate workspace-coverage-gate coverage-report arch-check sdk-performance-smoke gates ci install fullbuild sync-contract-fixtures

help:
	@printf '%s\n' \
		'Targets:' \
		'  make dev                    Start the core Docker Compose lab' \
		'  make dev-observability      Start the lab with the observability profile' \
		'  make stop                   Stop all Compose services, including observability' \
		'  make smoke                  Run the local 0.1.1 smoke checks' \
		'  make smoke-observability    Run the Prometheus/cardinality smoke checks' \
		'  make smoke-rak10-console    Run the console-trace smoke check' \
		'  make smoke-analyzer         Run the analyzer-service smoke check' \
		'  make smoke-cli              Run the m-trace CLI smoke check' \
		'  make sync-contract-fixtures Copy spec/contract-fixtures/analyzer/* to apps/api testdata' \
		'  make seed-rak9              Seed sessions/events for RAK-9 checks' \
		'  make browser-e2e            Run browser E2E checks' \
		'  make test                   Run API Docker tests and workspace tests' \
		'  make lint                   Run API Docker lint and workspace lint' \
		'  make build                  Build API runtime image and workspace packages' \
		'  make coverage-gate          Run API, SDK, dashboard and analyzer coverage gates' \
		'  make coverage-report        Export the API coverage report' \
		'  make arch-check             Run the API architecture boundary check' \
		'  make sdk-performance-smoke  Run the Player-SDK performance smoke check' \
		'  make gates                  Run test, lint, coverage and architecture gates' \
		'  make ci                     Run gates plus build' \
		'  make install                pnpm install --frozen-lockfile' \
		'  make fullbuild              Install + workspace/api build + gates (CI-äquivalent von clean)' \
		'' \
		'Variables:' \
		'  COMPOSE="docker compose" PNPM=pnpm API_MAKE="$(MAKE) -C apps/api"' \
		'  COVERAGE_THRESHOLD=90 THRESHOLD=$(THRESHOLD)'

dev:
	$(COMPOSE) up --build

dev-observability:
	OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317 OTEL_EXPORTER_OTLP_PROTOCOL=grpc OTEL_TRACES_EXPORTER=otlp OTEL_METRICS_EXPORTER=otlp $(COMPOSE) --profile observability up --build

stop:
	$(COMPOSE) --profile observability down

smoke:
	bash scripts/smoke-0.1.1.sh

smoke-observability:
	bash scripts/smoke-observability.sh

smoke-rak10-console:
	OTEL_TRACES_EXPORTER=console $(COMPOSE) up -d --build api
	bash scripts/smoke-rak10-console.sh

smoke-analyzer:
	$(COMPOSE) up -d --build analyzer-service api mediamtx stream-generator
	bash scripts/smoke-analyzer.sh

# smoke-cli verifiziert den Lastenheft-Aufruf `pnpm m-trace check <url>`
# (plan-0.3.0 §8 Tranche 7). Hängt am workspace-build, damit das CLI-
# Bundle (packages/stream-analyzer/dist/cli/main.cjs) vorliegt.
smoke-cli: workspace-build
	bash scripts/smoke-cli.sh

# Spec ist die Quelle der Wahrheit; Go-Tests konsumieren Kopien aus
# apps/api/.../testdata/, weil der api-Docker-Build-Context nur
# apps/api/ kennt. `make sync-contract-fixtures` kopiert die
# Spec-Dateien in den Go-Pfad — manueller Trigger, weil derselbe
# TS-Test (workspace-test) den Drift bereits hart prüft.
sync-contract-fixtures:
	cp spec/contract-fixtures/analyzer/success-master.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-master.json
	cp spec/contract-fixtures/analyzer/error-fetch-blocked.json apps/api/adapters/driven/streamanalyzer/testdata/contract-error-fetch-blocked.json
	@echo "[sync-contract-fixtures] copied 2 fixture(s) into apps/api/.../streamanalyzer/testdata/"

seed-rak9:
	bash scripts/seed-rak9.sh

browser-e2e:
	bash scripts/test-browser-e2e.sh

test: api-test workspace-test

api-test:
	$(API_MAKE) test

# Workspace-Pakete mit pnpm-Workspace-Deps (analyzer-service →
# stream-analyzer) brauchen die `dist/`-Artefakte ihrer Dependencies,
# bevor Tests/Lint/Coverage laufen können. `pnpm -r run build`
# respektiert den Topo-Sort und baut Dependencies vor Consumern; wir
# binden den Build deshalb als harte Vorbedingung ein.
workspace-test: workspace-build
	$(PNPM) run test

lint: api-lint workspace-lint

api-lint:
	$(API_MAKE) lint

workspace-lint: workspace-build
	$(PNPM) run lint

build: api-build workspace-build

api-build:
	$(API_MAKE) build

workspace-build:
	$(PNPM) run build

coverage-gate: api-coverage-gate workspace-coverage-gate

api-coverage-gate:
	$(API_MAKE) coverage-gate THRESHOLD="$(THRESHOLD)"

workspace-coverage-gate: workspace-build
	$(PNPM) --filter @npm9912/player-sdk run test:coverage
	$(PNPM) --filter @npm9912/m-trace-dashboard run test:coverage
	$(PNPM) --filter @npm9912/stream-analyzer run test:coverage
	$(PNPM) --filter @npm9912/analyzer-service run test:coverage

coverage-report:
	$(API_MAKE) coverage-report THRESHOLD="$(THRESHOLD)"

arch-check:
	$(API_MAKE) arch-check

sdk-performance-smoke:
	$(PNPM) --filter @npm9912/player-sdk run performance:smoke

gates: test lint coverage-gate arch-check

ci: gates build

install:
	$(PNPM) install --frozen-lockfile

# fullbuild ist der kanonische End-zu-End-Lauf vom frischen Clone:
# Dependencies installieren, alles bauen (workspace + api Docker)
# und alle Gates laufen lassen. Spiegelt das, was CI ausführt.
fullbuild: install ci
