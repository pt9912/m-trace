# Implementation Plan — `0.1.0` (OTel-native Local Demo)

> **Status**: In Arbeit. Pre-MVP-Vorbereitung (Tranche 0) abgeschlossen, Architektur-Skelett-Doku (Tranche 0a) und Spike-Code-Korrekturen (Tranche 0b) teilweise umgesetzt.  
> **Bezug**: [Lastenheft `1.0.0`](./lastenheft.md) §13.1 (RAK-1..RAK-10), §18 (MVP-DoD); [Roadmap](./roadmap.md) §1.2, §2, §3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert — Commit-Hash genannt; das Item ist im Code/in der Doku enthalten.
- `[ ]` offen — kein Commit, kein Code dahinter.
- 🟡 in Arbeit — partiell umgesetzt mit weiteren offenen Sub-Items.

Architektur-Soll steht in [`architecture.md`](./architecture.md) und enthält **kein** Status-Tracking. Differenzen Code↔Soll werden hier als offene `[ ]`-DoD-Items getrackt.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Pre-MVP-Vorbereitung — Spike-Sieger auf `main`, Lastenheft `1.0.0`, README/Roadmap, Risiken-Backlog | ✅ |
| 0a | Architektur- und Plan-Doku — `architecture.md`, `releasing.md`, `plan-0.1.0.md`, `telemetry-model.md`, `local-development.md` | 🟡 |
| 0b | Spike-Code-Korrekturen aus Code-Reviews — Auth-vor-Body, InvalidEvents-Scope, OTel-Counter, Step-Numbering | 🟡 |
| 1 | MVP-Implementierung — Dashboard, Player-SDK, Docker-Compose-Lab, Observability-Stack | ⬜ |

---

## 2. Tranche 0 — Pre-MVP-Vorbereitung

Roadmap §1.2 Schritte 1–4. Ausgeliefert in Commits `f2f3e44`, `e073040`, `09b2e23`, `486bd08`, `0881c23`, `7dc5d92`, `6b75fe1`, `08811cb`, `f36bbc0`, `953c678`.

### 2.1 Schritt 1 — `spike/go-api` → `apps/api` auf `main`

DoD:

- [x] Spike-Sieger-Branch `spike/go-api` als `--no-ff`-Merge auf `main` integriert (`f2f3e44`).
- [x] Modulpfad-Rename `github.com/example/m-trace/apps/api` → `github.com/pt9912/m-trace/apps/api` (16 Dateien, `e073040`).
- [x] Docker-Targets `test`, `lint`, `build` lokal grün verifiziert nach Modulpfad-Rename (`e073040`).
- [x] `apps/api/README.md` post-Spike aktualisiert (Titel, Adapter-Tree, Pflicht-Endpoints) (`09b2e23`, korrigiert in `486bd08`).
- [x] Roadmap §1.2 und §2 Schritt 1 auf ✅ (`09b2e23`).

### 2.2 Schritt 2 — Lastenheft `1.0.0`

DoD:

- [x] Header: Version `0.7.0` → `1.0.0`, Status „Entwurf" → „Verbindlich" (`0881c23`).
- [x] Primärer-Stack-Zeile auf Go 1.22 + stdlib + Prometheus + OTel + Distroless konkretisiert (`0881c23`).
- [x] §6, §7.3, §7.5.7, §17 Schritt 2: „Go oder Micronaut nach technischem Spike" → „Go (ADR-0001)" (`0881c23`).
- [x] §9.1 Backend-Entscheidung: aus Offenhaltung wird Festlegung mit Sieger-Markierung (`0881c23`).
- [x] §10.1 Backend: konkrete Stack-Spezifikation (Sprache, HTTP, Metriken, Tracing, Logging, Build/Runtime, Linting, Tests, Workflow, Modulpfad) (`0881c23`).
- [x] §16.2: OE-2 und OE-9 als `resolved` markiert mit Auflösung (`0881c23`).
- [x] §17 Schritt 0 (Backend-Spike) als „abgeschlossen" mit Verweisen auf ADR und Spike-Doku (`0881c23`).
- [x] Roadmap §5 nach §7-Wartungsregel bereinigt (Repo-Hosting, OE-2, OE-9 entfernt) (`0881c23`).
- [x] Roadmap §1.2 und §2 Schritt 2 auf ✅ (`0881c23`).
- [x] OE-Verweise in Roadmap §1.2/§2 für Schritte 1+2 ergänzt (`7dc5d92`).
- [x] Verweis-Spalte in Roadmap §1.2/§2 auf reine ID-Form (kein „§"-Zeichen) (`6b75fe1`).

### 2.3 Schritt 3 — `README.md` Tech-Overview

DoD:

- [x] Status-Quote: „frühe Planungs-/Architekturphase" → „Pre-MVP `0.1.0`" (`08811cb`).
- [x] §Architekturprinzipien › Backend: konkrete Go-1.22-Spezifikation (stdlib `net/http`, Prometheus, OTel, `slog`, Distroless) (`08811cb`).
- [x] Hexagon-Tree korrekt mit `port/{driving,driven}` und allen fünf driven-Adaptern (auth, metrics, persistence, ratelimit, telemetry) (`08811cb`).
- [x] §Backend-Technologie-Spike auf Vergangenheitsmodus, Pointer auf ADR + Spike-Doku (`08811cb`).
- [x] §Aktueller Stand auf Pre-MVP `0.1.0`; Doku-Liste um `roadmap.md` und `adr/0001-backend-stack.md` ergänzt (`08811cb`).
- [x] Repo-URL im Clone-Beispiel auf `github.com/pt9912/m-trace` (`08811cb`).
- [x] Roadmap §1.2 und §2 Schritt 3 auf ✅ (`08811cb`).

### 2.4 Schritt 4 — Risiken-Backlog

DoD:

- [x] Issue-Backlog-Form entschieden: Markdown-Datei `docs/risks-backlog.md` analog cmake-xray/d-migrate-Pattern (`953c678`).
- [x] R-1 Hexagon-Boundaries (Disziplin-basiert, kein Compile-Time-Enforcement) eingetragen mit Verweis auf Folge-ADR „`apps/api` Multi-Modul-Aufteilung" (`953c678`).
- [x] R-2 CGO/SRT bricht distroless-static eingetragen mit Verweis auf Folge-ADR „SRT-Binding-Stack" (`953c678`).
- [x] R-3 Go-WebSocket-Ökosystem fragmentiert eingetragen mit Verweis auf Folge-ADR „WebSocket vs. SSE" (`953c678`).
- [x] Roadmap §1.2 und §2 Schritt 4 auf ✅ (`953c678`).
- [x] Roadmap §5 „Issue-Backlog-Form" entfernt (resolved) (`953c678`).
- [x] OE-6-Trigger korrigiert auf „CI-Setup (vor `0.1.0`-DoD); MVP-32" (`f36bbc0`).

---

## 3. Tranche 0a — Architektur- und Plan-Doku

Roadmap §2 Schritte 5–7 + zwei roadmap-externe Plan-Dokumente. Status: 🟡 in Arbeit.

### 3.1 `docs/architecture.md`

DoD:

- [x] Initiale Fassung mit §0..§10, vier Mermaid-Diagrammen (Systemkontext, Hexagon-Zerlegung, Event-Ingest-Sequence, Build-Stages + Lokal-Lab) (`932f0bd`).
- [x] Findings aus Code-Review-Runde 1 (Validierungsreihenfolge, Status-Codes, OTel-Wording, Auth-Pfad, §5.4-Verweis, Tippfehler) korrigiert (`1f21c54`).
- [x] §5.1 Sequenzdiagramm vereinfacht (Cross-Actor + Note-Blöcke statt Self-Messages) für Lesbarkeit (`2b11ad0`).
- [x] Mermaid-ThemeVariables für Boxen, Pfeile, Text, Notes (Kontrast-Tuning) (`af539d9`).
- [x] Mermaid-Canvas-Background `#f8fafc` für hellen Diagramm-Hintergrund unabhängig vom Renderer-Mode (`f27a530`).
- [x] Findings aus Code-Review-Runde 2 (geplante Pfade, OTel-Spans-Pfeil, Session-Aggregation-Wording, Soll/Ist-Konvention) korrigiert (`13c2d27`).
- [x] Variante B integriert: Auth-vor-Body-Reihenfolge in Sequenzdiagramm und Fehler-Tabelle (`40d79d9`).
- [x] Restrukturierung zu reinem Zielbild (Soll/Ist-Trennung): „geplant"-Markierungen entfernt, OTel-Wording auf Soll, §4.1+§4.2 vereint, §5.2 Pull-Richtung, §5.1 Auth-Counter-Scope, Domain-Errors-Wording (`6ab96f1`).
- [x] Roadmap §2 Schritt 5 auf ✅ (`932f0bd`, `08811cb`). (§1.2 enthält nur die Pre-MVP-Schritte 1–4.)

### 3.2 `docs/risks-backlog.md`

DoD:

- [x] Datei angelegt, R-1..R-3 (siehe §2.4 oben) (`953c678`).

### 3.3 `docs/releasing.md` (Skeleton)

DoD:

- [x] Skeleton-Datei angelegt mit §0..§7, Platzhaltern und expliziten TODOs (`67b5aeb`).
- [x] Roadmap §3 verlinkt auf `releasing.md` (`67b5aeb`).
- [ ] §2 Verifikation konkretisieren, sobald **OE-6** (CI-Zielplattformen) entschieden ist.
- [ ] §3 Branching-Modell und Tag-Format konkretisieren, sobald **OE-7** (Release-Konvention) entschieden ist.
- [ ] §4 Asset-Liste, Source-Bundle, Container-Image-Pfad konkretisieren.
- [ ] §6 Rollback-Szenarien analog d-migrate-Pattern.

### 3.4 `docs/plan-0.1.0.md` (dieses Dokument)

DoD:

- [x] Datei angelegt mit Tranchen 0/0a/0b/1 und vollständigem Lieferstand (`6530502`).
- [x] Roadmap verweist auf Plan-Doku — Bezug-Liste, §1.2-Hinweis auf Tranche-0-Detail, §2-Hinweis auf granularen Lieferstand, §3-Akzeptanzkriterien-Spalte (`c172e0c`).

### 3.5 `docs/telemetry-model.md` (Schritt 6)

Deckt alle Telemetrie- und Cardinality-relevanten F-Kennungen aus Lastenheft §7.9, §7.10 und §7.11.

DoD:

- [ ] OTel-Schema und Naming-Konvention für Spans/Counter spezifizieren (Bezug **F-89..F-94**).
- [ ] Cardinality- und Datenmodell-Regeln dokumentieren (Bezug **F-95..F-100** Lastenheft §7.10 sowie **F-101..F-105** als MVP-Variante).
- [ ] Wire-Format für Player-Events spezifizieren — Pflichtfelder, Schema-Version, Versand-Pfad (Bezug **F-106..F-115** Lastenheft §7.11.1–§7.11.3).
- [ ] Backpressure- und Limit-Regeln (Batch-Größe, Rate-Limit, Drop-Politik) dokumentieren (Bezug **F-118..F-123**).
- [ ] Time-Stempel-Felder, Skew-Behandlung und Sequenz-Ordering dokumentieren (Bezug **F-124..F-130**).
- [ ] Schema-Versionierung und Evolution dokumentieren.
- [ ] Roadmap §2 Schritt 6 auf ✅.

### 3.6 `docs/local-development.md` (Schritt 7)

DoD:

- [ ] Quickstart `make dev` dokumentieren (Bezug AK-1, AK-2).
- [ ] Voraussetzungen pro Plattform (Linux/macOS/Windows-WSL).
- [ ] Compose-Stack-Topologie dokumentieren.
- [ ] Test-/Lint-/Coverage-Workflows lokal.
- [ ] Roadmap §2 Schritt 7 auf ✅.

---

## 4. Tranche 0b — Spike-Code-Korrekturen

Findings aus Code-Reviews der Spike-Implementation. Status: 🟡 in Arbeit.

### 4.1 Auth-vor-Body-Reihenfolge kodifiziert (Variante B)

DoD:

- [x] `docs/spike/backend-api-contract.md` §5: Reihenfolge auf 1=Auth-Header, 2=Body, 3=Auth-Token, 4=Rate-Limit, 5=Schema, 6/7=Batch-Form, 8=Event-Felder, 9=Token-Bindung, 10=Erfolg (`40d79d9`).
- [x] `docs/spike/backend-api-contract.md` §5 Tabelle und Folge-Hinweis ergänzt: Body > 256 KB ohne Auth-Header → 401 (`40d79d9`).
- [x] `docs/spike/backend-api-contract.md` §11 neuer Pflichttest „401 bei Body über 256 KB ohne Auth-Header" (`40d79d9`).
- [x] `docs/spike/backend-api-contract.md` Frozen-Status präzisiert (Spike-Vergleichs-Schutz, danach Pflege-Erlaubnis) (`40d79d9`).
- [x] `docs/architecture.md` §5.1 Sequenzdiagramm und Step-Nummern auf neue Reihenfolge (`40d79d9`).
- [x] `apps/api/adapters/driving/http/handler_test.go`: neuer Test `TestHTTP_401_BodyTooLarge_NoToken` (`40d79d9`).
- [x] Docker-Pflichttests grün (`40d79d9`).

### 4.2 Counter-Scope: invalid_events nur für 400/422, dropped_events nur für Backpressure (Hoch + Mittel C1)

Soll laut [API-Kontrakt §7](./spike/backend-api-contract.md) (präzisiert in Commit `9fddfa1`):

- `mtrace_invalid_events_total` zählt **abgelehnte Events** mit Status `400` oder `422`. Auth-Fehler (`401`) zählen nicht. Bei leerem Batch (`events.length == 0`) bleibt der Counter unverändert (Ablehnung sichtbar über HTTP-Status und Access-Logs).
- `mtrace_dropped_events_total` ist für **interne Backpressure-Drops** reserviert (z. B. überlaufender Async-Queue-Puffer) und darf konstant `0` sein. Synchron fehlgeschlagenes `Append` ist kein Drop und inkrementiert den Counter nicht — Sichtbarkeit erfolgt über HTTP-5xx-Histogramm und Logs.

DoD:

- [ ] `apps/api/hexagon/application/register_playback_event_batch.go` Token-Bindung-Branch: `u.metrics.InvalidEvents(len(in.Events))`-Aufruf entfernen. *Step-Mapping*: Kontrakt §5 Step 9; im Code aktuell als Step 7 kommentiert (siehe §4.4 für die Numerierungs-Sync).
- [ ] `apps/api/hexagon/application/register_playback_event_batch.go` Batch-leer-Branch: `u.metrics.InvalidEvents(0)`-Aufruf entfernen — Counter um 0 zu erhöhen ist ein No-Op. *Step-Mapping*: Kontrakt §5 Step 6; im Code aktuell der erste `if len(in.Events) == 0`-Branch innerhalb des kombinierten Code-Step 5 (Batch shape).
- [ ] `apps/api/hexagon/application/register_playback_event_batch.go` Persistenz-Branch: `u.metrics.DroppedEvents(len(parsed))`-Aufruf entfernen. *Step-Mapping*: Kontrakt §5 Step 10 (Persist) bei Repository-Fehler; im Code aktuell als Step 8.
- [ ] `apps/api/hexagon/application/register_playback_event_batch_test.go`: Unit-Test bei `project_id`/Token-Mismatch verifiziert, dass `InvalidEvents` **nicht** inkrementiert wird.
- [ ] `apps/api/hexagon/application/register_playback_event_batch_test.go`: Unit-Test bei leerem Batch verifiziert, dass `InvalidEvents` **nicht** inkrementiert wird.
- [ ] `apps/api/hexagon/application/register_playback_event_batch_test.go`: Unit-Test bei Repository-Fehler (Append → Error) verifiziert, dass `DroppedEvents` **nicht** inkrementiert wird; Use Case gibt den Fehler zurück, HTTP-Adapter liefert `500`.
- [ ] Docker-Targets `test` und `lint` grün.

### 4.3 Telemetry-Driven-Port + OTel-Counter + Request-Span (Mittel-Finding)

Soll laut [API-Kontrakt §8](./spike/backend-api-contract.md) (präzisiert in Commit `9fddfa1`) und Architecture §5.3: OTel-Aufrufe aus dem Use Case laufen ausschließlich über einen frameworkneutralen Driven Port `Telemetry`; Request-Spans erzeugt der HTTP-Adapter direkt. `hexagon/`-Pakete dürfen kein OTel importieren.

DoD:

- [ ] Neuer Port `apps/api/hexagon/port/driven/telemetry.go` mit Interface `Telemetry { BatchReceived(ctx context.Context, size int) }`.
- [ ] Use-Case-Konstruktor `NewRegisterPlaybackEventBatchUseCase` um `telemetry driven.Telemetry`-Parameter erweitert; Aufruf `u.telemetry.BatchReceived(ctx, len(in.Events))` am Eintritt.
- [x] Boundary-Test-Skript `apps/api/scripts/check-architecture.sh` (per `make arch-check` aufrufbar) prüft, dass `hexagon/` keine direkten Imports auf Adapter, OTel, Prometheus, `database/sql` oder `net/http` enthält und die Schichtengrenzen domain → application → port respektiert sind. Aktueller Code besteht den Test (`5784f6e`).
- [ ] Boundary-Test in CI eingebunden, sobald OE-6 entschieden ist.
- [ ] `apps/api/hexagon/`-Pakete importieren weiterhin **kein** OTel — per Boundary-Test verifiziert.
- [ ] Adapter `apps/api/adapters/driven/telemetry/otel.go`: `OTelTelemetry`-Implementierung der Schnittstelle mit OTel-`Int64Counter` `mtrace.api.batches.received`; Attribut `batch.size`.
- [ ] `apps/api/cmd/api/main.go` verdrahtet die `OTelTelemetry`-Implementierung in den Use Case.
- [ ] HTTP-Adapter `apps/api/adapters/driving/http/handler.go`: Request-Span via `otel.Tracer` um den Use-Case-Aufruf; Span-Name `http.handler POST /api/playback-events` o. ä.; Attribute für Status-Code und (bei Erfolg) `batch.size`.
- [ ] `Setup` in `adapters/driven/telemetry`: `MeterProvider` und `TracerProvider` registrieren Reader/Span-Exporter über `go.opentelemetry.io/contrib/exporters/autoexport`. Damit antwortet die Konfiguration auf die Standard-OTel-Env-Vars (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`).
- [ ] Setup ruft autoexport mit explizitem **No-Op-Fallback** auf (`autoexport.WithFallbackMetricReader` / `autoexport.WithFallbackSpanExporter`) — sonst defaultet autoexport ohne Env-Vars auf OTLP, was lokale Dev-Setups versuchen lässt, gegen einen nicht vorhandenen Collector zu pushen. Mit Fallback bleibt der Provider ohne Env-Vars silent.
- [ ] `autoexport`-Modul-Abhängigkeit in `apps/api/go.mod` ergänzt; konkrete Modulversion gepinnt (Default-Protokoll für OTLP — `grpc` vs `http/protobuf` — variiert zwischen autoexport-Versionen, deshalb Pin nötig).
- [ ] Unit-Test `RegisterPlaybackEventBatchTest`: Telemetry-Stub zählt `BatchReceived`-Aufrufe.
- [ ] Adapter-Test `OTelTelemetryTest`: nach N `BatchReceived`-Aufrufen liefert ein `metric.ManualReader` Counter-Wert N (oder die Standard-OTel-Test-Mechanik).
- [ ] Docker-Targets `test` und `lint` grün.

### 4.4 Code-Step-Numbering an Kontrakt anpassen

Code-Kommentare in `register_playback_event_batch.go` nutzen weiterhin die Spike-Reihenfolge (Step 2..8); Architecture und Kontrakt nutzen die neue Reihenfolge (Step 1..10). Beim nächsten Code-Touch synchron ziehen.

DoD:

- [ ] `apps/api/hexagon/application/register_playback_event_batch.go`: Step-Kommentare auf neue Numerierung 3..10.
- [ ] `apps/api/adapters/driving/http/handler.go`: Step-Kommentare 1+2 für Auth-Header und Body.
- [ ] Hinweis auf alte Numerierung in `docs/architecture.md` §5.1 entfernen (sobald Code aktualisiert).

---

## 5. Tranche 1 — MVP-Implementierung

Roadmap §2 Schritte 8–11; Lastenheft RAK-1..RAK-10. Status: ⬜ offen.

### 5.1 Schritt 8 — Dashboard-App und Session-Pfad (`apps/dashboard` + Backend-Erweiterung)

Bezug: MVP-3, MVP-16, F-23..F-28; RAK-7; OE-3 (Datenhaltung MVP) und OE-4 (Frontend-Styling) werden hier entschieden.

DoD Backend (`apps/api`):

- [ ] Domain-Aggregation: `StreamSession` wird aus eingehenden `PlaybackEvent`-Batches abgeleitet — bei jedem Event mit unbekanntem `session_id` wird eine `StreamSession` mit Default-State `Active` erzeugt; Session-Lifecycle (Active → Ended) ist Bonus, falls Time-Budget reicht.
- [ ] **MVP-16** Lokale Speicherung der Sessions und Events: In-Memory ist Pflicht-Default; SQLite als Soll-Erweiterung über OE-3-Folge-ADR. Beide Implementierungen leben hinter dem `EventRepository`-Port plus einem neuen `SessionRepository`-Port (oder vereinheitlicht — Design-Entscheidung im Use Case).
- [ ] Neuer Use Case `ListStreamSessions` und `GetStreamSession` (oder erweiterung des bestehenden); Domain-Sicht auf `StreamSession` mit Event-Zählern.
- [ ] Zwei neue MVP-Endpoints aus Lastenheft §7.3 — `GET /api/stream-sessions` (Liste) und `GET /api/stream-sessions/{id}` (Detail mit Event-Liste). Aktuell sind nur die drei Spike-Pflicht-Endpoints implementiert.
- [ ] Tests: Use-Case-Test für Session-Aggregation aus Event-Batches; HTTP-Integrationstest für die zwei Stream-Sessions-Endpoints.

DoD Dashboard (`apps/dashboard`):

- [ ] SvelteKit-App-Skelett unter `apps/dashboard/` (TypeScript, pnpm).
- [ ] Startseite mit Layout.
- [ ] Dashboard-Route `/sessions` zeigt Liste, ruft `GET /api/stream-sessions` auf.
- [ ] Dashboard-Route `/sessions/:id` zeigt Detail mit Event-Liste, ruft `GET /api/stream-sessions/{id}` auf.
- [ ] Dashboard-Route `/demo` — Test-Player mit hls.js + Player-SDK-Referenzintegration. Pfad in der App: `apps/dashboard/src/routes/demo/` (SvelteKit-Konvention, Lastenheft §7.5.3).
- [ ] API-Client mit typisierten Anfragen.
- [ ] Frontend-Styling: OE-4 entscheiden (eigenes CSS / Tailwind / UI-Library).
- [ ] Dashboard-Build im Docker-Compose-Lab (Schritt 10) eingebunden.

### 5.2 Schritt 9 — Player-SDK (`packages/player-sdk`)

Bezug: MVP-5, F-63..F-67; OE-8 (Paketnamen für npm) wird hier entschieden.

DoD:

- [ ] TypeScript-Package unter `packages/player-sdk/`.
- [ ] **F-63**: Anbindung an ein `HTMLVideoElement` über einen klar abgegrenzten Browser-Adapter (`adapters/hlsjs/` initial; weitere Player als spätere Adapter).
- [ ] **F-64**: Erfassung von Playback-Events aus dem hls.js-Stream (Manifest, Segment, Bitrate-Switch, Rebuffer, Error, …).
- [ ] **F-65**: Erfassung einfacher Metriken pro Session (Startup-Time, Rebuffer-Dauer, ...).
- [ ] **F-66**: Versand der Events via HTTP an `POST /api/playback-events` mit dem Wire-Format aus `docs/telemetry-model.md`. Batching und Sampling konfigurierbar; OpenTelemetry Web SDK als optionaler zweiter Transport-Pfad.
- [ ] **F-67**: Trennung von Browser-Adapter (`adapters/hlsjs/`) und fachlicher Tracking-Logik (`core/`) — strukturelle Boundary, kein gegenseitiger Zugriff: `core/` darf den Browser-Adapter nicht direkt importieren.
- [ ] Browser-Build (ESM + UMD).
- [ ] OE-8 entscheiden (Paketname, Scope).
- [ ] Demo-Integration in `apps/dashboard/src/routes/demo/`.

### 5.3 Schritt 10 — Docker-Compose-Lab (Core)

Bezug: MVP-7..MVP-9, F-82..F-88; RAK-1.

Compose-Stack ist in zwei Profile geteilt: **Core** (Pflicht für RAK-1) und **observability** (Soll-Add-On laut MVP-28/MVP-29). Tempo bleibt explizit aus dem MVP ausgeschlossen (MVP-22 ist Nicht-MVP).

DoD:

- [ ] `docker-compose.yml` im Repo-Wurzelverzeichnis mit dem **Core-Profil** (Default): `apps/api`, `apps/dashboard`, MediaMTX, FFmpeg-Generator — die vier Services aus `architecture.md` §8.2.
- [ ] MediaMTX als `services/media-server/` mit Konfiguration für HLS.
- [ ] FFmpeg-Generator als `services/stream-generator/` mit Teststream.
- [ ] `apps/api`-Container mit ENV-Variablen-Parametrisierung (Listen-Adresse, OTel-Endpoint, OTel-Exporter-Konfig laut `architecture.md` §5.3).
- [ ] `apps/dashboard`-Container im Production-Build oder Vite-Dev-Mode.
- [ ] `make dev` startet das **Core-Profil** und erfüllt damit RAK-1 — Observability-Services starten nur über das observability-Profil (siehe §5.4).
- [ ] `make stop` beendet sauber.
- [ ] Core-Stack mindestens unter Linux verifiziert (Bezug AK-1).

### 5.4 Schritt 11 — Observability-Stack

Bezug: MVP-10 (Muss), MVP-15 (Muss), MVP-28 (Soll Grafana), MVP-29 (Soll OTel-Collector); F-89..F-94 (Muss); Mindestmetriken laut Lastenheft §7.9; **MVP-22 (Tempo) ist explizit Nicht-MVP**.

Soll-Komponenten (Grafana, OTel-Collector) leben im `observability`-Compose-Profil und werden über `make dev-observability` (oder `docker compose --profile observability up`) ergänzend zum Core-Stack gestartet. RAK-1 ist mit dem Core-Stack erfüllt; das observability-Profil ist additiv und nicht für die DoD-Abnahme von Schritt 10 erforderlich.

DoD Pflicht-Anteile (Muss, in `apps/api` direkt):

- [ ] **F-89** Strukturierte Logs in `apps/api` (`log/slog` + JSON-Handler ist bereits aus dem Spike vorhanden; im Compose-Stack stdout-fähig konfiguriert).
- [ ] **F-90** Health Check `/api/health` ist bereits aus dem Spike vorhanden — Verifikation, dass der Endpoint im Compose-Stack `200` liefert (Bezug RAK-3).
- [ ] **F-91** OpenTelemetry-Unterstützung — durch Tranche-0b §4.3 (`Telemetry`-Port + OTLP-Anbindung via `autoexport`) bereits abgedeckt.
- [ ] **F-92** Playback-Events sind als Metriken oder Traces exportierbar — über den `Telemetry`-Port-Counter (Metriken) sowie HTTP-Adapter-Spans (Traces). Aktivierung erfolgt über `OTEL_*`-Env-Vars; im Core-Stack ohne observability-Profil bleiben sie silent.
- [ ] **F-93** Prometheus-Konfiguration unter `observability/prometheus/` mit Scrape-Job für den `api`-Compose-Service (`targets: ["api:8080"]`, `metrics_path: "/api/metrics"`); Compose-Service-Name wird in Schritt 10 verbindlich festgelegt. Prometheus läuft im observability-Profil.
- [ ] Mindestmetriken aus Lastenheft §7.9 in `apps/api` instrumentiert: bereits vorhanden sind die vier API-Kontrakt-Counter (`mtrace_playback_events_total`, `mtrace_invalid_events_total`, `mtrace_rate_limited_events_total`, `mtrace_dropped_events_total`); ergänzend für `0.1.0`: `mtrace_active_sessions`, `mtrace_api_requests_total`, `mtrace_playback_errors_total`, `mtrace_rebuffer_events_total`, `mtrace_startup_time_ms`. Cardinality-Regeln aus Lastenheft §7.10 sind einzuhalten.

DoD Soll-Anteile (`observability`-Profil, MVP-28/MVP-29):

- [ ] **F-94 + MVP-28** Grafana-Container im observability-Profil mit einem einfachen Beispiel-Dashboard unter `observability/grafana/`. Dashboard zeigt mindestens die vier API-Kontrakt-Counter; weitere Mindestmetriken aus §7.9 als Bonus. *Hinweis*: F-94 ist Lastenheft-Muss, MVP-28 ist Soll — die Tabellen widersprechen sich; der Plan folgt der granuläreren MVP-28-Klassifikation.
- [ ] **MVP-29** OTel-Collector unter `services/otel-collector/` im observability-Profil; nimmt OTLP von `apps/api` entgegen und exportiert Metriken zu Prometheus. Trace-Backend (z. B. Jaeger) ist Bonus, **kein** Pflicht-Bestandteil — Tempo ist per MVP-22 Nicht-MVP.
- [ ] `make dev-observability` (oder gleichwertiges Compose-Profile-Aufruf) startet beide Soll-Services additiv zum Core.

### 5.5 Release-Akzeptanzkriterien (Lastenheft §13.1: RAK-1..RAK-10)

Diese zehn Punkte sind die kanonische Abnahmeprüfung für den `0.1.0`-Release. Jeder Eintrag wird ausgehakt, sobald die zugehörige Tranche-1.x-Implementierung ihn erfüllt.

DoD:

- [ ] **RAK-1** `make dev` startet alle notwendigen Dienste (Tranche 5.3).
- [ ] **RAK-2** Dashboard ist erreichbar (Tranche 5.1, 5.3).
- [ ] **RAK-3** API ist erreichbar (Tranche 5.3 — `apps/api` läuft im Compose-Stack auf Port 8080, `/api/health` liefert `200`).
- [ ] **RAK-4** Teststream läuft über MediaMTX (Tranche 5.3).
- [ ] **RAK-5** Player-SDK sendet hls.js-basierte Events (Tranche 5.2).
- [ ] **RAK-6** API nimmt Events an (Tranche 5.1/5.2 End-to-End-Pfad).
- [ ] **RAK-7** Dashboard zeigt empfangene Events und einfache Session-Zusammenhänge (Tranche 5.1).
- [ ] **RAK-8** README beschreibt den Ablauf reproduzierbar — Quickstart-Pfad in `README.md`/`docs/local-development.md` (Tranche 0a §3.6 + Release-Doku).
- [ ] **RAK-9** Prometheus enthält nur aggregierte Metriken — Compose-Stack-Verifikation, dass keine hochkardinalen Labels exportiert werden (Tranche 5.4 + Spot-Check).
- [ ] **RAK-10 (Soll)** Player-Session-Traces sind vorbereitet oder exemplarisch sichtbar — entweder OTel-Spans pro Session oder Dashboard-Trace-Ansicht (Tranche 5.1 + 5.4; minimal: ein Beispiel-Trace im Dashboard oder Tempo).

### 5.6 Übergreifende DoD (Lastenheft §18)

Lastenheft §18 ergänzt die RAKs um dokumentations- und prozessbezogene Items, die kein eigenes RAK haben aber für die Abnahme erforderlich sind:

DoD:

- [x] Architektur in `docs/architecture.md` beschrieben (Tranche 0a §3.1 ausgeliefert; siehe dort für Commit-Liste).
- [ ] Eventmodell in `docs/telemetry-model.md` beschrieben (Tranche 0a §3.5).
- [ ] Tests für zentrale Use Cases vorhanden — Application-Tests für `RegisterPlaybackEventBatch`, HTTP-Integrationstests für alle MVP-Endpoints (Spike-Pflicht: `POST /api/playback-events`, `GET /api/health`, `GET /api/metrics`; ergänzt in Tranche 5.1: `GET /api/stream-sessions`, `GET /api/stream-sessions/{id}`), Tests für die Tranche-0b-Code-Korrekturen.
- [ ] CI führt mindestens Build und Tests aus (verknüpft mit OE-6, MVP-32).
- [ ] `CHANGELOG.md` enthält Eintrag für `0.1.0` (Release-Vorgehen siehe `docs/releasing.md`).
- [ ] Test-Player kann den Stream abspielen (manueller Smoke-Test, deckt RAK-2/4/5 zusammen ab).

---

## 6. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen (Format ```Item-Beschreibung (`<hash>`)```), gegebenenfalls Sub-Items detaillieren.
- Neue Findings landen entweder als neues `[ ]`-Item in der passenden Tranche oder, wenn architekturrelevant und langfristig, in [`risks-backlog.md`](./risks-backlog.md) als `R-X`.
- Beim Schritt-Abschluss: `roadmap.md` §1.2/§2 auf ✅ flippen.
- Nach `0.1.0`-Release: dieses Dokument als historisch archivieren oder in ein `0.2.0`-Plan-Dokument fortschreiben.
