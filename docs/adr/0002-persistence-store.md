# 0002 — Persistenz-Store für Sessions und Playback-Events

> **Status**: Draft  
> **Datum**: 2026-04-30  
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)  
> **Bezug**: `docs/lastenheft.md` OE-3, MVP-16, MVP-27, MVP-40;
> `docs/architecture.md` §11; `docs/roadmap.md` §4/§5;
> `docs/plan-0.2.0.md` Tranche 6.

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

## 4. Vorläufige Bewertung

| Kriterium | A: In-Memory | B: SQLite | C: PostgreSQL |
|---|---:|---:|---:|
| Lokale Einfachheit | hoch | hoch | mittel |
| Restart-Durability | niedrig | hoch | hoch |
| Cursor-Stabilität nach Restart | niedrig | hoch | hoch |
| Betriebsaufwand im Compose-Lab | niedrig | niedrig | mittel |
| Produktionsnähe | niedrig | mittel | hoch |
| Migrationsaufwand aus aktuellem Hexagon | niedrig | mittel | mittel |

Vorläufige Tendenz: SQLite ist der passendste nächste Schritt für lokale
Durability, solange Multi-Instance- und Multi-Tenant-Produktionsbetrieb noch
nicht Ziel des nächsten Releases sind. PostgreSQL bleibt Kandidat, sobald
Deployment-/Skalierungsanforderungen konkret werden.

## 5. Nicht-Entscheidung für `0.2.0`

Für `0.2.0` ist kein finaler Persistenzwechsel blockierend.

Begründung:

- `0.2.0` liefert ein publizierbares Player-SDK, keine neue Storage-API.
- Das SDK erzwingt keine neuen Persistenzgarantien gegenüber `0.1.2`.
- MVP-16 erlaubt lokale Speicherung per In-Memory oder SQLite.
- Die bestehenden Cursor invalidieren Restart bewusst über
  `process_instance_id`; das ist dokumentiert und bleibt bis zur
  Persistenzentscheidung akzeptiert.

Folgearbeit: Dieser Draft muss vor einer dauerhaften Session-/Event-Historie
oder vor stabilen Restart-Cursorn in einen `Accepted` ADR überführt werden.

## 6. Offene Punkte

- Tabellenlayout für Events, Sessions und Project-Konfiguration.
- Migrationswerkzeug für SQLite bzw. PostgreSQL.
- Retention-Defaults für das lokale Lab.
- Cursor-Format ohne `process_instance_id`-Invalidierung.
- Entscheidung, ob SQLite und PostgreSQL beide über dieselben Repository-Ports
  unterstützt werden oder ob SQLite nur lokaler MVP-Store bleibt.
