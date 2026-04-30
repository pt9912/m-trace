# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `0.1.1` Workspace-Bootstrap mit pnpm-Workspace, Node/pnpm-Pinning und Root-Scripts für Build/Test/Lint/Check.
- Player-SDK-Skelett unter `packages/player-sdk` mit Core-Tracker, HTTP-Transport, hls.js-Adapter, Browser-Build und Unit-Tests.
- Player-SDK erfasst einfache Session-Metriken: Startup-Dauer sowie Rebuffer-Dauer und kumulierte Rebuffer-Zeit als optionale Event-`meta`-Felder.
- Dashboard-Skelett unter `apps/dashboard` mit SvelteKit, typisiertem API-Client, Session-/Detail-/Error-/Status-Routen und hls.js-Demo-Player.
- Compose-Lab startet das Dashboard als vierten Core-Service und `make smoke` prüft API, Dashboard, Demo-Route, HLS-Manifest und Session-Ingest.
- Containerisierter Playwright-Browser-E2E via `make browser-e2e` prüft Demo-Player → API → Dashboard in Chromium und Firefox.

### Changed

- Lastenheft `1.1.5` löst OE-8 auf: Player-SDK-Paketname `@m-trace/player-sdk`.
- Lastenheft `1.1.6` löst OE-4 auf: Dashboard-Styling im MVP nutzt eigenes CSS ohne Tailwind/UI-Library.
- Root-Targets `make test`, `make lint` und `make build` decken zusätzlich den pnpm-Workspace ab.

### Fixed

- Dashboard-Lint baut das Player-SDK vor `svelte-check`, damit frische CI-Checkouts die Workspace-Typen auflösen.
- API-CORS setzt `Access-Control-Allow-Origin` jetzt auch auf echten Dashboard-GET-Antworten, nicht nur auf Preflight-Responses.

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
