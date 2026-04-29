# Local Development — m-trace

> **Status**: Skeleton — Inhalts-Sections sind als Platzhalter angelegt und werden im Zuge der `0.1.0`-Phase (Compose-Lab Core) befüllt.  
> **Bezug**: [Lastenheft `1.1.3`](./lastenheft.md) AK-1, AK-2, RAK-8, MVP-7; [Roadmap](./roadmap.md) §2 Schritt 7; [Plan `0.1.0`](./plan-0.1.0.md) §3.6 (Wartung) und §5.2 (Compose-Lab Core).

## 0. Zweck

Quickstart-Doku für ein neues Entwickler-Setup: Repo klonen, Voraussetzungen installieren, lokales Lab starten, häufige Test-/Lint-/Build-Workflows kennen. Erfüllt RAK-8 („README beschreibt den Ablauf reproduzierbar"); ergänzt das `README.md` um Detail-Schritte, die zu lang für den README-Quickstart sind.

## 1. Voraussetzungen

> **Status: TODO** — wird mit `0.1.0` Compose-Lab Core befüllt.

Zu dokumentieren:

- Linux: Docker Engine ≥ 24, Docker Compose v2, GNU Make.
- macOS: Docker Desktop ≥ 4.x; Hinweis zu Resource-Allocation (CPUs, RAM).
- Windows: WSL2 + Docker Desktop; nicht direkt nativ unterstützt.
- Lokale Go-Toolchain ist **nicht** erforderlich (Docker-only-Workflow gemäß `apps/api/README.md`).

## 2. Quickstart

> **Status: TODO** — wird mit `0.1.0` Compose-Lab Core befüllt.

Zu dokumentieren (Soll-Kommandos):

```bash
git clone https://github.com/pt9912/m-trace.git
cd m-trace
make dev               # startet Core-Profil (api, mediamtx, stream-generator)
# (in 0.1.1) make dev             # zusätzlich dashboard
# (in 0.1.2) make dev-observability  # additiv otel-collector, prometheus, grafana
make stop              # beendet sauber
```

Plus Hinweise zu:

- Erstmaligem Pull-Vorgang (Image-Größen, Cache-Layer).
- Erwarteten Service-Endpoints und Ports (siehe §3).
- Smoke-Test: `curl http://localhost:8080/api/health` → `200`; `curl -H 'X-MTrace-Token: demo-token' -d '{"schema_version":"1.0","events":[]}' http://localhost:8080/api/playback-events` → `422` (leerer Batch).

## 3. Compose-Stack-Topologie

> **Status: TODO** — wird mit `0.1.0`/`0.1.1`/`0.1.2` befüllt.

Zu dokumentieren:

- Services pro Sub-Release (`0.1.0`: api+mediamtx+stream-generator; `0.1.1`: + dashboard; `0.1.2`: + observability-Profil).
- Ports und Zwecke (api 8080, dashboard 5173, mediamtx 8888 HLS / 9997 API, prometheus 9090, grafana 3000).
- Inter-Service-Konfiguration (`OTEL_EXPORTER_OTLP_ENDPOINT`, MediaMTX-URL für Player-SDK-Demo).
- Service-Abhängigkeiten in Compose (`depends_on`, Healthchecks).
- **NF-37 CSP-Beispiele für `connect-src`**: Empfohlener `Content-Security-Policy`-Header für die Dashboard-Auslieferung (z. B. `default-src 'self'; connect-src 'self' http://localhost:8080`) — getrennte Beispiele für Dev-Mode (Vite-Proxy) und Compose-Production-Build (separater API-Origin).

## 4. Test-/Lint-/Coverage-Workflows

> **Status: TODO** — wird ergänzt während `0.1.0`-Implementierung.

Zu dokumentieren:

- `apps/api`: `make test`, `make lint`, `make build`, `make arch-check` (siehe `apps/api/Makefile`).
- `packages/player-sdk` (ab `0.1.1`): pnpm-Test-/Lint-Workflows.
- `apps/dashboard` (ab `0.1.1`): SvelteKit-Test-/Lint-Workflows.
- CI-Pendant (sobald OE-6 entschieden, MVP-32).
- Coverage-Schwellwert (Folge-ADR aus Roadmap §4).

## 5. Häufige Probleme und Workarounds

> **Status: TODO** — entsteht durch Erfahrung während `0.1.0`-Phase.

Beispiele für mögliche Einträge:

- Port-Konflikte (z. B. lokal laufende Services auf 8080).
- WSL2-Filesystem-Performance (Empfehlung: Repo unter `\\wsl$`).
- macOS-Docker-Resource-Limits.
- MediaMTX-Stream startet nicht (FFmpeg-Generator-Logs prüfen).
