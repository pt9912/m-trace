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

- Postgres-Schema herstellen — **nicht** aus dem eingecheckten
  `schema.yaml` (V1-only, 5 Tabellen; `make schema-generate` regeneriert
  bewusst nur `V1__m_trace.sql` via `--target sqlite --version 1`, V2–V7
  sind handgepflegt). Live = **V1 + V2–V7** (13 Tabellen:
  `ingest_streams`/`project_token_generations`/`auth_issuance_counters`/… +
  V6/V7-ALTERs). **Bevorzugter Weg über d-migrate `v0.9.9`** (noch zu
  bauen): `schema reverse --source <sqlite-url>` introspiziert die live,
  V1–V7-migrierte SQLite in ein vollständiges neutrales Schema, `export
  flyway --target postgresql` erzeugt das PG-DDL — automatisiert, keine
  Hand-Portage, keine `schema.yaml`-Kollision. **Fallback**: V1–V7
  hand-portieren (für etwaige Reverse-Lücken). **Verworfen**:
  `schema.yaml`-Vollausbau + Generieren — `--version 1` ließe alle 13
  Tabellen in *eine* V1-Datei laufen → Kollision mit den handgepflegten
  V2–V7 (Neuschnitt der Migrationshistorie, unverhältnismäßig).
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
  **DB-autoritativ** werden. **`SELECT nextval('…')` ist der
  port-erhaltende Weg**: synchron, liefert `int64`, hält
  `driven.IngestSequencer.Next() int64` 1:1 → SQLite-/InMemory-Impl **und**
  die Call-Site (`register_playback_event_batch.go`, Pre-Assign
  `IngestSequence: u.sequencer.Next()`) bleiben unangetastet (minimaler
  Blast-Radius). **`IDENTITY` + `RETURNING` ist die Alternative —
  vermeiden**: der Wert ist erst *nach* dem INSERT bekannt, bricht den
  Pre-Assign-Flow und erzwingt einen Use-Case-/Domain-Event-/Drei-Adapter-
  Refactor. **Perf-Vorbehalt (Tranche 6):** per-Event-`nextval` ist ein
  DB-Roundtrip (heute `atomic.Add`, 0 I/O); bei Batch-Append @ ~3800+ ev/s
  konfundiert das die „lineare Skalierung". Mitigation **hinter dem Port**:
  Block-Allokation pro Batch (Sequence-`CACHE n`, ein `nextval`-Block, oder
  `generate_series(1,N)`) — der `Next()`-Vertrag bleibt. **R-28**: operativ
  zuerst (ohne ihn maskieren PK-Kollisionen alles).
- **Read-Side: Keyset-Pagination-Skip/Dup unter Concurrent-Writern (R-27).**
  Die Cursor-Pagination ([ADR-0004](../../adr/0004-cursor-strategy.md))
  ordnet **primär über `server_received_at ASC`**, dann `sequence_number`,
  und `ingest_sequence` nur als **finalen Tie-Breaker**
  (`event_repository.go`: `ORDER BY server_received_at, COALESCE(sequence_number,…),
  ingest_sequence`). `server_received_at` wird im App-Layer zur
  **Empfangszeit** gesetzt (`ServerReceivedAt: now`, dieselbe Stelle wie
  `Next()`), **nicht** zur Commit-Zeit. Das Skip-/Dup-Risiko entsteht damit,
  **sobald es nebenläufige Postgres-Writer gibt** (= Adapter existiert,
  Tranche 2) — primär über `server_received_at` und **unabhängig** von R-28:
  ein Writer, der ein Event mit `server_received_at = T-1` *nach* dem Reader
  committet, der schon an `T` vorbeipaginiert ist, wird übersprungen — auch
  bei perfekt monotonem `ingest_sequence`. **Mitigation**: `REPEATABLE READ`
  allein reicht **nicht** (Pagination = mehrere Queries über mehrere
  Snapshots); nur ein **commit-order-stabiles Wasserzeichen** (nicht über
  noch-nicht-sichtbare Commits hinauspaginieren) trägt cross-page. R-27
  hängt also **nicht** an R-28 — „R-28 zuerst" ist nur **operativ** (sonst
  maskieren PK-Kollisionen alles).
- **Schema-Herkunft (eingecheckter `schema.yaml` ist V1-only).**
  `schema.yaml` = V1-Baseline (5 Tabellen); Live = V1 + handgepflegte
  V2–V7 (13 Tabellen). **Auflösung über d-migrate `v0.9.9`** (noch zu
  bauen; aktueller Pin `DMIGRATE_IMAGE` = `0.9.5`): dessen Dev-Tree hat
  `schema reverse` (live-DB → neutrales Schema) **und** `--target
  postgresql` (`driver-postgresql`, e2e `E2ERoundTripPostgresTest`).
  Tranche 1 reversed die live, V1–V7-migrierte SQLite in ein vollständiges
  Schema und generiert daraus PG-DDL — **kein** Postgres-Support-Risiko,
  aber `v0.9.9` ist eine **externe Voraussetzung** (s. §2, Tranche 1, §6).
- **Transaktions-Isolation.** SQLite-`BEGIN IMMEDIATE` → Postgres-Default
  `READ COMMITTED` vs. `REPEATABLE READ`; der Batch-Append + Lifecycle-
  Tick muss dasselbe Verhalten wie der SQLite-Pfad zeigen (Contract-Suite
  als Wächter).
- **Dialekt-Differenzen.** Platzhalter (`?` → `$1`), `INTEGER PRIMARY KEY
  AUTOINCREMENT` → `BIGINT GENERATED ... AS IDENTITY`/`BIGSERIAL`,
  `INSERT ... ON CONFLICT`, Boolean (`INTEGER 0/1` → `boolean`), JSON-
  Spalten (`meta TEXT` → `text`/`jsonb`). Wird über das d-migrate-Target
  + adapter-lokale Query-Konstanten gekapselt.
- **Connection-Pooling + `max_connections` als Scale-out-Decke.** SQLite =
  ein File-Handle; Postgres braucht `pgxpool`-Pool-Sizing. **Reale
  Harness-Decke**: `N Replicas × pool_size ≤ max_connections` (Default 100)
  — sonst misst Tranche 6 das Connection-Limit statt der Store-Skalierung;
  ggf. `pgbouncer` davor.
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
- **Getrackt: R-28** (Write-Side-Sequencer-Redesign, Tranche 2) und **R-27**
  (Read-Side-Keyset-Skip/Dup, Tranche 3) — **beide unabhängig**, je
  DoD-blockierend vor dem Scale-out-Test (Tranche 6). R-28 wird nur
  **operativ zuerst** gebaut (sonst maskieren PK-Kollisionen R-27); keine
  kausale Abhängigkeit.

## 4. Tranchen

| Tranche | Inhalt | Gate |
| --- | --- | --- |
| 1 | **Vollständiges PG-Schema via d-migrate `v0.9.9`.** (0) **Voraussetzung**: d-migrate `v0.9.9` bauen + `DMIGRATE_IMAGE` `0.9.5`→`0.9.9` bumpen (externe Abhängigkeit, §6). (1) `schema reverse --source <sqlite-url>` der live, V1–V7-migrierten SQLite → vollständiges neutrales Schema (13 Tabellen); `export flyway --target postgresql` → PG-DDL. Fallback: V1–V7 hand-portieren. (2) `driverName`/DSN parametrisieren (`migrate.go:37` hartkodiert `sqlite`); Postgres-Container im Lab. | Frische PG-DB trägt **alle** Live-Tabellen/Spalten (V1–V7, inkl. V2/V4-Tabellen + V6/V7-Spalten), nicht nur die V1-Baseline; SQLite-Migrationspfad + `generated-drift-check` unverändert grün. |
| 2 | **Postgres-Adapter (6 Ports) + Sequencer-Redesign (R-28).** `persistence/postgres` für fünf Ports als Dialekt-Kapselung; der **`ingest_sequencer` ist ein Redesign, kein Spiegel**: DB-autoritativ via **`SELECT nextval`** — **port-erhaltend** (`Next() int64` unverändert → SQLite-/InMemory-Impl **und** Call-Site `register_playback_event_batch.go` bleiben). `IDENTITY`+`RETURNING` **vermeiden** (bricht den Pre-Assign-Flow). Gegen den per-Event-Roundtrip: Block-Allokation pro Batch **hinter dem Port**. | Sequencer DB-autoritativ via `nextval`, Port-Vertrag unverändert; SQLite-/InMemory-Sequencer grün; Batch-Block-Allokation gegen Roundtrip-Konfundierung. |
| 3 | **Port-Korrektheit gegen Postgres.** Contract-Suite (3 Ports: `Sessions`/`Events`/`Sequencer`) gegen Postgres grün; **plus** portierte Postgres-Tests für die drei Nicht-Contract-Ports (`project_token`/`srt_health`/`ingest_stream` — Dialekt: `ON CONFLICT`, Boolean, Rotation); **plus** ein Concurrent-Writer-Test, der **R-28** (kein Dup / keine PK-Kollision über N parallele Writer) **und** **R-27** belegt — letzterer explizit mit **out-of-order `server_received_at`-Commit** (ein spät committeter Writer mit früherem `server_received_at` wird vom Cursor-Walk *nicht* übersprungen), nicht nur Sequenz-Kollision. | Alle **sechs** Ports gegen Postgres getestet (3 via Contract-Suite + 3 portiert); R-28- + R-27-Test (inkl. out-of-order `server_received_at`) grün; SQLite-Pfad unverändert. |
| 4 | **Wiring + CI-Matrix.** `MTRACE_PERSISTENCE=postgres` + DSN in `main.go` (Default `sqlite` byte-stabil); CI fährt die Persistenz-Tests gegen beide Stores. | `MTRACE_PERSISTENCE=sqlite` unverändert; `=postgres` boot't + gleicher Smoke grün. |
| 5 | **Multi-Replica-Harness.** Compose-Profil: ≥ 2 api-Replicas + 1 Postgres + LB (z. B. nginx). Pool-Sizing so, dass `N × pool_size ≤ max_connections` (Default 100); ggf. `pgbouncer`. | Stack startet; beide Replicas teilen den Store; Health grün; `N × pool_size ≤ max_connections` eingehalten. |
| 6 | **Scale-out-Lasttest (die R-26-c-Evidenz).** `smoke-load.sh` gegen den LB. **Readback braucht einen Postgres-Zweig**: kein GLOB, kein geteiltes File-Volume → `psql`-`count(*)` mit `LIKE 'prefix-%'` (`_` escapen) als **eine** Query gegen den geteilten Store (sauberer als der SQLite-`--volumes-from`-GLOB-Hack aus R-25). Messung: Durchsatz 1 vs. 2 vs. N Replicas, kein Verlust/Dup, `ingest_sequence`-Integrität. Multi-Tenant-Teil (R-26 b) erst **nach** dem shared Ingest-Limiter sinnvoll — bis dahin ist `N × Capacity` (kein Fairness-Nachweis) das **vorhergesagte** Verhalten, kein Befund. | Verdict: horizontale Durchsatz-Skalierung belegt, `persisted == accepted` global, 0 Duplikate über Replicas. |
| 7 | **Doku/Closeout.** ADR-0006 von „Accepted" auf „belegt" referenzieren; `budgets.md` §7 um Scale-out-Datenpunkte; **R-26 c → gelöst** (b/Multi-Tenant bleibt offen, s. R-26 b); Lastenheft RAK-91-Patch; CHANGELOG. | `make docs-check`; R-26 c aufgelöst mit Messwert. |

## 5. DoD

- [ ] **Vollständiges PG-Schema (alle 13 Tabellen, V1–V7)** via d-migrate
  `v0.9.9` `schema reverse` (live-SQLite) + `export flyway --target
  postgresql` hergestellt (Hand-Portage nur Fallback); `DMIGRATE_IMAGE` auf
  `0.9.9` gebumpt; `driverName`/DSN parametrisiert; SQLite-Pfad +
  `generated-drift-check` unverändert grün.
- [ ] `persistence/postgres`-Adapter implementiert die sechs Driven-Ports;
  der **`ingest_sequencer` ist DB-autoritativ via `nextval`** (R-28,
  **port-erhaltend**: `Next() int64` + Call-Site unverändert; **nicht**
  `IDENTITY`+`RETURNING`), mit Batch-Block-Allokation gegen den per-Event-
  Roundtrip; SQLite-/InMemory-Sequencer unverändert grün.
- [ ] **Alle sechs Ports** gegen Postgres getestet: Contract-Suite (3
  Ports) **plus** portierte Postgres-Tests für `project_token`/
  `srt_health`/`ingest_stream`; grün gegen SQLite **und** Postgres in CI.
- [ ] Concurrent-Writer-Test belegt **R-28** (kein Dup / keine
  PK-Kollision über parallele Writer) **und** **R-27** (Cursor-Walk sieht
  jedes Event genau einmal, inkl. **out-of-order `server_received_at`-
  Commit**) — beide unabhängig, DoD-blockierend vor Tranche 6.
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
  muss DB-autoritativ werden — **`nextval` (port-erhaltend) bevorzugt**,
  `IDENTITY`+`RETURNING` vermeiden (bricht den Pre-Assign-Flow); Block-
  Allokation gegen den per-Event-Roundtrip. In
  [`risks-backlog.md`](../in-progress/risks-backlog.md), Tranche 2,
  DoD-blockierend.
- **R-27 — Read-Side-Keyset-Skip/Dup unter Concurrent-Writern**: primär
  über `server_received_at` (App-gesetzt, Cursor-Primärschlüssel),
  **unabhängig** von R-28 — entsteht ab dem ersten nebenläufigen PG-Writer.
  `REPEATABLE READ` allein reicht nicht (mehrere Snapshots über mehrere
  Queries); nur ein commit-order-stabiles Wasserzeichen trägt cross-page.
  DoD-blockierend, vor Tranche 6 mit einem out-of-order-`server_received_at`-
  Test zu schließen.
- **d-migrate `v0.9.9` ist externe Voraussetzung** (noch zu bauen;
  aktueller Pin `0.9.5`). Postgres-Support selbst ist **kein** Risiko mehr
  (verifiziert im d-migrate-Dev-Tree: `driver-postgresql`, `schema reverse`,
  e2e `--target postgresql`); das Risiko ist die **Verfügbarkeit/das Bauen**
  von `0.9.9` plus ein etwaiger **Reverse-Gap** (Objekt, das `schema reverse`
  nicht erfasst → gezielte Hand-Portage als Fallback). Tranche 1 hängt am
  `0.9.9`-Bump. Der Schema-Portage-Aufwand (V1–V7, 13 Tabellen) ist damit
  weitgehend automatisiert, nicht mehr „zweiten Dialekt an fertigen Anker
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
