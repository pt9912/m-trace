# Implementation Plan — `0.23.0` Postgres-Runtime-Adapter + Scale-out-Evidenz

> **Status**: 🚧 **Geplant** — Antwort auf die einzige verbleibende,
> unbelegte Architektur-Achse: **horizontale Production-Scale-out-Evidenz**
> (R-26 Teil c). Reaktiviert RAK-91 („defer" → „proceed, optional") über
> [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md), das die
> Postgres-Vertagung aus [ADR-0005](../../adr/0005-production-ops-backends.md)
> amendet. Noch nicht gebaut — tranchenweise Umsetzung mit Gate je Tranche.
>
> **Bezug**: RAK-91 (Lastenheft, Postgres-Entscheidung); NF-20/NF-22/NF-23
> (Lastfähigkeit); [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md);
> R-26 in [`risks-backlog.md`](../in-progress/risks-backlog.md);
> Vorläufer [`plan-0.22.5-load-smoke`](../done/plan-0.22.5-load-smoke.md)
> (Single-Instance-Lastfähigkeit belegt).

## 1. Ziel

`plan-0.22.5-load-smoke` hat die **vertikale** Lab-Lastfähigkeit belegt
(eine Instanz, ein Tenant, 55,3 Mio Events, p95 ms-Bereich, kein stiller
Verlust). Offen bleibt — festgehalten als **R-26** — die **horizontale**
Achse: Multi-Replica-Betrieb (≥ 2 API-Instanzen, geteilter Store) und
Multi-Tenant-Isolation unter Last. Multi-Replica ist mit SQLite
strukturell unmöglich (single-writer, file-gebunden); es braucht zwingend
Postgres als netzwerkfähigen Store.

Dieser Plan liefert den **optionalen Postgres-Runtime-Adapter** und die
**Multi-Replica-Harness + den Scale-out-Lasttest**, der die R-26-Lücke
mit Messwerten schließt — nicht „Scale-out ist Postgres-Gebiet"
(Behauptung), sondern „N Replicas auf Postgres halten X ev/s mit
linearer Skalierung, kein Verlust/Dup über Replicas" (Nachweis).

## 2. Scope / Abgrenzung

**In Scope:**

- Postgres-DDL aus dem neutralen `apps/api/internal/storage/schema.yaml`
  (Target `postgres`), `driverName`/DSN parametrisiert.
- Postgres-Persistenz-Adapter für die sechs Driven-Ports, die der
  SQLite-Adapter heute hält: `event_repository`, `session_repository`,
  `srt_health_repository`, `ingest_stream_repository`,
  `project_token_repository`, `ingest_sequencer`.
- Grünziehen der adapter-agnostischen Contract-Suite
  (`apps/api/adapters/driven/persistence/contract`) gegen Postgres.
- `MTRACE_PERSISTENCE=postgres`-Wiring (Default unverändert `sqlite`).
- Multi-Replica-Compose-Profil (≥ 2 api + 1 Postgres + LB).
- Scale-out-Lasttest als Erweiterung von `scripts/smoke-load.sh`.
- Lastenheft-Patch: RAK-91-Status „defer" → „proceed, optional".

**Nicht in Scope** (siehe [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md)
§Grenzen):

- Keine automatische SQLite→Postgres-Datenmigration bestehender Läufe
  (Postgres startet mit frischem Store).
- Kein Cloud-Managed-Postgres-Zwang, keine Backup-/PITR-/Replikations-
  Topologie als Pflicht (RPO/RTO = ADR-0005-Trigger #2, separat).
- SQLite bleibt Default — kein Ersetzen, kein Pflichtpfad.
- Analytics-Backends (RAK-92), K8s (RAK-93/NF-18) unverändert.
- **Multi-Tenant** (R-26 Teil b) ist orthogonal und kann eine eigene,
  kleinere Tranche/Plan sein (N-Projekt-Seeding + Token-Fan-out, läuft
  auf SQLite *und* Postgres). Hier nur mitgedacht, wo es den Scale-out-
  Test schärft (Tenant-Verteilung über Replicas).

## 3. Architektur-Berührung & Crux-Risiken

Die Architektur ist vorbereitet (hexagonal, neutraler Schema-Anker,
Contract-Suite, Redis-Rate-Limiter für Multi-Host bereits gelöst). Die
nicht-trivialen Punkte:

- **`ingest_sequence`-Monotonie unter Concurrent-Writern (Crux).** SQLite
  serialisiert per Single-Writer + `_txlock=immediate`; die
  `ingest_sequence` (AUTOINCREMENT) ist dort trivial monoton. Postgres-
  MVCC erlaubt echte Parallel-Writer (das ist der Sinn) — eine
  IDENTITY/Sequence ist eindeutig, aber **Commit-Reihenfolge ≠
  Sequence-Reihenfolge** (Lücken, out-of-order sichtbar-werden). Die
  Cursor-Pagination ([ADR-0004](../../adr/0004-cursor-strategy.md)) ordnet
  über `(server_received_at, sequence_number, ingest_sequence)` — es ist
  zu belegen, dass Keyset-Pagination unter Concurrent-Insert keine Events
  überspringt (z. B. via geeigneter Snapshot-/Isolation-Semantik oder
  einer commit-order-stabilen Sequenzquelle). **Eigene Risiko-Zeile.**
- **Transaktions-Isolation.** SQLite-`BEGIN IMMEDIATE` → Postgres-Default
  `READ COMMITTED` vs. `REPEATABLE READ`; der Batch-Append + Lifecycle-
  Tick muss dasselbe Verhalten wie der SQLite-Pfad zeigen (Contract-Suite
  als Wächter).
- **Dialekt-Differenzen.** Platzhalter (`?` → `$1`), `INTEGER PRIMARY KEY
  AUTOINCREMENT` → `BIGINT GENERATED ... AS IDENTITY`/`BIGSERIAL`,
  `INSERT ... ON CONFLICT`, Boolean (`INTEGER 0/1` → `boolean`), JSON-
  Spalten (`meta TEXT` → `text`/`jsonb`). Wird über das d-migrate-Target
  + adapter-lokale Query-Konstanten gekapselt.
- **Connection-Pooling.** SQLite = ein File-Handle; Postgres braucht
  `pgxpool`/`database/sql`-Pool-Sizing, abgestimmt auf die Replica-Zahl.

## 5. Tranchen

| Tranche | Inhalt | Gate |
| --- | --- | --- |
| 1 | **Schema + Migrations-Target.** d-migrate Postgres-DDL aus `schema.yaml`; `driverName`/DSN parametrisieren (heute `migrate.go` hartkodiert `sqlite`); Postgres-Container im Lab. | `make`-Target migriert eine frische Postgres-DB fehlerfrei; Schema-Parität zu SQLite (gleiche Tabellen/Indizes aus einem Anker). |
| 2 | **Postgres-Adapter (6 Ports).** `persistence/postgres` mit denselben Driven-Port-Implementierungen; Dialekt-Kapselung. | Adapter kompiliert, alle sechs Ports implementiert. |
| 3 | **Contract-Suite grün.** Die adapter-agnostische `contract`-Suite gegen Postgres laufen lassen + Postgres-spezifische Tests (Concurrent-Writer-Sequenz, Cursor-Walk unter Last). | Contract-Suite grün gegen SQLite **und** Postgres in CI. |
| 4 | **Wiring + CI-Matrix.** `MTRACE_PERSISTENCE=postgres` + DSN in `main.go` (Default `sqlite` byte-stabil); CI fährt die Persistenz-Tests gegen beide Stores. | `MTRACE_PERSISTENCE=sqlite` unverändert; `=postgres` boot't + gleicher Smoke grün. |
| 5 | **Multi-Replica-Harness.** Compose-Profil: ≥ 2 api-Replicas + 1 Postgres + LB (z. B. nginx). | Stack startet; beide Replicas teilen den Store; Health grün. |
| 6 | **Scale-out-Lasttest (die R-26-Evidenz).** `smoke-load.sh` gegen den LB; Reconciliation muss über **alle** Replicas zählen (Readback bleibt korrekt, da gegen den geteilten Store); Messung: Durchsatz 1 vs. 2 vs. N Replicas, kein Verlust/Dup, `ingest_sequence`-Integrität. Optional: Multi-Tenant-Verteilung (R-26 b). | Verdict-Zahlen: horizontale Skalierung belegt, `persisted == accepted` global, 0 Duplikate. |
| 7 | **Doku/Closeout.** ADR-0006 von „Accepted" auf „belegt" referenzieren; `budgets.md` §7 um Scale-out-Datenpunkte; R-26 → 🟢; Lastenheft RAK-91-Patch; CHANGELOG. | `make docs-check`; R-26 aufgelöst mit Messwert. |

## 6. DoD

- [ ] Postgres-DDL aus `schema.yaml` generiert; `driverName`/DSN
  parametrisiert; SQLite-Pfad byte-stabil.
- [ ] `persistence/postgres`-Adapter implementiert die sechs Driven-Ports.
- [ ] Contract-Suite grün gegen SQLite **und** Postgres in CI.
- [ ] `ingest_sequence`-Monotonie / Cursor-Walk unter Concurrent-Writern
  per Test belegt (Crux-Risiko geschlossen).
- [ ] `MTRACE_PERSISTENCE=postgres` opt-in, Default unverändert `sqlite`.
- [ ] Multi-Replica-Compose-Profil (≥ 2 api + Postgres + LB) startbar.
- [ ] **Scale-out-Lasttest mit Verdict**: horizontale Durchsatz-
  Skalierung gemessen (1/2/N Replicas), `persisted == accepted` global,
  0 Duplikate über Replicas, `ingest_sequence` intakt — **R-26 c gelöst**.
- [ ] Lastenheft-Patch RAK-91 „defer" → „proceed, optional";
  [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md) als belegt
  referenziert; `budgets.md` §7 + CHANGELOG nachgetragen.

## 7. Risiken

- **R-neu (Sequenz-/Cursor-Ordering unter Concurrent-Writern)** — siehe
  §3 Crux. Wird vor Tranche 6 mit einem dedizierten Concurrent-Insert-
  Test geschlossen; bei Bedarf eigene `risks-backlog`-Zeile.
- **CI-Kosten**: doppelte Persistenz-Matrix (SQLite + Postgres) verlängert
  die Pipeline; Postgres-Tests laufen container-gebunden. Abwägung: nur
  die Persistenz-Suite doppeln, nicht die ganze `gates`-Kette.
- **Scope-Disziplin**: Postgres bleibt optionaler Adapter, kein Default-
  Wechsel; jede Versuchung zu „Postgres als neuer Standard" widerspricht
  [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md) §Grenzen.
