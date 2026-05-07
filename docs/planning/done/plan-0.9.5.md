# Implementation Plan — `0.9.5` (Quality-Gates Wave 2: Benchmarks + Fuzzing + Mutation)

> **Status**: ✅ released (Tag `v0.9.5` am 2026-05-07; Plan-File am
> selben Tag von `docs/planning/in-progress/` nach
> `docs/planning/done/` verschoben). Tranchen 0, 2, 3, 4, 5 ✅.
> Tranche 1 ist als Lieferung **ausgeliefert**, läuft aber mit
> einer 🟡-markierten Beobachtungsphase weiter (Nightly-Workflow
> `benchmark-observation.yml` `continue-on-error: true`); ein
> Folge-Commit nach N=3..5 grünen Läufen entfernt die `continue-
> on-error`-Marker und nimmt `make benchmark-smoke` in
> `make gates` auf — siehe Plan-DoD §2-6 und Roadmap §1.2 Folge-
> Punkt „Benchmark-Smoke PR-Blockierung". Vorgänger `v0.9.1`
> (Wartungs-Patch nach `v0.9.0`) bleibt im Release-Verlauf;
> `0.9.5` ist Patch-Release ohne User-Surface-Änderung —
> Inhalt: Benchmark-Smoke, Nightly-`benchstat`-Regressionen,
> selektives Fuzzing + TS-Property-Tests, Mutation-Testing als
> Nightly-Report.
>
> **Release-Typ**: Patch-Release nach `0.9.0`/`0.9.1` (Konvention aus
> [`plan-0.8.5.md`](../done/plan-0.8.5.md) §0.6 /
> [`docs/user/releasing.md`](../../user/releasing.md) §3.1).
>
> **Lastenheft-Status**: kein Lastenheft-Patch nötig (Quality-Gates,
> keine User-Surface).
>
> **Bezug**: [`extra-gates.md`](../in-progress/extra-gates.md) §3.2 (Benchmark-
> Smoke), §3.3 (Nightly-`benchstat`-Regressionen), §3.5 (Selektives
> Fuzzing / Property Tests), §3.6 (Mutation Testing) — die vier
> statistisch-langlaufenden Wave-2-Gates aus dem Master-Backlog;
> [`plan-0.8.5.md`](../done/plan-0.8.5.md) Wave 1 (Security + Generated-
> Drift) als Vorlage für die Patch-Release-Mechanik;
> [`plan-0.4.0.md`](../done/plan-0.4.0.md) Tranche 1 (Cursor
> v2 — bevorzugter Fuzz-Kandidat aus §3.5);
> [`packages/stream-analyzer/`](../../../packages/stream-analyzer/)
> (Manifest-Parser — bevorzugter Property-Test-Kandidat);
> [`docs/perf/budgets.md`](../../perf/budgets.md) (initiale
> Performance-Budgets pro Modul, Tranche 0).
>
> **Nachfolger**: offen — kein `plan-0.10.0.md` vorbereitet.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

Scope-Grenze: dieser Plan liefert die vier statistisch-/langlaufenden
Quality-Gates aus `extra-gates.md` (Wave 2). Im Gegensatz zu Wave 1
(`plan-0.8.5.md`) sind diese Gates **nicht** alle PR-blockierend —
einige laufen Nightly oder Release-only. Die Trennung folgt der
Benchmarking-Policy aus `extra-gates.md` §4.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- **`0.9.0` ist released** (Tag `v0.9.0`). Der Drift-Smoke aus
  [`plan-0.9.0.md`](../done/plan-0.9.0.md) Tranche 1 ist live;
  SRS-Lab und DASH-Analyse sind ausgeliefert. `0.9.5` ist das
  Quality-Hardening-Patch **nach** dem Feature-Release.
- `0.8.5` ist released (Tag `v0.8.5`); Wave-1-Gates (Security +
  Generated-Drift) sind aktiv und PR-blockierend.
- `extra-gates.md` §6 (offene Entscheidungen) ist bei Tranche 0
  geklärt: Baseline-Pfad (Git-Repo, Actions-Artefakt oder Release-
  Asset), initiale Performance-Budgets pro Modul, Quarantäne-
  Policy für laute Benchmarks.

### 0.2 Out-of-Scope-Klauseln (durchgängig)

- Kein `govulncheck`/Container-Scan und kein Generated-Drift-Gate —
  diese sind Wave 1 (`plan-0.8.5.md`).
- Keine Wire-Vertrags-Änderung; reine CI-/Tooling-Lieferung.
- Keine produktiven Telemetrie-Pfade (kein neuer
  `mtrace_*`-Counter); Benchmark-Ergebnisse fließen über
  CI-Artefakte, nicht über die m-trace-API.
- Kein eigenständiger Lastenheft-Patch.

### 0.3 Sequenzierung und harte Gates

1. Tranche 0 (Plan-Aktivierung + Baseline-Entscheidungen) ist
   Pflicht.
2. Tranche 1 (Benchmark-Smoke) ist Voraussetzung für Tranche 2
   (Nightly-`benchstat`-Regressionen) — `benchstat` braucht die
   Benchmarks als Quelle. Reihenfolge ist erzwungen.
3. Tranche 3 (Fuzzing) und Tranche 4 (Mutation Testing) sind
   **unabhängig** voneinander und von Tranche 1+2.
4. Tranche 5 (Closeout) erst nach allen vier inhaltlichen Tranchen.

### 0.4 Implementierungsleitplanken

**Benchmark-Smoke (Tranche 1)**: Bevorzugte Form ist eine
go-Benchmark-Suite (`go test -bench=. -benchmem`) für die in
`extra-gates.md` §3.2 gelisteten API-Kandidaten plus eine
TypeScript-Benchmark-Suite (vitest-bench oder Tinybench) für die
Stream-Analyzer-Kandidaten. PR-Budgets sind absolute Schwellen,
keine Vergleiche. Erste Beobachtungsläufe sind nicht-blockierend
(N=3-5 grüne Runs), bevor das Budget PR-blockierend wird.

**Nightly-`benchstat` (Tranche 2)**: Bevorzugte Form ist ein
neuer GitHub-Actions-Workflow `.github/workflows/benchmark.yml`
(`on: schedule: cron`). Baseline ist ein Git-Branch
`benchmark-baseline` oder ein dedizierter Release-Asset (in
Tranche 0 entschieden). Bei Regression > 15 % auf einem statistisch
signifikanten Benchmark wird ein Issue auto-erstellt.

**Fuzzing (Tranche 3)**: Bevorzugte Form ist Go-Fuzzing
(`go test -fuzz=...`) für die in `extra-gates.md` §3.5 gelisteten
Go-Kandidaten (Cursor, HTTP-Validation, Event-Metadaten,
SRT-Health-Mapping); für TypeScript Property-Tests via
`fast-check`. Fuzzing läuft **nicht** PR-blockierend (zu lang); ein
opt-in `make fuzz-check`-Target läuft mit kurzem `-fuzztime` (z. B.
`30s`), Nightly mit längerem Budget.

**Mutation Testing (Tranche 4)**: Bevorzugte Form ist
`go-mutesting` für ein bis zwei kritische Module (Vorschlag:
`apps/api/hexagon/application/event_meta_validation.go` und
`packages/player-sdk/src/adapters/webrtc/sampling.ts` über
StrykerJS). Initial **nicht-blockierend**; Output als Nightly-
Report-Artefakt. PR-Blockierung erst, wenn die Reports zeigen, dass
die Mutation-Score stabil > 80 % ist.

### 0.5 Test-Fixture-Versions-Drift bei Patch-Release

Identisch zu `plan-0.8.5.md` §0.5: `xargs sed -i` über
Test-Fixture-Files. Folge-Backlog-Item bleibt offen (Tests-aus-
package.json-lesen).

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung + Baseline-Entscheidungen aus `extra-gates.md` §6 (Baseline-Pfad, initiale Budgets, Quarantäne-Policy) | ✅ |
| 1 | Benchmark-Smoke für API + Stream-Analyzer mit konservativen Budgets, opt-in PR-blockierend nach N grünen Beobachtungsläufen | 🟡 |
| 2 | Nightly-`benchstat`-Regressionen mit Baseline-Vergleich; CI-Workflow `benchmark.yml` (cron) | ✅ |
| 3 | Selektives Fuzzing (Go) + Property Tests (TypeScript) für Cursor/Parser/URL-Klassifizierung | ✅ |
| 4 | Mutation Testing als nicht-blockierender Nightly-Report für ein bis zwei kritische Module | ✅ |
| 5 | Release-Doku, Versions-Bump 0.9.0 → 0.9.5, Plan nach `done/`, Tag `v0.9.5` | ✅ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Baseline-Entscheidungen

Bezug: `extra-gates.md` §6 (offene Entscheidungen).

DoD:

- [x] Plan-Skelett von `docs/planning/open/plan-0.9.5.md` nach
  `docs/planning/in-progress/plan-0.9.5.md` verschoben (Status
  `⬜ → 🟡 → ✅` für Tranche 0; Cross-Refs in `roadmap.md` §1.2/§3
  und `risks-backlog.md`-Header nachgezogen) (Tranche-0-Commit).
- [x] **Baseline-Pfad entschieden**: **Git-Branch
  `benchmark-baseline`**. Begründungen:
  1. Deterministische Historie via `git log` — `benchstat`-Vergleiche
     sind über jeden Commit-SHA reproduzierbar.
  2. Kein Retention-Limit (GitHub-Actions-Artefakte verfallen
     standardmäßig nach 90 Tagen; Release-Assets sind an Tags
     gebunden und nicht rolling-fortschreibbar ohne Tag-Move).
  3. Identisch zur Default-Empfehlung in §0.4 dieses Plans und in
     [`extra-gates.md`](../in-progress/extra-gates.md) §3.3.
  Mechanik (für Tranche 2 Implementation): Nightly-Workflow
  pusht den `benchstat`-tauglichen Output nach `benchmark-baseline`
  als orphan-branch-File `benchmarks/<module>.txt`; PR-Reviews
  vergleichen mit `benchstat baseline.txt new.txt` (Tranche-0-Commit).
- [x] **Initiale Performance-Budgets pro Modul dokumentiert** in
  [`docs/perf/budgets.md`](../../perf/budgets.md) — eigene
  Single-Source-of-Truth außerhalb des Plans, weil die Tabelle
  über mehrere Tranchen wächst (Tranche 1 Schärfung nach realen
  Messungen, Tranche 2 Drift-Schwellen). Initial-Werte sind
  Architektur-basiert großzügig (Plan-DoD: „Faktor 2-3 über
  aktueller Messung"), explizit als „Tranche-0-Stand,
  noch nicht mess-basiert" markiert; Tranche-1-Beobachtungsphase
  schärft sie nach N=3-5 grünen Läufen pro Hot-Path
  (Tranche-0-Commit).
- [x] **Quarantäne-Policy für laute Benchmarks**: ein Benchmark
  darf maximal **30 Tage** in Quarantäne (markiert `t.Skip` mit
  Begründung-Kommentar plus Backlog-Item in
  `docs/planning/in-progress/risks-backlog.md`). Danach: entweder
  Drift-Fix landed, oder der Benchmark wird aus dem Smoke
  entfernt. Quarantäne-Beginn und Trigger werden im jeweiligen
  Backlog-Item dokumentiert; reine „flaky"-Vermutungen ohne
  Drift-Hinweis brauchen einen Skip-Grund. Verlängerung über 30
  Tage hinaus ist eine Plan-DoD-Item-Änderung im jeweiligen
  Folge-Plan, kein stiller Re-Skip (Tranche-0-Commit).

---

## 2. Tranche 1 — Benchmark-Smoke für API + Stream-Analyzer

Bezug: `extra-gates.md` §3.2.

Ziel: PR-blockierende Budget-Smokes für die kritischen Hot-Paths.
Budgets sind absolute Schwellen, nicht Diffs.

DoD:

- [x] Go-Benchmark-Suite in `apps/api/...` für vier Hot-Paths aus
  `docs/perf/budgets.md` §3:
  `BenchmarkRegisterPlaybackEventBatch_Typical` und `_MaxBatch` in
  `hexagon/application/register_playback_event_batch_bench_test.go`,
  `BenchmarkEventRepository_AppendBatch_100` in
  `adapters/driven/persistence/sqlite/event_repository_bench_test.go`,
  `BenchmarkSessionsService_ListSessions_DefaultPage` in
  `hexagon/application/sessions_service_bench_test.go`,
  `BenchmarkCursorEncodeDecode_Pair` in
  `adapters/driving/http/cursor_bench_internal_test.go`. Test-
  Helper-Stubs werden aus den jeweiligen `_test.go`-Files
  wiederverwendet (Tranche-1a-Commit `afbcafd`).
- [x] TypeScript-Benchmark-Suite in
  `packages/stream-analyzer/benchmarks/analyzer.bench.ts` (eingebaute
  Vitest-Bench-API, separate Config `vitest.bench.config.ts`) für
  sieben Hot-Paths aus `docs/perf/budgets.md` §4: HLS Master klein
  (5 Variants + 1 Rendition), HLS Master groß (50 Variants + 20
  Renditions), HLS Media (1.000 Segmente), DASH-MPD VOD/Live,
  Detector über 256-KiB-Body, SSRF-URL-Klassifizierung 100 Calls.
  Synthetische Fixtures werden im Bench-File generiert
  (deterministisch, repo-lokal); `master.m3u8` wird aus dem
  bestehenden Fixtures-Pfad gelesen (Tranche-1b-Commit).
- [x] `make api-benchmark-smoke` und `make analyzer-benchmark-smoke`
  im Root-`Makefile`; Wrapper `make benchmark-smoke` (Plan-DoD §2-3).
  Beide drucken zuerst Runner-Info (`scripts/print-bench-runner-
  info.sh`), dann den Bench-Lauf. `make api-benchmark-smoke`
  delegiert an `apps/api/Makefile::benchmark-smoke`, das einen
  golang:1.26-Container startet und `go test -bench=. -benchmem
  -benchtime 1s ./hexagon/... ./adapters/...` ausführt.
  `make analyzer-benchmark-smoke` ruft `pnpm run bench` über das
  Workspace-Filter (`@npm9912/stream-analyzer`) auf
  (Tranche-1b-Commit).
- [x] Output enthält Laufzeit, Allokationen (`-benchmem`) und
  Durchsatz lesbar: Go-Benchmarks drucken `ns/op`, `B/op`,
  `allocs/op`; Vitest-Bench druckt `hz`, `mean`, `p75`, `p99`,
  `rme`. Budget-Verletzung mit eindeutiger Ist/Soll-Fehlermeldung
  über `scripts/check-bench-budgets.mjs` (Tranche-1c-Commit) — der
  Validator parst die Text-Tabellen beider Bench-Backends (`--kind
  ts` für Vitest-Bench-stdout, `--kind go` für Go-Bench-stdout) und
  vergleicht `mean` und `p99` (TS) bzw. `ns/op→ms` (Go) gegen die
  Budget-Tabelle (Single-Source: `docs/perf/budgets.md` §3 / §4,
  parallel als JS-Object im Validator-Skript). Output-Form:
  `[bench-budget] FAIL <name>: ist=<X> ms soll=<Y> ms (over by
  <Z>%)`. Vitest-JSON-Reporter mit `--outputFile` ist nicht nutzbar
  — vitest 4.1 wirft dort einen internen Server-Setup-Fehler;
  Text-Parsing ist resilient und versionsstabil.
- [x] Fixtures sind stabil und versioniert (keine Netzwerk-
  Abhängigkeit). Go-Benchmarks nutzen In-Memory-Stubs aus
  `_test.go`-Files plus `b.TempDir()` für SQLite. TS-Benchmarks
  nutzen `master.m3u8` aus dem Fixtures-Pfad und generieren große
  Master/Media-Manifeste plus DASH-MPDs synthetisch im Bench-File
  (Tranche-1b-Commit).
- [/] Beobachtungsphase: erste 3-5 grüne CI-Läufe sind nicht-
  blockierend; danach wird `make benchmark-smoke` PR-blockierend
  (in `make gates` aufgenommen). **Stand 2026-05-07
  (Tranche-1c-Commit)**: Beobachtungsphase **läuft**.
  CI-Workflow `.github/workflows/benchmark-observation.yml` ist
  aktiv (Cron `30 2 * * *` UTC plus `workflow_dispatch`); beide
  Bench-Steps und der Job selbst tragen `continue-on-error: true`,
  damit Drift-Failures während der Beobachtung den Workflow nicht
  rot markieren. Bench-Output wird als Artefakt
  `bench-observation-<run_id>` mit 14 Tagen Retention hochgeladen
  — Vorbereitung für Tranche 2 (Nightly-`benchstat`-Regressionen).
  Folge-Commit (nach N=3-5 grünen Läufen) entfernt die drei
  `continue-on-error: true`-Marker und nimmt `make benchmark-smoke`
  in `make gates` auf; Tranchen-Tabelle Tranche 1 wechselt dann
  von 🟡 auf ✅.
- [x] Jeder Lauf druckt Runner-OS, CPU-Modell und Runtime-Versionen
  (Go, Node, pnpm), damit Budget-Failures einordenbar bleiben:
  `scripts/print-bench-runner-info.sh` läuft als erstes Step in
  beiden Make-Targets — druckt Date, Kernel, OS-PrettyName,
  CPU-Modell + Cores, Node-/pnpm-/Go-Versionen
  (Tranche-1b-Commit).

---

## 3. Tranche 2 — Nightly-`benchstat`-Regressionen

Bezug: `extra-gates.md` §3.3.

Ziel: Statistisch belastbare Trend-Erkennung mit Baseline-Vergleich.
Nicht im PR-Pfad; Nightly + Release-blockierend.

DoD:

- [x] CI-Workflow `.github/workflows/benchmark.yml` (`on: schedule:
  cron '0 4 * * *'` UTC plus `workflow_dispatch`) führt die
  Go-Benchmarks via `cd apps/api; go test -bench=. -benchmem
  -count=10 -benchtime=2s ./hexagon/... ./adapters/...` aus, lädt
  die Baseline aus dem Tranche-0-Pfad (orphan-Branch
  `benchmark-baseline` als File `benchmarks/api-bench.txt`) und
  vergleicht via `benchstat` aus `golang.org/x/perf/cmd/benchstat`.
  Ohne Baseline läuft der Workflow als observation-only mit Notice
  und exitet ohne Vergleich (Tranche-2a-Commit).
- [x] Regressions-Schwelle ist explizit: **+15 % auf statistisch
  signifikantem Ergebnis (p < 0.05)**. Implementiert in
  `scripts/check-benchstat-regression.mjs`; Schwelle als
  `--threshold-percent`-Flag konfigurierbar, Default 15. Schwelle
  ist im Workflow als Aufruf-Argument dokumentiert; parallel-
  Eintrag in `docs/perf/budgets.md` §5 Wartung folgt mit Tranche-
  2b-Commit (Tranche-2a-Commit).
- [x] `benchstat`-Output wird als Workflow-Artefakt
  `bench-regression-<run_id>` gespeichert mit **30 Tagen
  Retention** — enthält `current.txt`, `baseline.txt` (falls
  vorhanden) und `comparison.txt` (Tranche-2a-Commit).
- [x] Bei Regression: Auto-Issue mit Workflow-Run-URL,
  benchstat-Diff im `comparison.txt`-Block, lokaler Repro-Befehl
  und Drift-Akzeptanz-Pfad. Labels `performance,benchmark,
  plan-0.9.5`. Issue wird unconditional erstellt (kein
  `secrets.*`-Gate; Performance-Drift ist immer team-relevant)
  (Tranche-2a-Commit).
- [x] Release-Gate in `docs/user/releasing.md` neue §2.5
  („Benchmark-Regression-Gate") referenziert den letzten grünen
  Nightly-Run vor Release-Tag als **Pflicht-Voraussetzung für
  Minor-Releases** (`0.X.0`); Patch-Releases (`0.X.Y`) sind
  ausgenommen, weil sie die Performance-Charakteristik nicht
  ändern und über `make benchmark-smoke` (PR-Pfad) abgesichert
  werden. Block dokumentiert auch Quarantäne-Tag-Format und
  Operator-Pfad bei Regression-Issue (Tranche-2b-Commit).
- [x] Quarantäne-Mechanik: ein Benchmark kann via
  `// bench:quarantine YYYY-MM-DD reason: <text>`-Kommentar direkt
  über der `func BenchmarkX(...)` (Go) bzw. dem `bench("...",
  ...)`-Aufruf (TS) aus dem Vergleich genommen werden. Skript
  `scripts/check-bench-quarantines.mjs` (executable) scant
  `apps/api` und `packages/stream-analyzer` rekursiv, mappt jeden
  Tag auf den nachfolgenden Bench-Namen (max. 5 Zeilen Suchradius)
  und failed mit Exit-1, sobald ein Tag älter als die
  konfigurierbare Maximal-Quarantäne (Default 30 Tage) ist. Liste
  aktiver Quarantänen wird optional als JSON unter `--output
  <path>` für Folge-Konsumenten geschrieben. Workflow ruft den
  Check als ersten Step nach Setup; expired Tag = Workflow-Failure
  vor benchstat-Lauf. Operator-Doku in `docs/user/releasing.md`
  §2.5 (Tranche-2b-Commit).

---

## 4. Tranche 3 — Selektives Fuzzing + Property Tests

Bezug: `extra-gates.md` §3.5.

Ziel: Edge-Case-Robustheit für Parser/Decoder/Validierung. Nicht
PR-blockierend; opt-in `make fuzz-check` mit kurzem Budget,
Nightly mit längerem Budget.

DoD:

- [x] Go-Fuzz-Targets für mindestens: Cursor Encode/Decode (aus
  ADR-0004), HTTP-Validation für Playback-Event-Batches,
  Event-Meta-Validation (`webrtc.*`-Allowlist aus `0.8.0`),
  SRT-Health-Mapping. Sechs Fuzz-Targets in vier Packages
  (`apps/api/adapters/driving/http/cursor_fuzz_internal_test.go`,
  `apps/api/adapters/driving/http/wire_fuzz_internal_test.go`,
  `apps/api/hexagon/application/event_meta_validation_fuzz_internal_test.go`,
  `apps/api/adapters/driven/srt/mediamtxclient/mapping_fuzz_internal_test.go`).
  Erstfund über `FuzzMapMediaMtxItem`: `mbpsLinkCapacity=-1` leakte
  in `AvailableBandwidthBPS=-1_000_000`; Fix in `mapping.go`
  (Tranche-3a-Commit `53adbab`).
- [x] TypeScript-Property-Tests via `fast-check` (4.4.0,
  devDependency in `packages/stream-analyzer` und
  `packages/player-sdk`) für die drei Pflicht-Bereiche:
  - `packages/stream-analyzer/tests/hls-parser.property.test.ts`:
    zwei Properties — beliebige Eingaben mit `#EXTM3U`-Header
    produzieren ein deterministisches `AnalysisResult` (kein
    Crash, `analyzerKind:"hls"`, `playlistType` ∈ `master`/
    `media`/`unknown`); non-HLS/non-DASH-Bodies ergeben hart
    `manifest_not_supported`.
  - `packages/stream-analyzer/tests/dash-parser.property.test.ts`:
    drei Properties — Detector-Klassifikation `<?xml`/`<MPD`-
    Präfix → `dash`; well-formed MPD → deterministisches Result
    mit `analyzerKind:"dash"` / `playlistType:"dash"` /
    `details.type` (static/dynamic) und passendem
    `summary.itemCount`. Plan-§4 verlangte DASH „sobald 0.9.0
    Tranche 3 produktiv ist" — seit Commit `b241b7d` der Fall.
  - `packages/player-sdk/tests/redact.property.test.ts`: drei
    Properties — bounded ASCII/Sentinel-Inputs throwen nicht;
    JWT-Shape-Pfadsegmente werden zu `:redacted`; lange Hex-
    Pfadsegmente werden zu `:redacted`. Lehre aus dem ersten
    Bench-Lauf: `fc.webUrl(...)` und `fc.stringMatching(...).filter(...)`
    haben fast-check 4.4 in einen Discard-Loop geschickt
    (vitest-Workers liefen 30+ min auf 97% CPU). Lösung in der
    Spec dokumentiert: alle Properties nutzen deterministische
    Generators mit fixer Länge (`fc.constantFrom` + `fc.array`),
    kein `.filter()`-Pfad. Plus `interruptAfterTimeLimit: 4_000`
    als Schutznetz; Folge-Backlog-Item für vollständigen URL-
    Redaction-Korpus mit `fc.webUrl` sobald fast-check den
    Discard-Pfad hardened.
  Tests laufen über `make ts-test` (Plan-DoD §4-2-Item;
  vitest-Bench-fertig); 14 zusätzliche Property-Tests dazu, alle
  grün.
- [x] `make fuzz-check`-Target im Root-`Makefile` mit kurzem
  `-fuzztime` (Default `30s`); CI-Stage opt-in (manueller Trigger
  oder Nightly). `make fuzz-check` ist Wrapper auf
  `apps/api/Makefile::fuzz-check` (Container-basierter Go-Fuzz-Lauf
  in `golang:1.26`) plus die TS-Property-Tests via `make ts-test`.
  Greppt alle `^func Fuzz...` aus `*_fuzz_test.go`/
  `*_fuzz_internal_test.go` automatisch — keine Registry-Pflege.
  Override `FUZZTIME` per Env (`FUZZTIME=120s make api-fuzz-check`).
  **Opt-in**, nicht in `make gates` (Tranche-3c-Commit).
- [x] Nightly-Workflow erweitert: längeres Fuzz-Budget (z. B.
  10 min pro Target); gefundene Regressions werden als Issue
  auto-erstellt mit Repro-Test. `.github/workflows/fuzz.yml`:
  Cron `0 5 * * *` UTC plus `workflow_dispatch`-Input `fuzztime`
  (Default `5m` pro Target ⇒ ≈ 30 min Gesamt-Laufzeit über sechs
  Targets). Crash-Inputs werden via `find -newer go.mod` aus
  `testdata/fuzz/<Target>/` eingesammelt, als Artefakt
  `fuzz-nightly-<run_id>` mit 30 Tagen Retention hochgeladen, und
  ein Issue mit Labels `fuzz,quality,plan-0.9.5` wird automatisch
  geöffnet — der Issue-Body verweist auf den Repo-Pfad
  `apps/api/<package>/testdata/fuzz/<Target>/<id>` als permanenter
  Regression-Seed (Tranche-3c-Commit).
- [x] Doku in `docs/dev/fuzzing.md` (oder ähnlich): Liste der
  aktiven Fuzz-Targets, lokale Reproduktion, Sample-Korpus-Pfad.
  [`docs/dev/fuzzing.md`](../../dev/fuzzing.md) listet alle sechs
  Go-Fuzz-Targets plus die drei TS-Property-Test-Suites mit
  Pflicht-Invariante, dokumentiert den `make fuzz-check`-Pfad,
  die `gh run download`-Mechanik aus dem Nightly-Workflow, die
  zwei Korpus-Schichten (`f.Add`-Seeds vs. generierte
  `testdata/fuzz/`-Files) und die Tranche-3b-Lehre zu
  fast-check-Discard-Loops (Tranche-3c-Commit).

---

## 5. Tranche 4 — Mutation Testing (Nightly-Report, nicht-blockierend)

Bezug: `extra-gates.md` §3.6.

Ziel: Test-Qualität messen über Coverage hinaus. Initial nicht-
blockierend; nur Reporting.

DoD:

- [x] Mutation-Tool entschieden: **gremlins** (`github.com/go-
  gremlins/gremlins`) für Go statt go-mutesting (Substitution
  begründet in [`docs/dev/mutation-testing.md`](../../dev/mutation-testing.md) §1:
  go-mutesting seit 2022 unmaintained, AST-Brüche auf Go 1.21+);
  Modul: `apps/api/hexagon/application/event_meta_validation.go`
  (gemutiert als Teil des `hexagon/application`-Packages).
  **StrykerJS** + `@stryker-mutator/vitest-runner` für TypeScript;
  Modul: `packages/player-sdk/src/adapters/webrtc/sampling.ts`
  (Stryker `mutate`-Scope auf das eine File begrenzt). Beide
  Module sind sicherheits-relevant (Event-Meta-Reserved-
  Namespace + WebRTC-`getStats()`-Wire-Mapping); Test-Surface
  jeweils > 1× LoC der Pilot-Datei. Auswahl-Begründung in
  Doku §2 (Tranche-4-Commit).
- [x] `make mutation-report`-Target im Root-`Makefile`; läuft auf
  einem Modul gleichzeitig (nicht repo-weit). Wrapper für
  `make api-mutation-report` (gremlins via golang:1.26-Container,
  `go install`-zur-Laufzeit) und `make ts-mutation-report`
  (StrykerJS via `pnpm dlx`, kein devDep-Pinning im player-sdk).
  Beide Sub-Targets sind opt-in (NICHT in `make gates`); Stryker-
  Konfig in `packages/player-sdk/stryker.conf.cjs` mit `mutate`-
  Scope auf `src/adapters/webrtc/sampling.ts` (Tranche-4-Commit).
- [x] Nightly-Workflow erweitert: führt das Target aus, lädt den
  HTML-Report als Artefakt hoch. Neuer Workflow
  `.github/workflows/mutation.yml` mit Cron `0 6 * * *` UTC
  (Slot nach fuzz.yml 05:00, kein Konflikt). Zwei Jobs
  (`mutation-go` + `mutation-ts`), beide `continue-on-error:
  true` (Plan-DoD §5: nicht-blockierend). Artefakte
  `mutation-go-<run_id>` (gremlins-JSON + stdout) und
  `mutation-ts-<run_id>` (Stryker-HTML + JSON) mit 30 Tagen
  Retention (Tranche-4-Commit).
- [x] Score-Schwelle dokumentiert (z. B. > 70 % Mutation-Score als
  Wunsch-Ziel; PR-Blockierung erst, wenn das Modul die Schwelle
  drei Beobachtungsläufe in Folge erreicht).
  [`docs/dev/mutation-testing.md`](../../dev/mutation-testing.md) §3 dokumentiert die
  Übergangs-Mechanik: < 60 % → Folge-Backlog-Item; 60-70 % →
  Beobachtungsphase-Mittelfeld; > 70 % drei Nightly-Runs in
  Folge → Folge-Commit nimmt `continue-on-error: true` raus und
  setzt `--threshold-break=70`. > 80 % gilt als „grün" (Stryker
  `thresholds.high`). Score-Senkungen sind begründungspflichtig
  (Tranche-4-Commit).
- [x] Doku in `docs/dev/mutation-testing.md`: Liste der Module,
  aktueller Score, lokale Reproduktion.
  [`docs/dev/mutation-testing.md`](../../dev/mutation-testing.md) listet die zwei
  Pilot-Module mit Test-Surface, dokumentiert Tool-Substitution
  (gremlins statt go-mutesting), Score-Schwelle und Übergangs-
  Pfad zur PR-Blockierung, lokale Reproduktion (`make
  mutation-report` / `api-mutation-report` /
  `ts-mutation-report`), Trend-Tracking via `gh run download`,
  und Quarantäne-Politik (kein expliziter Quarantäne-Pfad —
  Mutation-Tests sind nicht-blockierend, flaky Läufe rauschen im
  Trend durch). Aktueller Score wird **nicht** statisch
  eingetragen — der nächste Nightly liefert den ersten Wert
  (Tranche-4-Commit).

---

## 6. Tranche 5 — Release-Doku, Closeout

Bezug: `plan-0.8.5.md` §6 als Vorlage.

DoD:

- [x] `docs/user/releasing.md` §3 referenziert Wave-2-Gates
  (`make benchmark-smoke` PR-blockierend, `make fuzz-check` und
  `make mutation-report` opt-in/Nightly); Release-Voraussetzung
  ist „letzter Nightly-Benchmark grün". `releasing.md` §2.6 neu
  hinzugefügt (Fuzz- und Mutation-Beobachtungs-Gates), §3.1
  ergänzt um Wave-2-Quality-Gates-Voraussetzung pro Release-Tag
  (Tranche-5-Commit).
- [x] `README.md` Status-Block erwähnt `0.9.5` als Patch-Release
  mit Quality-Gates Wave 2. Header-Block und Sektion „Aktueller
  Stand" listen die vier Tranche-Lieferungen explizit
  (Tranche-5-Commit).
- [x] Versions-Bump 0.9.0 → 0.9.5 (analog `plan-0.8.5.md` §4
  Closeout-Mechanik). 39 Dateien per Bulk-`xargs sed` von
  `"0.9.1"` → `"0.9.5"`; drei zusätzliche Stellen mit Versions-
  Strings in Test-Error-Messages (`apps/api/.../streamanalyzer/
  contract_test.go`, `http_test.go`, `packages/stream-analyzer/
  tests/version.test.ts`) per Edit nachgezogen (Tranche-5-Commit).
- [x] CHANGELOG: [Unreleased]-Block in `[0.9.5] - 2026-05-07`
  umgewandelt; Block listet die vier Tranche-Lieferungen plus
  den Erstfund aus `FuzzMapMediaMtxItem` und den Versions-Bump
  (Tranche-5-Commit).
- [x] `make docs-check` grün; `make gates` grün;
  `make benchmark-smoke` grün; Nightly-Workflow läuft sauber
  durch. `make gates` nach dem Closeout-Commit (Plan-DoD §0.5
  Pflicht: Drift-Gate vergleicht working-tree gegen HEAD; vor-
  Commit würde der Versions-Bump als Drift gewertet).
  `benchmark-smoke` ist opt-in und läuft im Beobachtungs-Nightly,
  daher nicht Teil von `make gates` — separat verifiziert
  (Tranche-5-Commit).
- [x] `plan-0.9.5.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben; Cross-Refs angepasst;
  Roadmap §3 zeigt `0.9.5` ✅ (Tranche-5-Commit).
- [x] Tag `v0.9.5` annotiert; Push opt-in; GitHub-Release
  (Tranche-5-Commit; Tag-Hash siehe Release-Notes).

---

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen.
- `extra-gates.md` ist Master-Backlog für die sechs Gates. Wenn
  ein neues Gate hinzukommt (z. B. License-Compliance), wird es
  zuerst dort erfasst, dann in einem Folge-Plan konkretisiert.
- Quarantäne-Tags in Benchmarks sind self-expiring nach 30 Tagen;
  Tranche-2-Workflow setzt das automatisch um.
- Beobachtungsphase für PR-Blockierung von Wave-2-Gates: jeder
  neue Gate startet nicht-blockierend, wird erst nach
  3-5 grünen Läufen blockierend. Diese Übergänge werden im Plan
  als Wartungs-Eintrag vermerkt.
