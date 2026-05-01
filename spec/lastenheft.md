# Lastenheft: m-trace

**Projektname:** m-trace<br>
**Dokumenttyp:** Lastenheft<br>
**Version:** 1.1.7<br>
**Status:** Verbindlich<br>
**Lizenz:** MIT<br>
**Architekturstil:** Mono-Repo mit hexagonaler Architektur<br>
**Primärer Stack:** Go 1.22 (stdlib `net/http`, Prometheus, OpenTelemetry, Distroless-Runtime), SvelteKit, TypeScript, Docker — Backend-Stack entschieden in `docs/adr/0001-backend-stack.md`.

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
- Backend-API in Go (siehe `docs/adr/0001-backend-stack.md`)
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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-1 | Muss | Das Repository muss alle Hauptbestandteile des Projekts enthalten. |
| F-2 | Muss | Anwendungen müssen unter `apps/` liegen. |
| F-3 | Muss | Wiederverwendbare Libraries müssen unter `packages/` liegen. |
| F-4 | Muss | Hilfsdienste müssen unter `services/` liegen. |
| F-5 | Muss | Beispiele müssen unter `examples/` liegen. |
| F-6 | Muss | Observability-Konfigurationen müssen unter `observability/` liegen. |
| F-7 | Muss | Deployment-Artefakte müssen unter `deploy/` liegen. |
| F-8 | Muss | Dokumentation muss unter `docs/` liegen. |
| F-9 | Muss | Skripte müssen unter `scripts/` liegen. |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-10 | Muss | Fachlogik muss im Ordner `hexagon/` liegen. |
| F-11 | Muss | Technische Ein- und Ausgänge müssen im Ordner `adapters/` liegen. |
| F-12 | Muss | Abhängigkeiten müssen von außen nach innen zeigen. |
| F-13 | Muss | Die Domain darf keine Framework-, HTTP-, Datenbank- oder Docker-Abhängigkeiten enthalten. |
| F-14 | Muss | Ports müssen als Schnittstellen definiert werden. |
| F-15 | Muss | Adapter müssen Ports implementieren oder Use Cases aufrufen. |
| F-16 | Muss | DTOs dürfen nicht Teil der Domain sein. |

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

Die API-Anwendung muss unter `apps/api` liegen. Backend-Technologie ist Go gemäß `docs/adr/0001-backend-stack.md`; Spec in §10.1.

#### Hauptaufgaben

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-17 | Muss | Annahme von Playback-Events |
| F-18 | Muss | Verwaltung von Stream-Sessions |
| F-19 | Muss | Bereitstellung von Metriken |
| F-20 | Muss | Weitergabe von Telemetrie an OpenTelemetry |
| F-21 | Muss | Bereitstellung von Daten für das Dashboard |
| F-22 | Muss | Architektur-Vorbereitung in `apps/api` für Stream Analyzer (Port-Hook); volle Integration ab Phase `0.3.0`. |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-23 | Muss | Anzeige laufender Stream-Sessions |
| F-24 | Muss | Anzeige aktueller Playback-Metriken |
| F-25 | Muss | Anzeige von Fehlern und Warnungen |
| F-26 | Muss | Anzeige einfacher Stream-Health-Zustände |
| F-27 | Muss | Anzeige von Backend- und Telemetrie-Status |
| F-28 | Muss | Integration eines Test-Players |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-29 | Muss | Playback-Events annehmen |
| F-30 | Muss | Stream-Sessions verwalten |
| F-31 | Muss | Metriken vorbereiten oder exportieren |
| F-32 | Muss | Daten für Dashboard bereitstellen |
| F-33 | Muss | Stream Analyzer anbinden |
| F-34 | Muss | Health Checks bereitstellen |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-35 | Muss | Live-Übersicht anzeigen |
| F-36 | Muss | Test-Player bereitstellen |
| F-37 | Muss | Playback-Events anzeigen |
| F-38 | Muss | Stream-Sessions anzeigen |
| F-39 | Muss | API-Status anzeigen |
| F-40 | Muss | Links zu Grafana, Prometheus und Media-Server-Konsole anzeigen |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-41 | Kann | HLS-Teststream abspielen |
| F-42 | Kann | Player-SDK isoliert integrieren |
| F-43 | Kann | erzeugte Events sichtbar machen |
| F-44 | Kann | SDK-Konfiguration testen |
| F-45 | Kann | als minimale Referenzintegration für externe Nutzer dienen |

Warum nicht im MVP:

Das Dashboard kann die Demo-Funktion zunächst ausreichend abdecken. Eine eigene App würde Build-, Deployment- und Dokumentationsaufwand erhöhen, ohne den ersten Nutzwert wesentlich zu steigern.


---

#### 7.5.4 `apps/ingest-gateway`

`apps/ingest-gateway` ist ein späterer Dienst zur Verwaltung von Ingest-Flows, Stream-Keys und Routing-Regeln.

Status im MVP: **Kann**

Hauptaufgaben:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-46 | Kann | Stream-Keys verwalten |
| F-47 | Kann | Ingest-Endpunkte beschreiben |
| F-48 | Kann | Routing-Regeln für Streams definieren |
| F-49 | Kann | Webhooks bei Stream-Start und Stream-Ende auslösen |
| F-50 | Kann | SRT-/RTMP-Konfigurationen vorbereiten |
| F-51 | Kann | Media-Server-Konfigurationen generieren oder validieren |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-52 | Kann | HLS-URL entgegennehmen |
| F-53 | Kann | Manifest analysieren |
| F-54 | Kann | Analyseergebnis als JSON liefern |
| F-55 | Kann | Fehler und Warnungen normalisieren |
| F-56 | Kann | spätere DASH-/CMAF-Analyse anbieten |
| F-57 | Kann | Sicherheitsgrenzen für externe URL-Abrufe schaffen |

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
| `apps/api` | zentrale Backend-API | Muss | Go (ADR-0001) |
| `apps/dashboard` | Web-Dashboard | Muss | SvelteKit |
| `apps/demo-player` | SDK-Referenz und Testplayer | Nicht MVP, zunächst `/demo`-Route | SvelteKit oder Vite |
| `apps/ingest-gateway` | Stream-Key, Ingest und Routing | Kann | Go (analog ADR-0001) |
| `apps/analyzer-api` | separater Analyse-Service | Kann | Go oder Node.js |
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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-58 | Kann | dash.js |
| F-59 | Kann | Shaka Player |
| F-60 | Kann | Video.js |
| F-61 | Kann | native Safari HLS |
| F-62 | Kann | WebRTC `getStats()`, separat in späterer Phase |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-63 | Muss | Anbindung an ein `HTMLVideoElement` |
| F-64 | Muss | Erfassung von Playback-Events |
| F-65 | Muss | Erfassung einfacher Metriken |
| F-66 | Muss | Versand der Events über OpenTelemetry Web SDK oder HTTP an die API |
| F-67 | Muss | Trennung von Browser-Adapter und fachlicher Tracking-Logik |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-68 | Muss | Abruf von HLS-Manifesten |
| F-69 | Muss | Analyse einfacher Manifest-Eigenschaften |
| F-70 | Muss | Prüfung von Segment-Dauern |
| F-71 | Muss | Erkennung offensichtlicher Inkonsistenzen |
| F-72 | Muss | Bereitstellung einer API für Backend und CLI |
| F-73 | Muss | Vorbereitung für DASH- und CMAF-Analyse |

#### Mindestfunktionen für den MVP

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-74 | Muss | HLS Master Playlist erkennen |
| F-75 | Muss | HLS Media Playlist erkennen |
| F-76 | Muss | Varianten und Renditions extrahieren |
| F-77 | Muss | Segment-Anzahl bestimmen |
| F-78 | Muss | durchschnittliche Segment-Dauer berechnen |
| F-79 | Muss | Abweichungen bei Segment-Dauern erkennen |
| F-80 | Muss | einfache Live-Latenz-Schätzung |
| F-81 | Muss | Analyseergebnis als JSON liefern |

#### CLI-Ziel

```bash
pnpm m-trace check https://example.com/live/master.m3u8
```

---

### 7.8 Lokales Streaming-Lab

Das Projekt muss eine lokale Streaming-Testumgebung bereitstellen.

#### Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-82 | Muss | Start per Docker Compose |
| F-83 | Muss | Media Server für lokale Tests |
| F-84 | Muss | FFmpeg-basierter Teststream |
| F-85 | Muss | API erreichbar unter `localhost` |
| F-86 | Muss | Dashboard erreichbar unter `localhost` |
| F-87 | Muss | Prometheus und Grafana optional verfügbar |
| F-88 | Muss | OpenTelemetry Collector optional verfügbar |

#### Mindestdienste

Die Dienste sind in zwei Klassen gegliedert (harmonisiert mit F-87/F-88
und MVP-28/MVP-29 in Patch `1.0.2`):

**Pflicht (Muss, im Default-Compose-Profil):**

| Dienst | Zweck |
|---|---|
| `api` | Backend-API |
| `dashboard` | SvelteKit UI |
| `mediamtx` | lokaler Media Server |
| `stream-generator` | FFmpeg-Teststream |

**Soll (optional, im `observability`-Compose-Profil):**

| Dienst | Zweck | Bezug |
|---|---|---|
| `otel-collector` | OpenTelemetry Collector | F-88 (optional verfügbar), MVP-29 |
| `prometheus` | Metrikspeicherung | F-87 (optional verfügbar) |
| `grafana` | Visualisierung | F-87 (optional verfügbar), MVP-28 |

> **Hinweis (Patch `1.1.1`)**: Die obigen Tabellen beschreiben den
> **End-Zustand der `0.1.x`-Phase**. Mit der Sub-Release-Schneidung
> aus Patch `1.1.0` (§13.1–§13.3) werden die Pflicht-Mindestdienste
> stufenweise aktiviert: `0.1.0` startet `api`, `mediamtx`,
> `stream-generator`; `0.1.1` ergänzt `dashboard`; `0.1.2` aktiviert
> die Soll-Tabelle über das `observability`-Compose-Profil. Der
> jeweilige Sub-Release-Pflicht-Subset steht im zugehörigen
> Plan-Dokument (`docs/planning/plan-0.1.x.md`).

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-89 | Muss | API muss strukturierte Logs erzeugen. |
| F-90 | Muss | API muss Health Checks bereitstellen. |
| F-91 | Muss | API soll OpenTelemetry unterstützen. |
| F-92 | Muss | Playback-Events sollen als Metriken oder Traces exportierbar sein. |
| F-93 | Muss | Prometheus soll technische Metriken erfassen können. |
| F-94 | Soll | Grafana kann mit einem einfachen Beispiel-Dashboard ausgeliefert werden (harmonisiert mit MVP-28). |

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
| `mtrace_invalid_events_total` | Anzahl wegen Schema-/Validierungsfehlern (`400`/`422`) abgelehnter Events; Auth-Fehler (`401`) zählen nicht (harmonisiert mit API-Kontrakt §7 in Patch `1.1.2`). |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-95 | Muss | Prometheus darf nur für aggregierte Metriken verwendet werden. |
| F-96 | Muss | `session_id` darf nicht als Prometheus-Label verwendet werden. |
| F-97 | Muss | Per-Session-Daten sollen als Traces oder Events modelliert werden. |
| F-98 | Muss | Für hochvolumige Eventdaten muss eine spätere Storage-Option vorgesehen werden. |
| F-99 | Muss | Das System muss Sampling vorbereiten. |
| F-100 | Muss | Das Telemetrie-Modell muss Datenschutz und Cardinality gemeinsam berücksichtigen. |

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

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-101 | Muss | Prometheus nur für Aggregate |
| F-102 | Muss | Player-Sessions als OpenTelemetry-Traces vorbereiten |
| F-103 | Muss | In-Memory-Speicherung nur für lokale Demo |
| F-104 | Muss | keine produktive Langzeitspeicherung im MVP |
| F-105 | Muss | keine `session_id`-Labels in Prometheus |

---

### 7.11 Telemetry Ingest, Event-Schema und SDK-Budget

Die Telemetrie-Schnittstelle ist ein Kernbestandteil des Projekts und muss früh spezifiziert werden.

#### Authentifizierung von Player-Events

Das Browser-SDK darf nicht dauerhaft gegen einen vollständig offenen Ingest-Endpunkt senden.

MVP-Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-106 | Muss | Events enthalten eine `project_id`. |
| F-107 | Muss | Events werden mit einem öffentlichen Project Token oder einem kurzlebigen Ingest Token versehen. |
| F-108 | Muss | Das Backend validiert erlaubte Origins. |
| F-109 | Muss | Tokens dürfen keine Secrets mit hoher Kritikalität sein, da Browser-Code öffentlich ist. |
| F-110 | Muss | Rate Limits gelten pro Project, Origin und IP-Bereich. |

Spätere Erweiterungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-111 | Kann | serverseitig signierte Session Tokens |
| F-112 | Kann | rotierbare Project Tokens |
| F-113 | Kann | tenant-spezifische Ingest Policies |

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
    "name": "@npm9912/player-sdk",
    "version": "0.2.0"
  }
}
```

Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-114 | Muss | neue Felder müssen abwärtskompatibel sein |
| F-115 | Muss | unbekannte Felder dürfen nicht zum Fehler führen |
| F-116 | Muss | entfernte Felder müssen über mindestens eine Minor-Version toleriert werden |
| F-117 | Muss | Breaking Changes erfordern neue Major-Version der Event-Schemas |

#### Backpressure und Rate Limiting

Die Ingest-API muss Überlastung kontrolliert behandeln.

MVP-Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-118 | Muss | maximale Event-Batch-Größe definieren |
| F-119 | Muss | maximale Request-Rate pro Project definieren |
| F-120 | Muss | HTTP `429` bei Rate Limit |
| F-121 | Muss | HTTP `202` für angenommene Events |
| F-122 | Muss | Events dürfen bei lokaler Überlast verworfen werden, wenn dies als Dropped-Event-Metrik sichtbar wird |
| F-123 | Muss | SDK muss Sampling und Batch-Größe konfigurieren können |

#### Zeitstempel und Time Skew

Browser-Clocks sind unzuverlässig. Das Backend muss daher zwischen Client-Zeit und Server-Zeit unterscheiden.

Pflichtfelder:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-124 | Muss | `client_timestamp` |
| F-125 | Muss | `server_received_at` |
| F-126 | Muss | optional `client_time_origin` |
| F-127 | Muss | optional `sequence_number` |

Regeln:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| F-128 | Muss | Ordering innerhalb einer Session bevorzugt über `sequence_number` |
| F-129 | Muss | Latenzberechnungen niemals blind nur aus Client-Zeit ableiten |
| F-130 | Muss | Backend muss auffälligen Time Skew markieren können |

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
| `spec/architecture.md` | Architekturüberblick |
| `docs/user/local-development.md` | lokale Entwicklung |
| `spec/telemetry-model.md` | Telemetrie- und Eventmodell |
| `spec/player-sdk.md` | Player-SDK-Nutzung |
| `docs/stream-analyzer.md` | Stream Analyzer |
| `docs/planning/roadmap.md` | geplante Entwicklung |

---

## 8. Nichtfunktionale Anforderungen

### 8.1 Plattform

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-1 | Muss | Entwicklung muss unter Linux möglich sein. |
| NF-2 | Muss | Entwicklung muss mit VS Code kompatibel sein. |
| NF-3 | Muss | Lokaler Betrieb muss über Docker möglich sein. |
| NF-4 | Muss | Build-Prozesse müssen ohne proprietäre Dienste funktionieren. |

### 8.2 Wartbarkeit

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-5 | Muss | Fachlogik muss testbar sein, ohne externe Infrastruktur zu starten. |
| NF-6 | Muss | Domain-Klassen dürfen keine Framework-Abhängigkeiten enthalten. |
| NF-7 | Muss | Ports müssen klar benannt und dokumentiert sein. |
| NF-8 | Muss | Adapter müssen austauschbar sein. |
| NF-9 | Muss | Technische Implementierungen dürfen nicht in die Domain-Schicht lecken. |

### 8.3 Erweiterbarkeit

Das Projekt muss vorbereitet sein für spätere Erweiterungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-10 | Muss | MediaMTX-Adapter |
| NF-11 | Muss | SRT-Ingest-Metriken |
| NF-12 | Muss | DASH-Analyse |
| NF-13 | Muss | CMAF-Analyse |
| NF-14 | Muss | WebRTC-Metriken |
| NF-15 | Muss | Datenbankpersistenz |
| NF-16 | Muss | Authentifizierung |
| NF-17 | Muss | Multi-Stream-Betrieb |
| NF-18 | Muss | Kubernetes Deployment |
| NF-19 | Muss | CI-basierte Stream-Checks |

### 8.4 Performance

Für den MVP gelten einfache Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-20 | Muss | API muss lokale Demo-Last problemlos verarbeiten. |
| NF-21 | Muss | Player-SDK darf Playback nicht merklich beeinflussen. |
| NF-22 | Muss | Dashboard muss bei mehreren aktiven Sessions bedienbar bleiben. |
| NF-23 | Muss | Event-Erfassung muss asynchron oder leichtgewichtig erfolgen. |

### 8.5 Sicherheit

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-24 | Muss | Keine Secrets im Repository. |
| NF-25 | Muss | `.env.example` muss Beispielwerte enthalten. |
| NF-26 | Muss | Produktive Secrets müssen über Umgebungsvariablen gesetzt werden. |
| NF-27 | Muss | CORS muss im lokalen Setup kontrolliert konfiguriert sein. |
| NF-28 | Muss | Externe URLs für Stream-Analyse müssen später abgesichert werden, um SSRF-Risiken zu vermeiden. |
| NF-29 | Muss | Security-Meldungen müssen über `SECURITY.md` beschrieben werden. |

#### CORS- und CSP-Grundregeln für Player-Telemetrie

Für Browser-SDK-Telemetrie muss Cross-Origin-Kommunikation kontrolliert werden.

MVP-Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-30 | Muss | erlaubte Origins werden pro Project konfiguriert |
| NF-31 | Muss | SDK-Requests nutzen standardmäßig `credentials: "omit"` |
| NF-32 | Muss | keine Cookies für Player-Telemetrie im MVP |
| NF-33 | Muss | Preflight-fähige CORS-Konfiguration |
| NF-34 | Muss | `Access-Control-Allow-Origin` darf nicht pauschal `*` sein, sobald Project Tokens genutzt werden |
| NF-35 | Muss | erlaubte Methoden zunächst auf `POST` und `OPTIONS` begrenzen |
| NF-36 | Muss | erlaubte Header explizit definieren, z. B. `Content-Type`, `X-MTrace-Project`, `X-MTrace-Token` |
| NF-37 | Muss | CSP-Beispiele für `connect-src` müssen dokumentiert werden |

Beispiel-CSP für eine Demo-Integration:

```text
Content-Security-Policy: connect-src 'self' https://m-trace.example.com;
```


### 8.6 Datenschutz und GDPR

Player-Telemetrie kann personenbezogene oder personenbeziehbare Daten enthalten. Dazu gehören insbesondere IP-Adressen, User-Agents, Session-IDs und grobe Standortinformationen.

Anforderungen:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-38 | Muss | IP-Adressen dürfen im MVP nicht unnötig gespeichert werden. |
| NF-39 | Muss | User-Agent-Daten müssen reduzierbar oder anonymisierbar sein. |
| NF-40 | Muss | Session-IDs müssen pseudonym sein. |
| NF-41 | Muss | Ein konfigurierbarer Anonymisierungs-Layer im Collector soll vorbereitet werden. |
| NF-42 | Muss | Das Projekt muss dokumentieren, welche Telemetriedaten erhoben werden. |
| NF-43 | Muss | Datenschutzfreundliche Defaults haben Vorrang vor maximaler Analyse-Tiefe. |
| NF-44 | Muss | Für EU-Nutzung muss eine GDPR-freundliche Betriebsweise möglich sein. |

### 8.7 Qualität

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| NF-45 | Muss | Automatisierte Tests für Domain- und Application-Schicht |
| NF-46 | Muss | Linting für TypeScript |
| NF-47 | Muss | Tests für zentrale Backend-Use-Cases |
| NF-48 | Muss | CI-Pipeline für Build und Test |
| NF-49 | Muss | klare Commit- und Release-Konventionen |
| NF-50 | Muss | CHANGELOG-Pflege ab dem ersten Release |

---

## 9. Technologie-Strategie und Architekturentscheidungen

Streaming-Observability-relevante Komponenten und Communities sind stark durch Go, Rust und TypeScript geprägt:

- Media-Server und Streaming-Infrastruktur häufig in Go
- OpenTelemetry Collector in Go
- Browser- und Player-Ökosystem stark in TypeScript
- performante Analyzer- und CLI-Werkzeuge häufig in Go oder Rust

### 9.1 Backend-Entscheidung

**Entschieden: Go.** Die Wahl ist in `docs/adr/0001-backend-stack.md` (Status: Accepted) festgehalten und beruht auf zwei Mini-Prototypen mit identischem Muss-Scope (`spec/backend-api-contract.md`); das Spike-Protokoll liegt in `docs/spike/backend-stack-results.md`.

Historischer Tradeoff (Stand vor dem Spike):

| Option | Vorteil | Nachteil |
|---|---|---|
| **Go** ✅ | passt kulturell gut zu OTel, MediaMTX und Infrastruktur-Tools | — |
| JVM (Micronaut) | vertrauter JVM-Stack, gute DI, gute Testbarkeit | kleinerer Contributor-Pool im Streaming-OSS-Umfeld |

Konkrete Stack-Spezifikation in §10.1.

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

Backend-Technologie: **Go**, entschieden in `docs/adr/0001-backend-stack.md`.

| Bereich | Festlegung |
|---|---|
| Sprache | Go 1.22 oder höher |
| HTTP | Standard-Library `net/http` |
| Metriken | `prometheus/client_golang` |
| Tracing | `go.opentelemetry.io/otel` |
| Logging | `log/slog`, JSON-Formatter |
| Build/Runtime | Distroless-static (`gcr.io/distroless/static-debian12:nonroot`) |
| Linting | `golangci-lint` mit Default-Lintern (`govet`, `errcheck`, `staticcheck`, `unused`, `ineffassign`) |
| Tests | `testing` + `httptest`, keine externen Frameworks |
| Workflow | Docker-only (`docker build --target {test,lint,build,runtime}`); lokales Go optional |
| Modulpfad | `github.com/pt9912/m-trace/apps/api` |

Mindestanforderungen an die Implementierung:

- HTTP API für Event-Ingest gemäß `spec/backend-api-contract.md`
- Health Check
- strukturierte Logs (`slog`)
- OpenTelemetry-kompatibles Eventmodell
- klare Trennung von Domain, Application und Adapters (Hexagon-Layout `hexagon/{domain,application,port/{driving,driven}}`, `adapters/{driving,driven}/...`)
- Containerisierung per Docker

Multi-Modul-Aufteilung über `go.work` ist nicht im MVP erforderlich; erst on demand bei wachsender Codebase (siehe `docs/planning/roadmap.md` §4 Folge-ADR).


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

Muss-Anforderungen für die `0.1.x`-Phase (Gesamt-MVP). Die einzelnen
Items werden mit Patch `1.1.0` auf die drei Sub-Releases verteilt
(siehe §13.1–§13.3 für die RAK-Zuordnung):

- `0.1.0` Backend Core + Demo-Lab — Pflicht: MVP-1, MVP-2, MVP-7
  (Compose Core), MVP-8, MVP-9, MVP-11, MVP-16, MVP-17, MVP-18.
  Nicht in `0.1.0`: MVP-3 (Dashboard), MVP-4 (Demo-Route), MVP-5
  (Player-SDK), MVP-12..MVP-14 (Dashboard-Anzeigen), MVP-10/MVP-15
  (Observability).
- `0.1.1` Player-SDK + Dashboard — kommen MVP-3, MVP-4, MVP-5,
  MVP-6, MVP-12, MVP-13, MVP-14 hinzu.
- `0.1.2` Observability-Stack — kommen MVP-10, MVP-15
  (Pflicht-Anteile, siehe §7.9 Hinweis aus Patch `1.1.1`/`1.1.2`)
  hinzu; Soll-Anteile MVP-28, MVP-29 als observability-Profil.

Muss-Anforderungen für die Gesamt-`0.1.x`-Phase:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-1 | Muss | Mono-Repo-Struktur |
| MVP-2 | Muss | eine Backend-App unter `apps/api` |
| MVP-3 | Muss | eine Web-App unter `apps/dashboard` |
| MVP-4 | Muss | Demo-Player als `/demo`-Route im Dashboard, nicht als separate App |
| MVP-5 | Muss | `packages/player-sdk` mit hls.js-Adapter |
| MVP-6 | Muss | pragmatische SDK-Struktur ohne vollständige Hexagon-Ceremony |
| MVP-7 | Muss | Docker Compose Setup |
| MVP-8 | Muss | MediaMTX als erster Media Server |
| MVP-9 | Muss | FFmpeg-Teststream |
| MVP-10 | Muss | OpenTelemetry-kompatibles Eventmodell |
| MVP-11 | Muss | API-Endpunkt für Playback-Event-Batches |
| MVP-12 | Muss | einfache Session-Liste |
| MVP-13 | Muss | einfache Event-Anzeige |
| MVP-14 | Muss | einfache eingebaute Session-/Trace-Ansicht im Dashboard |
| MVP-15 | Muss | Prometheus nur für aggregierte Metriken |
| MVP-16 | Muss | lokale Speicherung per In-Memory oder SQLite |
| MVP-17 | Muss | README mit Schnellstart |
| MVP-18 | Muss | CHANGELOG mit initialem Eintrag |

Nicht im `0.1.0`-MVP:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-19 | Muss | separate `apps/demo-player` |
| MVP-20 | Muss | separate `apps/analyzer-api` |
| MVP-21 | Muss | `packages/stream-analyzer` als fertiges Paket |
| MVP-22 | Muss | Tempo als Pflichtkomponente |
| MVP-23 | Muss | Mimir oder ClickHouse |
| MVP-24 | Muss | WebRTC |
| MVP-25 | Muss | SRT-Health-View |
| MVP-26 | Muss | Multi-Tenant-Betrieb |


### 12.2 Soll-Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-27 | Soll | SQLite-Persistenz statt reinem In-Memory |
| MVP-28 | Soll | Grafana-Dashboard für Aggregate |
| MVP-29 | Soll | einfache OTel-Collector-Konfiguration |
| MVP-30 | Soll | rudimentäre HLS-Manifest-Prüfung als interner Spike |
| MVP-31 | Soll | WebSocket oder SSE für Live-Updates |
| MVP-32 | Soll | CI mit GitHub Actions |


### 12.3 Kann-Anforderungen

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| MVP-33 | Kann | eigenständiger Stream Analyzer als `packages/stream-analyzer` |
| MVP-34 | Kann | CLI für Stream Analyzer |
| MVP-35 | Kann | Tempo-Integration |
| MVP-36 | Kann | SRS-Beispiel |
| MVP-37 | Kann | DASH-Analyse |
| MVP-38 | Kann | SRT-Ingest-Beispiel |
| MVP-39 | Kann | SRT-Health-View |
| MVP-40 | Kann | Persistenz mit PostgreSQL |
| MVP-41 | Kann | ClickHouse- oder VictoriaMetrics-Anbindung |
| MVP-42 | Kann | Kubernetes-Manifeste |
| MVP-43 | Kann | Devcontainer |
| MVP-44 | Kann | Release-Automatisierung |


---

## 13. Release-Plan

Die `0.1.x`-Phase ist in drei Sub-Releases geschnitten (Patch `1.1.0`),
damit jeder Schritt einen demonstrierbaren Eigenwert hat und der
Gesamt-MVP-Scope nicht in einem einzelnen Release-Cycle landet:

- `0.1.0` Backend Core + Demo-Lab — curl-driven End-to-End.
- `0.1.1` Player-SDK + Dashboard — Browser-getrieben, UI sichtbar.
- `0.1.2` Observability-Stack — Aggregat-Metriken und Traces.

`0.2.0` schließt mit dem publizierbaren Player-SDK an wie ursprünglich.

### 13.1 Version 0.1.0: Backend Core + Demo-Lab

Ziel: Ein Entwickler kann das Repository klonen und ein lokales Lab mit MediaMTX-basiertem Teststream und Backend-API starten; Player-Events werden per `curl` (oder anderem HTTP-Client) an die API gesendet, Sessions sind über die API abfragbar.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-1 | Muss | `make dev` startet die in `0.1.0` erforderlichen Pflicht-Dienste (`api`, `mediamtx`, `stream-generator`). |
| RAK-3 | Muss | API ist erreichbar (`/api/health` liefert `200`, drei Pflicht-Endpoints aus dem Spike plus die zwei Stream-Sessions-Endpoints). |
| RAK-4 | Muss | Teststream läuft über MediaMTX. |
| RAK-6 | Muss | API nimmt Events an (`POST /api/playback-events` mit gültigem Token). |
| RAK-8 | Muss | README/Local-Development-Doku beschreibt den `0.1.0`-Quickstart reproduzierbar. |

### 13.2 Version 0.1.1: Player-SDK + Dashboard

Ziel: Browser-Player auf der Demo-Route sendet echte Playback-Events an die API; Dashboard zeigt Sessions, Events und Status.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-2 | Muss | Dashboard ist erreichbar; `make dev` startet zusätzlich den `dashboard`-Service. |
| RAK-5 | Muss | Player-SDK sendet hls.js-basierte Events. |
| RAK-7 | Muss | Dashboard zeigt empfangene Events und einfache Session-Zusammenhänge. |

### 13.3 Version 0.1.2: Observability-Stack

Ziel: Optionales `observability`-Compose-Profil bringt Prometheus, Grafana und OTel-Collector additiv zum Core-Stack; Aggregat-Metriken und Traces sind sichtbar.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-9 | Muss | Prometheus enthält nur aggregierte Metriken (Cardinality-Regeln aus §7.10 eingehalten). |
| RAK-10 | Soll | Player-Session-Traces sind vorbereitet oder exemplarisch sichtbar. |

RAK-8 wird in `0.1.2` ergänzt um die Doku zum `observability`-Profil.

---

### 13.4 Version 0.2.0: Publizierbares Player SDK

Ziel: Das Player-SDK wird vom MVP-Prototyp zu einem eigenständig nutzbaren und dokumentierten npm-Paket ausgebaut.

Abgrenzung zu `0.1.0`:

`0.1.0` beweist den End-to-End-Pfad mit hls.js-Adapter und Event-Ingest.  
`0.2.0` stabilisiert das SDK als wiederverwendbares Paket mit Public API, Tests, Dokumentation und Versionierungsstrategie.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-11 | Muss | SDK ist als npm-Paket baubar und lokal installierbar. |
| RAK-12 | Muss | Public API ist dokumentiert. |
| RAK-13 | Muss | Event-Schema ist versioniert. |
| RAK-14 | Muss | hls.js-Adapter ist getestet. |
| RAK-15 | Muss | HTTP-Transport ist getestet. |
| RAK-16 | Soll | OTel-Transport ist vorbereitet oder experimentell nutzbar. |
| RAK-17 | Muss | SDK unterstützt Batching, Sampling und Retry-Grenzen. |
| RAK-18 | Muss | SDK hält das definierte Performance-Budget ein. |
| RAK-19 | Muss | Browser-Support-Matrix ist dokumentiert. |
| RAK-20 | Muss | Beispielintegration in der Dashboard-Route `/demo` ist dokumentiert. |
| RAK-21 | Muss | Kompatibilität zwischen SDK-Version und Event-Schema wird in CI geprüft. |


---

### 13.5 Version 0.3.0: Stream Analyzer

Ziel: HLS-Streams können analysiert werden.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-22 | Muss | HLS Manifest kann geladen werden. |
| RAK-23 | Muss | Master Playlist kann erkannt werden. |
| RAK-24 | Muss | Media Playlist kann erkannt werden. |
| RAK-25 | Muss | Segment-Dauern werden geprüft. |
| RAK-26 | Muss | Ergebnis wird als JSON ausgegeben. |
| RAK-27 | Muss | API kann Analyzer nutzen. |
| RAK-28 | Muss | CLI-Grundlage existiert. |

---

### 13.6 Version 0.4.0: Erweiterte Trace-Korrelation

Ziel: Die in `0.1.0` vorbereitete OTel-Grundlage wird zu einer nutzbaren Korrelationsschicht ausgebaut.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-29 | Muss | Player-Session-Traces werden konsistent erzeugt. |
| RAK-30 | Soll | Manifest-Requests, Segment-Requests und Player-Events werden in einem Trace zusammengeführt, soweit technisch möglich. |
| RAK-31 | Kann | Tempo kann optional als Trace-Backend verwendet werden. |
| RAK-32 | Muss | Dashboard kann Session-Verläufe auch ohne Tempo einfach anzeigen. |
| RAK-33 | Muss | Prometheus bleibt auf aggregierte Metriken beschränkt. |
| RAK-34 | Muss | Dropped-, Rate-Limited- und Invalid-Event-Metriken sind sichtbar. |
| RAK-35 | Muss | Dokumentation beschreibt Cardinality-Grenzen und Sampling-Strategie. |


---

### 13.7 Version 0.5.0: Multi-Protocol Lab

Ziel: Das lokale Lab unterstützt weitere Streaming-Szenarien.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-36 | Muss | MediaMTX-Beispiel vorhanden. |
| RAK-37 | Muss | SRT-Beispiel vorhanden. |
| RAK-38 | Muss | DASH-Beispiel vorhanden. |
| RAK-39 | Soll | WebRTC-Beispiel vorbereitet. |
| RAK-40 | Muss | Beispiele sind dokumentiert. |

---

### 13.8 Version 0.6.0: SRT Health View

Ziel: SRT-Contribution-Workflows technisch sichtbar machen.

Akzeptanzkriterien:

| Kennung | Prioritaet | Akzeptanzkriterium |
|---|---|---|
| RAK-41 | Muss | SRT-Testsetup vorhanden. |
| RAK-42 | Muss | SRT-Verbindungsmetriken werden erfasst oder importiert. |
| RAK-43 | Muss | RTT, Packet Loss, Retransmissions und Bandbreite werden angezeigt. |
| RAK-44 | Muss | Dashboard enthält eine SRT-Health-Ansicht. |
| RAK-45 | Muss | Dokumentation erklärt typische SRT-Fehlerbilder. |
| RAK-46 | Muss | SRT-Metriken werden OTel-kompatibel modelliert. |

---

## 14. Akzeptanzkriterien für das Gesamtprojekt

Das Projekt gilt in der ersten Phase als erfolgreich, wenn folgende Punkte erfüllt sind:

| Kennung | Prioritaet | Anforderung |
|---|---|---|
| AK-1 | Muss | Ein neuer Entwickler kann das Projekt unter Linux lokal starten. |
| AK-2 | Muss | Die Startanleitung funktioniert ohne manuelle Sonderkonfiguration. |
| AK-3 | Muss | Die Architektur ist klar nachvollziehbar. |
| AK-4 | Muss | Die Domain-Schicht ist frameworkfrei. |
| AK-5 | Muss | Die Adapter sind technisch klar getrennt. |
| AK-6 | Muss | Mindestens ein Teststream kann abgespielt werden. |
| AK-7 | Muss | Playback-Events werden vom Browser an die API gesendet. |
| AK-8 | Muss | Events sind im Dashboard sichtbar. |
| AK-9 | Muss | Basis-Metriken sind über Observability-Komponenten sichtbar oder vorbereitet. |
| AK-10 | Muss | Das Repository ist Open-Source-tauglich dokumentiert. |
| AK-11 | Muss | Die erste Version ist als GitHub-Release veröffentlichbar. |

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

| Kennung | Status | Entscheidung |
|---|---|---|
| OE-1 | resolved | Projektlizenz: **MIT** (siehe `LICENSE`) |
| OE-2 | resolved | Backend-Technologie final: **Go** (siehe `docs/adr/0001-backend-stack.md`) |
| OE-3 | offen | Datenhaltung im MVP: rein In-Memory oder SQLite/PostgreSQL |
| OE-4 | resolved | Frontend-Styling im MVP: **eigenes CSS ohne Tailwind/UI-Library** |
| OE-5 | offen | Live-Updates: Polling, WebSocket oder Server-Sent Events |
| OE-6 | resolved | CI-Zielplattform: **GitHub Actions `ubuntu-24.04`** |
| OE-7 | resolved | Release-Konvention: **trunk-based auf `main`, annotierte SemVer-Tags `vX.Y.Z`, GitHub Release aus `CHANGELOG.md`** |
| OE-8 | resolved | npm-Paketname Player-SDK: **`@npm9912/player-sdk`** ab `0.2.0`; Lieferstand `0.1.x` wurde unter `@m-trace/player-sdk` ausgeliefert (Lastenheft `1.1.5`, `bae4a2a`), aber nie öffentlich publishet. |
| OE-9 | resolved | Go Module Name final: **`github.com/pt9912/m-trace/apps/api`** |

---

## 17. Erste empfohlene Umsetzungsschritte

### Schritt 0: Backend-Technologie-Spike — abgeschlossen

Backend-Technologie wurde durch zwei lauffähige Mini-Prototypen (Go,
Micronaut) im identischen Muss-Scope entschieden. Dokumentation:

- Spike-Spezifikation: `docs/spike/0001-backend-stack.md`
- Implementierungsplan: `docs/planning/plan-spike.md`
- API-Kontrakt: `spec/backend-api-contract.md`
- Spike-Protokoll: `docs/spike/backend-stack-results.md`
- Entscheidung: `docs/adr/0001-backend-stack.md` (Status: Accepted) — **Go**

Sieger-Branch `spike/go-api` ist auf `main` als `apps/api` integriert
(siehe `docs/planning/roadmap.md` §1).

---

### Schritt 1: Repository initialisieren

- Mono-Repo-Struktur anlegen
- README.md erstellen
- CHANGELOG.md erstellen
- LICENSE hinzufügen
- Makefile hinzufügen
- `.env.example` hinzufügen

### Schritt 2: API-Grundgerüst

- Backend-App unter `apps/api` in Go (siehe `docs/adr/0001-backend-stack.md`)
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
- Die Architektur in `spec/architecture.md` beschrieben ist.
- Das Eventmodell in `spec/telemetry-model.md` beschrieben ist.
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
