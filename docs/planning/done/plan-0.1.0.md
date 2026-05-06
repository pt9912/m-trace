# Implementation Plan — `0.1.0` (Backend Core + Demo-Lab)

> **Status**: ✅ abgeschlossen. `0.1.0` ist ausgeliefert; dieses Dokument bleibt als historischer Lieferstand erhalten.  
> **Bezug**: [Lastenheft `1.1.6`](../../../spec/lastenheft.md) §13.1 (RAK-1, 3, 4, 6, 8 für `0.1.0`), §18 (MVP-DoD); [Roadmap](../in-progress/roadmap.md) §1.2, §2, §3; [Architektur (Zielbild)](../../../spec/architecture.md); [API-Kontrakt](../../../spec/backend-api-contract.md); [Risiken-Backlog](../open/risks-backlog.md).
> **Folge-Pläne**: [`plan-0.1.1.md`](./plan-0.1.1.md) (Player-SDK + Dashboard), [`plan-0.1.2.md`](./plan-0.1.2.md) (Observability-Stack).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert — Commit-Hash genannt; das Item ist im Code/in der Doku enthalten.
- `[ ]` offen — kein Commit, kein Code dahinter.
- `[!]` blockiert durch Lastenheft-Inkonsistenz — Item kann erst angegangen werden, wenn ein Lastenheft-Patch (Tranche 0c, siehe §4a) den Widerspruch auflöst. Siehe `roadmap.md` §7.1 für die Konvention.
- 🟡 in Arbeit — partiell umgesetzt mit weiteren offenen Sub-Items.

Architektur-Soll steht in [`architecture.md`](../../../spec/architecture.md) und enthält **kein** Status-Tracking. Differenzen Code↔Soll werden hier als offene `[ ]`-DoD-Items getrackt.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Pre-MVP-Vorbereitung — Spike-Sieger auf `main`, Lastenheft `1.0.0`, README/Roadmap, Risiken-Backlog | ✅ |
| 0a | Architektur- und Plan-Doku — `architecture.md`, `releasing.md`, `plan-0.1.0.md`, `telemetry-model.md`, `local-development.md` | ✅ |
| 0b | Spike-Code-Korrekturen aus Code-Reviews — Auth-vor-Body, InvalidEvents-Scope, OTel-Counter, Step-Numbering | ✅ |
| 0c | Lastenheft-Patches (fortlaufend) — `1.0.1`, `1.0.2`, `1.1.0` (Restrukturierung), `1.1.1`, `1.1.2`, `1.1.3`, `1.1.4`, `1.1.5`, `1.1.6`, `1.1.7`, `1.1.8`, `1.1.9`, `1.1.10` | ✅ bis `1.1.10`; fortlaufend |
| 1 | MVP `0.1.0` — Backend-Erweiterung (Sessions-Endpoints, MVP-16 Persistenz, Lifecycle, F-22-Hook) + Compose-Lab Core | ✅ |

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

- [x] Issue-Backlog-Form entschieden: Markdown-Datei `docs/planning/open/risks-backlog.md` analog cmake-xray/d-migrate-Pattern (`953c678`).
- [x] R-1 Hexagon-Boundaries (Disziplin-basiert, kein Compile-Time-Enforcement) eingetragen mit Verweis auf Folge-ADR „`apps/api` Multi-Modul-Aufteilung" (`953c678`).
- [x] R-2 CGO/SRT bricht distroless-static eingetragen mit Verweis auf Folge-ADR „SRT-Binding-Stack" (`953c678`).
- [x] R-3 Go-WebSocket-Ökosystem fragmentiert eingetragen mit Verweis auf Folge-ADR „WebSocket vs. SSE" (`953c678`).
- [x] Roadmap §1.2 und §2 Schritt 4 auf ✅ (`953c678`).
- [x] Roadmap §5 „Issue-Backlog-Form" entfernt (resolved) (`953c678`).
- [x] OE-6-Trigger korrigiert auf „CI-Setup (vor `0.1.0`-DoD); MVP-32" (`f36bbc0`).

---

## 3. Tranche 0a — Architektur- und Plan-Doku

Roadmap §2 Schritte 5–7 + zwei roadmap-externe Plan-Dokumente. Status: ✅ abgeschlossen.

### 3.1 `spec/architecture.md`

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

### 3.2 `docs/planning/open/risks-backlog.md`

DoD:

- [x] Datei angelegt, R-1..R-3 (siehe §2.4 oben) (`953c678`).

### 3.3 `docs/user/releasing.md` (Skeleton)

DoD:

- [x] Skeleton-Datei angelegt mit §0..§7, Platzhaltern und expliziten TODOs (`67b5aeb`).
- [x] Roadmap §3 verlinkt auf `releasing.md` (`67b5aeb`).
- [x] §2 Verifikation konkretisieren: `0.1.0` nutzt GitHub Actions auf `ubuntu-24.04`; Workflow `.github/workflows/build.yml` führt Root-Targets `make test`, `make lint`, `make coverage-gate`, `make arch-check`, `make build` aus (`46e45ec`).
- [x] §3 Branching-Modell und Tag-Format konkretisieren: trunk-based auf `main`, annotierte SemVer-Tags `vX.Y.Z`.
- [x] §4 Asset-Liste, Source-Bundle, Container-Image-Pfad konkretisieren (`a9f0c53`).
- [x] §6 Rollback-Szenarien analog d-migrate-Pattern (`a9f0c53`).

### 3.4 `docs/planning/done/plan-0.1.0.md` (dieses Dokument)

DoD:

- [x] Datei angelegt mit Tranchen 0/0a/0b/1 und vollständigem Lieferstand (`6530502`).
- [x] Roadmap verweist auf Plan-Doku — Bezug-Liste, §1.2-Hinweis auf Tranche-0-Detail, §2-Hinweis auf granularen Lieferstand, §3-Akzeptanzkriterien-Spalte (`c172e0c`).

### 3.5 `spec/telemetry-model.md` (Schritt 6)

Beschreibt das **Datenmodell** der Telemetrie — Wire-Format, Schema, Cardinality-Regeln. Implementierungs-/Setup-Aspekte (strukturierte Logs, Health-Endpoint, Prometheus- und Grafana-Konfiguration) gehören in [`plan-0.1.2.md`](./plan-0.1.2.md) Tranche 1/2 (Observability-Stack), nicht hierher.

DoD:

- [x] Skelett-Datei angelegt mit §0..§6-Sektionen und Bezug-Verweisen, alle Inhalts-Sections als „TODO"-Platzhalter (`c86e021`). Referenz-Anker für andere Plan-Dokumente (`plan-0.1.1.md` §2, `plan-0.1.2.md` §4) ist damit gesetzt.
- [x] OTel-Modell für Spans/Counter spezifizieren — Naming-Konvention, Resource-Attribute, Pflicht-Spans (Bezug **F-91, F-92**) (`e532e1e`).
- [x] Cardinality- und Datenmodell-Regeln dokumentieren — verbotene Labels, Trennung Aggregat/Per-Session (Bezug **F-95..F-100** Lastenheft §7.10 sowie **F-101..F-105** als MVP-Variante) (`e532e1e`).
- [x] Wire-Format für Player-Events spezifizieren — Pflichtfelder, Schema-Version, Versand-Pfad, SDK-Identifier (Bezug **F-106..F-115** Lastenheft §7.11.1–§7.11.3) (`e532e1e`).
- [x] Backpressure- und Limit-Regeln dokumentieren — Batch-Größe, Rate-Limit-Modell, Drop-Politik (Bezug **F-118..F-123**) (`e532e1e`).
- [x] Time-Stempel-Felder, Skew-Behandlung und Sequenz-Ordering dokumentieren (Bezug **F-124..F-130**) (`e532e1e`).
- [x] Schema-Versionierung und Evolution dokumentieren (`e532e1e`).
- [x] Roadmap §2 Schritt 6 auf ✅ (`e532e1e`).

### 3.6 `docs/user/local-development.md` (Schritt 7)

DoD:

- [x] Skelett-Datei angelegt mit §0..§5-Sektionen und Bezug-Verweisen, alle Inhalts-Sections als „TODO"-Platzhalter (`c86e021`). Referenz-Anker für `plan-0.1.2.md` §4.1 (RAK-8-Refinement) ist damit gesetzt.
- [x] Quickstart `make dev` dokumentieren (Bezug AK-1, AK-2) (`2eede43`).
- [x] Voraussetzungen pro Plattform (Linux/macOS/Windows-WSL) (`2eede43`).
- [x] Compose-Stack-Topologie dokumentieren (`2eede43`).
- [x] Test-/Lint-/Coverage-Workflows lokal (`2eede43`).
- [x] Roadmap §2 Schritt 7 auf ✅ (`2eede43`).

---

## 4. Tranche 0b — Spike-Code-Korrekturen

Findings aus Code-Reviews der Spike-Implementation. Status: ✅ abgeschlossen.

### 4.1 Auth-vor-Body-Reihenfolge kodifiziert (Variante B)

DoD:

- [x] `spec/backend-api-contract.md` §5: Reihenfolge auf 1=Auth-Header, 2=Body, 3=Auth-Token, 4=Rate-Limit, 5=Schema, 6/7=Batch-Form, 8=Event-Felder, 9=Token-Bindung, 10=Erfolg (`40d79d9`).
- [x] `spec/backend-api-contract.md` §5 Tabelle und Folge-Hinweis ergänzt: Body > 256 KB ohne Auth-Header → 401 (`40d79d9`).
- [x] `spec/backend-api-contract.md` §11 neuer Pflichttest „401 bei Body über 256 KB ohne Auth-Header" (`40d79d9`).
- [x] `spec/backend-api-contract.md` Frozen-Status präzisiert (Spike-Vergleichs-Schutz, danach Pflege-Erlaubnis) (`40d79d9`).
- [x] `spec/architecture.md` §5.1 Sequenzdiagramm und Step-Nummern auf neue Reihenfolge (`40d79d9`).
- [x] `apps/api/adapters/driving/http/handler_test.go`: neuer Test `TestHTTP_401_BodyTooLarge_NoToken` (`40d79d9`).
- [x] Docker-Pflichttests grün (`40d79d9`).

### 4.2 Counter-Scope: invalid_events nur für 400/422, dropped_events nur für Backpressure (Hoch + Mittel C1)

Soll laut [API-Kontrakt §7](../../../spec/backend-api-contract.md) (präzisiert in Commit `9fddfa1`):

- `mtrace_invalid_events_total` zählt **abgelehnte Events** mit Status `400` oder `422`. Auth-Fehler (`401`) zählen nicht. Bei leerem Batch (`events.length == 0`) bleibt der Counter unverändert (Ablehnung sichtbar über HTTP-Status und Access-Logs).
- `mtrace_dropped_events_total` ist für **interne Backpressure-Drops** reserviert (z. B. überlaufender Async-Queue-Puffer) und darf konstant `0` sein. Synchron fehlgeschlagenes `Append` ist kein Drop und inkrementiert den Counter nicht — Sichtbarkeit erfolgt über HTTP-5xx-Histogramm und Logs.

DoD:

- [x] `apps/api/hexagon/application/register_playback_event_batch.go` Token-Bindung-Branch (Step 9): `u.metrics.InvalidEvents(len(in.Events))`-Aufruf entfernen (`372a6d4`).
- [x] `apps/api/hexagon/application/register_playback_event_batch.go` Batch-leer-Branch (Step 6): `u.metrics.InvalidEvents(0)`-Aufruf entfernen — Counter um 0 zu erhöhen ist ein No-Op (`372a6d4`).
- [x] `apps/api/hexagon/application/register_playback_event_batch.go` Persistenz-Branch (Step 10): `u.metrics.DroppedEvents(len(parsed))`-Aufruf entfernen (`372a6d4`).
- [x] `apps/api/hexagon/application/register_playback_event_batch_test.go`: Unit-Test bei `project_id`/Token-Mismatch verifiziert, dass `InvalidEvents` **nicht** inkrementiert wird (`372a6d4`).
- [x] `apps/api/hexagon/application/register_playback_event_batch_test.go`: Unit-Test bei leerem Batch verifiziert, dass `InvalidEvents` **nicht** inkrementiert wird (`372a6d4`).
- [x] `apps/api/hexagon/application/register_playback_event_batch_test.go`: Unit-Test bei Repository-Fehler (Append → Error) verifiziert, dass `DroppedEvents` **nicht** inkrementiert wird; Use Case gibt den Fehler zurück, HTTP-Adapter liefert `500` (`372a6d4`).
- [x] Docker-Targets `test` und `lint` grün (`372a6d4`).

### 4.3 Telemetry-Driven-Port + OTel-Counter + Request-Span (Mittel-Finding)

Soll laut [API-Kontrakt §8](../../../spec/backend-api-contract.md) (präzisiert in Commit `9fddfa1`) und Architecture §5.3: OTel-Aufrufe aus dem Use Case laufen ausschließlich über einen frameworkneutralen Driven Port `Telemetry`; Request-Spans erzeugt der HTTP-Adapter direkt. `hexagon/`-Pakete dürfen kein OTel importieren.

**Scope-Abgrenzung gegenüber `0.1.2` Observability-Stack**: §4.3 liefert die **API-seitige Telemetrie-Vorbereitung** in `apps/api` (Port + Adapter-Implementierung + Request-Spans + autoexport-Setup). Die **Observability-Stack-Komponenten** (Prometheus-Service, Grafana-Service, OTel-Collector-Service mit ihren Compose-/Konfig-Artefakten) sind explizit nicht hier, sondern in [`plan-0.1.2.md`](./plan-0.1.2.md). Die Vorziehung der API-seitigen Vorbereitung in `0.1.0` ist bewusst — ohne den Telemetry-Port wäre die Architektur in `0.1.0` instabil (Hexagon-Boundary-Verletzungen müssten erst nachträglich aufgeräumt werden), und der API-Kontrakt §8 verlangt „mindestens einen Counter oder Span" als Spike-Erbe.

DoD:

- [x] Neuer Port `apps/api/hexagon/port/driven/telemetry.go` mit Interface `Telemetry { BatchReceived(ctx context.Context, size int) }` (`51b3812`).
- [x] Use-Case-Konstruktor `NewRegisterPlaybackEventBatchUseCase` um `telemetry driven.Telemetry`-Parameter erweitert; Aufruf `u.telemetry.BatchReceived(ctx, len(in.Events))` am Eintritt — vor Auth, damit auch fehlgeschlagene Auth-Requests im received-Counter erscheinen (`51b3812`).
- [x] Boundary-Test-Skript `apps/api/scripts/check-architecture.sh` (per `make arch-check` aufrufbar) prüft, dass `hexagon/` keine direkten Imports auf Adapter, OTel, Prometheus, `database/sql` oder `net/http` enthält und die Schichtengrenzen domain → application → port respektiert sind. Aktueller Code besteht den Test (`5784f6e`).
- [x] Boundary-Test in CI eingebunden (`make arch-check` im Workflow `.github/workflows/build.yml`) (`46e45ec`).
- [x] `apps/api/hexagon/`-Pakete importieren weiterhin **kein** OTel — per Boundary-Test verifiziert (`make arch-check` grün auf `51b3812`).
- [x] Adapter `apps/api/adapters/driven/telemetry/otel.go`: `OTelTelemetry`-Implementierung der Schnittstelle mit OTel-`Int64Counter` `mtrace.api.batches.received` (Punkt-Notation laut OTel-Semconv); Attribut `batch.size`. **Naming-Translation**: das OTel→Prometheus-Mapping ersetzt `.` durch `_`, daher erscheint der Counter in Prometheus als `mtrace_api_batches_received` (vom OTLP-Exporter automatisch konvertiert). Smoke-Test-Regex `^mtrace_.+` (Plan `0.1.2` §4) deckt beide Namen ab — den translated Counter sowie die direkten Prometheus-Counter aus `adapters/driven/metrics`. Dokumentation in `spec/telemetry-model.md` §2 erfasst diese Translation explizit (`51b3812`).
- [x] `apps/api/cmd/api/main.go` verdrahtet die `OTelTelemetry`-Implementierung in den Use Case (`51b3812`).
- [x] HTTP-Adapter `apps/api/adapters/driving/http/handler.go`: Request-Span via `otel.Tracer` um den Use-Case-Aufruf; Span-Name `http.handler POST /api/playback-events`; Attribute `http.method`, `http.route`, `http.status_code`, `batch.size` (sobald JSON geparst), `batch.outcome` (`51b3812`).
- [x] `Setup` in `adapters/driven/telemetry`: `MeterProvider` und `TracerProvider` registrieren Reader/Span-Exporter über `go.opentelemetry.io/contrib/exporters/autoexport`. Damit antwortet die Konfiguration auf die Standard-OTel-Env-Vars (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`) (`51b3812`).
- [x] **Autoexport-Modul-Pin** festgelegt: `go.opentelemetry.io/contrib/exporters/autoexport v0.57.0` (kompatibel mit OTel SDK `v1.32.0`, `WithFallback*` verfügbar) — gepinnt in `apps/api/go.mod` (`51b3812`).
- [x] Setup ruft autoexport mit explizitem **No-Op-Fallback** auf (`autoexport.WithFallbackMetricReader` mit `ManualReader`, `autoexport.WithFallbackSpanExporter` mit lokalem `noopSpanExporter`); ohne `OTEL_*`-Env-Vars bleibt der Provider silent (`51b3812`).
- [x] `autoexport`-Modul-Abhängigkeit in `apps/api/go.mod`/`go.sum` mit der gepinnten Version ergänzt; `make test` und `make lint` grün auf der frischen Module-Resolution (`51b3812`).
- [x] Unit-Test `register_playback_event_batch_test.go`: Telemetry-Stub zählt `BatchReceived`-Aufrufe; `TestTelemetryReceivedBeforeAuth` verifiziert Auth-Reject-Pfad (`51b3812`).
- [x] Adapter-Test `adapters/driven/telemetry/otel_test.go`: nach N `BatchReceived`-Aufrufen liefert ein `sdkmetric.ManualReader` einen Counter mit Wert 1 pro `batch.size`-Attribut-Bucket (`51b3812`).
- [x] Span-Test `adapters/driving/http/span_test.go`: `tracetest.SpanRecorder` verifiziert Span-Name, `http.method`/`http.route`/`http.status_code`/`batch.size`/`batch.outcome`-Attribute auf 202- und 401-Pfaden (`51b3812`).
- [x] Docker-Targets `test`, `lint` und `arch-check` grün (`51b3812`).

### 4.4 Code-Step-Numbering an Kontrakt anpassen

Code-Kommentare in `register_playback_event_batch.go` nutzen weiterhin die Spike-Reihenfolge (Step 2..8); Architecture und Kontrakt nutzen die neue Reihenfolge (Step 1..10). Beim nächsten Code-Touch synchron ziehen.

DoD:

- [x] `apps/api/hexagon/application/register_playback_event_batch.go`: Step-Kommentare auf neue Numerierung 3..10 (Steps 3 ResolveByToken, 4 RateLimit, 5 SchemaVersion, 6 BatchEmpty, 7 BatchTooLarge, 8 EventFields, 9 TokenBinding, 10 PersistAccept) (`dbdcb67`).
- [x] `apps/api/adapters/driving/http/handler.go`: Step-Kommentare 1+2 für Auth-Header und Body; Doc-Block auf Steps 3-10 im Use Case verweisen (`dbdcb67`).
- [x] Hinweis auf alte Numerierung in `spec/architecture.md` §5.1 entfernen — bereits in einem früheren Edit erfolgt; aktueller Stand zeigt nur die Kontrakt-Numerierung 1..10 (`dbdcb67`).

---

## 4a. Tranche 0c — Lastenheft-Patches

Aus Code-Reviews und User-Entscheidungen entstehende Lastenheft-Korrekturen (interne Inkonsistenzen, Wording-Schärfungen, Restrukturierung). Jeder Patch erhöht den Lastenheft-Versionsstand (`1.0.x` für Inhalts-Patches, `1.x.0` für strukturelle Bumps). **Diese Tranche ist fortlaufend** — sie ist auch dann „🟡 fortlaufend", wenn alle bisherigen Patch-Items abgeschlossen sind, weil weitere Patches während `0.1.x` jederzeit ergänzt werden können. Wartung: neue Patches werden als neuer §4a.X-Eintrag mit eigener Patch-Versionsnummer aufgenommen.

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

- [x] Lastenheft Header: Version `1.0.2` → `1.1.0` (`31ccb47`).
- [x] Lastenheft §13: §13.1 (`0.1.0`-RAKs) wird in §13.1 (`0.1.0` Backend Core + Demo-Lab), §13.2 (`0.1.1` Player-SDK + Dashboard) und §13.3 (`0.1.2` Observability-Stack) aufgeteilt; RAK-1..RAK-10 werden auf die drei Sub-Releases verteilt; nachfolgende §13.x-Sections (`0.2.0`..`0.6.0`) entsprechend renumeriert auf §13.4..§13.8 (`31ccb47`).
- [x] Lastenheft §7.3 F-22: Wording auf „Architektur-Vorbereitung in `apps/api` für Stream Analyzer (Port-Hook); volle Integration ab Phase `0.3.0`" — löst den Tranche-13-Findings-Block (F-22 vs MVP-21) (`31ccb47`).
- [x] `docs/planning/done/plan-0.1.0.md` schrumpft auf den neuen `0.1.0`-Scope (Backend + Lab); Player-SDK, Dashboard und Observability werden in eigene Plan-Dokumente ausgelagert (`31ccb47`).
- [x] `docs/planning/done/plan-0.1.1.md` neu angelegt — Player-SDK + Dashboard; Tranchen 0/0a/0b/0c werden referenzierend zu `plan-0.1.0.md` gehalten (`31ccb47`).
- [x] `docs/planning/done/plan-0.1.2.md` neu angelegt — Observability-Stack; analog referenzierend (`31ccb47`).
- [x] `docs/planning/in-progress/roadmap.md` §3 Release-Übersicht auf `0.1.0`/`0.1.1`/`0.1.2`/`0.2.0`/… umgestellt (`0c4cab6`).
- [x] Bezug-Pins (Plan §0, Architecture §0, README) auf `Lastenheft 1.1.0` aktualisiert (`0c4cab6`).

### 4a.4 Patch `1.1.1` — Mindestdienste-Hinweis für Sub-Releases

Aus Code-Review-Finding: Lastenheft §7.8 listet `dashboard` weiterhin als Pflicht-Mindestdienst, während `plan-0.1.0.md` §5.2 für `0.1.0` nur drei Core-Services (`api`, `mediamtx`, `stream-generator`) startet. Die Mindestdienste-Tabelle in §7.8 ist korrekt für den `0.1.x`-End-Zustand (nach `0.1.1`); für die Sub-Release-Subsets fehlte ein Hinweis. Patch `1.1.1` ergänzt diesen.

DoD:

- [x] Lastenheft Header: Version `1.1.0` → `1.1.1` (`85ef32a`).
- [x] Lastenheft §7.8 nach den Mindestdienste-Tabellen: Hinweisblock ergänzt, dass die Tabellen den `0.1.x`-End-Zustand beschreiben; Pflicht-Mindestdienste werden stufenweise mit `0.1.0`/`0.1.1`/`0.1.2` aktiviert; Sub-Release-Subsets stehen im jeweiligen Plan-Dokument (`85ef32a`).
- [x] Bezug-Pins (Plan §0, Architecture §0, README) auf `Lastenheft 1.1.1` aktualisiert (`85ef32a`).

### 4a.5 Patch `1.1.2` — `mtrace_invalid_events_total`-Wording

Aus Code-Review-Finding: Lastenheft §7.9 beschrieb `mtrace_invalid_events_total` als „Anzahl wegen Schema/Auth abgelehnter Events", während API-Kontrakt §7 (seit Patch `9fddfa1`) und Architecture §5.1 Auth-Fehler explizit aus dem Counter ausschließen. Das war eine Lastenheft-vs-Kontrakt-Inkonsistenz, die der Plan §4.2 unilateral aufgelöst hatte — Patch `1.1.2` zieht das Lastenheft korrekt nach.

DoD:

- [x] Lastenheft Header: Version `1.1.1` → `1.1.2` (`0d6ffae`).
- [x] Lastenheft §7.9 Mindestmetriken-Tabelle: Wording für `mtrace_invalid_events_total` auf „Anzahl wegen Schema-/Validierungsfehlern (`400`/`422`) abgelehnter Events; Auth-Fehler (`401`) zählen nicht (harmonisiert mit API-Kontrakt §7 in Patch `1.1.2`)" — entfernt „Auth" aus dem Counter-Scope (`0d6ffae`).
- [x] Bezug-Pins (Plan §0, Architecture §0, README) auf `Lastenheft 1.1.2` aktualisiert (`0d6ffae`).

### 4a.6 Patch `1.1.3` — §12 MVP-Umfang nach Sub-Release-Split + Roadmap-Schritte 6–11 + MediaMTX-Link

Aus Code-Review-Findings: nach der Sub-Release-Schneidung (Patch `1.1.0`) waren noch drei Stellen auf den ursprünglichen einzelnen `0.1.0`-Scope ausgerichtet — Lastenheft §12, Roadmap-Schritte 8–11, Roadmap-Schritt 6. Plus eine kleine MediaMTX-Bezeichnungs-Korrektur in `plan-0.1.1.md`.

DoD:

- [x] Lastenheft Header: Version `1.1.2` → `1.1.3` (`a39f943`).
- [x] Lastenheft §12.1: Header umformuliert auf „Muss-Anforderungen für die `0.1.x`-Phase (Gesamt-MVP)"; Hinweisblock ergänzt, der MVP-1..MVP-29 auf die Sub-Releases `0.1.0`/`0.1.1`/`0.1.2` verteilt analog zur RAK-Verteilung in §13.1–§13.3 (`a39f943`).
- [x] Roadmap §2 Schritt 6 enger gefasst: Beschreibung auf „Datenmodell, Wire-Format, Cardinality — kein Observability-Setup"; Verweis-IDs auf F-91, F-92, F-95..F-105, F-106..F-115, F-118..F-130, AK-9 (F-89/F-90/F-93/F-94 entfernt — gehören zu Roadmap §2 Schritt 11) (`a39f943`).
- [x] Roadmap §2 Schritte 8–11 Trigger an Sub-Release-Abhängigkeiten angeglichen: Schritt 8/9 „Nach `0.1.0`-Release" mit Verweis auf `plan-0.1.1.md`; Schritt 10 dreigeteilt (Core in `0.1.0`, dashboard in `0.1.1`, observability in `0.1.2`); Schritt 11 „Nach `0.1.1`-Release" mit Verweis auf `plan-0.1.2.md`. Schritt 9 Verweis-Range auf F-63..F-67 aktualisiert (war F-63..F-65, fehlten F-66/F-67 aus Patch der Dashboard-DoD) (`a39f943`).
- [x] `plan-0.1.1.md` §3 F-40-Item: MediaMTX-Link von Port `8888` (HLS) auf Port `9997` (API/Status) verlegt; „Web-UI"-Bezeichnung entfernt — MediaMTX hat keine native Web-UI (`a39f943`).
- [x] Bezug-Pins (Plan §0, Architecture §0, README) auf `Lastenheft 1.1.3` aktualisiert (`a39f943`).

---

### 4a.7 Patch `1.1.4` — OE-1/OE-6/OE-7 auflösen

Nach Lieferung des Compose-Labs und des GitHub-Actions-Workflows sind
die vor `0.1.0` blockierenden offenen Entscheidungen geklärt.

DoD:

- [x] Lastenheft Header: Version `1.1.3` → `1.1.4` (`a9f0c53`).
- [x] Lastenheft Header: Lizenzziel auf konkrete Lizenz **MIT** gesetzt; `LICENSE` ist bereits vorhanden (`a9f0c53`).
- [x] Lastenheft §16.2: OE-1 resolved — Projektlizenz MIT (`a9f0c53`).
- [x] Lastenheft §16.2: OE-6 resolved — CI-Zielplattform GitHub Actions `ubuntu-24.04` (`a9f0c53`).
- [x] Lastenheft §16.2: OE-7 resolved — trunk-based auf `main`, annotierte SemVer-Tags `vX.Y.Z`, GitHub Release aus `CHANGELOG.md` (`a9f0c53`).
- [x] Bezug-Pins (README, Architektur, Telemetry-Modell, Local-Development, Pläne `0.1.0`/`0.1.1`/`0.1.2`) auf `Lastenheft 1.1.4` aktualisiert (`a9f0c53`).

---

### 4a.8 Patch `1.1.5` — OE-8 Paketname Player-SDK auflösen

Mit dem `0.1.1`-Player-SDK-Skelett wird der npm-Paketname festgelegt.

DoD:

- [x] Lastenheft Header: Version `1.1.4` → `1.1.5` (`bae4a2a`).
- [x] Lastenheft §16.2: OE-8 resolved — Player-SDK-Paketname `@m-trace/player-sdk` (`bae4a2a`).
- [x] Bezug-Pins (README, Architektur, Telemetry-Modell, Local-Development, Pläne `0.1.0`/`0.1.1`/`0.1.2`) auf `Lastenheft 1.1.5` aktualisiert (`bae4a2a`).

---

### 4a.9 Patch `1.1.6` — OE-4 Frontend-Styling auflösen

Mit dem `0.1.1`-Dashboard-Skelett wird die Styling-Strategie festgelegt.

DoD:

- [x] Lastenheft Header: Version `1.1.5` → `1.1.6` (`1a6a6c7`).
- [x] Lastenheft §16.2: OE-4 resolved — eigenes CSS ohne Tailwind/UI-Library (`1a6a6c7`).
- [x] Bezug-Pins (README, Architektur, Telemetry-Modell, Local-Development, Pläne `0.1.0`/`0.1.1`/`0.1.2`) auf `Lastenheft 1.1.6` aktualisiert (`1a6a6c7`).

### 4a.10 Patch `1.1.7` — OE-8 Paketname Player-SDK neu entscheiden

Mit `0.2.0` wird das Player-SDK erstmals publizierbar. Der ursprünglich
für `0.1.x` dokumentierte npm-Scope `@m-trace` ist nicht als npm-Org
reserviert; Maintainer publisht Pakete bereits unter `@npm9912`. Daher
wird OE-8 vor der ersten Veröffentlichung neu entschieden. Details stehen
in [`docs/planning/done/migrate-package-name.md`](./migrate-package-name.md) (im `0.8.0`-Wartungs-Sweep nach `done/` archiviert).

DoD:

- [x] Lastenheft Header: Version `1.1.6` → `1.1.7`.
- [x] Lastenheft §16.2: OE-8 resolved — Player-SDK-Paketname `@npm9912/player-sdk` ab `0.2.0`; `0.1.x`-Lieferstand unter `@m-trace/player-sdk` bleibt historische Wahrheit, wurde aber nie öffentlich publishet.
- [x] Lebende Code-, Doku- und Package-Stellen folgen `docs/planning/done/migrate-package-name.md` §2.1; historische `0.1.x`-Artefakte bleiben gemäß §2.2 unverändert.

### 4a.11 Patch `1.1.8` — OE-3 und OE-5 auflösen

Mit dem `0.4.0`-Scope-Cut werden die zwei verbliebenen offenen
Entscheidungen aus §16.2 geschlossen: ADR-0002 legt SQLite als lokalen
Durable-Store fest, ADR-0003 legt SSE mit Polling-Fallback für
Dashboard-Live-Updates fest.

DoD:

- [x] Lastenheft Header: Version `1.1.7` → `1.1.8`.
- [x] Lastenheft §16.2: OE-3 resolved — SQLite als lokaler Durable-Store ab `0.4.0`.
- [x] Lastenheft §16.2: OE-5 resolved — Server-Sent Events mit Polling-Fallback; WebSocket nicht in `0.4.0`.
- [x] Roadmap, Architektur, Telemetry-Modell, Risiken-Backlog und `plan-0.4.0.md` referenzieren ADR-0002/ADR-0003 konsistent.

### 4a.12 Patch `1.1.9` — `0.7.0` WebRTC-Lab-Erweiterung als Folge zum `0.5.0`-Vorbereitungspfad

Mit dem `0.5.0`-Release (Multi-Protocol Lab) wurde der WebRTC-Pfad bewusst als Doku-only Vorbereitungspfad geschnitten (RAK-39): `examples/webrtc/` enthält ausschließlich Konfigurations-/Out-of-Scope-Doku, kein Compose-Stack, kein Smoke. Die produktive Lab-Erweiterung wird in `0.7.0` ausgeliefert; dafür braucht das Lastenheft eine eigene §13.9-Zielsektion mit fünf neuen RAKs (RAK-47..RAK-51), die den Übergang vom Vorbereitungs- zum Lab-Pfad sauber trennen vom `hls.js`-Demo-Pfad in `apps/dashboard`.

DoD:

- [x] Lastenheft Header: Version `1.1.8` → `1.1.9`.
- [x] Lastenheft §13.9 neu: WebRTC-Lab-Erweiterung mit RAK-47..RAK-51 (Compose-Stack, opt-in Smoke `make smoke-webrtc-prep`, `getStats()`-Allowlist + Schema-Drift-Strategie, Browser-Handcheck, optionaler Player-SDK-Adapter-Pfad).
- [x] [`docs/planning/done/plan-0.7.0.md`](./plan-0.7.0.md) §0.2 von „Vorschlag" auf „ausgeliefert in Lastenheft `1.1.9`" hochgezogen (im `0.7.0`-Closeout nach `done/` verschoben).
- [x] [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md) §3 0.7.0-Zeile: Hinweis „Lastenheft-Patch ausstehend" entfernt; RAK-Verteilung referenziert §13.9.

### 4a.13 Patch `1.1.10` — `0.8.0` Player-SDK-WebRTC-Adapter (RAK-51 Hochstufung + RAK-52..RAK-55)

Mit dem `0.7.0`-Release ist die WebRTC-Lab-Erweiterung ausgeliefert (RAK-47..RAK-50; `examples/webrtc/`-Compose, opt-in Smoke `make smoke-webrtc-prep`, Future-Telemetry-Notiz in `spec/telemetry-model.md` §3.5, R-12 als Schema-Drift-Review-Gate). RAK-51 (Player-SDK-WebRTC-Adapter) bleibt in §13.9 „Kann" / deferred. `0.8.0` zieht RAK-51 verbindlich aus dem Kann-Status; der Patch hebt RAK-51 in einer eigenen §13.10 zu „Muss" und ergänzt vier Sub-RAK (RAK-52..RAK-55) für Public-API + hls.js-Trennung, produktive WebRTC-Telemetrie auf bounded Allowlist und Compat-Tests. §13.9 bleibt unverändert als historische Aussage zum `0.7.0`-Stand; ein Hinweis dort verweist auf §13.10.

DoD:

- [x] Lastenheft Header: Version `1.1.9` → `1.1.10`.
- [x] Lastenheft §13.10 neu: `0.8.0` Player-SDK-WebRTC-Adapter mit RAK-51 (Hochstufung) + RAK-52..RAK-55 (Public-API/Pack-Smoke, produktive WebRTC-Telemetrie auf §3.2-Allowlist, `getStats()`-Sammlung mit Schema-Drift-Strategie, opt-in Browser-E2E).
- [x] Lastenheft §13.9 RAK-51-Zeile bekommt einen Hinweis auf die Hochstufung in §13.10; der historische `0.7.0`-Wortlaut bleibt erhalten.
- [x] [`docs/planning/done/plan-0.8.0.md`](./plan-0.8.0.md) §0.2 von „Vorschlag" auf „ausgeliefert in Lastenheft `1.1.10`" hochgezogen; §0.1 Vorgänger-Gate entsprechend nachgezogen (im `0.8.0`-Closeout nach `done/` verschoben).
- [x] [`docs/planning/in-progress/roadmap.md`](../in-progress/roadmap.md) §2 Schritt 40 abgehakt (✅); §1.2 verweist auf den ausgelieferten Patch.

---

## 5. Tranche 1 — MVP `0.1.0` (Backend Core + Demo-Lab)

Status: ✅ abgeschlossen — §5.1 Backend-Erweiterung, §5.2 Compose-Lab, §5.3 RAK-Verifikation, §5.4 CI-Setup und Public-Release-Vorbereitung sind ausgeliefert. Bezug: Lastenheft `1.1.6` §13.1 (RAK-1, RAK-3, RAK-4, RAK-6, RAK-8 für `0.1.0`); Roadmap §2 Schritt 10 (Compose-Lab Core) plus Backend-Erweiterungen aus Lastenheft §7.3.

Player-SDK + Dashboard sind in [`plan-0.1.1.md`](./plan-0.1.1.md), Observability-Stack in [`plan-0.1.2.md`](./plan-0.1.2.md) ausgelagert.

### 5.1 Backend-Erweiterung (`apps/api`)

Bezug: MVP-2, MVP-16, F-17..F-22; OE-3 (Datenhaltung MVP) wird hier entschieden. (MVP-3 = Dashboard wandert nach `0.1.1`, siehe `plan-0.1.1.md`.)

DoD:

- [x] Domain-Aggregation: `StreamSession` wird aus eingehenden `PlaybackEvent`-Batches abgeleitet — bei jedem Event mit unbekanntem `session_id` wird eine `StreamSession` mit Default-State `Active` erzeugt (`9842d39`).
- [x] Session-Lifecycle (`Active` → `Stalled` → `Ended`) als Pflicht — Voraussetzung für F-26 im Dashboard (`0.1.1`). Implementiert mit Stalled-Schwellwert 60 s und Ended-Schwellwert 5 min als Konstanten in `application` (ENV-Tunable folgt bei Lab-Bedarf); zusätzlich expliziter `event_name=session_ended`-Pfad. Background-Sweeper läuft alle 10 s aus `cmd/api`-Goroutine; Lifecycle ist idempotent (`835f258`).
- [x] **MVP-16** Lokale Speicherung der Sessions und Events: In-Memory ist Pflicht-Default; SQLite als Soll-Erweiterung über OE-3-Folge-ADR. Implementierung lebt hinter zwei getrennten Driven Ports (`EventRepository`, `SessionRepository`) — die Trennung erlaubt SQLite-Migration eines der beiden Ports ohne den anderen anzufassen (`9842d39`).
- [x] Neuer Use Case `ListStreamSessions` und `GetStreamSession` als `application.SessionsService` mit typisierten Cursor-Strukturen; Domain-Sicht auf `StreamSession` mit Event-Zählern (LastEventAt, EventCount) (`796aaa7`).
- [x] Zwei neue MVP-Endpoints aus Lastenheft §7.3 (`26a64e2`):
    - `GET /api/stream-sessions` (Liste). Default-Limit 100 Sessions, hartes Maximum 1000 (Query-Parameter `limit`); stabile Sortierung nach `(started_at desc, session_id asc)` — `session_id` als Tie-Breaker bei identischen `started_at`-Werten, damit Cursor-Pagination keine Sessions doppelt liefert oder überspringt. Cursor-basierte Pagination via Query-Parameter `cursor` (opaker Token, kapselt `(started_at, session_id)`-Position).
    - `GET /api/stream-sessions/{id}` (Detail mit Event-Liste). Event-Liste mit Default-Limit 100, hartes Maximum 1000 (Query-Parameter `events_limit`); stabile Sortierung nach `(server_received_at asc, sequence_number asc, ingest_sequence asc)` — `ingest_sequence` als finaler Tie-Breaker. Cursor-basierte Pagination via Query-Parameter `events_cursor`. Hintergrund: Cardinality/Storage-Risiko aus Lastenheft §7.10 — der Endpoint darf nicht unbeschränkt viele Events streamen.
- [x] **`ingest_sequence` als serverseitiges Pflichtfeld** im Domain-Modell `PlaybackEvent`: monoton steigender Counter pro `apps/api`-Prozess via `driven.IngestSequencer`-Port (atomarer InMemoryIngestSequencer-Adapter; SQLite-Migration ersetzt den Adapter, ohne den Use Case anzufassen). Gesetzt im Use Case direkt vor `Append`. Erste Sequence ist 1; nebenläufige Aufrufe ohne Lücken/Duplikate (Concurrent-Test) (`5d6d92c`).
- [x] **Cursor-Invalidierung nach Storage-Restart** (Pflicht für In-Memory-Default): Cursor kapselt eine `process_instance_id` (16-Byte-Hex via crypto/rand, in main.go beim Setup einmalig erzeugt) zusätzlich zu den Sortier-Feldern. Beim Inbound-Request wird die `process_instance_id` aus dem Cursor mit der aktuellen Prozess-`process_instance_id` verglichen — Mismatch → `400 Bad Request` mit Body `{"error":"cursor_invalid","reason":"storage_restart"}`. Tests `TestSessionsService_ListSessions_CursorMismatchInvalidatesPagination` und `TestHTTP_StreamSessions_StaleCursor` decken Use-Case- und HTTP-Pfad ab; `TestHTTP_StreamSessions_Pagination` deckt Cursor-Roundtrip ohne Restart ab. Single-Instance-Annahme weiterhin explizit (siehe Roadmap §4 für Folge-ADR) (`796aaa7`, `26a64e2`).
- [x] **CORS / Origin-Validierung** für Browser-SDK-Anbindung — Variante B implementiert (`c15d8e1`):
    - **F-108 + NF-30** Allowed-Origins pro Project konfigurierbar; Domain-Modell `AllowedOrigin` (oder Erweiterung `Project`); statische Konfiguration analog Spike-Pattern reicht für `0.1.0`.
    - Preflight-Verhalten (kein Body, keine Auth, keine Project-Auflösung):
        - **Player-SDK-Pfad** `OPTIONS /api/playback-events` — Bezug **NF-33 + NF-35 + NF-36**: Origin in globaler Union aller Allowed-Origins → `204 No Content` mit `Access-Control-Allow-Origin: <origin>` (konkret, niemals `*`), `Access-Control-Allow-Methods: POST, OPTIONS` (NF-35), `Access-Control-Allow-Headers: Content-Type, X-MTrace-Project, X-MTrace-Token` (NF-36), `Access-Control-Max-Age ≈ 600`. **Kein** `Access-Control-Allow-Credentials`-Header — credentialless CORS lässt diesen Header weg (NF-31 + NF-32: SDK nutzt `credentials: "omit"`); der einzige interoperable positive Wert wäre `true`, was zu Cookies-Erlaubnis führt und hier explizit unerwünscht ist. Origin nicht in der Union → `403 Forbidden`.
        - **Dashboard-Lese-Pfad** `OPTIONS /api/stream-sessions`, `OPTIONS /api/stream-sessions/{id}`, `OPTIONS /api/health` — Bezug **NF-33 + NF-34 + NF-36** (NF-35 gilt explizit nur für den SDK-Telemetrie-Pfad, Lastenheft §7.11): analog, aber `Access-Control-Allow-Methods: GET, OPTIONS`; ebenfalls **kein** `Access-Control-Allow-Credentials`-Header. Origin nicht in der Union → `403`.
        - **Vary**-Header in jeder Antwort (Preflight wie eigentlicher Request): `Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers` — verhindert falsches Caching durch CDN/Proxy.
    - Echte-Request-Verhalten:
        - **Player-SDK-Pfad** `POST /api/playback-events`: Origin-Validierung erfolgt **nach** Step 3 (Token-Auflösung) und **vor** Step 4 (Rate-Limit), damit ein Origin-Mismatch weder Rate-Limit-Tokens verbraucht noch Events persistiert noch `mtrace_playback_events_total` inkrementiert. Konkreter Ablauf: Step 3 löst das Project auf → Origin wird gegen die Allowed-Origins **dieses Projects** geprüft → Mismatch → `403 Forbidden` ohne Side-Effects. `Origin`-Header **fehlt** (CLI/curl/Lab-Flows): die Project-Bindung wird übersprungen, kein 403 — der Pfad bleibt für `curl`-Smoke-Tests und Headless-Tests offen.
        - **Dashboard-Lese-Pfad** (`GET ...`): keine Project-Token-Bindung (Lese-Pfad ist projektunabhängig im MVP); `Origin`-Validierung gegen globale Union genügt.
    - **NF-31 + NF-32** Hinweis im API-Kontrakt-/Telemetrie-Modell-Doku: SDK nutzt `credentials: "omit"`; keine Cookies werden gesetzt oder erwartet.
    - HTTP-Integrationstests verbindlich:
        1. Preflight `OPTIONS /api/playback-events` mit registriertem Origin → `204` mit konkretem `Access-Control-Allow-Origin: <origin>`, kein `*`.
        2. Preflight `OPTIONS /api/playback-events` mit unbekanntem Origin → `403`.
        3. `POST /api/playback-events` mit gültigem Project-A-Token, aber Origin aus Project-B-Allowlist → `403` (Project↔Origin-Mismatch).
        3a. **Side-Effect-Test**: derselbe Project-↔-Origin-Mismatch-Request darf weder den `EventRepository` (kein `Append`) noch den `RateLimiter` (kein Token-Verbrauch) noch `mtrace_playback_events_total` berühren — Origin-Check liegt vor Step 4.
        4. `POST /api/playback-events` ohne `Origin`-Header (CLI/curl) → unverändert akzeptiert (kein `403`).
        5. Antworten enthalten `Vary: Origin, Access-Control-Request-Method, Access-Control-Request-Headers`.
        6. Keine Antwort enthält `Access-Control-Allow-Origin: *`, sobald ein Project-Token im Spiel ist.
        7. Preflight `OPTIONS /api/stream-sessions` mit registriertem Origin → `204` mit `Access-Control-Allow-Methods: GET, OPTIONS`.
- [x] **Rate-Limit-Dimensionen erweitert (F-110)**: `RateLimiter`-Port nimmt jetzt `RateLimitKey{ProjectID, ClientIP, Origin}` an. Alle drei Dimensionen sind in 0.1.0 implementiert (jenseits der Pflicht-Mindestmenge); leere Werte werden übersprungen, der TokenBucket wendet eine all-or-nothing-Commit-Semantik an. Architektur-Sync mitgezogen: `spec/architecture.md` §3.3 (Driven-Ports-Code-Beispiel) zeigt die strukturierte Port-Signatur, das Sequenzdiagramm in §5.1 nennt Step 3b für Origin-Validierung; API-Kontrakt §6 listet die drei Dimensionen und die all-or-nothing-Semantik (`75e55e7`).
- [x] **F-22** Architektur-Vorbereitung — Port `hexagon/port/driven/StreamAnalyzer` mit konkretem `AnalyzeBatch(ctx, []domain.PlaybackEvent) error`-Vertrag (kein Marker-Interface). Adapter `adapters/driven/streamanalyzer/NoopStreamAnalyzer` mit Compile-Time-Check; Use Case verdrahtet den Slot, ruft die Methode aber nicht produktiv auf — Use-Case-Test verifiziert `analyzer.calls=0` (`2104092`).
- [x] Tests: Use-Case-Test für Session-Aggregation aus Event-Batches (`9842d39` `TestHappyPath`/`TestRepoFailureDoesNotCountAsDropped`) und Lifecycle-Transitions Active → Stalled → Ended (`835f258` `TestInMemorySessionRepository_Sweep_*`); HTTP-Integrationstests für die zwei Stream-Sessions-Endpoints (`26a64e2` `TestHTTP_StreamSessions*`/`TestHTTP_StreamSessionsByID_*`).

### 5.2 Docker-Compose-Lab (Core, `0.1.0`-Anteil)

Bezug: MVP-7..MVP-9, F-82..F-88 (nach Patch `1.0.2`); RAK-1, RAK-4.

Compose-Setup nutzt die Docker-Compose-Profile-Semantik korrekt: Core-Services werden **ohne** `profiles:`-Direktive deklariert und starten damit per Default bei `docker compose up`. Das observability-Profil ist additiv und wird in [`plan-0.1.2.md`](./plan-0.1.2.md) gepflegt; das `dashboard`-Service-Add-On wird in [`plan-0.1.1.md`](./plan-0.1.1.md) gepflegt. In `0.1.0` startet das Core-Lab drei Pflicht-Mindestdienste: `api`, `mediamtx`, `stream-generator`.

DoD:

- [x] `docker-compose.yml` im Repo-Wurzelverzeichnis. Core-Services für `0.1.0` (`api`, `mediamtx`, `stream-generator`) ohne `profiles:`-Direktive — sie starten per Default; entspricht den entsprechenden Pflicht-Mindestdiensten aus Lastenheft §7.8 (nach Patch `1.0.2`) (`504e4c9`).
- [x] MediaMTX als `services/media-server/` mit Konfiguration für HLS (Port `8888`) und HTTP-API/Status (Port `9997`). Beide Ports werden im Compose-Stack exposed; HTTP-API/Status ist Voraussetzung für den `0.1.1`-Dashboard-System-Status-Link (F-40) (`504e4c9`).
- [x] FFmpeg-Generator als `services/stream-generator/` mit Teststream (`jrottenberg/ffmpeg:8.1-ubuntu2404`) (`504e4c9`).
- [x] `apps/api`-Container mit ENV-Variablen-Parametrisierung für die Listen-Adresse (`MTRACE_API_LISTEN_ADDR`). OTel-Exporter-Konfig bleibt im Core-Compose bewusst unset, damit autoexport mit No-Op-Fallback silent bleibt; der OTLP-Endpoint wird erst im `observability`-Profil aus `plan-0.1.2.md` gesetzt (`504e4c9`).
- [x] `make dev` führt `docker compose up --build` ohne Profil-Flag aus — startet damit ausschließlich die Core-Services (`504e4c9`).
- [x] `make stop` beendet sauber (`docker compose down`, Profile-aware) (`504e4c9`).
- [x] Core-Stack mindestens unter Linux verifiziert: `docker compose up -d --build`, `make smoke`, `make stop`, danach `docker compose ps` leer (`504e4c9`).
- [x] Smoke-Test `0.1.0`: nach `make dev` liefert `curl http://localhost:8080/api/health` ein `200`; ein POST mit gültigem Token an `/api/playback-events` liefert `202`; ein GET an `/api/stream-sessions` listet die so erzeugte Session. Zusätzlich prüft `make smoke` das HLS-Manifest via MediaMTX (`504e4c9`).

### 5.3 Release-Akzeptanzkriterien `0.1.0` (Lastenheft `1.1.6` §13.1; RAK-Verteilung aus Patch `1.1.0`)

DoD:

- [x] **RAK-1** `make dev` startet die in `0.1.0` erforderlichen Pflicht-Dienste (`api`, `mediamtx`, `stream-generator`) (Tranche 5.2) (`504e4c9`).
- [x] **RAK-3** API ist erreichbar — `/api/health` liefert `200` im Compose-Stack; alle MVP-Endpoints (drei Spike-Pflicht plus die zwei Stream-Sessions-Endpoints) erreichbar (Tranche 5.1/5.2) (`26a64e2`, `504e4c9`).
- [x] **RAK-4** Teststream läuft über MediaMTX (Tranche 5.2) (`504e4c9`).
- [x] **RAK-6** API nimmt Events an (`POST /api/playback-events` mit gültigem Token liefert `202`) (Tranche 5.1/5.2) (`504e4c9`).
- [x] **RAK-8** README/Local-Development-Doku beschreibt den `0.1.0`-Quickstart reproduzierbar (Initial-Anteil; wird in `0.1.1` und `0.1.2` ergänzt). Bezug Tranche 0a §3.6 (`2eede43`, `504e4c9`).

### 5.4 Übergreifende DoD `0.1.0` (Lastenheft §18, `0.1.0`-Anteil)

Dokumentations- und prozessbezogene Items für den `0.1.0`-Release. RAK-spezifische Items stehen in §5.3.

DoD:

- [x] Architektur in `spec/architecture.md` beschrieben (Tranche 0a §3.1 ausgeliefert; siehe dort für Commit-Liste).
- [x] Eventmodell in `spec/telemetry-model.md` beschrieben (Tranche 0a §3.5) — Pflicht für `0.1.0`, weil das Wire-Format gegen die Spike-API-Kontrakt-Erweiterungen geprüft werden muss (`e532e1e`, `51b3812`).
- [x] Local-Development-Doku in `docs/user/local-development.md` (Tranche 0a §3.6) — Pflicht für RAK-8 (`2eede43`, `504e4c9`).
- [x] Tests für zentrale Use Cases vorhanden — Application-Tests für `RegisterPlaybackEventBatch` (inkl. Tranche-0b-Korrekturen) und neue Session-Use-Cases; HTTP-Integrationstests für alle `0.1.0`-MVP-Endpoints (`7148a8d`, `9842d39`, `796aaa7`, `26a64e2`, `835f258`, `504e4c9`).
- [x] CI führt Build und Tests aus (verknüpft mit OE-6, MVP-32): OE-6 ist für `0.1.0` entschieden als GitHub Actions auf `ubuntu-24.04`; Workflow `.github/workflows/build.yml` läuft auf Push nach `main` und Pull Requests mit `make test`, `make lint`, `make coverage-gate`, `make arch-check`, `make build` (`46e45ec`).
- [x] `CHANGELOG.md` enthält Eintrag für `0.1.0` (Release-Vorgehen siehe `docs/user/releasing.md`) (`95591a5`).

---

## 6. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen (Format ```Item-Beschreibung (`<hash>`)```), gegebenenfalls Sub-Items detaillieren.
- Neue Findings landen entweder als neues `[ ]`-Item in der passenden Tranche oder, wenn architekturrelevant und langfristig, in [`risks-backlog.md`](../open/risks-backlog.md) als `R-X`.
- Beim Schritt-Abschluss: `roadmap.md` §1.2/§2 auf ✅ flippen.
- Nach `0.1.0`-Release: dieses Dokument als historisch archivieren; Folge-Pläne sind [`plan-0.1.1.md`](./plan-0.1.1.md) und [`plan-0.1.2.md`](./plan-0.1.2.md), danach `plan-0.2.0.md`.
