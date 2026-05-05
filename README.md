# m-trace

**OpenTelemetry-native Observability für Live-Media-Streaming.**

m-trace ist ein selbst-gehosteter Observability- und Diagnose-Stack für Live-Media-Workflows.  
Er hilft, Media-Streams von der Ingest-Seite bis zum Player nachzuverfolgen, indem er Player-Telemetrie, Stream-Sessions, Infrastruktursignale, Prometheus-Metriken und ein OpenTelemetry-kompatibles Eventmodell zusammenführt.

> Status: `0.4.0` released — erweiterte Trace-Korrelation: SQLite-Persistenz, `correlation_id`/`trace_id`-Trennung, Dashboard-Session-Timeline ohne Tempo-Pflicht, optionales Tempo-Profil, Aggregat-Metriken-Sichtbarkeit, Cardinality-/Sampling-Doku.

---

## Was ist m-trace?

m-trace richtet sich an Entwickler, Selbsthoster, kleine Streaming-Plattformen, Broadcaster und technische Teams, die verstehen wollen, was in ihrer Streaming-Pipeline passiert — ohne sich von einem proprietären SaaS-Analytics-Silo abhängig zu machen.

Das erste Ziel ist einfach:

```text
MediaMTX + hls.js-Demo-Player + Playback-Events + Dashboard + OpenTelemetry-kompatibles Modell
```

Das langfristige Ziel ist breiter:

```text
Media-Streams von Ingest bis Player nachverfolgen.
```

---

## Warum m-trace?

Kommerzielle Plattformen wie Mux Data, Bitmovin Analytics, NPAW/YOUBORA und Conviva lösen viele klassische QoE- und Analytics-Probleme.  
m-trace fokussiert eine andere Lücke:

- selbstgehostete Streaming-Observability
- OpenTelemetry-natives Modellieren
- Korrelation von Ingest bis Player
- entwicklerfreundliche lokale Demos
- Streaming-Diagnose statt Business-Analytics
- praktisches Tooling für kleine Teams und Labs

Das Projekt versucht nicht, eine vollständige kommerzielle Video-Analytics-Plattform zu ersetzen.  
Es soll ein praxistauglicher Open-Source-Stack für technische Streaming-Diagnose werden.

---

## Kerngedanke

Ein typischer Live-Streaming-Flow sieht so aus:

```text
Encoder / FFmpeg / OBS
        ↓
Ingest
        ↓
MediaMTX
        ↓
HLS
        ↓
hls.js Player
        ↓
m-trace Player SDK
        ↓
m-trace API
        ↓
Dashboard / Metrics / OpenTelemetry
```

m-trace sammelt und normalisiert Signale aus Player und Backend, sodass Stream-Sessions inspiziert, debugged und langfristig mit Infrastruktur-Telemetrie korreliert werden können.

---

## MVP-Scope

Der erste MVP ist bewusst klein gehalten.

### Enthalten seit v0.1.0

- Mono-Repo-Struktur
- Backend-API unter `apps/api`
- MediaMTX-basiertes lokales Streaming-Setup
- FFmpeg-Teststream
- Aufnahme von Playback-Events
- Prometheus-kompatible Aggregat-Metriken
- OpenTelemetry-kompatibles Eventmodell
- In-Memory-Persistenz
- Docker-first lokale Entwicklung

### Enthalten seit v0.1.1

- Dashboard unter `apps/dashboard`
- Demo-Player als Dashboard-Route `/demo`
- Player-SDK unter `packages/player-sdk`
- hls.js-Adapter
- einfache Stream-Session-Ansicht
- einfache Event-Ansicht

### Enthalten seit v0.3.0

- Stream-Analyzer-Bibliothek `@npm9912/stream-analyzer` mit
  HLS-Klassifikator, SSRF-geschütztem URL-Loader und Master-/
  Media-Detail-Parsing
- Diskriminierter `AnalysisResult`-Typ mit Erweiterungspfad für
  DASH/CMAF (`analyzerKind`) und Stabilitätsregel
- Interner Node-Service `@npm9912/analyzer-service` als HTTP-
  Wrapper, in der Compose-Topologie verdrahtet
- API-Endpunkt `POST /api/analyze` (siehe
  `spec/backend-api-contract.md` §3.6) mit Problem-Shape-Fehlern
  und Prometheus-Counter `mtrace_analyze_requests_total`
- CLI `pnpm m-trace check <url-or-file>` mit JSON-stdout und
  definierten Exit-Codes
- `make smoke-analyzer` und `make smoke-cli` als End-to-End-Smokes

### Enthalten seit v0.4.0

- Durable SQLite-Persistenz für Sessions, Playback-Events und
  Ingest-Sequenz statt In-Memory-Store; Cursor sind Restart-stabil
  ([ADR-0002](docs/adr/0002-persistence-store.md))
- `correlation_id` als durable Source-of-Truth für Player-Sessions:
  ein Wert pro Session über alle Batches; `trace_id` ist optionale
  Tempo-Vertiefung pro Batch (`spec/telemetry-model.md` §2.5)
- Manifest-/Segment-/Player-Korrelation über `correlation_id` mit
  URL-Redaction am SDK-Boundary, `session_boundaries[]`-Wrapper für
  Degradationsfälle und `network_signal_absent[]`-Read-Shape
- Dashboard-Session-Timeline-Ansicht (`/sessions/<id>`) mit
  Server-Sent Events plus Polling-Fallback und Backfill-Cursor
  ([ADR-0003](docs/adr/0003-live-updates.md))
- Optionales Tempo-Profil `make dev-tempo` für Trace-Debugging — das
  Dashboard bleibt Tempo-unabhängig, Source-of-Truth ist SQLite
  (RAK-31 Kann-Scope, RAK-32 Pflicht)
- Pflichtcounter sichtbar in Grafana, OTel-Counter
  `mtrace_api_batches_received` ist label-frei wie die vier
  Prometheus-Pflichtcounter; verschärfter Cardinality-Smoke
- Endpoint-spezifische Auth: `POST /api/playback-events` und Session-
  /Event-Reads sind tokenpflichtig; `POST /api/analyze` ist nur bei
  gesetzter `correlation_id`/`session_id` tokenpflichtig
- Cursor-v3 mit Project-Scope, neuer `cursor_invalid_legacy`-Code
  für `0.1.x`/`0.2.x`/`0.3.x`-Cursor

### Nicht im aktuellen 0.3.x/0.4.x-MVP enthalten

- separate Demo-Player-App
- separate Analyzer-API
- produktive Multi-Tenancy
- WebRTC-Monitoring
- SRT-Health-Ansicht
- Tempo als Pflicht-Abhängigkeit
- Mimir oder ClickHouse
- Kubernetes-Produktionsbetrieb
- vollständiger HLS-/DASH-Manifest-Analyzer


---

## Architekturprinzipien

Die aktuelle Architektur ist in [spec/architecture.md](spec/architecture.md) beschrieben.

---

## Eventmodell

Player-Events nutzen ein versioniertes Wire-Format.

Beispiel:

```json
{
  "schema_version": "1.0",
  "events": [
    {
      "event_name": "rebuffer_started",
      "project_id": "demo",
      "session_id": "01J...",
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

Wichtige Konzepte:

- `schema_version`
- `project_id`
- `session_id`
- `client_timestamp`
- `server_received_at`
- `sequence_number`
- SDK-Name und -Version

Das Backend muss Schema-Evolution, Time Skew, Rate Limits und ungültige Event-Batches explizit behandeln.

---

## Metriken

Prometheus wird ausschließlich für Aggregat-Metriken genutzt. Die
drei Backends teilen die Verantwortung wie folgt (kanonische 3-Spalten-
Tabelle: [`spec/telemetry-model.md`](spec/telemetry-model.md) §3.3):

| Backend | Rolle | Cardinality |
|---|---|---|
| **Prometheus** | Aggregat-Metriken (Counter, Rates) | bounded — Forbidden-Liste aus [`spec/telemetry-model.md`](spec/telemetry-model.md) §3.1 release-blockierend |
| **SQLite** (ADR-0002) | Per-Session-Historie inkl. `session_id`, `correlation_id`, `trace_id`, redacted URLs | unbeschränkt |
| **OTel/Tempo** | Per-Request-Trace-Spans (sample-basiert) | nicht im Cardinality-Vertrag |

Beispiele für Prometheus-Counter (alle label-frei):

```text
mtrace_playback_events_total
mtrace_invalid_events_total
mtrace_rate_limited_events_total
mtrace_dropped_events_total
mtrace_active_sessions
mtrace_api_batches_received
```

Hochkardinale Werte wie `session_id`, `correlation_id`, `trace_id`,
`user_agent`, `segment_url`, `client_ip` oder Token-/Credential-Felder
dürfen **nicht** als Prometheus-Labels verwendet werden — die
vollständige Forbidden-Liste plus Suffix-Regeln (`*_url`, `*_uri`,
`*_token`, `*_secret`) steht in
[`spec/telemetry-model.md`](spec/telemetry-model.md) §3.1.

Per-Session-Debugging läuft über die durable SQLite-Persistenz und
optional über Tempo-Spans (`make dev-tempo`) — niemals über Prometheus.

---

## OpenTelemetry-Strategie

m-trace ist von Beginn an OpenTelemetry-nativ.

Das bedeutet:

- bestehende OTel-Semantic-Conventions nutzen, wo möglich
- media-spezifische Attribute nur dort definieren, wo nötig
- vendor-spezifische Telemetrieformate vermeiden
- Session-Daten trace-kompatibel halten
- Prometheus auf Aggregate beschränken
- spätere Korrelation über Ingest, Origin und Player vorbereiten

Ein zukünftiger Player-Session-Trace könnte so aussehen:

```text
Player Session Trace
├── manifest_request
├── segment_request
├── startup_time
├── bitrate_switch
├── rebuffer_event
└── playback_error
```

---

## Lokale Entwicklung

Das lokale Core-Lab startet Backend-API, Dashboard, MediaMTX und einen FFmpeg-Teststream:

```bash
git clone https://github.com/pt9912/m-trace.git
cd m-trace
make dev
```

Smoke-Test:

```bash
make smoke
```

SDK- und Demo-Dokumentation:

- [spec/player-sdk.md](spec/player-sdk.md) beschreibt Installation, Public API, Transport, Performance-Budget und Browser-Build.
- [docs/user/demo-integration.md](docs/user/demo-integration.md) beschreibt die Dashboard-Route `/demo` als lokale hls.js-/Player-SDK-Integration.
- [spec/browser-support.md](spec/browser-support.md) dokumentiert die Browser-Support-Matrix.

Lokaler SDK-/Demo-Pfad:

```bash
pnpm --filter @npm9912/player-sdk run pack:smoke
make dev
# dann http://localhost:5173/demo?session_id=readme-demo&autostart=1 öffnen
```

Optionaler Observability-Stack mit Prometheus, Grafana und OTel-Collector:

```bash
make dev-observability
make smoke-observability
```

Browser-End-to-End-Test im Playwright-Container:

```bash
make browser-e2e
```

- Dashboard: <http://localhost:5173>
- API: <http://localhost:8080/api/health>
- HLS-Teststream: <http://localhost:8888/teststream/index.m3u8>
- Prometheus: <http://localhost:9090> (`make dev-observability`)
- Grafana: <http://localhost:3000> (`admin`/`admin`, `make dev-observability`)
- OTel Collector: OTLP `localhost:4317`/`4318`, Health <http://localhost:13133>

Details stehen in [docs/user/local-development.md](docs/user/local-development.md).

## Roadmap

### v0.1.0 — OTel-native lokale Demo

- MediaMTX-Lokal-Setup
- hls.js-Demo-Route
- Aufnahme von Player-Events
- einfache Session-Ansicht
- einfache Event-Ansicht
- Prometheus-Aggregat-Metriken
- OpenTelemetry-kompatibles Eventmodell

### v0.2.0 — Publizierbares Player-SDK

- npm-Paket
- stabile Public API
- hls.js-Adapter-Tests
- Event-Schema-Kompatibilitätstests
- Batching und Sampling
- dokumentierter Browser-Support

### v0.3.0 — Stream-Analyzer (✅ veröffentlicht)

- HLS-Manifest-Parsing für Master und Media Playlists
- Segment-Dauer-Checks und Target-Duration-Verletzung
- API-Anbindung über internen analyzer-service
- CLI-Grundlage `pnpm m-trace check <url-or-file>`

### v0.4.0 — Erweiterte Trace-Korrelation (✅ veröffentlicht)

- durable SQLite-Persistenz mit `make wipe` als verbindlichem
  Reset-Pfad (ADR-0002)
- Player-Session-Korrelation über `correlation_id` (durable, Tempo-
  unabhängig); `trace_id` ist optionale Per-Batch-Vertiefung
- Manifest-/Segment-/Player-Trace mit URL-Redaction am SDK-Boundary
- Dashboard-Session-Timeline (`/sessions/<id>`) mit SSE und
  Polling-Fallback (ADR-0003)
- optionales Tempo-Profil `make dev-tempo` (RAK-31, Kann-Scope;
  Dashboard-Timeline bleibt Tempo-unabhängig — RAK-32)
- Cardinality-/Sampling-Doku: Pflichtcounter und
  `mtrace_api_batches_received` sind label-frei; Sampling-Grenze für
  `sampleRate < 1` dokumentiert

### v0.5.0 — Multi-Protokoll-Lab

- DASH-Beispiel
- SRS-Beispiel
- erweiterte MediaMTX-Beispiele

### v0.6.0 — SRT-Health-Ansicht

- SRT-Metriken
- RTT, Packet Loss, Retransmissions
- Link-Health-Dashboard
- SRT-Troubleshooting-Doku

---

## Browser-Support

Der MVP-Browser-Support ist bewusst eng gefasst.

| Umgebung                         | MVP-Status                |
| -------------------------------- | ------------------------- |
| Chrome Desktop, aktuelle Stable  | unterstützt               |
| Firefox Desktop, aktuelle Stable | unterstützt               |
| Safari Desktop, aktuelle Stable  | eingeschränkt             |
| Chromium-basierte Browser        | best effort               |
| iOS Safari                       | im MVP nicht erforderlich |
| Android Chrome                   | im MVP nicht erforderlich |
| Smart-TV-Browser                 | nicht im Scope            |
| Embedded WebViews                | nicht im Scope            |

Der MVP-Integrationspfad ist hls.js.  
Native Safari-HLS-Introspektion ist kein Ziel von v0.1.0.

---

## Sicherheit und Datenschutz

m-trace soll für selbstgehostete Umgebungen standardmäßig sicher sein.

MVP-Prinzipien:

- keine Secrets im Repository
- keine cookie-basierte Telemetrie-Annahme
- SDK-Requests nutzen standardmäßig `credentials: "omit"`
- erlaubte Origins werden pro Projekt konfiguriert
- Project-Tokens gelten als niedrig-kritische Browser-Tokens
- Rate Limits sind verpflichtend
- IP-Adressen sollen nicht unnötig gespeichert werden
- User-Agent-Daten sollen reduzierbar oder anonymisierbar sein
- GDPR-konformer Betrieb muss möglich sein

---

## Was m-trace nicht ist

m-trace ist nicht:

- ein Ersatz für kommerzielle QoE-Analytics
- ein Werbe-Analytics-System
- eine DRM-Analytics-Plattform
- ein CDN-Optimizer
- eine vollständige Multi-Tenant-SaaS-Plattform
- ein Ersatz für MediaMTX, FFmpeg, Grafana oder Prometheus

m-trace ist ein technisches Observability- und Diagnose-Projekt für Media-Streaming-Workflows.

---

## Aktueller Stand

Das Projekt steht bei `0.4.0` (released, Tag `v0.4.0`): SQLite-
Persistenz, Trace-Korrelation, Manifest-/Segment-Korrelation, Dashboard-
Session-Timeline mit SSE, optionales Tempo-Profil, Aggregat-Metriken-
Sichtbarkeit und Cardinality-/Sampling-Doku sind auf `main` integriert
und durch GitHub-Actions-`build` grün verifiziert. Plan-Datei:
[`docs/planning/done/plan-0.4.0.md`](docs/planning/done/plan-0.4.0.md).
Nächste Phase: `0.5.0` (Multi-Protokoll-Lab, RAK-36..RAK-40) — Scope-Cut
steht aus.

Leitende Dokumente:

- [spec/lastenheft.md](spec/lastenheft.md) — Anforderungen (verbindlich, 1.1.8)
- [docs/planning/in-progress/roadmap.md](docs/planning/in-progress/roadmap.md) — Status, Folge-ADRs, offene Entscheidungen
- [docs/adr/0001-backend-stack.md](docs/adr/0001-backend-stack.md) — Backend-Entscheidung (Accepted: Go)
- [docs/user/releasing.md](docs/user/releasing.md) — Release-Prozess
- [docs/user/quality.md](docs/user/quality.md) — Qualitätsrichtlinien

Nächste Schritte stehen in [docs/planning/in-progress/roadmap.md](docs/planning/in-progress/roadmap.md) §2.

---

## Lizenz

[MIT License](LICENSE).

---

## Name

`m-trace` steht für:

```text
Media Trace
```

Das Projektziel ist simpel:

```text
Media-Streams von Ingest bis Player nachverfolgen.
```
