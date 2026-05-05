# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

> Post-`0.6.0`-Code-Review-Fixes plus `0.7.0` Tranche 0 (Plan-
> Aktivierung + Toolchain-Hardening). Versions-Bump und finalen
> CHANGELOG-Block setzt der `0.7.0`-Closeout (Tranche 5).

### Added

- Browser-E2E-Tests für `/srt-health` (Playwright) mit fünf Specs
  gegen `page.route()`-Mocks: Empty-State, vier Pflichtmetriken
  in der Tabelle, Stale-Pill, Detail Current+History, Detail-404.
  Schließt eine `0.6.0`-DoD-Lücke (Tranche 7 „Dashboard-Test/E2E
  grün"); Lab-gestützter E2E bleibt operative Übung in
  `releasing.md` §2.1.
- ENV `MTRACE_SRT_REQUIRED_BANDWIDTH_BPS` für die SRT-Health-
  Bandbreitenbewertung. Adapter-Hookup
  `mediamtxclient.WithRequiredBandwidthBPS` setzt das Domain-Feld
  pro Sample; ohne ENV bleibt es `nil` (spec/telemetry-model.md
  §7.4 Verhalten unverändert: angezeigt, nicht bewertet).
- Opt-in-Pfad `SMOKE_INCLUDE_MTRACE_API=1` in
  `scripts/smoke-srt-health.sh` probt zusätzlich
  `GET /api/srt/health/{stream_id}` mit `X-MTrace-Token` und
  validiert die vier RAK-43-Pflichtwerte im Wire-Format aus spec
  §7a.2. Default-off, weil `examples/srt/compose.yaml` `apps/api`
  nicht startet — Operator schaltet ihn beim Release-Closeout an.
- Race-Detector-Stage `race` im `apps/api/Dockerfile`
  (`CGO_ENABLED=1 go test -race ./...`); Targets `make race` /
  `make api-race` mit `--no-cache-filter race`. **In `make gates`
  aufgenommen** (Race ist Superset von `make test`); ~33 s vs.
  ~20 s `api-test`.

### Changed

- Toolchain-Bump für `apps/api`: Go 1.22.7 → 1.26.0 (1.22 ist seit
  Februar 2025 EOL); `golang:1.22` → `golang:1.26` für deps/test/
  coverage/build-Stages und arch-check; `golangci-lint v1.62-alpine`
  (Sep 2024, Go 1.23) → `v2.12.1-alpine` (Mai 2026, Go 1.26.2).
  `.golangci.yml` über `golangci-lint migrate` auf v2-Schema
  gezogen (`disable-all: true` → `default: none`,
  `gomodguard` → `gomodguard_v2`, `run.timeout` entfällt). Runtime
  bleibt CGO-frei `distroless-static` (Race-Stage erbt nur von
  `deps`).
- `make gates` ruft jetzt `api-race ts-test` statt `test` — Go-
  Tests laufen mit Race-Detector als Pflicht-Step.
- `mockSrtHealthRepo` (Test-Helper) mit `sync.Mutex` +
  `appendedCount()`-Helper abgesichert. Race-Stage hatte einen
  echten Data-Race aufgedeckt: Mock schrieb aus Collector-
  Goroutine während Test-Body parallel `len(appended)` las.
- `plan-0.6.0.md` (archiviert in `done/`): Status-Häkchen
  konsistent zur Release-Realität nachgezogen — §1 Tranche 5/7
  von `🟡`/`⬜` auf ✅, alle DoD-Boxen in §8 mit Datum 2026-05-05
  abgehakt; neue §8.3 mit Post-Release-Code-Review-Befund-
  Tabelle (vier Befunde mit Schwere/Korrektur/Commit) plus drei
  Lehren für den `0.7.0`-Closeout.

### Fixed

- `srt_health_collector_test.go`: zwei Polling-Loops
  `for { if X >= N { break } }` → `for X < N { ... }`
  (staticcheck QF1006 quickfix in `golangci-lint v2.12.1`); zwei
  parallele Reads gegen den Mock thread-safe gemacht.

## [0.6.0] - 2026-05-05

> SRT-Health-View: lokaler Verbindungs-Health-Pfad mit MediaMTX-API
> als CGO-freier Quelle (Risiken-Backlog R-2 als CGO-frei aufgelöst);
> durabler Health-Store, Read-API plus Dashboard-Route. RAK-41..RAK-46
> erfüllt; Lieferstand der Tranchen 0–7 strukturiert nach Spec/Domain/
> Adapter/UI/Doku.

### Added

- **SRT-Health-Smoke (Tranche 2, RAK-41):**
  [`scripts/smoke-srt-health.sh`](scripts/smoke-srt-health.sh) +
  `make smoke-srt-health`. Probt HLS-Baseline plus MediaMTX-API
  `/v3/srtconns/list` und vier RAK-43-Pflichtwerte; auth-Override
  in [`examples/srt/mediamtx.yml`](examples/srt/mediamtx.yml) per
  `authInternalUsers`-Block.
- **Spec-Block für SRT-Health (Tranche 3 Sub-3.1, RAK-42/RAK-46):**
  Neue [`spec/telemetry-model.md`](spec/telemetry-model.md) §7 mit
  Datenmodell, Health-Schwellen, Source-Status-Tabelle, Cardinality-
  Vertrag; [`spec/backend-api-contract.md`](spec/backend-api-contract.md)
  §7a (Read-Vertrag) und §10.6 (Persistenz);
  [`spec/architecture.md`](spec/architecture.md) §5.4 Datenfluss-
  Diagramm. §3.1/§3.2 Allowlist um `health_state`/`source_status`/
  `source_error_code` erweitert; SRT-Source-Labels (`id`/`path`/
  `remoteAddr`/`state`) explizit verboten.
- **Domain-Modell + Driven-Ports (Sub-3.2):**
  `apps/api/hexagon/domain/srt_health.go` mit Enums (HealthState,
  SourceStatus, SourceErrorCode, ConnectionState) plus
  `SrtConnectionSample`/`SrtHealthSample`-Records;
  `port/driven/srt_source.go`, `srt_health_repository.go`,
  `srt_errors.go` (Sentinels). Application-Use-Case
  `SrtHealthCollector` mit reiner `Evaluate`-Funktion (RTT/Loss/
  Bandbreiten-Schwellen aus `DefaultThresholds`).
- **SQLite-Persistenz (Sub-3.3):** Migration `V5__srt_health_samples.sql`
  und durable Tabelle laut spec §10.6; Adapter
  `apps/api/adapters/driven/persistence/sqlite/srt_health_repository.go`
  mit Dedupe-Skip auf
  `(project_id, stream_id, connection_id, COALESCE(source_observed_at, source_sequence))`.
- **HTTP-Client-Adapter (Sub-3.4):**
  `apps/api/adapters/driven/srt/mediamtxclient/` implementiert
  `SrtSource` über HTTP-Pull gegen MediaMTX `/v3/srtconns/list`,
  CGO-frei. Auth via Basic-Auth aus ENV. Sentinel-Fehler-Wrapping
  für die drei Source-Status-Klassen. Fixture
  [`spec/contract-fixtures/srt/mediamtx-srtconns-list.json`](spec/contract-fixtures/srt/mediamtx-srtconns-list.json)
  pinnt das Wire-Format.
- **Polling-Loop + cmd/api-Wiring (Sub-3.5):** Run-Methode mit
  exponentiellem Backoff (5s → 60s); ENV-Konfig
  `MTRACE_SRT_SOURCE_URL` / `_USER` / `_PASS` /
  `_PROJECT_ID` / `_POLL_INTERVAL_SECONDS`. Collector ist opt-in,
  bleibt im Default-Lab deaktiviert.
- **Prometheus-Aggregate + OTel-Span (Sub-3.6):** drei bounded
  CounterVecs (`mtrace_srt_health_samples_total{health_state}`,
  `mtrace_srt_health_collector_runs_total{source_status}`,
  `mtrace_srt_health_collector_errors_total{source_error_code}`)
  plus Span `mtrace.srt.health.collect` mit `mtrace.srt.*`-Attributen.
- **Smoke-Erweiterung (Sub-3.7):** Integrationstest in
  `apps/api/adapters/driven/persistence/sqlite/srt_health_collector_integration_test.go`
  weist zwei Samples mit fortschreitender SourceSequence in real-
  SQLite nach; [`scripts/smoke-observability.sh`](scripts/smoke-observability.sh)
  prüft bounded Allowlist für `mtrace_srt_health_*` und liest
  Prometheus-Targets gegen `mediamtx`/`srt`-Muster.
- **Read-API (Tranche 4, RAK-43):** Endpoints `GET /api/srt/health`
  und `GET /api/srt/health/{stream_id}` mit Token-Auth analog
  `/api/stream-sessions`. Wire-Format trennt `metrics`/`derived`/
  `freshness`-Block (spec §7a.2); Snapshot-Test gegen
  [`spec/contract-fixtures/api/srt-health-detail.json`](spec/contract-fixtures/api/srt-health-detail.json).
- **Dashboard-Route (Tranche 5, RAK-43/RAK-44):** Sidebar-Tab
  „SRT health" plus Routes `/srt-health` (Tabelle pro Stream) und
  `/srt-health/[stream_id]` (Current + History, samples_limit=50);
  `isSrtSampleStale`-Helper, Stale-Pill-Variante (gelb), 5s-
  Polling. 18 Component-Tests in vitest decken Loading/Empty/
  Error/Stale/Polling ab.
- **Operator-Doku (Tranche 6, RAK-45):**
  [`docs/user/srt-health.md`](docs/user/srt-health.md) mit 12
  Sektionen — Quickstart, Datenfluss, Metriken (mit MediaMTX-
  Mapping), Health-Zustände, Counter-vs-Rate, Bandbreite-Caveat
  (Loopback-Gbps-Falle), Freshness/Stale, Source-Status-Tabelle,
  acht Fehlerbilder, Cardinality-/Datenschutzvertrag, Operator-
  Quickref, Deferred-Liste. Querverweise von
  `examples/srt/README.md`, `docs/user/local-development.md` §2.7.1
  und [`docs/user/releasing.md`](docs/user/releasing.md) §2.1
  (fünf manuelle 0.6.0-Prüfschritte).

### Changed

- **Risiken-Backlog:** R-2 (CGO/SRT-Bindings) durch Tranche 1 als
  CGO-frei aufgelöst und nach §1.2 verschoben — MediaMTX-API über
  HTTP trägt alle vier RAK-43-Pflichtwerte. Folge-ADR „SRT-Binding-
  Stack" als obsolet markiert. Stand-Notizen für R-5/R-7/R-9/R-10
  („0.6.0 Closeout: Triggerschwelle nicht ausgelöst"). Neues R-11
  für SRT-Health-Cursor-Pagination (samples_limit-only in 0.6.0;
  Cursor-Pfad als ErrNotImplemented gestubbed).
- **MetricsPublisher-Port** um drei SRT-Methoden erweitert
  (`SrtHealthSampleAccepted`/`SrtCollectorRun`/`SrtCollectorError`);
  Telemetry-Port um `SrtSampleRecorded`. Bestehende Mocks
  in Test-Suite (`spyMetrics`, `noopTelemetry`, `stubTelemetry`)
  no-op-erweitert.
- **Versions-Bump auf 0.6.0** über alle 5 `package.json`,
  `serviceVersion`, `PLAYER_SDK_VERSION`, `STREAM_ANALYZER_VERSION`,
  `sdk_version`, Pack-Smoke-Tarball, plus Test-Fixtures und
  Contract-Fixtures.

### Notes

- Browser-E2E für die Dashboard-Route ist als manueller 5-Schritte-
  Test in [`docs/user/releasing.md`](docs/user/releasing.md) §2.1
  dokumentiert; Automatisierung als Folge-Item.
- MediaMTX-`mbpsLinkCapacity` liefert in Loopback-Lab Gbps-Werte;
  Health-Bewertung ohne `required_bandwidth_bps` ist nur Anzeige
  (siehe [`docs/user/srt-health.md`](docs/user/srt-health.md) §4.2).

## [0.5.0] - 2026-05-05

> Multi-Protokoll-Lab: MediaMTX-, SRT-, DASH-Beispiele plus WebRTC-
> Vorbereitungspfad. Lieferstand der Tranchen 0–6 strukturiert nach
> Lab-/Beispiel-Bereichen.

### Added

- **examples/-Struktur (Tranche 1):** Konventions-Index
  [`examples/README.md`](examples/README.md) mit Mindeststruktur für
  Beispiel-READMEs, Compose-Form-Tabelle (Core-Lab vs. Eigenes
  Compose mit Project-Name `mtrace-<name>`), Smoke-Naming und
  Smoke-Skript-Konvention; vier Sub-Verzeichnisse mit konsistenter
  7-Punkt-README-Struktur.
- **MediaMTX-Beispiel (Tranche 2, RAK-36):** Core-Lab-Variante in
  [`examples/mediamtx/README.md`](examples/mediamtx/README.md);
  opt-in Smoke `make smoke-mediamtx` prüft den HLS-Pfad (200,
  `#EXTM3U`-Body, Media-Referenzen) gegen ein laufendes `make dev`.
- **SRT-Beispiel (Tranche 3, RAK-37):** eigenes Compose-Project
  `mtrace-srt` ([`examples/srt/`](examples/srt/)) mit FFmpeg-SRT-
  Publisher → MediaMTX-SRT-Listener → HLS auf Host-Port `8889`;
  opt-in Smoke `make smoke-srt` mit Auto-Start/-Stop. Keine SRT-
  Health-Metriken, kein CGO-Binding — Folge-Scope `0.6.0`
  (Risiken-Backlog R-2 unverändert).
- **DASH-Beispiel (Tranche 4, RAK-38):** eigenes Compose-Project
  `mtrace-dash` ([`examples/dash/`](examples/dash/)) mit FFmpeg-
  DASH-Live-Generator → shared Volume → nginx-Static-Server auf
  Host-Port `8891`; opt-in Smoke `make smoke-dash` mit Auto-Start/
  -Stop. Keine produktive DASH-Manifestanalyse oder dash.js-
  Adapter — `analyzerKind: "hls"` bleibt einzige produktive
  Variante; eine MPD an `POST /api/analyze` liefert `not_hls`
  (erwartet).
- **WebRTC-Vorbereitungspfad (Tranche 5, RAK-39):** Doku-only
  Beispielplatz in [`examples/webrtc/README.md`](examples/webrtc/README.md);
  kein Compose, kein Smoke in `0.5.0`. Folge-Pfad-Sektion benennt
  vier konkrete Schritte (Lab-Compose, README-Konkretisierung,
  `smoke-webrtc-prep`-Target, WebRTC-Telemetrie-Bewertung) für
  spätere Tranchen.
- **Doku-/Closeout-Updates (Tranche 6, RAK-40):**
  [`docs/user/local-development.md`](docs/user/local-development.md)
  §2.7 mit Quickref-Tabelle aller vier Beispiele und parallel-Stack-
  Port-Schnitt (Core-Lab `8888`/`9997`, `mtrace-srt` `8889`/`8890`/
  `9998`, `mtrace-dash` `8891`); README v0.5.0-Block;
  `docs/user/releasing.md` listet die drei `0.5.0`-Smokes; jede
  Beispiel-README hat Quickref-Verweis auf §2.7 zurück.

### Changed

- `apps/dashboard`-Demo-Pfad bleibt unverändert auf
  `hls.js`/Core-Lab-`teststream` — keine DASH-/WebRTC-Demo-Route.
- `@npm9912/player-sdk` Public-API unverändert; kein dash.js- oder
  WebRTC-Adapter in `0.5.0`.
- `@npm9912/stream-analyzer` und `POST /api/analyze` bleiben HLS-
  only; DASH-/CMAF-Erweiterung ist Folge-Scope (MVP-37).

## [0.4.0] - 2026-05-05

> Erweiterte Trace-Korrelation: SQLite-Persistenz, `correlation_id`/
> `trace_id`-Trennung, Dashboard-Session-Timeline ohne Tempo-Pflicht,
> optionales Tempo-Profil, Aggregat-Metriken-Sichtbarkeit, Cardinality-/
> Sampling-Doku. Lieferstand der Tranchen 1–7 strukturiert nach Trace-,
> Storage-, Dashboard-/SSE-, Tempo-, Metrik- und Doku-Bereich.

### Added

- **Persistenz (Tranche 1):** durable SQLite-Persistenz für
  `stream_sessions`, `playback_events` und `ingest_sequence`; Cursor
  sind Restart-stabil ([ADR-0002](docs/adr/0002-persistence-store.md));
  Reset-Pfad ist `make wipe`; Cursor-v3 mit Project-Scope plus
  `cursor_invalid_legacy`/`cursor_invalid_malformed`/`cursor_expired`-
  Codes (siehe `spec/backend-api-contract.md` §10.3).
- **Trace-Korrelation (Tranche 2):** `correlation_id` als durable
  Source-of-Truth pro Player-Session über alle Batches hinweg;
  `trace_id` ist optionale Per-Batch-Vertiefung. Hybrid-`traceparent`-
  Strategie: SDK propagiert optional einen W3C-Header; Server toleriert
  fehlende oder ungültige Header (Server-Span-Attribut
  `mtrace.trace.parse_error=true`). Span-Modell pro Batch mit
  `mtrace.session.correlation_id` für Single-Session-Batches; das
  `session_id`-Span-Attribut ist ab `0.4.0` verboten
  (`spec/telemetry-model.md` §2.5).
- **Manifest-/Segment-/Player-Korrelation (Tranche 3):** alle
  Network-Events (`manifest_loaded`, `segment_loaded`) tragen
  `correlation_id` und URL-Redaction am SDK-Boundary;
  `session_boundaries[]`-Wrapper und `network_signal_absent[]`-Read-
  Shape decken Browser-/CORS-/Service-Worker-/Native-HLS-/CDN-
  Degradationen ab. Endpoint-spezifische Auth: `POST /api/playback-
  events` und Session-/Event-Reads sind tokenpflichtig; ungebundene
  `POST /api/analyze`-Requests bleiben tokenfrei und liefern
  `session_link.status="detached"` ([R-6](docs/planning/open/risks-backlog.md)
  technisch geschlossen).
- **Dashboard-Session-Timeline (Tranche 4):** Timeline-Ansicht
  `/sessions/<id>` mit Server-Sent Events ([ADR-0003](docs/adr/0003-live-updates.md))
  plus Polling-Fallback und Backfill-Cursor; Mini-Status-Panels und
  konfigurierbare Service-Links (F-39/F-40); Tempo-unabhängig (RAK-32).
- **Optionales Tempo-Profil (Tranche 5):** `make dev-tempo` startet
  Tempo neben Prometheus/Grafana/OTel-Collector;
  `scripts/smoke-tempo.sh` deckt drei Stack-Zustände ab (`core`,
  `observability`, `tempo`); RAK-31 Kann-Scope erfüllt.
- **Aggregat-Metriken (Tranche 6):** vier Pflichtcounter
  (`mtrace_playback_events_total`, `mtrace_invalid_events_total`,
  `mtrace_rate_limited_events_total`, `mtrace_dropped_events_total`)
  bleiben label-frei; Backend-Tests pinnen Inkrement- und Null-
  Inkrement-Pfade in `metrics_counter_test.go`. Cardinality-Smoke
  (`scripts/smoke-observability.sh`) verschärft auf vollständige
  Forbidden-Liste plus Per-Pflichtcounter-Labelset-Whitelist und
  Cardinality-Cap < 50 Serien. Grafana-Dashboard
  `m-trace-overview.json` zeigt die vier Pflichtcounter.
- **Cardinality-/Sampling-Doku (Tranche 7):**
  `spec/backend-api-contract.md` §7 verweist auf
  `spec/telemetry-model.md` §3.1 als kanonische Forbidden-Liste; §3.1
  deckt §7-Mindestliste vollständig ab inklusive Suffix-Regeln
  (`*_url`/`*_uri`/`*_token`/`*_secret`). Der OTel-Counter
  `mtrace.api.batches.received` ist ab `0.4.0` Tranche 7 label-frei
  (`batch.size` lebt nur am Span); `batch_size` ist in Smoke
  `scripts/smoke-observability.sh` und in §3.1 verboten. Sampling-
  Nachweisgrenze für `sampleRate < 1` dokumentiert in
  `spec/player-sdk.md` und `packages/player-sdk/README.md`.
- **Doku:** `docs/user/local-development.md` §3.4 dokumentiert
  Storage-Retention ("unlimited mit dokumentiertem Reset-Pfad"), §3.5
  ergänzt Prometheus-Aggregate-Quickref; `docs/user/demo-integration.md`
  zeigt reproduzierbare Demo-Session inkl. Timeline-Verifikation und
  SQLite-Restart-Stabilität.

### Changed

- Lastenheft `1.1.8` löst OE-3 und OE-5 auf: SQLite ist ab `0.4.0`
  der lokale Durable-Store; SSE mit Polling-Fallback ist der
  Live-Update-Mechanismus.
- `POST /api/analyze` antwortet ab Tranche 3 für **alle**
  erfolgreichen Requests mit der Hülle `{analysis, session_link}`;
  ungebundene Requests erhalten `session_link.status="detached"`
  (Breaking Change gegenüber `0.3.x`).
- `mtrace.api.batches.received` ist ab Tranche 7 label-frei — der
  bisherige `batch.size`/`batch_size`-Counter-Attribut-Pfad ist
  entfernt; Span-Attribut bleibt für Trace-Debugging unverändert.

## [0.3.0] - 2026-05-01

### Added

- Workspace-Paket `@npm9912/stream-analyzer` mit HLS-Klassifikator,
  URL-Loader (Timeout/Größenlimit/SSRF-Sperrlisten),
  Master- und Media-Detail-Parser sowie diskriminierter Union-API
  `AnalysisResult` (`analyzerKind: "hls"`, `analyzerVersion`,
  Stabilitätsregel und Serialisierungsgarantien).
- Internes HTTP-Service-Paket `@npm9912/analyzer-service`
  (`apps/analyzer-service`) als Node-Wrapper um den Analyzer; läuft
  in der Compose-Topologie als `analyzer-service`-Container.
- API-Endpunkt `POST /api/analyze` mit Pass-through-Schema und
  Problem-Shape-Fehlern (`invalid_request`, `analyzer_unavailable`
  etc.). Go-Driven-Adapter `HTTPStreamAnalyzer` ruft den
  analyzer-service.
- `make smoke-analyzer` als End-to-End-Smoke (Master-Text-Input und
  SSRF-Negativfall) im laufenden Compose-Stack.
- CLI `m-trace check <url-or-file>` aus `@npm9912/stream-analyzer`:
  bin-Eintrag, Datei- und URL-Input (URL teilt den SSRF-geschützten
  Loader-Pfad), JSON auf stdout, Exit-Codes 0/1/2, `--help` und
  `--version`. Smoke `make smoke-cli` deckt `--help`, Master-Datei,
  Nicht-HLS-Datei, fehlende Datei, no-args, SSRF-URL und Bin-Symlink
  ab.
- Doku: `docs/user/stream-analyzer.md` (vollständiger 0.3.0-Stand)
  und `spec/backend-api-contract.md` §3.6 Analyzer-Endpunkt.
- Tranche-7.5-Härtung der API-Anbindung:
  - Prometheus-Counter `mtrace_analyze_requests_total{outcome,code}`
    zählt jeden `POST /api/analyze`-Aufruf (`outcome` ∈ `ok|error`,
    `code` ∈ `ok|invalid_request|invalid_json|unsupported_media_type|payload_too_large|invalid_input|fetch_blocked|manifest_not_hls|fetch_failed|manifest_too_large|internal_error|analyzer_unavailable`).
    Cardinality bleibt durch eine Whitelist im Publisher beschränkt.
  - `analyzer-service` respektiert `ANALYZER_ALLOW_PRIVATE_NETWORKS=true|1|yes|on`
    und reicht ein neues `FetchOptions.allowPrivateNetworks`-Flag an
    den Loader weiter. Default bleibt: SSRF-IP-Block aktiv. Aufrufer
    können das Flag nicht über den Body anfordern (Service-Whitelist).
  - `apps/analyzer-service/Dockerfile` baut ohne zweiten
    pnpm-install-Schritt — `pnpm deploy --prod --legacy /deploy` in
    der Build-Stage erzeugt ein selbsttragendes Bundle, die
    Runtime-Stage übernimmt es per `COPY`.
  - Cross-Process-Vertragstest TS↔Go: gemeinsame Fixtures unter
    `spec/contract-fixtures/analyzer/`; TS-Test pinnt
    `analyzeHlsManifest`-Output gegen Spec, Go-Test parst die Kopien
    in `apps/api/.../testdata/` via `go:embed`, plus ein TS-Drift-
    Check gegen die Spec-Quelle. `make sync-contract-fixtures`
    syncronisiert die Kopien per Knopfdruck.

### Tooling

- Wurzel-`Makefile` deckt jetzt `install`, `fullbuild`, `smoke-analyzer`,
  `smoke-cli` und `sync-contract-fixtures` ab; `workspace-test`,
  `workspace-lint` und `workspace-coverage-gate` hängen am
  `workspace-build`, damit Tests und Linter die Workspace-Dependencies
  in Topo-Sort erst bauen.
- `.gitattributes` setzt `text eol=lf` als Default plus harte Pflicht
  für `*.json`/`*.m3u8`/`*.sh`, damit Windows-Checkouts keine
  CRLF-Drift in den Contract-Fixtures erzeugen.

## [0.2.0] - 2026-04-30

### Added

- Publizierbares Player-SDK-Paket `@npm9912/player-sdk` mit ESM-, CJS-,
  Browser/IIFE- und Type-Definition-Builds.
- Pack-, Publish-Dry-Run-, Install- und Browser-Load-Smokes für das SDK.
- Projektweite SDK-Doku in `spec/player-sdk.md` sowie Paketdoku in
  `packages/player-sdk/README.md`.
- Maschinenlesbare Contract-Artefakte für Event-Schema und SDK↔Schema-
  Kompatibilität.
- CI-Kompatibilitätscheck für SDK-Version, `sdk.version`,
  `schema_version` und API-`SupportedSchemaVersion`.
- Vitest-Coverage-Gates für Player-SDK und Dashboard mit verbindlichen
  90-%-Schwellen.
- Performance-Smoke für das Player-SDK mit Bundle-, Hot-Path- und
  Queue-/Retry-Prüfungen.
- Browser-Support-Matrix in `spec/browser-support.md`.
- Demo-Integrationsdoku für die Dashboard-Route `/demo`.
- ADR-Draft `0002` zur Persistenzentscheidung In-Memory vs.
  SQLite/PostgreSQL.

### Changed

- Lastenheft `1.1.7` entscheidet OE-8 neu: Player-SDK wird ab `0.2.0` als `@npm9912/player-sdk` veröffentlicht. Der `0.1.x`-Lieferstand wurde nie öffentlich publishet, daher ist kein Migrations-Pfad für externe Konsumenten erforderlich.
- Player-SDK-Events senden die SDK-Version synchron aus
  `packages/player-sdk/package.json`.
- Player-SDK-Batches bleiben innerhalb der API-Grenzen: maximal 100 Events
  und maximal 256 KiB Request-Body.
- `HttpTransport` respektiert `Retry-After` bei `429`, retried nur
  transiente Fehler und vermeidet blindes Retry bei nicht-transienten `4xx`
  sowie `413`.
- Dashboard- und SDK-Paketnamen wurden auf den `@npm9912`-Scope migriert.

### Fixed

- Dashboard-Tests laufen in frischen CI-Checkouts ohne vorher gebautes
  SDK-`dist`, weil Vitest den SDK-Import im Testmodus auf einen lokalen Mock
  auflöst.
- `session_ended` wird beim Tracker-`destroy()` zuverlässig erzeugt und
  umgeht Sampling.

## [0.1.2] - 2026-04-30

### Added

- Observability-Compose-Profil mit Prometheus, Grafana und OTel-Collector.
- Prometheus-Konfiguration, Grafana-Provisioning und m-trace-Beispieldashboard.
- API-Mindestmetriken für aktive Sessions, API-Requests, Playback-Fehler, Rebuffer-Events und Startup-Zeit.
- RAK-9-Seed- und Smoke-Skripte für Prometheus-Cardinality-Checks.
- RAK-10-Console-Smoke für exemplarische OTel-Request-Spans.

## [0.1.1] - 2026-04-30

### Added

- `0.1.1` Workspace-Bootstrap mit pnpm-Workspace, Node/pnpm-Pinning und Root-Scripts für Build/Test/Lint/Check.
- Player-SDK-Skelett unter `packages/player-sdk` mit Core-Tracker, HTTP-Transport, hls.js-Adapter, Browser-Build und Unit-Tests.
- Player-SDK erfasst einfache Session-Metriken: Startup-Dauer sowie Rebuffer-Dauer und kumulierte Rebuffer-Zeit als optionale Event-`meta`-Felder.
- Dashboard-Skelett unter `apps/dashboard` mit SvelteKit, typisiertem API-Client, Session-/Detail-/Error-/Status-Routen und hls.js-Demo-Player.
- Compose-Lab startet das Dashboard als vierten Core-Service und `make smoke` prüft API, Dashboard, Demo-Route, HLS-Manifest und Session-Ingest.
- Containerisierter Playwright-Browser-E2E via `make browser-e2e` prüft Demo-Player → API → Dashboard in Chromium und Firefox.
- Dashboard-Route `/events` zeigt Playback-Events über aktuelle Sessions hinweg mit Session- und Event-Typ-Filter.
- Status-Ansicht kennzeichnet Prometheus, Grafana und OTel Collector einzeln als inaktiv, solange das Observability-Profil nicht läuft.

### Changed

- Lastenheft `1.1.5` löst OE-8 auf: Player-SDK-Paketname `@m-trace/player-sdk`.
- Lastenheft `1.1.6` löst OE-4 auf: Dashboard-Styling im MVP nutzt eigenes CSS ohne Tailwind/UI-Library.
- Root-Targets `make test`, `make lint` und `make build` decken zusätzlich den pnpm-Workspace ab.

### Fixed

- Dashboard-Lint baut das Player-SDK vor `svelte-check`, damit frische CI-Checkouts die Workspace-Typen auflösen.
- API-CORS setzt `Access-Control-Allow-Origin` jetzt auch auf echten Dashboard-GET-Antworten, nicht nur auf Preflight-Responses.
- Player-SDK begrenzt Batches auf maximal 100 Events, splittet größere lokale Queues und sendet beim `destroy()` ein `session_ended`-Event.
- `docs/planning/done/plan-0.1.0.md` spiegelt den abgeschlossenen `0.1.0`-Lieferstand wieder.
- README und Local-Development-Doku trennen den `0.1.0`- und `0.1.1`-Scope klarer.

## [0.1.0] - 2026-04-30

### Added

- `0.1.0` Compose-Lab Core mit `api`, `mediamtx` und `stream-generator`.
- Root-Targets `make dev`, `make stop` und `make smoke`.
- Root-Targets `make test`, `make lint`, `make coverage-gate`, `make arch-check` und `make build` für lokale CI-Parität.
- GitHub-Actions-Workflow `build.yml` für API-Test, Lint, Coverage-Gate, Architekturprüfung und Runtime-Build auf `ubuntu-24.04`.
- MediaMTX-Konfiguration für RTSP-Publish, HLS auf Port `8888` und HTTP-API auf Port `9997`.
- FFmpeg-Teststream via `jrottenberg/ffmpeg:8.1-ubuntu2404`.

### Changed

- API-Listen-Adresse ist über `MTRACE_API_LISTEN_ADDR` konfigurierbar.
- Local-Development-Doku beschreibt den verifizierten `0.1.0`-Smoke-Test.
- Lastenheft `1.1.4` löst OE-1, OE-6 und OE-7 auf: MIT-Lizenz, GitHub Actions auf `ubuntu-24.04`, trunk-based Releases mit annotierten `vX.Y.Z`-Tags.
