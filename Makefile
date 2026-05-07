COMPOSE ?= docker compose
PNPM ?= pnpm
API_MAKE ?= $(MAKE) -C apps/api

COVERAGE_THRESHOLD ?= 90
THRESHOLD ?= $(COVERAGE_THRESHOLD)

.DEFAULT_GOAL := help

.PHONY: help dev dev-observability dev-tempo stop wipe smoke smoke-observability smoke-tempo smoke-rak10-console smoke-analyzer smoke-mediamtx smoke-srt smoke-srt-health smoke-dash smoke-webrtc-prep smoke-webrtc-stats-drift smoke-srs smoke-cli seed-rak9 browser-e2e docs-check docs-refs test api-test api-race ts-test lint api-lint ts-lint build api-build ts-build coverage-gate api-coverage-gate ts-coverage-gate coverage-report arch-check sdk-pack-smoke sdk-performance-smoke gates ci install fullbuild sync-contract-fixtures schema-validate schema-generate vuln-check audit-ts image-scan security-gates generated-drift-check

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
		'  make smoke-webrtc-prep      Run the WebRTC lab prep smoke (starts/stops mtrace-webrtc project; endpoint-only)' \
		'  make smoke-webrtc-stats-drift Run the WebRTC getStats() drift smoke against mtrace-webrtc (plan-0.9.0 Tranche 1, RAK-56; opt-in)' \
		'  make smoke-srs              Run the SRS example smoke (starts/stops mtrace-srs project; endpoint-only; plan-0.9.0 Tranche 2, RAK-57)' \
		'  make smoke-cli              Run the m-trace CLI smoke check' \
		'  make sync-contract-fixtures Copy spec/contract-fixtures/analyzer/* to apps/api testdata' \
		'  make seed-rak9              Seed sessions/events for RAK-9 checks' \
		'  make browser-e2e            Run browser E2E checks' \
		'  make docs-check             Run documentation checks' \
		'  make test                   Run API Docker tests and TS workspace tests' \
		'  make api-race               Run API Go race-detector tests (CGO=1, -race; opt-in, in gates)' \
		'  make lint                   Run API Docker lint and TS workspace lint' \
		'  make build                  Build API runtime image and TS workspace packages' \
		'  make coverage-gate          Run API, SDK, dashboard and analyzer coverage gates' \
		'  make coverage-report        Export the API coverage report' \
		'  make arch-check             Run the API architecture boundary check' \
		'  make schema-validate        Validate apps/api schema.yaml via d-migrate' \
		'  make schema-generate        Re-generate apps/api SQLite DDL from schema.yaml' \
		'  make sdk-pack-smoke         Run the Player-SDK pack/public-entry smoke check' \
		'  make sdk-performance-smoke  Run the Player-SDK performance smoke check' \
		'  make vuln-check             Run govulncheck on apps/api Go dependencies (plan-0.8.5 Tranche 1)' \
		'  make audit-ts               Run pnpm audit --audit-level high on the TS workspace (plan-0.8.5 Tranche 1)' \
		'  make image-scan             Run Trivy scan on API/Dashboard/Analyzer runtime images' \
		'  make security-gates         Run vuln-check + audit-ts + image-scan together (plan-0.8.5 Tranche 1)' \
		'  make generated-drift-check  Re-run schema/contract/SDK generators and fail on drift (plan-0.8.5 Tranche 2)' \
		'  make gates                  Run api-race + TS/API quality, SDK smokes, schema and docs gates' \
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

# `make smoke-webrtc-prep` startet das WebRTC-Lab-Beispiel (plan-0.7.0
# §4 Tranche 3, RAK-48) als Project `mtrace-webrtc`: FFmpeg pushed via
# RTSP in MediaMTX, MediaMTX exposed WHIP/WHEP. Smoke ist endpoint-/
# compose-only — prüft API-Erreichbarkeit, Stream-Pfad-Registrierung
# und WHIP/WHEP-OPTIONS-Statuscodes (kein Browser, kein Playback,
# kein getStats). Opt-in (nicht in `make gates`).
smoke-webrtc-prep:
	bash scripts/smoke-webrtc-prep.sh

# `make smoke-webrtc-stats-drift` ist der Browser-Drift-Smoke aus
# plan-0.9.0 §2 Tranche 1 (RAK-56). Schließt R-12 als „automatisiert
# detektiert" — fährt das mtrace-webrtc-Lab hoch, läuft die
# Playwright-Spec tests/e2e/webrtc-stats-drift.spec.ts gegen die
# Default-Browser (chromium,firefox; WebKit opt-in via
# MTRACE_WEBRTC_DRIFT_BROWSERS) und vergleicht das `getStats()`-
# Schema gegen spec/telemetry-model.md §3.5.2 + §1.4. Opt-in
# (NICHT in `make gates`); produktiv über den Nightly-CI-Workflow
# `.github/workflows/webrtc-drift.yml`.
smoke-webrtc-stats-drift:
	bash scripts/smoke-webrtc-stats-drift.sh

# `make smoke-srs` ist der SRS-Lab-Smoke aus plan-0.9.0 §3 Tranche 2
# (RAK-57, MVP-36 als eingelöst). Fährt examples/srs/compose.yaml
# als Project mtrace-srs hoch (RTMP-Listener + HTTP-API +
# HTTP-FLV-Egress), prüft endpoint-/compose-only, dass die SRS-
# HTTP-API erreichbar ist, der FFmpeg-Publisher den Stream
# live/srs-test registriert hat und HTTP-FLV-Egress den FLV-Magic-
# Header liefert. Opt-in (NICHT in `make gates`).
smoke-srs:
	bash scripts/smoke-srs.sh

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
	cp spec/contract-fixtures/analyzer/success-dash-vod.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-vod.json
	cp spec/contract-fixtures/analyzer/success-dash-live.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-live.json
	cp spec/contract-fixtures/analyzer/error-fetch-blocked.json apps/api/adapters/driven/streamanalyzer/testdata/contract-error-fetch-blocked.json
	mkdir -p apps/api/adapters/driven/srt/mediamtxclient/testdata
	cp spec/contract-fixtures/srt/mediamtx-srtconns-list.json apps/api/adapters/driven/srt/mediamtxclient/testdata/mediamtx-srtconns-list.json
	mkdir -p apps/api/adapters/driving/http/testdata
	cp spec/contract-fixtures/api/srt-health-detail.json apps/api/adapters/driving/http/testdata/srt-health-detail.json
	@echo "[sync-contract-fixtures] copied 6 fixture(s) into apps/api/.../testdata/"

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

# Opt-in Race-Detector-Lauf für apps/api (Go Race Detector,
# `go test -race ./...`). Bewusst nicht Teil von `make test` /
# `make gates` — 5–10× langsamer und nur sinnvoll, wenn
# Concurrency-Code geändert wurde. Lauf vor Tag-Push, siehe
# `docs/user/releasing.md` §2 Smoke-Block.
api-race:
	$(API_MAKE) race

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

sdk-pack-smoke:
	$(PNPM) --filter @npm9912/player-sdk run pack:smoke

gates: api-race ts-test lint coverage-gate arch-check schema-validate generated-drift-check sdk-pack-smoke sdk-performance-smoke docs-check

# plan-0.8.5 Tranche 1 — Quality-Gates Wave 1. Security-Gates laufen
# parallel zu `make gates` (separater CI-Job in build.yml), nicht in
# `make gates`-Pipeline integriert: Vulnerability-Datenbank-Download
# kann lokal 30-60 s dauern, sollte den schnellen Inner-Loop nicht
# blockieren.

# govulncheck-Version explizit gepinnt (analog d-migrate-Image-Pin).
# v1.1.4 ist die letzte stable mit Go 1.26-Kompatibilitaet.
GOVULNCHECK_VERSION ?= v1.1.4

# Trivy-Image gepinnt (analog d-migrate-Image-Pin). 0.59.1 ist die
# stable Linie mit guter Default-Policy fuer CRITICAL/HIGH.
TRIVY_IMAGE ?= aquasec/trivy:0.59.1

# `make vuln-check` prueft Go-Dependencies in apps/api gegen die
# Go Vulnerability Database (https://pkg.go.dev/vuln/). govulncheck
# scannt nur tatsaechlich aufgerufene Funktionen — False-Positive-
# Rate ist niedriger als bei statischen Tools.
vuln-check:
	docker run --rm -v "$(CURDIR)/apps/api:/src" -w /src golang:1.26 \
		bash -c "go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) && govulncheck ./..."

# `make audit-ts` prueft die npm-Dependency-Closure des pnpm-Workspaces
# (apps/dashboard, apps/analyzer-service, packages/*) gegen den GitHub
# Advisory Feed. Schwelle = high — moderate/low werden lediglich
# berichtet, brechen aber den Lauf nicht. Pendant zu vuln-check fuer
# die TypeScript-Seite; ohne diesen Gate wuerde eine bekannte CVE in
# einer Frontend-/SDK-Dependency die Security-Wave bestehen.
audit-ts:
	$(PNPM) audit --audit-level high

# `make image-scan` baut die drei Runtime-Images und scannt sie mit
# Trivy. Policy: CRITICAL und HIGH brechen den Lauf; MEDIUM wird
# berichtet. Cache-Verzeichnis liegt unter .security/.trivy-cache,
# damit lokale Wiederholungen nicht jedes Mal die Vuln-DB neu laden.
#
# Dashboard- und Analyzer-Service-Images brauchen TS-Build-Artefakte
# (`pnpm run build` in den jeweiligen Workspaces). Wir bauen sie hier
# explizit, weil `make build` bislang nur api-build + ts-build
# ausfuehrt, nicht die Multi-Stage-Container fuer dashboard/analyzer-
# service.
image-scan:
	docker build --target runtime -t mtrace-api:scan apps/api
	$(PNPM) run build
	# Dashboard- und Analyzer-Service-Images referenzieren in
	# `COPY packages/...` und `COPY apps/.../package.json` Pfade
	# außerhalb von `apps/<svc>/`. Build-Context muss daher der
	# Repo-Root sein; das Dockerfile wird über `-f` adressiert.
	docker build -f apps/dashboard/Dockerfile -t mtrace-dashboard:scan .
	docker build -f apps/analyzer-service/Dockerfile -t mtrace-analyzer-service:scan .
	mkdir -p .security/.trivy-cache
	# `.security/.trivyignore` wird pro Image aus
	# `.security/vulnignore.yaml` generiert (single-source-of-truth +
	# audit trail). Der Generator bricht ab, falls ein Eintrag das
	# `expires`-Datum ueberschritten hat — Wartungsregel laut
	# plan-0.8.5 §2. Scope-Filterung verhindert, dass ein CVE-Ignore
	# fuer ein Runtime-Image global alle Image-Scans maskiert.
	bash scripts/render-trivyignore.sh mtrace-api
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v "$(CURDIR)/.security/.trivy-cache:/root/.cache/trivy" \
		-v "$(CURDIR)/.security/.trivyignore:/work/.trivyignore:ro" \
		$(TRIVY_IMAGE) image \
		--severity CRITICAL,HIGH \
		--exit-code 1 \
		--no-progress \
		--ignorefile /work/.trivyignore \
		mtrace-api:scan
	bash scripts/render-trivyignore.sh mtrace-dashboard
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v "$(CURDIR)/.security/.trivy-cache:/root/.cache/trivy" \
		-v "$(CURDIR)/.security/.trivyignore:/work/.trivyignore:ro" \
		$(TRIVY_IMAGE) image \
		--severity CRITICAL,HIGH \
		--exit-code 1 \
		--no-progress \
		--ignorefile /work/.trivyignore \
		mtrace-dashboard:scan
	bash scripts/render-trivyignore.sh mtrace-analyzer-service
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v "$(CURDIR)/.security/.trivy-cache:/root/.cache/trivy" \
		-v "$(CURDIR)/.security/.trivyignore:/work/.trivyignore:ro" \
		$(TRIVY_IMAGE) image \
		--severity CRITICAL,HIGH \
		--exit-code 1 \
		--no-progress \
		--ignorefile /work/.trivyignore \
		mtrace-analyzer-service:scan

security-gates: vuln-check audit-ts image-scan

# `make generated-drift-check` ruft die drei Generierungs-/Sync-
# Pfade auf und stellt sicher, dass keine erzeugten Artefakte vom
# committeten Stand abweichen. Bei Drift wird der konkrete
# Regenerier-Befehl pro Pfad gemeldet, damit der Fix nicht raten
# muss, welches Target die Quelle ist. Ohne Netzwerk lauffähig,
# sobald die `d-migrate`- und `golang:1.26`-Images lokal gepullt
# sind (CI-Cache trägt das mit).
#
# Geprüfte Artefakte (Single-Source-of-Truth links, Generated rechts):
#   - schema.yaml             → migrations/V1__m_trace.sql
#   - spec/contract-fixtures/ → apps/api/.../testdata/contract-*.json
#                               apps/api/.../testdata/mediamtx-*.json
#                               apps/api/.../testdata/srt-health-*.json
#   - packages/player-sdk/src/index.ts → public-api.snapshot.txt
#     (check-public-api.mjs ist read-only und exited bei Drift mit 1;
#     deshalb separater Aufruf, kein git-diff danach.)
generated-drift-check:
	@echo "[drift-check] Re-generating schema DDL (V1__m_trace.sql)..."
	@$(MAKE) --no-print-directory schema-generate >/dev/null
	@echo "[drift-check] Re-syncing contract fixtures..."
	@$(MAKE) --no-print-directory sync-contract-fixtures >/dev/null
	@echo "[drift-check] Verifying public API snapshot..."
	@$(PNPM) --filter @npm9912/player-sdk exec node scripts/check-public-api.mjs
	@echo "[drift-check] Verifying working tree is clean for generated paths..."
	@# `git diff --exit-code HEAD -- ...` vergleicht Working-Tree gegen
	@# HEAD (nicht gegen den Index), damit ein vorzeitiges `git add`
	@# einen Drift nicht maskiert. CI mit shallow checkout (depth=1)
	@# hat HEAD verfügbar.
	@if ! git diff --exit-code HEAD -- \
		apps/api/internal/storage/migrations/V1__m_trace.sql \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-master.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-vod.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-live.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-error-fetch-blocked.json \
		apps/api/adapters/driven/srt/mediamtxclient/testdata/mediamtx-srtconns-list.json \
		apps/api/adapters/driving/http/testdata/srt-health-detail.json; then \
		echo ""; \
		echo "Generated artifacts are out of sync with their sources."; \
		echo "  - schema DDL (V1__m_trace.sql)        --> run: make schema-generate"; \
		echo "  - api/streamanalyzer/testdata/*.json   --> run: make sync-contract-fixtures"; \
		echo "  - api/srt/mediamtxclient/testdata/...  --> run: make sync-contract-fixtures"; \
		echo "  - api/driving/http/testdata/...        --> run: make sync-contract-fixtures"; \
		echo "  - player-sdk public API snapshot       --> update packages/player-sdk/scripts/public-api.snapshot.txt"; \
		echo ""; \
		echo "Re-run 'make generated-drift-check' afterwards to verify."; \
		exit 1; \
	fi
	@echo "[drift-check] OK -- no drift detected."

ci: gates build

install:
	$(PNPM) install --frozen-lockfile

# fullbuild ist der kanonische End-zu-End-Lauf vom frischen Clone:
# Dependencies installieren, alles bauen (workspace + api Docker)
# und alle Gates laufen lassen. Spiegelt das, was CI ausführt.
fullbuild: install ci
