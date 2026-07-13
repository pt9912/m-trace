# Runbook: SQLite → Postgres Cutover (optional)

> **Status**: Operator-Runbook für die **optionale** Datenmigration eines
> bestehenden SQLite-Deployments auf den Postgres-Scale-out-Store.
> Bezug: [ADR-0007](../adr/0007-sqlite-postgres-data-cutover.md),
> `docs/planning/…/plan-0.24.0-sqlite-postgres-cutover.md`, Risiko `R-29`.
>
> **SQLite bleibt Default.** Der Cutover ist opt-in und nur für Deployments,
> die aktiv auf den (mit [ADR-0006](../adr/0006-postgres-scaleout-adapter.md)
> gelieferten) Postgres-Store wechseln. Läuft ausschließlich als ephemerer
> d-migrate-Ops-Container; die API-Runtime bleibt JDK-frei
> ([ADR-0002](../adr/0002-persistence-store.md)).

## 1. Was der Cutover tut (und nicht tut)

Überträgt die Historie einer bestehenden SQLite-Instanz zeilengetreu nach
Postgres, **sequenz-erhaltend** (die PG-`BIGSERIAL`-Zählerstände setzen den
SQLite-`MAX` fort — der erste neue DB-vergebene Wert kollidiert nicht) und mit
minimaler Downtime (Bulk + inkrementelles Nachziehen, dann ein kurzes
Quiesce-Fenster).

**Nicht** enthalten: Zero-Downtime-Live-Replikation (CDC), Multi-Quell-Cutover,
automatischer Trigger. Der Cutover wird manuell/operator-initiiert.

## 2. Voraussetzungen

- Das **Postgres-Ziel-Schema** ist bereits angelegt (die eingecheckte DDL
  `apps/api/internal/storage/migrations/postgres/`, z. B. durch einen ersten
  Boot der API mit `MTRACE_PERSISTENCE=postgres` gegen die frische DB, oder per
  `psql -f`). Der `doctor` prüft das.
- Das Ziel ist **leer** (frischer Store) — der Bulk fährt `--on-conflict abort`
  und bricht sonst ab. Für ein bereits teilbefülltes Ziel siehe `incremental`.
- Die **Quell-SQLite** ist für den d-migrate-Container-User **lesbar**. Mehr
  ist nicht nötig: d-migrate (≥ 0.9.12, `--read-only` Default) öffnet die
  Quelle für `profile` und die Transfer-Quellseite schreibgeschützt
  (`file:…?mode=ro`, keine `-wal`/`-shm`-Nebendateien). Der `doctor` prüft die
  Lesbarkeit als derselbe uid.
- `docker`, das Repo-Checkout (für den `DMIGRATE_IMAGE`-Pin) und `python3`
  (Profile-Auswertung) sind vorhanden. Ohne Repo: `DMIGRATE_IMAGE` explizit setzen.

## 3. Aufruf

```sh
export SQLITE_DB=/var/lib/mtrace/m-trace.db          # Quell-SQLite (Host-Pfad)
export PG_DSN='postgres://user:pass@host:5432/mtrace?sslmode=disable'  # Ziel
# Wenn der DSN-Host ein Container-Name ist, dem d-migrate-/psql-Container das Netz geben:
# export PG_NETWORK=<docker-netz>

make cutover ARGS=doctor        # Pre-Flight (Tooling / Quelle / Ziel-PG / Schema / leer)
make cutover ARGS=profile       # Phase 0: Typ-Kompatibilität der Quelle (Tripwire)
make cutover ARGS=bulk          # Phase 1: Erstübertragung + Parität/Sequenz-Verifikation
make cutover ARGS=incremental   # Phase 2: Delta nachziehen (wiederholbar, idempotent)
# ... Schritt 2 beliebig oft, bis das Delta klein ist ...
# --- Cutover-Fenster: Writer quiescen (siehe unten) ---
make cutover ARGS=switch        # Phase 3: finaler Re-Sync + Verifikation
```

## 4. Ablauf im Detail

### Phase 0 — `doctor` + `profile`

`doctor` verifiziert: d-migrate-Container lauffähig, Quelle lesbar (als
derselbe uid wie der d-migrate-Container), Ziel-PG erreichbar, Ziel-Schema
vorhanden (≥ 13 Tabellen), Ziel leer. `profile` prüft die **Typ-Gesundheit** der Quelle: bricht ab, wenn ein
Wert sich nicht in seinen eigenen Zieltyp abbilden lässt (echte Korruption);
Cross-Type-Warnungen und leere Tabellen sind Info, kein Abbruch.

### Phase 1 — `bulk`

`data transfer` aller App-Tabellen (`schema_migrations` + `sqlite_%`-Interna
ausgenommen — die verwaltet der Migrations-Runner auf beiden Seiten),
`--on-conflict abort`. Danach: Row-Count-Parität je Tabelle + Sequenz-Erhalt.
Ein zweiter `bulk` gegen ein nicht-leeres Ziel bricht ab (kein Doppel-Load).

### Phase 2 — `incremental`

Zieht das Delta nach: `--since-column ingest_sequence` filtert die
High-Volume-Tabelle `playback_events`; alle übrigen Tabellen werden voll
gescannt und per `--on-conflict skip` dedupliziert. **Idempotent** — beliebig oft
wiederholbar, um das finale Delta (und damit das Quiesce-Fenster) klein zu
halten. `SINCE` ist per ENV setzbar; Default = aktueller Ziel-`MAX`
(Auto-Resume).

> **Wichtig:** `incremental` (`--on-conflict skip`) propagiert **keine
> Änderungen an bestehenden Zeilen** — z. B. Session-State-Updates in
> `stream_sessions`. Diese werden erst im **quiescten `switch`** inhaltlich
> reconciled. Gegen eine laufende Quelle ist das auch nicht anders konsistent
> lösbar.

### Cutover-Fenster — Writer quiescen

**Vor `switch`** die Quell-Writer stoppen: die API anhalten oder read-only
schalten, sodass die SQLite-Quelle nicht mehr geschrieben wird. Nur dann ist der
finale Re-Sync konsistent. Der `switch` kann das **nicht** erzwingen und warnt
nur.

### Phase 3 — `switch`

Zwei Pässe: die **append-only** Tabellen (`APPEND_ONLY`, Default
`playback_events,srt_health_samples`) bekommen den finalen Delta über ihren
PK als Since-Spalte (`LOOKBACK`-Überlapp, Default 100000, absorbiert
Out-of-order-Commits am letzten Watermark); **alle übrigen** Tabellen werden
per `--on-conflict update` voll re-synchronisiert und fangen so alle
zwischenzeitlichen Mutationen. Danach Verifikation: Parität + Duplikatfreiheit +
Sequenz-Erhalt + `SUM`-Content-Aggregat.

## 5. Umschalten + Rollback

Nach grünem `switch` ist das Ziel konsistent mit der (quiescten) Quelle:

1. **Umschalten:** `MTRACE_PERSISTENCE=postgres` (+ `MTRACE_POSTGRES_DSN`)
   setzen und die API neu starten.
2. **Rollback:** bei `MTRACE_PERSISTENCE=sqlite` bleiben bzw. zurücksetzen. Die
   **SQLite-Quelle ist unangetastet** (der Cutover liest nur) — der Rollback ist
   ein reiner Konfig-/Restart-Schritt, solange Postgres noch nicht als
   Autorität bestätigt ist.

## 6. Exit-Codes

`0` ok · `1` hard FAIL (Transfer-/Verifikations-Fehler) · `2` Config-/
Nutzungsfehler · `3` Pre-Flight-Befund (Ziel nicht bereit) · `4` Stub.

## 7. ENV-Referenz

| ENV | Wirkung |
| --- | --- |
| `SQLITE_DB` | Host-Pfad der Quell-SQLite. |
| `PG_DSN` | Ziel-Postgres-DSN. |
| `PG_NETWORK` | Docker-Netz für Client-/d-migrate-Container (wenn DSN-Host = Container-Name). |
| `DMIGRATE_IMAGE` | Override; Default = Single-Source-Pin aus `apps/api/Makefile`. |
| `CHUNK_SIZE` | `data transfer`-Chunkgröße (Default 10000). |
| `SINCE` | `incremental`: `ingest_sequence`-Untergrenze (Default = Ziel-`MAX`). |
| `LOOKBACK` | `switch`: konservativer Re-Scan-Überlapp in Zeilen (Default 100000). |
| `APPEND_ONLY` | `switch`: immutable Tabellen (Default `playback_events,srt_health_samples`). Diese **müssen** insert-only sein; bekäme eine je einen UPDATE-Pfad, gehörte sie hier raus. |

## 8. Reproduzierbarer Smoke

`make smoke-cutover` fährt den vollen Ablauf (doctor inkl. read-only-Quelle ·
profile inkl. read-only-Quelle + Korrupt-Tripwire · bulk inkl. abort-Guard ·
incremental inkl. Idempotenz · switch inkl. Mutations-Beleg) gegen ein
ephemeres Lab. Opt-in, nicht in `make gates`.
