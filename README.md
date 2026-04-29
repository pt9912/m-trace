# m-trace

**OpenTelemetry-native Observability für Live-Media-Streaming.**

m-trace ist ein selbst-gehosteter Observability- und Diagnose-Stack für Live-Media-Workflows.  
Er hilft, Media-Streams von der Ingest-Seite bis zum Player nachzuverfolgen, indem er Player-Telemetrie, Stream-Sessions, Infrastruktursignale, Prometheus-Metriken und ein OpenTelemetry-kompatibles Eventmodell zusammenführt.

> Status: Pre-MVP `0.1.0` — Backend-Skelett auf `main`, Lastenheft `1.1.2` verbindlich.

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

### Enthalten in v0.1.0

- Mono-Repo-Struktur
- Backend-API unter `apps/api`
- Dashboard unter `apps/dashboard`
- Demo-Player als Dashboard-Route `/demo`
- Player-SDK unter `packages/player-sdk`
- hls.js-Adapter
- MediaMTX-basiertes lokales Streaming-Setup
- FFmpeg-Teststream
- Aufnahme von Playback-Events
- einfache Stream-Session-Ansicht
- einfache Event-Ansicht
- Prometheus-kompatible Aggregat-Metriken
- OpenTelemetry-kompatibles Eventmodell
- In-Memory- oder SQLite-Persistenz
- Docker-first lokale Entwicklung

### Nicht in v0.1.0 enthalten

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

## Geplante Repository-Struktur

```text
m-trace/
├── apps/
│   ├── api/
│   └── dashboard/
├── packages/
│   ├── player-sdk/
│   ├── stream-analyzer/
│   ├── shared-types/
│   └── config/
├── services/
│   ├── stream-generator/
│   ├── otel-collector/
│   └── media-server/
├── examples/
│   ├── mediamtx/
│   ├── hls/
│   ├── dash/
│   ├── srt/
│   └── webrtc/
├── observability/
│   ├── prometheus/
│   ├── grafana/
│   ├── tempo/
│   └── otel/
├── docs/
│   ├── adr/
│   └── spike/
├── scripts/
├── docker-compose.yml
├── Makefile
├── README.md
└── CHANGELOG.md
```

Nicht alle Verzeichnisse gehören zum ersten MVP.  
Einige sind Platzhalter für die Roadmap.

---

## Architekturprinzipien

m-trace nutzt pragmatische Architekturgrenzen.

### Backend

Backend-Stack ist **Go 1.22** (Standard-Library `net/http`,
`prometheus/client_golang`, `go.opentelemetry.io/otel`, `log/slog`,
Distroless-Runtime). Entscheidung in `docs/adr/0001-backend-stack.md`,
Spec im Lastenheft §10.1.

Workflow ist Docker-only: `make {test,lint,build,run}` in `apps/api/`.
Lokales Go ist nicht erforderlich.

Hexagon-Layout in `apps/api/`:

```text
apps/api/
├── cmd/api/main.go
├── hexagon/
│   ├── domain/
│   ├── port/{driving,driven}/
│   └── application/
├── adapters/
│   ├── driving/http/
│   └── driven/{auth,metrics,persistence,ratelimit,telemetry}/
└── Dockerfile
```

Abhängigkeitsrichtung:

```text
adapters → hexagon
```

Die Domain darf nicht von HTTP-, Datenbank-, Framework-, Docker- oder OpenTelemetry-Implementierungsdetails abhängen.

### Player-SDK

Das Player-SDK ist im MVP bewusst nicht voll hexagonal.

```text
packages/player-sdk/src/
├── core/
├── adapters/
│   └── hlsjs/
├── transport/
├── types/
└── index.ts
```

Die erste unterstützte Player-Integration ist:

```text
hls.js
```

Weitere Integrationen sind spätere Arbeit:

- dash.js
- Shaka Player
- Video.js
- natives Safari-HLS
- WebRTC `getStats`

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
        "name": "@m-trace/player-sdk",
        "version": "0.1.0"
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

## Lokales Entwicklungsziel

Die geplante Developer Experience:

```bash
git clone https://github.com/pt9912/m-trace.git
cd m-trace
make dev
```

Erwartete lokale Dienste:

| Dienst           | Zweck                            |
| ---------------- | -------------------------------- |
| API              | Event-Annahme und Session-API    |
| Dashboard        | Web-UI und `/demo`-Player-Route  |
| MediaMTX         | lokaler Media-Server             |
| FFmpeg-Generator | Teststream                       |
| Prometheus       | Aggregat-Metriken                |
| Grafana          | optionale Dashboards             |
| OTel Collector   | optionale Telemetrie-Pipeline    |

Dieses Setup ist noch nicht umgesetzt.  
Es ist das Ziel des ersten MVP.

---

## Backend-Technologie-Spike (abgeschlossen)

Die Backend-Technologie wurde durch zwei lauffähige Mini-Prototypen
(Go, Micronaut) im identischen Muss-Scope entschieden. Sieger ist Go.

Dokumentation:

- `docs/adr/0001-backend-stack.md` — Entscheidung (Status: Accepted)
- `docs/spike/backend-stack-results.md` — Spike-Protokoll
- `docs/spike/backend-api-contract.md` — API-Kontrakt (frozen)
- `docs/spike/0001-backend-stack.md` — Spike-Spezifikation
- `docs/plan-spike.md` — Implementierungsplan

Sieger-Branch `spike/go-api` ist auf `main` als `apps/api` integriert
(Modulpfad `github.com/pt9912/m-trace/apps/api`).

---

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

### v0.3.0 — Stream-Analyzer

- HLS-Manifest-Parsing
- Segment-Dauer-Checks
- Target-Duration-Checks
- Grundlage für eigenständige CLI

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

| Umgebung                          | MVP-Status                  |
| --------------------------------- | --------------------------- |
| Chrome Desktop, aktuelle Stable   | unterstützt                 |
| Firefox Desktop, aktuelle Stable  | unterstützt                 |
| Safari Desktop, aktuelle Stable   | eingeschränkt               |
| Chromium-basierte Browser         | best effort                 |
| iOS Safari                        | im MVP nicht erforderlich   |
| Android Chrome                    | im MVP nicht erforderlich   |
| Smart-TV-Browser                  | nicht im Scope              |
| Embedded WebViews                 | nicht im Scope              |

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

Das Projekt ist in der Pre-MVP-`0.1.0`-Phase: Backend-Spike abgeschlossen, Lastenheft `1.1.2` verbindlich, `apps/api`-Skelett auf `main` integriert.

Leitende Dokumente:

```text
docs/lastenheft.md           # Anforderungen (verbindlich, 1.1.2)
docs/roadmap.md              # Status, Folge-ADRs, offene Entscheidungen
docs/adr/0001-backend-stack.md   # Backend-Entscheidung (Accepted: Go)
docs/spike/                  # Spike-Spezifikation, API-Kontrakt, Protokoll
docs/plan-spike.md           # Spike-Implementierungsplan
```

Nächste Schritte stehen in `docs/roadmap.md` §2.

---

## Lizenz

[MIT License](LICENSE).

---

## Mitarbeit

Beiträge sind noch nicht offen, da sich das Repository in der initialen Planungsphase befindet.

Geplante Bereiche für Beiträge:

- Player-SDK
- hls.js-Telemetrie
- Backend-API
- MediaMTX-Beispiele
- OpenTelemetry-Modellierung
- Prometheus-/Grafana-Dashboards
- HLS-/DASH-Analyzer
- SRT-Metriken

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
