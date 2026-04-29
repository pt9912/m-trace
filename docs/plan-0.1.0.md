# Implementation Plan — `0.1.0` (Backend Core + Demo-Lab)

> **Status**: In Arbeit. Pre-MVP-Vorbereitung (Tranche 0) abgeschlossen, Architektur-Skelett-Doku (Tranche 0a) und Spike-Code-Korrekturen (Tranche 0b) teilweise umgesetzt.  
> **Bezug**: [Lastenheft `1.1.0`](./lastenheft.md) §13.1 (RAK-1, 3, 4, 6, 8 für `0.1.0`), §18 (MVP-DoD); [Roadmap](./roadmap.md) §1.2, §2, §3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).
> **Folge-Pläne**: [`plan-0.1.1.md`](./plan-0.1.1.md) (Player-SDK + Dashboard), [`plan-0.1.2.md`](./plan-0.1.2.md) (Observability-Stack).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert — Commit-Hash genannt; das Item ist im Code/in der Doku enthalten.
- `[ ]` offen — kein Commit, kein Code dahinter.
- `[!]` blockiert durch Lastenheft-Inkonsistenz — Item kann erst angegangen werden, wenn ein Lastenheft-Patch (Tranche 0c, siehe §4a) den Widerspruch auflöst. Siehe `roadmap.md` §7.1 für die Konvention.
- 🟡 in Arbeit — partiell umgesetzt mit weiteren offenen Sub-Items.

Architektur-Soll steht in [`architecture.md`](./architecture.md) und enthält **kein** Status-Tracking. Differenzen Code↔Soll werden hier als offene `[ ]`-DoD-Items getrackt.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Pre-MVP-Vorbereitung — Spike-Sieger auf `main`, Lastenheft `1.0.0`, README/Roadmap, Risiken-Backlog | ✅ |
| 0a | Architektur- und Plan-Doku — `architecture.md`, `releasing.md`, `plan-0.1.0.md`, `telemetry-model.md`, `local-development.md` | 🟡 |
| 0b | Spike-Code-Korrekturen aus Code-Reviews — Auth-vor-Body, InvalidEvents-Scope, OTel-Counter, Step-Numbering | 🟡 |
| 0c | Lastenheft-Patches aus Code-Reviews (`1.0.1`, `1.0.2`) und MVP-Restrukturierung (`1.1.0`) | 🟡 |
| 1 | MVP `0.1.0` — Backend-Erweiterung (Sessions-Endpoints, MVP-16 Persistenz, Lifecycle, F-22-Hook) + Compose-Lab Core | ⬜ |

Player-SDK + Dashboard sind in [`plan-0.1.1.md`](./plan-0.1.1.md), Observability-Stack in [`plan-0.1.2.md`](./plan-0.1.2.md) ausgegliedert (Lastenheft `1.1.0` Restrukturierung).

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

Beschreibt das **Datenmodell** der Telemetrie — Wire-Format, Schema, Cardinality-Regeln. Implementierungs-/Setup-Aspekte (strukturierte Logs, Health-Endpoint, Prometheus- und Grafana-Konfiguration) gehören in Tranche 5.4 (Observability-Stack), nicht hierher.

DoD:

- [ ] OTel-Modell für Spans/Counter spezifizieren — Naming-Konvention, Resource-Attribute, Pflicht-Spans (Bezug **F-91, F-92**).
- [ ] Cardinality- und Datenmodell-Regeln dokumentieren — verbotene Labels, Trennung Aggregat/Per-Session (Bezug **F-95..F-100** Lastenheft §7.10 sowie **F-101..F-105** als MVP-Variante).
- [ ] Wire-Format für Player-Events spezifizieren — Pflichtfelder, Schema-Version, Versand-Pfad, SDK-Identifier (Bezug **F-106..F-115** Lastenheft §7.11.1–§7.11.3).
- [ ] Backpressure- und Limit-Regeln dokumentieren — Batch-Größe, Rate-Limit-Modell, Drop-Politik (Bezug **F-118..F-123**).
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

## 4a. Tranche 0c — Lastenheft-Patches

Aus Code-Reviews entstehende Lastenheft-Korrekturen (interne Inkonsistenzen, Wording-Schärfungen). Jeder Patch erhöht den Lastenheft-Patch-Level (`1.0.0` → `1.0.1` → `1.0.2` …). Status: 🟡 in Arbeit.

### 4a.1 Patch `1.0.1` — F-94 / MVP-28 Harmonisierung (Grafana-Klassifikation)

Lastenheft-interner Widerspruch: §7.9 F-94 listete Grafana mit Priorität **Muss**, §12.2 MVP-28 mit Priorität **Soll**. Beide referenzieren dieselbe Komponente. Im Plan §5.4 hatte ich eigenmächtig zugunsten MVP-28 entschieden; per User-Entscheidung wird stattdessen das Lastenheft korrigiert.

DoD:

- [x] Lastenheft Header: Version `1.0.0` → `1.0.1` (`65405cb`).
- [x] Lastenheft §7.9 F-94: Priorität `Muss` → `Soll`; Wording auf „Grafana **kann** mit einem einfachen Beispiel-Dashboard ausgeliefert werden (harmonisiert mit MVP-28)" (`65405cb`).
- [x] Plan §5.4: Hinweis auf F-94/MVP-28-Widerspruch entfernt; Grafana bleibt im Soll-Block (observability-Profil) (`65405cb`).
- [x] Bezug-Listen (Plan §0, Architecture §0) auf `Lastenheft 1.0.1` aktualisiert (`65405cb`).
- [x] README Status- und Aktueller-Stand-Abschnitt auf `Lastenheft 1.0.1` aktualisiert (`65405cb`).

### 4a.2 Patch `1.0.2` — F-87 / F-88 / Mindestdienste-Harmonisierung

Lastenheft-interner Widerspruch in §7.8: F-87/F-88 klassifizierten Prometheus, Grafana und OTel-Collector als „optional verfügbar" mit Priorität **Muss**, während die Mindestdienste-Tabelle dieselben Dienste ohne Optional-Hinweis listete. Per User-Entscheidung Variante A: Mindestdienste-Tabelle in Pflicht- und Soll-Block aufgeteilt, konsistent mit F-87/F-88 und MVP-28/MVP-29.

DoD:

- [x] Lastenheft Header: Version `1.0.1` → `1.0.2` (`c2e7ac7`).
- [x] Lastenheft §7.8 Mindestdienste: in zwei Tabellen aufgeteilt — Pflicht-Block (`api`, `dashboard`, `mediamtx`, `stream-generator`), Soll-Block (`otel-collector`, `prometheus`, `grafana`) mit Bezug auf F-87/F-88 und MVP-28/MVP-29 (`c2e7ac7`).
- [x] Plan §5.3 `make dev`-Item von `[!]` zurück auf `[ ]` geflippt — Inkonsistenz aufgelöst (`c2e7ac7`).
- [x] Bezug-Listen (Plan §0, Architecture §0) auf `Lastenheft 1.0.2` aktualisiert (`c2e7ac7`).
- [x] README Status- und Aktueller-Stand-Abschnitt auf `Lastenheft 1.0.2` aktualisiert (`c2e7ac7`).

### 4a.3 Patch `1.1.0` — MVP-Phasen-Restrukturierung

Aus User-Entscheidung „Variante 2-A": der ursprüngliche `0.1.0`-MVP wird in drei Sub-Releases (`0.1.0`/`0.1.1`/`0.1.2`) geschnitten, damit jeder Schritt einen demonstrierbaren Eigenwert hat und der Gesamt-Scope nicht in einem einzelnen Cycle landet. Das ist eine **Restrukturierung**, kein Detail-Patch — daher Minor-Bump statt Patch-Level.

DoD:

- [ ] Lastenheft Header: Version `1.0.2` → `1.1.0`.
- [ ] Lastenheft §13: §13.1 (`0.1.0`-RAKs) wird in §13.1 (`0.1.0` Backend Core + Demo-Lab), §13.2 (`0.1.1` Player-SDK + Dashboard) und §13.3 (`0.1.2` Observability-Stack) aufgeteilt; RAK-1..RAK-10 werden auf die drei Sub-Releases verteilt; nachfolgende §13.x-Sections (`0.2.0`..`0.6.0`) entsprechend renumeriert auf §13.4..§13.8.
- [ ] Lastenheft §7.3 F-22: Wording auf „Architektur-Vorbereitung in `apps/api` für Stream Analyzer (Port-Hook); volle Integration ab Phase `0.3.0`" — löst den Tranche-13-Findings-Block (F-22 vs MVP-21).
- [ ] `docs/plan-0.1.0.md` schrumpft auf den neuen `0.1.0`-Scope (Backend + Lab); Player-SDK, Dashboard und Observability werden in eigene Plan-Dokumente ausgelagert.
- [ ] `docs/plan-0.1.1.md` neu angelegt — Player-SDK + Dashboard; Tranchen 0/0a/0b/0c werden referenzierend zu `plan-0.1.0.md` gehalten.
- [ ] `docs/plan-0.1.2.md` neu angelegt — Observability-Stack; analog referenzierend.
- [ ] `docs/roadmap.md` §3 Release-Übersicht auf `0.1.0`/`0.1.1`/`0.1.2`/`0.2.0`/… umgestellt.
- [ ] Bezug-Pins (Plan §0, Architecture §0, README) auf `Lastenheft 1.1.0` aktualisiert.

---

## 5. Tranche 1 — MVP `0.1.0` (Backend Core + Demo-Lab)

Status: ⬜ offen. Bezug: Lastenheft `1.1.0` §13.1 (RAK-1, RAK-3, RAK-4, RAK-6, RAK-8 für `0.1.0`); Roadmap §2 Schritt 10 (Compose-Lab Core) plus Backend-Erweiterungen aus Lastenheft §7.3.

Player-SDK + Dashboard sind in [`plan-0.1.1.md`](./plan-0.1.1.md), Observability-Stack in [`plan-0.1.2.md`](./plan-0.1.2.md) ausgelagert.

### 5.1 Backend-Erweiterung (`apps/api`)

Bezug: MVP-3, MVP-16, F-17..F-22; OE-3 (Datenhaltung MVP) wird hier entschieden.

DoD:

- [ ] Domain-Aggregation: `StreamSession` wird aus eingehenden `PlaybackEvent`-Batches abgeleitet — bei jedem Event mit unbekanntem `session_id` wird eine `StreamSession` mit Default-State `Active` erzeugt.
- [ ] Session-Lifecycle (`Active` → `Stalled` → `Ended`) als Pflicht — Voraussetzung für F-26 im Dashboard (`0.1.1`). Stalled = keine Events in einem Schwellwert-Fenster (z. B. 60 s, konfigurierbar); Ended = explizites End-Event aus dem SDK oder Inaktivität jenseits des Stalled-Fensters.
- [ ] **MVP-16** Lokale Speicherung der Sessions und Events: In-Memory ist Pflicht-Default; SQLite als Soll-Erweiterung über OE-3-Folge-ADR. Beide Implementierungen leben hinter dem `EventRepository`-Port plus einem neuen `SessionRepository`-Port (oder vereinheitlicht — Design-Entscheidung im Use Case).
- [ ] Neuer Use Case `ListStreamSessions` und `GetStreamSession`; Domain-Sicht auf `StreamSession` mit Event-Zählern.
- [ ] Zwei neue MVP-Endpoints aus Lastenheft §7.3 — `GET /api/stream-sessions` (Liste) und `GET /api/stream-sessions/{id}` (Detail mit Event-Liste).
- [ ] **F-22** (Lastenheft `1.1.0`): Architektur-Vorbereitung — Port `hexagon/port/driven/StreamAnalyzer` (oder vergleichbar) als leeres bzw. marker-Interface; Use Case bindet den Port nicht produktiv ein (keine Aufrufe), legt aber den Erweiterungspunkt fest. Volle Integration ab Phase `0.3.0`.
- [ ] Tests: Use-Case-Test für Session-Aggregation aus Event-Batches und Lifecycle-Transitions (Active → Stalled → Ended); HTTP-Integrationstest für die zwei Stream-Sessions-Endpoints.

### 5.2 Docker-Compose-Lab (Core, `0.1.0`-Anteil)

Bezug: MVP-7..MVP-9, F-82..F-88 (nach Patch `1.0.2`); RAK-1, RAK-4.

Compose-Setup nutzt die Docker-Compose-Profile-Semantik korrekt: Core-Services werden **ohne** `profiles:`-Direktive deklariert und starten damit per Default bei `docker compose up`. Das observability-Profil ist additiv und wird in [`plan-0.1.2.md`](./plan-0.1.2.md) gepflegt; das `dashboard`-Service-Add-On wird in [`plan-0.1.1.md`](./plan-0.1.1.md) gepflegt. In `0.1.0` startet das Core-Lab drei Pflicht-Mindestdienste: `api`, `mediamtx`, `stream-generator`.

DoD:

- [ ] `docker-compose.yml` im Repo-Wurzelverzeichnis. Core-Services für `0.1.0` (`api`, `mediamtx`, `stream-generator`) ohne `profiles:`-Direktive — sie starten per Default; entspricht den entsprechenden Pflicht-Mindestdiensten aus Lastenheft §7.8 (nach Patch `1.0.2`).
- [ ] MediaMTX als `services/media-server/` mit Konfiguration für HLS.
- [ ] FFmpeg-Generator als `services/stream-generator/` mit Teststream.
- [ ] `apps/api`-Container mit ENV-Variablen-Parametrisierung (Listen-Adresse, OTel-Endpoint, OTel-Exporter-Konfig laut `architecture.md` §5.3).
- [ ] `make dev` führt `docker compose up --build` ohne Profil-Flag aus — startet damit ausschließlich die Core-Services.
- [ ] `make stop` beendet sauber (`docker compose down`, Profile-aware).
- [ ] Core-Stack mindestens unter Linux verifiziert.
- [ ] Smoke-Test `0.1.0`: nach `make dev` liefert `curl http://localhost:8080/api/health` ein `200`; ein POST mit gültigem Token an `/api/playback-events` liefert `202`; ein GET an `/api/stream-sessions` listet die so erzeugte Session.

### 5.3 Release-Akzeptanzkriterien `0.1.0` (Lastenheft §13.1 nach Patch `1.1.0`)

DoD:

- [ ] **RAK-1** `make dev` startet die in `0.1.0` erforderlichen Pflicht-Dienste (`api`, `mediamtx`, `stream-generator`) (Tranche 5.2).
- [ ] **RAK-3** API ist erreichbar — `/api/health` liefert `200` im Compose-Stack; alle MVP-Endpoints (drei Spike-Pflicht plus die zwei Stream-Sessions-Endpoints) erreichbar (Tranche 5.1/5.2).
- [ ] **RAK-4** Teststream läuft über MediaMTX (Tranche 5.2).
- [ ] **RAK-6** API nimmt Events an (`POST /api/playback-events` mit gültigem Token liefert `202`) (Tranche 5.1/5.2).
- [ ] **RAK-8** README/Local-Development-Doku beschreibt den `0.1.0`-Quickstart reproduzierbar (Initial-Anteil; wird in `0.1.1` und `0.1.2` ergänzt). Bezug Tranche 0a §3.6.

### 5.4 Übergreifende DoD `0.1.0` (Lastenheft §18, `0.1.0`-Anteil)

Dokumentations- und prozessbezogene Items für den `0.1.0`-Release. RAK-spezifische Items stehen in §5.3.

DoD:

- [x] Architektur in `docs/architecture.md` beschrieben (Tranche 0a §3.1 ausgeliefert; siehe dort für Commit-Liste).
- [ ] Eventmodell in `docs/telemetry-model.md` beschrieben (Tranche 0a §3.5) — Pflicht für `0.1.0`, weil das Wire-Format gegen die Spike-API-Kontrakt-Erweiterungen geprüft werden muss.
- [ ] Local-Development-Doku in `docs/local-development.md` (Tranche 0a §3.6) — Pflicht für RAK-8.
- [ ] Tests für zentrale Use Cases vorhanden — Application-Tests für `RegisterPlaybackEventBatch` (inkl. Tranche-0b-Korrekturen) und neue Session-Use-Cases; HTTP-Integrationstests für alle `0.1.0`-MVP-Endpoints.
- [ ] CI führt mindestens Build und Tests aus (verknüpft mit OE-6, MVP-32). Falls OE-6 zum `0.1.0`-Release noch offen ist, wandert die CI-DoD in `plan-0.1.1.md`.
- [ ] `CHANGELOG.md` enthält Eintrag für `0.1.0` (Release-Vorgehen siehe `docs/releasing.md`).

---

## 6. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen (Format ```Item-Beschreibung (`<hash>`)```), gegebenenfalls Sub-Items detaillieren.
- Neue Findings landen entweder als neues `[ ]`-Item in der passenden Tranche oder, wenn architekturrelevant und langfristig, in [`risks-backlog.md`](./risks-backlog.md) als `R-X`.
- Beim Schritt-Abschluss: `roadmap.md` §1.2/§2 auf ✅ flippen.
- Nach `0.1.0`-Release: dieses Dokument als historisch archivieren; Folge-Pläne sind [`plan-0.1.1.md`](./plan-0.1.1.md) und [`plan-0.1.2.md`](./plan-0.1.2.md), danach `plan-0.2.0.md`.
