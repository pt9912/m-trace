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
  - `X-MTrace-Token` — Auth-Token; Pflicht je Endpoint gemäß §4.
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
    HTTP-konform case-insensitiv (`Traceparent`, `traceparent`,
    `TRACEPARENT` sind derselbe Header); SDKs schreiben den Namen
    lowercased. Der Header-Wert ist genau ein einzelner W3C-`traceparent`-
    Wert (55 Zeichen, Form `00-<32 hex>-<16 hex>-<2 hex>`). Führende
    und nachfolgende OWS (Spaces, Tabs) werden vom HTTP-Wire-Layer
    der Go-`net/http`-Standardbibliothek bereits beim Header-Lesen
    entfernt, bevor der Wert das Backend erreicht; ein OWS-umschlossener,
    sonst valider Wert wird daher als gültig behandelt und führt zur
    normalen Child-Span-Übernahme. Das Backend führt selbst kein
    zusätzliches Trim durch und verlässt sich für die OWS-Normalisierung
    ausschließlich auf den Wire-Layer; ein durchgereichter OWS-Wert
    (z. B. von einem Reverse-Proxy mit abweichender Header-Verarbeitung)
    fällt am defensiven `len == 55`-Check des Parsers auf den
    parse_error-Pfad zurück. Test-Anker:
    `TestHTTP_Span_TraceParent_LeadingTrailingWhitespace` für die
    Wire-Beobachtung und `TestParseTraceParent_Invalid` für den
    Defense-in-Depth-Pfad bei direktem Funktionsaufruf.
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
| `GET`  | `/api/stream-sessions/stream` | SSE-Live-Stream der Event-Append-Frames (plan-0.4.0 §5 H4) | `200 OK` (text/event-stream) oder `401` |
| `POST` | `/api/analyze`          | HLS-Manifest analysieren (plan-0.3.0 §7) | `200 OK` |
| `POST` | `/api/ingest/streams`   | Ingest-Stream anlegen; gibt Stream-Metadaten plus Klartext-Key genau einmal zurück (plan-0.11.0 §0.6, RAK-66) | `201 Created` |
| `GET`  | `/api/ingest/streams`   | Streams im aufgelösten Project listen, ohne Klartext-Key (RAK-66/RAK-70) | `200 OK` |
| `GET`  | `/api/ingest/streams/{id}` | Stream-Details inkl. Endpunkt und Routing-Regel lesen, ohne Klartext-Key (RAK-67) | `200 OK` oder `404 Not Found` |
| `POST` | `/api/ingest/streams/{id}/rotate-key` | Stream-Key rotieren; gibt neuen Klartext-Key genau einmal zurück (RAK-66) | `200 OK` |
| `POST` | `/api/ingest/streams/{id}/validate-key` | Lokalen Stream-Key gegen aktive `key_hash`-Werte prüfen; **kein** produktiver Media-Server-Auth-Pfad (RAK-65/RAK-66) | `200 OK` |
| `POST` | `/api/ingest/hooks/stream-started` | Lokales Start-Event empfangen oder Smoke-Event einspeisen (RAK-69) | `202 Accepted` |
| `POST` | `/api/ingest/hooks/stream-ended` | Lokales Ende-Event empfangen oder Smoke-Event einspeisen (RAK-69) | `202 Accepted` |
| `GET`  | `/api/ingest/media-server-config` | Generiertes/validiertes MediaMTX-Artefakt abrufen oder Diagnose liefern (RAK-68) | `200 OK` |
| `POST` | `/api/auth/session-tokens` | Kurzlebiges Session Token aus gültigem Project Token ausstellen (plan-0.12.0 §0.5, RAK-72) | `201 Created` |

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
bleiben ungültig. Es sind maximal 20 Boundaries pro Batch erlaubt.
Boundaries zählen in dasselbe `max_body_bytes`-Budget wie Events. Jede
Boundary muss eine `(project_id, session_id)`-Partition referenzieren,
für die im selben Batch mindestens ein Event vorhanden ist; andernfalls
ist der Batch `422 Unprocessable Entity`.

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
`"hls.js"`, `"native_hls"` oder `"unknown"`. Die zulässige Domäne und
das Längen-/Charset-Pattern für `reason` sind normativ in
`spec/telemetry-model.md` §1.4 definiert (gemeinsamer Reason-Enum mit
`meta["network.unavailable_reason"]`); `contracts/event-schema.json`
spiegelt sie für maschinenlesbare Validierung über
`session_boundaries.reasons_ref` und `session_boundaries.reason_pattern_ref`,
die auf `network_unavailable_reasons` bzw.
`network_unavailable_reason_pattern` zeigen. Andere Werte, rohe URLs,
Token-Strings oder HTML/Script-Fragmente werden mit
`422 Unprocessable Entity` abgelehnt. `project_id` muss wie bei Events
zum `X-MTrace-Token` passen; `session_id` ist Pflicht.

Der komplette Batch-Wrapper wird vor jedem Write validiert oder
gemeinsam transaktional persistiert. Ein invalider Boundary-Block
persistiert weder Events noch Boundaries und erhöht `accepted` nicht.

#### 3.4a Reservierter `webrtc.*`-Meta-Namespace (`0.8.0`)

Ab `plan-0.8.0.md` Tranche 3 ist `webrtc.*` ein reservierter Meta-
Namespace; der vollständige Schlüsselsatz, die Wertedomänen und die
Counter-Semantik sind normativ in `spec/telemetry-model.md` §1.4
und §3.5 verankert. `contracts/event-schema.json` (`reserved_meta_keys`
und `reserved_meta_namespace_webrtc`) spiegelt die Allowlist für
maschinenlesbare Validierung. Pflichtverhalten des API-Ingress:

- Jeder `webrtc.*`-Schlüssel muss in der Allowlist stehen, mit dem
  dort dokumentierten Typ und (bei Strings) der dort dokumentierten
  Enum-/Pattern-Domäne. Verstöße liefern `422 Unprocessable Entity`
  und werden nicht persistiert; eine `mtrace_webrtc_*`-Metrik wird
  nicht erzeugt.
- Per-Identifier-Felder aus `spec/telemetry-model.md` §3.1
  (`webrtc.track_id`, `webrtc.candidate_pair_id`, `webrtc.ssrc`,
  `webrtc.user_agent`, weitere) sind explizit verboten und liefern
  `422`.
- Nicht-`webrtc.*`-Meta-Keys bleiben gemäß additiver Forward-
  Compatibility-Regel (§3.4) unangetastet — alte Backends ignorieren
  unbekannte additive Keys, neue SDK-Versionen dürfen sie ergänzen.

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
SSE-Read-Pfade aus Tranche 4 folgen derselben Auth-Regel; ihre
Preflight-Routen müssen `GET, OPTIONS` und die Header `X-MTrace-Token`
`X-MTrace-Project` und `Last-Event-ID` erlauben. Fetch-basierte
SSE-Reconnects übertragen die Backfill-Position über `Last-Event-ID`.

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

### 3.8 Ingest-Control-Endpunkte (`0.11.0`, RAK-65..RAK-70)

Der `/api/ingest/*`-Pfad implementiert lokales Stream Control für
Lab-/Demo-Flows (`apps/api`-Modul, Variante B aus
[`docs/planning/done/plan-0.11.0.md`](../docs/planning/done/plan-0.11.0.md)
§0.3). Der Pfad ist **kein** produktiver Auth-Replacement und
**kein** mandantenfähiger Control-Plane-Pfad — Out-of-Scope-Liste
in [`docs/user/ingest-control.md`](../docs/user/ingest-control.md)
§5 und [`docs/planning/done/plan-0.11.0.md`](../docs/planning/done/plan-0.11.0.md)
§0.1.

**Auth-Matrix.** Alle `/api/ingest/*`-Endpunkte sind tokenpflichtig
und folgen der bestehenden `X-MTrace-Token`-/Project-Resolver-
Konvention aus §4. `project_id` wird serverseitig aus dem Token
abgeleitet; ein Request-`project_id`-Wert (z. B. in
`POST /api/ingest/streams`) darf nur als Konsistenzcheck dienen
und muss zum Token passen. Listen, Details, Rotation, Key-
Validierung und Lifecycle-Events sind immer auf das aufgelöste
Project gefiltert; ein Stream eines fremden Projects wird wie
nicht-existent behandelt (`404`, kein leakender Hinweis auf
Existenz).

**CORS-Preflight.** Standard-Browser-Aufrufe gegen
`/api/ingest/*` werden im Preflight wie der Dashboard-Lese-Pfad
behandelt (siehe Spec §10a für SSE; analog OPTIONS/Allow-Origin
gegen die globale Origin-Allowlist), weil Ingest-Control im
`0.11.0`-Scope kein produktiver Browser-Schreibpfad ist. Die
konkrete Origin-Politik wird mit dem Plan-Closeout beschrieben.

**Fehlerreihenfolge.** Der Handler prüft in dieser Reihenfolge,
analog zum bestehenden `/api/analyze`-Pfad:

1. Content-Type (`application/json`) → `415` bei Verstoß.
2. Body-Größe ≤ `maxIngestRequestBytes` (1 MiB Default) → `413`
   bei Überlauf.
3. JSON-Parsing → `400 invalid_json`.
4. `X-MTrace-Token` vorhanden und resolvierbar → `401 unauthorized`
   sonst.
5. Schema-Validierung (Pflicht-/Optionalfelder, `protocol`-
   Allowlist, `project_id`-Konsistenz) → `400 invalid_request`.
6. Domain-Vorbedingungen (Stream existiert, Endpoint/Target
   existiert, Routing-Regel aktiv) → `404 not_found` /
   `409 conflict` je nach Fall.
7. Use-Case-Aufruf.

**Wire-Skizzen.**

`POST /api/ingest/streams` Request:

```json
{
  "display_name": "Lab SRT",
  "protocol": "srt",
  "endpoint_id": "mediamtx-srt-local",
  "target_id": "mediamtx-local",
  "project_id": "demo"
}
```

`project_id` ist optional; fehlt das Feld, wird es serverseitig
aus dem Token abgeleitet. Stimmt es nicht mit dem Token überein,
liefert der Server `400 invalid_request` mit
`code:"project_id_mismatch"`.

`POST /api/ingest/streams` und
`POST /api/ingest/streams/{id}/rotate-key` Response:

```json
{
  "stream": {
    "id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "project_id": "demo",
    "display_name": "Lab SRT",
    "protocol": "srt",
    "endpoint_id": "mediamtx-srt-local",
    "target_id": "mediamtx-local",
    "routing_rule_id": "route_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "status": "ready",
    "created_at": "2026-05-09T10:00:00Z",
    "updated_at": "2026-05-09T10:00:00Z"
  },
  "stream_key": {
    "value": "mtr_ing_7YQ3pVh4v0hT8x2l9b6nR4c1A5sD0eF2gH3jK8mN9pQ",
    "fingerprint": "mtr_ing_7YQ3...N9pQ",
    "created_at": "2026-05-09T10:00:00Z"
  }
}
```

`stream_key.value` darf ausschließlich in Create-/Rotate-Antworten
erscheinen. List-, Detail-, Event-, Fehler- und Artefakt-Antworten
enthalten höchstens `key_fingerprint`.

`GET /api/ingest/streams` Response:

```json
{
  "streams": [
    {
      "id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
      "project_id": "demo",
      "display_name": "Lab SRT",
      "protocol": "srt",
      "endpoint_id": "mediamtx-srt-local",
      "target_id": "mediamtx-local",
      "routing_rule_id": "route_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
      "status": "ready",
      "key_fingerprint": "mtr_ing_7YQ3...N9pQ",
      "created_at": "2026-05-09T10:00:00Z",
      "updated_at": "2026-05-09T10:00:00Z"
    }
  ]
}
```

`POST /api/ingest/streams/{id}/validate-key` Request/Response:

```json
{ "stream_key": "mtr_ing_7YQ3pVh4v0hT8x2l9b6nR4c1A5sD0eF2gH3jK8mN9pQ" }
```

```json
{
  "valid": true,
  "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "key_fingerprint": "mtr_ing_7YQ3...N9pQ"
}
```

Das Validate-Endpoint ist explizit **kein** produktiver Auth-Pfad.
Es gibt keinen Hinweis auf Existenz fremder Streams (immer
`{ "valid": false }` ohne Detail-Stream-ID, wenn Token und Stream
nicht zum selben Project gehören).

`POST /api/ingest/hooks/stream-{started,ended}` Request:

```json
{
  "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "observed_at": "2026-05-09T10:05:30.123Z",
  "source": "local-smoke",
  "connection_id": "srtconn-1",
  "reason": "smoke_complete"
}
```

`POST /api/ingest/hooks/stream-{started,ended}` Response (`202`):

```json
{
  "accepted": true,
  "event_id": "evt_3f2a91c4...",
  "stream_id": "ing_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "type": "stream_started",
  "observed_at": "2026-05-09T10:05:30.123Z"
}
```

Das Event-`type`-Feld der Antwort wird ausschließlich aus dem
URL-Suffix abgeleitet — ein gegenteiliger `type`-Wert im Body hat
keinen Effekt. Lifecycle-Events tragen **keine** Klartext-Keys;
höchstens den `key_fingerprint`. `source` muss aus der Allowlist
`local-smoke` oder `mediamtx-hook` stammen; unbekannte Werte werden
auf `400 invalid_request` gemappt. `connection_id` und `reason` sind
optional und längenbegrenzt (≤ 256 Zeichen). Produktive ausgehende
Webhook-Zustellung an externe Systeme ist nicht Teil des
`0.11.0`-Vertrags.

`GET /api/ingest/media-server-config` Response:

```json
{
  "target_id": "mediamtx-local",
  "kind": "mediamtx",
  "config_path": "examples/ingest-control/mediamtx.generated.yml",
  "config_yaml": "paths:\n  publish:{stream_path}:\n    source: publisher\n",
  "warnings": []
}
```

`config_yaml` ist das tatsächlich von `apps/api` generierte oder
validierte Artefakt; `config_path` zeigt, wohin der Smoke das
Artefakt schreibt. Das Endpoint produziert kein I/O auf
laufenden externen Media-Servern.

**Fehler-Codes (zusätzlich zu den Schema-/Auth-Codes oben).**

| HTTP | `code` | Anlass |
| ---- | ------ | ------ |
| 400  | `invalid_request` | Pflichtfeld fehlt, `protocol` außerhalb der Allowlist, oder `project_id` widerspricht dem Token. |
| 400  | `project_id_mismatch` | Request-`project_id` weicht vom Token ab. |
| 401  | `unauthorized` | `X-MTrace-Token` fehlt oder ist ungültig. |
| 404  | `stream_not_found` | Stream-ID gehört nicht zum Project oder existiert nicht. |
| 404  | `endpoint_not_found` / `target_not_found` | Referenziertes Endpunkt-/Target-Objekt fehlt. |
| 409  | `stream_name_conflict` | Aktiver Stream mit gleichem `display_name` im Project existiert. |
| 409  | `routing_rule_disabled` | Routing-Regel ist deaktiviert; Lifecycle-Hook lehnt ab. |
| 422  | `key_invalid` | `validate-key`-Response trägt `valid:false` (Wire-Form `{valid:false}` ohne Stream-ID). |
| 503  | `media_server_config_unavailable` | Konfigurations-Artefakt konnte nicht generiert/validiert werden. |

---

### 3.9 Auth / Token Lifecycle (`0.12.0`, RAK-71..RAK-76)

`0.12.0` führt kurzlebige serverseitig signierte Session Tokens für
Browser-Telemetrie, rotierbare Project-Token-Generationen und
Project-gebundene Ingest Policies ein. Der Pfad ist eine Härtung des
bestehenden lokalen/API-nahen Auth-Modells aus §4 (Variante B,
Auth-Modul in `apps/api`) — **kein** vollständiger Identity-/SSO-/
OAuth-Pfad und **kein** mandantenfähiger Control-Plane-Pfad. Bestehende
`X-MTrace-Token`-Project-Token-Flows bleiben im
`0.12.0`-Compatibility-Fenster gültig (RAK-75); siehe Out-of-Scope-
Liste in
[`docs/planning/done/plan-0.12.0.md`](../docs/planning/done/plan-0.12.0.md)
§0 für Scope, Architektur und Threat Model.

**Auth-Matrix (zusätzlich zu §4).**

| Endpoint | Token-Pfad |
|---|---|
| `POST /api/auth/session-tokens` | Pflicht-`X-MTrace-Token` (Project Token); stellt kurzlebiges Session Token aus. |
| `POST /api/playback-events` | Pflicht — entweder `X-MTrace-Token` (Legacy/Project) **oder** `Authorization: Bearer mtr_st_*` **oder** `X-MTrace-Session-Token`. |
| `POST /api/ingest/*` | Pflicht-`X-MTrace-Token` (Project Token) wie in §3.8. **Kein** `0.12.0`-Project-Policy-Enforcement — der Pfad bleibt im `0.11.0`-Token-Validierungs-Modell (RAK-65, lokale/lab-nahe Stream-Verwaltung; `curl`-/Operator-driven, kein Browser-Konsument). Project-Policy für Ingest ist Folge-Scope; vollständige Out-of-Scope-Liste in [`docs/user/ingest-control.md`](../docs/user/ingest-control.md) §5. |
| `POST /api/analyze` mit `correlation_id`/`session_id` und `GET /api/stream-sessions[/{id}]` | Pflicht — zusätzlich zur `X-MTrace-Token`-Variante aus §4 ist `Authorization: Bearer mtr_st_*` oder `X-MTrace-Session-Token` erlaubt, sofern das Session Token den Project- und Session-Scope passend bindet. |

**Header-Priorität für Mehrfach-Auth.** Werden mehrere Auth-Header
gleichzeitig präsentiert, gilt:

1. `Authorization: Bearer mtr_st_*` ist der bevorzugte Session-Token-
   Pfad. Andere `Authorization`-Werte (z. B. fremde OAuth-/Reverse-
   Proxy-Header) sind für m-trace Auth nicht auswertbar und werden
   ignoriert, solange ein gültiger m-trace Header (`X-MTrace-Token`
   oder `X-MTrace-Session-Token`) vorhanden ist. Ohne gültigen
   m-trace Header liefern sie `401 auth_token_missing`.
2. `X-MTrace-Session-Token` ist der alternative Session-Token-Pfad
   für Umgebungen, in denen `Authorization` nicht verwendet werden
   soll.
3. `X-MTrace-Token` ist der Legacy-/Project-Token-Pfad und bleibt
   im `0.12.0`-Compatibility-Fenster gültig.
4. Wenn mehr als ein Auth-Mechanismus präsentiert wird, müssen alle
   präsentierten Tokens zum selben `project_id` passen. Widersprüche
   liefern `401 auth_project_mismatch`. Ein zusätzlich präsentiertes
   ungültiges Token liefert `401 auth_token_invalid`. Es gibt keinen
   stillen Fallback von einem ungültigen höher priorisierten Token
   auf ein gültiges niedriger priorisiertes Token.
5. Sind `Authorization: Bearer mtr_st_*` und `X-MTrace-Session-Token`
   beide gesetzt und enthalten unterschiedliche Session Tokens,
   liefert die API `401 auth_token_invalid`, auch wenn eines der
   beiden Tokens für sich gültig wäre.

**Fehlerpräzedenz für tokenpflichtige Requests.** Diese Reihenfolge
gilt zusätzlich zur Validierungsreihenfolge aus §5; ein Verstoß auf
früherer Stufe verhindert die Auswertung späterer Stufen.

| Priorität | Bedingung | Status / `code` |
| ---: | --- | --- |
| 1 | Pflicht-Auth fehlt vollständig | `401 auth_token_missing` |
| 2 | Präsentierter m-trace Token ist syntaktisch malformed oder Signatur/Hash ungültig | `401 auth_token_invalid` |
| 3 | Präsentierter Token ist widerrufen | `401 auth_token_revoked` |
| 4 | Präsentierter Token ist abgelaufen | `401 auth_token_expired` |
| 5 | Präsentierter Token ist noch nicht gültig (`nbf`) | `401 auth_token_not_yet_valid` |
| 6 | Alle Tokens sind für sich gültig, binden aber unterschiedliche Projects | `401 auth_project_mismatch` |
| 7 | Session Token ist für falsche Audience/Session/Origin gebunden | `403 auth_session_scope_denied` |
| 8 | Project Policy lehnt Origin/Methode/Header/Scope ab | `403 auth_policy_denied` |
| 9 | Endpoint-spezifisches Rate-Limit nach erfolgreicher Auth/Policy-Prüfung überschritten | `429` mit dem im Endpoint-Vertrag definierten Rate-Limit-Code |

`auth_issuance_rate_limited` ist ausschließlich der Rate-Limit-Code
von `POST /api/auth/session-tokens`. Andere Endpoints definieren ihren
eigenen Rate-Limit-Code (für `POST /api/playback-events` bleibt es
`429 Too Many Requests` mit `Retry-After` aus §6).

**CORS-Preflight-Modell.** Browser-Preflights (`OPTIONS`) enthalten in
der Praxis kein Project- oder Session-Token, das der Server
verlässlich validieren könnte. `0.12.0` nutzt deshalb für Preflights
eine globale, konservative und informationsarme Allowlist:

- erlaubte Methoden maximal `POST, OPTIONS`;
- erlaubte Header maximal `Content-Type`, `Authorization`,
  `X-MTrace-Token`, `X-MTrace-Session-Token`, `traceparent`;
- bekannte Origins aus der globalen Union aller konfigurierten
  Project-Origins werden mit dem konkreten Origin gespiegelt — nie
  mit `*`. Erfolgreiche Preflights liefern exakt `204`, leeren Body,
  `Access-Control-Allow-Origin: <Origin>`,
  `Access-Control-Allow-Methods: POST, OPTIONS`,
  `Access-Control-Allow-Headers` mit der erlaubten Header-Liste,
  `Access-Control-Max-Age: 600`, `Vary: Origin` und
  `Cache-Control: no-store`;
- unbekannte Origins erhalten eine minimale Ablehnung: exakt `204`,
  leerer Body, **kein** `Access-Control-Allow-Origin`, **kein**
  `Access-Control-Allow-Methods`, **kein**
  `Access-Control-Allow-Headers`, mit `Vary: Origin` und
  `Cache-Control: no-store`. Es gibt weder Project- noch
  Origin-Enumeration im Body oder in Headern;
- project-spezifische Policy-Entscheidungen (Origin, Methode, Header,
  Audience, Rate-Limit) werden erst beim tatsächlichen `POST`
  ausgewertet, sobald Project-/Session-Token verfügbar sind.

Akzeptiertes Informationsniveau: ein Client darf erkennen, ob sein
eigener Origin in der globalen Union bekannt ist; er darf **nicht**
erkennen, welche Projects existieren oder welche anderen Origins
konfiguriert sind. RAK-74 gilt damit für Request-Enforcement;
Preflight ist ein konservativer Browser-Kompatibilitätsfilter.

**Wire-Skizze: `POST /api/auth/session-tokens`.**

Request:

```json
{
  "project_id": "demo",
  "session_id": "sess_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
  "origin": "http://localhost:5173",
  "audience": "playback-events",
  "ttl_seconds": 900
}
```

Der Request ist mit einem gültigen Project Token authentifiziert
(`X-MTrace-Token`). `project_id` ist optional und dient als
Konsistenzcheck zum Token: fehlt es, wird das Project ausschließlich
aus dem Token abgeleitet und in der Response zurückgegeben; ist es
gesetzt und passt nicht zum Token, liefert die API
`401 auth_project_mismatch`. `audience` muss aus der Project-Policy-
Allowlist stammen — im `0.12.0`-Pflichtpfad ist `playback-events` die
einzige Muss-Audience. Unbekannte oder nicht erlaubte Audiences
liefern `403 auth_session_scope_denied`. `session_id` und `origin`
sind optional und binden den ausgestellten Token zusätzlich; sie
müssen, wenn gesetzt, mit den Werten des späteren konsumierenden
Requests übereinstimmen.

`ttl_seconds` hat eine harte globale Obergrenze von 900 Sekunden.
Project Policies dürfen eine niedrigere Grenze definieren, aber
keine höhere. Fehlt `project_max_ttl_seconds` für ein Project,
gilt exakt der Default `900`. Fehlt `ttl_seconds` im Request,
verwendet der Server `min(project_max_ttl_seconds, 900)`. Werte
`<= 0` oder oberhalb der wirksamen Grenze liefern
`422 auth_token_ttl_too_large` ohne stillen Clamp.

Response (`201 Created`):

```json
{
  "session_token": {
    "value": "mtr_st_eyJraWQiOiJrZXlfMjAyNi0wNSJ9...",
    "token_id": "st_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "project_id": "demo",
    "session_id": "sess_01HZXJ7A5K9V7W1E7BTKJ8V7N9",
    "audience": "playback-events",
    "expires_at": "2026-05-09T10:15:00Z"
  }
}
```

`session_token.value` darf ausschließlich in der Issuance-Antwort
erscheinen. Logs, Fehlerantworten, Metriken, Traces und Fixtures
enthalten höchstens `token_id` oder Fingerprints. `token_id` ist der
öffentliche Wire-Name des `jti`-Claims; beide Werte sind identisch.
Implementierung und Tests verwenden `jti` nur innerhalb der signierten
Claims und `token_id` in Responses, Logs und Doku.

Session-Token-Claims enthalten mindestens `iss`, `sub` (`project_id`),
`aud`, `iat`, `nbf`, `exp`, `jti`, optional `session_id` und `origin`.
Signaturschlüssel werden über einen serverseitigen Key-Ring (ENV/File-
Konfiguration) verwaltet; `kid` im Token-Header erlaubt parallele
Signing-Keys und alte Verify-Keys bleiben über Deployments und
Restarts geladen, bis alle damit signierten Tokens abgelaufen sind.
Unbekannter `kid` liefert `401 auth_token_invalid`.

**Konsumieren von Session Tokens.** `POST /api/playback-events`
akzeptiert in `0.12.0` zusätzlich zu `X-MTrace-Token`:

- `Authorization: Bearer mtr_st_*` — bevorzugter Browser-Pfad;
- `X-MTrace-Session-Token: mtr_st_*` — alternativer Pfad ohne
  `Authorization`-Header.

Beide Pfade folgen der Header-Priorität und Fehlerpräzedenz oben.
Cookies werden für Player-Telemetrie nicht eingeführt; SDK bleibt bei
`credentials:"omit"`.

**Project-Token-Rotation.** Project Tokens (`X-MTrace-Token`) werden
serverseitig als Generationen modelliert: `token_id`, `project_id`,
Hash/Fingerprint, Status, `not_before`, `grace_until?`, `expires_at?`,
`revoked_at?`, `created_at`, `rotated_from?`. Persistenz speichert
nie den Klartext-Token; Klartext erscheint nur bei Erzeugung/Rotation
oder in markierten Operator-/Test-Fixtures. Während einer
Rotations-Grace-Phase bleibt die alte Generation gültig, bis das
persistierte `grace_until` erreicht ist; `revoked_at` beendet Grace
sofort. `grace_until` darf nicht aus volatilem Prozesszustand oder
aus `rotated_from` rekonstruiert werden — es ist Source of Truth und
restart-stabil.

**Project Policies.** Project-gebunden konfigurierbar:

- erlaubte Origins (gegen den `Origin`-Header beim tatsächlichen
  `POST` validiert; leerer `Origin` bleibt für CLI/curl nur dort
  zulässig, wo der jeweilige Endpoint-Vertrag das ausdrücklich
  vorsieht);
- erlaubte Methoden und Header (Subset der globalen Preflight-
  Allowlist);
- erlaubte Session-Token-Audiences (Allowlist; im `0.12.0`-Pflichtpfad
  mindestens `playback-events`);
- maximale Session-Token-TTL (`project_max_ttl_seconds`, ≤ 900);
- Rate-Limit-Buckets pro Project. Globale und Project-Buckets sind
  Muss-Pfad. Origin- und IP-nahe Buckets sind nur dann Teil des
  `0.12.0`-Muss-Scopes, wenn die bestehende Rate-Limit-Infrastruktur
  sie ohne größere Architekturänderung tragen kann; andernfalls sind
  sie als Folge-Scope zu dokumentieren.

**Fehler-Codes (zusätzlich zu §4 und §3.8).**

| HTTP | `code` | Bedeutung |
| ---- | ------ | --------- |
| 401 | `auth_token_missing` | Pflicht-Auth-Header fehlt vollständig. |
| 401 | `auth_token_invalid` | Token syntaktisch malformed, Signatur/Hash ungültig oder unbekannter `kid`. |
| 401 | `auth_token_revoked` | Token-Generation widerrufen (`revoked_at` gesetzt). |
| 401 | `auth_token_expired` | Token-Generation oder Session-Token abgelaufen (`exp`/`expires_at` überschritten). |
| 401 | `auth_token_not_yet_valid` | Token-Generation noch nicht gültig (`nbf`/`not_before` in der Zukunft). |
| 401 | `auth_project_mismatch` | Mehrere präsentierte Tokens binden unterschiedliche Projects, oder Request-`project_id` widerspricht dem Token. |
| 403 | `auth_session_scope_denied` | Session Token passt nicht zur Audience oder zur gebundenen `session_id`/`origin`; oder gewünschte Audience nicht in der Project-Policy-Allowlist. |
| 403 | `auth_policy_denied` | Project Policy lehnt Origin/Methode/Header/Scope ab. |
| 422 | `auth_token_ttl_too_large` | Gewünschte `ttl_seconds` ≤ 0 oder oberhalb der wirksamen Project-TTL-Grenze (max. 900). |
| 429 | `auth_issuance_rate_limited` | Globale oder Project-spezifische Issuance-Quote von `POST /api/auth/session-tokens` überschritten. |

**Validierungsreihenfolge für `POST /api/auth/session-tokens`.**

1. Content-Type (`application/json`) → `415` bei Verstoß.
2. Body-Größe ≤ `maxAuthRequestBytes` (4 KiB Default) → `413` bei
   Überlauf.
3. JSON-Parsing → `400 invalid_json`.
4. Header-Auswahl und Auth-Validierung des Project Tokens
   (`X-MTrace-Token`) gemäß Header-Priorität und Fehlerpräzedenz
   oben.
5. `project_id`-Konsistenzcheck (falls im Body gesetzt).
6. Audience-Allowlist (Project Policy) → `403 auth_session_scope_denied`
   bei Verstoß.
7. `ttl_seconds`-Validierung gegen wirksame Project-TTL-Grenze →
   `422 auth_token_ttl_too_large` bei Verstoß.
8. Issuance-Rate-Limit → `429 auth_issuance_rate_limited` bei
   Überschreitung.
9. Session-Token-Erzeugung und `201`-Response.

**Validierungsreihenfolge für tokenpflichtige Konsum-Endpunkte (`POST
/api/playback-events`, `/api/ingest/*`, Read-Endpunkte mit Session-
Token-Pfad).** Die Stufen aus §5 bleiben gültig; Auth-Stufe 1 wird
durch die obige Header-Priorität und Fehlerpräzedenz konkretisiert.
Project-Policy- und Audience-Prüfung erfolgen zwischen Stufe 4
(Rate-Limit) und Stufe 5 (Schema-Version):

- Project Policy lehnt Origin/Methode/Header ab →
  `403 auth_policy_denied`;
- Session Token bindet falsche Audience/Session/Origin →
  `403 auth_session_scope_denied`.

Logs, Metriken, Traces und Contract-Fixtures enthalten **keine**
Klartext-Tokens (weder Project- noch Session-Tokens), sondern
ausschließlich `token_id` oder Fingerprints. Fremde
`Authorization`-Header ohne `Bearer mtr_st_*` werden nicht in
m-trace Audit-Logs übernommen.

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
| `state` | `"active"` oder `"ended"` | Aktueller Session-Zustand aus `stream_sessions.state`; `ended` gilt für explizites `session_ended` und Sweeper-Ende. |
| `started_at` | string, RFC3339 | Server-/Persistenzzeitpunkt der ersten Beobachtung. |
| `last_seen_at` | string, RFC3339 | Zeitpunkt der letzten Event-Beobachtung oder Session-Aktualisierung. |
| `ended_at` | string, RFC3339 oder `null` | Endezeitpunkt bei `state="ended"`, sonst `null`. |
| `end_source` | `"client"`, `"sweeper"` oder `null` | Ursache des Endzustands: explizites `session_ended` aus Client-/SDK-Pfad, Sweeper-Ende oder `null` bei `state="active"`. |
| `event_count` | int ≥ 0 | Persistierte Event-Anzahl der Session. |
| `correlation_id` | `string`; nicht-leer für ab §3.2 angelegte oder bereits selbst-geheilte Sessions, sonst `""` als Legacy-Fall | Spiegelt `stream_sessions.correlation_id`; identisch mit dem `correlation_id`-Wert auf ab §3.2 persistierten Events derselben Session. Historische Events vor §3.2 werden nicht backfilled. Dient dem Dashboard als primärer Korrelations-Schlüssel — Tempo-unabhängig. |
| `network_signal_absent` | Array, Default `[]` | Session-skopierte Degradationsgrenzen aus `session_boundaries[]`; nur im Session-Block, nie als synthetisches Event. |

`network_signal_absent[]`-Einträge haben die Form:

| Feld | Typ | Beschreibung |
|---|---|---|
| `kind` | `"manifest"` oder `"segment"` | Netzwerksignal-Typ. |
| `adapter` | `"hls.js"`, `"native_hls"` oder `"unknown"` | SDK-/Browserpfad, der die Grenze gemeldet hat. |
| `reason` | Reason-Enum aus `contracts/event-schema.json#network_unavailable_reasons` | Maschinenlesbarer Grund. |

Sortierung ist stabil nach `kind`, dann `adapter`, dann `reason`.
Doppelte Tripel werden im Read-Shape dedupliziert.

---

## 4. Authentifizierung

> **`0.12.0`-Erweiterung:** §3.9 ergänzt diese Basis um kurzlebige
> serverseitig signierte Session Tokens (`Authorization: Bearer
> mtr_st_*` und `X-MTrace-Session-Token`), rotierbare Project-Token-
> Generationen, Project-gebundene Ingest Policies, eine globale
> konservative CORS-Preflight-Allowlist und eine neunstufige
> Auth-Fehlerpräzedenz. Der `X-MTrace-Token`-Pfad aus diesem
> Abschnitt bleibt im `0.12.0`-Compatibility-Fenster gültig
> (RAK-75).

`X-MTrace-Token` ist endpoint-spezifisch:

| Endpoint | Token-Pflicht |
|---|---|
| `POST /api/playback-events` | Pflicht. |
| `GET /api/stream-sessions` und `GET /api/stream-sessions/{id}` | Pflicht ab `plan-0.4.0.md` Tranche 3. |
| `POST /api/analyze` ohne `correlation_id` und ohne `session_id` | Nicht erforderlich; die Analyse bleibt `session_link.status="detached"`. |
| `POST /api/analyze` mit `correlation_id` oder `session_id` | Pflicht; fehlt der Token oder ist er ungültig, liefert die API `401 Unauthorized` und führt keinen Session-Lookup aus. |

- Token-Validierung gegen eine **hardcodierte Map** (Spec §6.4):

  ```json
  {
    "demo": "demo-token"
  }
  ```

  Schlüssel ist `project_id`, Wert ist das erwartete Token.

- Regeln:
  - Fehlt ein nach Matrix pflichtiger `X-MTrace-Token` → `401 Unauthorized`.
  - Ein nach Matrix ausgewerteter Token ist ungültig → `401 Unauthorized`.
  - `project_id` im Event oder Link-Kontext passt nicht zum Token → `401 Unauthorized`.
  - `project_id` im Event oder Link-Kontext ist nicht in der Map → `401 Unauthorized`.

- Dynamische Project-Verwaltung und Endpunkte zum Anlegen oder Rotieren
  von Tokens sind nicht Teil dieses Kontrakts.

---

## 5. Validierungsregeln und Fehlerfälle

Reihenfolge der Validierung für tokenpflichtige Requests (Implementierungen
müssen sich daran halten, damit die Pflichttests deterministisch sind):

1. **Auth-Header**: ein nach §4 pflichtiger `X-MTrace-Token` fehlt →
   `401 Unauthorized`. Diese
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
| Pflichtiger Auth-Header fehlt            | `401` |
| Body > 256 KB (mit Auth-Header)          | `413` |
| Token unbekannt                          | `401` |
| `project_id`/Token-Mismatch              | `401` |
| Rate-Limit überschritten                 | `429` + `Retry-After` |
| `schema_version` ≠ `"1.0"`               | `400` |
| `events` leer oder fehlt                 | `422` |
| `events.length` > 100                    | `422` |
| Event ohne Pflichtfeld                   | `422` |
| Valider Batch                             | `202` |

Folge der Auth-vor-Body-Reihenfolge: ein tokenpflichtiger Request
**ohne** Auth-Header und mit Body > 256 KB liefert `401`, **nicht**
`413` (siehe Pflichttest in §11).

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
- Die vier Pflichtcounter tragen keine fachlichen Labels; erlaubt sind
  nur technische Prometheus-/Target-Metadaten außerhalb des Metric-Vektors.
  Insbesondere sind `project_id`, `session_id`, `viewer_id`, `request_id`,
  `user_agent`, `segment_url`, andere URL-/URL-Teil-Felder, `client_ip`,
  `trace_id`, `span_id`, `correlation_id` sowie Token-/Credential-Felder
  verboten. Eine spätere `project_id`-Ausnahme braucht eine explizite
  bounded Allowlist und einen passenden Cardinality-Smoke. Die normative,
  vollständige Forbidden-Liste über alle `mtrace_*`-Metriken hinweg
  (inklusive Suffix-Regeln `_url`/`_uri`/`_token`/`_secret`) steht in
  [`telemetry-model.md`](./telemetry-model.md) §3.1; `scripts/smoke-observability.sh`
  spiegelt diese Liste und ist release-blockierend.
- `mtrace_dropped_events_total` darf konstant `0` sein, wenn die API
  keinen expliziten Drop-Pfad hat — die Metrik **muss** aber existiert sein.
- `delivery_status: "duplicate_suspected"`-Events (siehe §10.2) zählen
  zu `mtrace_playback_events_total` (sie sind angenommen, nur als
  Duplikat klassifiziert) und **nicht** zu
  `mtrace_invalid_events_total` oder `mtrace_dropped_events_total`.
- Implementierungen dürfen weitere `mtrace_*`-Metriken ergänzen
  (z. B. `mtrace_active_sessions`, der OTel-translated `mtrace_api_batches_received`
  oder `mtrace_analyze_requests_total{outcome,code}`), sofern Cardinality
  kontrolliert ist: entweder label-frei oder ausschließlich bounded
  Aggregat-Labels aus [`telemetry-model.md`](./telemetry-model.md) §3.2.
  `mtrace_api_batches_received` ist ab `0.4.0` Tranche 7 explizit label-frei
  — `batch.size` lebt nur als Span-Attribut, nicht als Counter-Attribut
  (siehe `telemetry-model.md` §2.2 und `plan-0.4.0.md` §8.2).
- **SRT-Health-Aggregate** (`0.6.0`, plan-0.6.0 §4): Tranche 6
  liefert `mtrace_srt_health_samples_total{health_state}`,
  `mtrace_srt_health_collector_runs_total{source_status}` und
  `mtrace_srt_health_collector_errors_total{source_error_code}`.
  Erlaubte Labelwerte sind die Enums aus
  [`telemetry-model.md`](./telemetry-model.md) §7.4/§7.5; die
  Allowlist ist in §3.2 dort entsprechend ergänzt. Per-Verbindung-
  Felder (`stream_id`, `connection_id`, MediaMTX-`id`/`path`/
  `remoteAddr`/`state`) bleiben in §3.1 verboten und gehen
  ausschließlich in SQLite/OTel-Spans (siehe §10.6 SRT-Health-
  Persistenz und §7a SRT-Health-Read-Vertrag unten).

---

## 7a. SRT-Health-Read-Vertrag (`0.6.0`)

> Bezug: Lastenheft §13.8 RAK-43;
> [`plan-0.6.0.md`](../docs/planning/done/plan-0.6.0.md) §5
> (Tranche 4); [`telemetry-model.md`](./telemetry-model.md) §7.

`0.6.0` Tranche 4 liefert die Read-Endpoints. Vorgesehene Form
(Tranche 4 finalisiert):

### 7a.1 Endpoints

| Endpoint | Bedeutung |
|---|---|
| `GET /api/srt/health` | Aktuelle Health-Snapshots aller bekannten SRT-Streams (eine Zeile pro `stream_id`). |
| `GET /api/srt/health/{stream_id}` | Detail plus Verlauf der letzten N Samples. |

CORS- und Auth-Verhalten folgt den bestehenden Dashboard-Lese-Pfaden
aus §4 Authentifizierung — kein Project-Token im Read-Pfad,
Origin-Validierung gegen globale Allowlist (analog
`/api/stream-sessions`).

### 7a.2 Response-Struktur

Trennt Rohwerte, abgeleitete Werte und Bewertung explizit:

```json
{
  "stream_id": "srt-test",
  "connection_id": "00000000-0000-0000-0000-000000000001",
  "health_state": "healthy",
  "source_status": "ok",
  "source_error_code": "none",
  "connection_state": "connected",
  "metrics": {
    "rtt_ms": 0.231,
    "packet_loss_total": 0,
    "packet_loss_rate": 0,
    "retransmissions_total": 0,
    "available_bandwidth_bps": 3623031946,
    "throughput_bps": 1153142,
    "required_bandwidth_bps": 1500000
  },
  "derived": {
    "loss_per_window": 0,
    "retrans_per_window": 0,
    "bandwidth_headroom_factor": 2415.354
  },
  "freshness": {
    "source_observed_at": null,
    "source_sequence": "37208036",
    "collected_at": "2026-05-05T08:48:01Z",
    "ingested_at": "2026-05-05T08:48:01.250Z",
    "sample_age_ms": 250,
    "stale_after_ms": 15000
  }
}
```

`sample_age_ms` ist die Zeit seit dem letzten **Source-Wechsel**
(monoton steigender `source_sequence`-Wert), **nicht** seit
`ingested_at`. Wenn die Quelle keinen `source_observed_at` liefert
und `source_sequence` über N Polls identisch bleibt, steigt
`sample_age_ms` an, bis `stale_after_ms` überschritten wird —
dann setzt der Server `source_status: stale`.

### 7a.3 Pagination und Limits

- Default-Limit pro Stream-Verlauf: 100 Samples; hartes Maximum 1000
  (Query-Parameter `samples_limit`).
- Kanonische Sortierung: `(ingested_at desc, id desc)` (analog §10.4).
- Cursor-Pagination via Query-Parameter `samples_cursor` (opaker
  base64-url-kodierter Token, kapselt die Storage-Position
  `(ingested_at, id)` **plus den Collection-Scope `(project_id,
  stream_id)`** analog dem v3-Event-Cursor aus §10.3). Token-Inhalt
  ist servergetragen und vom Client als opak zu behandeln.
- Antwort liefert `next_cursor` (String) im Response-Body, wenn eine
  Folgeseite existiert. Auf der letzten Seite fehlt das Feld
  (`omit-empty`).
- Andere Query-Parameter-Namen oder Aliasse (z. B. `cursor`) sind
  nicht Teil des Kontrakts; ein gesetzter Alias wird vom Server
  ignoriert.

### 7a.4 Fehlerverhalten

- `404` bei unbekannter `stream_id`.
- `200` mit `health_state: unknown`, `source_status: unavailable`,
  `metrics: {}` falls die Quelle aktuell nicht erreichbar ist (Stream
  war früher bekannt, aktuell kein Sample vorhanden). Operator-
  sichtbar als „Health unbekannt" plus stale-Hinweis.
- Cursor-Reject analog §10.3-Tabelle:
  - `400 cursor_invalid_legacy` mit Body
    `{"error":"cursor_invalid_legacy","reason":"<kurze Erklärung>"}`
    für Cursor mit `v`-Feld 1/2 oder fehlendem `v`-Feld
    (Pre-§4.3-Format ohne Collection-Scope).
  - `400 cursor_invalid_malformed` mit Body
    `{"error":"cursor_invalid_malformed","reason":"<kurze
    Erklärung>"}` für Base64-/JSON-Decode-Fehler, unbekannten `v`-
    Wert, fehlende oder ungültige Pflichtfelder, oder Scope-
    Mismatch (Cursor aus Project A im Request-Kontext Project B,
    bzw. Cursor aus Stream X im Request-Kontext Stream Y).
  - `410 cursor_expired` mit Body
    `{"error":"cursor_expired","reason":"<kurze Erklärung>"}` für
    Cursor mit valider Position, die durch `make wipe` o. ä. nicht
    mehr existiert (siehe §10.3-Tabelle).
- `Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers`
  in jeder Antwort (analog §3.5).

### 7a.5 Pflichttest-Anker

- Snapshot-Test pinnt das oben gezeigte Response-Schema gegen
  ein OpenAPI-/Contract-Fixture (Tranche 4 finalisiert den Pfad).
- Health-State-Schwellen-Tests (Tranche 4) decken Normalfall,
  `degraded`, `critical`, `unknown`/`stale` ab.

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
  v3-List-Cursor enthalten den Project-Scope (`project_id` oder einen
  daraus abgeleiteten Scope-Hash) zusätzlich zur Storage-Position.
  v3-Event-Cursor enthalten den Collection-Scope (`project_id` +
  `session_id` oder einen daraus abgeleiteten Scope-Hash) zusätzlich
  zur Storage-Position.
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
| `accepted` | Token decodiert; `v == 2` vor Tranche 3 oder `v == 3` ab Tranche 3; alle Pflichtfelder vorhanden und valide; bei v3 passt der Project-Scope zum Request-Kontext und bei Event-Cursorn zusätzlich der Session-Scope zum Pfad `{id}`. | `200 OK`. | regulärer Listen-Response inkl. `next_cursor`. | weiter paginieren mit `next_cursor`. |
| `cursor_invalid_legacy` | Token decodiert; `v`-Feld fehlt oder enthält `1`; oder `pid`-Feld vorhanden; nach Aktivierung projekt-skopierter Read-Pfade auch `v == 2` für Session-Cursor ohne Project-Scope. | `400 Bad Request`. | `{"error":"cursor_invalid_legacy","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |
| `cursor_invalid_malformed` | Base64- oder JSON-Decode schlägt fehl; oder `v`-Feld enthält unbekannten Wert; oder Pflichtfeld fehlt/Format ungültig; oder unbekannte Zusatzfelder vorhanden; oder v3-Project-Scope passt nicht zum Request-Kontext; oder v3-Event-Cursor-Scope passt nicht zur Session im Pfad. | `400 Bad Request`. | `{"error":"cursor_invalid_malformed","reason":"<kurze Erklärung>"}`. | Cursor verwerfen, Snapshot ohne `cursor` neu laden. |
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

### 10.6 SRT-Health-Persistenz (`0.6.0`)

> Bezug: Lastenheft §13.8 RAK-42/RAK-46;
> [`plan-0.6.0.md`](../docs/planning/done/plan-0.6.0.md) §4
> (Tranche 3 Sub-3.3); [`telemetry-model.md`](./telemetry-model.md)
> §7.

SRT-Health-Samples sind durable in SQLite persistiert (ADR-0002).
Tabelle (Vorschlag, Tranche 3 Sub-3.3 finalisiert über
`apps/api/internal/storage/schema.yaml`):

| Spalte | Typ | Bemerkung |
|---|---|---|
| `id` | INTEGER PRIMARY KEY AUTOINCREMENT | Surrogat-PK. |
| `project_id` | TEXT NOT NULL | Aus `application/Project`-Resolver. |
| `stream_id` | TEXT NOT NULL | Lab-Stream-Name; nicht Prometheus-Label. |
| `connection_id` | TEXT NOT NULL | Quellseitige Verbindungs-ID. |
| `source_observed_at` | TEXT NULL | RFC3339-Timestamp; bei MediaMTX-Quelle in `0.6.0` `NULL`. |
| `source_sequence` | TEXT NULL | Source-Sequence-Surrogat; Pflicht bei `NULL`-`source_observed_at`. |
| `collected_at` | TEXT NOT NULL | Polling-Zeitpunkt des Collectors. |
| `ingested_at` | TEXT NOT NULL | SQLite-Persistenz-Zeitpunkt. |
| `rtt_ms` | REAL NOT NULL | Snapshot. |
| `packet_loss_total` | INTEGER NOT NULL | Counter. |
| `packet_loss_rate` | REAL NULL | Rate, optional. |
| `retransmissions_total` | INTEGER NOT NULL | Counter. |
| `available_bandwidth_bps` | INTEGER NOT NULL | Berechnet aus `mbpsLinkCapacity × 1_000_000`. |
| `throughput_bps` | INTEGER NULL | Optional. |
| `required_bandwidth_bps` | INTEGER NULL | Aus Lab-Konfig oder Stream-Konfiguration. |
| `sample_window_ms` | INTEGER NULL | Optional. |
| `source_status` | TEXT NOT NULL | Enum aus `telemetry-model.md` §7.5. |
| `source_error_code` | TEXT NOT NULL | Enum aus §7.5. |
| `connection_state` | TEXT NOT NULL | Enum aus §7.1. |
| `health_state` | TEXT NOT NULL | Enum aus §7.4. |

**Dedupe-/Upsert-Regel**: ein Sample ist eindeutig über
`(project_id, stream_id, connection_id, COALESCE(source_observed_at, source_sequence))`.
`collected_at` allein ist **kein** stabiler Dedupe-Schlüssel
(Wiederholtes Lesen identischer Quellen-Daten würde sonst neue Zeilen
erzeugen). Index auf `(project_id, stream_id, connection_id, ingested_at)`
für Latest-First-Reads.

**Retention**: in `0.6.0` analog zur SQLite-Demo-Daten-Politik
(unbegrenzt; `make wipe`-äquivalent reicht). Bounded Snapshot-
Historie mit dokumentiertem Reset-/Prune-Pfad ist Folge-Scope
(plan-0.6.0 §4 DoD).

**Schema-Migration** ist idempotent und mit Restart-Tests
abgesichert (analog `plan-0.4.0.md` §2.5 / `apps/api/internal/storage/`-
Migrationspfad). Tranche 7 verifiziert die Migration über
`make schema-validate`.

---

## 10a. SSE-Live-Stream (`GET /api/stream-sessions/stream`)

Ab `0.4.0` (`plan-0.4.0.md` §5 H4) bietet die API einen Server-Sent-
Events-Stream, der pro neu persistiertem Playback-Event einen
Mindestframe pushed. Vertragstext und Test-Anker:

**Auth.** Tokenpflichtig: fehlender oder ungültiger
`X-MTrace-Token` → `401`. Gültiger Token scoped Stream und
Backfill auf das aufgelöste Project.

**Content-Type.** `text/event-stream; charset=utf-8`. Cache-
Control: `no-store`. Connection: `keep-alive`.

**Frame-Format.** EventSource-kompatibel:

```
id: <ingest_sequence>
event: event_appended
data: {"project_id":"demo","session_id":"01J7K9...","ingest_sequence":42,"event_name":"manifest_loaded"}

```

(Trailing-Newline + Leerzeile nach jedem Frame, `\n\n`-Terminator.)
Pro `playback_events.ingest_sequence` wird genau ein Frame gepushed.
Der `event`-Typ ist konstant `event_appended`. Felder im JSON-`data`
sind exakt: `project_id`, `session_id`, `ingest_sequence`,
`event_name`. Konsumenten laden den vollen Read-Shape (Event-Felder
wie `server_received_at`, `correlation_id`, `meta` usw.) und den
aktuellen Session-Header über `GET /api/stream-sessions/{id}` nach;
der SSE-Frame ist Trigger, nicht vollständige Timeline-Zeile.

**Backfill via `Last-Event-ID`.** Reconnect-Clients dürfen den
Header `Last-Event-ID: <ingest_sequence>` setzen; der Server liefert
dann zuerst alle Events des Projects mit
`ingest_sequence > Last-Event-ID` als Backfill-Frames in
`ingest_sequence`-Aufsteigender Reihenfolge, danach erst Live-Frames.
Der Header ist optional; ohne Header startet der Stream rein live ab
dem nächsten neuen Event.

**Heartbeat.** Alle 15 Sekunden pushed der Server ein SSE-Comment-
Frame `: heartbeat\n\n`, damit Proxies den Stream nicht als
idle-Timeout schließen. Comments sind im EventSource-Vertrag
definiert (Zeilen mit führendem `:`) und werden von Konsumenten
ignoriert.

**Reconnect-Semantik.** Server liefert keinen `retry`-Hint;
Konsumenten nutzen ihre Default-Backoff-Strategie. Bei
Server-Restart oder Stream-Abbruch hält der Client den letzten
gesehenen `id` und reconnectet mit `Last-Event-ID`-Header — der
Backfill schließt die Lücke.

**Heart-Beat-only-Polling-Fallback.** Konsumenten, die kein SSE
sprechen (z. B. Proxy blockiert `text/event-stream`), bleiben auf
`GET /api/stream-sessions/{id}` mit Polling-Intervall ≥ 5s. Das
ist ausdrückliche, dokumentierte Abwärts-Kompat; Plan-DoD §5
verlangt den Fallback explizit.

**CORS-Preflight.** `OPTIONS /api/stream-sessions/stream` →
`Access-Control-Allow-Methods: GET, OPTIONS` und
`Access-Control-Allow-Headers: Content-Type, X-MTrace-Project,
X-MTrace-Token, Last-Event-ID`. Reihenfolge der `Allow-Headers` ist
nicht semantisch (Fetch-Spec behandelt den Wert als ungeordnete
Liste), aber pinnen die Pflichttests den exakten String, damit Spec-/
Code-Drift sofort sichtbar wird. Origin-Echo nur bei zugelassenem
Origin (CORS Variante B); sonst `403`.

**Cross-Project-Scope.** Frames werden nur für Events des
authentifizierten Projects gepushed. `Last-Event-ID`-Backfill
filtert ebenfalls nach `project_id`; ein Cross-Project-`ingest_
sequence`-Wert im Header liefert keine fremden Events, sondern nur
neuere des eigenen Projects.

**Limits.** `Last-Event-ID`-Backfill max. 1000 Events pro
Reconnect; bei größerer Lücke wird der Stream mit einem
`event: backfill_truncated\ndata: {"oldest_ingest_sequence":N}\n\n`-
Frame geöffnet, anschließend live ab `N`. Konsumenten müssen den
Detail-Snapshot dann erneut laden.

**Pflicht-Tests.** Backend-Tests pinnen: SSE-Header (Content-Type,
Cache-Control), EventSource-Format des Mindestframes, Heartbeats
nach Idle-Timeout, Client-Disconnect (Server stoppt Loop und gibt
Ressourcen frei), `Last-Event-ID`-Backfill mit
Project-Skopierung, fehlender/ungültiger Token → `401`,
CORS-Preflight, `backfill_truncated` ab > 1000 Events.

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
