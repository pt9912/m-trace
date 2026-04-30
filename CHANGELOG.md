# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - Unreleased

### Added

- Publizierbares Player-SDK-Paket `@npm9912/player-sdk` mit ESM-, CJS-,
  Browser/IIFE- und Type-Definition-Builds.
- Pack-, Publish-Dry-Run-, Install- und Browser-Load-Smokes für das SDK.
- Projektweite SDK-Doku in `docs/player-sdk.md` sowie Paketdoku in
  `packages/player-sdk/README.md`.
- Maschinenlesbare Contract-Artefakte für Event-Schema und SDK↔Schema-
  Kompatibilität.
- CI-Kompatibilitätscheck für SDK-Version, `sdk.version`,
  `schema_version` und API-`SupportedSchemaVersion`.
- Vitest-Coverage-Gates für Player-SDK und Dashboard mit verbindlichen
  90-%-Schwellen.
- Performance-Smoke für das Player-SDK mit Bundle-, Hot-Path- und
  Queue-/Retry-Prüfungen.
- Browser-Support-Matrix in `docs/browser-support.md`.
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
- `docs/plan-0.1.0.md` spiegelt den abgeschlossenen `0.1.0`-Lieferstand wieder.
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
