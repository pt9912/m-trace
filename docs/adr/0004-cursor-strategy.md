# 0004 — Dauerhaft konsistente Cursor-Strategie für Pagination

> **Status**: Accepted  
> **Datum**: 2026-05-02  
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)  
> **Bezug**: `spec/lastenheft.md` RAK-32; `docs/adr/0002-persistence-store.md`;
> `docs/planning/roadmap.md` §4 (Folge-ADR „Dauerhaft konsistente Cursor-Strategie");
> `docs/planning/plan-0.4.0.md` §2.1, §2.5; `spec/backend-api-contract.md` §10.

---

## 1. Kontext

`0.1.0` führt Cursor-basierte Pagination für `GET /api/stream-sessions`
und für die Event-Liste in `GET /api/stream-sessions/{id}` ein. Der
Cursor ist heute base64url(JSON) mit den Feldern `pid`
(`process_instance_id`), Zeitstempel und ID-/Sequenz-Werten.

Die `process_instance_id` ist ein zufälliger 32-Hex-String, der bei
jedem API-Start neu erzeugt wird. Dadurch invalidiert ein API-Restart
oder ein Cross-Instance-Routing alle bestehenden Cursor: der Server
erkennt den fremden `pid`-Wert und liefert `400 Bad Request` mit Body
`{"error":"cursor_invalid"}`. Clients müssen den Snapshot neu laden.

Das war für `0.1.x` bewusst akzeptabel, weil Storage rein In-Memory
und ohnehin neustart-flüchtig war.

ADR 0002 entscheidet SQLite als lokalen Durable-Store für `0.4.0`.
Damit wird Storage neustart-stabil — der Cursor sollte es auch sein.
Roadmap §4 hält dafür eine Folge-ADR „Dauerhaft konsistente
Cursor-Strategie" fest, parallel zur SQLite-Migration. Dieser ADR
schließt diese Lücke.

## 2. Entscheidungsfrage

Welche Cursor-Form ersetzt das `process_instance_id`-basierte
`0.1.x`-Format und bleibt nach API-Restart sowie nach späterer
Storage-Migration (z. B. Postgres-Folge-ADR) stabil?

| Option | Kurzbeschreibung |
|---|---|
| **A: Storage-Position-Token** | Cursor enthält ausschließlich durable Storage-Werte (`server_received_at`, optional `sequence_number`, `ingest_sequence` für Events; `started_at`, `session_id` für Sessions). Token bleibt JSON-base64url, aber mit explizitem `cursor_version`-Feld. |
| **B: Opaker, server-signierter Token** | Server erzeugt einen HMAC-signierten oder verschlüsselten Token, der intern eine durable Persistenz-ID kodiert. Inhalt für Clients undurchsichtig. |
| **C: Persistenz-ID-Cursor** | Cursor ist eine reine Persistenz-ID (z. B. SQLite-`ROWID` oder eigener Sequence-Wert). Server sortiert intern weiter nach kanonischer Reihenfolge, paginiert aber über die Persistenz-ID. |

## 3. Anforderungen

| Bereich | Anforderung |
|---|---|
| Restart-Stabilität | Cursor, der vor einem API-Restart erzeugt wurde, muss nach dem Restart entweder weiter funktionieren oder mit einem definierten Fehlercode aus einer Kompatibilitätsmatrix abgewiesen werden. |
| Versionierbarkeit | Wechsel zwischen Cursor-Formaten muss erkennbar sein, ohne zwischen Endpunkten oder Heuristiken zu raten. |
| Storage-Portabilität | Spätere Migration auf Postgres (Folge-ADR aus Roadmap §4) darf das Cursor-Format nicht erneut brechen. |
| Kanonische Sortierung | Pagination muss mit der API-seitig spezifizierten Sortierung (Sessions: `started_at desc, session_id asc`; Events: `server_received_at asc, sequence_number asc, ingest_sequence asc`) konsistent bleiben. |
| Recovery | Bei abgewiesenem Cursor muss der Client durch erneuten Snapshot-Load (Endpunkt ohne `cursor`) recovern können, ohne Retry-Loop. |
| Sicherheit | Cursor-Inhalt darf keine sensiblen Daten preisgeben; Server muss alle Felder serverseitig validieren und niemals als Filter blind übernehmen. |
| `0.4.0`-Scope | Die Lösung darf keine zusätzliche Krypto-Library oder Schlüsselverwaltung erzwingen, die nicht ohnehin schon vorhanden ist. |

## 4. Bewertung der Optionen

| Kriterium | A: Storage-Position-Token | B: Opaker signierter Token | C: Persistenz-ID-Cursor |
|---|---:|---:|---:|
| Restart-Stabilität | hoch | hoch | hoch |
| Implementierungsaufwand | niedrig | mittel-hoch | niedrig |
| Lesbarkeit / Debuggbarkeit | hoch | niedrig | hoch |
| Forward-Kompatibilität (Versionierung) | gut (`cursor_version`-Feld) | gut (Server kontrolliert Format) | mittel (neuer ID-Typ erfordert neuen Token) |
| Storage-Portabilität SQLite → Postgres | hoch (durable Sortier-Werte sind portabel) | hoch (Server kann Inhalt ändern) | mittel (`ROWID` ist SQLite-spezifisch; eigener Sequence-Generator nötig) |
| Krypto-/Key-Management | nicht nötig | Schlüssel-Rotation, Secret-Verwaltung | nicht nötig |
| Konsistenz mit kanonischer Sortierung | direkt (Cursor enthält Sortier-Tupel) | indirekt (Server muss Position rekonstruieren) | indirekt (Pagination ≠ Sortier-Reihenfolge möglich) |
| Risiko aus R-3 / ADR-0001 | niedrig | mittel (zusätzliche Lib oder Code) | niedrig |

## 5. Entscheidung

**Option A: Storage-Position-Token mit `cursor_version`-Feld** wird
gewählt.

Der Cursor ist weiter base64url-kodiertes JSON, bekommt aber ein
verbindliches `v`-Feld (Cursor-Version) und enthält ausschließlich
durable Storage-Werte. Das aktuelle `pid`-Feld
(`process_instance_id`) entfällt vollständig.

Konkrete Token-Inhalte für `cursor_version: 2`:

- **Sessions-Listen-Cursor**:
  `{"v":2,"sa":"<server_started_at, RFC3339Nano UTC>","sid":"<session_id>"}`
- **Session-Events-Cursor**:
  `{"v":2,"rcv":"<server_received_at, RFC3339Nano UTC>","seq":<int|null>,"ing":<int>}`

`ingest_sequence` (`ing`) bleibt der serverseitig durable Tie-Breaker
und wird als globale, monoton steigende Persistenz-Sequenz aus SQLite
geliefert (siehe ADR 0002 §8). Damit ist die kanonische Sortierung
restart-stabil, ohne dass der Cursor die `process_instance_id` enthält.

Option B wird verworfen, weil `0.4.0` keinen Krypto-/Key-Management-
Aufwand rechtfertigt: der Cursor enthält keine sensiblen Daten,
Server validiert alle Felder ohnehin defensiv, und HMAC-Signaturen
würden eine zusätzliche Geheimnis-Verwaltung im Compose-Lab erzwingen.

Option C wird verworfen, weil eine reine Persistenz-ID die kanonische
Sortier-Reihenfolge nicht direkt ausdrückt und bei späterer
Postgres-Migration ein eigener portabler Sequence-Generator nötig
würde. Das `(rcv, seq, ing)`-Tupel aus Option A bleibt unter beiden
Storage-Backends bedeutungsgleich.

## 6. Cursor-Kompatibilitätsmatrix

Server entscheidet pro eingehendem Cursor anhand der Decode-Reihenfolge:

| Klasse | Erkennung | HTTP-Status | Body | Client-Recovery |
|---|---|---|---|---|
| `accepted` | Token decodiert; `v == 2`; alle Pflichtfelder vorhanden und valide. | `200 OK` (Listen-Endpoint antwortet wie ohne Cursor, aber mit Filter). | regulärer Listen-Response. | weiter paginieren mit `next_cursor`. |
| `cursor_invalid_legacy` | Token decodiert; `v`-Feld fehlt oder enthält `1`; oder `pid`-Feld vorhanden (Hinweis auf `0.1.x`-Format). | `400 Bad Request`. | `{"error":"cursor_invalid_legacy","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |
| `cursor_invalid_malformed` | Base64- oder JSON-Decode schlägt fehl; oder `v`-Feld enthält unbekannten Wert; oder Pflichtfeld fehlt/Format ungültig. | `400 Bad Request`. | `{"error":"cursor_invalid_malformed","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |
| `cursor_expired` | Cursor referenziert eine Storage-Position, die durch Retention-/Wipe-Pfad nicht mehr existiert (z. B. nach `make wipe` oder zukünftiger TTL-Aufräumung). | `410 Gone`. | `{"error":"cursor_expired","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |

Keine der Fehlerklassen liefert `Retry-After`. Recovery ist
deterministisch: Snapshot ohne `cursor` laden; ein Retry-Loop mit
demselben Cursor ist nutzlos und gilt als Client-Fehler.

`cursor_invalid_legacy` ist eine **dauerhafte** Reject-Klasse: ein
einzelner Legacy-Cursor wird nicht „einmalig" akzeptiert. „Einmalig"
gilt nur für das Client-Verhalten *pro Cursor-Wert*: nach
Snapshot-Reload darf derselbe Legacy-Cursor nicht erneut gesendet
werden.

## 7. Konsequenzen

- **Wire-Format**: Cursor-Tokens enthalten `cursor_version` als
  `v`-Feld; Cursor ohne `v` werden als Legacy behandelt.
- **Server**: `apps/api/adapters/driving/http/cursor.go` wird so
  umgebaut, dass es `v`-Feld parst, Felder validiert und die obigen
  Fehlerklassen mit definiertem Body liefert.
- **Domain-Cursor-Typen**: `driving.ListSessionsCursor` und
  `driving.SessionEventsCursor` verlieren das
  `ProcessInstanceID`-Feld; Application-/Domain-Layer arbeitet nur
  noch mit durable Sortier-Werten.
- **`ingest_sequence` als globale Sequenz**: SQLite-Schema definiert
  `ingest_sequence` als `INTEGER PRIMARY KEY AUTOINCREMENT` auf der
  Event-Tabelle; Eindeutigkeit und Monotonie sind global
  (nicht nur pro `project_id` + `session_id`). Details in ADR 0002 §8
  und im API-Kontrakt §10.
- **Fehlerbody**: bisheriger Body `{"error":"cursor_invalid"}` wird
  durch die feiner aufgelösten Fehlerklassen aus §6 ersetzt; der
  bisherige Sammelbegriff wird in keinem Migrationspfad beibehalten.
- **Tests**: alle Matrix-Klassen aus §6 sind in Backend-Tests
  abgedeckt; insbesondere der Legacy-Reject-Test mit echtem
  `0.1.x`-Cursor-String, ein Malformed-Test pro Decode-Stufe
  (Base64, JSON, `v`-Wert, Pflichtfeld) und ein Restart-Stabilitäts-
  Test mit echter SQLite-Datei.
- **Doku**: `spec/backend-api-contract.md` §10 (Persistenz, Sub-Section
  „Pagination und Cursor") führt die Matrix als Vertrag; SDK-Doku
  zeigt das Recovery-Verhalten ohne Retry-Loop.
- **Keine SDK-Breaking-Change**: Player-SDK sendet keine Cursor; nur
  das Dashboard ist betroffen. Dashboard-Code wird in Tranche 4 (§5
  in `plan-0.4.0.md`) angepasst.

## 8. Offene Punkte

- Konkrete `cursor_expired`-Implementierung hängt davon ab, ob
  `0.4.0` einen Retention-Pfad einführt (siehe ADR 0002 §8 und
  `plan-0.4.0.md` §2.4 / §2.6). Solange Retention auf „unlimited mit
  dokumentiertem Reset-Pfad" steht, ist `cursor_expired` ausschließlich
  durch `make wipe` o. Ä. erreichbar.
- Postgres-Folge-ADR muss prüfen, ob `ingest_sequence` als globale
  Sequenz mit gleicher Semantik beibehalten wird oder ob ein
  expliziter portabler Sequence-Generator nötig wird. Cursor-Format
  selbst bleibt portabel.
- Falls eine spätere `cursor_version: 3` nötig wird (z. B.
  Multi-Tenant-Tagging), gelten dieselben Matrix-Regeln: alte
  Versionen werden als `cursor_invalid_legacy` abgewiesen, nicht
  „toleriert".
