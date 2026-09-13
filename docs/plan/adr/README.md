# ADR-Index — m-trace

**Derivativ:** Quelle der Wahrheit sind die ADR-Dateien; dieser Index ist
eine Bequemlichkeits-Sicht — bei jedem neuen/akzeptierten ADR mitziehen.

| ID | Titel | Status | Bezug |
|---|---|---|---|
| [0001](0001-backend-stack.md) | Backend-Technologie für `apps/api` | Accepted | `spec/lastenheft.md` §9.1, §10.1 |
| [0002](0002-persistence-store.md) | Persistenz-Store für Sessions und Playback-Events | Accepted | `spec/lastenheft.md` OE-3, MVP-16, MVP-27, MVP-40, RAK-32 |
| [0003](0003-live-updates.md) | Live-Updates für Dashboard-Session-Verläufe | Accepted | `spec/lastenheft.md` OE-5, MVP-31, RAK-29, RAK-32 |
| [0004](0004-cursor-strategy.md) | Dauerhaft konsistente Cursor-Strategie für Pagination | Accepted | `spec/lastenheft.md` RAK-32; [ADR-0002](0002-persistence-store.md) |
| [0005](0005-production-ops-backends.md) | Production- und Ops-Backends als optionale Seeds | Accepted | `spec/lastenheft.md` NF-18, MVP-40..MVP-44 |
| [0006](0006-postgres-scaleout-adapter.md) | Postgres-Runtime-Adapter für Production-Scale-out | Accepted | `spec/lastenheft.md` RAK-91 |
| [0007](0007-sqlite-postgres-data-cutover.md) | SQLite→Postgres-Datenmigration / Cutover (optional) | Accepted | `spec/lastenheft.md` RAK-91 |
| [0008](0008-benchmark-mutation-execution-in-docker.md) | Ausführungsort der Bench-/Mutation-Gates: Docker | Accepted | [`docs/perf/budgets.md`](../perf/budgets.md) §2/§3/§4 |
| [0009](0009-harness-baseline-v3.5.0.md) | Regelwerk-Baseline auf ai-harness-course v3.5.0: strukturelle Adoption | Accepted | [`harness/conventions.md`](../../harness/conventions.md) §Baseline |
| [0010](0010-closure-note-pflicht.md) | Closure-Note-Pflicht für abgeschlossene Pläne | Accepted | [ADR-0009](0009-harness-baseline-v3.5.0.md) |
| [0011](0011-harness-baseline-v3.5.1-bump.md) | Regelwerk-Baseline-Bump v3.5.0 → v3.5.1 (nicht-struktureller Re-Vendor) | Accepted | [ADR-0009](0009-harness-baseline-v3.5.0.md) |

## Konventionen

- ADRs sind nach `Accepted` **immutable** (siehe Baseline-Regelwerk `modul-04-adrs.md`).
- Schärfungen entstehen als neue ADR mit `Supersedes ADR-NNNN`.
- Bei `Accepted`: diesen Index aktualisieren (Status, Datum).
- Jede ADR deklariert im `**Schärft:**`-Feld *aufwärts*, welche Spec-Stelle
  sie verbindlich macht (Baseline-Regelwerk `grundlagen-referenz-richtung.md`
  §Referenz-Richtung (SDP)) — als Kennung (`SPEC-*`, `ARC-*`,
  `<PREFIX>-FA-*.<Buchstabe>`), ersatzweise als Abschnitt, wo die Sektion
  keine Kennungen vergibt. Prozess-ADRs ohne Spec-Stratum tragen `—`.
- ADR-0001..0008 tragen keinen `## Re-Evaluierungs-Trigger`-Abschnitt
  ([MR-009](../../harness/conventions.md#mr-009), grandfathered);
  ADR-0009..0011 haben ihn bereits.
