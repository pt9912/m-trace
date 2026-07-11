# Implementation Plan — `schema.yaml`-Refold (rolling-V1-Rekonsolidierung)

> **Status**: ✅ **Abgeschlossen (2026-07-11).** Der Tranche-1-Trockenlauf fand
> einen d-migrate-SQLite-Bug (`export flyway --target sqlite` verlor `NOT NULL`
> auf PRIMARY-KEY-Spalten; SQLites `PRIMARY KEY` ≠ NOT NULL) — **gefixt in
> d-migrate `0.9.10`** (Owner, Modell C3), `DMIGRATE_IMAGE` gebumpt. Danach alle
> Tranchen durch: **T1** `schema.yaml` = 13 Tabellen (`schema-validate` grün, 0
> Warnungen), Struktur-Diff `fresh(Einzel-V1)` vs `live(V1–V7)` == 0 (NOT NULL
> auf allen PK-Spalten erhalten; einziger Rest = kosmetische Auto-Index-
> Nummerierung, referenzlos). **T2** V1 regeneriert (13 Tab.), V2–V7 gelöscht,
> Fresh-Start-Test 7→1; Storage-/SQLite-/Auth-Tests grün. **T3** PG-DDL kommt
> jetzt via `export flyway --target postgresql --source schema.yaml`
> (`generate-postgres-schema.sh` umgestellt); PG-V1 ändert sich nur im Versions-
> Stempel; `smoke-pg-lab` grün. **T4** ADR-0002 §8.2 amendiert (Privileg
> reaktiviert, Begründung dokumentiert), CHANGELOG. Der d-migrate-Bug (Ursache +
> Fix-Ort) ist im [ADR-0002](../../adr/0002-persistence-store.md) §8.2 Amendment
> zusammengefasst.
>
> Stellt `schema.yaml` als
> **Single-Source-of-Truth für den vollen Schema-Stand** wieder her.
> Heute deckt `schema.yaml` nur die V1-Baseline (5 Tabellen); der Live-Store
> ist V1 + hand-gepflegte V2–V7 (13 Tabellen), und `generated-drift-check`
> bewacht nur V1↔`schema.yaml` → **V2–V7 sind drift-ungewacht**. Umsetzung
> als **Modell C** (rolling-V1-Rekonsolidierung, Muster von plan-0.8.5
> Tranche 2, Commit `dc0c705`): `schema.yaml` auf 13 Tabellen, **ein** V1
> daraus regenerieren, V2–V7 löschen, künftig `schema.yaml`-Edit → V1 neu
> generieren.
>
> **Sicherheit geprüft**: Der Apply-Runner (`internal/storage/migrate.go`)
> hat **keine Checksum-Validierung** — `schema_migrations` ist nur
> `(version INTEGER PRIMARY KEY, applied_at TEXT, dirty INTEGER)`,
> `appliedVersions` keyt rein auf die Versionsnummer. Bestehende
> Deployments (V1–V7 applied) **überspringen** das neue 13-Tabellen-V1
> (kein Re-Apply, kein Mismatch); frische DBs wenden das eine V1 an. Beide
> Pfade konvergieren zum selben 13-Tabellen-Endzustand.
>
> **Entscheidungen (2026-07-11)**: Modell C (gegen A=history-erhaltend +
> `schema migrate`, B=nur-Spiegel); Constraint-Namen **generisch `fk_0`
> aus dem Reverse akzeptiert** (kein semantisches Nachziehen);
> Reihenfolge **nach** Postgres-Tranche-2-Port-6/6 (erledigt, `e4e2011`).
>
> **Bezug**: [ADR-0002](../../adr/0002-persistence-store.md) §8.2
> (rolling-V1-Strategie + Pre-Production-Privileg — **wird amendiert**,
> weil das Privileg mit den v0.22.x-Releases formal abgelaufen ist);
> `plan-0.23.0-postgres-scaleout` (das PG-DDL wird in Tranche 3 auf
> `schema.yaml`-Provenienz umgestellt); `generate-postgres-schema.sh`
> (die Reverse-Pipeline, die Tranche 1/2 wiederverwenden).

## 1. Ziel

`schema.yaml` soll wieder **führen**: die volle 13-Tabellen-Wahrheit
deklarativ beschreiben **und** die Quelle sein, aus der das Baseline-DDL
(SQLite **und** Postgres) entsteht — nicht ein V1-Rest, an dem die
hand-gepflegten V2–V7 vorbeilaufen und den kein Gate gegen den Live-Stand
prüft.

## 2. Scope / Abgrenzung

**In Scope**: `schema.yaml` auf 13 Tabellen (Reverse der live V1–V7-SQLite);
`V1__m_trace.sql` daraus als 13-Tabellen-Baseline regeneriert; V2–V7
gelöscht; Fresh-Start-Test 7→1; PG-DDL-Provenienz auf `schema.yaml`
umgestellt; Drift-Gate auf den vollen Stand; ADR-0002-§8.2-Amendment;
CHANGELOG.

**Nicht in Scope**: Datenmigration bestehender DBs (nicht nötig — Runner
checksumfrei, applied-ohne-File wird ignoriert); Modell A/`schema migrate`
(verworfen); semantische Constraint-Namen (generische `fk_0` akzeptiert);
Postgres-Adapter-Logik (0.23.0, unberührt — der Live-Endzustand ändert
sich nicht).

## 3. Architektur-Berührung

- **Round-Trip ist erprobt**: `generate-postgres-schema.sh` baut bereits
  `live.db` aus V1–V7 (sort -V) und macht `schema reverse
  --sqlite-autoincrement-width 64` → `reverse.yaml`. Der reversierte
  **neutrale** Output *ist* das neue `schema.yaml`. Tranche 1/2 wechseln
  nur den Export-Target auf `--target sqlite`.
- **64-bit-PKs**: `--sqlite-autoincrement-width 64` ist Pflicht, sonst
  verengen `playback_events.ingest_sequence` + `srt_health_samples.id` auf
  int32 (ADR-0027 in d-migrate).
- **V3/V5-Rebuild-Tabellen**: der Reverse erfasst den **Endzustand** (nach
  den 12-Schritt-RENAME/DROP-Rebuilds) — genau das, was die neue V1 als
  einzelne `CREATE TABLE` braucht. Spalten-Parität wird verifiziert.
- **Kein Checksum** im Apply-Runner (belegt): V1-Inhaltswechsel bei
  gleicher Versionsnummer bricht keine Bestands-DB.

## 4. Tranchen

| # | Inhalt | Gate |
| --- | --- | --- |
| 1 | **`schema.yaml` → 13-Tabellen-Sollzustand.** `schema reverse` der live V1–V7-SQLite (`--sqlite-autoincrement-width 64`) → neues `schema.yaml`. Verifikation: round-trip `export flyway --target sqlite --version 1` reproduziert eine 13-Tabellen-V1; **volle Pragma-Struktur-Parität** (Spalten/Typen/**NOT NULL**/Default/PK/Indizes) gegen live-SQLite. **⛔ Blockiert:** d-migrate verliert `NOT NULL` auf PK-Spalten (s. Status). | `make schema-validate` grün; alle 13 Tabellen + V6/V7-Spalten + V3/V5-Rebuild-Endzustand; **Struktur-Diff `fresh(Einzel-V1)` vs `live(V1..V7)` == 0 (inkl. NOT-NULL auf allen PK-Spalten)**. |
| 2 | **V1 regenerieren + V2–V7 löschen.** `make schema-generate` → 13-Tabellen `V1__m_trace.sql`; V2–V7-SQLite-Files löschen; Fresh-Start-Test (`migrate_internal_test.go:62`) 7→1. | Frische DB spaltenweise == alter V1–V7-Endzustand; simulierte Bestands-DB (V1–V7 applied) startet ohne Re-Apply/Dirty; `make api-test` grün. |
| 3 | **PG-Provenienz auf `schema.yaml` umstellen.** `generate-postgres-schema.sh`/`make schema-generate-postgres`: `export flyway --target postgresql --source schema.yaml` statt `reverse` der Live-SQLite — schließt den Kreis, entfernt den 0.23.0-Sonderpfad. | Neues PG-`V1__m_trace.sql` == bisheriges reverse-basiertes (byte/semantisch, kein Drift); `make smoke-pg-lab` grün. |
| 4 | **Drift-Gate + ADR + Doku.** `generated-drift-check` deckt das volle 13-Tabellen-V1 (verifizieren, ruft schon `schema-generate`); **ADR-0002 §8.2 amenden** (Privileg reaktivieren: kein managed-prod-Store, PG startet frisch, Runner checksumfrei → Refold safe; historischer V2–V7-Hinweis wie der bestehende V2–V5-Block); CHANGELOG. | `make gates` grün; ADR konsistent. |

## 5. DoD

- [ ] `schema.yaml` beschreibt alle 13 Live-Tabellen (V1–V7-Endzustand).
- [ ] `V1__m_trace.sql` daraus regeneriert (13 Tabellen); V2–V7 gelöscht;
  Fresh-Start-Test 7→1; Bestands-DB-Kompatibilität verifiziert.
- [ ] PG-DDL kommt aus `schema.yaml` (`export flyway --target postgresql`);
  kein Drift zum bisherigen reverse-basierten PG-V1; `smoke-pg-lab` grün.
- [ ] `generated-drift-check` bewacht das volle V1; `make gates` grün.
- [ ] ADR-0002 §8.2 amendiert; CHANGELOG nachgetragen.

## 6. Risiken

- **ADR-Governance**: das Pre-Production-Privileg (§8.2) ist mit v0.22.x
  formal abgelaufen; das Refold braucht ein explizites Amendment, das den
  einmaligen Refold autorisiert bzw. das Privileg an „kein durable
  managed-prod-Store" knüpft. Ohne Amendment widerspricht der Refold der
  ADR-as-written.
- **Reverse-Stil-Drift**: der `reverse.yaml`-Output kann stilistisch vom
  hand-gepflegten `schema.yaml` abweichen (Feld-Reihenfolge, `version`,
  gestrippte Kommentare, `fk_0`-Namen). Akzeptiert (generische Namen
  entschieden); der Round-Trip-Check (Tranche 1) fängt semantische Drifts.
- **Interaktion mit 0.23.0**: die Konsolidierung ändert die *live* SQLite
  nicht (Einzel-V1 = gleicher Endzustand) → PG-Adapter unberührt. Tranche 3
  stellt nur die *Herkunft* des PG-DDL um.
