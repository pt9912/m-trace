# Telemetry-Model — m-trace

> **Status**: Verbindlich für `0.1.x`. Wire-Format und Backpressure-Limits sind harte Voraussetzung für `0.1.1` (Player-SDK); OTel-Modell und Cardinality-Regeln sind harte Voraussetzung für `0.1.2` (Observability-Stack).  
> **Bezug**: [Lastenheft `1.1.6`](./lastenheft.md) §7.10 (Cardinality), §7.11 (Telemetry Ingest, Event-Schema, SDK-Budget); [Roadmap](../docs/planning/in-progress/roadmap.md) §2 Schritt 6; [Plan `0.1.0`](../docs/planning/done/plan-0.1.0.md) §3.5; [API-Kontrakt](./backend-api-contract.md); [Architektur](./architecture.md) §5.

## 0. Zweck

Beschreibt das **Datenmodell** der Telemetrie — Wire-Format, OTel-Schema, Cardinality-Regeln, Time-Stempel-Konventionen, Backpressure-Politik. Implementierungs-/Setup-Aspekte (strukturierte Logs, Health-Endpoint, Prometheus- und Grafana-Konfiguration) gehören in [`plan-0.1.2.md`](../docs/planning/done/plan-0.1.2.md), nicht hierher.

Drei Wirkungsebenen pro Telemetrie-Datum:

1. **Wire** — was zwischen Browser-SDK und API über das Netz fliegt (§1).
2. **Ingest** — wie das Backend ankommende Daten validiert, normalisiert und persistiert (siehe `apps/api/hexagon/application/RegisterPlaybackEventBatch` und [API-Kontrakt §5](./backend-api-contract.md)).
3. **Beobachtung** — was als OTel-Span/-Counter und als Prometheus-Metrik nach außen tritt (§2, §3).

---

## 1. Wire-Format Player-Events

> Bezug: F-106..F-115; API-Kontrakt §3; Lastenheft §7.11.1–§7.11.3.

Maschinenlesbare Source of Truth für die Wire-Schema-Version ist
[`contracts/event-schema.json`](../contracts/event-schema.json). Die
SDK↔Schema-Kompatibilität steht in
[`contracts/sdk-compat.json`](../contracts/sdk-compat.json). Änderungen an
Schema-Version, SDK-Version oder API-`SupportedSchemaVersion` müssen diese
Contract-Artefakte im selben Commit aktualisieren.

### 1.1 Batch-Wrapper

Ein einzelner `POST /api/playback-events`-Request transportiert genau einen **Batch** mit 1..100 Events:

```json
{
  "schema_version": "1.0",
  "events": [ /* 1..100 EventInput-Objekte */ ]
}
```

Pflicht-Header (siehe API-Kontrakt §1):

- `Content-Type: application/json`
- `X-MTrace-Token: <project-token>` — siehe §1.5.
- optional `X-MTrace-Project: <project-id>` — reserviert für CORS-Allowlist und spätere strengere Project-Bindung; im `0.1.x` nicht ausgewertet.

### 1.2 Event-Pflichtfelder

| Feld | Typ | Bedeutung | Bezug |
|---|---|---|---|
| `event_name` | string, nicht-leer | bezeichnet das Event (siehe §1.3). | API-Kontrakt §3.2 |
| `project_id` | string, nicht-leer | identifiziert das Projekt; muss zum gelieferten Token passen, sonst `401` (Token-Bindung). | F-106; API-Kontrakt §5 Step 9 |
| `session_id` | string, nicht-leer | identifiziert die Player-Session; pseudonym (NF-40). | API-Kontrakt §3.2 |
| `client_timestamp` | string, RFC3339 mit ms | Erzeugungszeitpunkt am Client. | F-124; siehe §5 |
| `sdk.name` | string, nicht-leer | identifiziert das SDK (z. B. `@npm9912/player-sdk` ab `0.2.0`). | API-Kontrakt §3.2 |
| `sdk.version` | string, SemVer | identifiziert die SDK-Version. | API-Kontrakt §3.2 |

### 1.3 Erfasste Event-Typen im MVP

Für `0.1.x` werden mindestens die folgenden `event_name`-Werte unterstützt; weitere können additiv ergänzt werden, ohne Schema-Bump (F-114):

| event_name | Trigger im SDK | Pflicht ab |
|---|---|---|
| `manifest_loaded` | hls.js `MANIFEST_LOADED` | `0.1.1` |
| `segment_loaded` | hls.js `FRAG_LOADED` | `0.1.1` |
| `playback_started` | erstes `playing` nach `loadedmetadata` | `0.1.1` |
| `bitrate_switch` | hls.js `LEVEL_SWITCHED` | `0.1.1` |
| `rebuffer_started` | `waiting` während Wiedergabe | `0.1.1` |
| `rebuffer_ended` | erstes `playing` nach `waiting` | `0.1.1` |
| `playback_error` | hls.js Error oder `error`-Event auf `<video>` | `0.1.1` |
| `startup_time_measured` | erster Startup-Abschluss, `meta.duration_ms` enthält Startup-Dauer | `0.1.1` |
| `metrics_sampled` | SDK-seitiger Metrik-Snapshot, optional für spätere Erweiterungen | `0.1.1` |
| `session_ended` | Tab-Close oder explizites SDK-`stop()` | `0.1.1` |

### 1.4 Optionale Felder

| Feld | Typ | Bedeutung | Bezug |
|---|---|---|---|
| `sequence_number` | int64 ≥ 0 | monoton aufsteigend pro Session; primärer Ordnungsschlüssel innerhalb einer Session. | F-127, F-128 |
| `client_time_origin` | string, RFC3339 | Setup-Zeitpunkt des SDK; erlaubt skew-tolerantere Latenz-Auswertung. | F-126 |
| `meta` | object | beliebige event-spezifische Felder, z. B. `bitrate`, `duration_ms`, `error_code`. Schema-Erweiterung über §6. | F-114 |

Ab `plan-0.4.0.md` Tranche 3 nutzen Manifest-/Segment-nahe
Netzwerkereignisse additive, flache `meta`-Keys nach dem Muster
`network.*`. Der Punkt ist Teil des Key-Namens, kein verschachteltes
Objekt; das bleibt kompatibel mit der aktuellen SDK-Typisierung
`EventMeta = Record<string, string | number | boolean | null>`.
Der Degradationsmarker ist normativ:

| Feld | Typ | Bedeutung |
|---|---|---|
| `meta["network.kind"]` | string aus `{"manifest", "segment"}` | Netzwerkbezug des Events. |
| `meta["network.detail_status"]` | string aus `{"available", "network_detail_unavailable"}` | `available`, wenn Timing-/URL-Details nach Redaction nutzbar sind; `network_detail_unavailable`, wenn Browser, CORS, Resource Timing, Service Worker, Redirects oder native HLS die Detaildaten blockieren. |
| `meta["network.unavailable_reason"]` | string, optional | Maschinenlesbarer Grund aus dem normativen Reason-Enum: `native_hls_unavailable`, `hlsjs_signal_unavailable`, `browser_api_unavailable`, `resource_timing_unavailable`, `cors_timing_blocked`, `service_worker_opaque`; zusätzlich `^[a-z0-9_]{1,64}$`. **Diese Tabelle ist der einzige normative Anker des Reason-Enums** — `session_boundaries[].reason` (siehe unten), `spec/backend-api-contract.md` §3.4 und `contracts/event-schema.json` (`network_unavailable_reasons`, `network_unavailable_reason_pattern`) verweisen ausschließlich auf diese Werte. |
| `meta["network.redacted_url"]` | string, optional | Bereits redigierter URL-Repräsentant gemäß Redaction-Matrix; rohe URLs mit Query, Fragment, `userinfo` oder tokenartigen Pfadsegmenten sind unzulässig. |

`meta["network.unavailable_reason"]` ist nur zulässig, wenn
`meta["network.detail_status"]="network_detail_unavailable"` ist. Bei
`available` wird ein Reason-Wert als semantischer Widerspruch behandelt
und mit `422` abgelehnt.

Reservierte `network.*`- und `timing.*`-Keys werden inbound vor
Persistenz typvalidiert. Objekte und Arrays sind für diese Keys
unzulässig; `network.*`-Werte sind Strings mit den oben dokumentierten
Domänen/Redaction-Regeln, `timing.*`-Werte sind Zahlen oder explizit
dokumentierte RFC3339-Strings. Verstöße liefern `422` und werden nicht
persistiert.

`network_detail_unavailable` ist kein Fehlerstatus und darf allein
keinen 4xx auslösen. Das Event bleibt in der Session-Timeline sichtbar,
behält seine serverseitig vergebene `correlation_id` und kann als
Timeline-only-Ereignis ohne OTel-Span umgesetzt werden.

Wenn der Browser-/Native-HLS-Pfad gar kein Manifest-/Segment-Signal
liefert, erzeugt das SDK kein synthetisches Netzwerkereignis. SDK oder
Adapter muss stattdessen bei Session-Start bzw. Capability-Erkennung
einen optionalen Batch-Wrapper-Block `session_boundaries[]` an
`POST /api/playback-events` senden. Für Tranche 3 ist darin nur
`kind="network_signal_absent"` definiert; der Block enthält außerdem
`project_id`, `session_id`, `network_kind` (`manifest` oder `segment`),
`adapter` (`hls.js`, `native_hls` oder `unknown`), `reason` und
`client_timestamp`. Dieser Block ist kein Event, besitzt kein
`event_name`, zählt nicht in `accepted` und ändert die
Batch-`schema_version` nicht. Pro Batch sind maximal 20 Boundaries
zulässig, sie zählen ins Body-Size-Budget, und jede Boundary muss eine
`(project_id, session_id)`-Partition referenzieren, die mindestens ein
Event im selben Batch trägt. `reason` verwendet denselben normativen Reason-Enum wie
`meta["network.unavailable_reason"]` (Pflicht-Liste und Pattern siehe
Tabelle weiter oben in §1.4); andere Werte oder Verstöße gegen
`^[a-z0-9_]{1,64}$` führen zu `422`. Die Session-API markiert diese Grenze
außerhalb des Event-Streams im Session-Block als `network_signal_absent`:
Liste von Objekten mit `kind`, `adapter` und maschinenlesbarem `reason`.
Persistenzvehikel ist eine durable Session-Metadaten-Spalte oder ein
äquivalenter session-skopierter Capability-/Boundary-Record; der Wert
darf nicht nur aus flüchtigem Prozesszustand abgeleitet werden und muss
über API-Restart stabil bleiben. Dashboard-Sichtbarkeit wird im
`plan-0.4.0.md` Tranche-4-Scope umgesetzt.

Tranche 3 führt keine neuen `event_name`-Werte ein. Manifest- und
Segment-Netzwerkdetails werden über die bestehenden `manifest_loaded`
und `segment_loaded`-Events plus additive flache `network.*`-Meta-Keys
modelliert.

URL-Redaction für `network.*`-URL-Repräsentanten in `meta` und für alle
URL-verdächtigen generischen Meta-Keys (`url`, `uri`, `manifest_url`,
`segment_url`, `media_url`, `network.url`, `network.redacted_url`,
`request.url`, `response.url`, case-insensitiv) folgt einer festen
Matrix: Scheme, Host und nicht-sensitive Pfadsegmente dürfen erhalten
bleiben; Query und Fragment werden entfernt; `userinfo` wird entfernt;
signierte/credential-artige Query-Parameter (`token`, `signature`,
`sig`, `expires`, `key`, `policy`, case-insensitiv) werden nicht
gespeichert. Ein Pfadsegment ist tokenartig, wenn es mindestens 24
Zeichen lang ist und mindestens 80 % seiner Zeichen aus
`[A-Za-z0-9_-]` bestehen, wenn es ein Hex-String mit gerader Länge
mindestens 32 ist, oder wenn es bekannte JWT-/SAS-/Signed-URL-Muster
trägt. Tokenartige
Pfadsegmente werden ausschließlich als `:redacted` persistiert; es wird
kein stabiler Hash oder Gleichheitsmarker gespeichert. Unbekannte
Meta-Keys mit String-Werten, die als absolute URL parsebar sind oder
`://` enthalten, werden vor Persistenz redigiert oder verworfen.

#### 1.4.1 Bekannte Grenzen der Korrelation

`correlation_id` ist der Pflichtkontext für jede Player-Session-
Timeline und ist über alle Events derselben Session konstant. `trace_id`
ist eine optionale Debug-Vertiefung für Tempo-Cross-Trace-Suche, ist
batch-bezogen und darf eine Timeline-Zuordnung nicht alleine tragen
(siehe §2.5).

Die folgenden Grenzen sind ab `0.4.0` bekannt und akzeptiert:

- **Browser-APIs / Resource Timing**: nicht alle Browser exposen
  Resource-Timing-Daten für Cross-Origin-Fetches; gemessene
  Latenzen können fehlen. Das Event bleibt sichtbar, `network.detail_status`
  signalisiert `network_detail_unavailable` mit Reason
  `browser_api_unavailable` oder `resource_timing_unavailable`.
- **CORS**: bei fehlendem `Access-Control-Allow-Origin` oder
  `Timing-Allow-Origin` blockt der Browser Resource-Timing-Felder.
  Reason: `cors_timing_blocked`.
- **hls.js-spezifische Signal-Lücken**: einzelne hls.js-Versionen
  oder -Konfigurationen exposen Manifest-/Fragment-Signale
  unvollständig (z. B. fehlt `FRAG_LOADED.frag.url` in einer
  Beta-Version). Das SDK degradiert mit Reason
  `hlsjs_signal_unavailable` ohne den hls.js-Stack zu unterbrechen.
- **Service Worker**: ein abfangender Service Worker kann
  Manifest-/Segment-Loads ohne hls.js-Sichtbarkeit beantworten.
  Reason: `service_worker_opaque`.
- **CDN-Redirects / signierte URLs**: 3xx-Redirects auf
  signierte CDN-URLs können hls.js-`FRAG_LOADED` mit veränderter
  URL feuern. Die ursprüngliche Anfrage-URL wird **nicht**
  persistiert; nur der redigierte URL-Repräsentant. Tokenartige
  Pfadsegmente werden vor Persistenz durch `:redacted` ersetzt.
- **Native HLS** (Safari iOS/macOS, ohne hls.js): liefert keine
  Manifest-/Fragment-Events. Das SDK markiert das per
  `session_boundaries[]`-Eintrag mit `adapter="native_hls"`,
  Reason `native_hls_unavailable`. Der Read-Pfad zeigt das im
  Session-Block als `network_signal_absent[]`.
- **Sampling**: Konsumenten können das Player-SDK so
  konfigurieren, dass Events vor dem Send ausgesampelt werden
  (`sampleRate < 1`). Das ist ausdrückliche, dokumentierte
  Degradation und kein Vertragsbruch — die Session-Korrelation
  bleibt für jede tatsächlich beim Server eingehende Session
  konsistent, weil `correlation_id` serverseitig vergeben wird.
  Vollständig ausgesampelte Sessions (`sampleRate=0` oder zufällig
  alle Events gefiltert) erreichen den Server gar nicht erst und
  existieren in den Read-Antworten nicht — Operator-Konsequenz,
  kein Datenintegritätsproblem.

`POST /api/playback-events` akzeptiert `network.unavailable_reason`-Werte
nur aus dem Reason-Enum (`network_unavailable_reasons`) plus
`^[a-z0-9_]{1,64}$`; der gleiche Enum gilt für
`session_boundaries[].reason`. Andere Werte werden mit `422`
abgewiesen — die SDK-Adapter halten sich an den Enum, damit das
Backend defensiv enforcen kann.

### 1.5 SDK-Identifier und Tokens

- **Project Token (`X-MTrace-Token`)**: öffentlicher Token, der dem Browser ausgeliefert wird. Token bindet auf eine `project_id`; Mismatch bei Step 9 → `401` (siehe §5.3 unten und API-Kontrakt §5).
- Token sind laut F-109 **keine** hochkritischen Secrets — sie schützen vor zufälligem Misuse, nicht vor gezielten Angriffen. Rotation und tenant-spezifische Policies sind Kann-Anforderungen (F-111..F-113), nicht im `0.1.x`-Scope.
- **NF-37 CSP-Beispiele für SDK-`connect-src`**: für Drittanbieter, die das Player-SDK in eigene Seiten einbinden, wird folgendes Muster empfohlen:

  ```text
  Content-Security-Policy: default-src 'self'; connect-src 'self' https://collector.example.com
  ```

  Der API-Endpoint muss zusätzlich in der projektspezifischen Allowed-Origins-Liste des Backends stehen (F-108).

### 1.6 Beispiel-Payload (Happy-Path)

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
      },
      "meta": {
        "buffered_seconds": 1.8,
        "current_time_seconds": 23.4
      }
    }
  ]
}
```

---

## 2. OTel-Modell

> Bezug: F-91, F-92; API-Kontrakt §8; Architektur §5.3.

### 2.1 Spans

`apps/api/adapters/driving/http` erzeugt einen Span pro Request am HTTP-Boundary. Das ist der einzige Span-Pfad im `0.1.x` (Use Case spricht OTel ausschließlich über den `Telemetry`-Driven-Port — siehe §2.2).

| Span-Name | Wann | Pflicht-Attribute | Implementiert in |
|---|---|---|---|
| `http.handler POST /api/playback-events` | pro Request auf den Player-SDK-Pfad | `http.method=POST`, `http.route=/api/playback-events`, `http.status_code=<code>`, `batch.size=<int>` (sobald JSON geparst), `batch.outcome=<accepted\|invalid\|unauthorized\|too_large\|rate_limited\|error\|other>` | 0.1.0-pre, plan-0.1.0 §4.3 |
| `http.handler GET /api/stream-sessions` | pro Request auf den Listen-Endpoint | `http.method=GET`, `http.route=/api/stream-sessions`, `http.status_code` | 0.1.0, plan-0.1.0 §5.1 |
| `http.handler GET /api/stream-sessions/{id}` | pro Request auf den Detail-Endpoint | `http.method=GET`, `http.route=/api/stream-sessions/{id}`, `http.status_code` | 0.1.0, plan-0.1.0 §5.1 |
| `http.handler GET /api/health` | pro Health-Check | `http.method=GET`, `http.route=/api/health`, `http.status_code` | 0.1.0, plan-0.1.0 §5.1 |

`GET /api/metrics` erzeugt **keinen** Span — der Prometheus-Endpoint wird vom Scraper periodisch und in hoher Frequenz gepollt; ein Span pro Scrape würde Trace-Storage ohne Erkenntnisgewinn aufblähen.

Span-Attribute folgen [OTel HTTP Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/http/) wo anwendbar; m-trace-spezifische Erweiterungen nutzen den Namespace `mtrace.*` oder `batch.*`.

Vor `0.4.0` durften `session_id`, `user_agent` und `segment_url` als **Span-Attribute** verwendet werden (Cardinality-Regel gilt nur für Prometheus-Labels, nicht für Trace-Attribute). Ab `0.4.0` setzt der Server in **keinem** OTel-Span ein `session_id`-Attribut; Session-Suche in Traces läuft ausschließlich über `mtrace.session.correlation_id` (verbindlicher Vertrag in §2.5; Test-Anker `TestHTTP_Span_DoesNotSetSessionIDAttribute`). Diese Verschärfung gilt nicht nur für den `POST /api/playback-events`-Span, sondern für alle vom Server erzeugten Spans im `apps/api`-Pfad.

### 2.2 Counter

Der Use Case ruft am Eintritt jedes `RegisterPlaybackEventBatch`-Aufrufs einen frameworkneutralen `Telemetry`-Port auf (Architektur §3.3, §5.3). Der OTel-Adapter mappt diesen Aufruf auf einen Counter:

| OTel-Counter | Typ | Aufrufstelle | Attribute |
|---|---|---|---|
| `mtrace.api.batches.received` | `Int64Counter` | Use Case Step 0 (vor Auth) | (keine — Counter ist label-frei wie die vier Pflichtcounter aus §2.4) |

**Naming-Translation**: das OTel-→-Prometheus-Mapping ersetzt `.` durch `_`; in Prometheus erscheint der Counter als `mtrace_api_batches_received` (vom OTLP-Exporter automatisch konvertiert). Smoke-Test-Regex `^mtrace_.+` aus `plan-0.1.2.md` §4 deckt sowohl den translated OTel-Counter als auch die direkten Prometheus-Counter aus `adapters/driven/metrics` ab.

**Cardinality-Beschluss (`plan-0.4.0.md` §8.2, Variante (a))**: `mtrace.api.batches.received` wird **ohne** `batch.size`-Attribut inkrementiert. Der Counter läuft im Use-Case Step 0 vor jeder Validierung — eine `batch.size = len(in.Events)`-Annotation würde im Reject-Pfad (`events.length > 100` → `422`) eine unbegrenzte Wertedomäne erzeugen, die per Prometheus-Naming-Translation als `batch_size`-Label auf `mtrace_api_batches_received` landen würde. Die Per-Request-Sicht „wie groß war dieser Batch?" bleibt über das Span-Attribut `batch.size` auf dem `http.handler POST /api/playback-events`-Span (siehe §2.1) erhalten — Span-Cardinality ist sample-basiert und im Cardinality-Vertrag aus §3 nicht bindend. Über `mtrace_api_batches_received / mtrace_playback_events_total` lässt sich der Mittelwert weiterhin abschätzen. Smoke-Schutz gegen Re-Introduction: `batch_size` ist in `scripts/smoke-observability.sh` zur Forbidden-Liste aus §3.1 hinzugefügt.

### 2.3 Resource-Attribute

`adapters/driven/telemetry/Setup` registriert für Provider-Resource folgende Pflicht-Attribute (Architektur §5.3):

| Attribut | Wert | Quelle |
|---|---|---|
| `service.name` | `m-trace-api` | Konstante in `cmd/api/main.go` |
| `service.version` | `<release-tag>` (z. B. `0.1.0`) | Build-Info / ENV |
| `mtrace.component` | `api` | Konstante in `Setup` |

Die Default-Resource wird mit `resource.Default()` zusammengeführt, damit `process.*`-Attribute (PID, executable.name) automatisch ergänzt werden.

### 2.4 Beziehung zu Prometheus-Metriken

| Prometheus-Counter (Pflicht laut API-Kontrakt §7) | Quelle im Code | Erfassung |
|---|---|---|
| `mtrace_playback_events_total` | `adapters/driven/metrics.PrometheusPublisher.EventsAccepted(n)` | Step 10 — pro Batch-Event mit Status `202` |
| `mtrace_invalid_events_total` | `…InvalidEvents(n)` | Step 5/6/7/8 — pro Event mit Status `400`/`422` (siehe Lastenheft §7.9 Mindestmetriken-Hinweis nach Patch `1.1.2`) |
| `mtrace_rate_limited_events_total` | `…RateLimitedEvents(n)` | Step 4 — bei Rate-Limit-Treffer |
| `mtrace_dropped_events_total` | `…DroppedEvents(n)` | nur Backpressure-Drops, **nicht** synchrone Persistenz-Fehler (siehe API-Kontrakt §7) |

Zusätzlich zu den vier Pflicht-Countern werden in `0.1.2` die Mindestmetriken aus Lastenheft §7.9 instrumentiert (`mtrace_active_sessions`, `mtrace_api_requests_total`, …). Der OTel-translated Counter `mtrace_api_batches_received` (Quelle: `mtrace.api.batches.received` aus §2.2) ist ab `0.4.0` Tranche 7 ebenfalls label-frei — derselbe Cardinality-Vertrag wie für die vier Pflichtcounter (`__name__`/`instance`/`job`-Whitelist; jeder zusätzliche Label-Key ist release-blockierend).

### 2.5 Trace-Korrelation in `0.4.0`

> Bezug: RAK-29; RAK-32; ADR-0002 §8.1 (Schema-Spalten); `plan-0.4.0.md` §3.1.

`0.4.0` führt zwei getrennte Korrelations-Konzepte ein, damit Tempo-Sichtbarkeit (RAK-31, optional) und Tempo-unabhängige Dashboard-Timeline (RAK-32, Pflicht) sauber entkoppelt sind.

**Hybrid-Trace-ID-Strategie.** Player-SDK propagiert optional einen W3C-`traceparent`-Header gemäß [W3C Trace Context](https://www.w3.org/TR/trace-context/). Server-Verhalten:

| Eingehender `traceparent` | Server-Aktion | Resultierender Span |
|---|---|---|
| valider Header (Format, Version, Flags) | übernimmt `trace_id` und `parent_span_id` aus dem Header | Child-Span mit übernommenen Trace-Identifiers |
| ungültiger Header (Parse-Fehler) | generiert eigene `trace_id`; Header wird **nicht** auf 4xx geführt | Root-Span; Span-Attribut `mtrace.trace.parse_error=true` |
| kein Header | generiert eigene `trace_id` | Root-Span ohne Parse-Error-Attribut |

`trace_id` und `parent_span_id` aus dem `traceparent`-Header werden serverseitig defensiv geprüft (Hex-Form, Längen 32/16, Version `00`); jeder Verstoß landet im Server-Fallback-Pfad. Der Header-Name ist HTTP-konform case-insensitiv. Führende und nachfolgende OWS (Spaces, Tabs) im Header-Wert werden bereits vom HTTP-Wire-Layer der Go-`net/http`-Standardbibliothek entfernt, bevor der Wert das Backend erreicht; ein OWS-umschlossener, sonst valider W3C-Wert wird daher als gültig akzeptiert und führt zur normalen Child-Span-Übernahme. Das Backend führt selbst kein zusätzliches Trim durch und verlässt sich für die OWS-Normalisierung ausschließlich auf den Wire-Layer; ein durchgereichter OWS-Wert (z. B. von einem Reverse-Proxy mit abweichender Header-Verarbeitung) fällt am defensiven `len == 55`-Check des Parsers auf den parse_error-Pfad zurück und wird wie jeder andere Format-Verstoß behandelt. Verbindlicher Vertragstext und Test-Anker stehen in `spec/backend-api-contract.md` §1.

**`trace_id` ≠ `correlation_id`.** Beide sind getrennte Konzepte mit klarer Verantwortung:

| Feld | Persistenz | Quelle | Lebensdauer | Konsumenten |
|---|---|---|---|---|
| `trace_id` | `playback_events.trace_id` (TEXT, nullable, 32 Hex-Zeichen) | SDK-`traceparent` oder server-generiert pro Batch | pro Batch (= ein Server-Span) | Tempo (RAK-31, optional); Cross-System-Trace-Suche |
| `correlation_id` | `stream_sessions.correlation_id` und `playback_events.correlation_id` (TEXT; für ab §3.2 verarbeitete Events gesetzt, historische Leerwerte möglich) | server-generiert beim allerersten Event einer Session (UUIDv4 oder vergleichbar) oder per Self-Healing beim nächsten Event einer Legacy-Session | pro Session, konstant über alle ab §3.2 geschriebenen Batches | Dashboard-Timeline (RAK-32) — Tempo-unabhängig |

`correlation_id` ist die **durable Source-of-Truth** für die Korrelation einer Session über mehrere Batches hinweg. Drei aufeinanderfolgende Batches mit gleicher `session_id` produzieren drei verschiedene `trace_id`-Werte (jeder Batch ein Trace), aber dieselbe `correlation_id`.

**Legacy-Grenze:** Tranche 2 führt kein historisches Backfill für vor §3.2 persistierte `playback_events.correlation_id` aus. Bestehende Sessions ohne `stream_sessions.correlation_id` werden beim nächsten Event per Self-Healing nachgezogen; ältere Events derselben Session können im Read-Pfad weiter eine leere `correlation_id` tragen und sind ein degradierter Legacy-Fall, nicht die Norm für neu verarbeitete Events.

**Span-Modell: ein HTTP-Request-Span pro Batch.** Keine Child-Spans pro Event (Cardinality-Schutz). Pflicht- und optionale Attribute am Server-Span für `POST /api/playback-events`:

| Attribut | Pflicht | Wertebereich | Bedeutung |
|---|---|---|---|
| `mtrace.project.id` | Pflicht für accepted Batches und für jeden Pfad, in dem der Use-Case-Resolver ein Project erfolgreich aufgelöst hat; **bewusst unset** für Rejects vor Project-Auflösung (z. B. `auth_error` durch fehlenden/ungültigen Token) | Allowlist aus dem Use-Case-Resolver | identifiziert das Project; Test-Anker `TestHTTP_Span_SingleSessionBatch_SetsCorrelationID` |
| `mtrace.batch.size` | ja | int ≥ 0 | Anzahl Events im Batch |
| `mtrace.batch.outcome` | ja | `accepted`, `invalid`, `rate_limited`, `auth_error`, `too_large`, `error` | Klassifikation des HTTP-Outcomes; Mapping zu API-Kontrakt §5 unten |
| `mtrace.batch.session_count` | ja | int ≥ 0 | Anzahl distinkter `session_id` im Batch |
| `mtrace.session.correlation_id` | bei Single-Session-Batch (`session_count == 1`); **nicht gesetzt** sonst (kein Empty-String, keine Komma-Liste) | UUIDv4 als String | erlaubt Tempo-Suche nach Sessions ohne `session_id` zu exposen |
| `mtrace.trace.parse_error` | optional | Boolean | gesetzt, wenn `traceparent` ungültig war |
| `mtrace.time.skew_warning` | optional | Boolean | gesetzt, wenn mindestens ein Event im Batch `\|client_timestamp - server_received_at\| > 60s` (siehe §5.3) |

**`session_id`-Span-Attribut-Verbot.** Ab `0.4.0` setzt der Server in **keinem** OTel-Span ein `session_id`-Attribut (weder unter dem rohen Schlüssel `session_id` noch in den Varianten `mtrace.session.id` / `mtrace.session_id`). Single-Session-Suche in Traces läuft ausschließlich über `mtrace.session.correlation_id`. Die historische Aussage aus §2.1, dass `session_id` als Span-Attribut zulässig sei, gilt **nur** für `0.1.x` und ist nicht Teil des Tranche-2-Vertrags. Test-Anker: `TestHTTP_Span_DoesNotSetSessionIDAttribute`.

**Outcome → HTTP-Status-Mapping** (Validierungs-Reihenfolge aus API-Kontrakt §5):

| `mtrace.batch.outcome` | HTTP-Status | API-Kontrakt §5 |
|---|---|---|
| `auth_error` | `401 Unauthorized` | Header fehlt; Token unbekannt; Project-Mismatch; Project unbekannt |
| `too_large` | `413 Payload Too Large` | Body > 256 KB |
| `rate_limited` | `429 Too Many Requests` | Rate-Limit-Treffer |
| `invalid` | `400 Bad Request` (Schema-Version, JSON-Form) **oder** `422 Unprocessable Entity` (Batch-Form/-Größe, Event-Pflichtfeld) | siehe §5 Steps 5–8 |
| `accepted` | `202 Accepted` | Happy-Path |
| `error` | `5xx` | unerwartete Fehler (Persistenz, Telemetrie, etc.) |

**Cardinality-Regel.** Weder `trace_id`, `correlation_id` noch `span_id` werden als Prometheus-Labels verwendet. Span-Attribute (kontrolliert), Event-Persistenz-Spalten (durable) und Wire-Format-Felder (optional) sind die einzigen Konsumenten. Verstöße sind release-blocking; der CI-Cardinality-Smoke (`make smoke-observability`, `scripts/smoke-observability.sh`) prüft die Pflicht-Counter aus §2.4 auf hochkardinale Labels.

**Sampling-Auswirkung.** Server-Span pro Batch ist niedrige Cardinality (eine Span pro HTTP-Request). Auch ohne Sampling bleibt Tempo-Storage in 0.4.0 unauffällig. Spans werden via OTLP exportiert, wenn das Tempo-Profil aktiv ist (siehe `plan-0.4.0.md` §6). Ohne Profil und mit unset `OTEL_*` nutzt der autoexport-Fallback einen No-Op-Pfad ohne Exportversuch und ohne Log-Ausgabe; Logs entstehen nur, wenn bewusst ein Debug-/Console-Exporter oder der Collector-Debug-Exporter konfiguriert ist.

### 2.6 Trace-Suche in Tempo (`0.4.0`, optional unter `tempo`-Profil)

Wenn das `tempo`-Compose-Profil aktiv ist (`make dev-tempo`, `plan-0.4.0.md` §6), exportiert der OTel-Collector Spans nach Tempo. Für die Trace-Suche im Lab gilt verbindlich:

| Such-Pfad | Suchwert | Span-Attribut / Feld | Zweck |
|---|---|---|---|
| **Primary (Session-Korrelation)** | `correlation_id` aus API-Read-Antwort, Dashboard oder SQLite | Span-Attribut `mtrace.session.correlation_id` | Alle Server-Spans einer Single-Session-Batch-Verarbeitung; Tempo-API: TraceQL `GET /api/search?q={ span.mtrace.session.correlation_id = "<UUID>" }&start=<unix>&end=<unix>` |
| **Legacy (Session-Korrelation)** | `correlation_id` aus API-Read-Antwort, Dashboard oder SQLite | Span-Attribut `mtrace.session.correlation_id` | Fallback nur für ältere Tempo-Setups: `GET /api/search?tags=mtrace.session.correlation_id=<UUID>&start=<unix>&end=<unix>` |
| **Sekundär (Batch-Korrelation)** | `trace_id` aus `playback_events.trace_id` (nur Single-Event-Batch oder explizit gesuchter Batch) | Tempo-Trace-ID (`trace_id`-Hex) | Tempo-API: `GET /api/traces/<trace_id>` |

**Suchfenster-Pflicht.** Tempo-Search-Aufrufe müssen `start` und `end` als Unix-Zeitfenster setzen. `/api/search?tags=...` ohne Zeitraum ist nicht normativ; Tempo 2.x kann je nach Default-Resolution leere Treffer liefern, obwohl der Span bereits ingestiert ist.

**Multi-Trace-Disclaimer.** Eine Session kann mehrere `trace_id`-Werte haben — jeder Batch erzeugt einen neuen Server-Span (siehe §2.5). `trace_id` ist daher **kein Session-Schlüssel**. Eine Session-übergreifende Tempo-Suche per `trace_id` ist immer batchspezifisch; die vollständige Session-Trace-Liste (sortiert nach `ingest_sequence`) liefert nur das Dashboard plus Read-Pfad, nicht Tempo.

**Single-Session-Batch-Pflicht für `mtrace.session.correlation_id`.** Das Span-Attribut wird ausschließlich bei `mtrace.batch.session_count == 1` gesetzt; bei Multi-Session-Batches bleibt es unset (keine Komma-Liste, kein Empty-String — siehe §2.5-Tabelle). Tempo-Suche nach Multi-Session-Batches läuft daher nicht über `mtrace.session.correlation_id`, sondern muss aus dem Read-Pfad (Dashboard/SQLite) eine Single-Session-Batch-Span finden, die zur Session gehört. Diese bewusste Grenze schließt aus, dass eine Session-`correlation_id` versehentlich an einen Span gebunden wird, der mehrere Sessions umfasst.

**Tempo ist Debug-Tiefe, nicht Read-Pfad.** Die Dashboard-Session-Timeline (RAK-32, plan-0.4.0 §5) ist Tempo-unabhängig. Tempo erweitert die Sichtbarkeit auf Span-Ebene (Header-Verarbeitung, Outcome-Klassifikation, Resource-Attribute); jede Aussage über *Event-Persistenz* oder *Session-State* bleibt im Read-Pfad und in SQLite verbindlich.

---

## 3. Cardinality-Regeln

> Bezug: F-95..F-100 (Lastenheft §7.10), F-101..F-105 (MVP-Variante).

### 3.1 Verbotene Prometheus-Labels

Folgende Werte dürfen **nie** als Prometheus-Label erscheinen, weil sie Cardinality-Explosion verursachen oder Datenschutz-/Trace-Identifier sind. Diese Liste ist die normative Quelle für `scripts/smoke-observability.sh` und für jeden neuen `mtrace_*`-Metrik-Vorschlag — sie deckt die Mindest-Verbote aus API-Kontrakt §7 vollständig ab; kürzere Beispiel-Listen reichen nicht als Abnahme:

| Label | Begründung | Bezug |
|---|---|---|
| `session_id` | hochkardinale Pseudonym-IDs; potentiell Millionen Sessions/Tag. | F-96, F-105; API-Kontrakt §7 |
| `user_agent` | quasi-unbegrenzter Wertebereich. | API-Kontrakt §7 |
| `segment_url`, `manifest_url`, beliebige `*_url` / `*_uri` | URL-Variation pro CDN/Player-Adaptation. | API-Kontrakt §7 |
| `client_ip` | DSGVO-Risiko + hohe Cardinality. | API-Kontrakt §7 |
| beliebige `project_id` | bei Multi-Tenant explosiv; nur kontrollierte Allowlist erlaubt. | API-Kontrakt §7 |
| `viewer_id`, `request_id` | hochkardinale Per-Request-Identifier. | API-Kontrakt §7 |
| `trace_id`, `span_id`, `correlation_id` | Trace-/Session-Korrelations-Identifier; Cross-System-Suche läuft über Tempo/Read-Pfad, nicht Prometheus. | API-Kontrakt §7; §2.5 |
| `token`, `authorization`, beliebige `*_token` / `*_secret` | Credentials gehören niemals in eine Metrik-Serie. | API-Kontrakt §7 |
| `batch_size` | unbegrenzte Integer-Domäne, weil der OTel-Counter `mtrace.api.batches.received` vor der `MaxBatchSize=100`-Validierung läuft (siehe §2.2). Ab `0.4.0` Tranche 7 entfernt; `batch.size` bleibt nur als Span-Attribut. | §2.2; `plan-0.4.0.md` §8.2 |
| SRT-Source-Labels (`id`, `path`, `remoteAddr`, `state`, `connection_id`, `stream_id` als Per-Stream-Identifier) sowie URL-/IP-/Token-Varianten aus MediaMTX `/v3/srtconns/list` | hochkardinale Per-Verbindung-Identifier; werden in SQLite/OTel-Spans persistiert, **nie** als Prometheus-Label. | §7; `plan-0.6.0.md` §4 |

Erlaubt sind ausschließlich die bounded Aggregat-Labels aus §3.2. Die Forbidden-Liste in `scripts/smoke-observability.sh` deckt die Tabelle plus generische Suffixe (`_url`, `_uri`, `_token`, `_secret`) defensiv ab; jeder Treffer ist release-blockierend.

### 3.2 Erlaubte Aggregat-Labels

Erlaubt sind Labels mit kontrolliertem, kleinem Wertebereich. Jede neue `mtrace_*`-Metrik mit Vector-Labels muss ihren Labelsatz in dieser Tabelle (oder einer ausgewiesenen Erweiterung) belegen — beliebige Labelnamen sind nicht zulässig:

| Label | Wertebereich | Beispiel | Erlaubt auf |
|---|---|---|---|
| `event_type` | feste Enum aus §1.3 | `rebuffer_started`, `playback_error` | per-event-type-Aggregate (zukünftige Metriken) |
| `outcome` | feste Enum | `accepted`, `invalid`, `rate_limited`, `dropped`, `analyzer_unavailable`, `analyzer_error`, … | `mtrace_analyze_requests_total{outcome,code}` (Tranche 6 §3) |
| `code` | feste Fehler-/Ergebnis-Code-Domäne pro Metrik | `invalid_request`, `analyzer_unavailable`, `fetch_blocked` | `mtrace_analyze_requests_total{outcome,code}` |
| `health_state` | feste Enum aus §7.4: `healthy`, `degraded`, `critical`, `unknown` | `degraded` | `mtrace_srt_health_samples_total{health_state}` (`plan-0.6.0.md` §4) |
| `source_status` | feste Enum aus §7.5: `ok`, `unavailable`, `partial`, `stale`, `no_active_connection` | `stale` | `mtrace_srt_health_collector_runs_total{source_status}` (`plan-0.6.0.md` §4) |
| `instance` / `job` | OTel/Prometheus-Standard | `api:8080` | alle Metriken (Target-Metadaten) |

Die vier Pflichtcounter (`mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total`, `mtrace_dropped_events_total`) und der OTel-translated Counter `mtrace_api_batches_received` tragen **gar keine** fachlichen Vector-Labels (siehe API-Kontrakt §7 und §2.4). `batch_size` ist explizit nicht in der Allowlist (siehe §3.1) — die Per-Request-Sicht „Batchgröße" lebt nur am Span.

Prometheus-Series pro Mindest-Counter sollten ≤ einstellige Anzahl sein. RAK-9-Smoke-Test (`plan-0.1.2.md` §4) prüft dies via `count(count by (...) (...))`-PromQL; der verschärfte Smoke aus `plan-0.4.0.md` §7.3 (`scripts/smoke-observability.sh`) prüft pro Pflichtcounter zusätzlich, dass das Labelset auf `__name__`/`instance`/`job` beschränkt ist.

### 3.3 Trennung Aggregat vs Per-Session

Per-Session-Daten (Stream-Health, Event-Timeline, Trace-Identifier) gehen **nicht** in Prometheus. Die drei Backends teilen die Verantwortung wie folgt — diese Tabelle ist die normative Quelle für `README.md`, `docs/user/local-development.md` und jede neue Telemetrie-Diskussion:

| Backend | Daten | Cardinality-Verträglichkeit | Konsumenten |
|---|---|---|---|
| **Prometheus** | Aggregat-Metriken (counts, rates, optional gauges); ausschließlich bounded Aggregat-Labels aus §3.2 | hart begrenzt auf wenige Serien pro Metrik; Forbidden-Liste aus §3.1 ist release-blockierend | Grafana-Dashboards (`observability/grafana/dashboards/m-trace-overview.json`); Alerting; RAK-9-Cardinality-Smoke |
| **SQLite** (ADR-0002) | Session-/Event-Historie mit allen Per-Session-Identifiern (`session_id`, `correlation_id`, `trace_id`, `span_id`, redacted URLs, `network_signal_absent`-Boundary-Records) | unbeschränkt — durable Event-Store, kein Cardinality-Vertrag | Dashboard-Session-Timeline (RAK-32); Read-Pfad `GET /api/stream-sessions/...`; SDK-Cursor-Pagination |
| **OTel/Tempo** | Per-Request-Trace-Spans mit allen Span-Attributen (`mtrace.session.correlation_id`, `batch.size`, `mtrace.batch.outcome`, …); ein Server-Span pro Batch | sample-basiert; Span-Cardinality ist im Cardinality-Vertrag aus §3.1 nicht bindend | Tempo-Trace-Suche (`make dev-tempo`, RAK-31, optional); Span-Ebene-Debugging beim Header-Verarbeitung-/Outcome-Pfad |

Diese Trennung ist die zentrale Architektur-Aussage von F-97 und Lastenheft §7.10. Praktische Konsequenzen:

- **Aggregate-Anfrage** („wie viele 4xx-Antworten in den letzten 5 Minuten?") → Prometheus; nie SQLite, nie Tempo.
- **Konkrete-Session-Anfrage** („zeig mir die Timeline von `session_id = abc-123`") → Read-Pfad/Dashboard auf SQLite; nie Prometheus, nie zwingend Tempo.
- **Cross-System-Trace-Vertiefung** („was ist im Server-Span passiert, der diesen Batch verarbeitet hat?") → Tempo (falls aktives Profil); Per-Span-Detail, nicht Per-Session-Aggregat.

Tempo ist daher Debug-Tiefe, **nicht** Read-Pfad. Die Dashboard-Session-Timeline (RAK-32) ist ausdrücklich Tempo-unabhängig (siehe §2.6); jede Aussage über Event-Persistenz oder Session-State bleibt in SQLite verbindlich.

### 3.4 Datenschutz

Telemetrie-Modell und Datenschutz werden gemeinsam betrachtet (F-100):

- `session_id` ist pseudonym (NF-40 Lastenheft §8.6).
- IP-Adressen werden nicht unnötig persistiert; falls erfasst, dann nur in OTel-Spans, nicht in Prometheus-Labels.
- User-Agent-Felder dürfen reduzierbar sein (z. B. nur Major-Version).
- GDPR-konformer Betrieb: Event-Store muss eine Löschanfrage pro `session_id` bedienen können — Implementierung über das `EventRepository` in der jeweiligen Persistenz-Variante.

---

## 4. Backpressure und Limits

> Bezug: F-118..F-123; API-Kontrakt §3, §5, §6.

### 4.1 Batch-Größe

| Limit | Wert | Bezug |
|---|---|---|
| Mindest-Events pro Batch | 1 (leerer Batch → `422`) | F-118; API-Kontrakt §5 Step 6 |
| Maximal-Events pro Batch | 100 | F-118; API-Kontrakt §3, §5 Step 7 |
| Maximal-Body-Größe | 256 KB | API-Kontrakt §5 Step 2 |

Das SDK muss Batches selbst aufteilen, wenn lokal mehr Events vorliegen.

### 4.2 Rate-Limit-Modell

Token-Bucket pro drei Dimensionen (F-110, post-`1.0.2`-Mindestdienste-Klärung):

| Dimension | Default-Quote | Bemerkung |
|---|---|---|
| `project_id` | 100 Events/s | Spike-Pattern |
| `client_ip` | 100 Events/s | Schutz gegen einzelne Misuse-Browser |
| `origin` | 100 Events/s | Pflicht für Browser-Traffic ab `0.1.1` (siehe `plan-0.1.0.md` §5.1) |

Konfigurationsweise: Konstanten in `cmd/api/main.go` oder ENV-Variablen analog Spike. Verteilt-konsistente Rate-Limiter sind Bonus (F-110-Erweiterung in späteren Phasen).

### 4.3 Drop-Politik (Backpressure)

`mtrace_dropped_events_total` ist laut API-Kontrakt §7 ausschließlich für **interne Backpressure-Drops** reserviert (z. B. überlaufender Async-Channel-Puffer). Synchron fehlgeschlagenes `Append` ist **kein** Drop und inkrementiert den Counter nicht (F-122).

In `0.1.0` mit synchron-blockierendem `EventRepository.Append` gibt es keinen Backpressure-Pfad — der Counter darf konstant `0` bleiben (Lastenheft §7.9-Hinweis: „Metrik muss aber existieren"). Mit Wechsel auf einen Async-Persistenz-Pfad in einer Folge-Phase würde der Counter relevant.

### 4.4 SDK-Konfigurierbarkeit

Das SDK muss Sampling und Batch-Größe konfigurierbar anbieten (F-123, MVP-Soll):

| SDK-Parameter | Bedeutung | Default-Vorschlag |
|---|---|---|
| `sampleRate` | Anteil der erzeugten Events, die gesendet werden (0..1). | `1.0` (alle Events) |
| `batchSize` | maximale Events pro Batch, hart auf ≤ 100 begrenzt. | `10` |
| `flushIntervalMs` | maximale Wartezeit, bevor ein nicht-voller Batch geflusht wird; `0` deaktiviert den Timer. | `5000` |
| `maxQueueEvents` | lokales Queue-Limit für normale Playback-Events, bevor neue normale Events verworfen werden. | `1000` |
| `endpoint` | Backend-URL. | konfigurierbar, kein Default |

---

## 5. Time-Stempel und Sequenz-Ordering

> Bezug: F-124..F-130.

### 5.1 Pflicht- und Optional-Felder

Vom Client gesetzt:

- **`client_timestamp`** (Pflicht, F-124): Erzeugungszeitpunkt am Client; RFC3339 mit Millisekunden-Genauigkeit.
- **`client_time_origin`** (optional, F-126): Setup-Zeitpunkt des SDK; erlaubt skew-toleranten Latenz-Vergleich (`client_timestamp - client_time_origin` ist robust gegen Wall-Clock-Skew).
- **`sequence_number`** (optional, F-127): monoton aufsteigend pro Session.

Vom Server gesetzt:

- **`server_received_at`** (Pflicht, F-125): vom HTTP-Adapter direkt nach Body-Parsen gestempelt; in den Domain-`PlaybackEvent`-Datensatz übernommen.
- **`ingest_sequence`** (Pflicht, siehe `plan-0.1.0.md` §5.1): monoton aufsteigender Counter pro `apps/api`-Prozess; finaler Tie-Breaker für Cursor-Pagination.

### 5.2 Ordering innerhalb einer Session

Kanonische API- und Cursor-Reihenfolge:

1. `server_received_at` (Server-Eingangszeit, restart-stabil persistiert).
2. `sequence_number` (falls gesetzt — Client kontrolliert; sortiert nur innerhalb derselben Server-Zeitgruppe).
3. `ingest_sequence` (durabler Tie-Breaker und Pagination-Schlüssel).

Für fachliche Session-Analyse darf das Dashboard zusätzlich die
Client-Sequenz visualisieren. Die API liefert jedoch konsistent
`(server_received_at, sequence_number, ingest_sequence)`, damit Cursor nicht
durch fehlende oder fehlerhafte Client-Sequenzen instabil werden.

### 5.3 Latenzberechnung und Time-Skew

- Latenzen dürfen niemals blind aus reiner Client-Zeit abgeleitet werden (F-129) — Client-Uhren divergieren in der Praxis um Sekunden bis Minuten.
- Bevorzugt: Latenz = `server_received_at - client_time_origin` (skew-tolerant), nicht `server_received_at - client_timestamp`.
- Auffälliger Skew (F-130): liegt `|client_timestamp - server_received_at|` über einem Schwellwert (in `0.4.0` Konstante 60 s, kein Configuration-Item — siehe `plan-0.4.0.md` §3.2), markiert das Backend den Server-Span mit dem Attribut `mtrace.time.skew_warning=true` (siehe §2.5). Persistenz des Skew-Flags auf Event-Ebene (Domain-Flag, dedizierte Schema-Spalte, Dashboard-Anzeige) ist in `0.4.0` explizit deferred — siehe `docs/planning/open/risks-backlog.md` R-5 für das Folge-Item.

---

## 6. Schema-Versionierung

> Bezug: F-114..F-117.

### 6.1 Versionsfeld

Jeder Batch trägt eine `schema_version` (siehe §1.1). Format: SemVer-`MAJOR.MINOR` (Patch wird im Wire nicht ausgewertet).

| Wert in `0.1.x` | Bedeutung |
|---|---|
| `1.0` | aktuelle Wire-Format-Version laut diesem Dokument. |

### 6.2 Evolution-Regeln

- **Neue Felder** (F-114): müssen abwärtskompatibel sein — bestehende Clients dürfen sie ignorieren.
- **Unbekannte Felder** (F-115): das Backend darf nicht mit `400`/`422` reagieren, sondern muss sie ignorieren (Forward-Compatibility für Clients neuerer Schema-Version).
- **Entfernte Felder** (F-116): müssen über mindestens eine **Minor-Version** toleriert werden — Backend akzeptiert das Feld weiterhin, ignoriert es aber.
- **Breaking Changes** (F-117): erfordern eine neue **Major-Version** der Schema-Wire-Form. Ältere Major-Versionen können temporär weiter angenommen werden, müssen aber explizit deprecated und mit Sunset-Datum versehen sein.

### 6.3 Backend-Verhalten bei Schema-Versions-Mismatch

- `schema_version` ≠ `1.0` → API-Kontrakt §5 Step 5 → `400 Bad Request`. Strikte Major.Minor-Prüfung, kein Range-Match in `0.1.x`.
- Mit künftiger Multi-Version-Unterstützung wird das Step 5-Wording erweitert; Folge-ADR dokumentiert die Übergangsstrategie.

---

## 7. SRT-Health-Modell (`0.6.0`)

> Bezug: Lastenheft §4.3, §13.8 RAK-41..RAK-46;
> [`plan-0.6.0.md`](../docs/planning/in-progress/plan-0.6.0.md)
> §2 (Quellen-Entscheidung), §4 (Datenmodell + Storage + OTel-Vertrag);
> [`spec/contract-fixtures/srt/mediamtx-srtconns-list.json`](contract-fixtures/srt/mediamtx-srtconns-list.json).

SRT-Health-Metriken sind **getrenntes Verbindungs-/Ingest-Signal** und nicht
mit Player-Playback-Events vermischt (plan-0.6.0 §0.1). Die Quelle ist in
`0.6.0` MediaMTX-Control-API über HTTP (`/v3/srtconns/list`); `apps/api`
bleibt CGO-frei (R-2 in `risks-backlog.md` §1.2 aufgelöst).

### 7.1 Datenmodell

Ein **Sample** repräsentiert eine SRT-Verbindung zu einem Polling-Zeitpunkt
und ist in `apps/api` durable persistiert (SQLite, ADR-0002). Pflicht- und
Optional-Felder im Domain-Modell:

| Feld | Pflicht | Typ | Einheit / Bedeutung |
|---|---|---|---|
| `project_id` | Muss | string (bounded Project-Resolver) | Tenant-Anker; nicht als Prometheus-Label exposed (siehe §3.1). |
| `stream_id` | Muss | string | Lab-Stream-Name (z. B. `srt-test`); Per-Stream-Identifier, nur SQLite/OTel. |
| `connection_id` | Muss | string | Quellseitige Verbindungs-ID (in MediaMTX `items[].id`); Per-Verbindung-Identifier, nur SQLite/OTel. |
| `source_observed_at` | Soll | timestamp (RFC3339, ms) | Wann die Quelle den SRT-Zustand gemessen hat. **Optional**, weil MediaMTX-API in `0.6.0` keinen expliziten Timestamp liefert. |
| `source_sequence` | Pflicht ohne `source_observed_at` | string oder integer | Monotones Surrogat — z. B. `bytesReceived`-Counter aus dem Sample, Generation-ID, Sample-Window-Endzeit. Wird von der Stale-Bewertung als Source-Sequence gewertet. |
| `collected_at` | Muss | timestamp (RFC3339, ms) | Zeitpunkt des Polls durch den Collector (m-trace-eigene Uhr). Allein **nicht** ausreichend für Freshness-Bewertung. |
| `ingested_at` | Muss | timestamp (RFC3339, ms) | Zeitpunkt der SQLite-Persistenz. |
| `rtt_ms` | Muss | number | RTT in Millisekunden. Snapshot-Wert; bei MediaMTX-API: `msRTT`. |
| `packet_loss_total` | Muss | integer (counter, kumulativ) | Empfänger-Paketverlust seit Verbindungsstart. Bei MediaMTX-API: `packetsReceivedLoss`. |
| `packet_loss_rate` | Optional | number (0..1) | Verlustrate als Snapshot, falls Quelle sie zusätzlich liefert. Nicht release-blockierend, weil Counter-Diff abgeleitet werden kann. |
| `retransmissions_total` | Muss | integer (counter, kumulativ) | Empfänger-Retransmissions. Bei MediaMTX-API: `packetsReceivedRetrans`. Sender-seitige `packetsRetrans` ist optional. |
| `available_bandwidth_bps` | Muss | integer (bits/s) | Linkkapazitäts-Schätzung der Quelle. Bei MediaMTX-API: `mbpsLinkCapacity × 1_000_000`. **Caveat**: in localhost-/Loopback-Netzen liefert MediaMTX Werte im Gbps-Bereich, die kein realistischer „verfügbarer"-Wert sind — Health-Bewertung in §7.4 kompensiert das via `required_bandwidth_bps`-Vergleich. |
| `throughput_bps` | Optional | integer (bits/s) | Tatsächlich beobachteter Stream-Durchsatz. Bei MediaMTX-API: `mbpsReceiveRate × 1_000_000`. Erfüllt RAK-43 nicht allein. |
| `required_bandwidth_bps` | Optional | integer (bits/s) | Erwarteter Bandbreitenbedarf (aus Lab-Konfig oder Stream-Konfiguration). Ohne diese Schwelle darf `available_bandwidth_bps` angezeigt, aber **nicht** als Engpass bewertet werden. |
| `sample_window_ms` | Optional | integer | Zeitfenster für aus Countern abgeleitete Raten, falls relevant. |
| `source_status` | Muss | enum (§7.5) | `ok`, `unavailable`, `partial`, `stale`, `no_active_connection`. |
| `source_error_code` | Muss | enum (§7.5) | Stabile Fehlerklasse bei nicht-`ok`-Status. `none` bei `ok`. |
| `connection_state` | Muss | enum | `connected`, `no_active_connection`, `unknown`. Getrennt vom Quellenstatus, weil eine erreichbare Quelle ohne aktive Verbindung ein anderer Fall ist als eine nicht erreichbare Quelle. |
| `health_state` | Muss | enum (§7.4) | `healthy`, `degraded`, `critical`, `unknown`. Server-seitig berechnet aus den Pflicht-Werten plus Schwellen. |

### 7.2 Erweiterte SRT-Signale (deferred in `0.6.0`, sofern nicht ohne Zusatzrisiko aus der Quelle mitfallen)

Lastenheft §4.3 listet weitere SRT-Signale; plan-0.6.0 §0.1 priorisiert
explizit RAK-43-Pflichtwerte. Folgende Signale sind aus MediaMTX-API
verfügbar und können **als Zusatzfelder** im Datenmodell mitfallen,
sind aber nicht release-blockierend:

| Quellfeld | Bedeutung | Mapping (Vorschlag) |
|---|---|---|
| `msReceiveBuf` | Receiver-TSBPD-Buffer-Tiefe (ms) | `receive_buffer_ms` |
| `bytesReceiveBuf` | Receiver-Buffer-Bytes | `receive_buffer_bytes` |
| `packetsReceiveBuf` | Receiver-Buffer-Paketanzahl | `receive_buffer_packets` |
| `outboundFramesDiscarded` | Verworfene Frames | `frames_discarded_total` (counter) |
| `packetsReorderTolerance` | Reorder-Toleranz | `reorder_tolerance_packets` |

Send-/Receive-Buffer-Detail, Verbindungsstabilität, separater Link-
Health-Score und Failover-Zustände aus Lastenheft §4.3 bleiben
deferred.

### 7.3 Counter-vs-Rate und Sample-Window

- **Counter** (kumulativ ab Verbindungsstart): `packet_loss_total`,
  `retransmissions_total`, `frames_discarded_total`. Adapter speichert
  den absoluten Counter; Dashboard kann die Intervallrate aus zwei
  aufeinanderfolgenden Samples ableiten (Δ Counter / Δ
  `source_sequence`-Surrogat).
- **Snapshot** (Momentaufnahme): `rtt_ms`, `available_bandwidth_bps`,
  `throughput_bps`, `connection_state`, Buffer-Werte.
- **Reset-Verhalten**: Counter resetten bei Verbindungs-Reconnect
  (`connection_id`-Wechsel). Adapter erkennt Wechsel an neuer
  `connection_id` und beginnt mit neuem Counter-Verlauf.

### 7.4 Health-Bewertung

`health_state` ist server-seitig aus den Pflicht-Werten berechnet.
Schwellen sind dokumentiert und über Tests fixiert (Tranche 4 setzt
die finalen Schwellen):

| Zustand | Bedingung (Vorschlag, Tranche 4 final) |
|---|---|
| `healthy` | Alle Pflicht-Werte verfügbar; `rtt_ms < 100`; `packet_loss_total`-Δ pro Sample-Window unter 1 % der `bytesReceived`-Δ-äquivalenten Paketanzahl; `available_bandwidth_bps >= required_bandwidth_bps × 1.5` (oder kein `required_bandwidth_bps` bekannt → keine Bandbreiten-Bewertung). |
| `degraded` | `rtt_ms` zwischen 100 und 250 ms ODER Paketverlust 1–5 % ODER Retransmissions-Anteil > 0,5 %. |
| `critical` | `rtt_ms ≥ 250` ODER Paketverlust > 5 % ODER `available_bandwidth_bps < required_bandwidth_bps`. |
| `unknown` | `source_status ≠ ok`, oder Pflicht-Werte teilweise fehlen, oder Stale-Erkennung schlägt an. |

Bandbreiten-Health darf nur dann `degraded`/`critical` auslösen, wenn
`required_bandwidth_bps` bekannt ist. Ohne Schwelle wird die
Bandbreite nur angezeigt (siehe §7.1 `required_bandwidth_bps`).

### 7.5 Source-Status und Fehlerklassen

Stabile Codes (analog Probe-Befund aus
[`plan-0.6.0.md`](../docs/planning/in-progress/plan-0.6.0.md) §2.4):

| `source_status` | `source_error_code` | Auslöser |
|---|---|---|
| `ok` | `none` | Quelle erreichbar, alle Pflichtfelder gesetzt. |
| `no_active_connection` | `no_active_connection` | Quelle erreichbar, aber `items[]` enthält keine Verbindung mit erwartetem Pfad/State. |
| `partial` | `partial_sample` | Quelle erreichbar, Item gefunden, einzelne Pflichtfelder fehlen oder sind non-numeric. |
| `stale` | `stale_sample` | Quelle erreichbar, Pflichtfelder gesetzt, aber `bytesReceived` (oder das gewählte Source-Sequence-Surrogat) hat sich über N Polls nicht verändert, obwohl `state: publish`. |
| `unavailable` | `source_unavailable` | HTTP `4xx`/`5xx`, Connection refused, Timeout. |
| `unavailable` | `parse_error` | HTTP `200`, aber Body ist kein gültiges JSON oder Schema-Drift. |

### 7.6 Freshness-Strategie

- `source_observed_at` ist die Source-of-Truth, falls die Quelle ihn
  liefert. MediaMTX-API liefert ihn in `0.6.0` **nicht** — Adapter
  nutzt stattdessen `collected_at` plus ein **Source-Sequence-
  Surrogat**: monoton steigender `bytesReceived` zwischen Polls.
- **Stale-Erkennung**: identischer `bytesReceived` (oder gewähltes
  Surrogat) zwischen `N` aufeinanderfolgenden Polls trotz
  `connection_state = connected` → `source_status: stale` mit
  `source_error_code: stale_sample`. `N` ist konfigurierbar
  (Vorschlag Tranche 4: `N = 3`, ≈ 15 s bei 5-s-Polling).
- **Importzeit allein** (`collected_at` oder `ingested_at`) darf
  Freshness niemals beweisen. Wiederholt importierte Altwerte mit
  identischem Surrogat sind stale, auch wenn `collected_at` neu ist.

### 7.7 Cardinality-Vertrag

- `health_state` und `source_status` sind in §3.2 als bounded
  Aggregat-Labels freigegeben.
- Per-Verbindung-Felder (`stream_id`, `connection_id`, `id`,
  `remoteAddr`, `path`, `state`) sind in §3.1 verboten.
- Rohmetriken aus MediaMTX werden nicht in den Projekt-Prometheus
  gescraped (plan-0.6.0 §0.1). MediaMTX-eigene Prometheus-Targets
  bleiben außerhalb des m-trace-Stacks.

Erlaubte `mtrace_srt_*`-Aggregate (Tranche 6 finalisiert):

| Metrik | Typ | Labels |
|---|---|---|
| `mtrace_srt_health_samples_total` | Counter | `health_state` |
| `mtrace_srt_health_collector_runs_total` | Counter | `source_status` |
| `mtrace_srt_health_collector_errors_total` | Counter | `source_error_code` |

### 7.8 OTel-Modell

- Span pro Collector-Run: `mtrace.srt.health.collect`. Attribute:
  `mtrace.srt.connection_id`, `mtrace.srt.stream_id`, `mtrace.srt.health_state`,
  `mtrace.srt.source_status`, `mtrace.srt.rtt_ms`,
  `mtrace.srt.available_bandwidth_bps`. Keine Token-/IP-Felder.
- Counter (translated to Prometheus über §2.4): identisch zu §7.7.
- Resource-Attribute folgen §2.3.

### 7.9 Datenschutz

- `connection_id` und `remoteAddr` sind Per-Verbindung-Identifier;
  in MediaMTX-Lab sind das Docker-interne IPs ohne PII-Bezug. In
  produktiven Setups sind das ggf. öffentliche IPs — Persistenz nur
  in SQLite/OTel-Spans, **niemals** in Prometheus, und Retention
  folgt §3.4 plus dem allgemeinen GDPR-Pfad (`EventRepository`-
  Löschanfrage-Äquivalent für `SrtHealthRepository`).
- MediaMTX-Auth-Credentials für die API (z. B. `authInternalUsers`-
  Pass) gehören in ENV / Geheimnis-Store, nicht in Code oder Logs.
