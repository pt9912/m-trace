# m-trace

**OpenTelemetry-native Observability für Live-Media-Streaming.**

m-trace ist ein selbst-gehosteter Observability- und Diagnose-Stack für Live-Media-Workflows.  
Er hilft, Media-Streams von der Ingest-Seite bis zum Player nachzuverfolgen, indem er Player-Telemetrie, Stream-Sessions, Infrastruktursignale, Prometheus-Metriken und ein OpenTelemetry-kompatibles Eventmodell zusammenführt.

> Status: `0.3.0` — HLS-Stream-Analyzer mit Library, interner HTTP-Service, API-Endpunkt `POST /api/analyze` und CLI `m-trace check`.

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

### Nicht im aktuellen 0.3.x-MVP enthalten

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

Prometheus wird ausschließlich für Aggregat-Metriken genutzt.

Beispiele:

```text
mtrace_playback_events_total
mtrace_invalid_events_total
mtrace_rate_limited_events_total
mtrace_dropped_events_total
mtrace_active_sessions
```

Hochkardinale Werte wie `session_id`, `user_agent` oder `segment_url` dürfen nicht als Prometheus-Labels verwendet werden.

Per-Session-Debugging soll als Trace modelliert oder in einem geeigneten Event-/Session-Store abgelegt werden.

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

### v0.4.0 — Erweiterte Trace-Korrelation

- Player-Session-Traces
- optionale Tempo-Integration
- Session-Timeline-Ansicht
- Sampling-Strategie

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

Das Projekt steht bei `0.3.0`: Lastenheft `1.1.7` verbindlich, Player-SDK-Paketierung, Dashboard, Observability-Profil, Demo-Integration und der HLS-Stream-Analyzer (Library, analyzer-service, API-Endpunkt, CLI) sind auf `main` integriert.

Leitende Dokumente:

- [spec/lastenheft.md](spec/lastenheft.md) — Anforderungen (verbindlich, 1.1.7)
- [docs/planning/roadmap.md](docs/planning/roadmap.md) — Status, Folge-ADRs, offene Entscheidungen
- [docs/adr/0001-backend-stack.md](docs/adr/0001-backend-stack.md) — Backend-Entscheidung (Accepted: Go)
- [docs/user/releasing.md](docs/user/releasing.md) — Release-Prozess
- [docs/user/quality.md](docs/user/quality.md) — Qualitätsrichtlinien

Nächste Schritte stehen in [docs/planning/roadmap.md](docs/planning/roadmap.md) §2.

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
