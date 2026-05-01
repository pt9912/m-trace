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
| Sortierung | Session-Listen bleiben nach `started_at desc, session_id asc`; Event-Listen bleiben nach `server_received_at asc, ingest_sequence asc`. |
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
  Form ist Folge-ADR (siehe roadmap §4 „Durabel-konsistente Cursor-Strategie").
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

## 8. Offene Punkte (für die `0.4.0`-Eingangstranche)

- Tabellenlayout für Events, Sessions und Project-Konfiguration.
- Migrationswerkzeug für SQLite (z. B. `golang-migrate`, `goose`, eigenes
  Embed-Schema).
- Konkrete Retention-Defaults für das lokale Lab.
- Cursor-Format ohne `process_instance_id`-Invalidierung — als Folge-ADR.
