# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.5] - 2026-05-07

> **Quality-Gates Wave 2** — Patch-Release nach `0.9.0`/`0.9.1`
> (Patch-Release-Konvention `0.X.Y`, siehe
> [`docs/user/releasing.md`](docs/user/releasing.md) §3.1) ohne
> User-Surface-Änderung. Liefert die vier statistisch- bzw.
> langlaufenden Quality-Gates aus
> [`docs/planning/in-progress/extra-gates.md`](docs/planning/in-progress/extra-gates.md):
> Benchmark-Smoke (PR-Pfad) + Nightly-`benchstat`-Regressionen
> (§3.2/§3.3), selektives Fuzzing + TS-Property-Tests (§3.5) und
> Mutation-Testing als Nightly-Report (§3.6). Kein Lastenheft-Patch
> (Quality-Gates, keine User-Surface). Plan in
> [`done/plan-0.9.5.md`](docs/planning/done/plan-0.9.5.md).

### Added (Tranche 0 — Plan-Aktivierung + Baseline-Entscheidungen)

- [`docs/perf/budgets.md`](docs/perf/budgets.md) als Single-Source
  für Performance-Budgets pro Modul (API + Stream-Analyzer);
  initiale Werte als „Tranche-0-Stand, noch nicht mess-basiert"
  markiert; Tranche-1-Beobachtungsphase schärft sie nach.
- Baseline-Pfad für Tranche 2 entschieden: Git-Branch
  `benchmark-baseline` (orphan, File `benchmarks/api-bench.txt`),
  begründet in `plan-0.9.5.md` §1a Tranche 0.
- Quarantäne-Policy: maximal 30 Tage Skip mit Begründungs-
  Kommentar plus Backlog-Item in
  [`docs/planning/in-progress/risks-backlog.md`](docs/planning/in-progress/risks-backlog.md);
  Verlängerung ist Plan-DoD-Item-Änderung im jeweiligen Folge-Plan.

### Added (Tranche 1 — Benchmark-Smoke API + Stream-Analyzer)

- Go-Benchmark-Suite in `apps/api` für vier Hot-Paths (Plan-DoD
  §2-1): `BenchmarkRegisterPlaybackEventBatch_{Typical,MaxBatch}`,
  `BenchmarkEventRepository_AppendBatch_100`,
  `BenchmarkSessionsService_ListSessions_DefaultPage`,
  `BenchmarkCursorEncodeDecode_Pair`. Pfade in
  `*_bench_test.go`/`*_bench_internal_test.go`.
- TS-Benchmark-Suite
  `packages/stream-analyzer/benchmarks/analyzer.bench.ts` für
  sieben Hot-Paths (Plan-DoD §2-2): HLS Master/Media,
  DASH-MPD VOD/Live, Detector über 256-KiB-Body, SSRF-URL-
  Klassifizierung. Synthetische Fixtures deterministisch im
  Bench-File generiert; separater
  `vitest.bench.config.ts`.
- `make api-benchmark-smoke` + `make analyzer-benchmark-smoke` +
  Wrapper `make benchmark-smoke` (Plan-DoD §2-3); beide drucken
  zuerst Runner-Info via `scripts/print-bench-runner-info.sh`.
- `scripts/check-bench-budgets.mjs` parst beide Bench-Backends
  (Vitest-Bench-stdout + Go-Bench-stdout) gegen die Budget-
  Tabelle aus `docs/perf/budgets.md`; Output-Form
  `[bench-budget] FAIL <name>: ist=<X> ms soll=<Y> ms`.
- `.github/workflows/benchmark-observation.yml` (Cron `30 2 * * *`
  UTC + `workflow_dispatch`) als Beobachtungs-Nightly mit
  `continue-on-error: true`; Bench-Output als Artefakt
  `bench-observation-<run_id>` mit 14 Tagen Retention.
- `make benchmark-smoke` ist **nicht** in `make gates`; PR-
  Blockierung folgt nach N=3..5 grünen Beobachtungsläufen
  (Folge-Commit nimmt `continue-on-error` raus + Aufnahme in
  `make gates`).

### Added (Tranche 2 — Nightly-`benchstat`-Regressionen)

- `.github/workflows/benchmark.yml` (Cron `0 4 * * *` UTC +
  `workflow_dispatch`): `go test -bench=. -benchmem -count=10
  -benchtime=2s` auf API-Hot-Paths, Baseline aus orphan-Branch
  `benchmark-baseline` als File `benchmarks/api-bench.txt`,
  Vergleich via `benchstat` aus `golang.org/x/perf`.
- `scripts/check-benchstat-regression.mjs` mit
  `--threshold-percent`-Flag (Default 15): Regressions-Schwelle
  +15 % auf statistisch signifikantem Ergebnis (p < 0.05).
- benchstat-Output als Workflow-Artefakt
  `bench-regression-<run_id>` (30 Tage Retention) mit
  `current.txt`, `baseline.txt`, `comparison.txt`.
- Auto-Issue bei Regression mit Workflow-Run-URL,
  benchstat-Diff-Block, Repro-Befehl und Drift-Akzeptanz-Pfad;
  Labels `performance,benchmark,plan-0.9.5`.
- Quarantäne-Mechanik: `// bench:quarantine YYYY-MM-DD reason:
  <text>`-Kommentar direkt über `func BenchmarkX(...)` (Go) bzw.
  dem `bench("...", ...)`-Aufruf (TS). Skript
  `scripts/check-bench-quarantines.mjs` scant `apps/api` und
  `packages/stream-analyzer` und failed bei Tag älter als 30 Tage.
- Release-Gate-Doku in
  [`docs/user/releasing.md`](docs/user/releasing.md) §2.5
  („Benchmark-Regression-Gate"): Pflicht-Voraussetzung für
  Minor-Releases; Patch-Releases sind ausgenommen.

### Added (Tranche 3 — Selektives Fuzzing + Property Tests)

- Sechs Go-Fuzz-Targets in vier Packages (Plan-DoD §4-1):
  - `apps/api/adapters/driving/http/cursor_fuzz_internal_test.go`:
    `FuzzDecodeListSessionsCursor`, `FuzzDecodeSessionEventsCursor`.
  - `apps/api/adapters/driving/http/wire_fuzz_internal_test.go`:
    `FuzzWireBatchDecode`.
  - `apps/api/hexagon/application/event_meta_validation_fuzz_internal_test.go`:
    `FuzzValidateReservedEventMeta`, `FuzzValidateUnavailableReason`.
  - `apps/api/adapters/driven/srt/mediamtxclient/mapping_fuzz_internal_test.go`:
    `FuzzMapMediaMtxItem`.
- Drei TS-Property-Test-Suites via `fast-check@4.4.0` (Plan-DoD
  §4-2):
  - `packages/stream-analyzer/tests/hls-parser.property.test.ts`
    (zwei Properties).
  - `packages/stream-analyzer/tests/dash-parser.property.test.ts`
    (drei Properties).
  - `packages/player-sdk/tests/redact.property.test.ts` (drei
    Properties).
- `make api-fuzz-check` + `make fuzz-check` (Plan-DoD §4-3) mit
  `FUZZTIME`-Override (Default 30 s pro Target); greppt
  `^func Fuzz...` automatisch — keine Registry-Pflege. Opt-in
  (NICHT in `make gates`).
- `.github/workflows/fuzz.yml` (Cron `0 5 * * *` UTC +
  `workflow_dispatch`): Default 5 min pro Target ⇒ ≈ 30 min Total;
  Crash-Inputs via `find -newer go.mod` aus
  `testdata/fuzz/<Target>/`; Artefakt `fuzz-nightly-<run_id>` mit
  30 Tagen Retention; Auto-Issue mit Repo-Pfad
  `apps/api/<package>/testdata/fuzz/<Target>/<id>` als Regression-
  Seed (Labels `fuzz,quality,plan-0.9.5`).
- [`docs/dev/fuzzing.md`](docs/dev/fuzzing.md): Liste der Fuzz-
  Targets, lokale Reproduktion, Crash-Repro-Pfad aus dem Nightly,
  Korpus-Schichten, fast-check-Discard-Loop-Lehre.

### Fixed (Tranche 3 — Erstfund durch FuzzMapMediaMtxItem)

- `apps/api/adapters/driven/srt/mediamtxclient/mapping.go`:
  `mbpsLinkCapacity=-1` produzierte
  `AvailableBandwidthBPS=-1_000_000` (negativer Wert leakt durch
  in den Wire-Vertrag). Fix: `AvailableBandwidthBPS` wird nur
  gesetzt, wenn `mbpsLinkCapacity > 0`; sonst bleibt das Feld
  leer und `state` wechselt auf `unknown`. Erstfund vom Fuzz-
  Target im selben Tranche-3a-Commit.

### Added (Tranche 4 — Mutation Testing als Nightly-Report)

- Tool-Auswahl: **gremlins** (`github.com/go-gremlins/gremlins`)
  für Go statt go-mutesting (Substitution begründet in
  [`docs/dev/mutation-testing.md`](docs/dev/mutation-testing.md)
  §1: go-mutesting seit ~2022 unmaintained, AST-Brüche auf
  Go 1.21+); **StrykerJS** + `@stryker-mutator/vitest-runner` für
  TS.
- Pilot-Module:
  `apps/api/hexagon/application/event_meta_validation.go` (Go)
  und `packages/player-sdk/src/adapters/webrtc/sampling.ts` (TS)
  — beide sicherheits-relevant.
- `make api-mutation-report` (gremlins via golang:1.26-Container,
  `go install`-zur-Laufzeit) + `make ts-mutation-report` (Stryker
  via `pnpm dlx`, kein devDep-Pinning im player-sdk) + Wrapper
  `make mutation-report`. Stryker-Konfig
  `packages/player-sdk/stryker.conf.cjs` mit `mutate`-Scope auf
  das eine File. Opt-in (NICHT in `make gates`).
- `.github/workflows/mutation.yml` (Cron `0 6 * * *` UTC +
  `workflow_dispatch`): zwei Jobs (`mutation-go` + `mutation-ts`),
  beide `continue-on-error: true` (nicht-blockierend); Artefakte
  `mutation-{go,ts}-<run_id>` mit 30 Tagen Retention.
- Score-Schwelle dokumentiert in
  [`docs/dev/mutation-testing.md`](docs/dev/mutation-testing.md)
  §3: > 70 % Wunsch-Ziel; PR-Blockierung erst, wenn das Modul
  drei Nightly-Runs in Folge > 70 % erreicht.

### Changed

- Versions-Bump auf `0.9.5` (Patch-Release): alle 5 `package.json`
  (root, `apps/dashboard`, `apps/analyzer-service`,
  `packages/player-sdk`, `packages/stream-analyzer`),
  `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts` `PLAYER_SDK_VERSION`,
  `packages/player-sdk/scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` plus Test-Fixtures
  und Contract-Fixtures, die Versions-Strings hartkodieren.
  Gleicher Bulk-Sed-Pfad wie `0.8.5`/`0.9.0`/`0.9.1` Closeout.
- [`docs/user/releasing.md`](docs/user/releasing.md) §3 erwähnt
  Wave-2-Gates (`make benchmark-smoke` opt-in/PR-blockierend
  nach Beobachtungsphase, `make fuzz-check` und
  `make mutation-report` opt-in/Nightly).

## [0.9.1] - 2026-05-07

> Wartungs-Patch nach `0.9.0` (Patch-Release-Konvention `0.X.Y`,
> siehe [`docs/user/releasing.md`](docs/user/releasing.md) §3.1):
> WebRTC-Drift-Smoke robuster gegen reale Browser-Eigenheiten plus
> Pfad-Korrekturen nach dem `git mv` von `plan-0.9.0.md` zu `done/`.
> Kein Lastenheft-Patch, kein eigener Plan-File — alle Inhalte
> stammen aus dem Wartungs-Commit `f3a0ddf` direkt nach dem
> `0.9.0`-Tag.

### Fixed

- WebRTC-Drift-Smoke (`tests/e2e/webrtc-stats-drift.spec.ts`)
  robuster gemacht für reale Browser-Eigenheiten:
  - WHEP-Signalisierung läuft jetzt aus dem Playwright-Node-Kontext
    statt browserseitig — vermeidet Browser-CORS-Abhängigkeiten des
    lokalen MediaMTX-WHEP-Endpoints; die `RTCPeerConnection` und
    alle `getStats()`-Reports stammen weiterhin aus echten Browsern.
  - Firefox im Smoke audio-only (Chromium bleibt video+audio), weil
    die Playwright-Firefox-Linie in dieser Umgebung keinen
    kompatiblen Videocodec für den MediaMTX-Lab-Stream anbietet.
  - Fehlende `RTCStatsType.transport`-Reports werden als
    `[drift-soll]` geloggt statt als harter Drift-Fail gewertet —
    folgt damit `spec/telemetry-model.md` §3.5.3 („Metrik leer
    statt `unknown`-Surrogat", per Engine).
  - `peer-connection.connectionState` wird über die normative
    `pc.connectionState`-API geprüft, weil aktuelle Browser das
    Feld nicht durchgängig im `peer-connection`-Stats-Report
    spiegeln.
- `playwright.config.ts` und `scripts/smoke-webrtc-stats-drift.sh`
  unterstützen `PLAYWRIGHT_TEST_RESULTS_DIR` als Env-Override; der
  Drift-Smoke schreibt seine Artefakte standardmäßig nach
  `${TMPDIR:-/tmp}/mtrace-webrtc-drift-results-$$` statt in das
  lokale `test-results/` (vermeidet Rechte-Konflikte mit Docker-
  Compose-Ausgaben).
- `spec/telemetry-model.md` §3.5.2 Einleitungstext und §3.5.3
  Punkt 1 präzisiert: Muss-Felder sind Pflichtbedingung **pro
  Engine** für die jeweilige Aggregat-Metrik (nicht „über alle drei
  Browser stabil"); fehlt ein Muss-Feld, bleibt die Metrik in dieser
  Engine leer (`unknown`-Surrogat ist Cardinality-Risiko).
- `packages/player-sdk/README.md` Browser-Support-Matrix Zeile
  Firefox 120+ präzisiert: `RTCStatsType.transport` ist nicht in
  allen Playwright-Firefox-Builds sichtbar; Adapter folgt §3.5.3
  und droppt das `dtls_state`-Aggregat anstatt ein `unknown`-
  Surrogat zu emittieren.
- Veraltete `docs/planning/in-progress/plan-0.9.0.md`-Referenzen in
  Code-Kommentaren und Markdown-Verweisen
  (`packages/stream-analyzer/src/internal/parsers/dash.ts`,
  `examples/srs/README.md`, `docs/planning/done/plan-0.9.0.md` DoD-
  Texte) nach dem Closeout-`git mv` auf
  `docs/planning/done/plan-0.9.0.md` korrigiert.

### Changed

- Versions-Bump auf `0.9.1` (Patch-Release): alle 5 `package.json`
  (root, `apps/dashboard`, `apps/analyzer-service`,
  `packages/player-sdk`, `packages/stream-analyzer`),
  `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts` `PLAYER_SDK_VERSION`,
  `packages/player-sdk/scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` plus Test-Fixtures und
  Contract-Fixtures, die Versions-Strings hartkodieren. Gleicher
  Bulk-Sed-Pfad wie `0.8.5` und `0.9.0` Closeout.

## [0.9.0] - 2026-05-07

> Drift-Smoke + SRS-Lab + DASH-Manifest-Analyse — Minor-Release mit
> drei thematisch getrennten Liefergegenständen aus
> [`done/plan-0.9.0.md`](docs/planning/done/plan-0.9.0.md). RAK-56
> (Browser-Drift-Smoke, automatisiert detektiert R-12) Soll;
> RAK-57 (SRS-Lab `examples/srs/`) Kann (MVP-36 eingelöst);
> RAK-58 (DASH-Manifest-Analyse) Muss (NF-12 eingelöst, MVP-37
> hochgestuft auf Muss); RAK-59 (DASH-CLI-Pfad) Kann. Lastenheft-
> Patch `1.1.11` ergänzt §13.11 mit RAK-56..RAK-59 und zieht MVP-37
> entsprechend NF-12 von „Kann" auf „Muss". Lieferstand der
> Tranchen 0–5 strukturiert nach Plan-Aktivierung+Lastenheft-Patch
> → Drift-Smoke → SRS-Lab → DASH-Analyse → Doku-Pflege → Closeout.

### Added (Tranche 1 — Browser-Drift-Smoke / R-12 / RAK-56)

- `tests/e2e/webrtc-stats-drift.spec.ts` (neu, Playwright):
  öffnet im Page-Context eine eigene `RTCPeerConnection` mit
  recvonly video+audio gegen den WHEP-Endpoint
  `http://localhost:8892/webrtc-test/whep`; nach
  `connectionState=connected` sammelt die Spec `pc.getStats()` und
  validiert gegen `spec/telemetry-model.md` §3.5.2 Muss-Felder pro
  `RTCStatsType`-Gruppe (peer-connection.connectionState,
  transport.dtlsState, candidate-pair.state, inbound-rtp.packetsLost
  +bytesReceived). §1.4-Allowlists für `connectionState`,
  `iceConnectionState` und `dtlsState` ebenfalls geprüft;
  unbekannter Enum → Fail. Soll-Felder werden als `[drift-soll]`
  geloggt, brechen den Smoke aber nicht. Spec ist via
  `MTRACE_WEBRTC_STATS_DRIFT=1` opt-in.
- `make smoke-webrtc-stats-drift` (Skript
  `scripts/smoke-webrtc-stats-drift.sh`): fährt das `mtrace-webrtc`-
  Lab via `docker compose -p mtrace-webrtc up -d --build` hoch,
  delegiert Endpoint-Probe an `smoke-webrtc-prep.sh` und ruft
  `pnpm exec playwright test` mit `--project=$browser` für jeden
  Browser aus `MTRACE_WEBRTC_DRIFT_BROWSERS` (Default
  `chromium,firefox`; WebKit/Safari opt-in über
  `chromium,firefox,webkit`). Opt-in (NICHT in `make gates`).
- `.github/workflows/webrtc-drift.yml` (neu, Nightly via
  `schedule: cron '30 3 * * *'` plus `workflow_dispatch`):
  installiert Playwright-Browser, ruft den Smoke; bei Failure wird
  optional ein Issue mit Workflow-Run-URL und Reaktions-Pfad
  erstellt (opt-in über `secrets.DRIFT_AUTO_ISSUE=1`).
- `docs/planning/in-progress/risks-backlog.md` R-12 von
  „release-blockierend ab nächstem Browser-Major-Bump" auf
  „automatisiert detektiert, Drift bricht den Drift-Smoke"
  angehoben; Manuell-Review-Pflicht entfällt.
- `docs/user/releasing.md` neue §2.4.1 mit Cron-Zeit, Browser-Set
  und Reaktions-Pfad bei Befund.

### Added (Tranche 2 — SRS-Lab `examples/srs/` / RAK-57 / MVP-36)

- `examples/srs/compose.yaml`: SRS-Container `ossrs/srs:5` plus
  FFmpeg-RTMP-Publisher; Project-Name `mtrace-srs`; Host-Ports
  1935 (RTMP) / 1985 (HTTP-API) / 8088 (HTTP-FLV) kollisionsfrei
  zu Core-Lab und mtrace-srt/dash/webrtc.
- `examples/srs/srs.conf`: minimale Konfiguration (RTMP-Listener,
  HTTP-API, HTTP-FLV-Egress mit `http_remux`); kein HLS/DASH/
  WebRTC-Output (das sind die anderen Beispiele).
- `examples/srs/ffmpeg-rtmp-loop.sh`: synthetischer Test-Stream
  (testsrc2 720p30 + 1 kHz Sine) per RTMP an
  `rtmp://srs:1935/live/srs-test`.
- `examples/srs/README.md` auf 7-Punkt-Standard analog
  `examples/srt/`/`examples/dash/`/`examples/webrtc/`.
- `make smoke-srs` (Skript `scripts/smoke-srs.sh`): opt-in Smoke
  (auto-up/down auf `mtrace-srs`-Project) mit drei Probes:
  HTTP-API antwortet 200, Stream `live/srs-test` registriert mit
  `publish.active=true`, HTTP-FLV-Egress liefert FLV-Magic-Header
  `FLV`. NICHT in `make gates`.
- `examples/README.md` Smoke-Tabelle und Beispiele-Tabelle um
  SRS-Eintrag erweitert; `docs/user/local-development.md` §2.7
  Port-Quickref um `mtrace-srs`-Zeile (1935/1985/8088).
- `docs/user/releasing.md` neue §2.4.2 (SRS-Lab-Boot-Verifikation).

### Added (Tranche 3 — DASH-Manifest-Analyse / RAK-58 / RAK-59 / NF-12)

- `@npm9912/stream-analyzer` versteht DASH-MPD-Eingaben zusätzlich
  zu HLS-Manifesten:
  - `internal/parsers/detect.ts`: Detector klassifiziert den
    Body-Anfang als HLS (`#EXTM3U`-Header), DASH (`<?xml`/`<MPD`-
    Header) oder unsupported; BOM-Strip; liefert auch erste Zeile
    für Diagnose-Findings.
  - `internal/parsers/dash.ts`: regex-basierter MPD-Parser ohne
    externe XML-Dependency; deckt VOD- und einfache Live-MPDs ab;
    AdaptationSet-Inheritance für mimeType/codecs;
    bandwidth-Pflicht laut MPEG-DASH §5.3.5 mit Error-Finding bei
    Verstoß. SegmentTemplate-Edge-Cases (`$Time$`-Variablen,
    `availabilityStartTime`-Drift) sind Out-of-Scope.
  - `analyze.ts` dispatcht über den Detector; Public-Funktion
    `analyzeHlsManifest` bleibt Backward-Kompat-Alias, neuer
    generischer Name `analyzeManifest`.
  - `internal/loader/fetch.ts`: `loadHlsManifest` →
    `loadManifest`; Content-Type-Allowlist um
    `application/dash+xml`/`application/xml`/`text/xml`;
    Accept-Header generalisiert.
- Neue Result-Types: `AnalyzerKind = "hls" | "dash"`,
  `PlaylistType` union erweitert um `"dash"`,
  `DashAnalysisResult` mit Diskriminator-Paar
  `analyzerKind:"dash"` + `playlistType:"dash"`,
  `DashManifestDetails` / `DashAdaptationSet` /
  `DashRepresentation` für die MPD-Hierarchie.
- Neuer Public-Code `manifest_not_supported` als Detector-
  Sammelfehler (HTML/JSON/leerer Body); `manifest_not_hls` bleibt
  HLS-Parser-spezifisch. Beide Codes mappen auf HTTP 422.
- `apps/api`-Adapter: `domain.AnalyzerKind`-Type plus
  `AnalyzerKindHLS`/`AnalyzerKindDASH`; `PlaylistTypeDash`-
  Konstante; `StreamAnalysisManifestNotSupported`-ErrorCode;
  `mapAnalyzerKind`-Helper im HTTP-Adapter; Driving-HTTP gibt
  `analyzerKind` aus dem Result statt der HLS-Konstante;
  Prometheus-Allowlist `mtrace_analyze_requests_total` um
  `manifest_not_supported` erweitert.
- `spec/contract-fixtures/analyzer/`: zwei neue Fixtures
  (`success-dash-vod.json` + `success-dash-live.json`) plus
  Sync-Pfad nach `apps/api/.../testdata/`; `make sync-contract-
  fixtures` kopiert 6 statt 4 Files; `make generated-drift-check`
  validiert die Kopien.
- `make smoke-cli` (Skript `scripts/smoke-cli.sh`) erweitert:
  zusätzlich zur HLS-Master-Probe und zu den IO-/SSRF-/Bin-Tests
  jetzt auch `m-trace check <vod.mpd>` →
  `analyzerKind=dash`/`playlistType=dash`, plus negativer Pfad
  HTML-Body → `manifest_not_supported`. Live verifiziert.
- `docs/user/stream-analyzer.md` §2.3 listet beide Codes mit
  klarer Trennung; §3 Scope-Tabelle erweitert um DASH-MPD-Pfade
  (VOD ✅, Live ✅, URL ✅; CMAF Folge-Plan); §9 CLI-Block listet
  DASH-Beispiele.
- `packages/stream-analyzer/README.md`: Diskriminator-Tabelle für
  `analyzerKind` + `playlistType`, gekürztes JSON-Beispiel mit
  Verweis auf Spec-Fixture; CLI-Dispatcher-Sektion mit Auto-
  Detection und der `manifest_not_hls`/`manifest_not_supported`-
  Trennung; Status-Block auf `0.9.0`.
- `docs/user/releasing.md` neue §2.4.3 (DASH-CLI-Probe).

### Added (Tranche 0)

- `0.9.0`-Plan aus `open/` aktiviert und nach Closeout unter
  `docs/planning/done/plan-0.9.0.md` archiviert (Status
  `⬜ → 🟡 → ✅`); Tranche 0 abgeschlossen mit Plan-Move,
  Lastenheft-Patch `1.1.11` und
  Toolchain-Bump-Check ohne Bump-Bedarf (Go 1.26 / golangci-lint
  v2.12.1-alpine / Node 22-trixie-slim / pnpm 10 sind seit
  `0.7.0`/`0.8.5` aktuell).
- Lastenheft-Patch `1.1.11` (`spec/lastenheft.md` Header + §13.11
  + §12.3-Note für MVP-37; Patch-Log §4a.14 in
  `docs/planning/done/plan-0.1.0.md`): RAK-56 (Drift-Smoke, Soll),
  RAK-57 (SRS-Lab, Kann), RAK-58 (DASH-Manifest-Analyse, Muss),
  RAK-59 (DASH-CLI, Kann); MVP-37 von „Kann" auf „Muss"
  hochgezogen entsprechend NF-12.
- `docs/planning/in-progress/roadmap.md` §2 neue Schritte 42
  (Lastenheft-Patch ✅) und 43 (`0.9.0` ausliefern, jetzt ✅).

### Changed

- Versions-Bump auf `0.9.0` (Tranche 5 Closeout): alle 5
  `package.json` (root, `apps/dashboard`, `apps/analyzer-service`,
  `packages/player-sdk`, `packages/stream-analyzer`),
  `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts` `PLAYER_SDK_VERSION`,
  `packages/player-sdk/scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` plus alle Test-
  Fixtures und Contract-Fixtures, die Versions-Strings
  hartkodieren. Gleicher Bulk-Sed-Pfad wie `0.8.5` Closeout.

## [0.8.5] - 2026-05-07

> Erstmaliger **Patch-Release** im m-trace-Repo (Quality-Gates Wave 1):
> Security-Gates (`vuln-check`/`audit-ts`/`image-scan`/`security-gates`)
> und Generated-Artifact-Drift-Gate; Migrations-Konsolidierung als
> rolling V1; Image-Hardening (Trixie-slim + dev-dep-Snip + npm-
> Removal); OpenTelemetry-Stack-Bump als Vuln-Fix-Folge. Patch-Release-
> Konvention (`0.X.Y`) in `docs/user/releasing.md` §3.1 verankert. Keine
> User-Surface-Änderung, kein Lastenheft-Patch, keine RAK-
> Verifikationsmatrix (Plan-DoD-Items reichen). Lieferstand der
> Tranchen 0–3 in [`docs/planning/done/plan-0.8.5.md`](docs/planning/done/plan-0.8.5.md)
> archiviert.

### Added

- `0.8.5`-Plan unter `docs/planning/in-progress/plan-0.8.5.md` aus
  `open/` aktiviert (Status `🟡 in Arbeit`); Tranche 0 abgeschlossen
  (Plan-Move + Container-Scanner-Wahl Trivy + Toolchain-Check
  ohne Bump-Bedarf, weil Go 1.26 / golangci-lint v2.12.1 / Node 22
  LTS aus `0.7.0` Tranche 0 aktuell sind).
- Security-Gates Wave 1 (plan-0.8.5 Tranche 1):
  - `make vuln-check`: govulncheck (gepinnt auf `v1.1.4`) gegen
    `apps/api/...`-Go-Dependencies in einem `golang:1.26`-Container.
  - `make audit-ts`: `pnpm audit --audit-level high` über den
    gesamten pnpm-Workspace (`apps/dashboard`,
    `apps/analyzer-service`, `packages/*`); Schwelle = `high`,
    `moderate`/`low` werden berichtet, brechen aber den Lauf nicht.
    Pendant zu `vuln-check` für die TypeScript-Seite — bewusst
    Bestandteil derselben Wave, weil die Go-/Image-Gates allein
    den npm-Pfad nicht abdecken (`extra-gates.md` §3.1 nannte
    ursprünglich nur Go + Container, der TS-Gate ist die Lücke).
  - `make image-scan`: Trivy (gepinnt auf `aquasec/trivy:0.59.1`)
    scannt `apps/api`-Runtime-Image plus die Dashboard- und
    Analyzer-Service-Container; Policy `CRITICAL,HIGH` mit
    Exit-Code 1; Cache unter `.security/.trivy-cache`.
  - Wrapper
    `make security-gates: vuln-check audit-ts image-scan`.
  - `.security/vulnignore.yaml` als Ignorierregel-Pfad mit
    `expires`-Pflicht und Begründungs-Spalte; initial leer.
  - `.github/workflows/build.yml` um zweiten Job `security`
    erweitert (parallel zum bestehenden `build`-Job, PR-blockierend,
    Trivy-Cache als Workflow-Artefakt mit 7 Tagen Retention).
  - Image-Hardening als Tranche-1-Closeout (getriggert durch den
    ersten CI-Lauf): Dashboard- und Analyzer-Service-Dockerfile
    beide auf `node:22-trixie-slim`, dev-deps werden via
    `pnpm deploy --prod --legacy /deploy` ausgeschnitten,
    Runtime-Stages entfernen das gebündelte npm-Tooling
    (eliminiert die `picomatch@4.0.3`-CVE-Kopie aus npm).
    Analyzer-Service vorher `node:22-alpine`; Wechsel zu glibc,
    weil musl bei multi-threaded Workloads (libuv-Worker-Pool,
    V8-GC/JIT) pessimisiert. Drei verbleibende Trixie-OS-CVEs
    ohne Upstream-Fix (`CVE-2025-69720`, `CVE-2026-29111`,
    `CVE-2026-4878`) per `.security/vulnignore.yaml` mit 90-
    Tage-`expires` dokumentiert; Generator
    `scripts/render-trivyignore.sh` rendert `.trivyignore`
    daraus und bricht ab, sobald ein `expires` überschritten
    ist (Wartungsregel). `pnpm.overrides`-Block in Root-
    `package.json` hebt `picomatch` workspace-weit auf
    `^4.0.4`. Folge-Risiko R-13 in `risks-backlog.md`
    (Trixie-Point-Release-Re-Review oder Distroless-Wechsel
    vor 1.0).
- Generated-Artifact-Drift-Gate Wave 1 (plan-0.8.5 Tranche 2):
  - `make generated-drift-check`: ruft `make schema-generate`,
    `make sync-contract-fixtures` und das Player-SDK-Public-API-
    Snapshot-Skript auf und vergleicht anschließend die generierten
    Pfade per `git diff --exit-code HEAD --` mit dem committeten
    Stand. Drift-Befund nennt den konkreten Regenerier-Befehl pro
    Pfad. In `make gates` zwischen `schema-validate` und
    `sdk-pack-smoke` aufgenommen.
  - Geprüfte Artefakte: `apps/api/internal/storage/migrations/V1__m_trace.sql`,
    vier Contract-Fixtures (`testdata/contract-success-master.json`,
    `contract-error-fetch-blocked.json`, `mediamtx-srtconns-list.json`,
    `srt-health-detail.json`),
    `packages/player-sdk/scripts/public-api.snapshot.txt`.
- Patch-Release-Konvention `0.X.Y` (plan-0.8.5 Tranche 3, §0.6
  des Plans): `docs/user/releasing.md` §3.1 dokumentiert die drei
  Release-Typen (Patch/Minor/Major) als Tabelle inklusive
  Lastenheft-Pflicht und RAK-Verifikationsmatrix-Pflicht. Patch-
  Release umfasst alle Versions-Bump-Stellen, die ein Minor-Bump
  auch berührt — sonst entsteht Drift zwischen SDK-Bundle, API-
  Service-Version und CI-Smokes. `0.8.5` ist der erste Patch-
  Release im Repo.

### Changed

- Migrations-Konsolidierung: rolling V1 als Single-Source-of-Truth
  (plan-0.8.5 Tranche 2 Sub-2a). Die historischen V2..V5-Migrationen
  (`V2__project_session_pk.sql`, `V3__session_boundaries.sql`,
  `V4__session_end_source.sql`, `V5__srt_health_samples.sql`) wurden
  in der aus `schema.yaml` regenerierten V1 zusammengefasst und
  gelöscht; legitim, weil m-trace noch keinen Production-State
  erreicht hat. Vorbedingung war, dass der Composite-FK
  `stream_session_boundaries → stream_sessions(project_id, session_id)
  ON DELETE CASCADE` aus V3 als `constraints[]`-Eintrag mit
  `type: foreign_key` in `schema.yaml` ergänzt werden musste —
  vorher modellierte `schema.yaml` nur Single-Column-FKs und hätte
  den FK beim Konsolidieren verloren. Apply-Runner unverändert:
  ignoriert applied-Versionen ohne File, deshalb bleiben Dev-DBs
  mit V2..V5-applied funktional. ADR-0002 §8.2 dokumentiert die
  rolling-V1-Strategie und das Pre-Production-Privileg. Auslöser
  war ein Drift-Befund während der Tranche-2-Implementierung —
  schema-generate erzeugte 50 Zeilen Diff, weil V1 hinter
  `schema.yaml` zurückgefallen war.
- OpenTelemetry-Stack in `apps/api/go.mod` von `v1.32.0`/`v0.57.0`/
  `v0.8.0` auf `v1.43.0`/`v0.68.0`/`v0.19.0` angehoben — direkter
  Auslöser war `make vuln-check` (`GO-2026-4394`: PATH-Hijacking in
  `go.opentelemetry.io/otel/sdk@v1.32.0`, fixed in `v1.40.0`). Da
  die contrib-/exporter-Pakete denselben Release-Schwarm wie der
  Core nutzen, wurde der gesamte Stack koordiniert auf den aktuell
  jüngsten Stable-Stand bezahlt. Folge-Anpassung: `semconv`-Import
  in `apps/api/adapters/driven/telemetry/otel.go` von `v1.26.0`
  auf `v1.40.0` umgestellt, damit der Schema-URL-Merge im SDK-
  Default-Resource nicht mehr in einen `conflicting Schema URL`-
  Fehler läuft. `make api-test` und `make gates` grün; keine
  Anpassungen am restlichen Telemetrie-Code nötig.
- Versions-Bump auf `0.8.5` (plan-0.8.5 Tranche 3): alle 5
  `package.json` (root, `apps/dashboard`, `apps/analyzer-service`,
  `packages/player-sdk`, `packages/stream-analyzer`),
  `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts` `PLAYER_SDK_VERSION`,
  `packages/player-sdk/scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` plus alle Test-Fixtures
  und Contract-Fixtures, die Versions-Strings hartkodieren. Folge-
  Backlog-Item: Tests sollten die Version aus `package.json` lesen
  statt hartzukodieren (separater Plan).

### Removed

- `apps/api/internal/storage/migrations/V2__project_session_pk.sql`,
  `V3__session_boundaries.sql`, `V4__session_end_source.sql`,
  `V5__srt_health_samples.sql` — in der rolling V1 konsolidiert
  (s. Changed-Block oben).

## [0.8.0] - 2026-05-06

> Player-SDK-WebRTC-Adapter und produktive WebRTC-Telemetrie:
> `attachWebRtc(video, options, tracker)` als additiver Pfad neben
> `attachHlsJs`; reservierter `webrtc.*`-Meta-Namespace mit harter
> API-Validation; sechs `mtrace_webrtc_*`-Counter mit serverseitiger
> Delta-Berechnung über `(project_id, session_id, peer_connection_run_id,
> metric)` und Sample-ID-Idempotenz; `scripts/smoke-observability.sh`
> spiegelt §3.1-Forbidden und §3.2-Allowlist; R-12 release-blockierend
> ab nächstem Browser-Major-Bump; Browser-Support-Matrix Chromium
> 120+/Firefox 120+ Required, Safari 17+ Best-effort. RAK-51..RAK-55
> erfüllt. Lieferstand der Tranchen 0–5 strukturiert nach
> Plan-Aktivierung+Lastenheft-Patch `1.1.10` → Public-API-Vertrag
> → WHEP-Adapter-Implementation → produktive WebRTC-Telemetrie →
> Compat-Tests+Browser-Matrix → Closeout. Post-`0.7.0`-Sammelblock
> mitverarbeitet.

### Added (Tranche 4 — Compat-Tests + Browser-Support-Matrix)

- `packages/player-sdk/scripts/pack-smoke.mjs` validiert
  `attachWebRtc` in allen drei Bundle-Entries (ESM, CJS, IIFE) plus
  TypeScript-Type-Deklarationen (`WebRtcAdapter`,
  `WebRtcAdapterOptions`) in `dist/index.d.ts`.
- `packages/player-sdk/scripts/performance-smoke.mjs` importiert
  `attachWebRtc` aus dem produktiven Bundle und prüft die
  Funktion-Surface; Bundle-Size-Budget (< 30 KiB gzip ESM) bleibt
  unverändert und greift jetzt inklusive WebRTC-Adapter.
- `packages/player-sdk/README.md` §Performance and Browser Support
  um eine dedizierte WebRTC-Adapter-Browser-Matrix erweitert
  (Chromium 120+ Required, Firefox 120+ Required, Safari 17+ Best-
  effort) plus CI-Policy-Block (Release-blockierend vs. opt-in).
- `tests/e2e/dashboard-demo-webrtc.spec.ts` (neu, Playwright,
  RAK-55 Kann): rendert `/demo-webrtc?autostart=1` und verifiziert
  über `GET /api/stream-sessions/{id}`, dass mindestens ein Event
  mit reservierten `webrtc.*`-Meta-Keys in der Session-Timeline
  ankommt. Default prüft den Error-Pfad (`whep_signaling_failed`);
  `MTRACE_WEBRTC_LAB=1` flippt auf `playback_started`.

### Added (Tranche 3 — produktive WebRTC-Telemetrie)

- `webrtc.*`-Meta-Namespace ist ab `0.8.0` produktiv: `spec/
  telemetry-model.md` §1.4 listet die Allowlist (peer_connection_run_id,
  sample_id, connection_state/ice_state/dtls_state,
  packets_lost/bytes_received/bytes_sent, error_code, error_detail);
  `contracts/event-schema.json` (`reserved_meta_keys` +
  `reserved_meta_namespace_webrtc`) und `contracts/sdk-compat.json`
  (`reserved_meta_namespaces`) spiegeln den Vertrag.
- `apps/api/hexagon/application/event_meta_validation.go` weist
  unbekannte `webrtc.*`-Keys, falsche Typen, ungültige Enum-Werte,
  negative Counter, Pattern-Verletzungen und Per-Identifier-Felder
  (`webrtc.track_id`, `webrtc.ssrc`, …) hart mit HTTP 422 ab — keine
  `mtrace_webrtc_*`-Metrik wird in diesem Pfad erzeugt.
- `apps/api/adapters/driven/metrics/webrtc_metrics.go`: drei State-
  CounterVec (`mtrace_webrtc_{connection,ice,dtls}_state_total`) plus
  drei label-freie Delta-Counter (`packets_lost_total`,
  `bytes_received_total`, `bytes_sent_total`). Server-side Sample-
  State (`(project_id, session_id, peer_connection_run_id, metric)`-
  Map) berechnet nichtnegative Deltas; Sample-ID-Idempotenz für
  Duplicates; Reconnect mit neuer Run-ID startet eigene Baseline.
- `packages/player-sdk/src/adapters/webrtc/sampling.ts`:
  `collectAggregate()` extrahiert §3.5.2-Muss-Felder aus
  `RTCStatsReport`; fehlende Muss-Felder lassen das Sample fallen
  (kein unknown-Surrogat). `startSampling()` registriert ein
  setInterval-Tick gegen `pc.getStats()` und sendet
  `metrics_sampled`-Events.
- `attachWebRtc(...)` neue Option `samplingIntervalMs` (Default
  1000 ms; 0 deaktiviert). `peer_connection_run_id` aus
  `crypto.randomUUID()` wird in `playback_started`,
  `playback_error` und allen `metrics_sampled`-Events mitgeliefert.
- `scripts/smoke-observability.sh` erweitert: WebRTC-Forbidden-
  Identifier in der Forbidden-Liste plus Self-Tests; neue Allowlist-
  Sektion validiert State-Counter (nur State-Label) und Byte-/Loss-
  Counter (label-frei) gegen `mtrace_webrtc_*`-Series.
- `docs/planning/in-progress/risks-backlog.md` R-12 von „Triggerschwelle
  nicht ausgelöst" auf „release-blockierend ab nächstem Browser-
  Major-Bump" angehoben — produktive WebRTC-Telemetrie ist live;
  Drift-Review-Pflicht vor jedem Release-Tag.

### Added

- Lastenheft-Patch `1.1.10` (`spec/lastenheft.md` Header + §13.10 +
  §13.9-Notiz; Patch-Log in `done/plan-0.1.0.md` §4a.13): RAK-51
  von „Kann" auf „Muss" hochgezogen, RAK-52..RAK-55 für
  Public-API + hls.js-Trennung, produktive WebRTC-Telemetrie auf
  `spec/telemetry-model.md` §3.2-Allowlist, `getStats()`-Sammlung
  mit Schema-Drift-Strategie und opt-in Browser-E2E. Vorgänger-
  Gate für `0.8.0` Tranche 0 erfüllt.
- `0.8.0`-Plan unter `docs/planning/in-progress/plan-0.8.0.md` aus
  `open/` aktiviert (Status `🟡 in Arbeit`); Tranche 0 abgeschlossen
  (Plan-Move + Lastenheft-Patch + Toolchain-Check ohne Bump-Bedarf
  weil `0.7.0` Tranche 0 die Toolchain frisch gehoben hat).

## [0.7.0] - 2026-05-06

> WebRTC-Lab-Erweiterung: lokal startbares WHIP-/WHEP-Compose
> ([`examples/webrtc/`](examples/webrtc/), Project `mtrace-webrtc`)
> mit FFmpeg-RTSP-Publisher; opt-in `make smoke-webrtc-prep`
> (endpoint-/compose-only); WebRTC-Telemetrie-Vorbereitung in
> `spec/telemetry-model.md` §3.5 (bounded Allowlist, `getStats()`-
> Subset, Schema-Drift-Strategie); R-12 als Spec-/Adapter-Review-
> Gate. RAK-47..RAK-50 erfüllt; RAK-51 (Player-SDK-WebRTC-Adapter)
> deferred. Lieferstand der Tranchen 0–5 strukturiert nach
> Plan-Aktivierung+Toolchain → Lab-Compose → README → Smoke →
> Telemetrie-Bewertung → Closeout. Post-`0.6.0`-Befunde sind in
> diesem Block mit verarbeitet (siehe „Changed"/„Fixed").

### Added

- `spec/telemetry-model.md` §3.5 (Future-Telemetry-Notiz für WebRTC,
  RAK-49, plan-0.7.0 Tranche 4): `getStats()`-Subset pro
  `RTCStatsType`-Gruppe mit Muss-/Soll-Feldern, Schema-Drift-Strategie
  zwischen Chromium/Firefox/Safari, Fallback-Verhalten bei fehlenden
  Soll-Feldern, negative Cardinality-Prüfung (kein produktiver
  `mtrace_webrtc_*`-Counter im `0.7.0`-Scope), Out-of-Scope-Klauseln.
  §3.1 um WebRTC-Forbidden-Identifier (`peer_connection_id`, `ssrc`,
  …, Codec-/User-Agent-Strings) und §3.2 um drei W3C-Enum-Labels
  (`connection_state`, `ice_state`, `dtls_state`) ergänzt — Spec-
  Vorbereitung, Smoke-Spiegelung ist Folge-DoD.
- `docs/planning/in-progress/risks-backlog.md` R-12 (Spec-/Adapter-Review-Gate
  für WebRTC-`getStats()`-Schema-Drift; nicht R-11, der ist seit
  `0.6.0`-Closeout für SRT-Health-Cursor-Pagination vergeben). Trigger-
  schwelle: Browser-Major-Bump mit Schema-Änderung ODER Beginn
  produktiver WebRTC-Telemetrie-Implementierung.
- `make smoke-webrtc-prep` (RAK-48, plan-0.7.0 Tranche 3) startet/
  stoppt `mtrace-webrtc` und prüft endpoint-/compose-only fünf
  Stufen: MediaMTX-API ready, Stream-Pfad `webrtc-test` ready=true,
  `OPTIONS …/whep|whip → 204`, Negativ-Probe für unbekannten Pfad
  liefert ≠ 204. Kein Browser, kein Playback, kein `getStats()` —
  diese Aspekte deckt nur der manuelle Browser-Handcheck. Opt-in,
  nicht in `make gates`.
- WebRTC-Lab-Compose `examples/webrtc/compose.yaml` (Project
  `mtrace-webrtc`, RAK-47): MediaMTX `bluenviron/mediamtx:1` mit
  WHIP-/WHEP-Listener auf Host-Port `8892`, ICE-UDP `8189`,
  Control-API `9999` (kollisionsfrei zu Core-Lab/`mtrace-srt`/
  `mtrace-dash`). FFmpeg-RTSP-Publisher `ffmpeg-rtsp-loop.sh`
  (synthetischer H264+Opus-Stream); MediaMTX re-published auf
  WHEP-Pfad `/webrtc-test/whep`. `examples/webrtc/README.md` auf
  7-Punkt-Standard mit Browser-Handcheck (RAK-50, manuell) plus
  live verifiziertem Endpoint-Statussatz.
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
  `session_link.status="detached"` ([R-6](docs/planning/in-progress/risks-backlog.md)
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
