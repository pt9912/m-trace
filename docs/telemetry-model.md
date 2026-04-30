# Telemetry-Model — m-trace

> **Status**: Verbindlich für `0.1.x`. Wire-Format und Backpressure-Limits sind harte Voraussetzung für `0.1.1` (Player-SDK); OTel-Modell und Cardinality-Regeln sind harte Voraussetzung für `0.1.2` (Observability-Stack).  
> **Bezug**: [Lastenheft `1.1.5`](./lastenheft.md) §7.10 (Cardinality), §7.11 (Telemetry Ingest, Event-Schema, SDK-Budget); [Roadmap](./roadmap.md) §2 Schritt 6; [Plan `0.1.0`](./plan-0.1.0.md) §3.5; [API-Kontrakt](./spike/backend-api-contract.md); [Architektur](./architecture.md) §5.

## 0. Zweck

Beschreibt das **Datenmodell** der Telemetrie — Wire-Format, OTel-Schema, Cardinality-Regeln, Time-Stempel-Konventionen, Backpressure-Politik. Implementierungs-/Setup-Aspekte (strukturierte Logs, Health-Endpoint, Prometheus- und Grafana-Konfiguration) gehören in [`plan-0.1.2.md`](./plan-0.1.2.md), nicht hierher.

Drei Wirkungsebenen pro Telemetrie-Datum:

1. **Wire** — was zwischen Browser-SDK und API über das Netz fliegt (§1).
2. **Ingest** — wie das Backend ankommende Daten validiert, normalisiert und persistiert (siehe `apps/api/hexagon/application/RegisterPlaybackEventBatch` und [API-Kontrakt §5](./spike/backend-api-contract.md)).
3. **Beobachtung** — was als OTel-Span/-Counter und als Prometheus-Metrik nach außen tritt (§2, §3).

---

## 1. Wire-Format Player-Events

> Bezug: F-106..F-115; API-Kontrakt §3; Lastenheft §7.11.1–§7.11.3.

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
| `sdk.name` | string, nicht-leer | identifiziert das SDK (z. B. `@m-trace/player-sdk`). | API-Kontrakt §3.2 |
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
        "name": "@m-trace/player-sdk",
        "version": "0.1.0"
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

`session_id`, `user_agent`, `segment_url` dürfen als **Span-Attribute** verwendet werden (Cardinality-Regel gilt nur für Prometheus-Labels, nicht für Trace-Attribute).

### 2.2 Counter

Der Use Case ruft am Eintritt jedes `RegisterPlaybackEventBatch`-Aufrufs einen frameworkneutralen `Telemetry`-Port auf (Architektur §3.3, §5.3). Der OTel-Adapter mappt diesen Aufruf auf einen Counter:

| OTel-Counter | Typ | Aufrufstelle | Attribute |
|---|---|---|---|
| `mtrace.api.batches.received` | `Int64Counter` | Use Case Step 0 (vor Auth) | `batch.size=<int>` |

**Naming-Translation**: das OTel-→-Prometheus-Mapping ersetzt `.` durch `_`; in Prometheus erscheint der Counter als `mtrace_api_batches_received` (vom OTLP-Exporter automatisch konvertiert). Smoke-Test-Regex `^mtrace_.+` aus `plan-0.1.2.md` §4 deckt sowohl den translated OTel-Counter als auch die direkten Prometheus-Counter aus `adapters/driven/metrics` ab.

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

Zusätzlich zu den vier Pflicht-Countern werden in `0.1.2` die Mindestmetriken aus Lastenheft §7.9 instrumentiert (`mtrace_active_sessions`, `mtrace_api_requests_total`, …).

---

## 3. Cardinality-Regeln

> Bezug: F-95..F-100 (Lastenheft §7.10), F-101..F-105 (MVP-Variante).

### 3.1 Verbotene Prometheus-Labels

Folgende Werte dürfen **nie** als Prometheus-Label erscheinen, weil sie Cardinality-Explosion verursachen:

| Label | Begründung | Bezug |
|---|---|---|
| `session_id` | hochkardinale Pseudonym-IDs; potentiell Millionen Sessions/Tag. | F-96, F-105 |
| `user_agent` | quasi-unbegrenzter Wertebereich. | API-Kontrakt §7 |
| `segment_url` | URL-Variation pro CDN/Player-Adaptation. | API-Kontrakt §7 |
| `client_ip` | DSGVO-Risiko + hohe Cardinality. | API-Kontrakt §7 |
| beliebige `project_id` | bei Multi-Tenant explosiv; nur kontrollierte Allowlist erlaubt. | API-Kontrakt §7 |

### 3.2 Erlaubte Aggregat-Labels

Erlaubt sind Labels mit kontrolliertem, kleinem Wertebereich:

| Label | Wertebereich | Beispiel |
|---|---|---|
| `event_type` | feste Enum aus §1.3 | `rebuffer_started`, `playback_error` |
| `outcome` | feste Enum | `accepted`, `invalid`, `rate_limited`, `dropped` |
| `instance` / `job` | OTel/Prometheus-Standard | `api:8080` |

Prometheus-Series pro Mindest-Counter sollten ≤ einstellige Anzahl sein. RAK-9-Smoke-Test (`plan-0.1.2.md` §4) prüft dies via `count(count by (...) (...))`-PromQL.

### 3.3 Trennung Aggregat vs Per-Session

Per-Session-Daten (Stream-Health, Event-Timeline) gehen **nicht** in Prometheus, sondern in:

- **OTel-Spans** mit Span-Attributen (Cardinality-Limit gilt dort nicht — Spans sind sample-basiert).
- **Event-Store** im `apps/api`-Repository (in `0.1.0` In-Memory; in `0.2.0`+ über OE-3-Folge-ADR auf SQLite/PostgreSQL).
- **Dashboard-Trace-Ansicht** (MVP-14, plan-0.1.1.md §3) liest aus dem Event-Store, nicht aus Prometheus.

Diese Trennung ist die zentrale Architektur-Aussage von F-97 und Lastenheft §7.10.

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
| `batchMaxEvents` | maximale Events pro Batch, ≤ 100. | `25` |
| `batchMaxAgeMs` | maximale Wartezeit, bevor ein nicht-voller Batch geflusht wird. | `1000` |
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

Bevorzugte Reihenfolge (F-128):

1. `sequence_number` (falls gesetzt — Client kontrolliert).
2. `(server_received_at, ingest_sequence)` als Fallback und als Pagination-Sortierschlüssel im API.

Das Dashboard zeigt Events bevorzugt nach `sequence_number` an, wenn vorhanden; die API liefert das Default-Ordering aus §5.2 oben aber konsistent als `(server_received_at, sequence_number, ingest_sequence)`.

### 5.3 Latenzberechnung und Time-Skew

- Latenzen dürfen niemals blind aus reiner Client-Zeit abgeleitet werden (F-129) — Client-Uhren divergieren in der Praxis um Sekunden bis Minuten.
- Bevorzugt: Latenz = `server_received_at - client_time_origin` (skew-tolerant), nicht `server_received_at - client_timestamp`.
- Auffälliger Skew (F-130): liegt `|client_timestamp - server_received_at|` über einem Schwellwert (Default 60 s), markiert das Backend das Event mit einem Span-Attribut `mtrace.time.skew_warning=true` und einem Domain-Flag (Persistenz-Detail in `EventRepository`-Implementierung).

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
