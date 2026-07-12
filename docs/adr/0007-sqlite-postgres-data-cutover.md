# 0007 — SQLite→Postgres-Datenmigration / Cutover (optional)

> **Status**: **Accepted + GELIEFERT** (2026-07-12) — die vier Phasen
> (Profile-Check → Bulk → inkrementell → Switch) sind als opt-in-Ops-Werkzeug
> `scripts/cutover-sqlite-postgres.sh` (`make cutover`) implementiert, alle
> code-reviewt, `make smoke-cutover` (8 Cases) grün, Operator-Runbook
> [`docs/ops/postgres-cutover.md`](../ops/postgres-cutover.md). Watermark
> entschieden (`ingest_sequence` auf der Single-Instance-Quelle); mutable
> Tabellen werden im quiescten Switch per `--on-conflict update` reconciled.
> Ausführung tranchiert in
> [`plan-0.24.0-sqlite-postgres-cutover.md`](../planning/in-progress/plan-0.24.0-sqlite-postgres-cutover.md);
> `R-29` → 🟢. **Kein eigenes 0.24.0-Release** (Owner-Entscheidung: reine
> Ops-Tooling ohne Runtime-/Package-Änderung) — rollt in den nächsten Release.
> Zuvor: Proposed (2026-07-10), Accepted (2026-07-12).
> **Datum**: 2026-07-10 (Proposed) · 2026-07-12 (Accepted + geliefert)
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)
> **Bezug**: `spec/lastenheft.md` RAK-91 (Postgres „proceed, optional");
> [ADR-0005](0005-production-ops-backends.md) (Postgres deferred mit
> Triggern), [ADR-0006](0006-postgres-scaleout-adapter.md) (Runtime-Adapter);
> `docs/planning/in-progress/risks-backlog.md` R-26 / R-29;
> Roadmap-Anker „defer-**with-migration-seed**" ([`plan-0.15.0.md`](../planning/done/plan-0.15.0.md)
> Szenario A Tranche 5); Ausführung in
> [`plan-0.24.0-sqlite-postgres-cutover.md`](../planning/in-progress/plan-0.24.0-sqlite-postgres-cutover.md).

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

> **Entscheidung (Accepted 2026-07-12):** Einen **optionalen, ops-/deploy-
> zeitigen** SQLite→Postgres-**Cutover** über d-migrate `data transfer`
> bereitstellen, in vier Phasen: **Profile-Check → Bulk → inkrementell →
> Switch** (Profile-Check ist Pre-Flight, läuft zuerst). Details in
> [`plan-0.24.0-sqlite-postgres-cutover.md`](../planning/in-progress/plan-0.24.0-sqlite-postgres-cutover.md).

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

## Entschieden 2026-07-12 (vormals offen)

- **Watermark-Wahl entschieden**: `ingest_sequence` bleibt das `--since`-Delta-
  Signal — **keine** Commit-Zeit-Spalte am Cutover nötig. Die R-28-Non-Monotonie
  ist eine **Ziel**-Eigenschaft (Multi-Replica-PG); der Cutover liest per
  `--since-column` aus der **Quelle**, und die ist Single-Instance-SQLite →
  dort ist `ingest_sequence` monoton. Details in
  [`plan-0.24.0`](../planning/in-progress/plan-0.24.0-sqlite-postgres-cutover.md) §4.
- **Konsistenz während des Cutovers gelöst**: der Rest-Effekt (Zuweisung ≠
  Commit-Order auf einer Instanz) wird durch **Writer-Quiesce vor dem finalen
  Delta** + **`--on-conflict skip`-Idempotenz mit konservativem Lookback**
  getragen; die Vollständigkeits-Garantie liegt im quiescten finalen Lauf, nicht
  in den Zwischenläufen.

## Nicht Entschieden / Grenzen

- **`data profile`-Toleranz** (Abbruch- vs. Info-Warnings) und **Verifikations-
  Tiefe** (Row-Count vs. inhaltliche Stichprobe) bleiben Bau-Detail der
  plan-0.24.0-Tranchen (§8).
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
