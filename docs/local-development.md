# Local Development — m-trace

> **Status**: Verbindlich für `0.1.x`. Die Quickstart-Sektion wird mit jedem Sub-Release erweitert (`0.1.0` Backend Core, `0.1.1` Player-SDK + Dashboard, `0.1.2` Observability-Stack).  
> **Bezug**: [Lastenheft `1.1.4`](./lastenheft.md) AK-1, AK-2, RAK-8, MVP-7; [Roadmap](./roadmap.md) §2 Schritt 7; [Plan `0.1.0`](./plan-0.1.0.md) §3.6 (Wartung) und §5.2 (Compose-Lab Core); [Architektur](./architecture.md) §8.

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
| Node.js | 20 LTS (siehe `.nvmrc` im Repo-Root) | Build- und Test-Workflows für `apps/dashboard` und `packages/player-sdk`. |
| pnpm | 9 | Workspace-Manager (`pnpm-workspace.yaml`). |

`.nvmrc` und `.npmrc` werden mit dem Mono-Repo-Bootstrap aus `plan-0.1.1.md` §2 verbindlich gepinnt.

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

### 2.1 Erster Start (`0.1.0` Backend Core + Demo-Lab)

```bash
git clone https://github.com/pt9912/m-trace.git
cd m-trace
make dev
```

`make dev` führt `docker compose up --build` ohne Profil-Flag aus. Das startet das Core-Profil (drei Pflicht-Mindestdienste laut Lastenheft §7.8 nach Patch `1.1.1`):

- `api` auf `http://localhost:8080` (`apps/api`)
- `mediamtx` (HLS auf `http://localhost:8888`, HTTP-API/Status auf `http://localhost:9997`)
- `stream-generator` (FFmpeg-Teststream; sendet kontinuierlich an MediaMTX)

Erwartete Wartezeit beim Erst-Pull: 5–10 Minuten (Image-Download + Multi-Stage-Build von `apps/api`).

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
    "sdk": {"name": "@m-trace/player-sdk", "version": "0.1.0"}
  }]
}
JSON
# erwartet: 202 Accepted mit Body {"accepted":1}

# Sessions auflisten (ab 0.1.0 §5.1 implementiert)
curl http://localhost:8080/api/stream-sessions
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

`docker-compose.yml` setzt `depends_on` mit Healthchecks:

- `dashboard` wartet auf `api` (Healthcheck `/api/health` → 200).
- `api` wartet implizit auf nichts (kann ohne `mediamtx` starten).
- `prometheus` wartet auf `api` (Healthcheck) und `otel-collector` (im observability-Profil).
- `grafana` wartet auf `prometheus`.

Der Status pro Service ist über `docker compose ps` und `docker compose logs <service>` einsehbar.

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

Coverage-Tooling (`go test -cover`-Threshold) ist über `make coverage-gate` verfügbar und läuft im `0.1.0`-CI-Workflow.

### 4.2 Player-SDK (`packages/player-sdk`, ab `0.1.1`)

```bash
# vom Repo-Root, nicht aus packages/player-sdk
pnpm -r --filter player-sdk run build
pnpm -r --filter player-sdk run test
pnpm -r --filter player-sdk run lint
```

Alternative über Top-Level-Scripts (Mono-Repo-Bootstrap aus `plan-0.1.1.md` §2):

```bash
pnpm run build      # ruft alle Workspace-Pakete via pnpm -r
pnpm run test
pnpm run lint
```

### 4.3 Dashboard (`apps/dashboard`, ab `0.1.1`)

```bash
pnpm -r --filter dashboard run dev      # Vite-Dev-Mode mit /api/*-Proxy auf localhost:8080
pnpm -r --filter dashboard run build    # Production-Build für Compose-Service
pnpm -r --filter dashboard run check    # SvelteKit-Type-Check
pnpm -r --filter dashboard run test
```

Im Vite-Dev-Mode greift der SvelteKit-Proxy `/api/*` → `http://localhost:8080`, damit Browser-CORS für GET-Routen entfällt (`plan-0.1.1.md` §3 API-Origin-Strategie). Im Compose-Production-Build laufen Dashboard und API über getrennte Origins; CORS-Headers aus `plan-0.1.0.md` §5.1 greifen.

### 4.4 Architektur-Boundary-Check (Backend)

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

### 4.5 RAK-Smoke-Tests

| Smoke | Wann | Aufruf |
|---|---|---|
| `0.1.0`-Smoke (curl-basiert) | nach `make dev` | siehe §2.2 oben. |
| RAK-9-Smoke (Cardinality) | mit aktivem observability-Profil | `make seed-rak9 && bash scripts/check-rak9.sh` (siehe `plan-0.1.2.md` §2 + §4). |
| RAK-10-Smoke (OTel-Spans) | mit `OTEL_TRACES_EXPORTER=console` | siehe `plan-0.1.2.md` §4 Variante A. |

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
