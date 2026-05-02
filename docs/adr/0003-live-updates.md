# 0003 — Live-Updates für Dashboard-Session-Verläufe

> **Status**: Accepted  
> **Datum**: 2026-05-01  
> **Beteiligt**: m-trace-Owner (Solo-Entwicklung)  
> **Bezug**: `spec/lastenheft.md` OE-5, MVP-31, RAK-29, RAK-32;
> `docs/planning/in-progress/roadmap.md` §4/§5; `docs/planning/open/risks-backlog.md` R-3;
> `docs/planning/in-progress/plan-0.4.0.md`.

---

## 1. Kontext

`0.4.0` baut die vorbereitete OTel-Grundlage zu einer nutzbaren
Korrelationsschicht aus. Das Dashboard soll Session-Verläufe ohne Tempo
anzeigen (RAK-32) und laufende Sessions sinnvoll aktualisieren können.

OE-5 fragt, ob Live-Updates per Polling, WebSocket oder Server-Sent Events
umgesetzt werden. Aus ADR 0001 ist R-3 offen: Das WebSocket-Ökosystem in Go
hat keinen Standard-Library-Default und vergrößert die API-Surface gegenüber
HTTP-only-Endpunkten.

Die Update-Richtung für `0.4.0` ist einseitig: Das Dashboard muss neue
Session-/Event-Zustände vom Backend empfangen. Interaktive Befehle vom
Dashboard zurück zum Backend sind nicht Teil des Release-Scopes.

## 2. Entscheidungsfrage

Welcher Live-Update-Mechanismus wird für Dashboard-Session-Verläufe in
`0.4.0` verwendet?

| Option | Kurzbeschreibung |
|---|---|
| **A: Polling** | Dashboard fragt bestehende REST-Endpunkte periodisch neu ab. |
| **B: Server-Sent Events (SSE)** | Dashboard öffnet einen HTTP-Stream; Backend sendet einseitige Updates. |
| **C: WebSocket** | Dashboard und Backend nutzen eine bidirektionale Verbindung. |

## 3. Anforderungen

| Bereich | Anforderung |
|---|---|
| Richtung | Backend → Dashboard reicht für `0.4.0`; keine bidirektionalen Kommandos. |
| Snapshot | REST bleibt Quelle für initiale Session-Liste, Detaildaten und Backfill. |
| Fallback | Dashboard muss bei Stream-Abbruch nutzbar bleiben. |
| Backend-Fit | Lösung soll zu Go `net/http`, bestehendem Router und Docker-Lab passen. |
| Cardinality | Live-Update-Metadaten dürfen keine Prometheus-Cardinality erhöhen. |
| Betrieb | Lokales Lab soll ohne zusätzliche Infrastruktur auskommen. |

## 4. Bewertung der Optionen

| Kriterium | A: Polling | B: SSE | C: WebSocket |
|---|---:|---:|---:|
| Implementierungsaufwand | niedrig | mittel | mittel |
| Latenz bei laufender Session | mittel | gut | gut |
| Backend-Komplexität | niedrig | niedrig-mittel | mittel-hoch |
| Go-Standard-Library-Fit | hoch | hoch | niedrig |
| Bidirektionalität | nein | nein | ja |
| Passt zum `0.4.0`-Scope | mittel | hoch | niedrig-mittel |
| Risiko aus R-3 | niedrig | niedrig | hoch |

## 5. Entscheidung

**Option B: Server-Sent Events (SSE) mit Polling-Fallback** wird für
Dashboard-Live-Updates in `0.4.0` gewählt.

SSE deckt die benötigte Richtung Backend → Dashboard ab, bleibt HTTP-basiert
und ist mit Go `net/http` ohne zusätzliche WebSocket-Library umsetzbar.
Damit wird R-3 entschärft: `0.4.0` benötigt keine Entscheidung zwischen
`gorilla/websocket`, `nhooyr.io/websocket`, `coder/websocket` oder einer
anderen WebSocket-Abhängigkeit.

REST bleibt der autoritative Datenpfad. Das Dashboard lädt initiale Snapshots
und Backfill weiter über bestehende bzw. erweiterte
`GET /api/stream-sessions`-Endpunkte. SSE dient als Benachrichtigungs- oder
Delta-Kanal für laufende Updates; bei Verbindungsabbruch nutzt das Dashboard
Polling/Refresh gegen REST.

WebSocket wird nicht in `0.4.0` eingeführt. Falls spätere Releases echte
bidirektionale Steuerbefehle, Subscribe-/Unsubscribe-Kommandos oder
Multi-Client-Koordination brauchen, wird eine neue ADR geschrieben.

## 6. Konsequenzen

- **API-Surface bleibt HTTP-only.** `0.4.0` ergänzt einen oder mehrere
  `GET`-basierte SSE-Endpunkte; keine WebSocket-Route und keine WebSocket-
  Dependency.
- **REST bleibt Source of Truth.** SSE-Events dürfen klein sein und auf
  REST-Ressourcen verweisen. Verlorene SSE-Nachrichten werden über REST-
  Backfill korrigiert.
- **Polling-Fallback ist Pflicht.** Dashboard-Code muss ohne SSE-Verbindung
  weiter bedienbar bleiben, z. B. mit periodischem Refresh oder manuellem
  Reload.
- **CORS bleibt explizit.** SSE-Endpunkte gehören zum Dashboard-Lese-Pfad
  und nutzen die bestehende Dashboard-Origin-Allowlist; Credentials bleiben
  aus.
- **Prometheus bleibt aggregiert.** SSE-Verbindungs- oder Event-Metriken
  dürfen nur kontrollierte Labels wie `outcome` oder `route` verwenden,
  keine `session_id`.
- **Tests.** Backend-Tests müssen Stream-Header, Heartbeats/Keepalive,
  Abbruchverhalten und Fallback-freundliche Reconnect-Semantik abdecken.
  Dashboard-Tests decken SSE-Erfolg und Fallback ab.

## 7. Umsetzungsschnitt für `0.4.0`

Der genaue Endpunkt-Schnitt wird in `plan-0.4.0.md` umgesetzt. Der Zielkorridor
ist:

- `GET /api/stream-sessions/stream` für Session-Listen-Invalidierung oder
  kompakte Session-Summary-Updates.
- Optional `GET /api/stream-sessions/{id}/events/stream` für Detailansichten,
  falls ein globaler Stream plus REST-Backfill nicht ausreicht.
- EventSource-kompatibles Format mit `event:`, `id:` und `data:`; `id` darf
  nur durable Cursor-/Sequenzdaten enthalten, keine Prozess-ID.
- Heartbeat-Kommentare halten lokale Proxies und Browser-Verbindungen frisch.

## 8. Offene Punkte für die `0.4.0`-Tranche

- Ob ein globaler SSE-Stream genügt oder zusätzlich ein Session-Detail-Stream
  gebraucht wird.
- Konkretes Event-Payload-Schema und Reconnect-Backfill-Regel.
- ~~Ob `Last-Event-ID` direkt auf den durable Cursor aus der SQLite-Tranche
  abgebildet wird oder nur eine Stream-Sequenz ist.~~ **Resolved durch
  [ADR 0002 §8.1](./0002-persistence-store.md): `Last-Event-ID` ist die
  globale `ingest_sequence` aus der `playback_events`-Tabelle. Damit ist
  der Reconnect-Backfill restart-stabil, ohne Zusatzpräfix und ohne
  separaten Stream-Sequence-Generator.**
- Konkrete Dashboard-Fallback-Intervalle für Polling.
