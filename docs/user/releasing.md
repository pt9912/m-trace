# Releasing — m-trace

> **Bezug**: AK-11, F-131.

## 0. Zweck

Dieses Dokument beschreibt den minimalen, reproduzierbaren
Release-Ablauf für m-trace. Der Ablauf ist versionsunabhängig
formuliert und verwendet Platzhalter der Form `X.Y.Z`. Er gilt für
alle Releases aus dem Release-Plan (RAK-1..RAK-46).

## 1. Vorbereitung

```bash
VER="X.Y.Z"
TAG="v$VER"
```

Vor jedem Release:

- noch nicht veröffentlichte Änderungen stehen unter `## [Unreleased]`
  in [`CHANGELOG.md`](../../CHANGELOG.md); ein datierter Versionsabschnitt entsteht erst mit
  dem Release-Commit.
- [`CHANGELOG.md`](../../CHANGELOG.md) auf den Zielstand bringen.
- betroffene Plan-, Status- und Nutzungsdokumente aktualisieren
  ([`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md), [`spec/architecture.md`](../../spec/architecture.md), [`apps/api/README.md`](../../apps/api/README.md)).
- Roadmap §1.1 und §1.2 nach dem Release-Bump neu schreiben (siehe
  [`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md) Wartungsregel).
- offene `OE-X` und `R-X` durchsehen — Einträge, die mit dem Release
  aufgelöst werden, aus den Tabellen entfernen.

## 2. Verifikation

Vor Tag und GitHub-Release müssen die Root-Targets grün sein.

> **Ein-Befehl-Gate**: `MTRACE_RELEASE_APPROVED=1 make release-gate
> VER="$VER"` bündelt den kompletten Block unten (`gates` +
> `security-gates` inkl. Trivy-`image-scan` + `build` + Release-Smokes +
> Publish-Dry-Runs) und schließt mit dem Release-Guard. Stack-abhängige
> Smokes (`smoke-mediamtx` braucht das Core-Lab, `smoke-observability`
> den Observability-Stack) werden detached hoch- und wieder abgefahren.
> Die Verifikation läuft auch ohne Freigabe vollständig durch; nur der
> finale Guard-Stempel braucht `MTRACE_RELEASE_APPROVED=1`. Die
> Einzelaufrufe bleiben für gezielte Wiederholung gültig:

```bash
make gates                # CI-äquivalenter Komplettcheck (api-race+ts-test+lint+coverage+arch+schema+docs)
make build
make sdk-performance-smoke
make smoke-cli            # Lastenheft-Aufruf `pnpm m-trace check`
make smoke-analyzer       # manuelles Release-Gate, fährt Compose hoch
make smoke-observability  # Cardinality-Smoke; Observability-Stack muss laufen
make browser-e2e          # Dashboard-Timeline + hls.js-Demo-Flow
make dev-detached         # Core-Lab für smoke-mediamtx starten; danach `make stop`
make smoke-mediamtx       # MediaMTX-Beispiel (RAK-36); braucht laufendes Core-Lab
make smoke-srt            # SRT-Beispiel (RAK-37); startet/stoppt Project mtrace-srt
make smoke-srt-health     # SRT-Health-Smoke (RAK-41/RAK-42); startet/stoppt mtrace-srt + probt MediaMTX-API
make smoke-dash           # DASH-Beispiel (RAK-38); startet/stoppt Project mtrace-dash
make smoke-webrtc-prep    # WebRTC-Lab-Vorbereitungs-Smoke (RAK-48); startet/stoppt mtrace-webrtc; endpoint-only (kein Browser/Playback/getStats)
make smoke-webrtc-stats-drift # WebRTC-`getStats()`-Drift-Smoke (RAK-56); startet/stoppt mtrace-webrtc + ruft Playwright-Spec gegen Chromium/Firefox; opt-in lokal, produktiv im Nightly-Workflow `.github/workflows/webrtc-drift.yml`
make package-publish-dry-run # baut und prüft die zwei GitHub-Packages-npm-Artefakte ohne Veröffentlichung
make image-publish-dry-run VER="$VER" # baut und prüft die drei GHCR-Runtime-Images ohne Veröffentlichung
```

Erfolgskriterien:

- alle Targets exit code 0.
- `make gates` umfasst `make api-race` (Go-Tests mit Race-Detector,
  CGO=1; in gates statt `api-test`, weil
  Race-Detection ein Superset ist), `make ts-test`, `make lint`,
  `make coverage-gate`, `make arch-check`, `make schema-validate`
  und `make docs-check` — einzelne Aufrufe sind möglich, aber
  `make gates` ist die CI-äquivalente Eingangsstufe.
- `make coverage-gate` (Teil von `make gates`) umfasst API-,
  Player-SDK-, Dashboard-, stream-analyzer- und
  analyzer-service-Coverage.
- `golangci-lint`-Stage liefert keine Findings.
- `go test ./...` deckt mindestens die Pflichttests aus
  [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md) ab.
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
`browser-e2e` und `smoke-mediamtx`/`smoke-srt`/
`smoke-dash` (plus `smoke-srt-health`,
`smoke-webrtc-prep`) laufen lokal vor dem Tag (Compose-Stack-Up bzw.
Browser-Stack ist zu schwergewichtig für jeden PR-Run). CI-Zielplattform
ist GitHub Actions auf `ubuntu-24.04`, Workflow-Name: `build`.

### 2.1 Manuelle SRT-Health-Prüfungen

Zusätzlich zu den oben gelisteten Smokes braucht der Release
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
   RAK-43-Pflichtwerte im Wire-Format aus [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md).
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

### 2.2 Manuelle WebRTC-Lab-Prüfungen

Zusätzlich zu `make smoke-webrtc-prep` (auto-up/down, endpoint-only)
braucht der Release einen kurzen manuellen Browser-Handcheck
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
   [`spec/telemetry-model.md`](../../spec/telemetry-model.md) als Muss-Felder dokumentiert.
5. `docker compose -p mtrace-webrtc … down` räumt nur den
   `mtrace-webrtc`-Stack ab; Core-Lab und andere Beispiele bleiben
   unangetastet.

Vollständige Operator-Doku:
[`examples/webrtc/README.md`](../../examples/webrtc/README.md).

### 2.3 Manuelle Player-SDK-WebRTC-Prüfungen

Zusätzlich zu `make smoke-webrtc-prep` und dem WebRTC-Browser-
Handcheck braucht der Release einen produktiven End-to-End-
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
   [`spec/telemetry-model.md`](../../spec/telemetry-model.md)).
4. Optional automatisierter Browser-E2E:
   `MTRACE_WEBRTC_LAB=1 make browser-e2e` flippt
   `tests/e2e/dashboard-demo-webrtc.spec.ts` auf den Happy-Path
   (`playback_started` mit `webrtc.peer_connection_run_id`).
5. Stop: `docker compose -p mtrace-webrtc … down`. Greift weder
   Core-Lab noch andere Beispiele an.

> **Automatisierte Teilabnahme**: Der „1 kHz Sinuston"-Anteil von
> Schritt 2 ist über `make smoke-webrtc-tone`
> ([`plan-0.22.4-webrtc-tone-smoke`](../planning/done/plan-0.22.4-webrtc-tone-smoke.md))
> und den `webrtc-drift.yml`-Nightly automatisiert: ein
> FFT/Goertzel-Check auf dem RTSP-Egress belegt „ein sauberer
> 1-kHz-Ton liegt an und dominiert". Die *perzeptuelle* Abnahme
> (latenzarm, subjektiv sauber, Gesamt-Demo-Verhalten im echten
> Browser) bleibt manuell — der Smoke ist nicht-blockierend
> (`extra-gates.md` §3.8).

Vollständige Operator-Doku:
[`packages/player-sdk/README.md`](../../packages/player-sdk/README.md)
§Performance and Browser Support.

### 2.4 Manuelle Drift-Smoke- / DASH- / SRS-Prüfungen

Drei thematisch getrennte Liefergegenstände
; vor dem Release-Tag laufen drei
operative Verifikationspfade an, die alle als opt-in Smokes lokal
reproduzierbar sind und nicht in `make gates` enthalten:

#### 2.4.1 WebRTC-Drift-Smoke (RAK-56)

Der Drift-Review aus R-12 ist automatisiert.
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

#### 2.4.2 SRS-Lab-Boot (RAK-57)

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

#### 2.4.3 DASH-CLI-Probe (RAK-58, RAK-59)

`make smoke-cli` ist um einen DASH-Pfad
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

### 2.5 Benchmark-Regression-Gate

ist
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

### 2.6 Fuzz- und Mutation-Beobachtungs-Gates

Laufen aktuell zwei weitere Nightly-
Workflows als **nicht-blockierende Beobachtungs-Gates**:

- [`.github/workflows/fuzz.yml`](../../.github/workflows/fuzz.yml)
  (Cron `0 5 * * *` UTC, sechs Go-Fuzz-Targets, 5 min/Target).
  Crash-Funde landen als Issue mit Repo-Pfad
  `apps/api/<package>/testdata/fuzz/<Target>/<id>` (Labels
  `fuzz,quality`); offenes Crash-Issue **blockiert
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
- [`.github/workflows/load-smoke.yml`](../../.github/workflows/load-smoke.yml)
  (Cron `52 1 * * *` Europe/Berlin) — open-loop SLO (constant-arrival-
  rate, p95-Budget + `dropped_iterations`) gegen das Core-Lab,
  `continue-on-error`; Verdikt aus Job-Summary + Artefakt, nicht aus der
  Job-Farbe. Belegt die Lab-Lastfähigkeit (NF-20/NF-22/NF-23) und „kein
  stiller Verlust" per Readback gegen `playback_events`. Der Soak
  (Read-Retention-p95 gegen 2 s = `ADR-0005` Trigger #3, ≥ 10 Mio Events
  ~Stunden) läuft on-demand via `workflow_dispatch` (`mode=soak`).
  Details: [`extra-gates.md`](../planning/in-progress/extra-gates.md)
  §3.9.

PR-Pfad-Wrapper (opt-in, NICHT in `make gates`):

```bash
make fuzz-check        # FUZZTIME=30s pro Target (Default)
make mutation-report   # gremlins (Go) + StrykerJS (TS) auf den Pilot-Modulen
make smoke-load        # Last-Smoke (closed-loop, Reconciliation) gegen das Core-Lab
make smoke-load-slo    # open-loop SLO (constant-arrival-rate, p95-Budget)
make smoke-soak        # + Retention-Probe (ADR-0005 Trigger #3; lange DURATION fuer >=10M)
```

## 3. Release-Commit und Tag

Release-Konvention für `0.1.x`:

- trunk-based auf `main`.
- Release-Commit direkt auf `main`.
- annotierte SemVer-Tags im Format `vX.Y.Z`.
- kein Pre-Release-Suffix für Hauptreleases.
- keine automatische Veröffentlichung ohne explizite Freigabe.

### 3.0 Release-Guard

gibt es einen lokalen Guard vor Tag/Publish. Er ersetzt
keine Qualitäts-Gates, sondern prüft den manuellen Freigabepunkt und
die wichtigsten Release-Anker:

```bash
MTRACE_RELEASE_APPROVED=1 make release-guard VER="$VER"
```

Der Guard prüft:

- explizite Freigabe über `MTRACE_RELEASE_APPROVED=1`;
- Branch `main` und saubere Arbeitskopie;
- Tag `vX.Y.Z` existiert weder lokal noch auf `origin`;
- [`CHANGELOG.md`](../../CHANGELOG.md), API-`serviceVersion` und Root-`package.json` zeigen
  auf dieselbe Version.

Für lokale Tests am Guard selbst existieren drei bewusst benannte
Overrides: `MTRACE_RELEASE_ALLOW_NON_MAIN=1`,
`MTRACE_RELEASE_ALLOW_DIRTY=1` und
`MTRACE_RELEASE_ALLOW_OFFLINE=1`. Diese Overrides sind nicht für den
Release-Pfad zulässig und dürfen im Release-Log nicht gesetzt sein.

Zusätzlich gibt es einen lokalen Guard-Self-Test:

```bash
make release-guard-test
```

Der Test legt temporäre Git-Repositories an und prüft die wichtigsten
Fehlerfälle ohne Netzwerkzugriff: fehlende Freigabe, `v`-Prefix im
Versionsargument, falscher Branch, Dirty Worktree, bereits vorhandener
lokaler Tag und Versionsdrift in `package.json`. Er erzeugt keinen Tag
und keinen Release.

### 3.1 Patch-Release-Konvention (`0.X.Y`)

Patch-
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
  [`spec/contract-fixtures/analyzer/*.json`](../../spec/contract-fixtures/analyzer) (`analyzerVersion`-Feld)
- die 20 testdata-Kopien in
  `apps/api/adapters/driven/streamanalyzer/testdata/` (über
  `make sync-contract-fixtures`)
- alle Go- und TS-Tests mit hartkodierten Versions-Strings

Sonst entsteht Drift zwischen SDK-Bundle, API-Service-Version und
CI-Smokes. **Bump-Pattern-Sweep vor Tag** (mindestens diese drei
Patterns muss der grep abdecken; ein älterer Release hat zunächst
12 Stellen übersehen):

```bash
# In Source-Trees (apps/, packages/, spec/, contracts/):
grep -rn '"version":\s*"X\.Y\.Z"' --include='*.go' --include='*.ts' --include='*.json' --include='*.mjs'
grep -rn '"AnalyzerVersion":\s*"X\.Y\.Z"\|AnalyzerVersion:\s*"X\.Y\.Z"' --include='*.go' --include='*.ts' --include='*.json'
grep -rn '"sdk_version":\s*"X\.Y\.Z"\|PLAYER_SDK_VERSION\s*=\s*"X\.Y\.Z"\|serviceVersion\s*=\s*"X\.Y\.Z"' --include='*.go' --include='*.ts' --include='*.json'
```

Plan-DoD-Items ersetzen die RAK-Verifikationsmatrix; ein
Patch-Release-Plan trägt keinen `§6.1`-Block.

**Quality-Gates-Voraussetzung**: vor jedem
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
Verdict festhalten** — z. B. „Quality-Gates: benchmark.yml run
12345678 ✅, fuzz.yml run 12345679 ✅, mutation.yml letzte 3 Runs
Score-Trend stabil". Beim Release waren die Nightlies
grün, aber das Verdict wurde nicht dokumentiert (Lehre für
.

```bash
git commit -m "chore(release): vX.Y.Z"
git tag -a "$TAG" -m "Release X.Y.Z"
git push origin main
git push origin "$TAG"
```

> Der Push des `vX.Y.Z`-Tags löst `build.yml` (CI-Gates + Security-Gates
> inkl. `image-scan`) auf dem getaggten Commit aus — ein sichtbarer
> roter/grüner Lauf auf dem Release-Tag. Das ist kein harter Block (der
> Tag existiert bereits); die harte Vorab-Absicherung ist `make
> release-gate` weiter oben. Ein roter Tag-Lauf ist das Signal, vor dem
> GitHub-Release zu stoppen und ggf. per Rollback (siehe unten) den Tag
> zurückzuziehen.

## 4. GitHub-Release

Mindestumfang:

- Release-Notes aus dem [`CHANGELOG.md`](../../CHANGELOG.md)-Versionsabschnitt extrahieren.
- Release-Titel: `m-trace X.Y.Z`.
- Tag: `vX.Y.Z`.
- Assets: GitHub-Source-Archive (`zip`/`tar.gz`) plus
  GitHub-Packages-Publish für die publishbaren npm-Pakete und ab
  GHCR-Publish für die drei Runtime-Images.

```bash
gh release create "$TAG" \
    --title "m-trace $VER" \
    --notes-file <changelog-extract>
```

## 5. GitHub-Packages-Publish

  werden nur die zwei Library-/CLI-Pakete veröffentlicht:

- `@pt9912/player-sdk`
- `@pt9912/stream-analyzer`

Die Apps `@pt9912/m-trace-dashboard` und
`@pt9912/analyzer-service` bleiben `private: true` und werden nicht
als npm-Pakete veröffentlicht.

GitHub-Packages-Voraussetzungen:

- Paketnamen sind auf den GitHub-Owner-Scope `@pt9912` gemappt.
- `.npmrc` enthält `@pt9912:registry=https://npm.pkg.github.com`.
- `publishConfig.registry` zeigt in beiden publishbaren Paketen auf
  `https://npm.pkg.github.com`.
- `.github/workflows/publish-packages.yml` nutzt `GITHUB_TOKEN` mit
  `packages: write`; der Workflow kann manuell trocken oder produktiv
  gegen einen Tag laufen.

Lokaler Package-Dry-Run vor Tag:

```bash
make package-publish-dry-run
```

Manueller GitHub-Actions-Dry-Run nach Tag, aber vor GitHub-Release:

```bash
gh workflow run publish-packages.yml \
    --ref main \
    -f ref="$TAG" \
    -f dry_run=true
```

Produktiver Publish ohne GitHub-Release-Automatik, z. B. für eine
gezielte Reparatur, nachdem geprüft wurde, dass dieselbe Version noch
nicht veröffentlicht wurde:

```bash
gh workflow run publish-packages.yml \
    --ref main \
    -f ref="$TAG" \
    -f dry_run=false
```

Der normale Release-Pfad ist: erst Dry-Run, dann `gh release create`.
Der Workflow wird bei `release.published` automatisch ausgeführt und
veröffentlicht dann den Release-Tag. Der manuelle produktive Publish
ist nur ein Ersatzpfad, damit dieselbe Version nicht doppelt
veröffentlicht wird. Produktive Veröffentlichungen laufen intern über:

```bash
MTRACE_PACKAGE_PUBLISH_APPROVED=1 make package-publish
```

## 6. GHCR-Image-Publish

  werden die drei Runtime-Images versioniert auf GHCR
veröffentlicht:

- `ghcr.io/pt9912/m-trace-api:$VER`
- `ghcr.io/pt9912/m-trace-dashboard:$VER`
- `ghcr.io/pt9912/m-trace-analyzer-service:$VER`

`latest` wird bewusst nicht gesetzt. Der Release-Tag bleibt die
einzige veröffentlichte Tag-Quelle; K8s- und Compose-Beispiele sollen
weiterhin konkrete Versions-Tags referenzieren.

Lokaler Image-Dry-Run vor Tag:

```bash
make image-publish-dry-run VER="$VER"
```

Manueller GitHub-Actions-Dry-Run nach Tag, aber vor GitHub-Release:

```bash
gh workflow run publish-images.yml \
    --ref main \
    -f ref="$TAG" \
    -f image_tag="$VER" \
    -f dry_run=true
```

Der normale Release-Pfad ist: erst lokaler oder manueller Dry-Run,
dann `gh release create`. Der Workflow
`.github/workflows/publish-images.yml` wird bei `release.published`
automatisch ausgeführt und veröffentlicht den Release-Tag.

Produktive Veröffentlichungen laufen intern über:

```bash
MTRACE_IMAGE_PUBLISH_APPROVED=1 make image-publish VER="$VER"
```

Voraussetzungen:

- `make image-scan` ist vor dem Release grün.
- Der GitHub-Actions-Workflow hat `packages: write`.
- Die Veröffentlichung erfolgt nur für den Release-Tag und nur nach
  Human Approval durch GitHub Release oder expliziten manuellen
  Workflow-Aufruf.

## 7. Post-Release

- [`CHANGELOG.md`](../../CHANGELOG.md) öffnet einen neuen `## [Unreleased]`-Abschnitt.
- [`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md) (Release-Übersicht) aktualisiert den Status
  des veröffentlichten Releases (`⬜ → ✅`).
- Folge-ADRs, die mit dem Release entstehen oder fällig werden,
  in [`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md) ergänzen.
- **Spec-Header aktualisieren**, falls die Lieferung normative
  Spec-Sections berührt: Header in den betroffenen
  `spec/*.md`-Dokumenten (Stand-Marker und Lastenheft-Patch-Ref)
  auf den neuen Release-/Lastenheft-Stand setzen.
  Spec-Inline-Versionsmarker (`ab 0.X.Y`, `seit 0.X.Y`,
  `in 0.X.Y`) sind bewusst **nicht** Teil der Spec — das Lieferzeit-
  Audit-Trail steht im [`CHANGELOG.md`](../../CHANGELOG.md) und in den
  `docs/planning/done/plan-X.Y.Z.md`-Dokumenten (Variante-B-
  Konvention; siehe Spec-Kopf von [`spec/telemetry-model.md`](../../spec/telemetry-model.md)).
- **Spec referenziert das Lastenheft nur über Kennungen**
  (`F-XXX`, `NF-XXX`, `MVP-XXX`, `AK-XXX`, `RAK-XXX`), nicht über
  Paragraph-Nummern (`§7.10`, `§4.3` etc.). Die Lastenheft-
  Paragraph-Struktur wandert über die Patch-Versionen mit;
  Kennungen sind stabil und versionsversiegelt. Bei neuen
  RAK-Aufnahmen die Section-`Bezug:`-Refs auf die korrekte
  versionsversiegelte Lastenheft-Patch-Version setzen
  (Geburtsort der RAK-Familie, nicht aktueller Stand).
- **Spec referenziert keine Plan-Dokumente.** Pläne
  (`docs/planning/done/plan-X.Y.Z.md`) sind dem Lastenheft
  nachrangig — sie beschreiben *wie* eine Lieferung umgesetzt
  wurde, nicht *was* normativ wahr ist. Der Spec-Inhalt steht
  als eigenständige Wahrheit, gestützt nur durch Lastenheft-
  Kennungen, ADRs und stabile Code-Anchor (Modul-/Dateipfade
  unter `apps/api`, `packages/*`). Plan-Verweise im Spec-Text
  sind Doku-Drift und gehören in CHANGELOG-/Commit-Texte, nicht
  in die Spec.

## 8. Rollback

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

Package-Publish fehlgeschlagen:

- Solange kein Paket veröffentlicht wurde: Workflow-Fehler auf `main`
  beheben, neuen Commit erstellen und denselben Tag erneut publishen.
- Wenn ein Paket bereits veröffentlicht wurde: Version nicht löschen
  oder überschreiben; Folge-Patch-Release erstellen und beide Pakete
  erneut konsistent veröffentlichen.

Image-Publish fehlgeschlagen:

- Solange kein Image gepusht wurde: Workflow-Fehler auf `main`
  beheben und denselben Tag erneut über `publish-images.yml`
  publishen.
- Wenn nur ein Teil der Images veröffentlicht wurde: fehlende Images
  mit demselben Git-Ref nachpublishen; bereits veröffentlichte Tags
  nicht überschreiben.
- Wenn ein veröffentlichtes Image fehlerhaft ist: Folge-Patch-Release
  erstellen. Versionierte GHCR-Tags werden nicht mutiert.

## 9. Referenzen

- Akzeptanzkriterien (AK-11).
- Definition of Done.
- [`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md) — Release-Übersicht und RAK-Akzeptanzkriterien.
- [`docs/planning/in-progress/roadmap.md`](../planning/in-progress/roadmap.md) — Offene Entscheidungen.
- `CHANGELOG.md` — Versionsverlauf.
