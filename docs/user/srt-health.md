# SRT-Health-View

`0.6.0` ergänzt eine **lokale SRT-Verbindungs-Health-Sicht** für das
m-trace-Lab: ein Collector liest periodisch die MediaMTX-Control-API,
der Server bewertet Health-Zustände, das Dashboard zeigt sie.

> Bezug: [`spec/lastenheft.md`](../../spec/lastenheft.md) §4.3, §13.8
> RAK-41..RAK-46;
> [`docs/planning/done/plan-0.6.0.md`](../planning/done/plan-0.6.0.md);
> [`spec/telemetry-model.md`](../../spec/telemetry-model.md) §7;
> [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md)
> §7a/§10.6;
> [`spec/architecture.md`](../../spec/architecture.md) §3.4/§5.4.

## 1. Lieferumfang `0.6.0`

- **Lab-Smoke** [`make smoke-srt-health`](../../scripts/smoke-srt-health.sh)
  startet den `mtrace-srt`-Stack, prüft HLS plus MediaMTX-API
  `/v3/srtconns/list` und vier RAK-43-Pflichtwerte
  ([`examples/srt/README.md`](../../examples/srt/README.md)).
- **CGO-freier Adapter** `apps/api/adapters/driven/srt/mediamtxclient`
  liest die API über HTTP, `apps/api` bleibt `distroless-static`.
- **Durabler Health-Store** in SQLite (`srt_health_samples`-Tabelle,
  spec §10.6) plus exponentielles Polling-Backoff bei Source-Fehlern.
- **Read-Endpoints** `GET /api/srt/health` und
  `GET /api/srt/health/{stream_id}` (spec §7a).
- **Prometheus-Aggregate** mit bounded Labels:
  `mtrace_srt_health_samples_total{health_state}`,
  `mtrace_srt_health_collector_runs_total{source_status}`,
  `mtrace_srt_health_collector_errors_total{source_error_code}`.
- **OTel-Span** `mtrace.srt.health.collect` pro Sample mit
  `mtrace.srt.*`-Attributen.
- **Dashboard-Route** `/srt-health` mit Tabelle pro Stream und
  Detail-Ansicht mit Mini-Timeline der letzten 50 Samples.

Was `0.6.0` **nicht** liefert, steht in §11 (deferred Signale) und in
[`plan-0.6.0.md`](../planning/done/plan-0.6.0.md) §0.1.

## 2. Quickstart

### 2.1 Lab starten und Health-Smoke fahren

```bash
make smoke-srt-health
```

Das Target:

1. fährt das `examples/srt/`-Compose hoch (MediaMTX + FFmpeg-Publisher),
2. wartet, bis HLS auf `localhost:8889/srt-test/index.m3u8` antwortet,
3. probt `http://localhost:9998/v3/srtconns/list` und prüft `msRTT`,
   `packetsReceivedLoss`, `packetsReceivedRetrans`, `mbpsLinkCapacity > 0`,
4. räumt nach Abschluss auf.

### 2.2 Collector am API-Prozess aktivieren

Per Default ist der Collector deaktiviert. Setzen:

```bash
export MTRACE_SRT_SOURCE_URL="http://localhost:9998"
export MTRACE_SRT_SOURCE_USER="any"
export MTRACE_SRT_SOURCE_PASS=""
export MTRACE_SRT_PROJECT_ID="demo"        # default
export MTRACE_SRT_POLL_INTERVAL_SECONDS=5  # default
```

Im Compose-Lab ist die SQLite-Persistenz Pflicht — der Collector
schreibt nicht in In-Memory-DB.

### 2.3 Read-Endpoints

```bash
# Liste pro Stream
curl -H "X-MTrace-Token: demo-token" \
  http://localhost:8080/api/srt/health

# Detail mit Verlauf
curl -H "X-MTrace-Token: demo-token" \
  "http://localhost:8080/api/srt/health/srt-test?samples_limit=50"
```

Wire-Format-Beispiel:
[`spec/contract-fixtures/api/srt-health-detail.json`](../../spec/contract-fixtures/api/srt-health-detail.json).

### 2.4 Dashboard

Nach `make dev` die Dashboard-Sidebar öffnen → **„SRT health"**.

## 3. Datenfluss

```text
mediamtx (SRT :8890/udp + Control-API :9997)
  │  GET /v3/srtconns/list  (HTTP, Basic-Auth)
  ▼
adapters/driven/srt/mediamtxclient   ← parst Wire-Format gegen Fixture
  │ []domain.SrtConnectionSample
  ▼
hexagon/application/SrtHealthCollector  ← Single-Shot oder Polling-Loop
  │  - berechnet health_state aus Schwellen
  │  - bytesReceived-Δ als Source-Sequence-Surrogat
  │
  ├──► SrtHealthRepository   (sqlite, srt_health_samples)
  ├──► MetricsPublisher      (mtrace_srt_health_*)
  └──► Telemetry             (Span "mtrace.srt.health.collect")

GET /api/srt/health[/{stream_id}]
  │  application/SrtHealthQueryService.LatestByStream/HistoryByStream
  ▼
adapters/driving/http/SrtHealthListHandler / SrtHealthGetHandler
  │ JSON-Wire (metrics / derived / freshness — spec §7a.2)
  ▼
apps/dashboard /srt-health
```

Spec-Referenz: [`architecture.md`](../../spec/architecture.md) §5.4.

## 4. Metriken (RAK-43)

| Wert | Wire-Feld (`metrics.*`) | Einheit | Quelle (MediaMTX-Mapping) |
|------|-------------------------|---------|----------------------------|
| RTT | `rtt_ms` | Millisekunden | `msRTT` (Snapshot) |
| Packet Loss (Counter) | `packet_loss_total` | Pakete (kumulativ) | `packetsReceivedLoss` |
| Packet Loss (Rate, optional) | `packet_loss_rate` | 0..1 | `packetsReceivedLossRate` |
| Retransmissions | `retransmissions_total` | Pakete (kumulativ) | `packetsReceivedRetrans` |
| Verfügbare Bandbreite | `available_bandwidth_bps` | bit/s | `mbpsLinkCapacity × 1_000_000` |
| Tatsächlicher Durchsatz (optional) | `throughput_bps` | bit/s | `mbpsReceiveRate × 1_000_000` |
| Erwartete Bandbreite | `required_bandwidth_bps` | bit/s | aus Lab-/Stream-Konfig |
| Sample-Window | `sample_window_ms` | Millisekunden | optional |

### 4.1 Counter vs. Rate

- **Counter** (kumulativ ab Verbindungsstart, Reset bei
  Reconnect/`connection_id`-Wechsel): `packet_loss_total`,
  `retransmissions_total`. Adapter speichert den absoluten Wert;
  Dashboard kann die Intervallrate aus Δ Counter / Δ
  `source_sequence` (= `bytesReceived`) ableiten.
- **Snapshot** (Momentaufnahme, jeder Poll überschreibt): `rtt_ms`,
  `available_bandwidth_bps`, `throughput_bps`.

### 4.2 Bandbreite richtig lesen

Der `available_bandwidth_bps`-Wert ist die **SRT-eigene
Linkkapazitäts-Schätzung**, nicht der Stream-Durchsatz. In
localhost-Loopback liegt er typisch bei mehreren Gbit/s — das ist
**kein** Beweis für einen gesunden Pfad in einem realen Netzwerk.

Verlässlich ist der Wert nur, wenn `required_bandwidth_bps` gesetzt
ist (Stream-Konfiguration kennt den Bedarf):

- `available ≥ required × 1.5` → grünes Headroom-Verhältnis.
- `required ≤ available < required × 1.5` → `degraded`.
- `available < required` → `critical`.
- `required` nicht gesetzt → Bandbreite wird nur **angezeigt**, nicht
  bewertet (Dashboard zeigt „Required bandwidth: unset").

Spec-Referenz: [`telemetry-model.md`](../../spec/telemetry-model.md)
§7.4.

## 5. Health-Zustände

| Zustand | Bedingung (Default-Schwellen, `application.DefaultThresholds`) |
|---------|----------------------------------------------------------------|
| `healthy` | RTT < 100 ms, Loss-Rate < 1 %, `available ≥ required × 1.5` (oder kein `required`). |
| `degraded` | RTT 100–250 ms ODER Loss-Rate 1–5 % ODER `required ≤ available < required × 1.5`. |
| `critical` | RTT ≥ 250 ms ODER Loss-Rate > 5 % ODER `available < required`. |
| `unknown` | `source_status ≠ ok` (siehe §7) ODER Stale-Erkennung schlägt an (siehe §6) ODER Pflichtwerte fehlen. |

Schwellen sind im Server-Code als Konstanten gepinnt; das Dashboard
bewertet **nicht** selbst — UI-Pillen reagieren nur auf den
Server-`health_state`.

## 6. Freshness und Stale-Erkennung

MediaMTX liefert keinen Source-Sample-Timestamp. Der Adapter setzt
deshalb:

- `collected_at` = Zeitpunkt des HTTP-Polls,
- `source_sequence` = `bytesReceived` aus dem Sample (monoton steigend
  bei aktivem Stream).

Die Stale-Bewertung des Servers (`SourceStatusStale`) greift, wenn
`source_sequence` über `StaleAfterMillis` (Default 15 s) gleich
bleibt **und** `connection_state` weiterhin `connected` meldet.

Das Dashboard hat eine zusätzliche **client-seitige Stale-Sicht**:
`sample_age_ms > stale_after_ms` zum Lesezeitpunkt → Stale-Pill (gelb)
plus Text-Suffix `(stale)`. Das fängt den Fall ab, dass der Collector
nicht mehr läuft oder das Polling pausiert ist.

## 7. Source-Status-Tabelle

Die `source_status`/`source_error_code`-Klassifikation bestimmt, wann
Health auf `unknown` fällt. Stabile Codes aus
[`telemetry-model.md`](../../spec/telemetry-model.md) §7.5:

| `source_status` | `source_error_code` | Auslöser |
|---|---|---|
| `ok` | `none` | Quelle erreichbar, alle Pflichtfelder gesetzt. |
| `no_active_connection` | `no_active_connection` | Quelle erreichbar, aber `items[]` enthält keinen erwarteten Stream. |
| `partial` | `partial_sample` | Quelle erreichbar, Pflichtfeld fehlt oder ist non-numeric. |
| `stale` | `stale_sample` | `source_sequence` über N Polls eingefroren trotz `state: publish`. |
| `unavailable` | `source_unavailable` | HTTP 4xx/5xx, Connection refused, Timeout. |
| `unavailable` | `parse_error` | HTTP 200, Body kein gültiges JSON. |

Prometheus zählt jede Klasse über
`mtrace_srt_health_collector_errors_total{source_error_code}`.

## 8. Fehlerbilder

Operator-Faustregeln aus dem Lab plus Lastenheft §4.3:

| Symptom | Wertebild |
|---------|-----------|
| **Hohe RTT** | `rtt_ms` steigt deutlich (ab 100 ms `degraded`, ab 250 ms `critical`). Loss/Retrans bleiben bei kurzem Spike oft niedrig. |
| **Paketverlust** | `packet_loss_total` steigt; Retransmissions folgen meist binnen weniger Sekunden, weil SRT verlorene Pakete nachsendet. |
| **Retransmission-Spirale** | `retransmissions_total` steigt dauerhaft; `available_bandwidth_bps` schwankt oder sinkt; `throughput_bps` < Stream-Soll. Der Pfad ist kapazitätslimitiert. |
| **Bandbreitenengpass** | `available_bandwidth_bps < required_bandwidth_bps` → `critical`. **Nicht** allein anhand `throughput_bps` entscheiden — Durchsatz zeigt nur, was tatsächlich sendet. |
| **Keine Verbindung** | `items[]` leer → `source_status: no_active_connection` und `health_state: unknown`. Publisher prüfen (z. B. `docker logs mtrace-srt-srt-publisher-1`). |
| **Quelle stale** | Stream läuft, aber `source_sequence` ändert sich nicht → `source_status: stale`. Typisch wenn der Publisher zwar verbunden ist, aber keine Pakete schickt (Application-Bug, eingefrorene Encoder-Pipeline). |
| **API blockiert** | `mtrace_srt_health_collector_errors_total{source_error_code="source_unavailable"}` steigt. MediaMTX-`authInternalUsers` prüfen — `examples/srt/mediamtx.yml` ist der Lab-Default. |
| **Schema-Drift** | `parse_error` häuft sich. MediaMTX-Major-Version geändert? Fixture in `spec/contract-fixtures/srt/mediamtx-srtconns-list.json` mit dem realen Body abgleichen. |

## 9. Cardinality- und Datenschutzvertrag

- **Per-Verbindung-Identifier** (`stream_id`, `connection_id`,
  MediaMTX-`id`/`path`/`remoteAddr`/`state`) gehen ausschließlich in
  SQLite und OTel-Spans, **nie** in Prometheus-Labels.
- Erlaubte bounded Labels in Prometheus: `health_state`,
  `source_status`, `source_error_code` (spec §3.2).
- MediaMTX-eigene Prometheus-Targets werden **nicht** vom Projekt-
  Prometheus gescraped. `make smoke-observability` prüft das.
- IPs aus `remoteAddr` sind in Lab-Setups Docker-intern (`172.17.x.y`).
  In produktiven Setups wären das öffentliche IPs — Persistenz folgt
  dann dem GDPR-Pfad analog `EventRepository`.
- MediaMTX-Auth-Credentials (`MTRACE_SRT_SOURCE_PASS`) gehören in ENV
  / Geheimnis-Store, nie in Logs oder Span-Attributen.

## 10. Operator-Quickref

| Aufgabe | Befehl |
|---------|--------|
| Lab + Smoke | `make smoke-srt-health` |
| Nur Lab starten | `docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build` |
| MediaMTX-API roh prüfen | `curl -sS http://localhost:9998/v3/srtconns/list` |
| Health-Liste über API | `curl -H "X-MTrace-Token: demo-token" http://localhost:8080/api/srt/health` |
| Health-Detail | `curl -H "X-MTrace-Token: demo-token" "http://localhost:8080/api/srt/health/srt-test?samples_limit=50"` |
| Dashboard | <http://localhost:5173/srt-health> |
| Prometheus-Aggregate | `curl -sS http://localhost:9090/api/v1/query?query=mtrace_srt_health_samples_total` |

## 11. Deferred / Out-of-Scope für `0.6.0`

Lastenheft §4.3 listet weitere SRT-Signale, die `0.6.0` bewusst nicht
auswertet:

| Signal | Begründung |
|--------|------------|
| Send-/Receive-Buffer (`msReceiveBuf`, `bytesReceiveBuf`, `packetsReceiveBuf`) | Aus MediaMTX-API verfügbar; aktuell nicht im Domain-Modell. Spec [`telemetry-model.md`](../../spec/telemetry-model.md) §7.2 dokumentiert das Mapping als Folge-Item. |
| Detaillierte Verbindungsstabilität | Kann aus Health-Verlauf abgeleitet werden, aber kein eigener Pflichtwert in `0.6.0`. |
| Link Health-Score (separat zu `health_state`) | `health_state` deckt die operative Sicht; ein detaillierter Score (z. B. 0–100) ist Folge-Scope. |
| Failover-Zustände | `0.6.0` hat kein Multi-Path-/Failover-Lab; deferred. |
| Cursor-Pagination im History-Read-Pfad | spec §7a.3 dokumentiert das Wire-Format; SQLite-Adapter ist als ErrNotImplemented gestubbed (siehe [`plan-0.6.0.md`](../planning/done/plan-0.6.0.md) §4 Sub-3.3). |

Operator-Hinweis: wenn ein deferred Signal produktiv gebraucht wird,
ist der Pfad ein Folge-Plan (nicht ein lokaler Hotfix), weil die
Cardinality- und Spec-Verträge angepasst werden müssen.

## 12. Querverweise

- [`spec/telemetry-model.md`](../../spec/telemetry-model.md) §7 —
  Datenmodell, Health-Bewertung, Source-Status, Cardinality-Vertrag.
- [`spec/backend-api-contract.md`](../../spec/backend-api-contract.md)
  §7a (Read-Vertrag) und §10.6 (Persistenz).
- [`spec/architecture.md`](../../spec/architecture.md) §3.4
  (Adapter-Tabelle) und §5.4 (Datenfluss).
- [`examples/srt/README.md`](../../examples/srt/README.md) — Lab-Setup,
  `make smoke-srt-health`-Operator-Sicht.
- [`docs/user/local-development.md`](./local-development.md) §2.7 —
  Multi-Protocol-Lab-Quickref und Port-Schnitt.
- [`docs/user/releasing.md`](./releasing.md) — Release-Smokes.
