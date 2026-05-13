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
make dev-detached         # Core-Lab für smoke-mediamtx starten; danach `make stop`
make smoke-mediamtx       # ab 0.5.0: MediaMTX-Beispiel (RAK-36); braucht laufendes Core-Lab
make smoke-srt            # ab 0.5.0: SRT-Beispiel (RAK-37); startet/stoppt Project mtrace-srt
make smoke-srt-health     # ab 0.6.0: SRT-Health-Smoke (RAK-41/RAK-42); startet/stoppt mtrace-srt + probt MediaMTX-API
make smoke-dash           # ab 0.5.0: DASH-Beispiel (RAK-38); startet/stoppt Project mtrace-dash
make smoke-webrtc-prep    # ab 0.7.0: WebRTC-Lab-Vorbereitungs-Smoke (RAK-48); startet/stoppt mtrace-webrtc; endpoint-only (kein Browser/Playback/getStats)
make smoke-webrtc-stats-drift # ab 0.9.0: WebRTC-`getStats()`-Drift-Smoke (RAK-56); startet/stoppt mtrace-webrtc + ruft Playwright-Spec gegen Chromium/Firefox; opt-in lokal, produktiv im Nightly-Workflow `.github/workflows/webrtc-drift.yml`
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

### 2.4 Manuelle `0.9.0`-Prüfungen (Drift-Smoke + DASH + SRS)

`0.9.0` bündelt drei thematisch getrennte Liefergegenstände
(plan-0.9.0 Tranchen 1–3); vor dem Release-Tag laufen drei
operative Verifikationspfade an, die alle als opt-in Smokes lokal
reproduzierbar sind und nicht in `make gates` enthalten:

#### 2.4.1 WebRTC-Drift-Smoke (Tranche 1, RAK-56)

Seit `0.9.0` Tranche 1 ist der Drift-Review aus R-12 automatisiert.
Vor jedem Release-Tag (auch Patch) genügt ein Blick auf den letzten
Nightly-Lauf des Workflows
[`.github/workflows/webrtc-drift.yml`](../../.github/workflows/webrtc-drift.yml):

- **Cron**: `30 3 * * *` (UTC); `workflow_dispatch` für ad-hoc-
  Trigger nach einem Browser-Major-Release.
- **Browser-Set**: Chromium und Firefox aus dem Playwright-Bundle
  (Default); WebKit/Safari opt-in über
  `MTRACE_WEBRTC_DRIFT_BROWSERS=chromium,firefox,webkit`.
- **Befund-Pfad**: bei Schema-Drift bricht der Smoke; das Issue-
  Template steht im Workflow-`if`-Block (opt-in via
  `secrets.DRIFT_AUTO_ISSUE=1`).
- **Reaktion**: `webrtc.*`-Allowlist in
  [`spec/telemetry-model.md`](../../spec/telemetry-model.md) §1.4 +
  §3.5.2, `contracts/event-schema.json#reserved_meta_keys` und
  `packages/player-sdk/src/adapters/webrtc/sampling.ts` synchron
  aktualisieren; lokal `make smoke-webrtc-stats-drift` grün ziehen.

#### 2.4.2 SRS-Lab-Boot (Tranche 2, RAK-57)

`make smoke-srs` fährt das `mtrace-srs`-Compose hoch, prüft drei
Probes (HTTP-API erreichbar, Stream `live/srs-test` registriert,
HTTP-FLV-Egress liefert FLV-Magic-Header) und räumt den Stack ab.
Erwartete Endpoints:

```bash
make smoke-srs
# HTTP-API: http://localhost:1985/api/v1/streams/
# HTTP-FLV: http://localhost:8088/live/srs-test.flv
```

Ports `1935/1985/8088` müssen frei sein; Operator-Doku siehe
[`examples/srs/README.md`](../../examples/srs/README.md).

#### 2.4.3 DASH-CLI-Probe (Tranche 3, RAK-58/RAK-59)

`make smoke-cli` ist seit `0.9.0` Tranche 3 um einen DASH-Pfad
erweitert: zusätzlich zum HLS-Master-Test prüft der Smoke, dass
`pnpm m-trace check <vod.mpd>` ein Result mit
`analyzerKind:"dash"` / `playlistType:"dash"` und mindestens einer
`details.adaptationSets[]`-Entry liefert. Ein zweiter Block testet
den Negativ-Pfad: ein HTML-Body wird vom Detector als
`manifest_not_supported` zurückgewiesen (HTTP 422 in der API).

```bash
make smoke-cli  # 8 Probes inkl. DASH VOD und manifest_not_supported
```

Der DASH-Pfad nutzt die produktive Library, keine Stubs — die
Smoke-Schritte exerzieren den vollen Detector + MPD-Parser-Pfad.

#### Workflow-Übersicht

```bash
gh run watch --workflow build.yml
gh run list --workflow webrtc-drift.yml --limit 5
```

### 2.5 Benchmark-Regression-Gate (`0.9.5` Tranche 2, RAK-Wave-2)

Seit `plan-0.9.5` Tranche 2 ist
[`.github/workflows/benchmark.yml`](../../.github/workflows/benchmark.yml)
Nightly aktiv und failed bei statistisch signifikanten
Performance-Regressionen über +15 % (p < 0.05) gegenüber der
Baseline im orphan-Branch `benchmark-baseline`.

**Pflicht für Minor-Releases (`0.X.0`)**: Vor dem Release-Tag muss
der letzte Nightly-Lauf des `benchmark.yml`-Workflows **grün** sein
(oder als observation-only ohne Baseline gelaufen sein, falls die
Phase noch initial ist). Patch-Releases (`0.X.Y`) sind davon
ausgenommen — sie dürfen die Performance-Charakteristik nicht
ändern und werden über `make benchmark-smoke` (PR-Pfad)
abgesichert.

```bash
gh run list --workflow benchmark.yml --limit 5
gh run view <run-id>            # benchstat-Output im Log
gh run download <run-id>        # comparison.txt + current.txt + baseline.txt
```

Bei Regression öffnet der Workflow automatisch ein Issue mit
Workflow-Run-URL, vollständigem `benchstat`-Diff, lokalem
Repro-Befehl und dem Drift-Akzeptanz-Pfad. Ein offenes Issue
**blockiert** den Minor-Release-Tag, bis es geschlossen ist (Fix
landed oder Drift wurde durch Update des `benchmark-baseline`-
Branches akzeptiert).

**Quarantäne-Mechanik** (Plan-DoD-Wartungsregel): ein einzelner
lauter Benchmark kann temporär aus dem Vergleich genommen werden.
Format: ein Kommentar `// bench:quarantine YYYY-MM-DD reason: <text>`
**direkt** über der `func BenchmarkX(...)`-Definition (Go) bzw.
über dem `bench("...", ...)`-Aufruf (TS). Das Skript
[`scripts/check-bench-quarantines.mjs`](../../scripts/check-bench-quarantines.mjs)
läuft als erster Step im Workflow und failed, wenn ein Tag älter
als 30 Tage ist — Operator muss dann entweder den Bench fixen
und das Tag entfernen, oder das Tag mit einer Plan-DoD-Item-
Änderung im Folge-Plan verlängern (kein stiller Re-Skip).

```bash
# Manuelle Prüfung lokal:
node scripts/check-bench-quarantines.mjs apps/api packages/stream-analyzer
```

### 2.6 Fuzz- und Mutation-Beobachtungs-Gates (`0.9.5` Tranche 3+4, RAK-Wave-2)

Seit `plan-0.9.5` Tranchen 3 und 4 laufen zwei weitere Nightly-
Workflows als **nicht-blockierende Beobachtungs-Gates**:

- [`.github/workflows/fuzz.yml`](../../.github/workflows/fuzz.yml)
  (Cron `0 5 * * *` UTC, sechs Go-Fuzz-Targets, 5 min/Target).
  Crash-Funde landen als Issue mit Repo-Pfad
  `apps/api/<package>/testdata/fuzz/<Target>/<id>` (Labels
  `fuzz,quality,plan-0.9.5`); offenes Crash-Issue **blockiert
  den nächsten Release-Tag** (Patch *und* Minor), bis das
  Crash-File als Regression-Seed im Repo gelandet und der Bug
  gefixt ist.
  Operator-Doku: [`docs/dev/fuzzing.md`](../dev/fuzzing.md).
- [`.github/workflows/mutation.yml`](../../.github/workflows/mutation.yml)
  (Cron `0 6 * * *` UTC, gremlins für Go + StrykerJS für TS;
  beide Jobs `continue-on-error: true`). **Initial nicht-
  blockierend** (Plan-DoD §5: nur Reporting). Score-Trend wird
  über die HTML/JSON-Artefakte verfolgt; PR-Blockierung erst,
  wenn ein Modul drei Beobachtungsläufe in Folge > 70 % Score
  zeigt — Übergangs-Pfad in
  [`docs/dev/mutation-testing.md`](../dev/mutation-testing.md) §3.

PR-Pfad-Wrapper (opt-in, NICHT in `make gates`):

```bash
make fuzz-check        # FUZZTIME=30s pro Target (Default)
make mutation-report   # gremlins (Go) + StrykerJS (TS) auf den Pilot-Modulen
```

## 3. Release-Commit und Tag

Release-Konvention für `0.1.x`:

- trunk-based auf `main`.
- Release-Commit direkt auf `main`.
- annotierte SemVer-Tags im Format `vX.Y.Z`.
- kein Pre-Release-Suffix für Hauptreleases.
- keine automatische Veröffentlichung ohne explizite Freigabe.

### 3.0 Release-Guard (`0.13.0`)

Seit `0.13.0` gibt es einen lokalen Guard vor Tag/Publish. Er ersetzt
keine Qualitäts-Gates, sondern prüft den manuellen Freigabepunkt und
die wichtigsten Release-Anker:

```bash
MTRACE_RELEASE_APPROVED=1 make release-guard VER="$VER"
```

Der Guard prüft:

- explizite Freigabe über `MTRACE_RELEASE_APPROVED=1`;
- Branch `main` und saubere Arbeitskopie;
- Tag `vX.Y.Z` existiert weder lokal noch auf `origin`;
- `CHANGELOG.md`, API-`serviceVersion` und Root-`package.json` zeigen
  auf dieselbe Version.

Für lokale Tests am Guard selbst existieren drei bewusst benannte
Overrides: `MTRACE_RELEASE_ALLOW_NON_MAIN=1`,
`MTRACE_RELEASE_ALLOW_DIRTY=1` und
`MTRACE_RELEASE_ALLOW_OFFLINE=1`. Diese Overrides sind nicht für den
Release-Pfad zulässig und dürfen im Release-Log nicht gesetzt sein.

Ab `0.14.0` gibt es zusätzlich einen lokalen Guard-Self-Test:

```bash
make release-guard-test
```

Der Test legt temporäre Git-Repositories an und prüft die wichtigsten
Fehlerfälle ohne Netzwerkzugriff: fehlende Freigabe, `v`-Prefix im
Versionsargument, falscher Branch, Dirty Worktree, bereits vorhandener
lokaler Tag und Versionsdrift in `package.json`. Er erzeugt keinen Tag
und keinen Release.

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
Minor-Bump auch berührt:

- 5× `package.json` (Root + apps/analyzer-service + apps/dashboard +
  packages/player-sdk + packages/stream-analyzer)
- `apps/api/cmd/api/main.go` `serviceVersion`
- `packages/player-sdk/src/version.ts` `PLAYER_SDK_VERSION`
  (`pack-smoke.mjs` liest die erwartete Version dynamisch aus
  `package.json` — kein eigener Bump nötig)
- `contracts/sdk-compat.json` `sdk_version`
- alle 20 Analyzer-Spec-Fixtures unter
  `spec/contract-fixtures/analyzer/*.json` (`analyzerVersion`-Feld)
- die 20 testdata-Kopien in
  `apps/api/adapters/driven/streamanalyzer/testdata/` (über
  `make sync-contract-fixtures`)
- alle Go- und TS-Tests mit hartkodierten Versions-Strings

Sonst entsteht Drift zwischen SDK-Bundle, API-Service-Version und
CI-Smokes. **Bump-Pattern-Sweep vor Tag** (mindestens diese drei
Patterns muss der grep abdecken; das 0.12.0-Release hat zunächst
12 Stellen übersehen):

```bash
# In Source-Trees (apps/, packages/, spec/, contracts/):
grep -rn '"version":\s*"X\.Y\.Z"' --include='*.go' --include='*.ts' --include='*.json' --include='*.mjs'
grep -rn '"AnalyzerVersion":\s*"X\.Y\.Z"\|AnalyzerVersion:\s*"X\.Y\.Z"' --include='*.go' --include='*.ts' --include='*.json'
grep -rn '"sdk_version":\s*"X\.Y\.Z"\|PLAYER_SDK_VERSION\s*=\s*"X\.Y\.Z"\|serviceVersion\s*=\s*"X\.Y\.Z"' --include='*.go' --include='*.ts' --include='*.json'
```

Plan-DoD-Items ersetzen die RAK-Verifikationsmatrix; ein
Patch-Release-Plan trägt keinen `§6.1`-Block.

**Wave-2-Quality-Gates-Voraussetzung** (ab `0.9.5`): vor jedem
Release-Tag (Patch *und* Minor) zusätzlich prüfen *und im Release-
Log dokumentieren* (Plan-Closeout-Sektion oder Tag-Annotation
zitiert die geprüften Run-IDs):

- §2.5 Benchmark-Regression-Gate — letzter
  `benchmark.yml`-Nightly grün (Pflicht für Minor; Patch nur
  über `make benchmark-smoke` PR-Pfad).
  ```bash
  gh run list --workflow benchmark.yml --limit 1
  ```
- §2.6 Fuzz-Beobachtungs-Gate — kein offenes Issue mit Label
  `fuzz` aus dem letzten `fuzz.yml`-Nightly. Offenes
  Crash-Issue blockt den Tag, bis das Crash-File als
  Regression-Seed im Repo gelandet ist.
  ```bash
  gh run list --workflow fuzz.yml --limit 1
  gh issue list --label fuzz --state open
  ```
- §2.6 Mutation-Beobachtungs-Gate — Score-Trend in den letzten
  drei Nightly-Artefakten geprüft (kein hartes Gate; Score-
  Senkung ist begründungspflichtig, siehe
  [`docs/dev/mutation-testing.md`](../dev/mutation-testing.md)
  §3).
  ```bash
  gh run list --workflow mutation.yml --limit 3
  ```

**Im Release-Log (Plan §8 Closeout-DoD oder Tag-Annotation) das
Verdict festhalten** — z. B. „Wave-2-Gates: benchmark.yml run
12345678 ✅, fuzz.yml run 12345679 ✅, mutation.yml letzte 3 Runs
Score-Trend stabil". Beim `0.12.0`-Release waren die Nightlies
grün, aber das Verdict wurde nicht dokumentiert (Lehre für
`0.13.0`).

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
