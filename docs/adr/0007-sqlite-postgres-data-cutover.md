# 0007 — SQLite→Postgres-Datenmigration / Cutover (optional)

> **Status**: Proposed (Skizze / Stub — noch nicht entschieden)
> **Datum**: 2026-07-10
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: `spec/lastenheft.md` RAK-91 (Postgres „proceed, optional");
> [ADR-0005](0005-production-ops-backends.md) (Postgres deferred mit
> Triggern), [ADR-0006](0006-postgres-scaleout-adapter.md) (Runtime-Adapter);
> `docs/planning/in-progress/risks-backlog.md` R-26 / R-29;
> Roadmap-Anker „defer-**with-migration-seed**" ([`plan-0.15.0.md`](../planning/done/plan-0.15.0.md)
> Szenario A Tranche 5); Ausführung in
> [`plan-0.24.0-sqlite-postgres-cutover.md`](../planning/open/plan-0.24.0-sqlite-postgres-cutover.md).

## Kontext

[ADR-0006](0006-postgres-scaleout-adapter.md) liefert den **optionalen
Postgres-Runtime-Adapter** (Scale-out-Store). `plan-0.23.0-postgres-scaleout`
klammert die **Datenmigration** bestehender Läufe bewusst aus: *„Keine
automatische SQLite→Postgres-Datenmigration … Postgres startet mit frischem
Store."* Damit verliert ein bestehendes SQLite-Deployment beim Umstieg auf
Scale-out seine Historie — akzeptabel für den Erstnachweis (R-26 c), aber
nicht für einen echten Produktions-Cutover.

Der Roadmap-/Lastenheft-Stand hält Postgres schon länger als
**„defer-with-migration-seed"** (Szenario A, Tranche 5) — der Seed war also
immer mitgedacht, nur nicht ausgeführt.

Neue Erkenntnis (2026-07-10): d-migrate `v0.9.9` bietet nicht nur
Schema-Werkzeuge, sondern ein volles **Daten**-Migrations-Framework
(`data transfer`, `data profile`, `data export`/`import`). `data transfer`
migriert Zeilen **dialekt-übergreifend** SQLite→Postgres — batched
(`--chunk-size`), **inkrementell** (`--since-column`/`--since`), **idempotent**
(`--on-conflict abort|skip|update`) und **sequenz-getreu**
(`--sqlite-autoincrement-width 64`, konsistent mit dem in
[ADR-0006](0006-postgres-scaleout-adapter.md) gewählten 64-bit-BIGSERIAL).
Damit ist der in `plan-0.23.0` vertagte Backlog-Vorbehalt (BIGSERIAL-Sequence
startet bei 1 → bei Datenmigration Zählerstand-Erhalt nötig) **nativ lösbar**
statt Hand-Portage.

## Entscheidung

> **Vorschlag (noch offen):** Einen **optionalen, ops-/deploy-zeitigen**
> SQLite→Postgres-**Cutover** über d-migrate `data transfer` bereitstellen,
> in vier Phasen: **Bulk → inkrementell → Profile-Check → Switch**. Details in
> [`plan-0.24.0-sqlite-postgres-cutover.md`](../planning/open/plan-0.24.0-sqlite-postgres-cutover.md).

Kern der vorgeschlagenen Mechanik:

1. **Profile-Check** (`data profile`): Target-Typ-Kompatibilität + Quality-
   Warnings gegen die frische, drift-bewachte PG-Baseline
   (`migrations/postgres/V1__m_trace.sql`) prüfen, bevor Daten fließen.
2. **Bulk** (`data transfer`, `--chunk-size`, `--on-conflict abort`): die
   V1–V7-Tabellen einmalig übertragen; `--sqlite-autoincrement-width 64` hält
   `ingest_sequence`/`srt_health_samples.id` samt Sequenz-Zählerstand getreu.
3. **Inkrementell** (`--since-column ingest_sequence --since <watermark>`,
   `--on-conflict skip`): Delta seit dem Bulk-Watermark nachziehen → minimale
   Downtime (wiederholbar bis zum Cutover-Fenster).
4. **Switch**: Writer kurz quiescen, finales inkrementelles Nachziehen,
   `MTRACE_PERSISTENCE=postgres` schalten, Verifikation
   (Row-Counts/Watermark), Rollback = zurück auf SQLite (Postgres bleibt
   ephemer, solange nicht bestätigt).

Die JDK-freie API-Runtime ([ADR-0002](0002-persistence-store.md)) bleibt
unangetastet: d-migrate läuft als **ephemerer Ops-Container** (kein Runtime-
/Deploy-Image-Bloat), nicht im API-Prozess.

## Begründung

- **Nativer statt hand-portierter Pfad** — `data transfer` löst Sequenz-Erhalt,
  Batching und Idempotenz, die ein Eigenbau-Migrator alle selbst tragen müsste.
- **Konsistent mit dem Schema-Pfad** — dieselbe d-migrate-Version + dieselbe
  64-bit-Identity-Behandlung wie [ADR-0006](0006-postgres-scaleout-adapter.md)
  / `make schema-generate-postgres`.
- **Minimale Downtime** — Bulk + inkrementelles Nachziehen statt Stop-the-world.
- **Reversibel** — bis zum bestätigten Switch bleibt SQLite die Autorität.

## Nicht Entschieden / Grenzen

- **Watermark-Wahl** offen: `ingest_sequence` (app-monoton, aber R-28-Block-
  Allokation macht es connection-übergreifend non-monoton) vs. eine Commit-
  Zeit-Spalte (vgl. R-27). Der Cutover braucht ein **konsistentes** Delta-
  Signal — hängt an der R-27/R-28-Auflösung in `plan-0.23.0`.
- **Konsistenz während des Cutovers** (in-flight-Writes, out-of-order-Commits)
  ist der harte Teil — nicht in diesem Stub gelöst.
- **Multi-Tenant/Multi-Replica-Cutover** (mehrere Quell-DBs / geteilter
  Ziel-Store) nicht betrachtet.
- **Kein Zwang**: SQLite bleibt Default; der Cutover ist opt-in für
  Deployments, die auf Scale-out umsteigen wollen.

## What Ändert Sich

- Neuer opt-in **Ops-Pfad** (Skript/Make-Target, ephemerer d-migrate-Container)
  für den SQLite→Postgres-Cutover — kein Runtime-/API-Code-Impact.
- Der `plan-0.23.0`-Scope-Ausschluss „Postgres startet mit frischem Store" wird
  von einem harten Limit zu einer **Default-Option neben dem Cutover**.

## What Bleibt Unverändert

- API-Runtime JDK-frei ([ADR-0002](0002-persistence-store.md)); der
  eingebettete Go-Migrations-Runner (`internal/storage/migrate.go`) bleibt der
  Schema-Applier.
- SQLite bleibt Default-Store; das PG-Schema + der Drift-Check
  (`make schema-generate-postgres-check`) bleiben die Autorität für die
  Ziel-Struktur.
- [ADR-0006](0006-postgres-scaleout-adapter.md) (Runtime-Adapter) und die
  Tranche-1-Bausteine bleiben unberührt.
