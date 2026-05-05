# Local Development — m-trace

> **Status**: Verbindlich für `0.1.x`. Die Quickstart-Sektion wird mit jedem Sub-Release erweitert (`0.1.0` Backend Core, `0.1.1` Player-SDK + Dashboard, `0.1.2` Observability-Stack).  
> **Bezug**: [Lastenheft `1.1.6`](../../spec/lastenheft.md) AK-1, AK-2, RAK-8, MVP-7; [Roadmap](../planning/in-progress/roadmap.md) §2 Schritt 7; [Plan `0.1.0`](../planning/done/plan-0.1.0.md) §3.6 (Wartung) und §5.2 (Compose-Lab Core); [Architektur](../../spec/architecture.md) §8.

## 0. Zweck

Quickstart-Doku für ein neues Entwickler-Setup: Repo klonen, Voraussetzungen installieren, lokales Lab starten, häufige Test-/Lint-/Build-Workflows kennen. Erfüllt RAK-8 („README beschreibt den Ablauf reproduzierbar"); ergänzt das `README.md` um Detail-Schritte, die zu lang für den README-Quickstart sind.

Leitprinzip: ein neuer Entwickler ist innerhalb von 15 Minuten von `git clone` zu einem laufenden lokalen Lab — ohne lokale Go-Toolchain, ohne tiefe System-Konfiguration, ohne manuelle Sonderfälle (AK-2).

## 1. Voraussetzungen

### 1.1 Pflicht-Tooling

| Tool | Mindest-Version | Zweck |
|---|---|---|
| Docker Engine | 24.0 | Container-Runtime; alle Build-/Test-/Lint-Targets laufen über `docker build` (Docker-only-Workflow laut `apps/api/README.md`). |
| Docker Compose | v2.20 | Lokales Lab via `docker compose up`. |
| GNU Make | 3.81 (macOS-Default) | Wrapper für die Compose-/Build-Targets. |
| Git | 2.30 | Klonen und Branch-Operationen. |

Lokale Go-Toolchain ist **nicht erforderlich** — der `golang:1.22`-Container in `apps/api/Dockerfile` enthält die Toolchain.

Ab `0.1.1` zusätzlich (Player-SDK + Dashboard sind TypeScript-Pakete):

| Tool | Mindest-Version | Zweck |
|---|---|---|
| Node.js | 22 (siehe `.nvmrc` im Repo-Root) | Build- und Test-Workflows für `apps/dashboard` und `packages/player-sdk`. |
| pnpm | 10 (siehe `package.json` `packageManager`) | Workspace-Manager (`pnpm-workspace.yaml`). |

`.nvmrc`, `.npmrc`, `package.json`, `pnpm-workspace.yaml` und `pnpm-lock.yaml` sind im Repo-Root versioniert.

### 1.2 Plattform-Hinweise

**Linux**: nativ unterstützt; Docker Engine direkt installieren (Docker-Engine-Repos bevorzugt vor Distro-Paketen).

**macOS**: Docker Desktop ≥ 4.x; in den Settings → Resources mindestens 4 CPUs und 4 GB RAM zuweisen, sonst werden die Multi-Stage-Builds in `apps/api/Dockerfile` spürbar langsam.

**Windows**: WSL2 + Docker Desktop. Nicht direkt nativ unterstützt. Empfehlung: Repository unter dem WSL2-Filesystem (`\\wsl$`) klonen, nicht im Windows-Filesystem über `/mnt/c/` — letzteres bremst Docker-Build-Layer-Caching auf wenige MB/s.

### 1.3 Optionale Helfer

| Tool | Zweck |
|---|---|
| `jq` | JSON-Antworten der API auf der Kommandozeile filtern. |
| `httpie` (`http`) | bequemere HTTP-Aufrufe gegen die API als reines `curl`. |
| `direnv` | automatisches Setzen der Compose-Profile-Env-Vars pro Repo-Wechsel. |

---

## 2. Quickstart

### 2.1 Erster Start (`0.1.1` Core-Lab)

```bash
git clone https://github.com/pt9912/m-trace.git
cd m-trace
make dev
```

`make dev` führt `docker compose up --build` ohne Profil-Flag aus. Das startet das Core-Profil (vier Pflicht-Mindestdienste ab `0.1.1` laut Lastenheft §7.8):

- `api` auf `http://localhost:8080` (`apps/api`)
- `dashboard` auf `http://localhost:5173` (`apps/dashboard`)
- `mediamtx` (HLS auf `http://localhost:8888`, HTTP-API/Status auf `http://localhost:9997`)
- `stream-generator` (FFmpeg-Teststream; sendet kontinuierlich an MediaMTX)

Erwartete Wartezeit beim Erst-Pull: 5–10 Minuten (Image-Download + Multi-Stage-Builds von `apps/api` und `apps/dashboard`).

### 2.2 Smoke-Test

Nach erfolgreichem Start kann der API-Pfad direkt geprüft werden:

```bash
# Health-Check
curl http://localhost:8080/api/health
# erwartet: 200 OK mit Body {"status":"ok"}

# Event senden (gültiger Spike-Token "demo-token")
curl -i -X POST http://localhost:8080/api/playback-events \
  -H 'Content-Type: application/json' \
  -H 'X-MTrace-Token: demo-token' \
  --data-binary @- <<'JSON'
{
  "schema_version": "1.0",
  "events": [{
    "event_name": "rebuffer_started",
    "project_id": "demo",
    "session_id": "smoke-test-1",
    "client_timestamp": "2026-04-29T10:00:00.000Z",
    "sequence_number": 1,
    "sdk": {"name": "@npm9912/player-sdk", "version": "0.2.0"}
  }]
}
JSON
# erwartet: 202 Accepted mit Body {"accepted":1}

# Sessions auflisten (ab 0.1.0 §5.1 implementiert; ab 0.4.0 §4.3
# tokenpflichtig)
curl -H 'X-MTrace-Token: demo-token' \
  http://localhost:8080/api/stream-sessions
# erwartet: 200 OK mit JSON-Array, das die Session "smoke-test-1" enthält

# Prometheus-Counter abfragen
curl http://localhost:8080/api/metrics | grep ^mtrace_
# erwartet: alle vier Pflicht-Counter sichtbar
```

Stream-Auslieferung lokal (HLS):

```bash
# HLS-Manifest des Teststreams
curl -L http://localhost:8888/teststream/index.m3u8
# erwartet: 200 OK mit HLS-Manifest
```

Analyzer-Endpunkt (ab `0.3.0` Tranche 6):

```bash
# Master-Manifest analysieren (Text-Input, exerciert API → analyzer-service)
curl -i -X POST http://localhost:8080/api/analyze \
  -H 'Content-Type: application/json' \
  --data-binary @- <<'JSON'
{
  "kind": "text",
  "text": "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1280000\nvideo/720p.m3u8\n"
}
JSON
# erwartet: 200 OK mit { status:"ok", playlistType:"master", … }

# Smoke-Lauf inkl. Stack-Up und SSRF-Negativfall
make smoke-analyzer
```

URL-Inputs gegen interne Compose-Services funktionieren im Lab-
Setup, weil `docker-compose.yml` den `analyzer-service` mit
`ANALYZER_ALLOW_PRIVATE_NETWORKS=true` startet (`docs/user/stream-
analyzer.md` §6). Außerhalb der Compose-Topologie greift der
SSRF-Block per Default und URL-Inputs auf RFC1918-Adressen werden
mit 400 `fetch_blocked` abgelehnt — dann Text-Input verwenden oder
eine öffentliche HLS-URL.

CLI (ab `0.3.0` Tranche 7):

```bash
# Build des CLI-Bundles
make ts-build

# Datei oder URL prüfen
pnpm --silent m-trace check ./packages/stream-analyzer/tests/fixtures/master.m3u8
pnpm --silent m-trace check https://cdn.example.test/manifest.m3u8

# Smoke (--help, Master-Datei, Nicht-HLS, fehlende Datei)
make smoke-cli
```

`pnpm --silent` unterdrückt das pnpm-Skript-Banner; ohne `--silent`
landet vor dem Analyzer-JSON eine pnpm-Statuszeile auf stdout, die
beim Pipen in `jq` stören würde.

### 2.3 Stack erweitern (`0.1.1` Dashboard, `0.1.2` Observability)

Ab `0.1.1` kommt der `dashboard`-Service ins Core-Profil (vier Pflicht-Mindestdienste); `make dev` startet ihn automatisch. Erreichbar unter `http://localhost:5173` (oder Compose-equivalent).

Ab `0.1.2` ist der Observability-Stack (Prometheus, Grafana, OTel-Collector) als optionales Profil verfügbar:

```bash
make dev-observability
# entspricht: docker compose --profile observability up --build
# zusätzlich:
#   prometheus auf http://localhost:9090
#   grafana    auf http://localhost:3000  (Default-Login admin/admin, dann Bonus-Dashboard)
#   otel-collector ohne Web-UI; Logs via docker compose logs otel-collector
```

Ohne Profil-Flag bleibt der Stack im Core-Modus (`make dev`).

### 2.4 Beenden

```bash
make stop
# entspricht: docker compose down
# Profile-aware: beendet auch das observability-Profil, falls aktiv.
```

### 2.5 Tempo (optional, ab `0.4.0`)

Tempo ist ein **optionales** Trace-Backend für Debug-Tiefe (RAK-31,
Kann-Scope) und kein Ersatz für die eingebaute Session-Timeline
(RAK-32). Die Dashboard-Timeline und alle Read-Pfade
(`GET /api/stream-sessions/...`) sind Tempo-unabhängig — Source-of-
Truth ist SQLite (ADR-0002). Tempo erweitert die Sicht auf Span-Ebene
(Header-Verarbeitung, Outcome-Klassifikation, Resource-Attribute).

```bash
make dev-tempo
# entspricht: docker compose --profile observability --profile tempo up --build
# zusätzlich:
#   tempo auf http://localhost:3200  (HTTP-API für Trace-Search)
#   Collector lädt observability/otel-collector/config-tempo.yaml und
#   exportiert Traces zusätzlich an `otlp/tempo` (sonst nur `debug`).
```

Trace-Suche im Lab nutzt Span-Attribute aus
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §2.6:

- **Primary**: Session-bezogen über
  `mtrace.session.correlation_id` (gesetzt nur bei Single-Session-
  Batches). Tempo-Search primär via TraceQL mit explizitem Zeitfenster:
  `GET /api/search?q={ span.mtrace.session.correlation_id = "<UUID>" }&start=<unix>&end=<unix>`.
  `GET /api/search?tags=mtrace.session.correlation_id=<UUID>&start=<unix>&end=<unix>`
  ist nur Legacy-Fallback.
- **Sekundär**: batchspezifisch über `trace_id`. Eine Session kann
  mehrere `trace_id`-Werte haben (jeder Batch ein Trace) — `trace_id`
  ist daher kein Session-Schlüssel.

Smoke: `make smoke-tempo` (Default-State `tempo`) postet einen Single-
Session-Batch, liest `correlation_id` über die API zurück und
verifiziert den Tempo-Trace-Treffer. Andere Stack-Zustände
(`SMOKE_STATE=core` ohne OTLP-Versuch, `SMOKE_STATE=observability`
ohne Tempo-Verbindungsversuch) sind eigene Smoke-Aufrufe — siehe
`scripts/smoke-tempo.sh`.

`make stop` und `make wipe` adressieren das `tempo`-Profil mit; `wipe`
entfernt das benannte Volume `mtrace-tempo-data` zusätzlich zum
SQLite-Volume.

### 2.6 Korrelations-Identifier in Read-Antworten

Ab `0.4.0` (Tranche 3) tragen Session- und Event-Read-Antworten zwei
unabhängige Korrelations-Identifier (Spec
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §2.5):

- **`correlation_id`** ist der **Pflichtkontext** der Session-
  Timeline. Server-vergeben, durable, konstant über alle Events
  derselben Session. Dashboard und API-Konsumenten korrelieren
  ausschließlich über diesen Wert.
- **`trace_id`** ist eine **optionale Debug-Vertiefung** für
  Tempo/Jaeger. Standard ohne `traceparent`-Header: batch-bezogen —
  der Server erzeugt pro POST-Batch einen Root-Span, also tragen
  Events verschiedener Batches einer Session unterschiedliche
  `trace_id`s. Mit propagiertem `traceparent`-Header übernimmt der
  Server dagegen die `trace_id` aus dem Header und sie bleibt über
  alle Batches mit demselben Header konstant (Test-Anker
  `TestE2E_TraceparentPropagation_SameTraceID` in
  `apps/api/adapters/driving/http/correlation_e2e_test.go`).

Bekannte Korrelations-Lücken (Browser-APIs, CORS, Service Worker,
CDN-Redirects, Native HLS, Sampling) sind in
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §1.4.1
dokumentiert und werden über `network.detail_status=
network_detail_unavailable` (Event-Meta) bzw. `network_signal_absent`
(Session-Block) sichtbar gemacht.

### 2.7 Multi-Protocol-Lab-Beispiele (ab `0.5.0`)

`examples/` bündelt protokollspezifische Lab-Setups —
Konventionen, Project-Name-Pflicht und Smoke-Skript-Regeln stehen in
[`examples/README.md`](../../examples/README.md). Jedes Beispiel hat
einen opt-in Smoke und ist nicht Teil von `make gates`.

| Beispiel | Variante | Start | Smoke |
|---|---|---|---|
| [`examples/mediamtx/`](../../examples/mediamtx/) | Core-Lab-Beispiel | `make dev` | `make smoke-mediamtx` |
| [`examples/srt/`](../../examples/srt/) | Eigenes Compose, Project `mtrace-srt` | `docker compose -p mtrace-srt -f examples/srt/compose.yaml up -d --build` | `make smoke-srt` (auto-up/down) plus `make smoke-srt-health` ab `0.6.0` für Health-API-Probe |
| [`examples/dash/`](../../examples/dash/) | Eigenes Compose, Project `mtrace-dash` | `docker compose -p mtrace-dash -f examples/dash/compose.yaml up -d --build` | `make smoke-dash` (auto-up/down) |
| [`examples/webrtc/`](../../examples/webrtc/) | Doku-only Vorbereitungspfad | — | — (kein Smoke in `0.5.0`; siehe `examples/webrtc/README.md` „Folge-Pfad") |

Host-Ports sind so geschnitten, dass alle Beispiele **parallel** zum
Core-Lab und untereinander laufen können:

| Stack | Ports |
|---|---|
| Core-Lab (`make dev`) | `8080`/`5173`/`8888`/`9997` (API, Dashboard, MediaMTX-HLS, MediaMTX-Control) |
| `mtrace-srt` | `8889/tcp` (HLS-Out) · `8890/udp` (SRT-Listener) · `9998/tcp` (MediaMTX-Control) |
| `mtrace-dash` | `8891/tcp` (DASH-HTTP) |
| `mtrace-webrtc` (geplant) | `8889/tcp` — kollidiert in `0.5.0` mit `mtrace-srt`; eine Folge-Tranche muss den Schnitt neu vornehmen |

Reset pro Beispiel-Stack: `docker compose -p <project> -f
examples/<name>/compose.yaml down [--volumes]`. Greift nur den
jeweiligen Project-Namen — andere Stacks (Core-Lab, andere Beispiele)
bleiben unangetastet.

### 2.7.1 SRT-Health-View (`0.6.0`)

`0.6.0` ergänzt einen lokalen SRT-Verbindungs-Health-Pfad:

- **Lab-Smoke** `make smoke-srt-health` (verifiziert HLS + MediaMTX-API
  `/v3/srtconns/list` + vier RAK-43-Pflichtwerte).
- **Collector** im API-Prozess (opt-in via `MTRACE_SRT_SOURCE_URL`).
- **Read-Endpoints** `GET /api/srt/health[/{stream_id}]`.
- **Dashboard-Route** `/srt-health` mit Tabelle + Mini-Timeline.

Vollständige Operator-Doku (Metriken, Schwellen, Fehlerbilder,
Cardinality-Grenzen, Deferred-Liste): [`srt-health.md`](./srt-health.md).

---

## 3. Compose-Stack-Topologie

### 3.1 Service-Übersicht (Stand `0.1.x`-Endzustand)

| Service | Pflicht? | Port(s) | Zweck |
|---|---|---|---|
| `api` | Pflicht | 8080 | Backend-API (`apps/api`); HTTP-Endpoints + Prometheus-Scrape-Target. |
| `dashboard` | Pflicht ab `0.1.1` | 5173 | SvelteKit Web-UI. |
| `mediamtx` | Pflicht | 8888 (HLS), 9997 (HTTP-API/Status) | Lokaler Media-Server. |
| `stream-generator` | Pflicht | — | FFmpeg-Teststream → MediaMTX. |
| `prometheus` | Soll, observability-Profil | 9090 | Aggregat-Metriken, scraped `api:8080/api/metrics`. |
| `grafana` | Soll, observability-Profil | 3000 | Visualisierung der Pflicht-Counter. |
| `otel-collector` | Soll, observability-Profil | OTLP gRPC 4317, OTLP HTTP 4318 | Empfängt OTel-Spans/-Counter aus `api`. |

CSP-Beispiele für `connect-src` (NF-37):

- Dashboard-Auslieferung im Core-Stack: `Content-Security-Policy: default-src 'self'; connect-src 'self' http://localhost:8080`.
- Mit aktivem observability-Profil ergänzen sich Prometheus und Grafana — die Status-Ansicht des Dashboards verlinkt diese als Footer-Links (F-40), nicht als `connect-src`-Targets.

### 3.2 Inter-Service-Konfiguration

| Variable | Konsument | Wert (Compose-Default) | Bemerkung |
|---|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `api` | `http://otel-collector:4317` (im observability-Profil), sonst leer | Aktiviert OTLP-Export erst, wenn observability-Profil läuft. |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `api` | `grpc` | Protokoll-Wahl, siehe `architecture.md` §5.3. |
| `OTEL_TRACES_EXPORTER` / `OTEL_METRICS_EXPORTER` | `api` | unset (Default → autoexport mit No-Op-Fallback, silent) bzw. `console` für RAK-10-Smoke-Test | siehe `plan-0.1.2.md` §4. |
| `MEDIAMTX_URL` | `dashboard` | `http://mediamtx:8888` | Dashboard-Demo-Route lädt das HLS-Manifest. |

### 3.3 Service-Abhängigkeiten

`docker-compose.yml` setzt einfache `depends_on`-Beziehungen. Compose wartet damit auf gestartete Container (`service_started`), nicht auf applikative Healthchecks:

- `dashboard` startet nach `api` und `mediamtx`.
- `api` wartet implizit auf nichts (kann ohne `mediamtx` starten).
- `prometheus` startet nach `api` und `otel-collector` (im observability-Profil).
- `grafana` startet nach `prometheus`.

Der Status pro Service ist über `docker compose ps` und `docker compose logs <service>` einsehbar.

### 3.4 SQLite-Persistenz und Reset

Ab `0.4.0` persistiert der `api`-Service Sessions, Playback-Events und
Ingest-Sequenzen durable in einer SQLite-Datei (siehe
[ADR-0002](../adr/0002-persistence-store.md) §8.1). Konfiguration und
Reset-Pfad:

| Aspekt | Default im Compose-Lab | Override |
|---|---|---|
| Persistenz-Modus | `MTRACE_PERSISTENCE=sqlite` | `inmemory` für Tests/Restart-flüchtige Sessions |
| SQLite-Pfad | `MTRACE_SQLITE_PATH=/var/lib/mtrace/m-trace.db` | beliebiger Pfad (CI: `t.TempDir()`) |
| Volume-Name | `mtrace-data` (compose-internes Naming: `<project>_mtrace-data`) | nicht überschreibbar im Default-Compose |

Nach jedem `make dev`-Cycle bleibt die SQLite-Datei im benannten
Volume erhalten — der nächste Start liest die gleiche Session-Historie.
`make stop` (= `docker compose down` ohne `--volumes`) lässt das Volume
unangetastet.

**Storage-Retention in `0.4.0`**: „unlimited mit dokumentiertem
Reset-Pfad" gemäß [ADR-0002](../adr/0002-persistence-store.md) §8.4.
Sessions, Playback-Events und Ingest-Sequenzen wachsen ungebunden, bis
der Reset-Pfad ausgeführt wird. Konkrete Retention-Werte (Zeitfenster,
Pro-Projekt-Limit) sind ausdrücklich für eine Folge-Phase deferred —
Schema und Indizes (`occurred_at`, `started_at`, `project_id`,
`session_id`) sind so gewählt, dass eine spätere Retention-Implementierung
ohne erneute Migration ansetzen kann.

**Reset (destruktiv)**: ausschließlich über

```bash
make wipe
```

Das Target stoppt alle Services und entfernt gezielt das
`mtrace-data`-Volume. Sessions, Events, Cursor-States und
Ingest-Sequenz sind danach weg; der nächste `make dev`-Start migriert
ein leeres Schema. Andere Reset-Wege (manuelles `docker volume rm`,
Filesystem-Eingriffe) sind nicht Teil des Vertrags.

**Cursor-Recovery**: Cursor-v2 (Wire-Format aus
[ADR-0004](../adr/0004-cursor-strategy.md)) trägt nur durable
Storage-Werte (kein `process_instance_id` mehr) und bleibt nach
API-Restart gültig. Treten dennoch Cursor-Fehler auf, mappt der
Server folgende Klassen (siehe API-Kontrakt §10.3):

| HTTP | `error` | Wann | Client-Handlung |
|---|---|---|---|
| 400 | `cursor_invalid_legacy` | `0.1.x`/`0.2.x`/`0.3.x`-Cursor mit `pid`-Feld | Cursor verwerfen, Snapshot ohne `cursor` neu laden — kein Retry-Loop |
| 400 | `cursor_invalid_malformed` | Decode-/Schema-Verletzung | dito |
| 410 | `cursor_expired` | Storage-Position weg (nach `make wipe`) | dito |

Kein `Retry-After`-Header — Recovery ist deterministisch durch
Snapshot-Reload.

**Migrationen**: das initiale Schema-DDL liegt unter
`apps/api/internal/storage/migrations/V1__m_trace.sql` (aus
`schema.yaml` per d-migrate generiert). Beim Container-Start wendet
der eingebettete Apply-Runner offene Migrationen an; ein bereits
applizierter Schema-State ist no-op. Schlägt eine Migration fehl,
markiert der Runner den State als `dirty` und der nächste Start
weigert sich (`storage: schema is in dirty state`). Reparatur:
`make wipe` und neu starten.

### 3.5 Prometheus-Aggregate (Quickref)

Prometheus ist Aggregat-Backend, nicht Per-Session-Store. Die drei
Backends teilen die Verantwortung wie folgt (kanonische 3-Spalten-
Tabelle:
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §3.3):

| Backend | Rolle | Lab-Endpoint |
|---|---|---|
| Prometheus | Aggregat-Metriken (Counter, Rates) mit bounded Aggregat-Labels | `http://localhost:9090` (im observability-Profil) |
| SQLite | Per-Session-Historie (Sessions, Events, Cursor, `session_boundaries`) | `/var/lib/mtrace/m-trace.db` im API-Container; gelesen über `GET /api/stream-sessions/...` |
| OTel/Tempo | Per-Request-Trace-Spans für Debug-Tiefe | `http://localhost:3200` (im `tempo`-Profil — siehe §2.5) |

Die vier Pflichtcounter aus
[`spec/backend-api-contract.md`](../../spec/backend-api-contract.md) §7
und der OTel-translated `mtrace_api_batches_received` sind ab `0.4.0`
Tranche 7 alle **label-frei** (kein `batch_size`, kein `session_id`,
kein `project_id`); andere `mtrace_*`-Metriken dürfen ausschließlich
bounded Aggregat-Labels aus
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §3.2 tragen
(`event_type`, `outcome`, `code`, `instance`/`job`). Die normative
Forbidden-Liste über alle Metriken steht in
[`telemetry-model.md`](../../spec/telemetry-model.md) §3.1; der
verschärfte Cardinality-Smoke (`make smoke-observability`) prüft sie
release-blockierend.

Hochkardinale Werte (`session_id`, `correlation_id`, URL, User-Agent,
Token, …) gehören grundsätzlich **nicht** in Prometheus-Labels — sie
landen in SQLite (durable Read-Pfad) und/oder Tempo-Spans (sample-
basiertes Debug-Backend).

---

## 4. Test-/Lint-/Coverage-Workflows

Alle Checks laufen Docker-only — keine lokale Go-/Node-Toolchain erforderlich.

### 4.1 Backend (`apps/api`)

```bash
cd apps/api

make test           # docker build --target test  (go test ./...)
make lint           # docker build --target lint  (golangci-lint, Default-Linter)
make build          # docker build --target runtime  (Distroless-Final-Image)
make arch-check     # Boundary-Test: hexagon/ darf keine Adapter/OTel/Prometheus importieren

make compile        # schneller go-build-Feedback-Loop
make deps           # nur dependencies auflösen (Cache-Layer)
make clean          # entfernt alle apps/api-spike:*-Images
```

Coverage-Tooling ist über `make coverage-gate` verfügbar. Das Root-Target
prüft die API-Coverage per Docker sowie Player-SDK- und Dashboard-Coverage
per Vitest.

### 4.2 Player-SDK (`packages/player-sdk`, ab `0.1.1`)

```bash
# vom Repo-Root, nicht aus packages/player-sdk
pnpm --filter @npm9912/player-sdk run build
pnpm --filter @npm9912/player-sdk run test
pnpm --filter @npm9912/player-sdk run test:coverage
pnpm --filter @npm9912/player-sdk run performance:smoke
pnpm --filter @npm9912/player-sdk run lint
```

Der Performance-Smoke baut das SDK und prüft das `0.2.0`-Budget
für Bundle-Größe, Event-Hot-Path und Queue-/Retry-Grenzen.

Alternative über Top-Level-Scripts (Mono-Repo-Bootstrap aus `plan-0.1.1.md` §2):

```bash
pnpm run build      # ruft alle Workspace-Pakete via pnpm -r
pnpm run test
pnpm run lint
```

### 4.3 Dashboard (`apps/dashboard`, ab `0.1.1`)

```bash
pnpm --filter @npm9912/m-trace-dashboard run dev      # Vite-Dev-Mode mit /api/*-Proxy auf localhost:8080
pnpm --filter @npm9912/m-trace-dashboard run build    # Production-Build für Compose-Service
pnpm --filter @npm9912/m-trace-dashboard run check    # SvelteKit-Type-Check
pnpm --filter @npm9912/m-trace-dashboard run test
pnpm --filter @npm9912/m-trace-dashboard run test:coverage
```

Im Vite-Dev-Mode greift der SvelteKit-Proxy `/api/*` → `http://localhost:8080`, damit Browser-CORS für GET-Routen entfällt (`plan-0.1.1.md` §3 API-Origin-Strategie). Im Compose-Production-Build laufen Dashboard und API über getrennte Origins; CORS-Headers aus `plan-0.1.0.md` §5.1 greifen.

### 4.4 Gepacktes SDK gegen Dashboard/API testen (`0.2.0`)

Der lokale Veröffentlichungs-Sanity-Check trennt Paketierung und
End-to-End-Integration:

```bash
pnpm --filter @npm9912/player-sdk run pack:smoke
make dev
```

`pack:smoke` baut das SDK, erzeugt ein Tarball-Artefakt, installiert es in ein
temporäres Beispielprojekt und prüft ESM, CJS und den Browser/IIFE-Einstieg.
Damit ist abgesichert, dass das Paket lokal installierbar ist.

Das laufende Lab nutzt denselben Package-Entry-Point im Dashboard-Build. Die
Demo-Route testet das SDK gegen die echte lokale API:

```text
http://localhost:5173/demo?session_id=local-pack-demo&autostart=1
```

Nach einigen Sekunden kann die erzeugte Session geprüft werden:

```bash
curl -H 'X-MTrace-Token: demo-token' \
  http://localhost:8080/api/stream-sessions/local-pack-demo/events
```

Erwartet sind Playback-Events aus der Demo und nach `Stop` bzw. beim Verlassen
der Route ein `session_ended` Event. Die Details der Beispielintegration stehen
in [`docs/user/demo-integration.md`](./demo-integration.md).

### 4.5 Architektur-Boundary-Check (Backend)

```bash
cd apps/api
make arch-check
```

Prüft nach `apps/api/scripts/check-architecture.sh`, dass:

- `hexagon/` keine direkten Imports auf Adapter, OTel, Prometheus, `database/sql`, `net/http` zieht.
- `hexagon/domain/` keine Application-/Port-Imports zieht.
- `hexagon/application/` keine Adapter-Imports zieht.
- `hexagon/port/` keine Adapter-Imports zieht.

Schlägt der Test fehl, listet er den verstoßenden Import explizit. Verbindlich vor jedem PR.

### 4.6 RAK-Smoke-Tests

| Smoke | Wann | Aufruf |
|---|---|---|
| `0.1.0`-Smoke (curl-basiert) | nach `make dev` | siehe §2.2 oben. |
| `0.1.1`-Browser-E2E (Playwright im Container) | ohne laufenden Stack; Script startet API, MediaMTX, FFmpeg und `dashboard-e2e` selbst | `make browser-e2e` |
| RAK-9-Smoke (Cardinality) | mit aktivem observability-Profil | `make smoke-observability` (siehe `plan-0.1.2.md` §2 + §4). |
| RAK-10-Smoke (OTel-Spans) | startet `api` mit `OTEL_TRACES_EXPORTER=console` | `make smoke-rak10-console` |

---

## 5. Häufige Probleme und Workarounds

### 5.1 Port-Konflikte

`make dev` schlägt fehl mit `bind: address already in use` auf `8080`/`5173`/`8888`/`3000`/`9090`.

**Ursache**: lokal laufende Services (z. B. eine andere App auf 8080, ein lokaler Grafana-Server).

**Workaround**: konfliktende Services lokal stoppen, oder Compose-Override-Datei mit alternativen Ports anlegen (`docker-compose.override.yml`).

### 5.2 macOS-Docker-Resource-Limits

Multi-Stage-Builds in `apps/api/Dockerfile` brauchen mindestens 2 GB RAM für das Build-Image. Bei Standard-Docker-Desktop-Defaults (oft 2 GB RAM für die VM) bricht der Build mit OOM ab.

**Workaround**: Docker Desktop → Settings → Resources → Memory auf ≥ 4 GB heben.

### 5.3 WSL2-Filesystem-Performance

Build-Targets unter `/mnt/c/...` laufen auf wenige MB/s wegen des Windows↔Linux-Filesystem-Overhead.

**Workaround**: Repository unter `\\wsl$` klonen (`/home/<user>/m-trace` im WSL2-Linux-Filesystem), nicht im Windows-Pfad.

### 5.4 MediaMTX-Stream startet nicht

`curl http://localhost:8888/teststream/index.m3u8` liefert `404` oder leeres Manifest.

**Diagnose**:

```bash
docker compose logs stream-generator | tail -20
docker compose logs mediamtx       | tail -20
```

Häufige Ursachen:

- FFmpeg-Generator startet vor MediaMTX bereit ist → Restart nach 5–10 s reicht (`docker compose restart stream-generator`).
- MediaMTX-Konfiguration in `services/media-server/mediamtx.yml` setzt `protocols`-Allowlist nicht — beim Default sollten RTSP/RTMP/HLS aktiv sein.

### 5.5 OTel-Endpoint nicht erreichbar

`apps/api`-Logs zeigen `connect: connection refused` für OTLP-Export, obwohl `make dev` ohne observability-Profil läuft.

**Ursache**: `OTEL_EXPORTER_OTLP_ENDPOINT` ist in der Compose-Default-Config gesetzt, der Collector ist aber im opt-in-Profil.

**Workaround**: ohne observability-Profil bleibt das Setup silent — `OTEL_EXPORTER_OTLP_ENDPOINT` darf in der Core-Compose-Config **nicht** gesetzt sein. Der `autoexport`-Fallback (siehe `architecture.md` §5.3) liefert dann No-Op-Reader/-Exporter, kein Verbindungsversuch. Wenn das Setup bei dir trotzdem zu Push-Versuchen führt, prüfe lokale `OTEL_*`-Env-Vars (z. B. aus `direnv` oder Shell-Profil) und unset sie für die Core-Variante.

### 5.6 Kommandos hängen ohne sichtbare Ausgabe

Lange `docker build`-Stages liefern keine Zwischenausgabe, wenn `--progress=plain` nicht gesetzt ist.

**Workaround**: für Diagnose explizit `docker build --progress=plain ...` aufrufen oder `BUILDKIT_PROGRESS=plain make build` setzen — zeigt jeden Layer-Build-Schritt einzeln.
