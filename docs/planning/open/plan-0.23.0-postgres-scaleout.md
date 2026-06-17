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

- Postgres-Schema herstellen — **nicht** als One-Shot aus `schema.yaml`:
  der neutrale `schema.yaml` deckt nur die **V1-Baseline** (5 Tabellen),
  `make schema-generate` regeneriert auch nur `V1__m_trace.sql` (`--target
  sqlite`). Das Live-Schema ist **V1 + handgepflegte V2–V7** (~8 weitere
  Tabellen wie `ingest_streams`/`project_token_generations`/
  `auth_issuance_counters` + V6/V7-Spalten-ALTERs auf `playback_events`/
  `stream_sessions`). Tranche 1 entscheidet zwischen **V1–V7-Portage** nach
  Postgres-Dialekt **oder** `schema.yaml`-Vollausbau + Generierung (dann
  den V1-only-`generated-drift-check` reconcilen); vorab ist `d-migrate
  --target postgres` **nachzuweisen** (heute nur `--target sqlite` belegt).
  `driverName`/DSN parametrisiert.
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

Die Architektur ist hexagonal sauber, aber **nicht** „fertig zum
Andocken eines zweiten Dialekts" — zwei tragende Annahmen tragen so nicht
(erste zwei Bullets, je eine Risiko-Zeile). Die nicht-trivialen Punkte:

- **Write-Side: der `ingest_sequencer` ist heute ein In-Process-RAM-
  Counter — Multi-Replica-Blocker (R-28).** `NewIngestSequencer` seedet
  einen `atomic.Int64` einmalig aus `SELECT MAX(ingest_sequence)`,
  `Next()` = `counter.Add(1)`, der Adapter inserted den **app-zugewiesenen**
  Wert explizit (`AUTOINCREMENT` nur Defense-in-Depth). Zwei Replicas
  seeden aus demselben MAX und vergeben **dieselben** Werte →
  PK-Kollisionen / doppelte `ingest_sequence` genau im Multi-Replica-
  Betrieb (= wofür der Plan existiert). Der Postgres-Sequencer muss
  **DB-autoritativ** werden (`nextval` einer Postgres-Sequence bzw.
  `IDENTITY` + `RETURNING`) — ein **Port-/Implementierungs-Redesign**,
  nicht „dieselbe Implementierung mit Dialekt-Kapselung", und berührt ggf.
  den `driven.IngestSequencer`-Vertrag (SQLite-/InMemory-Impl müssen grün
  bleiben). **R-28**, Voraussetzung für überhaupt funktionierendes
  Multi-Replica.
- **Read-Side: Cursor-Ordering unter Concurrent-Writern (R-27).** *Sobald*
  der Sequencer DB-autoritativ ist, wird der Wert zur Insert-/Commit-Zeit
  vergeben → **Commit-Reihenfolge ≠ Sequence-Reihenfolge** (Lücken,
  out-of-order-Sichtbarkeit). Die Keyset-Cursor-Pagination
  ([ADR-0004](../../adr/0004-cursor-strategy.md)) ordnet über
  `(server_received_at, sequence_number, ingest_sequence)`; es ist zu
  belegen, dass sie unter Concurrent-Insert kein Event überspringt oder
  dupliziert (Snapshot-/`REPEATABLE READ`-Isolation oder commit-order-
  stabile Quelle). **R-27 setzt das R-28-Redesign voraus** — erst danach
  ist der Read-Side-Test echt.
- **Schema-Herkunft (kein fertiger Anker).** `schema.yaml` = V1-Baseline
  (5 Tabellen); Live = V1 + handgepflegte V2–V7. Das Postgres-Schema
  entsteht durch Portage der Migrationshistorie, nicht aus einem Anker —
  und `d-migrate --target postgres` ist erst nachzuweisen (s. §2, Tranche 1).
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
- **Getrackt: R-28** (Write-Side-Sequencer-Redesign, Tranche 2,
  Voraussetzung für Multi-Replica) **→ R-27** (Read-Side-Cursor-Ordering,
  wird erst nach R-28 echt) — beide DoD-blockierend vor dem Scale-out-Test
  (Tranche 6).

## 4. Tranchen

| Tranche | Inhalt | Gate |
| --- | --- | --- |
| 1 | **Schema + Migrations-Target (mehr als „Anker andocken").** (0) **Spike**: `d-migrate --target postgres` nachweisen — falls unsupported, Fallback (handgeschriebenes PG-DDL). (1) Vollständiges PG-Schema herstellen: V1–V7 nach Postgres-Dialekt portieren **oder** `schema.yaml` auf den V1–V7-Vollstand heben + generieren und den V1-only-`generated-drift-check` reconcilen. (2) `driverName`/DSN parametrisieren (heute `migrate.go:37` hartkodiert `sqlite`); Postgres-Container im Lab. | `--target postgres` nachgewiesen (oder Fallback dokumentiert); frische PG-DB trägt **alle** Live-Tabellen/Spalten (V1–V7, inkl. V2/V4-Tabellen + V6/V7-Spalten), nicht nur die V1-Baseline; SQLite-Migrationspfad + drift-check unverändert grün. |
| 2 | **Postgres-Adapter (6 Ports) + Sequencer-Redesign.** `persistence/postgres` für fünf Ports als Dialekt-Kapselung; der **`ingest_sequencer` ist ein Redesign, kein Spiegel** (R-28): DB-autoritativ (`nextval`/`IDENTITY`+`RETURNING`) statt In-Process-`MAX`+`atomic.Add`. `driven.IngestSequencer`-Vertrag ggf. anpassen — SQLite-/InMemory-Impl müssen grün bleiben. | Adapter kompiliert; fünf Ports als Dialekt-Kapselung; Sequencer DB-autoritativ; SQLite-/InMemory-Sequencer unverändert grün. |
| 3 | **Port-Korrektheit gegen Postgres.** Contract-Suite (3 Ports: `Sessions`/`Events`/`Sequencer`) gegen Postgres grün; **plus** portierte Postgres-Tests für die drei Nicht-Contract-Ports (`project_token`/`srt_health`/`ingest_stream` — Dialekt: `ON CONFLICT`, Boolean, Rotation); **plus** ein Concurrent-Writer-Test, der **R-28** (kein Dup / keine PK-Kollision über N parallele Writer) **und** **R-27** (Cursor-Walk sieht jedes Event genau einmal) belegt. | Alle **sechs** Ports gegen Postgres getestet (3 via Contract-Suite + 3 portiert); R-28- + R-27-Test grün; SQLite-Pfad unverändert. |
| 4 | **Wiring + CI-Matrix.** `MTRACE_PERSISTENCE=postgres` + DSN in `main.go` (Default `sqlite` byte-stabil); CI fährt die Persistenz-Tests gegen beide Stores. | `MTRACE_PERSISTENCE=sqlite` unverändert; `=postgres` boot't + gleicher Smoke grün. |
| 5 | **Multi-Replica-Harness.** Compose-Profil: ≥ 2 api-Replicas + 1 Postgres + LB (z. B. nginx). | Stack startet; beide Replicas teilen den Store; Health grün. |
| 6 | **Scale-out-Lasttest (die R-26-c-Evidenz).** `smoke-load.sh` gegen den LB. **Readback braucht einen Postgres-Zweig**: kein GLOB, kein geteiltes File-Volume → `psql`-`count(*)` mit `LIKE 'prefix-%'` (`_` escapen) als **eine** Query gegen den geteilten Store (sauberer als der SQLite-`--volumes-from`-GLOB-Hack aus R-25). Messung: Durchsatz 1 vs. 2 vs. N Replicas, kein Verlust/Dup, `ingest_sequence`-Integrität. Multi-Tenant-Teil (R-26 b) erst **nach** dem shared Ingest-Limiter sinnvoll — bis dahin ist `N × Capacity` (kein Fairness-Nachweis) das **vorhergesagte** Verhalten, kein Befund. | Verdict: horizontale Durchsatz-Skalierung belegt, `persisted == accepted` global, 0 Duplikate über Replicas. |
| 7 | **Doku/Closeout.** ADR-0006 von „Accepted" auf „belegt" referenzieren; `budgets.md` §7 um Scale-out-Datenpunkte; **R-26 c → gelöst** (b/Multi-Tenant bleibt offen, s. R-26 b); Lastenheft RAK-91-Patch; CHANGELOG. | `make docs-check`; R-26 c aufgelöst mit Messwert. |

## 5. DoD

- [ ] **Vollständiges PG-Schema (V1–V7)** hergestellt (V1–V7-Portage oder
  `schema.yaml`-Vollausbau); `d-migrate --target postgres` nachgewiesen
  oder Fallback dokumentiert; `driverName`/DSN parametrisiert; SQLite-Pfad
  + `generated-drift-check` unverändert grün.
- [ ] `persistence/postgres`-Adapter implementiert die sechs Driven-Ports;
  der **`ingest_sequencer` ist DB-autoritativ** (R-28, `nextval`/`IDENTITY`
  statt In-Process-`MAX`+`atomic.Add`), SQLite-/InMemory-Sequencer
  unverändert grün.
- [ ] **Alle sechs Ports** gegen Postgres getestet: Contract-Suite (3
  Ports) **plus** portierte Postgres-Tests für `project_token`/
  `srt_health`/`ingest_stream`; grün gegen SQLite **und** Postgres in CI.
- [ ] Concurrent-Writer-Test belegt **R-28** (kein Dup / keine
  PK-Kollision über parallele Writer) **und** **R-27** (Cursor-Walk sieht
  jedes Event genau einmal) — beide DoD-blockierend vor Tranche 6.
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

- **R-28 — Write-Side-Sequencer-Redesign** (Voraussetzung für überhaupt
  funktionierendes Multi-Replica): der `ingest_sequencer` ist heute
  In-Process (`MAX`-Seed + `atomic.Add`) → PK-Kollisionen über Replicas;
  muss DB-autoritativ werden (`nextval`/`IDENTITY`). Port-Redesign, nicht
  Dialekt-Kapselung. In
  [`risks-backlog.md`](../in-progress/risks-backlog.md), Tranche 2,
  DoD-blockierend.
- **R-27 — Read-Side-Cursor-Ordering unter Concurrent-Writern**: wird
  *nach* R-28 echt (DB-autoritativer Wert erst zur Insert-/Commit-Zeit,
  Commit-Reihenfolge ≠ Sequence-Reihenfolge). DoD-blockierend, vor
  Tranche 6 mit einem Concurrent-Insert-/Cursor-Walk-Test zu schließen.
- **`d-migrate --target postgres` unverifiziert**: jeder bestehende
  Aufruf ist `--target sqlite`. Kann 0.9.5 kein Postgres-Target, braucht
  die Generierung einen Fallback (handgeschriebenes PG-DDL) — Tranche 1
  beginnt mit diesem Nachweis, bevor das Gate sich festlegt.
- **Schema-Portage-Aufwand**: `schema.yaml` ist V1-only; das volle Schema
  (V1–V7, ~8 zusätzliche Tabellen + Spalten-ALTERs) ist die eigentliche
  Tranche-1-Arbeit, nicht „einen zweiten Dialekt an einen fertigen Anker
  andocken".
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
