# Backend-API-Kontrakt

> **Status**: Verbindlich; Änderungen werden synchron mit dem Code in
> `apps/api/` gepflegt, im Commit-Body begründet und aus den
> Pflichttests in §11 ableitbar gemacht.
>
> **Bezug**: `docs/spike/0001-backend-stack.md` §6, `docs/planning/done/plan-spike.md` §7.1, §12.3.
> **Historie**: Dieses Dokument entstand im Backend-Spike für zwei
> Prototypen. Seit ADR-0001 (Accepted) ist es der laufende API-Kontrakt
> des Sieger-Codes (`apps/api`).

Dieser Kontrakt ist die normative Schnittstelle der m-trace API.

---

## 1. Verbindliche Identifier

- **HTTP-Header**:
  - `X-MTrace-Token` — **Pflicht** (Auth, siehe §4)
  - `X-MTrace-Project` — reserviert für CORS-Allowlist und spätere
    strengere Project-Bindung; `project_id` kommt im aktuellen
    Wire-Format aus dem Payload.
  - `Content-Type: application/json` — Pflicht für `POST`.
  - `traceparent` — **optional** ab `0.4.0` auf `POST /api/playback-events`
    (W3C Trace Context, [Spec](https://www.w3.org/TR/trace-context/)).
    Wenn vorhanden und valide, übernimmt der Server `trace_id` und
    `parent_span_id` aus dem Header. Bei ungültigem Header gibt es
    **kein** 4xx — der Server fällt auf eine eigene `trace_id` zurück
    und setzt das Span-Attribut `mtrace.trace.parse_error=true`
    (siehe `spec/telemetry-model.md` §2.5). Der Header-Name ist
    case-insensitiv; der Header-Wert wird als einzelner W3C-Wert
    interpretiert. Das Verhalten bei führender/abschließender OWS wird
    im `plan-0.4.0.md`-§3.4c-Closeout gegen Code und Tests finalisiert.
  - `Retry-After` — Server-Antwort bei `429`.
- **Prometheus-Metrik-Prefix**: `mtrace_`
- **OTel-Attribut-Prefix**: `mtrace.*`

---

## 2. HTTP-Endpunkte

| Methode | Pfad | Zweck | Erfolgs-Status |
|---|---|---|---|
| `POST` | `/api/playback-events` | Batch von 1–100 Events annehmen | `202 Accepted` |
| `GET`  | `/api/health`           | Liveness-Check                  | `200 OK`        |
| `GET`  | `/api/metrics`          | Prometheus-Exposition           | `200 OK`        |
| `GET`  | `/api/stream-sessions`  | Stream-Sessions listen          | `200 OK`        |
| `GET`  | `/api/stream-sessions/{id}` | Stream-Session mit Events lesen | `200 OK` oder `404 Not Found` |
| `POST` | `/api/analyze`          | HLS-Manifest analysieren (plan-0.3.0 §7) | `200 OK` |

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
        "name": "@npm9912/player-sdk",
        "version": "0.2.0"
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

Ab `plan-0.4.0.md` Tranche 3 darf der Batch optional
`session_boundaries` enthalten. Dieser Block ist kein Event-Stream,
zählt nicht in `accepted`, besitzt kein `event_name` und ändert
`schema_version: "1.0"` nicht. Boundary-only-Batches ohne `events`
bleiben ungültig.

```json
{
  "schema_version": "1.0",
  "events": [{ "...": "PlaybackEvent" }],
  "session_boundaries": [
    {
      "kind": "network_signal_absent",
      "project_id": "demo",
      "session_id": "01J7K9X4Z2QHB6V3WS5R8Y4D1F",
      "network_kind": "segment",
      "adapter": "native_hls",
      "reason": "native_hls_unavailable",
      "client_timestamp": "2026-04-28T12:00:00.000Z"
    }
  ]
}
```

Für Tranche 3 ist nur `kind="network_signal_absent"` definiert.
`network_kind` ist `"manifest"` oder `"segment"`, `adapter` ist
`"hls.js"`, `"native_hls"` oder `"unknown"`, `reason` ist einer aus
`{"native_hls_unavailable", "hlsjs_signal_unavailable",
"browser_api_unavailable", "resource_timing_unavailable",
"cors_timing_blocked", "service_worker_opaque"}` und muss zusätzlich
`^[a-z0-9_]{1,64}$` erfüllen. Andere Werte, rohe URLs, Token-Strings
oder HTML/Script-Fragmente werden mit `422 Unprocessable Entity`
abgelehnt. `project_id` muss wie bei Events zum `X-MTrace-Token`
passen; `session_id` ist Pflicht.

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

### 3.6 Analyzer-Endpunkt `POST /api/analyze`

`POST /api/analyze` reicht eine HLS-Manifest-Analyse an den
internen `analyzer-service` weiter (plan-0.3.0 §7 Tranche 6) und
gibt das `AnalysisResult`-JSON aus `@npm9912/stream-analyzer`
zurück. Der Endpunkt ist authentifizierungsfrei in 0.3.0 — der
Service ist nur über das interne Netz erreichbar; ein öffentlich
exponierter Deploy braucht eine Egress-Firewall oder einen
Folge-ADR mit Token-Schicht.

**Request** (`Content-Type: application/json`):

```json
{ "kind": "url", "url": "https://cdn.example.test/manifest.m3u8" }
```

oder

```json
{ "kind": "text", "text": "#EXTM3U\n…", "baseUrl": "https://cdn.example.test/" }
```

Pflicht: `kind` (`"url" | "text"`). Bei `kind="url"` ist `url`
Pflicht; bei `kind="text"` ist `text` Pflicht und `baseUrl`
optional. Body-Limit auf API-Ebene: 1 MiB (Defense-in-Depth; der
analyzer-service hat sein eigenes Limit beim Manifest-Loading).

Ab `plan-0.4.0.md` Tranche 3 darf der Request optional eine
Session-Bindung tragen:

```json
{
  "kind": "url",
  "url": "https://cdn.example.test/manifest.m3u8",
  "correlation_id": "2f6f1a3c-9fb9-4c0b-a78f-2f41d8f6e1e7",
  "session_id": "01J7K9X4Z2QHB6V3WS5R8Y4D1F"
}
```

Verlinkung ist nur mit gültigem Project-Kontext erlaubt. Ein Request,
der `correlation_id` oder `session_id` setzt, muss `X-MTrace-Token`
erfolgreich auf ein `project_id` auflösen (und später, falls aktiv,
`X-MTrace-Project` konsistent dazu liefern). Fehlt dieser Kontext bei
gesetzten Link-Feldern oder ist der Token ungültig, antwortet die API
mit dem Auth-/Kontextfehler aus §5 (`401 Unauthorized`) und führt
keinen Session-Lookup aus. Nur Requests ohne Link-Felder dürfen ohne
Project-Kontext erfolgreich bleiben; sie erhalten
`session_link.status="detached"`.

`correlation_id` hat innerhalb dieses Project-Kontexts Vorrang vor
`session_id`. `correlation_id` allein ohne Treffer im Project liefert
`session_link.status="not_found_detached"`. Wenn beide Felder gesetzt
sind, muss zuerst `correlation_id` im Project existieren; eine bekannte
`session_id` darf eine unbekannte oder project-fremde `correlation_id`
nicht retten. Existiert die `correlation_id`, muss `session_id` im
gleichen Project zur Session mit dieser `correlation_id` auflösen; bei
Mismatch bleibt das Analyzer-Ergebnis eine unabhängige Manifestanalyse
und wird nicht in die Player-Timeline gemischt. Die API bleibt `200 OK`,
wechselt ab Tranche 3 aber auf eine Hülle:

```json
{
  "analysis": { "...": "AnalysisResult" },
  "session_link": { "status": "conflict_detached" }
}
```

Kompatibilitätsentscheidung: Ab Tranche 3 gibt `POST /api/analyze`
für alle erfolgreichen Requests diese Hülle zurück, auch wenn der
Request keine Link-Felder enthält. Ungebundene Requests erhalten
`session_link.status="detached"`; es gibt kein bedingtes
Response-Shape-Branching.

`session_link.status` ist eines aus `{"linked", "detached",
"conflict_detached", "not_found_detached"}`. Nur `session_id` ist als
Fallback zulässig, wenn sie auf eine bestehende oder bereits
selbst-geheilte Session auflösbar ist. Eine unbekannte `session_id`
erzeugt keine neue Session und liefert `not_found_detached`. Ohne beide
Link-Felder ist die Analyse bewusst session-los (`detached`). Alle
Link-Lookups verwenden
`(project_id, correlation_id)` bzw. `(project_id, session_id)`. Diese
Bindungsfelder ändern das Analyzer-`AnalysisResult` nicht; sie steuern
nur die optionale Dashboard-/Timeline-Verknüpfung.

**Erfolgsantwort** (`200 OK`): bis einschließlich `0.3.x`
vollständiges `AnalysisResult` aus `docs/user/stream-analyzer.md` §2.2.
Ab `plan-0.4.0.md` Tranche 3 wird dieses Resultat bei jedem
erfolgreichen Request unverändert unter `analysis` in der oben
beschriebenen Hülle transportiert.

Session-Read-Pfade sind ab Tranche 3 projekt-skopiert und
authentifiziert: Session-Liste, Session-Detail, Event-Reads und
Cursor-Reuse müssen `X-MTrace-Token` erfolgreich auf ein `project_id`
auflösen. Fehlender oder ungültiger Token liefert `401 Unauthorized`.
Der aufgelöste `project_id` ist Filter für alle Read-Pfade; Cursor aus
einem Project dürfen nicht für ein anderes Project akzeptiert werden.

**Fehler-Mapping** (Problem-Shape `{status, code, message, details?}`):

API-Eingabevalidierung (Request-Form):

| HTTP | `code`                  | Anlass                                                                  |
| ---- | ----------------------- | ----------------------------------------------------------------------- |
| 400  | `invalid_request`       | Pflichtfelder fehlen / kind unbekannt / leerer `text`/`url`-Wert.       |
| 400  | `invalid_json`          | Body ist kein gültiges JSON.                                            |
| 415  | `unsupported_media_type`| `Content-Type` ist nicht `application/json`.                            |
| 413  | `payload_too_large`     | Request-Body übersteigt 1 MiB.                                          |

Analyzer-Domain-Fehler (analyzer-service hat den Aufruf bewusst
abgelehnt; der `code` stammt aus `@npm9912/stream-analyzer` und
wird durchgereicht; `details` enthält strukturierte Zusatzinfos
aus dem Analyzer-Result, nicht die freie Adapter-Message):

| HTTP | `code`                | Anlass                                                                                |
| ---- | --------------------- | ------------------------------------------------------------------------------------- |
| 400  | `invalid_input`       | Analyzer hat die Manifest-Eingabe als formal ungültig zurückgewiesen.                 |
| 400  | `fetch_blocked`       | Analyzer-SSRF-Schutz hat die URL abgelehnt (privat/loopback/Credentials/Schema).      |
| 422  | `manifest_not_hls`    | Geladenes Manifest ist kein HLS-Inhalt — Eingabe semantisch nicht verarbeitbar.       |
| 502  | `fetch_failed`        | Analyzer konnte die URL nicht laden (Netzwerk, Status, Content-Type).                 |
| 502  | `manifest_too_large`  | Geladenes Manifest übersteigt das Loader-Größenlimit.                                 |
| 502  | `internal_error`      | Unerwarteter Fehler im Analyzer-Stack.                                                |

Transport- und Verfügbarkeitsfehler (analyzer-service nicht
erreichbar, JSON-Decode, Antwort über Größenlimit, fremder
HTTP-Status):

| HTTP | `code`                  | Anlass                                                                |
| ---- | ----------------------- | --------------------------------------------------------------------- |
| 502  | `analyzer_unavailable`  | analyzer-service nicht erreichbar, lieferte malformed JSON, oder gab einen unerwarteten HTTP-Status. Der Antwort-Body trägt **keine** rohe Adapter-Fehler-Message; Details landen strukturiert im API-Log. |

Der analyzer-service-Pfad bekommt einen 30-Sekunden-Timeout vom
HTTP-Adapter sowie ein Antwortgrößen-Limit von 4 MiB. Beides ist
Defense-in-Depth gegen einen kompromittierten oder hängenden
Service; die Limits sind nicht öffentlich konfigurierbar.

---

### 3.7 Server-vergebene Read-Felder

Die folgenden Felder werden ausschließlich vom Server vergeben und
erscheinen in den Read-Antworten von `GET /api/stream-sessions/{id}`:

| Feld | Typ | Verfügbar ab | Beschreibung |
|---|---|---|---|
| `ingest_sequence` | `int64`, ≥ 1, monoton steigend, global eindeutig | `0.1.x` | Durable Persistenz-Sequenz, durch das Storage-Backend vergeben (siehe §10.1, §10.4 und [ADR 0002 §8.1](../docs/adr/0002-persistence-store.md)). Tie-Breaker der kanonischen Event-Sortierung. |
| `delivery_status` | `string` aus `{"accepted", "duplicate_suspected", "replayed"}` | `0.4.0` (ab `plan-0.4.0.md` §2.3-Closeout) | Timeline-Klassifikation jedes Events; siehe §10.2. Default ist `"accepted"`. Vor §2.3-Closeout liefern Read-Antworten dieses Feld nicht. |
| `correlation_id` | `string` (UUIDv4 oder vergleichbar), **nicht-leer in ab §3.2 verarbeiteten Events**; bei vor §3.2-Closeout persistierten Events kann der Wert `""` sein (Read-Pfad liefert ihn dann als JSON-`""`, siehe Migrations-Hinweis unten) | `0.4.0` (ab `plan-0.4.0.md` §3.2-Closeout) | Server-generierte, durable Source-of-Truth für die Tempo-unabhängige Dashboard-Korrelation einer Session. Konstant über alle ab §3.2 verarbeiteten Events derselben Session; auch in der Session-Header-Response exposed (siehe §3.7.1). Siehe `spec/telemetry-model.md` §2.5. |
| `trace_id` | `string`, 32 Hex-Zeichen, optional (`null` zulässig wenn weder `traceparent` noch Server-Trace gesetzt — Edge-Case) | `0.4.0` (ab `plan-0.4.0.md` §3.2-Closeout) | W3C-Trace-ID des Batches, in dem das Event registriert wurde. Vom SDK propagiert (`traceparent`-Header, siehe §1) oder server-generiert. Primär für Tempo-Cross-Trace-Suche; Dashboard-Korrelation läuft über `correlation_id`. |

Diese vier Felder sind im POST-Wire-Format (§3.2/§3.3) **nicht** zulässig;
Clients dürfen sie nur aus Read-Antworten interpretieren. Die genaue
Vertragssemantik (Sortierung, Idempotenz, Cursor) steht in §10;
Trace-Korrelations-Vertrag in `spec/telemetry-model.md` §2.5.

**Migration von Pre-§3.2-Persistenz**: Sessions und Events, die vor
`0.4.0`-§3.2 angelegt wurden, haben kein `correlation_id`. Tranche 2
führt **kein historisches Event-Backfill** aus: ältere
`playback_events.correlation_id`-Leerwerte bleiben im Read-Pfad als
JSON-`""` sichtbar und sind ein degradierter Legacy-Fall. Der Use-Case
führt beim nächsten Event einer solchen Session ein Self-Healing durch
(siehe `resolveCorrelationIDs` in der Application-Schicht), das die
Session-`correlation_id` einmalig nachträglich setzt und die neu
persistierten Events mit dieser `correlation_id` schreibt. Clients
sollten leere `correlation_id`-Felder bei historischen Events als
„vor §3.2 nicht gesetzt" interpretieren — nicht als Vertragsbruch.

#### 3.7.1 Session-Header-Read-Felder

Die Session-Header-Antwort von `GET /api/stream-sessions` und
`GET /api/stream-sessions/{id}` (Session-Block, nicht Event-Block)
trägt ab `0.4.0` (§3.2-Closeout) zusätzlich:

| Feld | Typ | Beschreibung |
|---|---|---|
| `correlation_id` | `string`; nicht-leer für ab §3.2 angelegte oder bereits selbst-geheilte Sessions, sonst `""` als Legacy-Fall | Spiegelt `stream_sessions.correlation_id`; identisch mit dem `correlation_id`-Wert auf ab §3.2 persistierten Events derselben Session. Historische Events vor §3.2 werden nicht backfilled. Dient dem Dashboard als primärer Korrelations-Schlüssel — Tempo-unabhängig. |

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

- Dynamische Project-Verwaltung und Endpunkte zum Anlegen oder Rotieren
  von Tokens sind nicht Teil dieses Kontrakts.

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

Der `traceparent`-Header (siehe §1) ist **nicht** Teil dieser
Validierungs-Reihenfolge: ein ungültiger Wert führt nie zu `4xx`,
sondern wird über das Span-Attribut `mtrace.trace.parse_error=true`
markiert (Vertrag in `spec/telemetry-model.md` §2.5).

Antwort-Body bei Fehlerfällen ist **nicht** Teil des Pflicht-Kontrakts —
Implementierungen dürfen einen JSON-Body mit Fehlerbeschreibung senden,
müssen aber.

Cursor-Endpunkte (`GET /api/stream-sessions`, `GET /api/stream-sessions/{id}`)
folgen einer eigenen Validierungs- und Fehlerklassen-Matrix; siehe §10.3.

---

## 6. Rate Limiting

- **Quote**: 100 Events/Sekunde **pro Dimension**. Drei unabhängige
  Dimensionen werden geprüft (plan-0.1.0.md §5.1, F-110): `project_id`,
  `client_ip`, `origin`. Mismatch in einer Dimension reicht für `429`;
  ein 429 in einer Dimension verbraucht keine Tokens in den anderen
  („all-or-nothing"-Commit).
- **Algorithmus**: Token-Bucket pro (Dimension, Wert), in-memory, pro
  API-Prozess. Capacity und Refill teilen sich alle Dimensionen.
- **Antwort bei Überschreitung**: `429 Too Many Requests` mit Header
  `Retry-After: <seconds>`.
- **Granularität**: Quote zählt **Events**, nicht Requests; ein Batch
  von 50 Events verbraucht 50 Tokens pro gesetzter Dimension.
- **Leere Dimensions-Werte** (z. B. `origin=""` im CLI/curl-Pfad) werden
  übersprungen — keine künstlichen Sentinels, kein Rate-Limit gegen den
  leeren String.
- Verteiltes Rate-Limiting ist nicht Teil dieses Kontrakts.

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
- `mtrace_dropped_events_total` darf konstant `0` sein, wenn die API
  keinen expliziten Drop-Pfad hat — die Metrik **muss** aber existiert sein.
- `delivery_status: "duplicate_suspected"`-Events (siehe §10.2) zählen
  zu `mtrace_playback_events_total` (sie sind angenommen, nur als
  Duplikat klassifiziert) und **nicht** zu
  `mtrace_invalid_events_total` oder `mtrace_dropped_events_total`.
- Implementierungen dürfen weitere `mtrace_*`-Metriken ergänzen
  (z. B. `mtrace_active_sessions`), sofern Cardinality kontrolliert ist.

---

## 8. OpenTelemetry

- Der Use Case spricht OTel ausschließlich über einen
  frameworkneutralen Driven Port (z. B. `Telemetry`) an — `hexagon/`
  Pakete dürfen **nicht** direkt OTel importieren.
- Spans am Request-Boundary darf der HTTP-Adapter direkt erzeugen.
- Reader und Span-Exporter werden über
  `go.opentelemetry.io/contrib/exporters/autoexport` aufgelöst,
  jeweils mit explizitem No-Op-Fallback
  (`autoexport.WithFallbackMetricReader` /
  `autoexport.WithFallbackSpanExporter`) — sonst defaultet autoexport
  auf OTLP, sobald die Env-Vars unset sind. Mit Fallback gilt:
  ohne Env-Vars silent; mit `OTEL_TRACES_EXPORTER=otlp` /
  `OTEL_METRICS_EXPORTER=otlp` (oder weiteren Standard-OTel-Env-Vars)
  registriert autoexport den entsprechenden Exporter. Kein
  zusätzlicher Code-Pfad für „Dev vs. Prod".
- Konkrete Attribute und Resource-Konfiguration sind
  Implementierungs-Detail; die verbindlichen Telemetrie-Attribute stehen
  in [`telemetry-model.md`](./telemetry-model.md).

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

`0.1.x`–`0.3.x` nutzten In-Memory-Repositories (Datenverlust bei
Neustart, beabsichtigt). Ab `0.4.0` ist der lokale Durable-Store
SQLite (siehe [ADR 0002](../docs/adr/0002-persistence-store.md)). Die
nachfolgenden Sub-Sections sind Vertrag gegenüber API-Konsumenten —
sie beschreiben das beobachtbare Verhalten, nicht die interne
Implementierung.

### 10.1 Storage-Stand

- Sessions, Playback-Events und Ingest-Sequenzen werden in einer
  lokalen SQLite-Datei persistiert; ein API-Restart verliert keine
  bereits angenommenen Sessions oder Events.
- Reset des lokalen Storage geschieht ausschließlich über das
  dedizierte `make wipe`-Target (siehe `docs/user/local-development.md`);
  `make stop` räumt nicht auf. Andere Reset-Pfade (manuelles Löschen
  des Volumes, etc.) sind nicht Teil des Kontrakts.
- Postgres und andere Stores sind in `0.4.0` nicht im Scope (Folge-ADR
  aus Roadmap §4).

### 10.2 Idempotenz und Event-Deduplikation

- **Session-State-Updates** sind idempotent. Insbesondere ist
  `session_ended` vom Client mehrfach sendbar; der Server setzt das
  Session-Ende beim ersten Eintreffen und wertet nachfolgende
  Wiederholungen als no-op (Antwort bleibt `202 Accepted`, kein
  Fehlerbody).
- **Event-Deduplikation** erfolgt server-seitig als
  Timeline-Klassifikation, nicht als Hard-Reject:
  - Dedup-Key: `(project_id, session_id, sequence_number)` für Events
    mit gesetzter `sequence_number` (siehe §3.3).
  - Trifft ein Event mit demselben Dedup-Key auf einen bereits als
    `accepted` persistierten Vorgänger, wird das neue Event ebenfalls
    persistiert und im Read-Pfad mit `delivery_status: "duplicate_suspected"`
    ausgeliefert.
  - Events ohne `sequence_number` werden immer als
    `delivery_status: "accepted"` aufgenommen; ohne expliziten
    Dedup-Schlüssel führt der Server keine automatische Erkennung
    durch.
- Möglicher `delivery_status`-Wertebereich im Read-Pfad:
  `accepted`, `duplicate_suspected`, `replayed`. `replayed` ist in
  `0.4.0` reserviert und wird nur durch explizite Use-Case-Pfade
  gesetzt.
- POST-Antworten (`202 Accepted`) ändern sich durch die
  Dedup-Klassifikation **nicht**: jeder im Batch enthaltene Event
  zählt für `accepted` im Response-Body und für die
  `mtrace_playback_events_total`-Metrik (Cardinality-Regeln aus §7
  bleiben gültig).

### 10.3 Pagination und Cursor

Cursor-basierte Pagination gilt für `GET /api/stream-sessions`
(Query-Parameter `cursor`) und für die Event-Liste in
`GET /api/stream-sessions/{id}` (Query-Parameter `events_cursor`).
Andere Query-Parameter-Namen oder Aliasse sind nicht Teil des
Kontrakts.

- **Wire-Format**: Cursor-Tokens sind base64url-kodiertes JSON ohne
  Padding und enthalten ab `0.4.0` ein verbindliches `v`-Feld
  (Cursor-Version). Aktuelle Version ist bis Tranche 2 `2`; mit
  `plan-0.4.0.md` Tranche 3 wechseln Session-List- und
  Session-Event-Cursor wegen projekt-skopierter Read-Pfade auf `3`.
  v3-Cursor enthalten den Project-Scope (`project_id` oder einen
  daraus abgeleiteten Scope-Hash) zusätzlich zur Storage-Position.
  Token-Inhalt ist servergetragen und sollte vom Client als opak
  behandelt werden.
- **Versionierung**: Cursor ohne `v`-Feld oder mit `v: 1` werden als
  Legacy-Format (`process_instance_id`-basiert, `0.1.x`/`0.2.x`/`0.3.x`)
  erkannt und dauerhaft abgewiesen. Nach Aktivierung der
  projekt-skopierten Read-Pfade werden auch v2-Session-Cursor ohne
  Project-Scope dauerhaft als Legacy abgewiesen. Die feingranulare
  Begründung steht in [ADR 0004 — Cursor-Strategie](../docs/adr/0004-cursor-strategy.md).

**Kompatibilitätsmatrix**:

| Klasse | Erkennung | HTTP-Status | Body | Client-Recovery |
|---|---|---|---|---|
| `accepted` | Token decodiert; `v == 2` vor Tranche 3 oder `v == 3` ab Tranche 3; alle Pflichtfelder vorhanden und valide; bei v3 passt der Project-Scope zum Request-Kontext. | `200 OK`. | regulärer Listen-Response inkl. `next_cursor`. | weiter paginieren mit `next_cursor`. |
| `cursor_invalid_legacy` | Token decodiert; `v`-Feld fehlt oder enthält `1`; oder `pid`-Feld vorhanden; nach Aktivierung projekt-skopierter Read-Pfade auch `v == 2` für Session-Cursor ohne Project-Scope. | `400 Bad Request`. | `{"error":"cursor_invalid_legacy","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |
| `cursor_invalid_malformed` | Base64- oder JSON-Decode schlägt fehl; oder `v`-Feld enthält unbekannten Wert; oder Pflichtfeld fehlt/Format ungültig; oder unbekannte Zusatzfelder vorhanden; oder v3-Project-Scope passt nicht zum Request-Kontext. | `400 Bad Request`. | `{"error":"cursor_invalid_malformed","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |
| `cursor_expired` | Cursor decodiert valide; Token-Inhalt referenziert aber eine Storage-Position, die durch Reset/Retention nicht mehr existiert. In `0.4.0` ohne TTL praktisch nur nach `make wipe` erreichbar. | `410 Gone` (Token syntaktisch valide, Ziel weg). | `{"error":"cursor_expired","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |

**Recovery-Verhalten**:

- Keine der Fehlerklassen enthält `Retry-After`. Ein Retry-Loop mit
  demselben Cursor ist ein Client-Fehler.
- `cursor_invalid_legacy` ist eine **dauerhafte** Reject-Klasse. Der
  einzelne Legacy-Cursor wird nicht „einmalig" akzeptiert; nach dem
  ersten `400` muss der Client den Snapshot neu laden und den Cursor
  vergessen.

### 10.4 Kanonische Sortierung

API-Listen sind **restart-stabil** sortiert. Die Reihenfolge wird vom
Server garantiert und ist nicht durch Cursor-Verhalten überspielbar:

| Endpoint | Sortier-Tupel | Tie-Breaker |
|---|---|---|
| `GET /api/stream-sessions` | `started_at desc`, `session_id asc`. | `session_id` ist innerhalb `project_id` eindeutig (siehe §1). |
| `GET /api/stream-sessions/{id}` (Events) | `server_received_at asc`, `sequence_number asc` (falls vorhanden), `ingest_sequence asc`. | `ingest_sequence` ist global monoton steigend und durable (siehe ADR 0002 §8.1); damit eindeutig auch ohne `project_id`/`session_id`-Komposit. |

`ingest_sequence` ist serverseitig pflichtend und überlebt API-Restart.
Damit ist die Event-Reihenfolge zweier Listen-Aufrufe vor und nach
einem Restart bei identischem Cursor identisch (sofern keine neuen
Events angenommen wurden).

### 10.5 Retention

- `0.4.0` führt keine automatische Retention ein. Daten bleiben
  erhalten, bis ein expliziter Reset (siehe §10.1) erfolgt.
- Konkrete TTL- oder Pro-Projekt-Limits werden Folge-Arbeit, sobald
  Multi-Tenant-Last entsteht; bis dahin gibt der Server keinen
  Retention-Header aus.
- `cursor_expired` (§10.3) ist in `0.4.0` ohne TTL effektiv nur durch
  `make wipe` erreichbar — Server-Implementierung muss den Pfad aber
  vorsehen, damit Clients Retention-Folge-Arbeit ohne Wire-Format-
  Bruch unterstützen können.

---

## 11. Pflichttests für die API

Ursprünglich aus `docs/planning/done/plan-spike.md` §7.1 abgeleitet; weiterhin
Pflichtabdeckung für den Ingest-Pfad:

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
      "sdk": { "name": "@npm9912/player-sdk", "version": "0.2.0" }
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
- Vertragsänderungen müssen synchron mit `apps/api`, den Tests und den
  maschinenlesbaren Contract-Artefakten gepflegt werden.
- Schema-Version `1.0` ist der aktuell akzeptierte Wert. Eine Erhöhung
  (z. B. auf `1.1` oder `2.0`) erfordert eine Aktualisierung von
  `contracts/event-schema.json`, `contracts/sdk-compat.json`, API und SDK
  im selben Commit.
