# Performance-Budgets

> **Status**: Initial-Tabelle — **Tranche-0-Stand** aus
> [`docs/planning/done/plan-0.9.5.md`](../planning/done/plan-0.9.5.md)
> §1a. Werte sind bewusst großzügig (Architektur-basiert, **noch
> nicht** mess-basiert). Tranche 1 ersetzt die Initial-Werte nach
> N=3-5 grünen Beobachtungsläufen durch realistische Schwellen
> (Plan-DoD: „bewusst großzügig (Faktor 2-3 über aktueller Messung)";
> diese Tabelle hält die obere Grenze, nicht die aktuelle Messung).

## 1. Zweck

Single-Source-of-Truth für die `make api-benchmark-smoke` /
`make analyzer-benchmark-smoke` / `make benchmark-smoke`-Targets aus
[`extra-gates.md`](../planning/open/extra-gates.md) §3.2. Jeder
Smoke schlägt als PR-Block fehl, wenn ein Hot-Path die hier
gelistete Schwelle überschreitet. Die Schwellen sind **absolute
Obergrenzen**, kein Vergleich gegen den letzten Commit (das ist
`benchstat`-Aufgabe in Tranche 2).

## 2. Konvention

- **Plattform**: GitHub Actions `ubuntu-24.04` (PR-Pfad). Lokale
  Läufe können niedrigere Werte zeigen; Budgets sind für CI gesetzt.
- **Messprotokoll**: jeder Smoke-Run druckt Runner-OS, CPU-Modell
  und relevante Runtime-Versionen (Go, Node, pnpm) — damit ein
  Budget-Failure einordenbar bleibt.
- **Beobachtungsphase**: neue oder geänderte Budgets laufen erst
  N=3-5 grüne CI-Beobachtungsläufe nicht-blockierend mit, bevor sie
  PRs blockieren. Die Beobachtungsphase wird im Plan-DoD pro
  Budget-Zeile vermerkt.
- **Aktualisierung**: jede Schärfung eines Budgets ist eine
  Plan-DoD-Item-Änderung; reine Lockerung (Budget hochsetzen) braucht
  einen risks-backlog- oder Folge-Item-Eintrag.

## 3. API-Hot-Paths (`apps/api`, Go)

Quelle: `extra-gates.md` §3.2 API-Kandidaten. Budgets sind
Wall-Clock pro Aufruf bzw. pro N-Items, gemessen mit
`go test -bench=. -benchmem`.

| Modul | Hot-Path | Budget (initial) | Begründung (Tranche 0) |
| --- | --- | --- | --- |
| `apps/api/hexagon/application` | `RegisterPlaybackEventBatch` (typische 100-Event-Batch, In-Memory-Repo) | ≤ 10 ms / Batch | Cardinality-Validierung + Domain-Mapping + Sequenzvergabe; CI-Runner ist konservativ. |
| `apps/api/hexagon/application` | `RegisterPlaybackEventBatch` (Maximal-Batch laut spec/telemetry-model.md §4.1: 100 Events / 256 KiB Body) | ≤ 25 ms / Batch | gleicher Pfad, aber inkl. Per-Event-Meta-Validation und Batch-Ende-Lifecycle-Tick. |
| `apps/api/adapters/driven/persistence/sqlite` | Event-Append + Sequence-Allocation (typische 100-Event-Batch) | ≤ 100 ms / Batch | SQLite-WAL plus Sequenzvergabe; PR-CI ohne `tmpfs`-Boost. |
| `apps/api/hexagon/application` | `ListStreamSessions` (Default-Limit 100, gefüllte 1k-Session-DB) | ≤ 50 ms / Page | Cursor-Decode + Index-Scan + Domain-Hydratation. |
| `apps/api/adapters/driven/persistence/sqlite` | `cursor.Encode/Decode` (Cursor-v3 inkl. Process-Instance-Stamp) | ≤ 250 µs / Pair | Reine String-Konversion plus HMAC-Sign-Free-Path. |

## 4. Stream-Analyzer-Hot-Paths (`packages/stream-analyzer`, TypeScript)

Quelle: `extra-gates.md` §3.2 Stream-Analyzer-Kandidaten. Budgets
sind Wall-Clock pro Aufruf, gemessen mit Tinybench oder
vitest-bench (Tranche 0 entscheidet sich nicht zwischen den beiden;
Tranche 1 wählt). DASH-Pfad ist seit `0.9.0` Tranche 3 produktiv
und wird ab Tranche 1 mit gemessen.

| Modul | Hot-Path | Budget (initial) | Begründung (Tranche 0) |
| --- | --- | --- | --- |
| `internal/parsers/master.ts` | HLS Master Playlist klein (1-5 Variants + 1-3 Renditions) | ≤ 5 ms | Pure-Function-Parser ohne IO; ein-Pass-Scan. |
| `internal/parsers/master.ts` | HLS Master Playlist groß (50+ Variants, 20+ Renditions) | ≤ 25 ms | gleicher Pfad, aber Variant-Cross-Check und Group-ID-Lookups skalieren mit n. |
| `internal/parsers/media.ts` | HLS Media Playlist mit 1.000 `#EXTINF`-Segmenten | ≤ 50 ms | Segment-Aggregat-Statistiken plus Toleranzregel-Findings. |
| `internal/parsers/dash.ts` | DASH-MPD VOD (1 Period, 2 AdaptationSets, 5 Representations) | ≤ 5 ms | Regex-Parser ohne Dependency; 0.9.0-Tranche-3-Spec-Stand. |
| `internal/parsers/dash.ts` | DASH-MPD Live (`type=dynamic`, 3 AdaptationSets, 10 Representations) | ≤ 10 ms | wie VOD, plus Live-Felder; SegmentTemplate-Edge-Cases out-of-scope. |
| `internal/parsers/detect.ts` | Detector über ein 256-KiB-Body-Sample | ≤ 500 µs | erster lokaler Bench-Lauf (Tranche 1b, 2026-05-07): mean 207 µs / p75 268 µs auf Dev-Rechner — `firstNonEmptyLine` scannt aktuell den ganzen Body via `split(/\r?\n/)` statt nur das Präfix; Optimierung ist Folge-Plan, Budget bleibt großzügig (Faktor ~2× über mean) damit der CI-Runner-Faktor und der p75-Drift Headroom haben. |
| `internal/loader/ssrf.ts` | URL-Klassifizierung (typischer Allowlist- + Blocklist-Mix, 100 Calls) | ≤ 5 ms / 100 Calls | regex-basierte Hostname-Klassifikation plus IPv4/IPv6-Parser. |

## 5. Wartung

- **Beobachtungsphase**: jeder neue Budget-Eintrag startet
  nicht-blockierend (warning-only). Erst nach N=3-5 grünen
  CI-Beobachtungsläufen ohne Drift wird der Eintrag PR-blockierend
  geschaltet — die DoD-Item-Zeile im Plan dokumentiert den
  Übergang.
- **Drift-Strategie**: ein einzelner Failure ist Diagnose-Anlass,
  keine Sofort-Schärfung. Tranche 1 dokumentiert die Quarantäne-
  Policy für laute Benchmarks (max. 30 Tage; Plan-Tranche 1 §1a).
- **Aktualisierung dieser Datei**: jede Schärfung kommt mit einem
  Plan-DoD-Item-Update; die Begründungs-Spalte trägt das Datum +
  „nach N grünen Läufen geschärft".

## 6. Out-of-Scope (Tranche 0)

- Kein Mikrobenchmark-Vergleich (`benchstat`) — das ist Tranche 2,
  Nightly-Pfad mit Baseline-Branch.
- Kein WebRTC-Stats-Sampling-Benchmark — der Player-SDK-Adapter
  läuft im Browser; Performance ist über das `0.8.0`-Bundle-
  Budget aus `packages/player-sdk/scripts/performance-smoke.mjs`
  abgedeckt.
- Keine End-to-End-/Lab-Performance-Smokes — Compose-Stacks bleiben
  außerhalb des Budget-Smokes.
