# Telemetry-Model — m-trace

> **Status**: Skeleton — Inhalts-Sections sind als Platzhalter angelegt und werden im Zuge der `0.1.0`-Phase befüllt.  
> **Bezug**: [Lastenheft `1.1.3`](./lastenheft.md) §7.10 (Cardinality), §7.11 (Telemetry Ingest, Event-Schema, SDK-Budget); [Roadmap](./roadmap.md) §2 Schritt 6; [Plan `0.1.0`](./plan-0.1.0.md) §3.5; [API-Kontrakt](./spike/backend-api-contract.md); [Architektur](./architecture.md) §5.

## 0. Zweck

Beschreibt das **Datenmodell** der Telemetrie — Wire-Format, Schema, Cardinality-Regeln, Time-Stempel-Konventionen, Backpressure-Politik. Implementierungs-/Setup-Aspekte (strukturierte Logs, Health-Endpoint, Prometheus- und Grafana-Konfiguration) gehören in [`plan-0.1.2.md`](./plan-0.1.2.md), nicht hierher.

## 1. Wire-Format Player-Events

> **Status: TODO** — Bezug F-106..F-115 (Lastenheft §7.11.1–§7.11.3).

Zu spezifizieren:

- Pflichtfelder pro Event (`event_name`, `project_id`, `session_id`, `client_timestamp`, `sdk.name`, `sdk.version`, …).
- Optionale Felder (`sequence_number`, `client_time_origin`).
- Schema-Version-Feld (Lastenheft §7.11.3 / API-Kontrakt §3).
- Beispiel-Payloads (Happy-Path, leere Sub-Felder).
- **NF-37 CSP-Beispiele für SDK-`connect-src`**: für Drittanbieter, die das Player-SDK in eigene Seiten einbinden, ein Mustertext (z. B. `connect-src 'self' https://collector.example.com`); Hinweis auf Origin-Allowlist im Backend (F-108).

## 2. OTel-Modell

> **Status: TODO** — Bezug F-91, F-92.

Zu spezifizieren:

- Naming-Konvention für Spans (`http.handler ...`, `application.RegisterPlaybackEventBatch`, …).
- Naming-Konvention für Counter (`mtrace.api.batches.received`, …) und Mapping zu den Prometheus-Mindestmetriken aus Lastenheft §7.9.
- Resource-Attribute (`service.name`, `service.version`, `mtrace.component`).
- Span-Attribute (`batch.size`, `batch.outcome`, `event.session_id` als Span-Attribut zulässig — High-Cardinality-Verbot gilt nur für Prometheus-Labels).

## 3. Cardinality-Regeln

> **Status: TODO** — Bezug F-95..F-100 (Lastenheft §7.10) und F-101..F-105 (MVP-Variante).

Zu spezifizieren:

- Verbotene Prometheus-Labels (`session_id`, `user_agent`, `segment_url`, `client_ip`, beliebige `project_id`).
- Erlaubte Aggregat-Labels (z. B. `event_type` mit kontrollierter Wertemenge).
- Trennung Aggregat (Prometheus) vs. Per-Session (Trace/Event-Store).
- Spike-Spec-Regel-Verweis (API-Kontrakt §7).

## 4. Backpressure und Limits

> **Status: TODO** — Bezug F-118..F-123.

Zu spezifizieren:

- Max-Batch-Größe (100 Events laut API-Kontrakt §3).
- Max-Body-Größe (256 KB laut API-Kontrakt §5).
- Rate-Limit-Modell (Token-Bucket pro `project_id`/`origin`/`client_ip`, F-110).
- Drop-Politik bei interner Backpressure (`mtrace_dropped_events_total` ausschließlich für Backpressure-Drops, nicht für synchron fehlgeschlagenes `Append`).

## 5. Time-Stempel und Sequenz-Ordering

> **Status: TODO** — Bezug F-124..F-130.

Zu spezifizieren:

- Pflicht-Felder: `client_timestamp`, `server_received_at`.
- Optionale Felder: `client_time_origin`, `sequence_number`.
- Ordering innerhalb einer Session über `sequence_number` (F-128).
- Latenzberechnung (F-129): nicht blind aus Client-Zeit; Time-Skew-Korrektur über `server_received_at`.
- Time-Skew-Markierung (F-130): wenn `client_timestamp` und `server_received_at` über einer Schwelle (z. B. 60 s) divergieren, markiert das Backend das Event.

## 6. Schema-Versionierung

> **Status: TODO** — Bezug F-116, F-117.

Zu spezifizieren:

- Schema-Versionen-Verwaltung (SemVer-ähnlich oder integer).
- Major-/Minor-Inkompatibilitäten.
- Toleranz-Regel: entfernte Felder werden über mindestens eine Minor-Version toleriert (F-116).
- Erweiterbarkeit ohne Schema-Bump (zusätzliche optionale Felder).
