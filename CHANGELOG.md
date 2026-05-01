# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- ADR-0003 entscheidet OE-5: Dashboard-Live-Updates in `0.4.0`
  nutzen Server-Sent Events mit Polling-Fallback; WebSocket bleibt
  deferred.
- `docs/planning/plan-0.4.0.md` als Plan für Erweiterte
  Trace-Korrelation angelegt.

### Changed

- Lastenheft `1.1.8` löst OE-3 und OE-5 auf: SQLite ist ab `0.4.0`
  der lokale Durable-Store; SSE mit Polling-Fallback ist der
  Live-Update-Mechanismus.

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
- `docs/planning/plan-0.1.0.md` spiegelt den abgeschlossenen `0.1.0`-Lieferstand wieder.
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
