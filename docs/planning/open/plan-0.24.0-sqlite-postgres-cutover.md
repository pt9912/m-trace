# Implementation Plan (Skizze) — `0.24.0` SQLite→Postgres-Cutover

> **Status**: 📝 **Skizze / Draft** — noch nicht gestartet, noch nicht
> committed als Scope. Folge-Kandidat zu
> [`plan-0.23.0-postgres-scaleout`](../done/plan-0.23.0-postgres-scaleout.md)
> (Runtime-Adapter + Scale-out-Evidenz). Die Versionsnummer `0.24.0` ist
> provisorisch.
>
> **Bezug**: [ADR-0007](../../adr/0007-sqlite-postgres-data-cutover.md)
> (Cutover-Entscheidung, Proposed); [ADR-0006](../../adr/0006-postgres-scaleout-adapter.md)
> (Runtime-Adapter); RAK-91; Roadmap-Anker „defer-**with-migration-seed**"
> ([`plan-0.15.0.md`](../done/plan-0.15.0.md) Szenario A Tranche 5);
> R-26 / **R-29** in [`risks-backlog.md`](../in-progress/risks-backlog.md).

## 1. Ziel

Ein **bestehendes SQLite-Deployment** soll seine Historie **nach Postgres
migrieren** können, statt beim Umstieg auf Scale-out bei null anzufangen.
`plan-0.23.0` klammert das aus („Postgres startet mit frischem Store"); dieser
Plan liefert den **optionalen Cutover** — nativ über d-migrate `data transfer`,
nicht als Eigenbau-Migrator.

Nicht-Ziel: SQLite ablösen. SQLite bleibt Default; der Cutover ist opt-in für
Deployments, die aktiv auf den Postgres-Scale-out-Store wechseln.

## 2. Scope / Abgrenzung

**In Scope:**

- Ops-/deploy-zeitiger **Cutover-Ablauf** (Skript + `make`-Target, ephemerer
  d-migrate-Container) für SQLite→Postgres.
- **Sequenz-/Identity-Erhalt** (`--sqlite-autoincrement-width 64`) für
  `ingest_sequence` + `srt_health_samples.id` — der PG-Sequenz-Zählerstand
  muss den SQLite-`MAX` fortsetzen, sonst kollidiert der erste DB-vergebene
  Wert.
- **Verifikation** nach dem Transfer (Row-Counts je Tabelle, Watermark-
  Konsistenz, Profile-Report) und ein **Rollback-Pfad** (zurück auf SQLite).

**Nicht in Scope (dieser Tranche):**

- Zero-Downtime-Live-Replikation (CDC/logische Replikation) — hier reicht
  Bulk + kurzes inkrementelles Nachziehen mit definiertem Cutover-Fenster.
- Multi-Tenant-/Multi-Replica-Cutover (mehrere Quell-DBs → geteilter Store).
- Automatischer Trigger — der Cutover wird manuell/operator-initiiert.

## 3. Vorgehen (vier Phasen)

Reihenfolge **Bulk → inkrementell → Profile-Check → Switch**; der
Profile-Check läuft als Pre-Flight **vor** dem Bulk.

| Phase | d-migrate | Zweck |
| ----- | --------- | ----- |
| **0. Profile-Check** | `data profile --source sqlite://…` | Target-Typ-Kompatibilität + Quality-Warnings gegen die drift-bewachte PG-Baseline; Pre-Flight-Abbruch bei Inkompatibilität. |
| **1. Bulk** | `data transfer --source sqlite://… --target postgres://… --sqlite-autoincrement-width 64 --on-conflict abort --chunk-size N` | V1–V7-Tabellen einmalig übertragen (Schema via `plan-0.23.0`-DDL bereits angelegt); Sequenz-Zählerstände getreu. Watermark (max `ingest_sequence`) festhalten. |
| **2. Inkrementell** | `data transfer --since-column ingest_sequence --since <watermark> --on-conflict skip` | Delta seit Bulk nachziehen; wiederholbar bis zum Cutover-Fenster → minimale Downtime. |
| **3. Switch** | — | Writer kurz quiescen → finales inkrementelles Nachziehen → `MTRACE_PERSISTENCE=postgres` schalten → Verifikation → (Rollback: zurück auf SQLite). |

## 4. Offene Fragen / Crux

- **Watermark-Signal** (Crux, hängt an `plan-0.23.0` R-27/R-28): `ingest_sequence`
  ist app-monoton, aber die R-28-Block-Allokation macht es connection-
  übergreifend non-monoton; ein inkrementelles `--since` braucht ein
  **lückenlos-konsistentes** Delta-Signal. Möglicherweise Commit-Zeit-Spalte
  (vgl. R-27) statt `ingest_sequence`. **Blocker bis R-27/R-28 stehen.**
- **In-flight-Konsistenz** während des Switch-Fensters (out-of-order-Commits
  dürfen nicht zwischen finalem Delta und Umschaltung verloren gehen).
- **Verifikations-Tiefe**: Row-Count je Tabelle reicht als Smoke; eine
  inhaltliche Stichprobe (Hash/Checksumme je Chunk) wäre stärker.
- **`data profile`-Toleranz**: welche Quality-Warnings sind Abbruch- vs.
  Info-Kriterium?

## 5. Akzeptanz / DoD (Skizze)

- Frische, per `plan-0.23.0`-DDL angelegte PG-DB trägt nach dem Cutover
  **alle** SQLite-Zeilen (Row-Count-Parität je Tabelle).
- `ingest_sequence`-Sequenz in PG setzt den SQLite-`MAX` fort (erster neuer
  DB-vergebener Wert kollidiert nicht).
- Inkrementeller Re-Run ist **idempotent** (`--on-conflict skip` → keine
  Duplikate).
- Rollback zurück auf SQLite ist dokumentiert und getestet.
- Der Cutover-Ablauf ist als opt-in-Smoke reproduzierbar (analog
  [`smoke-pg-lab`](../../../scripts/smoke-pg-lab.sh)); SQLite-Default-Pfad +
  `make gates` unverändert grün.

## 6. Voraussetzungen / Sequenzierung

- **Setzt `plan-0.23.0` voraus**: PG-Schema + Runtime-Adapter + der
  R-27/R-28-Auflösung (Commit-Order-Signal), an der das Watermark hängt.
- d-migrate `data transfer` ist ab `v0.9.9` verfügbar (bereits gepinnt).
- Erst nach einem konkreten Cutover-Bedarf zu starten (R-29-Trigger), nicht
  vorgezogen.
