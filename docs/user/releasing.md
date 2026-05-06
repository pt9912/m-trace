# Releasing — m-trace

> **Status**: Verbindlich für alle Releases (zuletzt verifiziert mit
> `0.4.0`). CI-Verifikation, Branching-Modell und Tag-Format sind
> stabil; Container-Image-Veröffentlichung bleibt deferred.
> Bezug: AK-11, DoD §18 (Lastenheft).

## 0. Zweck

Dieses Dokument beschreibt den minimalen, reproduzierbaren
Release-Ablauf für m-trace. Der Ablauf ist versionsunabhängig
formuliert und verwendet Platzhalter der Form `X.Y.Z`. Er gilt für
alle Releases aus dem Release-Plan (Lastenheft §13, Roadmap §3 —
RAK-1..RAK-46).

## 1. Vorbereitung

```bash
VER="X.Y.Z"
TAG="v$VER"
```

Vor jedem Release:

- noch nicht veröffentlichte Änderungen stehen unter `## [Unreleased]`
  in `CHANGELOG.md`; ein datierter Versionsabschnitt entsteht erst mit
  dem Release-Commit.
- `CHANGELOG.md` auf den Zielstand bringen.
- betroffene Plan-, Status- und Nutzungsdokumente aktualisieren
  (`docs/planning/in-progress/roadmap.md`, `spec/architecture.md`, `apps/api/README.md`).
- Roadmap §1.1 und §1.2 nach dem Release-Bump neu schreiben (siehe
  `docs/planning/in-progress/roadmap.md` §7 Wartungsregel).
- offene `OE-X` und `R-X` durchsehen — Einträge, die mit dem Release
  aufgelöst werden, aus den Tabellen entfernen.

## 2. Verifikation

Vor Tag und GitHub-Release müssen die Root-Targets grün sein:

```bash
make gates                # CI-äquivalenter Komplettcheck (api-race+ts-test+lint+coverage+arch+schema+docs)
make build
make sdk-performance-smoke
make smoke-cli            # ab 0.3.0: Lastenheft-Aufruf `pnpm m-trace check`
make smoke-analyzer       # ab 0.3.0: manuelles Release-Gate, fährt Compose hoch
make smoke-observability  # ab 0.4.0: Cardinality-Smoke; Observability-Stack muss laufen
make browser-e2e          # ab 0.4.0: Dashboard-Timeline + hls.js-Demo-Flow
make smoke-mediamtx       # ab 0.5.0: MediaMTX-Beispiel (RAK-36); braucht laufendes `make dev`
make smoke-srt            # ab 0.5.0: SRT-Beispiel (RAK-37); startet/stoppt Project mtrace-srt
make smoke-srt-health     # ab 0.6.0: SRT-Health-Smoke (RAK-41/RAK-42); startet/stoppt mtrace-srt + probt MediaMTX-API
make smoke-dash           # ab 0.5.0: DASH-Beispiel (RAK-38); startet/stoppt Project mtrace-dash
make smoke-webrtc-prep    # ab 0.7.0: WebRTC-Lab-Vorbereitungs-Smoke (RAK-48); startet/stoppt mtrace-webrtc; endpoint-only (kein Browser/Playback/getStats)
```

Erfolgskriterien:

- alle Targets exit code 0.
- `make gates` umfasst `make api-race` (Go-Tests mit Race-Detector,
  CGO=1; ab `0.7.0` Tranche 0 in gates statt `api-test`, weil
  Race-Detection ein Superset ist), `make ts-test`, `make lint`,
  `make coverage-gate`, `make arch-check`, `make schema-validate`
  und `make docs-check` — einzelne Aufrufe sind möglich, aber
  `make gates` ist die CI-äquivalente Eingangsstufe.
- `make coverage-gate` (Teil von `make gates`) umfasst API-,
  Player-SDK-, Dashboard-, stream-analyzer- und (ab `0.3.0`)
  analyzer-service-Coverage.
- `golangci-lint`-Stage liefert keine Findings.
- `go test ./...` deckt mindestens die Pflichttests aus
  `spec/backend-api-contract.md` §11 ab.
- Coverage-Gate liegt bei mindestens 90 %.
- Architektur-Grenzen bleiben laut `make arch-check` intakt.
- `make smoke-observability` setzt einen laufenden Observability-Stack
  voraus (`make dev-observability` bzw. Compose mit
  `--profile observability`); ohne aktiven Stack schlägt der Smoke
  release-blockierend fehl.
- `make browser-e2e` startet API/MediaMTX/FFmpeg/Dashboard im
  Container und prüft die `/demo`-Route inklusive Session-Timeline-
  Read-Pfad in Chromium und Firefox.

CI deckt `make gates`, `make build`, `make sdk-performance-smoke` und
`make smoke-cli` ab; `smoke-analyzer`, `smoke-observability`,
`browser-e2e` und ab `0.5.0` `smoke-mediamtx`/`smoke-srt`/
`smoke-dash` (plus ab `0.6.0` `smoke-srt-health`, ab `0.7.0`
`smoke-webrtc-prep`) laufen lokal vor dem Tag (Compose-Stack-Up bzw.
Browser-Stack ist zu schwergewichtig für jeden PR-Run). CI-Zielplattform
ist GitHub Actions auf `ubuntu-24.04`, Workflow-Name: `build`.

### 2.1 Manuelle `0.6.0`-Prüfungen (SRT-Health-View)

Zusätzlich zu den oben gelisteten Smokes braucht der `0.6.0`-Release
eine kurze manuelle Operator-Prüfung gegen ein laufendes Lab:

1. `make dev` plus `examples/srt/`-Stack (`docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build`).
2. ENV `MTRACE_SRT_SOURCE_URL=http://localhost:9998` und optional
   `MTRACE_SRT_REQUIRED_BANDWIDTH_BPS=1500000` auf den `apps/api`-
   Prozess setzen und neu starten — Log meldet
   „srt-health collector enabled".
2a. Optional automatisierte API-Probe:
   `SMOKE_INCLUDE_MTRACE_API=1 make smoke-srt-health` —
   probt zusätzlich zum MediaMTX-Pfad gegen
   `GET /api/srt/health/{stream_id}` und verifiziert die vier
   RAK-43-Pflichtwerte im Wire-Format aus spec §7a.2.
3. Dashboard-Route <http://localhost:5173/srt-health> öffnen — die
   Tabelle muss `srt-test` mit Health-Pill `healthy`, RTT < 5 ms und
   Bandbreite im Mbit/s-Bereich zeigen.
4. Detail-Route `/srt-health/srt-test` — History muss mindestens
   zwei Samples mit fortschreitender Source-Sequence haben (Polling
   alle 5 s).
5. Stale-Pfad: Publisher kurz stoppen
   (`docker compose -p mtrace-srt stop srt-publisher`); nach
   ≥ 15 s muss die Pill auf `healthy (stale)` (gelb) wechseln.

Vollständige Operator-Doku:
[`srt-health.md`](./srt-health.md).

### 2.2 Manuelle `0.7.0`-Prüfungen (WebRTC-Lab-Erweiterung)

Zusätzlich zu `make smoke-webrtc-prep` (auto-up/down, endpoint-only)
braucht der `0.7.0`-Release einen kurzen manuellen Browser-Handcheck
gegen ein laufendes Lab (RAK-50, „Kann"; nicht release-blockierend,
aber Bestandteil der dokumentierten Verifikationspfade):

1. `docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml up -d --build`.
2. `make smoke-webrtc-prep` (oder `SMOKE_WEBRTC_AUTOSTART=0
   bash scripts/smoke-webrtc-prep.sh` gegen den laufenden Stack)
   muss alle fünf Probes grün durchlaufen.
3. Browser-Read-Demo öffnen: <http://localhost:8892/webrtc-test> in
   Chromium 120+ oder Firefox 120+ — Test-Pattern + 1 kHz Sinuston
   müssen latenzarm laufen.
4. `chrome://webrtc-internals` (Chromium) bzw. `about:webrtc`
   (Firefox) zeigt eine aktive `RTCPeerConnection` mit
   `connection_state=connected`, `ice_state=connected`,
   `dtls_state=connected`. Diese Werte sind in
   `spec/telemetry-model.md` §3.5.2 als Muss-Felder dokumentiert.
5. `docker compose -p mtrace-webrtc … down` räumt nur den
   `mtrace-webrtc`-Stack ab; Core-Lab und andere Beispiele bleiben
   unangetastet.

Vollständige Operator-Doku:
[`examples/webrtc/README.md`](../../examples/webrtc/README.md).

### 2.3 Manuelle `0.8.0`-Prüfungen (Player-SDK-WebRTC-Adapter)

Zusätzlich zu `make smoke-webrtc-prep` und dem `0.7.0`-Browser-
Handcheck braucht der `0.8.0`-Release einen produktiven End-to-End-
Lauf des Player-SDK-WebRTC-Adapters gegen das laufende Lab. Pflicht-
Schritte (RAK-51..RAK-54):

1. `make dev` (Core-Lab, API + MediaMTX + Dashboard) plus
   `mtrace-webrtc`-Stack (`docker compose -p mtrace-webrtc -f
   examples/webrtc/compose.yaml up -d --build`).
2. Browser auf <http://localhost:5173/demo-webrtc?autostart=1>
   öffnen. Erwartung in Chromium 120+ und Firefox 120+:
   - Test-Pattern + 1 kHz Sinuston spielen latenzarm.
   - `chrome://webrtc-internals` (Chromium) bzw. `about:webrtc`
     (Firefox) zeigt eine aktive `RTCPeerConnection` mit
     `connection_state=connected`,
     `ice_state` in `connected`/`completed`,
     `dtls_state=connected`.
3. `curl -sS http://localhost:8080/api/metrics | grep
   '^mtrace_webrtc_'` listet die sechs Counter:
   `mtrace_webrtc_connection_state_total{connection_state="connected"}`,
   `mtrace_webrtc_ice_state_total{ice_state}`,
   `mtrace_webrtc_dtls_state_total{dtls_state}`,
   `mtrace_webrtc_packets_lost_total`,
   `mtrace_webrtc_bytes_received_total`,
   `mtrace_webrtc_bytes_sent_total`. Keine `peer_connection_run_id`-,
   `ssrc`-, `track_id`-Labels (Cardinality-Vertrag aus
   `spec/telemetry-model.md` §3.1).
4. Optional automatisierter Browser-E2E:
   `MTRACE_WEBRTC_LAB=1 make browser-e2e` flippt
   `tests/e2e/dashboard-demo-webrtc.spec.ts` auf den Happy-Path
   (`playback_started` mit `webrtc.peer_connection_run_id`).
5. Stop: `docker compose -p mtrace-webrtc … down`. Greift weder
   Core-Lab noch andere Beispiele an.

Vollständige Operator-Doku:
[`packages/player-sdk/README.md`](../../packages/player-sdk/README.md)
§Performance and Browser Support.

```bash
gh run watch --workflow build.yml
```

## 3. Release-Commit und Tag

Release-Konvention für `0.1.x`:

- trunk-based auf `main`.
- Release-Commit direkt auf `main`.
- annotierte SemVer-Tags im Format `vX.Y.Z`.
- kein Pre-Release-Suffix für Hauptreleases.

### 3.1 Patch-Release-Konvention (`0.X.Y`, ab `0.8.5`)

Erstmals eingeführt mit `0.8.5` (Quality-Gates Wave 1). Patch-
Releases gelten für CI-/Tooling-/Doku-Lieferungen ohne neue
User-Surface:

| Release-Typ | Versions-Schema | Lastenheft-Patch | RAK-Verifikationsmatrix | Beispielhaftes Material |
| --- | --- | --- | --- | --- |
| **Patch** | `0.X.Y` | nicht nötig | nicht geführt | Quality-/Security-Gates, Generated-Artifact-Drift, CI-Tooling, Doku-only-Bugfixes |
| **Minor** | `0.X.0` | Pflicht (`1.1.X`-Patch mit neuen RAK) | Pflicht in §6.1 des Plans | neue User-Surface, neue Wire-Verträge, neue Lab-/Adapter-Pfade |
| **Major** | `1.0.0` | Pflicht plus Folge-ADR | Pflicht plus Public-API-Versprechen | erstmaliges öffentliches Public-API-Versprechen (aktuell Folge-ADR-Thema) |

Versions-Bump bei Patch-Release umfasst alle Stellen, die ein
Minor-Bump auch berührt (`package.json` × 5, `main.go`
`serviceVersion`, `version.ts`, `pack-smoke.mjs` `expectedVersion`,
`contracts/sdk-compat.json` `sdk_version`, Test-Fixtures mit
hartkodierten Versions-Strings) — sonst entsteht Drift zwischen
SDK-Bundle, API-Service-Version und CI-Smokes. Plan-DoD-Items
ersetzen die RAK-Verifikationsmatrix; ein Patch-Release-Plan trägt
keinen `§6.1`-Block.

```bash
git commit -m "chore(release): vX.Y.Z"
git tag -a "$TAG" -m "Release X.Y.Z"
git push origin main
git push origin "$TAG"
```

## 4. GitHub-Release

Mindestumfang:

- Release-Notes aus dem `CHANGELOG.md`-Versionsabschnitt extrahieren.
- Release-Titel: `m-trace X.Y.Z`.
- Tag: `vX.Y.Z`.
- Assets: GitHub-Source-Archive (`zip`/`tar.gz`) genügen für `0.1.0`.
  Container-Image-Veröffentlichung folgt in einem späteren Release.

```bash
gh release create "$TAG" \
    --title "m-trace $VER" \
    --notes-file <changelog-extract>
```

## 5. Post-Release

- `CHANGELOG.md` öffnet einen neuen `## [Unreleased]`-Abschnitt.
- `docs/planning/in-progress/roadmap.md` §3 (Release-Übersicht) aktualisiert den Status
  des veröffentlichten Releases (`⬜ → ✅`).
- Folge-ADRs, die mit dem Release entstehen oder fällig werden,
  in `docs/planning/in-progress/roadmap.md` §4 ergänzen.

## 6. Rollback

Tag noch nicht gepusht:

```bash
git tag -d "$TAG"
```

Tag bereits gepusht, GitHub-Release noch nicht erstellt:

```bash
git push origin ":refs/tags/$TAG"
git tag -d "$TAG"
```

GitHub-Release bereits erstellt:

```bash
gh release delete "$TAG"
git push origin ":refs/tags/$TAG"
git tag -d "$TAG"
```

CI-Build nach Release fehlgeschlagen: Release auf GitHub als
Pre-Release/Draft zurückstufen oder löschen, Fehler auf `main`
beheben, neuen Release-Commit erstellen und Tag neu setzen. Kein
Force-Push auf `main`.

## 7. Referenzen

- Lastenheft §14 — Akzeptanzkriterien (AK-11).
- Lastenheft §18 — Definition of Done für den MVP.
- `docs/planning/in-progress/roadmap.md` §3 — Release-Übersicht und RAK-Akzeptanzkriterien.
- `docs/planning/in-progress/roadmap.md` §5 — Offene Entscheidungen.
- `CHANGELOG.md` — Versionsverlauf.
