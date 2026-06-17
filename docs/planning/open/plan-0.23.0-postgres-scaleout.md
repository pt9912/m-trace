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
- Korrektheit gegen den Port-Vertrag: die adapter-agnostische
  Contract-Suite (`apps/api/adapters/driven/persistence/contract`) deckt
  heute **drei** der sechs Ports (`Sessions`, `Events`, `Sequencer`); die
  anderen drei (`project_token`, `srt_health`, `ingest_stream`) bekommen
  **portierte Postgres-Tests** aus den heutigen adapter-lokalen
  SQLite-Tests (echte Dialekt-Unterschiede: `ON CONFLICT`, Boolean,
  Token-Rotation). Optional: Contract-Suite auf sechs Ports erweitern.
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
- **Zweiter Single-Instance-Block (außerhalb des Stores).** Der
  Per-Projekt-Ingest-Limiter ist in-process
  (`ratelimit.NewTokenBucketRateLimiter`, `main.go`; keine Redis-Variante —
  nur Origin/Issuance haben Redis, R-22/R-17). Hinter N Replicas →
  effektive Per-Projekt-Decke `N × Capacity`, Fairness **nicht** repliken-
  übergreifend. Für den R-26-c-Nachweis (Durchsatz/kein Verlust, global
  gegen den Store gezählt) irrelevant; für den **Multi-Tenant-Fairness**-
  Teil (R-26 b, Tranche 6 optional) ist ein shared (Redis) Ingest-Limiter
  **Voraussetzung** — der Limiter-Umbau ist R-26-b-Scope, hier nur als
  Vorbehalt festgehalten (Fairness-Bruch wird **vorhergesagt**, nicht
  „entdeckt").
- **Crux-Risiko ist getrackt als R-27** (Sequenz-/Cursor-Ordering unter
  Concurrent-Writern, s. erster Bullet) — DoD-blockierend für Tranche 3.

## 4. Tranchen

| Tranche | Inhalt | Gate |
| --- | --- | --- |
| 1 | **Schema + Migrations-Target.** d-migrate Postgres-DDL aus `schema.yaml`; `driverName`/DSN parametrisieren (heute `migrate.go` hartkodiert `sqlite`); Postgres-Container im Lab. | `make`-Target migriert eine frische Postgres-DB fehlerfrei; Schema-Parität zu SQLite (gleiche Tabellen/Indizes aus einem Anker). |
| 2 | **Postgres-Adapter (6 Ports).** `persistence/postgres` mit denselben Driven-Port-Implementierungen; Dialekt-Kapselung. | Adapter kompiliert, alle sechs Ports implementiert. |
| 3 | **Port-Korrektheit gegen Postgres.** Contract-Suite (3 Ports: `Sessions`/`Events`/`Sequencer`) gegen Postgres grün; **plus** portierte Postgres-Tests für die drei Nicht-Contract-Ports (`project_token`/`srt_health`/`ingest_stream` — Dialekt: `ON CONFLICT`, Boolean, Rotation); **plus** der Concurrent-Writer-Sequenz-/Cursor-Walk-Test aus **R-27**. | Alle **sechs** Ports gegen Postgres getestet (3 via Contract-Suite + 3 portiert); R-27-Test grün; SQLite-Pfad unverändert. |
| 4 | **Wiring + CI-Matrix.** `MTRACE_PERSISTENCE=postgres` + DSN in `main.go` (Default `sqlite` byte-stabil); CI fährt die Persistenz-Tests gegen beide Stores. | `MTRACE_PERSISTENCE=sqlite` unverändert; `=postgres` boot't + gleicher Smoke grün. |
| 5 | **Multi-Replica-Harness.** Compose-Profil: ≥ 2 api-Replicas + 1 Postgres + LB (z. B. nginx). | Stack startet; beide Replicas teilen den Store; Health grün. |
| 6 | **Scale-out-Lasttest (die R-26-c-Evidenz).** `smoke-load.sh` gegen den LB. **Readback braucht einen Postgres-Zweig**: kein GLOB, kein geteiltes File-Volume → `psql`-`count(*)` mit `LIKE 'prefix-%'` (`_` escapen) als **eine** Query gegen den geteilten Store (sauberer als der SQLite-`--volumes-from`-GLOB-Hack aus R-25). Messung: Durchsatz 1 vs. 2 vs. N Replicas, kein Verlust/Dup, `ingest_sequence`-Integrität. Multi-Tenant-Teil (R-26 b) erst **nach** dem shared Ingest-Limiter sinnvoll — bis dahin ist `N × Capacity` (kein Fairness-Nachweis) das **vorhergesagte** Verhalten, kein Befund. | Verdict: horizontale Durchsatz-Skalierung belegt, `persisted == accepted` global, 0 Duplikate über Replicas. |
| 7 | **Doku/Closeout.** ADR-0006 von „Accepted" auf „belegt" referenzieren; `budgets.md` §7 um Scale-out-Datenpunkte; **R-26 c → gelöst** (b/Multi-Tenant bleibt offen, s. R-26 b); Lastenheft RAK-91-Patch; CHANGELOG. | `make docs-check`; R-26 c aufgelöst mit Messwert. |

## 5. DoD

- [ ] Postgres-DDL aus `schema.yaml` generiert; `driverName`/DSN
  parametrisiert; SQLite-Pfad byte-stabil.
- [ ] `persistence/postgres`-Adapter implementiert die sechs Driven-Ports.
- [ ] **Alle sechs Ports** gegen Postgres getestet: Contract-Suite (3
  Ports) **plus** portierte Postgres-Tests für `project_token`/
  `srt_health`/`ingest_stream`; grün gegen SQLite **und** Postgres in CI.
- [ ] `ingest_sequence`-Monotonie / Cursor-Walk unter Concurrent-Writern
  per Test belegt — **R-27 geschlossen** (DoD-blockierend für Tranche 3).
- [ ] `MTRACE_PERSISTENCE=postgres` opt-in, Default unverändert `sqlite`.
- [ ] Multi-Replica-Compose-Profil (≥ 2 api + Postgres + LB) startbar.
- [ ] **Scale-out-Lasttest mit Verdict**: horizontale Durchsatz-
  Skalierung gemessen (1/2/N Replicas), `persisted == accepted` global,
  0 Duplikate über Replicas, `ingest_sequence` intakt — **R-26 c gelöst**.
  (Multi-Tenant-Fairness, R-26 b, bleibt offen bis shared Ingest-Limiter.)
- [ ] Lastenheft-Patch RAK-91 „defer" → „proceed, optional" (**Variante B
  beim Spec-Edit: nur Kennungen, kein Plan-/§-Ref im Lastenheft**);
  [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md) als belegt
  referenziert; `budgets.md` §7 + CHANGELOG nachgetragen.

## 6. Risiken

- **R-27 — Sequenz-/Cursor-Ordering unter Concurrent-Writern** (das
  zentrale technische Risiko, siehe §3 Crux + DoD): in
  [`risks-backlog.md`](../in-progress/risks-backlog.md) angelegt,
  DoD-blockierend für Tranche 3, vor Tranche 6 mit einem dedizierten
  Concurrent-Insert-Test zu schließen.
- **In-Process-Ingest-Limiter** (siehe §3): Multi-Tenant-Fairness über
  Replicas (R-26 b) braucht einen shared (Redis) Ingest-Limiter; ohne ihn
  ist die effektive Per-Projekt-Decke `N × Capacity`. Gescopt in die
  R-26-b-Arbeit, **nicht** in 0.23.0 — hier nur Vorbehalt für den
  Multi-Tenant-Teil von Tranche 6.
- **CI-Kosten**: doppelte Persistenz-Matrix (SQLite + Postgres) verlängert
  die Pipeline; Postgres-Tests laufen container-gebunden. Abwägung: nur
  die Persistenz-Suite (Contract + die drei portierten Tests) gegen beide
  Stores doppeln, **nicht** die ganze `gates`-Kette.
- **Scope-Disziplin**: Postgres bleibt optionaler Adapter, kein Default-
  Wechsel; jede Versuchung zu „Postgres als neuer Standard" widerspricht
  [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md) §Grenzen.
