# Performance-Budgets

> **Status**: PR-blockierende Budget-Smokes — initiale Tabelle aus
> [`docs/planning/done/plan-0.9.5.md`](../planning/done/plan-0.9.5.md)
> §1a, in `0.22.0` nach fünf grünen Beobachtungsläufen
> (`benchmark-observation.yml` Runs `25592982776`..`25769811661`)
> PR-blockierend über `make gates` geschaltet. Werte bleiben bewusst
> großzügige Obergrenzen, nicht die aktuelle Messung.

## 1. Zweck

Single-Source-of-Truth für die `make api-benchmark-smoke` /
`make analyzer-benchmark-smoke` / `make benchmark-smoke`-Targets aus
[`extra-gates.md`](../planning/in-progress/extra-gates.md) §3.2. Jeder
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
- **Beobachtungsphase**: neue oder geänderte Budgets laufen weiterhin
  erst N=3-5 grüne CI-Beobachtungsläufe nicht-blockierend mit, bevor
  sie PRs blockieren. Die aktuelle Budget-Tabelle ist seit `0.22.0`
  blockierend.
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
- Keine End-to-End-/Lab-Performance-Smokes im **Budget-Smoke**
  (`make benchmark-smoke`, PR-blockierend) — Compose-Stacks bleiben dort
  außen vor. Lab-Lastfähigkeit deckt der separate, opt-in Load-/Soak-Smoke
  ab (§7), nicht-blockierend.

## 7. Load-/Soak-Smoke (`make smoke-load`, opt-in)

`make smoke-load` (`scripts/smoke-load.sh` + `scripts/load/playback-events.k6.js`)
belegt die Lab-Lastfähigkeit der Ingest→Persistenz→Read-Kette
(NF-20/NF-22/NF-23) unter echter Parallelität — komplementär zu den
isolierten Hot-Path-Budgets oben. **Opt-in + Nightly, NICHT in
`make gates`** (lastabhängig hardware-/runner-flaky). Zwei Modi:
`MODE=capacity` (Rate-Limit angehoben → echte Ingest-Kapazität),
`MODE=contract` (Default-100/s → Limiter-Verhalten).

**Gates (Pass/Fail), keine reinen Latenz-Budgets:**

| Kriterium | Schwelle | Begründung |
|---|---|---|
| Kein stiller Verlust | `persisted >= accepted` (Readback gegen die echte `playback_events`-Tabelle, **nicht** `event_count`) | `persisted` = tatsächlich in `playback_events` liegende Events der Lauf-Sessions, im Autostart-/CI-Pfad per direktem `SELECT count(*)` gegen das api-Volume gezählt (O(1)); der `events[]`-Array-Readback des Detail-Endpoints bleibt portabler Fallback für `SMOKE_LOAD_AUTOSTART=0` (s. risks-backlog R-25). Jedes client-bestätigte (`202`) Event muss dort liegen. `persisted < accepted` = stiller Verlust = FAIL. Ein Überschuss (`persisted > accepted`) ist **at-least-once unter Überlast** (Append erfolgreich, Client sah aber ein Timeout/`5xx`), kein Verlust. `event_count` (Session-Zähler, im Upsert vor dem Append getickt) taugt dafür ausdrücklich nicht. |
| Fehlerquote | `<= MAX_ERROR_PCT` (Default 5 %) | Anteil Events mit Status ≠ `202`/`429`. An der SQLite-Sättigung sind einzelne **explizite** Fehler erwartbar (graceful degradation); nur eine katastrophale Quote bricht. |
| Limiter (contract) / Override (capacity) | `429 > 0` bzw. `accepted > 0` | Sanity, dass der jeweilige Modus tatsächlich greift. |

Eine Latenz-Obergrenze ist **bewusst kein** harter Gate: unter
unbeschränkter VU-Last an der Kapazitätsgrenze ist hohe p95 inhärent
(siehe Referenz). Hot-Path-Latenz deckt §3/§4 ab.

**Referenz-Messungen (runner-abhängig, nicht PR-blockierend):**

CI (`ubuntu-24.04`, `MODE=capacity`, Batch 20):

- **Open-loop SLO**, 2000 ev/s offered / 10 min (Run `27696206008`):
  2000 ev/s akzeptiert, **p95 7,6 ms** (Budget 1000 ms), 0 `429`, 0
  Fehler, 0 dropped, `persisted == accepted`. Großer Headroom — hier weit
  von der Sättigung.
- **Closed-loop Soak**, 20 VUs / 4 h (Run `27665523746`): **3842 ev/s**
  akzeptiert, **p95 636 ms** / max 7,4 s, 0 `429`, 0,01 % Fehler,
  55.327.560 Events, `persisted == accepted`.

Dev-Laptop (historisch, 20 VUs / 30 s / Batch 20):

- **Kapazitäts-Modus**: ~**800 Events/s** akzeptiert, p95 ~**2,3 s** an
  der Sättigung — dieser Wert war laptop-/disk-gebunden; der CI-Runner
  hält bei höherem Durchsatz deutlich niedrigere p95, die ~800 sind also
  **nicht** die SQLite-Decke.
- **Contract-Modus** (Default 100/s): ~**103 Events/s** akzeptiert,
  Rest `429`, p95 ~**2,6 ms**, 0 Fehler, Reconciliation exakt.

**Kern-Befund: der SQLite-Single-Writer bleibt der Ingest-Engpass** (ein
serialisierter Writer), aber die belastbare Decke liegt hardware-abhängig
**oberhalb von ~3842 ev/s** (CI, dort noch graceful) — die frühere
~800-ev/s-Zahl war ein Dev-Laptop-Artefakt. Empirische Grundlage für die
Evaluierung des [ADR-0005](../adr/0005-production-ops-backends.md)-Postgres-
Triggers. Die Soak-Variante (`make smoke-soak`) misst zusätzlich die
Read-Retention-p95 gegen 2 s als **Trigger-#3-Evidenz (Proxy)**: die
Read-API serviert nur keyset-indizierte (größenunabhängige) Reads, keinen
Full-Scan-/Aggregat-Retention-Query. **Stand 2026-06-17**: der ≥-10-Mio-
Lauf ist gefahren — bei 55,3 Mio Events Read-Retention-p95 **12 ms**
(list 2,1 ms / detail 11,8 ms) ≪ 2 s → **ADR-0005-Trigger #3 nicht
ausgelöst**. Käme je eine Korpus-Scan-Query in die API, muss die Probe um
genau diese erweitert werden.

Abgrenzung: Single-Instance, **ein** Projekt/Token. Produktive
Multi-Tenant-Isolation und Multi-Replica-Skalierung sind hier **nicht**
belegt — getrackt als risks-backlog **R-26** (Machbarkeit) bzw.
ADR-0005-Trigger #1 (Multi-Replica → Postgres). Die Multi-Replica-Achse
(R-26 c) belegt §8.

## 8. Scale-out-Lasttest (`make smoke-scaleout-load`, opt-in)

`make smoke-scaleout-load` (`scripts/smoke-scaleout-load.sh`) treibt
k6-Playback-Event-Last gegen den Multi-Replica-Stack
(`docker-compose.scaleout.yml`: 2 API-Replicas + **geteilter** Postgres +
nginx-LB) und liefert die horizontale Scale-out-Evidenz aus
[ADR-0006](../adr/0006-postgres-scaleout-adapter.md) (**R-26 c**). Readback
via `psql` gegen den geteilten Postgres (kein SQLite-GLOB/Volume-Hack):
`persisted = COUNT(*)`, `distinct = COUNT(DISTINCT ingest_sequence)` je
Session-Prefix. Opt-in, **nicht** in `make gates`.

**Gate (Pass/Fail): Korrektheit über Replicas.**

| Kriterium | Schwelle | Begründung |
|---|---|---|
| Kein stiller Verlust über Replicas | `persisted == accepted` | Jedes client-bestätigte (`202`) Event muss im geteilten Store liegen. |
| Keine Duplikate über Replicas | `COUNT(DISTINCT ingest_sequence) == COUNT(*)` | Der DB-autoritative Sequencer (**R-28**, `nextval`+Block-Allokation) muss über parallele Writer kollisionsfreie `ingest_sequence` vergeben — explizit gezählt, nicht anzahl-inferentiell. |

**Referenz-Messungen (2026-07-11, 20-Kern-Host, 50 VUs / 60 s / Batch 20):**

Beide Läufe (throttled *und* unthrottled) über **~1,4 Mio Events**:
`persisted == accepted` und `distinct == COUNT(*)` — **0 Verlust, 0
Duplikate** über 2 konkurrierende Replicas.

| Modus | 1 Replica | 2 Replicas | Skalierung | Flaschenhals |
|---|---|---|---|---|
| **Throttled** (Default 100 ev/s/Projekt) | 101 ev/s | 203 ev/s | **2,01×** (linear) | Per-Replica-In-Memory-Limiter (App) |
| **Unthrottled** (Limiter aus) | **12.405 ev/s** | 11.221 ev/s | **0,9×** (flach) | Geteilter Single-Postgres |

**Kern-Befund: die Durchsatz-Skalierung ist flaschenhals-abhängig — und die
Korrektheit ist unabhängig davon wasserdicht.**

- **App-gebunden** (Limiter greift): 1→2 Replicas ist **linear (2,01×)**, weil
  der Ingest-Limiter **pro Replica** in-process liegt → N Replicas geben dem
  Projekt N× effektives Ratebudget. Das ist zugleich die **R-26-b-Lücke,
  jetzt gemessen**: ohne shared (Redis) Ingest-Limiter ist die Per-Projekt-
  Decke nicht repliken-übergreifend fair.
- **Store-gebunden** (Limiter aus): der **einzelne geteilte Postgres ist die
  Decke** (~12k ev/s); eine 2. App-Replica hebt den Durchsatz **nicht**
  (0,9×, minimal schlechter durch Contention auf einem Writer).
  `docker stats`-Attribution: Postgres **~9,5 Kerne** vs gesamte API-Schicht
  **~4 Kerne**, Host nur **~14/20 Kerne** genutzt → **Per-Instanz-Grenze des
  einen Postgres** (WAL/Commit-Serialisierung, `nextval`-/Lock-Contention),
  **kein** Host-CPU-Mangel. Ein einzelner Postgres-Replica hält **12.405
  ev/s** — 3× über dem SQLite-Single-Soak (§7, 3842 ev/s).

**Konsequenz.** Horizontale App-Repliken heben den Durchsatz nur bei
**app-gebundener** Last; bei **store-gebundener** Bulk-Ingest-Last ist der
Hebel der **Store** (größerer/partitionierter Postgres, `COPY`-Batching,
pgbouncer), nicht die Replica-Zahl. Der Scale-out-Adapter bleibt damit ein
belegter Korrektheits- und Betriebspfad; ein Durchsatz-Scaling *jenseits*
eines Single-Postgres ist ein eigener Folge-Scope. Connection-Budget-
Vorbehalt: `N × database/sql-Pool + Startup/Migration` muss unter Postgres'
`max_connections` bleiben (Default-Pool unbounded → beim Hochskalieren
begrenzen oder pgbouncer).
