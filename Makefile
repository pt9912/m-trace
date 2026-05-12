COMPOSE ?= docker compose
PNPM ?= pnpm
API_MAKE ?= $(MAKE) -C apps/api
TS_IMAGE ?= m-trace-ts
TS_DOCKER_BUILD ?= docker build -f Dockerfile

COVERAGE_THRESHOLD ?= 90
THRESHOLD ?= $(COVERAGE_THRESHOLD)

.DEFAULT_GOAL := help

.PHONY: help dev dev-observability dev-tempo stop wipe smoke smoke-observability smoke-tempo smoke-rak10-console smoke-analyzer smoke-mediamtx smoke-mediamtx-auth smoke-srt smoke-srt-health smoke-srt-health-pagination smoke-dash smoke-webrtc-prep smoke-webrtc-stats-drift smoke-srs smoke-ingest-control smoke-key-rotation smoke-issuance-replica smoke-issuance-multi-host smoke-origin-rate-limit smoke-vault-approle smoke-kms-skeleton smoke-mediaserver-provision smoke-browser-ingest smoke-outbound-webhook smoke-cli seed-rak9 browser-e2e docs-check docs-refs test api-test api-race ts-test lint api-lint ts-lint build api-build ts-build coverage-gate api-coverage-gate ts-coverage-gate coverage-report arch-check sdk-pack-smoke sdk-performance-smoke k8s-validate devcontainer-validate release-guard release-guard-test gates ci install lock-refresh fullbuild sync-contract-fixtures schema-validate schema-generate vuln-check audit-ts image-scan security-gates generated-drift-check api-benchmark-smoke analyzer-benchmark-smoke benchmark-smoke fuzz-check api-fuzz-check api-mutation-report ts-mutation-report mutation-report

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
		'  make smoke-srt-health-pagination Run the SRT health smoke incl. cursor-pagination probes (plan-0.12.6 Tranche 2 / RAK-86; opt-in)' \
		'  make smoke-dash             Run the DASH example smoke (starts/stops mtrace-dash project)' \
		'  make smoke-webrtc-prep      Run the WebRTC lab prep smoke (starts/stops mtrace-webrtc project; endpoint-only)' \
		'  make smoke-webrtc-stats-drift Run the WebRTC getStats() drift smoke against mtrace-webrtc (plan-0.9.0 Tranche 1, RAK-56; opt-in)' \
		'  make smoke-srs              Run the SRS example smoke (starts/stops mtrace-srs project; endpoint-only; plan-0.9.0 Tranche 2, RAK-57)' \
		'  make smoke-key-rotation     Run the multi-key signing rotation smoke (plan-0.12.5 Tranche 1, RAK-78; opt-in)' \
		'  make smoke-issuance-replica Run the shared-state issuance limiter smoke (plan-0.12.5 Tranche 2, RAK-77; opt-in)' \
		'  make smoke-issuance-multi-host Run the Redis multi-host issuance + origin limiter smoke (plan-0.12.6 Tranche 7, RAK-88; opt-in)' \
		'  make smoke-mediaserver-provision Run the MediaMTX provision smoke (plan-0.12.6 Tranche 9, RAK-87; opt-in)' \
		'  make smoke-vault-approle    Run the Vault AppRole + Kubernetes auth smoke (plan-0.12.6 Tranche 8, RAK-89; opt-in)' \
		'  make smoke-kms-skeleton     Run the KMS skeleton adapter smoke (plan-0.12.6 Tranche 8, RAK-89; opt-in)' \
		'  make smoke-origin-rate-limit Run the origin-/IP-rate-limiter smoke (plan-0.12.6 Tranche 6, RAK-90; opt-in)' \
		'  make smoke-browser-ingest   Run the browser-ingest policy smoke (plan-0.12.5 Tranche 4, RAK-80; opt-in)' \
		'  make smoke-mediamtx-auth    Run the MediaMTX externalAuth bridge smoke (plan-0.12.5 Tranche 5, RAK-81; opt-in)' \
		'  make smoke-outbound-webhook Run the outbound webhook dispatcher smoke (plan-0.12.5 Tranche 5, RAK-82; opt-in)' \
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
		'  make k8s-validate           Validate optional deploy/k8s examples without a cluster' \
		'  make devcontainer-validate  Validate the optional devcontainer seed' \
		'  make release-guard VER=X.Y.Z Run the manual release approval guard in dry-run mode' \
		'  make release-guard-test     Run local release-guard failure-path tests' \
		'  make vuln-check             Run govulncheck on apps/api Go dependencies (plan-0.8.5 Tranche 1)' \
		'  make audit-ts               Run pnpm audit --audit-level high on the TS workspace (plan-0.8.5 Tranche 1)' \
		'  make image-scan             Run Trivy scan on API/Dashboard/Analyzer runtime images' \
		'  make security-gates         Run vuln-check + audit-ts + image-scan together (plan-0.8.5 Tranche 1)' \
		'  make api-benchmark-smoke    Run Go API hot-path benchmarks (plan-0.9.5 Tranche 1, opt-in/observation; not in gates)' \
		'  make analyzer-benchmark-smoke Run TypeScript stream-analyzer hot-path benchmarks (plan-0.9.5 Tranche 1, opt-in/observation)' \
		'  make benchmark-smoke        Run both api- and analyzer-benchmark-smokes (plan-0.9.5 Tranche 1)' \
		'  make api-fuzz-check         Run Go fuzz targets (-fuzztime, default 30s; plan-0.9.5 Tranche 3, opt-in)' \
		'  make fuzz-check             Run all fuzz targets (Go + TS property tests; opt-in)' \
		'  make api-mutation-report    Run Go mutation report for the API pilot module (plan-0.9.5 Tranche 4, opt-in)' \
		'  make ts-mutation-report     Run TS mutation report for the player-sdk pilot module (plan-0.9.5 Tranche 4, opt-in)' \
		'  make mutation-report        Run Go + TS mutation reports (plan-0.9.5 Tranche 4, opt-in/nightly)' \
		'  make generated-drift-check  Re-run schema/contract/SDK generators and fail on drift (plan-0.8.5 Tranche 2)' \
		'  make gates                  Run api-race + TS/API quality, SDK smokes, schema and docs gates' \
		'  make ci                     Run gates plus build' \
		'  make install                Build the TS dependency image without host node_modules' \
		'  make lock-refresh           Update pnpm-lock.yaml in Docker without host node_modules' \
		'  make fullbuild              Install + ts/api build + gates (CI-äquivalent von clean)' \
		'' \
		'Variables:' \
		'  COMPOSE="docker compose" PNPM=pnpm API_MAKE="$(MAKE) -C apps/api" TS_IMAGE=m-trace-ts' \
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

# `make smoke-srt-health-pagination` (plan-0.12.6 Tranche 2 / RAK-86)
# fährt den existierenden `smoke-srt-health`-Pfad mit den Cursor-
# Probe-Sub-Checks (samples_cursor/next_cursor + cursor_invalid_-
# malformed). Setzt SMOKE_INCLUDE_MTRACE_API=1 plus
# MTRACE_SRT_HEALTH_PAGINATION=1 voraus, was der Wrapper hier
# automatisch aktiviert. Opt-in (nicht in `make gates`).
smoke-srt-health-pagination:
	SMOKE_INCLUDE_MTRACE_API=1 MTRACE_SRT_HEALTH_PAGINATION=1 bash scripts/smoke-srt-health.sh

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

# `make smoke-ingest-control` ist der Lab-Smoke aus plan-0.11.0
# Tranche 4 (RAK-69). Erstellt einen Stream über die HTTP-API und
# spielt einen Start-/Ende-Lifecycle-Hook ein; verifiziert
# `accepted:true` und unterschiedliche `event_id`-Werte. Erwartet
# eine erreichbare apps/api (Default `MTRACE_API_URL=http://localhost:8080`)
# und ein gültiges Token (`MTRACE_API_TOKEN=demo-token`).
# Opt-in (NICHT in `make gates`).
smoke-ingest-control:
	bash examples/ingest-control/smoke-lifecycle.sh

# `make smoke-key-rotation` — plan-0.12.5 Tranche 1 / RAK-78
# Multi-Key-Signing-Rotation. Wickelt den End-to-End-
# Rotation-Unit-Test (`TestParseSigningKeysEnv_RotationEndToEnd`)
# in ein reproduzierbares Make-Target ein und prüft das in
# `docs/user/auth.md` §5.3.1 dokumentierte Operator-Workflow-
# Verhalten: Token unter `kid_a` signieren, ACTIVE auf `kid_b`
# umschalten, altes Token muss weiterhin verifizieren. Opt-in
# (NICHT in `make gates`); echte API-Restart-Variante ist
# Folge-Item nach Multi-Replica-Compose-Bedarf (R-17).
smoke-key-rotation:
	bash scripts/smoke-key-rotation.sh

# `make smoke-issuance-replica` — plan-0.12.5 Tranche 2 / RAK-77
# Shared-State-Issuance-Limiter (R-17). Wickelt den End-to-End-
# Sharing-Test (`TestSqliteIssuanceRateLimiter_SharedAcrossInstances`)
# in ein reproduzierbares Make-Target ein: zwei `*sql.DB`-Verbindungen
# auf dieselbe SQLite-Datei (Single-Host + Shared-Volume-Pfad) — eine
# Replica verbraucht das Project-Bucket, die andere muss den nächsten
# Allow als „denied" sehen. Opt-in (NICHT in `make gates`); echte
# Compose-Multi-Container-Variante bleibt Folge-Item.
smoke-issuance-replica:
	bash scripts/smoke-issuance-replica.sh

# `make smoke-issuance-multi-host` — plan-0.12.6 Tranche 7 / RAK-88
# (R-17). Multi-Host-Variante des Shared-State-Limiters: zwei
# RedisIssuanceRateLimiter-Instances teilen sich Buckets ueber
# miniredis-Mock; deckt auch den Refund-Pfad und beide Fail-Modi
# (fail-closed Default + fail-open Opt-in). Opt-in (NICHT in
# `make gates`).
smoke-issuance-multi-host:
	bash scripts/smoke-issuance-multi-host.sh

# `make smoke-mediaserver-provision` — plan-0.12.6 Tranche 9 / RAK-87
# (R-15). Verifiziert den MediaMTX-Adapter (happy/idempotent/auth/
# server-error/unreachable) plus den Use-Case-Pfad (Provision=false/
# true ohne Adapter → disabled). Opt-in (NICHT in `make gates`).
smoke-mediaserver-provision:
	bash scripts/smoke-mediaserver-provision.sh

# `make smoke-vault-approle` — plan-0.12.6 Tranche 8 / RAK-89 (R-20).
# Verifiziert den AppRole- und Kubernetes-Auth-Pfad des Vault-Adapters
# gegen einen `httptest.Server`-Mock (kein echter Vault-Server noetig).
# Opt-in (NICHT in `make gates`).
smoke-vault-approle:
	bash scripts/smoke-vault-approle.sh

# `make smoke-kms-skeleton` — plan-0.12.6 Tranche 8 / RAK-89 (R-20).
# Verifiziert den KMS-Skelett-Adapter mit Stub-Decrypter +
# LabPassThrough-Decrypter. Production-AWS-SDK-Wiring ist Folge-Item.
# Opt-in (NICHT in `make gates`).
smoke-kms-skeleton:
	bash scripts/smoke-kms-skeleton.sh

# `make smoke-origin-rate-limit` — plan-0.12.6 Tranche 6 / RAK-90
# (R-22). Bestaetigt den Origin-Limiter live: drei aufeinander-
# folgende `POST /api/auth/session-tokens` aus derselben Quelle,
# erwartet 201/201/429 mit Body `{"error":"origin_rate_limited"}`.
# Voraussetzung: m-trace-API laeuft mit
# `MTRACE_ORIGIN_RATE_LIMITER=memory` und Capacity ≥ 2. Opt-in
# (NICHT in `make gates`).
smoke-origin-rate-limit:
	bash scripts/smoke-origin-rate-limit.sh

# `make smoke-browser-ingest` — plan-0.12.5 Tranche 4 / RAK-80
# Browser-Ingest-Policy (R-21). Wickelt die End-to-End-Browser-
# Ingest-Tests (`TestBrowserIngest*`) in ein reproduzierbares
# Make-Target ein: Preflight-Verhalten mit und ohne aktivierter
# Policy, POST-Enforcement (Origin-Pin, CSRF). Opt-in (NICHT in
# `make gates`); deckt den Wire-Vertrag aus `auth.md` §5.6.
smoke-browser-ingest:
	bash scripts/smoke-browser-ingest.sh

# `make smoke-mediamtx-auth` — plan-0.12.5 Tranche 5 / RAK-81
# MediaMTX-Auth-Bridge (R-14). Wickelt die Auth-Hook-Tests
# (`TestMediaMTXAuthHook_*`) in ein reproduzierbares Make-Target
# ein — Wire-Vertrag aus `auth.md` §5.7. Echte Compose-Variante
# mit MediaMTX-Container bleibt Folge-Item.
smoke-mediamtx-auth:
	bash scripts/smoke-mediamtx-auth.sh

# `make smoke-outbound-webhook` — plan-0.12.5 Tranche 5 / RAK-82
# Outbound-Webhook-Dispatcher (R-16). Wickelt die Dispatcher-Tests
# (`TestOutboundWebhook_*`) in ein reproduzierbares Make-Target ein:
# HMAC-Signatur, Retry mit Exponential-Backoff, Dead-Letter-Pfad.
# Opt-in (NICHT in `make gates`).
smoke-outbound-webhook:
	bash scripts/smoke-outbound-webhook.sh

# `make api-benchmark-smoke` ist die Go-Hot-Path-Bench-Suite aus
# plan-0.9.5 §2 Tranche 1 (extra-gates.md §3.2). Druckt zuerst die
# Runner-Identifikation (OS, CPU, Go-Stand) damit Budget-Failures
# einordenbar bleiben (Plan-DoD §2-7), dann läuft die Bench-Suite
# in einem golang:1.26-Container über alle `Benchmark*`-Funktionen
# in apps/api/.../**/*_bench_test.go. Initial-Budgets sind in
# `docs/perf/budgets.md` §3 dokumentiert; PR-Blockierung erst
# nach Beobachtungsphase (DoD §2-6).
#
# Workflow: apps/api/Makefile::benchmark-smoke schreibt den Go-
# Bench-Output nach .tmp/bench/api-bench.txt (im Container an
# /src/.tmp gemountet); `scripts/check-bench-budgets.mjs --kind go`
# parst per stdin und prüft Budgets aus §3.
api-benchmark-smoke:
	@bash scripts/print-bench-runner-info.sh
	@mkdir -p .tmp/bench
	$(API_MAKE) benchmark-smoke | tee .tmp/bench/api-bench.txt
	node scripts/check-bench-budgets.mjs --kind go < .tmp/bench/api-bench.txt

# `make analyzer-benchmark-smoke` ist das TS-Pendant aus plan-0.9.5
# §2 Tranche 1 für `@npm9912/stream-analyzer` (extra-gates.md §3.2
# Stream-Analyzer-Kandidaten). Nutzt die eingebaute Vitest-Bench-
# API (`vitest bench --run --config vitest.bench.config.ts`); keine
# zusätzliche Tinybench-Dependency. Initial-Budgets in
# `docs/perf/budgets.md` §4. Opt-in (NICHT in `make gates`).
#
# Workflow: vitest-bench-stdout wird nach
# `.tmp/bench/analyzer-bench.txt` gespiegelt; `scripts/check-bench-
# budgets.mjs --kind ts` parst die Texttabelle und prüft jeden Bench
# gegen das Budget aus §4 (Plan-DoD §2-4 „Budget-Verletzung erzeugt
# eindeutige Fehlermeldung mit Ist/Soll").
analyzer-benchmark-smoke:
	@bash scripts/print-bench-runner-info.sh
	@mkdir -p .tmp/bench
	@bash -o pipefail -c '$(PNPM) --filter @npm9912/stream-analyzer run bench 2>&1 | tee .tmp/bench/analyzer-bench.txt'
	node scripts/check-bench-budgets.mjs --kind ts < .tmp/bench/analyzer-bench.txt

# `make benchmark-smoke` bündelt beide Bench-Smokes in einem
# Aufruf. Plan-DoD §2-3: Wrapper-Target. Bleibt opt-in, bis die
# Beobachtungsphase abgeschlossen und PR-Blockierung in Tranche 1c
# eingeschaltet ist.
benchmark-smoke: api-benchmark-smoke analyzer-benchmark-smoke

# `make api-fuzz-check` ist die Go-Fuzz-Suite aus plan-0.9.5 §4
# Tranche 3 (extra-gates.md §3.5). Läuft alle `Fuzz*`-Targets aus
# `apps/api/.../**/*_fuzz_test.go` sequenziell mit kurzem
# `-fuzztime` (Default 30s; override via `FUZZTIME=120s make
# api-fuzz-check`). Crash-Funde werden von go test fuzz automatisch
# unter `testdata/fuzz/Fuzz<X>/<id>` als deterministische
# Reproduktion abgelegt — beim nächsten regulären `make api-test`-
# Lauf wirken sie als Regression-Tests. Opt-in (NICHT in
# `make gates`).
FUZZTIME ?= 30s
api-fuzz-check:
	@bash scripts/print-bench-runner-info.sh
	$(API_MAKE) fuzz-check FUZZTIME=$(FUZZTIME)

# `make fuzz-check` bündelt Go-Fuzz und die TS-Property-Tests (die
# über `make ts-test` ohnehin laufen, hier als expliziter Aufruf
# für den Tranche-3-Pfad). Plan-DoD §4-3: opt-in (NICHT in
# `make gates`); Nightly-CI hat eigene Längere-Budget-Stage.
fuzz-check: api-fuzz-check ts-test

# `make api-mutation-report` ist der Go-Mutation-Test (plan-0.9.5
# §5 Tranche 4, RAK-Wave-2 / extra-gates.md §3.6). Pilot-Modul:
# `apps/api/hexagon/application/event_meta_validation.go` (gemutiert
# als Teil des `hexagon/application`-Packages). Tool: gremlins
# (Substitution für unmaintainted go-mutesting; Begründung in
# `docs/dev/mutation-testing.md` §1). Output:
#   - `apps/api/.tmp/mutation/api-mutation-report.txt` (stdout-Spiegel)
#   - `apps/api/.tmp/mutation/api-mutation-report.json` (Maschinen-Form)
# Initial nicht-blockierend; opt-in (NICHT in `make gates`). Lokaler
# Lauf zieht das gremlins-CLI per `go install` zur Laufzeit, daher
# Netz erforderlich (selbe Mechanik wie `benchmark-smoke`).
api-mutation-report:
	$(API_MAKE) mutation-report

# `make ts-mutation-report` ist das TS-Pendant. Pilot-Modul:
# `packages/player-sdk/src/adapters/webrtc/sampling.ts`. Tool:
# StrykerJS via `pnpm dlx` (kein devDep im player-sdk-Manifest, damit
# der Stryker-Versions-Bump nicht im Lockfile pinned). Vitest-Runner
# (selbe Vitest-Version wie `make ts-test`). Output:
#   - `packages/player-sdk/reports/mutation/mutation.html` (visuell)
#   - `packages/player-sdk/reports/mutation/mutation.json` (Trend-Tracking)
# Initial nicht-blockierend; opt-in. Lokaler Lauf braucht Node + pnpm
# (host-side, kein Container — selbe Voraussetzung wie `make ts-test`).
ts-mutation-report:
	$(PNPM) --filter @npm9912/player-sdk run mutation

# `make mutation-report` bündelt Go + TS in einem Aufruf
# (Plan-DoD §5-2 Wrapper). Bleibt opt-in.
mutation-report: api-mutation-report ts-mutation-report

# smoke-cli verifiziert den Lastenheft-Aufruf `pnpm m-trace check <url>`
# (plan-0.3.0 §8 Tranche 7). Der Lauf passiert im Root-Dockerfile,
# damit weder `node_modules` noch `.pnpm-store` im Host-Workspace
# entstehen.
smoke-cli:
	$(TS_DOCKER_BUILD) --target cli-smoke -t $(TS_IMAGE):cli-smoke .

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
	cp spec/contract-fixtures/analyzer/success-hls-cmaf-vod.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-cmaf-vod.json
	cp spec/contract-fixtures/analyzer/success-hls-ts-negative.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-ts-negative.json
	cp spec/contract-fixtures/analyzer/success-hls-master-codecs-only.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-master-codecs-only.json
	cp spec/contract-fixtures/analyzer/success-hls-map-byterange.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-map-byterange.json
	cp spec/contract-fixtures/analyzer/success-hls-media-byterange.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-media-byterange.json
	cp spec/contract-fixtures/analyzer/success-dash-mp4-mime-only.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-mp4-mime-only.json
	cp spec/contract-fixtures/analyzer/success-dash-cmaf-vod.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-cmaf-vod.json
	cp spec/contract-fixtures/analyzer/success-dash-no-cmaf-signals.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-no-cmaf-signals.json
	cp spec/contract-fixtures/analyzer/success-dash-baseurl-inheritance.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-baseurl-inheritance.json
	cp spec/contract-fixtures/analyzer/success-dash-segmentlist.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-segmentlist.json
	cp spec/contract-fixtures/analyzer/error-cmaf-binary-validation.json apps/api/adapters/driven/streamanalyzer/testdata/contract-error-cmaf-binary-validation.json
	cp spec/contract-fixtures/analyzer/error-cmaf-invalid-box-structure.json apps/api/adapters/driven/streamanalyzer/testdata/contract-error-cmaf-invalid-box-structure.json
	cp spec/contract-fixtures/analyzer/success-cmaf-skipped-too-large.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-too-large.json
	cp spec/contract-fixtures/analyzer/success-cmaf-skipped-content-type.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-content-type.json
	cp spec/contract-fixtures/analyzer/success-cmaf-skipped-binary-disabled.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-binary-disabled.json
	cp spec/contract-fixtures/analyzer/success-cmaf-skipped-not-planned.json apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-not-planned.json
	mkdir -p apps/api/adapters/driven/srt/mediamtxclient/testdata
	cp spec/contract-fixtures/srt/mediamtx-srtconns-list.json apps/api/adapters/driven/srt/mediamtxclient/testdata/mediamtx-srtconns-list.json
	mkdir -p apps/api/adapters/driving/http/testdata
	cp spec/contract-fixtures/api/srt-health-detail.json apps/api/adapters/driving/http/testdata/srt-health-detail.json
	cp spec/contract-fixtures/api/srt-health-cursor-invalid-legacy.json apps/api/adapters/driving/http/testdata/srt-health-cursor-invalid-legacy.json
	cp spec/contract-fixtures/api/srt-health-cursor-invalid-malformed.json apps/api/adapters/driving/http/testdata/srt-health-cursor-invalid-malformed.json
	cp spec/contract-fixtures/api/ingest-stream-create.json apps/api/adapters/driving/http/testdata/ingest-stream-create.json
	cp spec/contract-fixtures/api/ingest-stream-list.json apps/api/adapters/driving/http/testdata/ingest-stream-list.json
	cp spec/contract-fixtures/api/ingest-stream-rotate.json apps/api/adapters/driving/http/testdata/ingest-stream-rotate.json
	cp spec/contract-fixtures/api/ingest-stream-validate-blind.json apps/api/adapters/driving/http/testdata/ingest-stream-validate-blind.json
	cp spec/contract-fixtures/api/ingest-error-unauthorized.json apps/api/adapters/driving/http/testdata/ingest-error-unauthorized.json
	cp spec/contract-fixtures/api/ingest-error-stream-not-found.json apps/api/adapters/driving/http/testdata/ingest-error-stream-not-found.json
	cp spec/contract-fixtures/api/ingest-lifecycle-hook-success.json apps/api/adapters/driving/http/testdata/ingest-lifecycle-hook-success.json
	cp spec/contract-fixtures/api/ingest-lifecycle-hook-error-disabled.json apps/api/adapters/driving/http/testdata/ingest-lifecycle-hook-error-disabled.json
	cp spec/contract-fixtures/api/auth-session-token-issue.json apps/api/adapters/driving/http/testdata/auth-session-token-issue.json
	cp spec/contract-fixtures/api/auth-error-token-expired.json apps/api/adapters/driving/http/testdata/auth-error-token-expired.json
	cp spec/contract-fixtures/api/auth-error-policy-denied.json apps/api/adapters/driving/http/testdata/auth-error-policy-denied.json
	cp spec/contract-fixtures/api/auth-error-ttl-too-large.json apps/api/adapters/driving/http/testdata/auth-error-ttl-too-large.json
	cp spec/contract-fixtures/api/auth-error-issuance-rate-limited.json apps/api/adapters/driving/http/testdata/auth-error-issuance-rate-limited.json
	cp spec/contract-fixtures/api/auth-project-token-generation.json apps/api/adapters/driving/http/testdata/auth-project-token-generation.json
	@echo "[sync-contract-fixtures] copied 38 fixture(s) into apps/api/.../testdata/"

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

# Workspace-Pakete mit pnpm-Workspace-Deps (analyzer-service ->
# stream-analyzer) brauchen die `dist/`-Artefakte ihrer Dependencies,
# bevor Tests/Lint/Coverage laufen koennen. Die TS-Stages im Root-
# Dockerfile kapseln install/build/test/lint vollstaendig in Docker,
# damit der Host-Workspace keinen `node_modules`-Baum braucht.
# Pre-Setup für `pnpm install --frozen-lockfile --ignore-scripts`
# läuft über `make install` (Root-Target); Lockfile-Updates für
# neue Dependencies über `make lock-refresh`.
ts-test:
	$(TS_DOCKER_BUILD) --target test -t $(TS_IMAGE):test .

lint: api-lint ts-lint

api-lint:
	$(API_MAKE) lint

ts-lint:
	$(TS_DOCKER_BUILD) --target lint -t $(TS_IMAGE):lint .

build: api-build ts-build

api-build:
	$(API_MAKE) build

ts-build:
	$(TS_DOCKER_BUILD) --target build -t $(TS_IMAGE):build .

coverage-gate: api-coverage-gate ts-coverage-gate

api-coverage-gate:
	$(API_MAKE) coverage-gate THRESHOLD="$(THRESHOLD)"

ts-coverage-gate:
	$(TS_DOCKER_BUILD) --target coverage -t $(TS_IMAGE):coverage .

coverage-report:
	$(API_MAKE) coverage-report THRESHOLD="$(THRESHOLD)"

arch-check:
	$(API_MAKE) arch-check

schema-validate:
	$(API_MAKE) schema-validate

schema-generate:
	$(API_MAKE) schema-generate

sdk-performance-smoke:
	$(TS_DOCKER_BUILD) --target sdk-performance-smoke -t $(TS_IMAGE):sdk-performance-smoke .

sdk-pack-smoke:
	$(TS_DOCKER_BUILD) --target sdk-pack-smoke -t $(TS_IMAGE):sdk-pack-smoke .

release-guard:
	@test -n "$(VER)" || (echo 'release-guard: set VER=X.Y.Z' >&2; exit 2)
	MTRACE_RELEASE_DRY_RUN=1 bash scripts/release-guard.sh "$(VER)"

k8s-validate:
	bash scripts/validate-k8s-examples.sh

devcontainer-validate:
	bash scripts/validate-devcontainer.sh

release-guard-test:
	bash scripts/test-release-guard.sh

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
	docker run --rm -v "$(CURDIR)/apps/api:/src" -w /src golang:1.26.3 \
		bash -c "go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) && govulncheck ./..."

# `make audit-ts` prueft die npm-Dependency-Closure des pnpm-Workspaces
# (apps/dashboard, apps/analyzer-service, packages/*) gegen den GitHub
# Advisory Feed. Schwelle = high — moderate/low werden lediglich
# berichtet, brechen aber den Lauf nicht. Pendant zu vuln-check fuer
# die TypeScript-Seite; ohne diesen Gate wuerde eine bekannte CVE in
# einer Frontend-/SDK-Dependency die Security-Wave bestehen.
audit-ts:
	$(TS_DOCKER_BUILD) --target audit -t $(TS_IMAGE):audit .

# `make image-scan` baut die drei Runtime-Images und scannt sie mit
# Trivy. Policy: CRITICAL und HIGH brechen den Lauf; MEDIUM wird
# berichtet. Cache-Verzeichnis liegt unter .security/.trivy-cache,
# damit lokale Wiederholungen nicht jedes Mal die Vuln-DB neu laden.
#
# Dashboard- und Analyzer-Service-Images bauen ihre TS-Artefakte in den
# eigenen Multi-Stage-Dockerfiles. Host-seitige pnpm-Artefakte sind fuer
# den Image-Scan nicht erforderlich.
image-scan:
	docker build --target runtime -t mtrace-api:scan apps/api
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
	@$(TS_DOCKER_BUILD) --target public-api-check -t $(TS_IMAGE):public-api-check . >/dev/null
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
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-cmaf-vod.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-ts-negative.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-master-codecs-only.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-map-byterange.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-hls-media-byterange.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-mp4-mime-only.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-cmaf-vod.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-no-cmaf-signals.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-baseurl-inheritance.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-dash-segmentlist.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-error-cmaf-binary-validation.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-error-cmaf-invalid-box-structure.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-too-large.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-content-type.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-binary-disabled.json \
		apps/api/adapters/driven/streamanalyzer/testdata/contract-success-cmaf-skipped-not-planned.json \
		apps/api/adapters/driven/srt/mediamtxclient/testdata/mediamtx-srtconns-list.json \
		apps/api/adapters/driving/http/testdata/srt-health-detail.json \
		apps/api/adapters/driving/http/testdata/srt-health-cursor-invalid-legacy.json \
		apps/api/adapters/driving/http/testdata/srt-health-cursor-invalid-malformed.json; then \
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
	$(TS_DOCKER_BUILD) --target deps -t $(TS_IMAGE):deps .

lock-refresh:
	$(TS_DOCKER_BUILD) --target lock-refresh-tool -t $(TS_IMAGE):lock-refresh-tool .
	docker run --rm \
		--user "$$(id -u):$$(id -g)" \
		-e XDG_CACHE_HOME=/tmp/.cache \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		$(TS_IMAGE):lock-refresh-tool \
		pnpm install --lockfile-only --ignore-scripts

# fullbuild ist der kanonische End-zu-End-Lauf vom frischen Clone:
# Dependencies installieren, alles bauen (workspace + api Docker)
# und alle Gates laufen lassen. Spiegelt das, was CI ausführt.
fullbuild: install ci
