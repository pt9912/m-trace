# 0002 — Persistenz-Store für Sessions und Playback-Events

> **Status**: Accepted  
> **Datum**: 2026-04-30 (Draft) · 2026-05-01 (Accepted)  
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)  
> **Bezug**: `spec/lastenheft.md` OE-3, MVP-16, MVP-27, MVP-40, RAK-32;
> `spec/architecture.md` §11; `docs/planning/roadmap.md` §4/§5;
> `docs/planning/plan-0.2.0.md` Tranche 6; `docs/planning/plan-0.3.0.md` §10.

---

## 1. Kontext

`0.1.x` nutzt bewusst In-Memory-Repositories für Playback-Events,
Stream-Sessions und Ingest-Sequenzen. Das ist für das lokale Demo-Lab
zulässig, verliert Daten aber bei Neustart und invalidiert Cursor über die
`process_instance_id`.

OE-3 fragt, ob die MVP-Datenhaltung rein In-Memory bleibt oder auf
SQLite/PostgreSQL wechselt. Dieser ADR (ursprünglich als Entwurf
2026-04-30 gestartet, am 2026-05-01 als `Accepted` angenommen)
beantwortet OE-3 in §6.

`0.2.0` stabilisiert das Player-SDK. Dabei wurde der Storage-Vertrag nicht
geändert: Das Wire-Format bleibt `schema_version: "1.0"`, Events werden
weiterhin append-only aufgenommen, und die bestehenden Session-/Event-Read-
Endpoints bleiben unverändert.

`0.3.0` liefert den Stream-Analyzer ausschliesslich stateless aus
(`POST /api/analyze` persistiert kein Ergebnis, plan-0.3.0 §10). Damit ist
die Persistenz-Frage durch `0.3.0` weder vergrössert noch entschärft worden;
sie kommt unverändert in die `0.4.0`-Planung.

## 2. Entscheidungsfrage

Welche Persistenz soll m-trace als nächstes für lokale und frühe MVP-Nutzung
unterstützen?

| Option | Kurzbeschreibung |
|---|---|
| **A: In-Memory bleibt MVP-Default** | Keine neue Storage-Abhängigkeit; Datenverlust bei Restart bleibt dokumentiert. |
| **B: SQLite als lokaler Durable-Store** | Eine lokale Datei speichert Events, Sessions und Ingest-Sequenzen; Docker-Lab bleibt einfach. |
| **C: PostgreSQL als primärer Store** | Produktionsnäherer Store mit klarer Multi-User-/Retention-Perspektive; höherer Betriebsaufwand. |

## 3. Anforderungen aus SDK-/Schema-Sicht

Folgende Punkte musste die Entscheidung berücksichtigen (siehe §6):

| Bereich | Anforderung |
|---|---|
| Event-Schema-Version | `schema_version` muss pro Batch bzw. gespeicherten Events nachvollziehbar bleiben; Migrationen dürfen Schema `1.0` nicht unlesbar machen. |
| SDK-Version | `sdk.name` und `sdk.version` müssen in Event-Records speicherbar und abfragbar bleiben. |
| Event-Meta | `meta` bleibt flexibel; der Store braucht ein JSON-fähiges Feld oder eine verlustfreie JSON-Serialisierung. |
| Ingest-Sequenz | `ingest_sequence` muss monoton und nach Restart konsistent fortgesetzt werden, sobald Cursor langlebig sein sollen. |
| Session-Ende | `session_ended` und Sweeper-Zustände müssen idempotent auf Session-State wirken. |
| Cursor-Stabilität | Cursor dürfen bei durablem Store nicht mehr pauschal an `process_instance_id` hängen; Restart darf Pagination nicht unnötig invalidieren. |
| Sortierung | Session-Listen bleiben nach `started_at desc, session_id asc`; Event-Listen bleiben nach `server_received_at asc, sequence_number asc, ingest_sequence asc`, wobei `ingest_sequence` der durable Tie-Breaker ist. |
| Retention | Store muss spätere Retention nach Zeit, Projekt und Session erlauben. |
| Lokales Lab | `make dev` soll ohne externen Cloud-Dienst reproduzierbar bleiben. |

## 4. Bewertung der Optionen

| Kriterium | A: In-Memory | B: SQLite | C: PostgreSQL |
|---|---:|---:|---:|
| Lokale Einfachheit | hoch | hoch | mittel |
| Restart-Durability | niedrig | hoch | hoch |
| Cursor-Stabilität nach Restart | niedrig | hoch | hoch |
| Betriebsaufwand im Compose-Lab | niedrig | niedrig | mittel |
| Produktionsnähe | niedrig | mittel | hoch |
| Migrationsaufwand aus aktuellem Hexagon | niedrig | mittel | mittel |

## 5. Verlauf bis zur Entscheidung

Die Persistenzentscheidung wurde bewusst zweimal verschoben:

- **`0.2.0`** liefert ein publizierbares Player-SDK, keine neue Storage-API;
  das SDK erzwingt keine neuen Persistenzgarantien gegenüber `0.1.2`. MVP-16
  erlaubt lokale Speicherung per In-Memory oder SQLite, und die bestehenden
  Cursor invalidieren Restart bewusst über `process_instance_id`.
- **`0.3.0`** liefert den Stream-Analyzer stateless aus; `POST /api/analyze`
  persistiert nichts, jeder Aufruf ist eigenständig (plan-0.3.0 §10).

Damit blieb die Frage „In-Memory bleibt zulässig?" bis zum `0.4.0`-Scope-Cut
offen. Sie wird mit diesem ADR entschieden.

## 6. Entscheidung

**Option B: SQLite als lokaler Durable-Store** wird als nächste Persistenz
für m-trace gewählt.

Auslöser ist **RAK-32** in `0.4.0` (Lastenheft §13.6, *Muss*): „Dashboard
kann Session-Verläufe auch ohne Tempo einfach anzeigen." Das verlangt eine
durable Session-/Event-Historie, die einen API-Restart überlebt — Option A
(In-Memory) erfüllt das nicht. Tempo (RAK-31) ist *Kann*; das durable
Session-View-Backend muss also lokal in m-trace selbst leben.

Gegen Option C (PostgreSQL) sprechen weiterhin:

- `make dev` soll ohne externen Cloud-Dienst reproduzierbar bleiben (§3,
  „Lokales Lab").
- Multi-Instance-Deployment ist für `0.4.0` nicht im Scope; die
  Skalierungs-Treiber für Postgres existieren noch nicht.
- Der Repository-Port-Schnitt bleibt formatneutral, sodass ein späterer
  Wechsel SQLite → PostgreSQL als Folge-ADR möglich bleibt, ohne `0.4.0`
  zu blockieren.

PostgreSQL bleibt als zukünftiger Folge-ADR offen, sobald Multi-Instance-,
Multi-Tenant- oder Retention-Last-Anforderungen konkret werden.

## 7. Konsequenzen

- **Repository-Port-Migration als `0.4.0`-Eingangstranche.** Die heute
  In-Memory-implementierten Driven-Adapter (Sessions, Events,
  Ingest-Sequenz) bekommen eine SQLite-Implementierung hinter den
  bestehenden Ports. Driving-Adapter und Anwendungsdienste bleiben
  unverändert.
- **Cursor-Format wird durable.** Das heutige `process_instance_id`-Cursor-
  Format invalidiert nach Restart — mit SQLite wird es durch eine
  Storage-getragene Form ersetzt (Sequence-ID oder opakes Token). Genaue
  Form ist Folge-ADR (siehe roadmap §4 „Dauerhaft konsistente Cursor-Strategie").
- **Schema-Versionierung.** SQLite-Schema bekommt eine eigene
  Migrations-Versionsspur, getrennt vom Event-Wire-Schema (`schema_version`).
  Migrationswerkzeug-Wahl ist Teil der `0.4.0`-Eingangstranche.
- **Retention-Defaults im Lab.** SQLite-Datei hat ein dokumentiertes
  Default-Retention-Fenster pro Projekt/Session; konkrete Defaults landen
  in `plan-0.4.0.md` und in `docs/user/local-development.md`.
- **Compose-Lab.** Die SQLite-Datei lebt in einem benannten Volume des
  `api`-Service, damit `make dev`-Cycles die Daten überdauern; `make stop`
  räumt das Volume *nicht* automatisch.
- **Doku-Update.** `spec/architecture.md` §11 (Storage), `docs/user/local-development.md`
  (Retention/Reset-Hinweis), `README.md` Status-Zeile (Datenhaltung) werden
  in der Tranche, die SQLite einführt, mitgezogen.

## 8. Geschlossene Punkte (`0.4.0`-Eingangstranche)

Die ursprünglich offenen Punkte für `0.4.0` sind im Rahmen von
`plan-0.4.0.md` §2.1 wie folgt entschieden. Die feingranulare
Implementierung (Spalten-Datentypen, Indexnamen, exakte Defaults) liegt
in der Schema-YAML aus §2.2; dieser Abschnitt fixiert die
Architektur-Aussagen, die §2.2 nicht eigenmächtig verändern darf.

### 8.1 Tabellen-Skizze

`0.4.0` führt vier Tabellen ein, plus eine vom Migrations-Apply-Runner
verwaltete Versions-Tabelle. Spaltenliste hier auf das Architektur-
Notwendige reduziert:

| Tabelle | Pflicht-Spalten | Schlüssel und Indizes |
|---|---|---|
| `schema_migrations` | `version`, `applied_at`, `dirty` | PK `version`. Vom Apply-Runner verwaltet, nicht aus der Schema-YAML generiert. |
| `projects` | `project_id` | PK `project_id`. Minimal in `0.4.0`; weitere Project-Konfigurationsfelder bleiben Use-Case-Service-intern. |
| `stream_sessions` | `session_id`, `project_id`, `started_at`, `last_seen_at`, `ended_at`, `state`, `event_count`, `correlation_id` | PK `session_id`. Index für Session-Listing nach kanonischer Sortierung (`started_at desc, session_id asc`), gefiltert nach `project_id`. `event_count` (BIGINTEGER NOT NULL DEFAULT 0) wird vom Adapter beim ersten Event auf 1 gesetzt und bei Folge-Events inkrementiert; Read-Pfad spiegelt das im JSON-Feld `event_count` der Session-Detail-Response (Vertrag siehe `spec/backend-api-contract.md` §3.7 / §10). `correlation_id` ist nullable und wird in Tranche 2 (`plan-0.4.0.md` §3) befüllt, sobald die Trace-/Korrelations-Strategie steht. |
| `playback_events` | `ingest_sequence`, `project_id`, `session_id`, `event_name`, `client_timestamp`, `server_received_at`, `sequence_number`, `sdk_name`, `sdk_version`, `schema_version`, `meta`, `delivery_status`, `trace_id`, `span_id`, `correlation_id` | PK `ingest_sequence` als `INTEGER PRIMARY KEY AUTOINCREMENT` (global monoton, durable). Index für kanonische Event-Sortierung pro Session. Nicht-eindeutiger Index `idx_playback_events_dedup` auf `(project_id, session_id, sequence_number)` für die Dedup-Lookup in §8.3 (Race-Schutz dort über `BEGIN IMMEDIATE`, nicht über DB-Constraint). `trace_id`/`span_id`/`correlation_id` sind nullable und werden in Tranche 2 befüllt; das Schema reserviert sie jetzt, damit §2.2 sie nicht nachträglich hinzufügen muss. |

`ingest_sequence` ist global, nicht pro `project_id` + `session_id`.
Damit ist der Cursor-Tie-Breaker aus ADR 0004 §5 erfüllt und
`Last-Event-ID` für SSE (ADR 0003) bekommt eine eindeutige globale ID
ohne Zusatzpräfix. `INTEGER PRIMARY KEY AUTOINCREMENT` vermeidet
ROWID-Reuse über die `sqlite_sequence`-Tabelle; Lücken nach Rollback
sind erlaubt und vom Cursor-Kontrakt (§10.4 in
`spec/backend-api-contract.md`) **nicht** ausgeschlossen — gefordert
ist nur Monotonie, keine Lückenfreiheit.

`meta` wird im neutralen Schema als `json` deklariert. Beim
DDL-Generate per d-migrate mappt das auf:

- SQLite: `TEXT` (keine Typ-Validierung — SQLite kennt keinen JSON-Typ).
- MySQL: `JSON` (Typ-Level-Validierung beim INSERT).
- PostgreSQL: `JSONB` (Typ-Level-Validierung beim INSERT).

JSON-Validierung erfolgt **im Application-Layer** vor dem Insert
(z. B. via `encoding/json.Valid()` oder Decoder im Adapter). Eine
DB-CHECK-Constraint mit `json_valid(meta)` wäre nicht portabel
(SQLite/MySQL kennen die Funktion, PostgreSQL nicht) und gehört
deshalb nicht in das neutrale Schema. Adapter-Tests (`plan-0.4.0.md`
§2.3) decken den Validierungspfad ab; eine Schema-Garantie über die
*innere* JSON-Struktur gibt der Store ohnehin nicht — die liegt im
Wire-Format-Vertrag (`spec/backend-api-contract.md` §3).

`trace_id`, `span_id` und `correlation_id` sind in `0.4.0` als TEXT-
Spalten reserviert. Konkrete Wertebereiche (W3C-Trace-Context: 32-Hex
für `trace_id`, 16-Hex für `span_id`; `correlation_id` als
serverseitig vergebener Fallback) und die Befüllungs-Regeln
(Quelle: SDK-Header oder serverseitig generiert; Vererbung von
Session zu Event) werden in Tranche 2 (`plan-0.4.0.md` §3, RAK-29)
festgelegt. Read-Pfad-Exposure im API-Kontrakt §3.7 erfolgt zusammen
mit Tranche 2.

### 8.2 Migrationswerkzeug

`d-migrate` (Container `ghcr.io/pt9912/d-migrate`, Quellen unter
<https://github.com/pt9912/d-migrate>; konkreter Versions-Pin im
`apps/api/Makefile`) wird als build-/dev-time-Werkzeug für Schema-
Definition und DDL-Generierung gewählt. Begründung:

- Schema-YAML ist neutrale Single-Source-of-Truth; Postgres-Folge-ADR
  bekommt `--target postgresql` ohne erneute manuelle Schema-Pflege.
- d-migrate bietet `schema validate`, `schema generate`, `schema compare`,
  `schema reverse` und `export flyway|liquibase|django|knex`. Für `0.4.0`
  werden `schema validate` (CI-Gate) und `export flyway --target sqlite`
  (Baseline-DDL-Erzeugung im Flyway-File-Format `V<n>__<desc>.sql`) genutzt.
  `export flyway` ist byte-deterministisch ohne Generated-Timestamp im
  Header — Re-Generation ohne Schema-Änderung erzeugt keinen git-Diff,
  keine Workaround-Pipeline (sed/perl) nötig. `schema generate` bleibt
  als Alternative verfügbar, ist hier aber nicht der genutzte Pfad.
- d-migrate wird in m-trace **nicht zur Laufzeit** eingesetzt: das
  API-Image bleibt JDK-frei.
- d-migrate steht unter Kontrolle desselben Owners wie m-trace. Format-
  Stabilität ist damit nicht von einem Drittanbieter-Release-Zyklus
  abhängig.

Apply-Runner zur Laufzeit ist ein eigener kleiner Go-Bestandteil in
einem neuen Paket unter `apps/api/internal/`; den konkreten Pfad
und Paketnamen legt `plan-0.4.0.md` §2.2 fest, sobald die Schema-
YAML steht. Der Runner wendet die per `embed.FS` eingebetteten,
aus der Schema-YAML generierten SQL-Files in Reihenfolge an.
Verantwortlich für:

- Lesen der `schema_migrations`-Tabelle, Erkennen offener Versionen.
- Anwenden in einer Transaktion pro Migration; bei Fehler `dirty=1`
  setzen und Start abbrechen.
- Re-Run gegen sauberen Stand ist no-op.
- Re-Start gegen `dirty=1` weigert sich („refuse to start") — das
  Refuse-Verhalten ist verbindlich; `plan-0.4.0.md` §2.6 darf
  Reparatur-Wording und Schritte ergänzen, das Refuse-to-Start-
  Verhalten aber **nicht** weichspülen (kein automatischer Retry,
  keine Warn-only-Variante). Die Reparatur-Prozedur (manueller Reset
  via `make wipe` oder gezieltes Beheben) steht in
  `docs/user/local-development.md` (Update in §2.6).

`golang-migrate` und `goose` werden verworfen, weil d-migrate die
neutrale Schema-Definition zusätzlich liefert und keine
Drittanbieter-Library im API-Image landet.

**Migrations-File-Konvention**: das Initial-Baseline-DDL
(`V1__m_trace.sql`) wird aus `schema.yaml` per
`d-migrate export flyway --target sqlite --version 1` erzeugt; das
File ist regenerierbar und nicht hand-zu-pflegen. Folge-Migrationen
(`V2__…sql`, `V3__…sql`) sind hand-gepflegt, bis `d-migrate
schema migrate` (Diff-basiert, geplant) verfügbar ist. Apply-Runner
behandelt beide File-Klassen identisch (Pattern `V<n>__.+\.sql`). Falls
eine Tranche ein DDL-Feature braucht, das d-migrate noch nicht
modelliert, wird das **dort** entschieden — entweder ist das Feature
anders lösbar (z. B. Race-Schutz über `BEGIN IMMEDIATE` statt
Partial-Index in §8.3), oder d-migrate selbst wird erweitert.

### 8.3 Idempotenz und Event-Deduplikation

- **Session-State-Updates** sind idempotent: `session_ended` und
  Sweeper-Übergänge nutzen ein UPSERT-Muster auf `session_id` und
  setzen `ended_at` nur, wenn noch nicht gesetzt.
- **Event-Dedup** erfolgt als Timeline-Klassifikation, **nicht** als
  Hard-Reject. Race-Schutz unter konkurrenten Writern erfolgt über
  SQLite-Schreibserialisierung mit `BEGIN IMMEDIATE`, **nicht** über
  einen DB-Constraint:
  - Dedup-Key ist `(project_id, session_id, sequence_number)` für
    Events mit gesetzter `sequence_number`.
  - Adapter-Algorithmus pro Insert (in einer einzigen Transaktion,
    eröffnet mit `BEGIN IMMEDIATE`, das den SQLite-DB-Write-Lock
    direkt akquiriert und alle anderen Writer bis zum Commit
    blockiert):
    1. SELECT mit Index-Lookup, um eine bereits `accepted`-Zeile
       mit demselben Dedup-Key zu finden.
    2. Falls vorhanden: INSERT mit
       `delivery_status = 'duplicate_suspected'`. Commit.
    3. Falls nicht vorhanden: INSERT mit
       `delivery_status = 'accepted'`. Commit.
  - SQLite serialisiert alle Writer per DB-Lock; ein zweiter
    Writer mit demselben Dedup-Key, der erst nach dem Commit des
    ersten startet, sieht in seinem SELECT die `accepted`-Zeile und
    klassifiziert deterministisch als `duplicate_suspected`. Race
    bleibt damit unmöglich, solange der Adapter `BEGIN IMMEDIATE`
    verbindlich nutzt; Tests aus §2.3 prüfen das mit Concurrent-
    Writern.
  - Events ohne `sequence_number` werden immer als
    `delivery_status = 'accepted'` aufgenommen; es gibt keinen
    automatischen Dedup ohne expliziten Schlüssel.
- Der Wert `delivery_status = 'replayed'` ist in der Spalte vorhanden,
  wird in `0.4.0` aber nur dann gesetzt, wenn ein Use-Case-Service
  explizit einen Replay-Pfad signalisiert. Für `0.4.0` bleibt der
  Replay-Pfad ungenutzt; Spalte und Wert sind reserviert für spätere
  Tranchen.
- Dashboard zeigt `duplicate_suspected` als sichtbare Klassifikation
  in der Timeline (Detailumsetzung in `plan-0.4.0.md` §5 Tranche 4).

### 8.4 Retention-Defaults

Für `0.4.0` gilt **„unlimited mit dokumentiertem Reset-Pfad"**:

- Keine automatische TTL-Aufräumung in `0.4.0`. SQLite-Datei wächst
  bis zum manuellen Reset.
- Reset-Pfad ist `make wipe` — verbindlicher Target-Name, nicht
  beispielhaft. Es löscht das benannte Volume des `api`-Service und
  erzeugt beim nächsten Start ein leeres Schema. `make stop` löscht
  das Volume **nicht**.
- Konkrete Retention-Werte (Zeitfenster, Pro-Projekt-Limit) werden
  Folge-Arbeit, sobald reale Datenmengen oder Multi-Tenant-Last
  auftreten. Bis dahin ist die Spalten-Skizze (`server_received_at`,
  `started_at`, `project_id`, `session_id`) für eine spätere Retention
  ausreichend.

### 8.5 Cursor-Format

Wird in [ADR 0004 — Dauerhaft konsistente Cursor-Strategie](./0004-cursor-strategy.md)
entschieden. Wesentlich für ADR 0002:

- Cursor-Tokens enthalten kein `process_instance_id` mehr.
- `ingest_sequence` aus §8.1 ist die durable Tie-Breaker-Sequenz,
  auf die der Event-Cursor zurückgreift.
- Schema-Migrationen müssen `ingest_sequence`-Monotonie und
  Eindeutigkeit garantieren, damit ADR 0004 §6 `cursor_expired` nur
  durch echten Reset (Retention/Wipe), nicht durch ID-Reuse,
  erreichbar ist.
