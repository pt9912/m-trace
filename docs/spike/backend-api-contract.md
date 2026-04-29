# Backend-API-Kontrakt — Spike

> **Status**: Frozen während Spike-Vergleich; nach ADR-0001 (Accepted)
> ist Post-Spike-Pflege erlaubt — Änderungen werden im Commit-Body
> begründet und aus den Pflichttests in §11 ableitbar gemacht.  
> **Bezug**: `docs/spike/0001-backend-stack.md` §6, `docs/plan-spike.md` §7.1, §12.3.  
> **Änderungen**: nur synchron mit dem Code in `apps/api/`; während
> des Spikes mussten beide Prototypen identisch sein, jetzt führt der
> Sieger-Code (`apps/api`) den Kontrakt.

Dieser Kontrakt ist die normative Schnittstelle, die beide Prototypen
(`spike/go-api`, `spike/micronaut-api`) identisch implementieren müssen.

---

## 1. Verbindliche Identifier

- **HTTP-Header**:
  - `X-MTrace-Token` — **Pflicht** (Auth, siehe §4)
  - `X-MTrace-Project` — reserviert für CORS-Allowlist, **im Spike nicht
    verlangt** (Plan §12.3, §14.x); `project_id` kommt aus dem Payload.
  - `Content-Type: application/json` — Pflicht für `POST`.
  - `Retry-After` — Server-Antwort bei `429`.
- **Prometheus-Metrik-Prefix**: `mtrace_`
- **OTel-Attribut-Prefix**: `mtrace.*`
- **Docker-Image**: `m-trace-api-spike:<stack>` mit `<stack>` ∈
  {`go`, `micronaut`}

---

## 2. HTTP-Endpunkte

| Methode | Pfad | Zweck | Erfolgs-Status |
|---|---|---|---|
| `POST` | `/api/playback-events` | Batch von 1–100 Events annehmen | `202 Accepted` |
| `GET`  | `/api/health`           | Liveness-Check                  | `200 OK`        |
| `GET`  | `/api/metrics`          | Prometheus-Exposition           | `200 OK`        |

Listen-/Detail-Endpunkte für Stream-Sessions sind **Bonus-Scope** (Spec
§7) und nicht Teil dieses Pflicht-Kontrakts.

---

## 3. Event-Schema (Wire-Format)

### 3.1 Beispielpayload für `POST /api/playback-events`

```json
{
  "schema_version": "1.0",
  "events": [
    {
      "event_name": "rebuffer_started",
      "project_id": "demo",
      "session_id": "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
      "client_timestamp": "2026-04-28T12:00:00.000Z",
      "sequence_number": 42,
      "sdk": {
        "name": "@m-trace/player-sdk",
        "version": "0.1.0"
      }
    }
  ]
}
```

### 3.2 Pflichtfelder pro Event

| Feld | Typ | Bedeutung |
|---|---|---|
| `event_name`        | string                       | Event-Typ (z. B. `rebuffer_started`) |
| `project_id`        | string                       | Projekt-Kennung; **muss zum `X-MTrace-Token` passen** |
| `session_id`        | string                       | Wiedergabe-Session, ULID oder UUID |
| `client_timestamp`  | string (RFC 3339, mit `Z`)   | Browser-Uhr; nur informell, nicht autoritativ |
| `sdk.name`          | string                       | SDK-Identifier |
| `sdk.version`       | string (SemVer)              | SDK-Version |

### 3.3 Optionale Felder pro Event

| Feld | Typ | Bedeutung |
|---|---|---|
| `sequence_number`     | int (≥ 0)                  | Monotone Reihenfolge pro Session |
| `server_received_at`  | string (RFC 3339, mit `Z`) | Server setzt das Feld; vom Client gesendete Werte werden verworfen |

### 3.4 Pflichtfelder im Batch-Wrapper

| Feld | Typ | Wert |
|---|---|---|
| `schema_version` | string                    | exakt `"1.0"` |
| `events`         | array of Event-Objekten   | Länge **1–100** |

Unbekannte Felder dürfen nicht zum Fehler führen (Vorwärtskompatibilität).

### 3.5 Antwort bei Erfolg

`POST /api/playback-events` antwortet mit `202 Accepted`:

```json
{
  "accepted": 1
}
```

`accepted` ist die Anzahl angenommener Events. Weitere Antwortfelder sind
nicht spezifiziert; Implementierungen dürfen sie ergänzen, müssen sich aber
abwärtskompatibel verhalten.

---

## 4. Authentifizierung

- Header `X-MTrace-Token` ist Pflicht.
- Token-Validierung gegen eine **hardcodierte Map** (Spec §6.4):

  ```json
  {
    "demo": "demo-token"
  }
  ```

  Schlüssel ist `project_id`, Wert ist das erwartete Token.

- Regeln:
  - Fehlt `X-MTrace-Token` → `401 Unauthorized`.
  - Token ungültig → `401 Unauthorized`.
  - `project_id` im Event passt nicht zum Token → `401 Unauthorized`.
  - `project_id` im Event ist nicht in der Map → `401 Unauthorized`.

- Im Spike gibt es **keine** dynamische Project-Verwaltung und **keinen**
  Endpunkt zum Anlegen/Rotieren von Tokens.

---

## 5. Validierungsregeln und Fehlerfälle

Reihenfolge der Validierung pro Request (Implementierungen müssen sich daran
halten, damit die Pflichttests deterministisch sind):

1. **Auth-Header**: `X-MTrace-Token` fehlt → `401 Unauthorized`. Diese
   Prüfung läuft im HTTP-Adapter, vor dem Body-Read, damit
   unauthentifizierte Requests einen Fast-Reject-Pfad erhalten und
   keine Body-Bandbreite konsumieren.
2. **Body-Größe**: > 256 KB → `413 Payload Too Large`.
3. **Auth-Token**: Token unbekannt → `401 Unauthorized`. Diese Prüfung
   läuft im Use Case (`ResolveByToken`).
4. **Rate-Limit** für `project_id` überschritten → `429 Too Many Requests`
   mit `Retry-After`-Header (Sekunden).
5. **Schema-Version**: `schema_version` ≠ `"1.0"` → `400 Bad Request`.
6. **Batch-Form**: `events` fehlt oder ist leer → `422 Unprocessable Entity`.
7. **Batch-Größe**: `events.length` > 100 → `422 Unprocessable Entity`.
8. **Event-Pflichtfelder**: ein Event ohne `event_name`, `project_id`,
   `session_id`, `client_timestamp` oder `sdk.{name,version}` → der
   gesamte Batch wird mit `422 Unprocessable Entity` abgelehnt.
9. **`project_id`/Token-Bindung**: ein Event mit `project_id` ≠ Token-
   Projekt → `401 Unauthorized` für den Batch.
10. **Erfolg**: Batch wird angenommen → `202 Accepted`.

Übersicht:

| Bedingung | Status |
|---|---:|
| Auth-Header fehlt                        | `401` |
| Body > 256 KB (mit Auth-Header)          | `413` |
| Token unbekannt                          | `401` |
| `project_id`/Token-Mismatch              | `401` |
| Rate-Limit überschritten                 | `429` + `Retry-After` |
| `schema_version` ≠ `"1.0"`               | `400` |
| `events` leer oder fehlt                 | `422` |
| `events.length` > 100                    | `422` |
| Event ohne Pflichtfeld                   | `422` |
| Valider Batch                             | `202` |

Folge der Auth-vor-Body-Reihenfolge: ein Request **ohne** Auth-Header
und mit Body > 256 KB liefert `401`, **nicht** `413` (siehe Pflichttest
in §11).

Antwort-Body bei Fehlerfällen ist **nicht** Teil des Pflicht-Kontrakts —
Implementierungen dürfen einen JSON-Body mit Fehlerbeschreibung senden,
müssen aber.

---

## 6. Rate Limiting

- **Quote**: 100 Events/Sekunde pro `project_id`.
- **Algorithmus**: Token-Bucket oder vergleichbar, in-memory, pro Prototyp.
- **Antwort bei Überschreitung**: `429 Too Many Requests` mit Header
  `Retry-After: <seconds>`.
- **Granularität**: Quote zählt **Events**, nicht Requests; ein Batch von
  50 Events verbraucht 50 Tokens.
- Verteiltes Rate-Limiting ist nicht Teil des Spikes.

---

## 7. Metriken (`GET /api/metrics`)

Format: Prometheus-Text-Exposition. Pflichtmetriken (Plan §5.4 +
Spec §6.6):

| Metrik | Typ | Bedeutung |
|---|---|---|
| `mtrace_playback_events_total`        | counter | angenommene Events (Status `202`) |
| `mtrace_invalid_events_total`         | counter | Events abgelehnt mit `400` oder `422` |
| `mtrace_rate_limited_events_total`    | counter | Events abgelehnt mit `429` |
| `mtrace_dropped_events_total`         | counter | intern verworfene Events (Backpressure) |

Verbindliche Regeln:

- Alle vier Counter zählen **Events**, nicht Batches. Bei einem Batch
  mit `events.length == 0`, der mit `422` abgelehnt wird, bleibt
  `mtrace_invalid_events_total` folglich unverändert — die Ablehnung
  ist über HTTP-Status und Access-Logs sichtbar, nicht über die Metrik.
- Auth-Fehler (`401`) laufen nicht in `mtrace_invalid_events_total`,
  weder der HTTP-seitige Header-Check noch Use-Case-seitiges
  `ResolveByToken` und Token-Bindung. `mtrace_invalid_events_total`
  ist auf `400` und `422` beschränkt.
- **Keine** hochkardinalen Labels: `session_id`, `user_agent`,
  `segment_url`, `client_ip` und beliebige `project_id` sind verboten.
- `mtrace_dropped_events_total` darf konstant `0` sein, wenn der Prototyp
  keinen expliziten Drop-Pfad hat — die Metrik **muss** aber existiert sein.
- Implementierungen dürfen weitere `mtrace_*`-Metriken ergänzen
  (z. B. `mtrace_active_sessions`), sofern Cardinality kontrolliert ist.

---

## 8. OpenTelemetry

- Im Spike-Muss-Scope: minimaler OTel-Setup im Code (Meter oder Tracer
  initialisiert, mindestens ein Counter oder Span erzeugt).
- Im Post-Spike-Soll: der Use Case spricht OTel ausschließlich über
  einen frameworkneutralen Driven Port (z. B. `Telemetry`) an —
  `hexagon/`-Pakete dürfen **nicht** direkt OTel importieren.
  Spans am Request-Boundary darf der HTTP-Adapter direkt erzeugen.
- Reader und Span-Exporter werden über
  `go.opentelemetry.io/contrib/exporters/autoexport` aufgelöst.
  Default ist silent (No-Op-Reader/-Exporter); OTLP-Anbindung
  aktiviert sich automatisch, sobald die Standard-OTel-Env-Vars
  gesetzt sind (`OTEL_EXPORTER_OTLP_ENDPOINT`,
  `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_TRACES_EXPORTER`,
  `OTEL_METRICS_EXPORTER`, …). Kein zusätzlicher Code-Pfad für
  „Dev vs. Prod".
- Konkrete Attribute und Resource-Konfiguration sind
  Implementierungs-Detail; bewertet wurde im Spike die **Ergonomie**
  der OTel-Integration im jeweiligen Stack (Bewertungsraster Spec §9).

---

## 9. Logging

- Format: strukturierte JSON-Logs.
- Pflichtfelder, sofern im Kontext vorhanden:
  - `project_id`
  - `session_id`
  - `status_code`
  - `error_type`
- Soll-Felder: `trace_id`, `request_id`.
- Lokale Test-Ausführung: Logs auf `stdout`.

---

## 10. Persistenz

- Im Spike: ausschließlich In-Memory.
- Daten überleben keinen Neustart — beabsichtigt.
- Keine SQLite, kein Redis, kein ORM.

---

## 11. Pflichttests pro Prototyp

Aus `docs/plan-spike.md` §7.1 (deckungsgleich mit Spec §6.12):

- Unit-Test `RegisterPlaybackEventBatch`: Happy Path
- Unit-Test zentrale Domain-Validierung: Pflichtfelder, Schema-Version
- Integrationstest `POST /api/playback-events` Happy Path mit gültigem Token
- Integrationstest `400` bei abweichender `schema_version`
- Integrationstest `401` bei fehlendem oder falschem Token
- Integrationstest `401` bei `project_id`/Token-Mismatch
- Integrationstest `401` bei unbekanntem `project_id`
- Integrationstest `413` bei Body über 256 KB (mit gültigem Auth-Header)
- Integrationstest `401` bei Body über 256 KB **ohne** Auth-Header — verifiziert die Auth-vor-Body-Reihenfolge aus §5
- Integrationstest `422` bei ungültigem Event (Pflichtfeld fehlt)
- Integrationstest `422` bei leerem oder fehlendem `events`-Feld
- Integrationstest `422` bei mehr als 100 Events im Batch
- Integrationstest `429` bei Rate-Limit-Überschreitung mit `Retry-After`-Header

---

## 12. Beispiele (curl)

### 12.1 Happy Path

```bash
curl -i -X POST http://localhost:8080/api/playback-events \
  -H 'Content-Type: application/json' \
  -H 'X-MTrace-Token: demo-token' \
  --data-binary @- <<'JSON'
{
  "schema_version": "1.0",
  "events": [
    {
      "event_name": "rebuffer_started",
      "project_id": "demo",
      "session_id": "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
      "client_timestamp": "2026-04-28T12:00:00.000Z",
      "sdk": { "name": "@m-trace/player-sdk", "version": "0.1.0" }
    }
  ]
}
JSON
```

Erwartet:

```text
HTTP/1.1 202 Accepted
Content-Type: application/json

{"accepted": 1}
```

### 12.2 Auth-Fehler (`401`)

```bash
curl -i -X POST http://localhost:8080/api/playback-events \
  -H 'Content-Type: application/json' \
  -H 'X-MTrace-Token: wrong-token' \
  --data-binary '{"schema_version":"1.0","events":[]}'
```

Erwartet: `HTTP/1.1 401 Unauthorized`.

### 12.3 Schema-Version-Fehler (`400`)

```bash
curl -i -X POST http://localhost:8080/api/playback-events \
  -H 'Content-Type: application/json' \
  -H 'X-MTrace-Token: demo-token' \
  --data-binary '{"schema_version":"2.0","events":[]}'
```

Erwartet: `HTTP/1.1 400 Bad Request`.

### 12.4 Health-Check

```bash
curl -i http://localhost:8080/api/health
```

Erwartet:

```text
HTTP/1.1 200 OK
```

Antwort-Body ist nicht spezifiziert; ein leeres `{}` oder `OK` reicht.

### 12.5 Prometheus-Metriken

```bash
curl http://localhost:8080/api/metrics | grep ^mtrace_
```

Erwartet: alle vier Pflichtmetriken aus §7 sichtbar.

---

## 13. Geltung und Versionsfortschreibung

- Diese Datei ist normativ ab dem Merge nach `main`.
- Vertragsänderungen während AP-1 oder AP-2 müssen
  - in beiden Prototypen identisch landen,
  - im Spike-Protokoll (`docs/spike/backend-stack-results.md`) begründet sein,
  - eine Änderung dieses Dokuments mit klarem Diff erzeugen.
- Schema-Version `1.0` ist der einzige im Spike akzeptierte Wert.
  Eine Erhöhung (z. B. auf `1.1` oder `2.0`) ist Phase-2-Thema.
