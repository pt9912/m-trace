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
SQLite/PostgreSQL wechselt. Dieser ADR-Entwurf bereitet die Entscheidung vor,
entscheidet sie aber noch nicht final.

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

Diese Punkte muss eine spätere Entscheidung berücksichtigen:

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
| `stream_sessions` | `session_id`, `project_id`, `started_at`, `last_seen_at`, `ended_at`, `state` | PK `session_id`. Index für Session-Listing nach kanonischer Sortierung (`started_at desc, session_id asc`), gefiltert nach `project_id`. |
| `playback_events` | `ingest_sequence`, `project_id`, `session_id`, `event_name`, `client_timestamp`, `server_received_at`, `sequence_number`, `sdk_name`, `sdk_version`, `schema_version`, `meta`, `delivery_status` | PK `ingest_sequence` als `INTEGER PRIMARY KEY AUTOINCREMENT` (global monoton, durable). Index für kanonische Event-Sortierung pro Session. Dedup-Index siehe §8.3. |

`ingest_sequence` ist global, nicht pro `project_id` + `session_id`.
Damit ist der Cursor-Tie-Breaker aus ADR 0004 §5 erfüllt und
`Last-Event-ID` für SSE (ADR 0003) bekommt eine eindeutige globale ID
ohne Zusatzpräfix.

`meta` wird als TEXT-Spalte mit JSON-Inhalt geführt; SQLite erlaubt
dazu JSON-Funktionen, eine Schema-Garantie über die JSON-Struktur
gibt der Store nicht — die liegt im Wire-Format-Vertrag
(`spec/backend-api-contract.md` §3).

### 8.2 Migrationswerkzeug

`d-migrate` (siehe `/Development/d-migrate`,
`ghcr.io/pt9912/d-migrate:latest`) wird als build-/dev-time-Werkzeug
für Schema-Definition und DDL-Generierung gewählt. Begründung:

- Schema-YAML ist neutrale Single-Source-of-Truth; Postgres-Folge-ADR
  bekommt `--target postgresql` ohne erneute manuelle Schema-Pflege.
- d-migrate bietet `schema validate`, `schema generate`, `schema compare`
  und `schema reverse`. Für `0.4.0` werden `schema validate` (CI-Gate)
  und `schema generate --target sqlite` (DDL-Erzeugung) genutzt.
- d-migrate wird in m-trace **nicht zur Laufzeit** eingesetzt: das
  API-Image bleibt JDK-frei.
- d-migrate steht unter Kontrolle desselben Owners wie m-trace. Format-
  Stabilität ist damit nicht von einem Drittanbieter-Release-Zyklus
  abhängig.

Apply-Runner zur Laufzeit ist ein eigener kleiner Go-Bestandteil in
`apps/api/internal/storage/` (oder vergleichbarer Pfad), der die per
`embed.FS` eingebetteten generierten SQL-Files in Reihenfolge
anwendet. Verantwortlich für:

- Lesen der `schema_migrations`-Tabelle, Erkennen offener Versionen.
- Anwenden in einer Transaktion pro Migration; bei Fehler `dirty=1`
  setzen und Start abbrechen.
- Re-Run gegen sauberen Stand ist no-op.
- Re-Start gegen `dirty=1` weigert sich („refuse to start"); die
  Reparatur-Prozedur (manueller Reset oder gezieltes Beheben) steht
  in `docs/user/local-development.md` (Update in §2.6 von
  `plan-0.4.0.md`).

`golang-migrate` und `goose` werden verworfen, weil d-migrate die
neutrale Schema-Definition zusätzlich liefert und keine
Drittanbieter-Library im API-Image landet.

### 8.3 Idempotenz und Event-Deduplikation

- **Session-State-Updates** sind idempotent: `session_ended` und
  Sweeper-Übergänge nutzen ein UPSERT-Muster auf `session_id` und
  setzen `ended_at` nur, wenn noch nicht gesetzt.
- **Event-Dedup** erfolgt als Timeline-Klassifikation, **nicht** als
  Hard-Reject:
  - Dedup-Key ist `(project_id, session_id, sequence_number)` für
    Events mit gesetzter `sequence_number`.
  - Server prüft beim Insert, ob bereits ein Event mit demselben
    Dedup-Key und `delivery_status = 'accepted'` existiert.
  - Falls ja: neuer Event wird trotzdem persistiert, aber mit
    `delivery_status = 'duplicate_suspected'`. Cursor-Sortierung und
    Trace-Korrelation bleiben dadurch unbeschädigt.
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
- Reset-Pfad ist der dedizierte Make-Target (z. B. `make wipe`), der
  das benannte Volume des `api`-Service löscht und beim nächsten
  Start ein leeres Schema erzeugt. `make stop` löscht das Volume
  **nicht**.
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
