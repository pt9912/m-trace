COMPOSE ?= docker compose
PNPM ?= pnpm
API_MAKE ?= $(MAKE) -C apps/api

COVERAGE_THRESHOLD ?= 90
THRESHOLD ?= $(COVERAGE_THRESHOLD)

.DEFAULT_GOAL := help

.PHONY: help dev dev-observability dev-tempo stop wipe smoke smoke-observability smoke-tempo smoke-rak10-console smoke-analyzer smoke-mediamtx smoke-srt smoke-srt-health smoke-dash smoke-cli seed-rak9 browser-e2e docs-check docs-refs test api-test ts-test lint api-lint ts-lint build api-build ts-build coverage-gate api-coverage-gate ts-coverage-gate coverage-report arch-check sdk-performance-smoke gates ci install fullbuild sync-contract-fixtures schema-validate schema-generate

help:
	@printf '%s\n' \
		'Targets:' \
		'  make dev                    Start the core Docker Compose lab' \
		'  make dev-observability      Start the lab with the observability profile' \
		'  make dev-tempo              Start observability + Tempo profiles (RAK-31)' \
		'  make stop                   Stop all Compose services, including observability + tempo' \
		'  make wipe                   Stop services AND delete SQLite + Tempo volumes (destructive)' \
		'  make smoke                  Run the local 0.1.1 smoke checks' \
		'  make smoke-observability    Run the Prometheus/cardinality smoke checks' \
		'  make smoke-tempo            Run the Tempo three-state smoke check' \
		'  make smoke-rak10-console    Run the console-trace smoke check' \
		'  make smoke-analyzer         Run the analyzer-service smoke check' \
		'  make smoke-mediamtx         Run the MediaMTX example smoke check (needs make dev)' \
		'  make smoke-srt              Run the SRT example smoke (starts/stops mtrace-srt project)' \
		'  make smoke-srt-health       Run the SRT health smoke (HLS + MediaMTX-API; plan-0.6.0 Tranche 2)' \
		'  make smoke-dash             Run the DASH example smoke (starts/stops mtrace-dash project)' \
		'  make smoke-cli              Run the m-trace CLI smoke check' \
		'  make sync-contract-fixtures Copy spec/contract-fixtures/analyzer/* to apps/api testdata' \
		'  make seed-rak9              Seed sessions/events for RAK-9 checks' \
		'  make browser-e2e            Run browser E2E checks' \
		'  make docs-check             Run documentation checks' \
		'  make test                   Run API Docker tests and TS workspace tests' \
		'  make lint                   Run API Docker lint and TS workspace lint' \
		'  make build                  Build API runtime image and TS workspace packages' \
		'  make coverage-gate          Run API, SDK, dashboard and analyzer coverage gates' \
		'  make coverage-report        Export the API coverage report' \
		'  make arch-check             Run the API architecture boundary check' \
		'  make schema-validate        Validate apps/api schema.yaml via d-migrate' \
		'  make schema-generate        Re-generate apps/api SQLite DDL from schema.yaml' \
		'  make sdk-performance-smoke  Run the Player-SDK performance smoke check' \
		'  make gates                  Run test, lint, coverage, architecture, schema and docs gates' \
		'  make ci                     Run gates plus build' \
		'  make install                pnpm install --frozen-lockfile' \
		'  make fullbuild              Install + ts/api build + gates (CI-äquivalent von clean)' \
		'' \
		'Variables:' \
		'  COMPOSE="docker compose" PNPM=pnpm API_MAKE="$(MAKE) -C apps/api"' \
		'  COVERAGE_THRESHOLD=90 THRESHOLD=$(THRESHOLD)'

dev:
	$(COMPOSE) up --build

dev-observability:
	OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317 OTEL_EXPORTER_OTLP_PROTOCOL=grpc OTEL_TRACES_EXPORTER=otlp OTEL_METRICS_EXPORTER=otlp $(COMPOSE) --profile observability up --build

# `make dev-tempo` startet observability + tempo gemeinsam (plan-0.4.0
# §6). Der Collector-Service ist nicht doppelt definiert; derselbe
# Container fährt in §6.3 mit der Tempo-Pipeline-Konfig hoch (env-
# gesteuerter Config-Pfad). RAK-31 ist Kann-Scope — ohne Profil
# bleibt die Dashboard-Timeline (RAK-32) vollständig funktional.
dev-tempo:
	OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317 OTEL_EXPORTER_OTLP_PROTOCOL=grpc OTEL_TRACES_EXPORTER=otlp OTEL_METRICS_EXPORTER=otlp COLLECTOR_CONFIG=config-tempo.yaml $(COMPOSE) --profile observability --profile tempo up --build

stop:
	$(COMPOSE) --profile observability --profile tempo down

# `make wipe` ist der einzige unterstützte Reset-Pfad für die SQLite-
# Datei (ADR-0002 §8.4, API-Kontrakt §10.1). Stoppt zuerst alle
# Services (sonst bleibt das Volume vom api-Container in Benutzung)
# und entfernt anschließend gezielt das `mtrace-data`-Volume.
#
# Gezieltes Targeting (statt `down --volumes`), damit später
# hinzukommende benannte Volumes (z. B. `m-trace_postgres-data`)
# nicht versehentlich mitgewipt werden. `COMPOSE_PROJECT_NAME` wird
# berücksichtigt, damit Override-Stacks ihre eigenen Volumes wipen.
WIPE_COMPOSE_PROJECT ?= $(if $(COMPOSE_PROJECT_NAME),$(COMPOSE_PROJECT_NAME),$(shell basename $(CURDIR)))
WIPE_VOLUME ?= $(WIPE_COMPOSE_PROJECT)_mtrace-data
WIPE_TEMPO_VOLUME ?= $(WIPE_COMPOSE_PROJECT)_mtrace-tempo-data
wipe:
	@echo "[wipe] destructive: removing volumes $(WIPE_VOLUME) and $(WIPE_TEMPO_VOLUME)"
	@echo "[wipe] sessions, events and Tempo traces will be lost"
	$(COMPOSE) --profile observability --profile tempo down
	docker volume rm "$(WIPE_VOLUME)" 2>/dev/null || \
		echo "[wipe] volume $(WIPE_VOLUME) not present (already wiped or never started)"
	docker volume rm "$(WIPE_TEMPO_VOLUME)" 2>/dev/null || \
		echo "[wipe] volume $(WIPE_TEMPO_VOLUME) not present (already wiped or never started)"

smoke:
	bash scripts/smoke-0.1.1.sh

smoke-observability:
	bash scripts/smoke-observability.sh

# `make smoke-tempo` deckt die drei Startzustände aus plan-0.4.0 §6.4
# ab. Default-State ist `tempo` (RAK-31-Roundtrip via Tempo-Search-API);
# `core` und `observability` lassen sich über `SMOKE_STATE=...` testen
# (Stack vorher mit `make dev` bzw. `make dev-observability` starten).
smoke-tempo:
	bash scripts/smoke-tempo.sh

smoke-rak10-console:
	OTEL_TRACES_EXPORTER=console $(COMPOSE) up -d --build api
	bash scripts/smoke-rak10-console.sh

smoke-analyzer:
	$(COMPOSE) up -d --build analyzer-service api mediamtx stream-generator
	bash scripts/smoke-analyzer.sh

# `make smoke-mediamtx` verifiziert den bestehenden Core-Lab-MediaMTX-
# Pfad (plan-0.5.0 §3 Tranche 2, RAK-36): MediaMTX-API erreichbar,
# teststream ready, HLS-Manifest auflösbar. Erwartet ein laufendes
# Core-Lab (`make dev`) — der Smoke startet nichts selbst, damit der
# Operator entscheidet, welche Stack-Variante er prüfen will.
# Opt-in (nicht in `make gates`).
smoke-mediamtx:
	bash scripts/smoke-mediamtx.sh

# `make smoke-srt` startet das eigene SRT-Beispiel (plan-0.5.0 §4
# Tranche 3, RAK-37) als Project `mtrace-srt`, prüft den HLS-Pfad
# auf 8889 (FFmpeg→SRT→MediaMTX→HLS) und beendet den Stack wieder.
# Opt-in (nicht in `make gates`).
smoke-srt:
	bash scripts/smoke-srt.sh

# `make smoke-srt-health` erweitert smoke-srt um eine API-Probe gegen
# MediaMTX `/v3/srtconns/list` (plan-0.6.0 §3 Tranche 2). Verifiziert
# zusätzlich vier RAK-43-Pflichtwerte (msRTT, packetsReceivedLoss,
# packetsReceivedRetrans, mbpsLinkCapacity > 0). Opt-in (nicht in
# `make gates`); braucht python3 für JSON-Validierung.
smoke-srt-health:
	bash scripts/smoke-srt-health.sh

# `make smoke-dash` startet das DASH-Beispiel (plan-0.5.0 §5 Tranche 4,
# RAK-38) als Project `mtrace-dash`: FFmpeg generiert DASH in ein
# Volume, nginx serviert es auf 8891. Smoke prüft MPD + Init-Segment
# und beendet den Stack wieder. Opt-in (nicht in `make gates`).
smoke-dash:
	bash scripts/smoke-dash.sh

# smoke-cli verifiziert den Lastenheft-Aufruf `pnpm m-trace check <url>`
# (plan-0.3.0 §8 Tranche 7). Hängt am ts-build, damit das CLI-
# Bundle (packages/stream-analyzer/dist/cli/main.cjs) vorliegt; ein
# zweiter `pnpm install` materialisiert die Bin-Symlinks (workspace-
# Pakete können das beim ersten Install nicht, wenn `dist/` noch
# fehlt — gleiches Verhalten in CI und auf frischen Clones).
smoke-cli: ts-build
	$(PNPM) install --frozen-lockfile
	bash scripts/smoke-cli.sh

# Spec ist die Quelle der Wahrheit; Go-Tests konsumieren Kopien aus
# apps/api/.../testdata/, weil der api-Docker-Build-Context nur
# apps/api/ kennt. `make sync-contract-fixtures` kopiert die
# Spec-Dateien in den Go-Pfad — manueller Trigger, weil derselbe
# TS-Test (ts-test) den Drift bereits hart prüft.
sync-contract-fixtures:
	cp spec/contract-fixtures/analyzer/success-master.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-master.json
	cp spec/contract-fixtures/analyzer/error-fetch-blocked.json apps/api/adapters/driven/streamanalyzer/testdata/contract-error-fetch-blocked.json
	mkdir -p apps/api/adapters/driven/srt/mediamtxclient/testdata
	cp spec/contract-fixtures/srt/mediamtx-srtconns-list.json apps/api/adapters/driven/srt/mediamtxclient/testdata/mediamtx-srtconns-list.json
	@echo "[sync-contract-fixtures] copied 3 fixture(s) into apps/api/.../testdata/"

seed-rak9:
	bash scripts/seed-rak9.sh

browser-e2e:
	bash scripts/test-browser-e2e.sh

docs-check:
	bash scripts/verify-doc-refs.sh

docs-refs: docs-check

test: api-test ts-test

api-test:
	$(API_MAKE) test

# Workspace-Pakete mit pnpm-Workspace-Deps (analyzer-service →
# stream-analyzer) brauchen die `dist/`-Artefakte ihrer Dependencies,
# bevor Tests/Lint/Coverage laufen können. `pnpm -r run build`
# respektiert den Topo-Sort und baut Dependencies vor Consumern; wir
# binden den Build deshalb als harte Vorbedingung ein.
ts-test: ts-build
	$(PNPM) run test

lint: api-lint ts-lint

api-lint:
	$(API_MAKE) lint

ts-lint: ts-build
	$(PNPM) run lint

build: api-build ts-build

api-build:
	$(API_MAKE) build

ts-build:
	$(PNPM) run build

coverage-gate: api-coverage-gate ts-coverage-gate

api-coverage-gate:
	$(API_MAKE) coverage-gate THRESHOLD="$(THRESHOLD)"

ts-coverage-gate: ts-build
	$(PNPM) --filter @npm9912/player-sdk run test:coverage
	$(PNPM) --filter @npm9912/m-trace-dashboard run test:coverage
	$(PNPM) --filter @npm9912/stream-analyzer run test:coverage
	$(PNPM) --filter @npm9912/analyzer-service run test:coverage

coverage-report:
	$(API_MAKE) coverage-report THRESHOLD="$(THRESHOLD)"

arch-check:
	$(API_MAKE) arch-check

schema-validate:
	$(API_MAKE) schema-validate

schema-generate:
	$(API_MAKE) schema-generate

sdk-performance-smoke:
	$(PNPM) --filter @npm9912/player-sdk run performance:smoke

gates: test lint coverage-gate arch-check schema-validate docs-check

ci: gates build

install:
	$(PNPM) install --frozen-lockfile

# fullbuild ist der kanonische End-zu-End-Lauf vom frischen Clone:
# Dependencies installieren, alles bauen (workspace + api Docker)
# und alle Gates laufen lassen. Spiegelt das, was CI ausführt.
fullbuild: install ci
