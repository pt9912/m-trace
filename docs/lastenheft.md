# Lastenheft: m-trace

**Projektname:** m-trace  
**Dokumenttyp:** Lastenheft  
**Version:** 0.7.0  
**Status:** Entwurf  
**Lizenzziel:** Open Source, bevorzugt Apache-2.0 oder MIT  
**Architekturstil:** Mono-Repo mit hexagonaler Architektur  
**Primärer Stack:** Go oder Micronaut nach technischem Spike, SvelteKit, TypeScript, Docker, OpenTelemetry  

---

## 1. Ziel des Projekts

m-trace ist ein Open-Source-Projekt zur lokalen und produktionsnahen Beobachtung, Analyse und Diagnose von Media-Streaming-Workflows.

Das Projekt soll Entwicklern, DevOps-Teams und Streaming-Betreibern ermöglichen, Live-Streams lokal und später auch in realen Umgebungen zu überwachen, Playback-Metriken zu erfassen, HLS-/DASH-Streams zu analysieren und Streaming-Probleme schneller einzugrenzen.

Der erste Fokus liegt auf einem reproduzierbaren lokalen Streaming-Labor mit Dashboard, Backend, Player-SDK, OpenTelemetry-Anbindung und Beispiel-Streaming-Server.

---

## 2. Ausgangssituation

Media-Streaming-Systeme bestehen häufig aus mehreren lose gekoppelten Komponenten:

- Encoder, z. B. OBS oder FFmpeg
- Ingest-Protokolle, z. B. RTMP oder SRT
- Media-Server, z. B. MediaMTX oder SRS
- Ausspielung über HLS, DASH, WebRTC oder ähnliche Protokolle
- Browser-Player
- Monitoring- und Logging-Systeme

In der Praxis ist die Fehlersuche oft schwierig, weil Informationen über Player-Verhalten, Stream-Zustand, Segment-Probleme, Latenz und Infrastrukturmetriken über mehrere Systeme verteilt sind.

m-trace soll diese Lücke schließen, indem es ein einfach startbares, erweiterbares und beobachtbares Streaming-Lab bereitstellt.

---

## 3. Projektvision

m-trace soll langfristig ein offenes Werkzeug für Streaming Observability und Stream-Diagnose werden.

Die langfristige Vision umfasst:

- lokale Streaming-Testumgebung per Docker Compose
- Browser-Player-SDK für Playback-Metriken
- API zur Annahme und Verarbeitung von Playback- und Stream-Events
- Dashboard für Live-Metriken und Sessions
- HLS-/DASH-/CMAF-Analyse
- OpenTelemetry-Export
- Prometheus- und Grafana-Integration
- SRT-, RTMP-, HLS-, DASH- und WebRTC-Beispiele
- erweiterbare Adapter für verschiedene Media-Server

---

## 4. Differenzierung und Marktpositionierung

Der Markt für Media-Streaming-Observability ist bereits gut besetzt. Kommerzielle Anbieter wie Mux Data, Bitmovin Analytics, NPAW/YOUBORA und Conviva decken viele klassische QoE- und Analytics-Anwendungsfälle ab.

m-trace soll sich deshalb nicht als allgemeines Video-Analytics-Produkt positionieren, sondern als offener, selbsthostbarer und OpenTelemetry-nativer Diagnose-Stack für Streaming-Infrastruktur.

### 4.1 Zentrale Differenzierung

Die zentrale Lücke liegt in der gemeinsamen Betrachtung von:

- Ingest
- Media Server / Origin
- Manifesten und Segmenten
- Player-Sessions
- Observability-Pipelines

Das Alleinstellungsmerkmal soll sein:

```text
OpenTelemetry-native streaming observability from ingest to player.
```

### 4.2 OpenTelemetry-native Ansatz

m-trace soll Player-Sessions, Stream-Ereignisse und Infrastrukturzustände so modellieren, dass sie in bestehende OpenTelemetry-Pipelines passen.

Ziel ist nicht ein weiteres isoliertes Monitoring-Silo, sondern Integration mit bestehenden Systemen wie:

- OpenTelemetry Collector
- Tempo
- Loki
- Mimir
- Prometheus
- Grafana
- ClickHouse oder VictoriaMetrics für hochvolumige Events

Ein wichtiges Zielbild ist die Modellierung einer Player-Session als Trace.

Beispielhafte Trace-Struktur:

```text
Player Session Trace
├── manifest_request
├── segment_request
├── segment_request
├── startup_time
├── bitrate_switch
├── rebuffer_event
└── playback_error
```

Damit wird eine spätere End-to-End-Korrelation zwischen Encoder, Ingest, Origin und Player möglich.

### 4.3 SRT als späterer starker Hebel

SRT ist für Contribution-Workflows, Broadcaster und Remote-Produktion besonders interessant.

m-trace soll später SRT-spezifische Metriken sichtbar machen, insbesondere:

- RTT
- Packet Loss
- Retransmissions
- verfügbare Bandbreite
- Send- und Receive-Buffer
- Verbindungsstabilität
- Link Health
- Failover-Zustände

Dieser Bereich ist für spätere Versionen ein hohes Differenzierungspotenzial, aber nicht Bestandteil des ersten MVP.

### 4.4 Manifest Analyzer als eigenständiger Wert

Der HLS-/DASH-Manifest-Analyzer soll als eigenständige Library und CLI betrachtet werden, nicht nur als internes Dashboard-Feature.

Besonders relevant sind:

- HLS-Compliance
- DASH-Compliance
- Segment-Drift
- Target-Duration-Verletzungen
- `EXT-X-DISCONTINUITY`-Plausibilität
- Varianten-/Rendition-Konsistenz
- Codec-/Container-Hinweise

Eine offene, gut diagnostizierende Alternative zu schwer zugänglichen oder proprietären Validatoren kann eigenständig wertvoll sein.

### 4.5 Bewusste Abgrenzung

m-trace soll im ersten MVP nicht versuchen, kommerzielle QoE-Plattformen vollständig zu ersetzen.

Nicht der Fokus im MVP:

- vollständige Business-Analytics
- Zuschauer-Tracking
- A/B-Testing
- DRM-Analytics
- Ad-Analytics
- WebRTC-Monitoring
- Multi-CDN-Kostenoptimierung
- umfangreiche Endgeräte-Kompatibilitätsmatrix

Der erste Fokus liegt auf technischer Diagnose und OpenTelemetry-Integration.

---


## 5. Zielgruppen

### 5.1 Primäre Zielgruppen

- Softwareentwickler im Media-Streaming-Umfeld
- DevOps- und Plattformteams
- Betreiber kleiner und mittlerer Streaming-Plattformen
- Entwickler von Playern, Streaming-Backends oder Video-Workflows
- Open-Source-Contributors mit Interesse an Media-Infrastruktur

### 5.2 Sekundäre Zielgruppen

- Vereine, Bildungseinrichtungen und Event-Teams mit Self-Hosted-Streaming
- Unternehmen mit internen Live-Streaming-Workflows
- Entwickler, die Streaming-Protokolle lernen oder testen möchten

---

## 6. Geltungsbereich

Dieses Lastenheft beschreibt die Anforderungen an die erste öffentliche Projektphase von m-trace.

Der Fokus liegt auf:

- Mono-Repo-Struktur
- hexagonaler Architektur
- lokaler Entwicklungsumgebung
- lauffähigem Docker-Compose-Setup
- Backend-API in Go oder Micronaut nach technischem Spike
- SvelteKit Dashboard
- TypeScript Player-SDK
- einfachem Stream Analyzer
- OpenTelemetry-Grundlagen
- Dokumentation und Open-Source-Projektstruktur

Nicht Bestandteil der ersten Projektphase sind:

- vollständige Produktionsplattform
- Mandantenfähigkeit
- Abrechnungssystem
- DRM
- Benutzerverwaltung mit SSO
- Kubernetes-Produktionsbetrieb
- hochverfügbare Streaming-Infrastruktur
- kommerzielles CDN-Management

---

## 7. Funktionale Anforderungen

### 7.1 Mono-Repo

Das Projekt muss als Mono-Repo organisiert werden.

#### Anforderungen

- Das Repository muss alle Hauptbestandteile des Projekts enthalten.
- Anwendungen müssen unter `apps/` liegen.
- Wiederverwendbare Libraries müssen unter `packages/` liegen.
- Hilfsdienste müssen unter `services/` liegen.
- Beispiele müssen unter `examples/` liegen.
- Observability-Konfigurationen müssen unter `observability/` liegen.
- Deployment-Artefakte müssen unter `deploy/` liegen.
- Dokumentation muss unter `docs/` liegen.
- Skripte müssen unter `scripts/` liegen.

#### Zielstruktur

```text
m-trace/
├── apps/
│   ├── api/                    # Backend/API
│   ├── dashboard/              # SvelteKit Web UI
│   ├── ingest-gateway/         # optionaler Ingest-/Routing-Service
│   ├── analyzer-api/           # optionaler Analyse-Service
│   ├── control-plane/          # spätere Verwaltungs-/Admin-App
│   └── demo-player/            # isolierte Player-Demo-App
├── packages/
│   ├── player-sdk/
│   ├── stream-analyzer/
│   ├── shared-types/
│   ├── ui/
│   └── config/
├── services/
│   ├── stream-generator/
│   ├── otel-collector/
│   └── media-server/
├── examples/
│   ├── srs/
│   ├── mediamtx/
│   ├── hls/
│   ├── dash/
│   ├── srt/
│   └── webrtc/
├── observability/
│   ├── prometheus/
│   ├── grafana/
│   └── otel/
├── deploy/
│   ├── compose/
│   ├── docker/
│   └── k8s/
├── docs/
├── scripts/
├── docker-compose.yml
├── Makefile
├── README.md
└── CHANGELOG.md
```

---

### 7.2 Hexagonale Architektur

Die fachlich relevanten Anwendungen und Libraries müssen nach hexagonaler Architektur strukturiert werden.

#### Anforderungen

- Fachlogik muss im Ordner `hexagon/` liegen.
- Technische Ein- und Ausgänge müssen im Ordner `adapters/` liegen.
- Abhängigkeiten müssen von außen nach innen zeigen.
- Die Domain darf keine Framework-, HTTP-, Datenbank- oder Docker-Abhängigkeiten enthalten.
- Ports müssen als Schnittstellen definiert werden.
- Adapter müssen Ports implementieren oder Use Cases aufrufen.
- DTOs dürfen nicht Teil der Domain sein.

#### Standardstruktur

```text
src/
├── hexagon/
│   ├── domain/
│   ├── port/
│   │   ├── in/
│   │   └── out/
│   └── application/
└── adapters/
    ├── in/
    └── out/
```

#### Abhängigkeitsregel

```text
adapters → hexagon
```

Nicht erlaubt:

```text
hexagon → adapters
```

---

### 7.3 API-Anwendung

Die API-Anwendung muss unter `apps/api` liegen. Die finale Backend-Technologie wird nach einem technischen Spike entschieden.

#### Hauptaufgaben

- Annahme von Playback-Events
- Verwaltung von Stream-Sessions
- Bereitstellung von Metriken
- Weitergabe von Telemetrie an OpenTelemetry
- Bereitstellung von Daten für das Dashboard
- Integration des Stream Analyzers

#### Mindest-Endpunkte für den MVP

| Methode | Pfad | Zweck |
|---|---|---|
| `POST` | `/api/playback-events` | Annahme eines Playback-Events |
| `GET` | `/api/stream-sessions` | Liste bekannter Stream-Sessions |
| `GET` | `/api/stream-sessions/{id}` | Details einer Stream-Session |
| `GET` | `/api/health` | Health Check |
| `GET` | `/api/metrics` | technische Metriken, sofern aktiviert |

#### Beispielhafte API-Domänenobjekte

- `Project`
- `ProjectId`
- `ProjectToken`
- `AllowedOrigin`
- `StreamSession`
- `StreamId`
- `PlaybackEvent`
- `PlaybackMetric`
- `PlaybackError`
- `StreamHealth`
- `LatencyMeasurement`

---

### 7.4 Dashboard

Das Dashboard muss unter `apps/dashboard` liegen und mit SvelteKit umgesetzt werden.

#### Hauptaufgaben

- Anzeige laufender Stream-Sessions
- Anzeige aktueller Playback-Metriken
- Anzeige von Fehlern und Warnungen
- Anzeige einfacher Stream-Health-Zustände
- Anzeige von Backend- und Telemetrie-Status
- Integration eines Test-Players

#### Mindestansichten für den MVP

| Ansicht | Zweck |
|---|---|
| Startseite | Überblick über lokale Demo |
| Stream Sessions | Liste aktiver und vergangener Sessions |
| Session Details | Detailansicht zu Metriken und Events |
| Test Player | HLS-Testplayer mit eingebundenem Player-SDK |
| System Status | Status von API, Media Server und Observability |

#### Frontend-Architektur

Das Dashboard muss nicht zwingend vollständig hexagonal aufgebaut werden. Es soll eine pragmatische Feature-Struktur verwenden.

```text
apps/dashboard/src/
├── lib/
│   ├── api/
│   ├── components/
│   ├── features/
│   ├── stores/
│   └── types/
└── routes/
```

---

### 7.5 Weitere Anwendungen im Mono-Repo

Neben `apps/api` und `apps/dashboard` soll das Mono-Repo so vorbereitet werden, dass weitere Anwendungen sauber ergänzt werden können.

Die Detailarchitektur der Pflicht-Apps wird nur einmal verbindlich beschrieben. Spätere App-Beschreibungen dürfen diese Struktur nicht duplizieren, sondern nur Verantwortlichkeiten und Abgrenzungen ergänzen. Nicht jede App muss im ersten MVP vollständig implementiert sein, aber ihre fachliche Rolle, Abgrenzung und spätere Architektur sollen im Lastenheft definiert sein.

#### Grundregel

Jede Anwendung unter `apps/` ist eine eigenständig startbare Anwendung oder ein klar abgegrenzter Dienst mit eigenem Build, eigener Konfiguration und eigener Verantwortlichkeit.

Wiederverwendbare Fachlogik gehört nicht direkt in eine App, sondern in `packages/`.

---

#### 7.5.1 `apps/api`

`apps/api` ist die zentrale Backend-API für Playback-Events, Stream-Sessions, Dashboard-Daten und Telemetrie.

Status im MVP: **Muss**

Hauptaufgaben:

- Playback-Events annehmen
- Stream-Sessions verwalten
- Metriken vorbereiten oder exportieren
- Daten für Dashboard bereitstellen
- Stream Analyzer anbinden
- Health Checks bereitstellen

Architektur:

```text
apps/api/
├── src/
│   ├── hexagon/
│   │   ├── domain/
│   │   ├── port/
│   │   │   ├── in/
│   │   │   └── out/
│   │   └── application/
│   └── adapters/
│       ├── in/
│       │   ├── http/
│       │   └── websocket/
│       └── out/
│           ├── persistence/
│           ├── telemetry/
│           └── analyzer/
└── Dockerfile
```

---

#### 7.5.2 `apps/dashboard`

`apps/dashboard` ist die Weboberfläche für lokale Demo, Stream-Sessions, Playback-Events, Test-Player und Systemstatus.

Status im MVP: **Muss**

Hauptaufgaben:

- Live-Übersicht anzeigen
- Test-Player bereitstellen
- Playback-Events anzeigen
- Stream-Sessions anzeigen
- API-Status anzeigen
- Links zu Grafana, Prometheus und Media-Server-Konsole anzeigen

Architektur:

```text
apps/dashboard/src/
├── lib/
│   ├── api/
│   ├── components/
│   ├── features/
│   ├── stores/
│   └── types/
└── routes/
```

Hinweis: Das Dashboard muss nicht strikt hexagonal aufgebaut werden. Wenn später echte Fachlogik entsteht, kann innerhalb einzelner Features eine kleine Hexagon-Struktur eingeführt werden.

---

#### 7.5.3 `apps/demo-player`

`apps/demo-player` ist keine MVP-App.

Im MVP wird die Player-Demo als Route im Dashboard umgesetzt:

```text
apps/dashboard/src/routes/demo/
```

Eine separate App `apps/demo-player` wird erst sinnvoll, wenn der Player-SDK als eigenständiges Produktpaket demonstriert werden soll.

Status im MVP: **Nicht Bestandteil**

Spätere Hauptaufgaben:

- HLS-Teststream abspielen
- Player-SDK isoliert integrieren
- erzeugte Events sichtbar machen
- SDK-Konfiguration testen
- als minimale Referenzintegration für externe Nutzer dienen

Warum nicht im MVP:

Das Dashboard kann die Demo-Funktion zunächst ausreichend abdecken. Eine eigene App würde Build-, Deployment- und Dokumentationsaufwand erhöhen, ohne den ersten Nutzwert wesentlich zu steigern.


---

#### 7.5.4 `apps/ingest-gateway`

`apps/ingest-gateway` ist ein späterer Dienst zur Verwaltung von Ingest-Flows, Stream-Keys und Routing-Regeln.

Status im MVP: **Kann**

Hauptaufgaben:

- Stream-Keys verwalten
- Ingest-Endpunkte beschreiben
- Routing-Regeln für Streams definieren
- Webhooks bei Stream-Start und Stream-Ende auslösen
- SRT-/RTMP-Konfigurationen vorbereiten
- Media-Server-Konfigurationen generieren oder validieren

Mögliche Endpunkte:

| Methode | Pfad | Zweck |
|---|---|---|
| `POST` | `/api/ingest/streams` | neuen Ingest-Stream registrieren |
| `GET` | `/api/ingest/streams` | Ingest-Streams listen |
| `POST` | `/api/ingest/streams/{id}/rotate-key` | Stream-Key erneuern |
| `POST` | `/api/ingest/hooks/stream-started` | Start-Webhook empfangen |
| `POST` | `/api/ingest/hooks/stream-ended` | Ende-Webhook empfangen |

Architektur:

```text
apps/ingest-gateway/
├── src/
│   ├── hexagon/
│   │   ├── domain/
│   │   │   ├── model/
│   │   │   └── service/
│   │   ├── port/
│   │   │   ├── in/
│   │   │   └── out/
│   │   └── application/
│   └── adapters/
│       ├── in/
│       │   ├── http/
│       │   └── webhook/
│       └── out/
│           ├── persistence/
│           ├── media_server/
│           └── telemetry/
└── Dockerfile
```

Mögliche Domain-Objekte:

- `IngestStream`
- `StreamKey`
- `IngestEndpoint`
- `RoutingRule`
- `MediaServerTarget`
- `IngestProtocol`
- `StreamLifecycleEvent`

---

#### 7.5.5 `apps/analyzer-api`

`apps/analyzer-api` ist ein optionaler separater HTTP-Service für Stream-Analysen. Er kapselt `packages/stream-analyzer` und stellt Analysefunktionen über HTTP bereit.

Status im MVP: **Kann**

Warum optional:

Im ersten MVP kann `apps/api` den Analyzer direkt als Library nutzen. Ein separater Analyse-Service lohnt sich erst, wenn Analysen schwerer werden, unabhängig skaliert werden sollen oder unsichere externe URLs isoliert verarbeitet werden müssen.

Hauptaufgaben:

- HLS-URL entgegennehmen
- Manifest analysieren
- Analyseergebnis als JSON liefern
- Fehler und Warnungen normalisieren
- spätere DASH-/CMAF-Analyse anbieten
- Sicherheitsgrenzen für externe URL-Abrufe schaffen

Mögliche Endpunkte:

| Methode | Pfad | Zweck |
|---|---|---|
| `POST` | `/api/analyze/hls` | HLS-Stream analysieren |
| `POST` | `/api/analyze/dash` | DASH-Stream analysieren, später |
| `GET` | `/api/analyze/jobs/{id}` | Analysejob abfragen, später |

Architektur:

```text
apps/analyzer-api/src/
├── hexagon/
│   ├── domain/
│   ├── port/
│   │   ├── in/
│   │   └── out/
│   └── application/
└── adapters/
    ├── in/
    │   └── http/
    └── out/
        ├── analyzer/
        ├── http_fetcher/
        └── telemetry/
```

Mögliche Domain-Objekte:

- `AnalysisJob`
- `StreamAnalysisRequest`
- `StreamAnalysisResult`
- `ManifestWarning`
- `ManifestError`
- `SegmentTimingIssue`

---

#### 7.5.6 `apps/control-plane`

`apps/control-plane` ist eine spätere Verwaltungsanwendung für produktionsnahe m-trace-Installationen.

Status im MVP: **Nicht Bestandteil**, nur vorbereitet

Hauptaufgaben in späteren Versionen:

- Konfiguration mehrerer m-trace-Instanzen
- Verwaltung von Media-Servern
- Verwaltung von Stream-Profilen
- Verwaltung von Teams und Projekten
- Audit-Log
- API-Keys
- Integrationen
- spätere Benutzerverwaltung

Wichtige Abgrenzung:

`apps/control-plane` darf im MVP nicht gebaut werden. Sonst entsteht zu früh eine Plattform, bevor das eigentliche Streaming-Diagnoseproblem gelöst ist.

Mögliche spätere Architektur:

```text
apps/control-plane/
├── backend/
└── frontend/
```

Oder bei klarer Trennung:

```text
apps/control-plane-api/
apps/control-plane-ui/
```

Die finale Aufteilung ist erst sinnvoll, wenn echte Anforderungen für Mehrbenutzerbetrieb und Administration vorliegen.

---

#### 7.5.7 App-Übersicht nach Priorität

| App | Zweck | MVP-Status | Technologie |
|---|---|---|---|
| `apps/api` | zentrale Backend-API | Muss | Go oder Micronaut, nach Spike |
| `apps/dashboard` | Web-Dashboard | Muss | SvelteKit |
| `apps/demo-player` | SDK-Referenz und Testplayer | Nicht MVP, zunächst `/demo`-Route | SvelteKit oder Vite |
| `apps/ingest-gateway` | Stream-Key, Ingest und Routing | Kann | Go oder Micronaut, nach Spike |
| `apps/analyzer-api` | separater Analyse-Service | Kann | Micronaut oder Node.js |
| `apps/control-plane` | spätere Verwaltungsplattform | Später | offen |

---

#### 7.5.8 Empfehlung für die erste Umsetzung

Für den ersten lauffähigen Release sollen nur folgende Apps aktiv implementiert werden:

```text
apps/
├── api/
└── dashboard/
```

Der Demo-Player wird zunächst als Route im Dashboard umgesetzt:

```text
apps/dashboard/src/routes/demo/
```

Folgende Apps sollen zunächst höchstens als dokumentierte Platzhalter existieren:

```text
apps/
├── ingest-gateway/
├── analyzer-api/
└── control-plane/
```

Das verhindert Architektur-Overhead und hält den ersten Release realistisch.

---

### 7.6 Player-SDK

Das Player-SDK muss unter `packages/player-sdk` liegen und in TypeScript umgesetzt werden.

#### MVP-Abgrenzung

Im MVP unterstützt das Player-SDK nur `hls.js`.

Weitere Player-Adapter sind spätere Erweiterungen:

- dash.js
- Shaka Player
- Video.js
- native Safari HLS
- WebRTC `getStats()`, separat in späterer Phase

Ein Player-SDK von Grund auf ist ein eigenes Subprojekt und darf nicht unterschätzt werden. Unterschiedliche Player liefern unterschiedliche Events, Timing-Modelle und Metriken. Safari mit nativem HLS bietet besonders wenig Introspektion.

#### Browser-Support im MVP

Der MVP definiert bewusst eine enge Browser-Matrix, um den Testaufwand realistisch zu halten.

| Umgebung | Status im MVP |
|---|---|
| Chrome Desktop, aktuelle stabile Version | unterstützt |
| Firefox Desktop, aktuelle stabile Version | unterstützt |
| Safari Desktop, aktuelle stabile Version | eingeschränkt, nur Basis-Playback |
| Chromium-basierte Browser | best effort |
| iOS Safari | nicht verpflichtend im MVP |
| Android Chrome | nicht verpflichtend im MVP |
| Smart-TV Browser | explizit nicht im Scope |
| Embedded WebViews | explizit nicht im Scope |

Für den MVP gilt:

- hls.js ist der primäre Integrationspfad.
- Native Safari-HLS-Introspektion ist nicht Ziel von `0.1.0`.
- Mobile Browser werden später gezielt getestet.
- Smart-TV- und Set-Top-Box-Umgebungen sind vorerst ausgeschlossen.

#### Hauptaufgaben

- Anbindung an ein `HTMLVideoElement`
- Erfassung von Playback-Events
- Erfassung einfacher Metriken
- Versand der Events über OpenTelemetry Web SDK oder HTTP an die API
- Trennung von Browser-Adapter und fachlicher Tracking-Logik

#### Zu erfassende Events im MVP

| Event | Beschreibung |
|---|---|
| `playback_started` | Wiedergabe wurde gestartet |
| `playback_paused` | Wiedergabe wurde pausiert |
| `playback_ended` | Wiedergabe wurde beendet |
| `startup_time_measured` | Startup-Zeit wurde gemessen |
| `rebuffer_started` | Buffering hat begonnen |
| `rebuffer_ended` | Buffering wurde beendet |
| `quality_changed` | Qualitäts-/Bitratenwechsel erkannt |
| `playback_error` | Player-Fehler erkannt |
| `metrics_sampled` | Regelmäßiger Metrik-Snapshot |

#### Zielstruktur im MVP

Das Player-SDK wird im MVP bewusst pragmatisch aufgebaut. Es nutzt keine vollständige Hexagon-Ceremony.

```text
packages/player-sdk/src/
├── core/
│   ├── session.ts
│   ├── event-buffer.ts
│   └── event-normalizer.ts
├── adapters/
│   └── hlsjs/
│       └── hlsjs-tracker.ts
├── transport/
│   ├── http-transport.ts
│   └── otel-transport.ts
├── types/
│   ├── events.ts
│   ├── config.ts
│   └── schema.ts
└── index.ts
```

Eine strengere Port-/Adapter-Struktur wird erst eingeführt, wenn mehr als ein Player-Adapter produktiv unterstützt wird.

---

### 7.7 Stream Analyzer

Der Stream Analyzer muss unter `packages/stream-analyzer` liegen und in TypeScript umgesetzt werden.

#### Hauptaufgaben

- Abruf von HLS-Manifesten
- Analyse einfacher Manifest-Eigenschaften
- Prüfung von Segment-Dauern
- Erkennung offensichtlicher Inkonsistenzen
- Bereitstellung einer API für Backend und CLI
- Vorbereitung für DASH- und CMAF-Analyse

#### Mindestfunktionen für den MVP

- HLS Master Playlist erkennen
- HLS Media Playlist erkennen
- Varianten und Renditions extrahieren
- Segment-Anzahl bestimmen
- durchschnittliche Segment-Dauer berechnen
- Abweichungen bei Segment-Dauern erkennen
- einfache Live-Latenz-Schätzung
- Analyseergebnis als JSON liefern

#### CLI-Ziel

```bash
pnpm m-trace check https://example.com/live/master.m3u8
```

---

### 7.8 Lokales Streaming-Lab

Das Projekt muss eine lokale Streaming-Testumgebung bereitstellen.

#### Anforderungen

- Start per Docker Compose
- Media Server für lokale Tests
- FFmpeg-basierter Teststream
- API erreichbar unter `localhost`
- Dashboard erreichbar unter `localhost`
- Prometheus und Grafana optional verfügbar
- OpenTelemetry Collector optional verfügbar

#### Mindestdienste

| Dienst | Zweck |
|---|---|
| `api` | Backend-API |
| `dashboard` | SvelteKit UI |
| `mediamtx` | lokaler Media Server |
| `stream-generator` | FFmpeg-Teststream |
| `otel-collector` | OpenTelemetry Collector |
| `prometheus` | Metrikspeicherung |
| `grafana` | Visualisierung |

#### Erwarteter Startbefehl

```bash
make dev
```

Oder direkt:

```bash
docker compose up --build
```

---

### 7.9 Observability

Das Projekt muss Observability von Beginn an berücksichtigen.

#### Anforderungen

- API muss strukturierte Logs erzeugen.
- API muss Health Checks bereitstellen.
- API soll OpenTelemetry unterstützen.
- Playback-Events sollen als Metriken oder Traces exportierbar sein.
- Prometheus soll technische Metriken erfassen können.
- Grafana soll mit einem einfachen Beispiel-Dashboard ausgeliefert werden.

#### Mindestmetriken

| Metrik | Beschreibung |
|---|---|
| `mtrace_playback_events_total` | Anzahl empfangener Playback-Events |
| `mtrace_playback_errors_total` | Anzahl empfangener Playback-Fehler |
| `mtrace_active_sessions` | Anzahl aktiver Sessions |
| `mtrace_rebuffer_events_total` | Anzahl Buffering-Ereignisse |
| `mtrace_startup_time_ms` | gemessene Startup-Zeit |
| `mtrace_api_requests_total` | API Requests |
| `mtrace_dropped_events_total` | Anzahl verworfener Events |
| `mtrace_rate_limited_events_total` | Anzahl durch Rate Limits abgelehnter Events |
| `mtrace_invalid_events_total` | Anzahl wegen Schema/Auth abgelehnter Events |

---

### 7.10 Datenmodell, Cardinality und Storage

m-trace muss von Beginn an zwischen aggregierten Metriken, hochvolumigen Events und per-Session-Daten unterscheiden.

#### Problem

Prometheus ist nicht geeignet für hochkardinale Labels wie:

- `session_id`
- `viewer_id`
- `client_ip`
- `user_agent`
- `segment_url`
- `request_id`

Diese Labels können bei Player-Telemetrie sehr schnell zu unkontrollierbarer Cardinality führen.

#### Anforderungen

- Prometheus darf nur für aggregierte Metriken verwendet werden.
- `session_id` darf nicht als Prometheus-Label verwendet werden.
- Per-Session-Daten sollen als Traces oder Events modelliert werden.
- Für hochvolumige Eventdaten muss eine spätere Storage-Option vorgesehen werden.
- Das System muss Sampling vorbereiten.
- Das Telemetrie-Modell muss Datenschutz und Cardinality gemeinsam berücksichtigen.

#### Empfohlene Zuordnung

| Datentyp | Geeigneter Speicher | Zweck |
|---|---|---|
| aggregierte technische Metriken | Prometheus / Mimir | Dashboards, Alerts |
| Player-Session-Verläufe | Tempo / Traces | Debugging einzelner Sessions |
| hochvolumige Events | ClickHouse / VictoriaMetrics / später | Analyse und Historie |
| Logs | Loki | technische Fehlersuche |
| Konfiguration | PostgreSQL / SQLite / später | persistente Projekt- und Streamdaten |

#### MVP-Entscheidung

Im ersten MVP sollen folgende Regeln gelten:

- Prometheus nur für Aggregate
- Player-Sessions als OpenTelemetry-Traces vorbereiten
- In-Memory-Speicherung nur für lokale Demo
- keine produktive Langzeitspeicherung im MVP
- keine `session_id`-Labels in Prometheus

---

### 7.11 Telemetry Ingest, Event-Schema und SDK-Budget

Die Telemetrie-Schnittstelle ist ein Kernbestandteil des Projekts und muss früh spezifiziert werden.

#### Authentifizierung von Player-Events

Das Browser-SDK darf nicht dauerhaft gegen einen vollständig offenen Ingest-Endpunkt senden.

MVP-Anforderungen:

- Events enthalten eine `project_id`.
- Events werden mit einem öffentlichen Project Token oder einem kurzlebigen Ingest Token versehen.
- Das Backend validiert erlaubte Origins.
- Tokens dürfen keine Secrets mit hoher Kritikalität sein, da Browser-Code öffentlich ist.
- Rate Limits gelten pro Project, Origin und IP-Bereich.

Spätere Erweiterungen:

- serverseitig signierte Session Tokens
- rotierbare Project Tokens
- tenant-spezifische Ingest Policies

#### Schema-Versionierung

Jedes Event muss eine Schema-Version enthalten.

Pflichtfelder im Wire-Format:

```json
{
  "schema_version": "1.0",
  "event_name": "rebuffer_started",
  "project_id": "demo",
  "session_id": "01J...",
  "client_timestamp": "2026-04-28T12:00:00.000Z",
  "sdk": {
    "name": "@m-trace/player-sdk",
    "version": "0.1.0"
  }
}
```

Anforderungen:

- neue Felder müssen abwärtskompatibel sein
- unbekannte Felder dürfen nicht zum Fehler führen
- entfernte Felder müssen über mindestens eine Minor-Version toleriert werden
- Breaking Changes erfordern neue Major-Version der Event-Schemas

#### Backpressure und Rate Limiting

Die Ingest-API muss Überlastung kontrolliert behandeln.

MVP-Anforderungen:

- maximale Event-Batch-Größe definieren
- maximale Request-Rate pro Project definieren
- HTTP `429` bei Rate Limit
- HTTP `202` für angenommene Events
- Events dürfen bei lokaler Überlast verworfen werden, wenn dies als Dropped-Event-Metrik sichtbar wird
- SDK muss Sampling und Batch-Größe konfigurieren können

#### Zeitstempel und Time Skew

Browser-Clocks sind unzuverlässig. Das Backend muss daher zwischen Client-Zeit und Server-Zeit unterscheiden.

Pflichtfelder:

- `client_timestamp`
- `server_received_at`
- optional `client_time_origin`
- optional `sequence_number`

Regeln:

- Ordering innerhalb einer Session bevorzugt über `sequence_number`
- Latenzberechnungen niemals blind nur aus Client-Zeit ableiten
- Backend muss auffälligen Time Skew markieren können

#### Performance-Budget für das Player-SDK

Das SDK darf Playback nicht stören.

MVP-Budget:

| Kennzahl | Ziel |
|---|---|
| Bundle-Größe | kleiner als 30 KB gzip ohne hls.js |
| Event-Verarbeitung | unter 5 ms pro Event im Normalfall |
| Hot Path | keine synchronen Netzwerkaufrufe |
| Transport | batchingfähig |
| Fehlerverhalten | niemals Playback abbrechen |
| Sampling | konfigurierbar |

#### OpenTelemetry Semantic Conventions

m-trace soll sich an bestehenden OpenTelemetry-Konventionen orientieren und eigene Media-Konventionen nur dort ergänzen, wo keine passende Konvention existiert.

Strategie:

- bestehende HTTP-, Client-, Browser- und Runtime-Konventionen nutzen
- eigene Attribute mit stabilem Prefix definieren, z. B. `mtrace.*`
- Media-spezifische Semantik dokumentieren
- spätere Kompatibilität mit entstehenden OTel-Media-Konventionen einplanen

---

### 7.12 Dokumentation

Das Projekt muss eine entwicklerfreundliche Dokumentation enthalten.

#### Pflichtdokumente

| Datei | Zweck |
|---|---|
| `README.md` | Einstieg und Schnellstart |
| `CHANGELOG.md` | Änderungsverlauf |
| `CONTRIBUTING.md` | Beitragsregeln |
| `LICENSE` | Lizenz |
| `SECURITY.md` | Sicherheitsmeldungen |
| `docs/architecture.md` | Architekturüberblick |
| `docs/local-development.md` | lokale Entwicklung |
| `docs/telemetry-model.md` | Telemetrie- und Eventmodell |
| `docs/player-sdk.md` | Player-SDK-Nutzung |
| `docs/stream-analyzer.md` | Stream Analyzer |
| `docs/roadmap.md` | geplante Entwicklung |

---

## 8. Nichtfunktionale Anforderungen

### 8.1 Plattform

- Entwicklung muss unter Linux möglich sein.
- Entwicklung muss mit VS Code kompatibel sein.
- Lokaler Betrieb muss über Docker möglich sein.
- Build-Prozesse müssen ohne proprietäre Dienste funktionieren.

### 8.2 Wartbarkeit

- Fachlogik muss testbar sein, ohne externe Infrastruktur zu starten.
- Domain-Klassen dürfen keine Framework-Abhängigkeiten enthalten.
- Ports müssen klar benannt und dokumentiert sein.
- Adapter müssen austauschbar sein.
- Technische Implementierungen dürfen nicht in die Domain-Schicht lecken.

### 8.3 Erweiterbarkeit

Das Projekt muss vorbereitet sein für spätere Erweiterungen:

- MediaMTX-Adapter
- SRT-Ingest-Metriken
- DASH-Analyse
- CMAF-Analyse
- WebRTC-Metriken
- Datenbankpersistenz
- Authentifizierung
- Multi-Stream-Betrieb
- Kubernetes Deployment
- CI-basierte Stream-Checks

### 8.4 Performance

Für den MVP gelten einfache Anforderungen:

- API muss lokale Demo-Last problemlos verarbeiten.
- Player-SDK darf Playback nicht merklich beeinflussen.
- Dashboard muss bei mehreren aktiven Sessions bedienbar bleiben.
- Event-Erfassung muss asynchron oder leichtgewichtig erfolgen.

### 8.5 Sicherheit

- Keine Secrets im Repository.
- `.env.example` muss Beispielwerte enthalten.
- Produktive Secrets müssen über Umgebungsvariablen gesetzt werden.
- CORS muss im lokalen Setup kontrolliert konfiguriert sein.
- Externe URLs für Stream-Analyse müssen später abgesichert werden, um SSRF-Risiken zu vermeiden.
- Security-Meldungen müssen über `SECURITY.md` beschrieben werden.

#### CORS- und CSP-Grundregeln für Player-Telemetrie

Für Browser-SDK-Telemetrie muss Cross-Origin-Kommunikation kontrolliert werden.

MVP-Anforderungen:

- erlaubte Origins werden pro Project konfiguriert
- SDK-Requests nutzen standardmäßig `credentials: "omit"`
- keine Cookies für Player-Telemetrie im MVP
- Preflight-fähige CORS-Konfiguration
- `Access-Control-Allow-Origin` darf nicht pauschal `*` sein, sobald Project Tokens genutzt werden
- erlaubte Methoden zunächst auf `POST` und `OPTIONS` begrenzen
- erlaubte Header explizit definieren, z. B. `Content-Type`, `X-MTrace-Project`, `X-MTrace-Token`
- CSP-Beispiele für `connect-src` müssen dokumentiert werden

Beispiel-CSP für eine Demo-Integration:

```text
Content-Security-Policy: connect-src 'self' https://m-trace.example.com;
```


### 8.6 Datenschutz und GDPR

Player-Telemetrie kann personenbezogene oder personenbeziehbare Daten enthalten. Dazu gehören insbesondere IP-Adressen, User-Agents, Session-IDs und grobe Standortinformationen.

Anforderungen:

- IP-Adressen dürfen im MVP nicht unnötig gespeichert werden.
- User-Agent-Daten müssen reduzierbar oder anonymisierbar sein.
- Session-IDs müssen pseudonym sein.
- Ein konfigurierbarer Anonymisierungs-Layer im Collector soll vorbereitet werden.
- Das Projekt muss dokumentieren, welche Telemetriedaten erhoben werden.
- Datenschutzfreundliche Defaults haben Vorrang vor maximaler Analyse-Tiefe.
- Für EU-Nutzung muss eine GDPR-freundliche Betriebsweise möglich sein.

### 8.7 Qualität

- Automatisierte Tests für Domain- und Application-Schicht
- Linting für TypeScript
- Tests für zentrale Backend-Use-Cases
- CI-Pipeline für Build und Test
- klare Commit- und Release-Konventionen
- CHANGELOG-Pflege ab dem ersten Release

---

## 9. Technologie-Strategie und Architekturentscheidungen

Die ursprüngliche Präferenz für Java/Micronaut ist technisch machbar, aber im Streaming-Observability-Umfeld nicht automatisch die strategisch stärkste Wahl.

Viele relevante Komponenten und Communities in diesem Bereich sind stark durch Go, Rust und TypeScript geprägt:

- Media-Server und Streaming-Infrastruktur häufig in Go
- OpenTelemetry Collector in Go
- Browser- und Player-Ökosystem stark in TypeScript
- performante Analyzer- und CLI-Werkzeuge häufig in Go oder Rust

### 9.1 Backend-Entscheidung

Für den MVP gibt es zwei realistische Optionen:

| Option | Vorteil | Nachteil |
|---|---|---|
| Go Backend | passt kulturell gut zu OTel, MediaMTX und Infrastruktur-Tools | weniger passend zur ursprünglichen Micronaut-Präferenz |
| JVM Backend (Micronaut) | vertrauter JVM-Stack, gute DI, gute Testbarkeit | kleinerer Contributor-Pool im Streaming-OSS-Umfeld |

Empfehlung für OSS-Adoption:

```text
Go für neue m-trace-Core-Services bevorzugen.
```

Pragmatische Alternative:

```text
Micronaut bleibt erlaubt, wenn die persönliche Umsetzungsgeschwindigkeit wichtiger ist als maximale OSS-Adoption.
```

Das Lastenheft muss diese Entscheidung bewusst offenhalten, bis der erste technische Spike abgeschlossen ist.

### 9.2 Hexagonale Architektur

Hexagonale Architektur soll nicht dogmatisch für alle Komponenten gelten.

Verbindliche Regel:

```text
Hexagonal nur dort, wo echte fachliche Anwendungslogik entsteht.
```

Empfohlene Anwendung:

| Komponente | Architektur |
|---|---|
| `apps/api` | hexagonal |
| `packages/stream-analyzer` | hexagonal oder klar geschichtete Library |
| `packages/player-sdk` | pragmatisch, keine vollständige Hexagon-Ceremony |
| `apps/dashboard` | Feature-Struktur |
| `apps/demo-player` | keine eigene App im MVP, höchstens Route im Dashboard |

Für das Player-SDK genügt eine leichte Adapter-Struktur:

```text
packages/player-sdk/src/
├── core/
├── adapters/
│   └── hlsjs/
├── transport/
└── types/
```

Ports und Use Cases sind dort erst nötig, wenn mehrere Player-Adapter tatsächlich implementiert werden.

### 9.3 Selbsthoster-first Konsequenz

Da der MVP auf Selbsthoster, kleine Plattformen, Broadcaster-Labs und technische Teams zielt, muss die Architektur zuerst einfach betreibbar sein.

Für den MVP bedeutet das:

- keine Mimir-Pflicht
- keine ClickHouse-Pflicht
- keine große Multi-Tenant-Architektur
- keine getrennte Demo-Player-App
- keine getrennte Analyzer-API
- bevorzugt lokale Speicherung mit SQLite oder In-Memory
- eingebaute Trace-/Session-Anzeige im Dashboard als Alternative zu Tempo
- Tempo, Mimir und ClickHouse nur als optionale spätere Integrationen

---

## 10. Technische Rahmenbedingungen

### 10.1 Backend

Die Backend-Technologie ist bis zum technischen Spike bewusst offen.

Zulässige Optionen für den MVP:

| Option | Sprache | Framework/Ansatz |
|---|---|---|
| Go | Go | Standard-Library, Chi, Fiber oder vergleichbar |
| JVM | Java oder Kotlin | Micronaut oder vergleichbar |

Entscheidungskriterien:

- Umsetzungsgeschwindigkeit
- Contributor-Fit im Streaming-/Observability-Umfeld
- OpenTelemetry-Integration
- Docker-Build-Komplexität
- Testbarkeit
- langfristige Wartbarkeit

Bis zur finalen Entscheidung müssen Architektur- und Verzeichnisbeispiele technologie-neutral verstanden werden.

Mindestanforderungen unabhängig vom Stack:

- HTTP API für Event-Ingest
- Health Check
- strukturierte Logs
- OpenTelemetry-kompatibles Eventmodell
- klare Trennung von Domain, Application und Adapters
- Containerisierung per Docker


### 10.2 Frontend

- Sprache: TypeScript
- Framework: SvelteKit
- Package Manager: pnpm
- Styling: zunächst pragmatisch, später UI-Package möglich
- Kommunikation: REST, später WebSocket oder SSE

### 10.3 Player-SDK

- Sprache: TypeScript
- Zielumgebung: Browser
- Build: pnpm
- Ausgabeformat: ESM
- Kernlogik frameworkfrei
- Adapter für Browser und HTTP

### 10.4 Stream Analyzer

- Sprache: TypeScript
- Zielumgebung: Node.js
- HLS zuerst
- DASH später
- CLI später
- API-kompatible JSON-Ergebnisse

### 10.5 Infrastruktur

- Docker
- Docker Compose
- MediaMTX als erster Media Server
- FFmpeg als Teststream-Generator
- OpenTelemetry Collector
- Prometheus
- Grafana

---

## 11. Abgrenzung zu ähnlichen Projekten

m-trace soll kein Ersatz sein für:

- OBS
- FFmpeg
- SRS
- MediaMTX
- Wowza
- Mux Data
- Grafana
- Prometheus
- kommerzielle Streaming-Plattformen

m-trace soll diese Systeme ergänzen, indem es lokale Reproduzierbarkeit, Player-Metriken, Stream-Diagnose und Observability verbindet.

---

## 12. MVP-Umfang

Der erste funktionsfähige MVP muss folgende Bestandteile enthalten:

### 12.1 Muss-Anforderungen

Der MVP wird bewusst enger gefasst. Für eine Solo-Umsetzung ist der ursprüngliche Scope zu groß. Realistisch ist ein kleiner, durchgängiger Pfad.

MVP-Ziel:

```text
MediaMTX + hls.js Demo Route + Player Events + OTel-kompatibles Eventmodell + Dashboard-Anzeige
```

Muss-Anforderungen für `0.1.0`:

- Mono-Repo-Struktur
- eine Backend-App unter `apps/api`
- eine Web-App unter `apps/dashboard`
- Demo-Player als `/demo`-Route im Dashboard, nicht als separate App
- `packages/player-sdk` mit hls.js-Adapter
- pragmatische SDK-Struktur ohne vollständige Hexagon-Ceremony
- Docker Compose Setup
- MediaMTX als erster Media Server
- FFmpeg-Teststream
- OpenTelemetry-kompatibles Eventmodell
- API-Endpunkt für Playback-Event-Batches
- einfache Session-Liste
- einfache Event-Anzeige
- einfache eingebaute Session-/Trace-Ansicht im Dashboard
- Prometheus nur für aggregierte Metriken
- lokale Speicherung per In-Memory oder SQLite
- README mit Schnellstart
- CHANGELOG mit initialem Eintrag

Nicht im `0.1.0`-MVP:

- separate `apps/demo-player`
- separate `apps/analyzer-api`
- `packages/stream-analyzer` als fertiges Paket
- Tempo als Pflichtkomponente
- Mimir oder ClickHouse
- WebRTC
- SRT-Health-View
- Multi-Tenant-Betrieb


### 12.2 Soll-Anforderungen

- SQLite-Persistenz statt reinem In-Memory
- Grafana-Dashboard für Aggregate
- einfache OTel-Collector-Konfiguration
- rudimentäre HLS-Manifest-Prüfung als interner Spike
- WebSocket oder SSE für Live-Updates
- CI mit GitHub Actions


### 12.3 Kann-Anforderungen

- eigenständiger Stream Analyzer als `packages/stream-analyzer`
- CLI für Stream Analyzer
- Tempo-Integration
- SRS-Beispiel
- DASH-Analyse
- SRT-Ingest-Beispiel
- SRT-Health-View
- Persistenz mit PostgreSQL
- ClickHouse- oder VictoriaMetrics-Anbindung
- Kubernetes-Manifeste
- Devcontainer
- Release-Automatisierung


---

## 13. Release-Plan

### 13.1 Version 0.1.0: OTel-native Local Demo

Ziel: Ein Entwickler kann das Repository klonen und lokal einen MediaMTX-basierten Teststream mit hls.js-Player, Player-Events und OpenTelemetry-Grundmodell sehen.

Akzeptanzkriterien:

- `make dev` startet alle notwendigen Dienste.
- Dashboard ist erreichbar.
- API ist erreichbar.
- Teststream läuft über MediaMTX.
- Player-SDK sendet hls.js-basierte Events.
- API nimmt Events an.
- Dashboard zeigt empfangene Events und einfache Session-Zusammenhänge.
- README beschreibt den Ablauf reproduzierbar.
- Prometheus enthält nur aggregierte Metriken.
- Player-Session-Traces sind vorbereitet oder exemplarisch sichtbar.

---

### 13.2 Version 0.2.0: Publizierbares Player SDK

Ziel: Das Player-SDK wird vom MVP-Prototyp zu einem eigenständig nutzbaren und dokumentierten npm-Paket ausgebaut.

Abgrenzung zu `0.1.0`:

`0.1.0` beweist den End-to-End-Pfad mit hls.js-Adapter und Event-Ingest.  
`0.2.0` stabilisiert das SDK als wiederverwendbares Paket mit Public API, Tests, Dokumentation und Versionierungsstrategie.

Akzeptanzkriterien:

- SDK ist als npm-Paket baubar und lokal installierbar.
- Public API ist dokumentiert.
- Event-Schema ist versioniert.
- hls.js-Adapter ist getestet.
- HTTP-Transport ist getestet.
- OTel-Transport ist vorbereitet oder experimentell nutzbar.
- SDK unterstützt Batching, Sampling und Retry-Grenzen.
- SDK hält das definierte Performance-Budget ein.
- Browser-Support-Matrix ist dokumentiert.
- Beispielintegration in der Dashboard-Route `/demo` ist dokumentiert.
- Kompatibilität zwischen SDK-Version und Event-Schema wird in CI geprüft.


---

### 13.3 Version 0.3.0: Stream Analyzer

Ziel: HLS-Streams können analysiert werden.

Akzeptanzkriterien:

- HLS Manifest kann geladen werden.
- Master Playlist kann erkannt werden.
- Media Playlist kann erkannt werden.
- Segment-Dauern werden geprüft.
- Ergebnis wird als JSON ausgegeben.
- API kann Analyzer nutzen.
- CLI-Grundlage existiert.

---

### 13.4 Version 0.4.0: Erweiterte Trace-Korrelation

Ziel: Die in `0.1.0` vorbereitete OTel-Grundlage wird zu einer nutzbaren Korrelationsschicht ausgebaut.

Akzeptanzkriterien:

- Player-Session-Traces werden konsistent erzeugt.
- Manifest-Requests, Segment-Requests und Player-Events werden in einem Trace zusammengeführt, soweit technisch möglich.
- Tempo kann optional als Trace-Backend verwendet werden.
- Dashboard kann Session-Verläufe auch ohne Tempo einfach anzeigen.
- Prometheus bleibt auf aggregierte Metriken beschränkt.
- Dropped-, Rate-Limited- und Invalid-Event-Metriken sind sichtbar.
- Dokumentation beschreibt Cardinality-Grenzen und Sampling-Strategie.


---

### 13.5 Version 0.5.0: Multi-Protocol Lab

Ziel: Das lokale Lab unterstützt weitere Streaming-Szenarien.

Akzeptanzkriterien:

- MediaMTX-Beispiel vorhanden.
- SRT-Beispiel vorhanden.
- DASH-Beispiel vorhanden.
- WebRTC-Beispiel vorbereitet.
- Beispiele sind dokumentiert.

---

### 13.6 Version 0.6.0: SRT Health View

Ziel: SRT-Contribution-Workflows technisch sichtbar machen.

Akzeptanzkriterien:

- SRT-Testsetup vorhanden.
- SRT-Verbindungsmetriken werden erfasst oder importiert.
- RTT, Packet Loss, Retransmissions und Bandbreite werden angezeigt.
- Dashboard enthält eine SRT-Health-Ansicht.
- Dokumentation erklärt typische SRT-Fehlerbilder.
- SRT-Metriken werden OTel-kompatibel modelliert.

---

## 14. Akzeptanzkriterien für das Gesamtprojekt

Das Projekt gilt in der ersten Phase als erfolgreich, wenn folgende Punkte erfüllt sind:

- Ein neuer Entwickler kann das Projekt unter Linux lokal starten.
- Die Startanleitung funktioniert ohne manuelle Sonderkonfiguration.
- Die Architektur ist klar nachvollziehbar.
- Die Domain-Schicht ist frameworkfrei.
- Die Adapter sind technisch klar getrennt.
- Mindestens ein Teststream kann abgespielt werden.
- Playback-Events werden vom Browser an die API gesendet.
- Events sind im Dashboard sichtbar.
- Basis-Metriken sind über Observability-Komponenten sichtbar oder vorbereitet.
- Das Repository ist Open-Source-tauglich dokumentiert.
- Die erste Version ist als GitHub-Release veröffentlichbar.

---

## 15. Risiken

### 15.1 Technische Risiken

| Risiko | Bewertung | Gegenmaßnahme |
|---|---|---|
| Projekt wird zu groß | Hoch | MVP strikt begrenzen |
| Streaming-Protokolle werden zu komplex | Mittel | HLS zuerst, andere später |
| Hexagonale Architektur wird übertrieben | Mittel | Nur dort einsetzen, wo Fachlogik existiert |
| Lokales Docker-Setup wird instabil | Mittel | einfache Defaults, klare Health Checks |
| Observability wird zu früh zu komplex | Mittel | erst minimale Metriken, später Ausbau |
| Browser-Verhalten unterscheidet sich stark | Hoch | MVP nur hls.js, weitere Adapter später |
| Prometheus-Cardinality explodiert | Hoch | keine Session-Labels, Traces für Per-Session-Daten |
| Player-SDK wird unterschätzt | Hoch | als eigenes Subprojekt mit Adapter-Schichten planen |
| WebRTC verwässert den MVP | Hoch | WebRTC aus Phase 1 entfernen |
| Datenschutz bremst Adoption | Mittel | Anonymisierung und sparsame Defaults früh vorsehen |
| Schema-Evolution bricht externe SDK-Versionen | Mittel | Schema-Versionierung, Contract-Tests und Kompatibilitätsprüfungen in CI |
| Project Token im Browser-Code wird zweckentfremdet | Mittel | niedrige Kritikalität, Origin-Pinning, Rate Limits und kurze Token-Rotation |

### 15.2 Projektbezogene Risiken

| Risiko | Bewertung | Gegenmaßnahme |
|---|---|---|
| Zu wenig sichtbarer Nutzen | Hoch | Demo-first Ansatz |
| README unklar | Hoch | Schnellstart prominent platzieren |
| Keine Contributor gewinnen | Mittel | gute Issues, Roadmap, klare Architektur |
| Zu viele unfertige Module | Mittel | Platzhalter reduzieren, Fokus auf lauffähigen Pfad |

---

## 16. Offene Punkte

Folgende Entscheidungen müssen noch früh getroffen werden, weil sie Architektur und Storage stark beeinflussen:

### 16.1 Zielgruppenentscheidung

Die wichtigste Produktentscheidung lautet:

```text
Selbsthoster und kleine Teams oder Plattform-Betreiber mit hunderten parallelen Streams?
```

Diese Entscheidung beeinflusst:

- Storage
- Sampling
- Cardinality
- Multi-Tenant-Fähigkeit
- Betriebsmodell
- Dashboard-Komplexität
- notwendige Alerting-Funktionen

Empfehlung für den MVP:

```text
Fokus auf Selbsthoster, kleine Plattformen, Broadcaster-Labs und technische Teams.
```

Große Plattform-Betreiber sollen erst später adressiert werden.

### 16.2 Weitere offene Entscheidungen

- Projektlizenz: MIT oder Apache-2.0
- Backend-Technologie final: Go oder Micronaut nach technischem Spike
- Datenhaltung im MVP: rein In-Memory oder SQLite/PostgreSQL
- Frontend-Styling: eigenes CSS, Tailwind oder UI-Library
- Live-Updates: Polling, WebSocket oder Server-Sent Events
- CI-Zielplattformen
- Release-Konvention
- Paketnamen für npm
- Go Module Name oder JVM Package Namespace final

---

## 17. Erste empfohlene Umsetzungsschritte

### Schritt 0: Backend-Technologie-Spike

Vor der eigentlichen MVP-Implementierung muss die Backend-Technologie entschieden werden.

Vorgehen:

- Branch `spike/go-api`
- Branch `spike/micronaut-api`
- identischer Mini-Scope in beiden Branches:
  - `POST /api/playback-events`
  - In-Memory Event Repository
  - strukturierte Logs
  - einfacher Health Check
  - OpenTelemetry Export oder vorbereiteter OTLP-Pfad
  - Dockerfile
- Entscheidung dokumentieren in `docs/adr/0001-backend-stack.md`

Akzeptanzkriterium:

Die Entscheidung wird nicht theoretisch getroffen, sondern anhand zweier lauffähiger Mini-Prototypen.

---

### Schritt 1: Repository initialisieren

- Mono-Repo-Struktur anlegen
- README.md erstellen
- CHANGELOG.md erstellen
- LICENSE hinzufügen
- Makefile hinzufügen
- `.env.example` hinzufügen

### Schritt 2: API-Grundgerüst

- Backend-App unter `apps/api`, in Go oder Micronaut nach technischem Spike
- Hexagon-Struktur anlegen
- Domain-Modelle für StreamSession und PlaybackEvent
- Use Case `RegisterPlaybackEventUseCase`
- In-Memory Repository
- HTTP Controller

### Schritt 3: Dashboard- und Demo-Player-Grundgerüst

- SvelteKit-App unter `apps/dashboard`
- Startseite
- Test-Player-Seite
- Stream-Sessions-Seite
- API Client
- Demo-Player-Route unter `apps/dashboard/src/routes/demo/`
- SDK-Referenzintegration innerhalb des Dashboards vorbereiten

### Schritt 4: Player-SDK

- TypeScript-Package unter `packages/player-sdk`
- HTMLVideoElement Adapter
- HTTP Event Publisher
- einfache Event-Erfassung

### Schritt 5: Docker Lab

- Docker Compose
- MediaMTX Service
- FFmpeg Teststream
- API Service
- Dashboard Service

### Schritt 6: Observability

- OTel Collector
- Prometheus
- Grafana
- erste Metriken
- Dokumentation

---

## 18. Definition of Done für den MVP

Der MVP ist fertig, wenn:

- `make dev` erfolgreich startet.
- Der Teststream lokal läuft.
- Das Dashboard im Browser erreichbar ist.
- Der Test-Player den Stream abspielen kann.
- Das Player-SDK Events erzeugt.
- Die API Events annimmt.
- Das Dashboard Events anzeigt.
- Die Architektur in `docs/architecture.md` beschrieben ist.
- Das Eventmodell in `docs/telemetry-model.md` beschrieben ist.
- Tests für zentrale Use Cases vorhanden sind.
- CI mindestens Build und Tests ausführt.
- CHANGELOG.md einen Eintrag für `0.1.0` enthält.

---

## 19. Glossar

| Begriff | Bedeutung |
|---|---|
| Adapter | Technische Implementierung eines Eingangs oder Ausgangs |
| DASH | MPEG-DASH, adaptives Streaming-Protokoll |
| Domain | Fachlicher Kern der Anwendung |
| HLS | HTTP Live Streaming |
| Hexagon | Architekturmodell mit Ports und Adapters |
| Inbound Adapter | Adapter, der die Anwendung von außen aufruft, z. B. HTTP Controller |
| Media Server | Server zur Annahme, Verarbeitung und Auslieferung von Streams |
| MediaMTX | Media Server mit Unterstützung für RTSP, RTMP, HLS, WebRTC und SRT |
| Mono-Repo | Repository, das mehrere Anwendungen und Pakete gemeinsam enthält |
| OpenTelemetry | Standard für Logs, Metriken und Traces |
| Outbound Adapter | Adapter, mit dem die Anwendung externe Systeme nutzt |
| Player-SDK | Browser-Bibliothek zur Erfassung von Playback-Metriken |
| Port | Schnittstelle zwischen Hexagon und Außenwelt |
| RTMP | Real-Time Messaging Protocol |
| SRS | Simple Realtime Server |
| SRT | Secure Reliable Transport |
| Stream Analyzer | Komponente zur Analyse von Streaming-Manifesten |
| Stream Session | zusammenhängende Betrachtung einer Wiedergabe- oder Streaming-Sitzung |
| Use Case | fachlicher Anwendungsfall |
| CMAF | Common Media Application Format, Container-/Segmentierungsstandard für adaptive Streaming-Workflows |
| LL-HLS | Low-Latency HLS, Variante von HLS für geringere Latenz |
| QoE | Quality of Experience, nutzerbezogene Qualitätswahrnehmung beim Playback |
| Cardinality | Anzahl unterschiedlicher Zeitreihen-Kombinationen durch Labels, besonders relevant für Prometheus |
| OTLP | OpenTelemetry Protocol für den Transport von Traces, Metriken und Logs |
| Time Skew | Abweichung zwischen Client-Uhr und Server-Uhr |

---

## 20. Zusammenfassung

m-trace soll als Open-Source-Mono-Repo ein praxisnahes Werkzeug für Media-Streaming-Observability werden.

Der entscheidende Erfolgsfaktor ist nicht maximale Funktionsbreite, sondern ein sofort nutzbarer lokaler Demo-Pfad:

```bash
git clone <repo>
cd m-trace
make dev
```

Danach soll ein Entwickler im Browser sehen können:

- ein laufender Teststream
- Player-Events
- Stream-Sessions
- erste Metriken
- technische Diagnoseinformationen

Die Architektur muss sauber genug sein, um langfristig wartbar zu bleiben, aber pragmatisch genug, damit der MVP schnell nutzbar wird.
