# m-trace

**OpenTelemetry-native observability for live media streaming.**

m-trace is a self-hosted observability and diagnostics stack for live media workflows.  
It helps trace media streams from ingest to player by combining player telemetry, stream sessions, infrastructure signals, Prometheus metrics, and OpenTelemetry-compatible event modeling.

> Status: early planning / architecture phase

---

## What is m-trace?

m-trace is designed for developers, self-hosters, small streaming platforms, broadcasters, and technical teams who want to understand what happens inside their streaming pipeline without depending on a proprietary SaaS analytics silo.

The first goal is simple:

```text
MediaMTX + hls.js demo player + playback events + dashboard + OpenTelemetry-compatible model
```

The long-term goal is broader:

```text
Trace live media streams from ingest to player.
```

---

## Why m-trace?

Commercial platforms such as Mux Data, Bitmovin Analytics, NPAW/YOUBORA, and Conviva solve many QoE and analytics problems.  
m-trace focuses on a different gap:

- self-hosted streaming observability
- OpenTelemetry-native modeling
- ingest-to-player correlation
- developer-friendly local demos
- streaming diagnostics instead of business analytics
- practical tooling for small teams and labs

The project is not trying to replace a full commercial video analytics platform.  
It aims to become a practical open-source stack for technical streaming diagnosis.

---

## Core idea

A typical live streaming flow looks like this:

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

m-trace collects and normalizes signals from the player and backend so that stream sessions can be inspected, debugged, and eventually correlated with infrastructure telemetry.

---

## MVP scope

The first MVP is intentionally small.

### Included in v0.1.0

- mono-repo structure
- backend API under `apps/api`
- dashboard under `apps/dashboard`
- demo player as dashboard route `/demo`
- player SDK under `packages/player-sdk`
- hls.js adapter
- MediaMTX-based local streaming setup
- FFmpeg test stream
- playback event ingestion
- basic stream session view
- basic event view
- Prometheus-compatible aggregate metrics
- OpenTelemetry-compatible event model
- in-memory or SQLite persistence
- Docker-first local development

### Not included in v0.1.0

- separate demo-player app
- separate analyzer API
- production multi-tenancy
- WebRTC monitoring
- SRT health view
- Tempo as required dependency
- Mimir or ClickHouse
- Kubernetes production deployment
- full HLS/DASH manifest analyzer

---

## Planned repository structure

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

Not all directories are part of the first MVP.  
Some are placeholders for the roadmap.

---

## Architecture principles

m-trace uses pragmatic architecture boundaries.

### Backend

The backend technology is still open until the backend spike is completed.

Candidates:

- Go
- Micronaut / JVM

The backend should follow a hexagonal architecture where it adds real value:

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
│       │   └── http/
│       └── out/
│           ├── persistence/
│           ├── telemetry/
│           └── metrics/
└── Dockerfile
```

Dependency direction:

```text
adapters → hexagon
```

The domain must not depend on HTTP, database, framework, Docker, or OpenTelemetry implementation details.

### Player SDK

The player SDK is intentionally not fully hexagonal in the MVP.

```text
packages/player-sdk/src/
├── core/
├── adapters/
│   └── hlsjs/
├── transport/
├── types/
└── index.ts
```

The first supported player integration is:

```text
hls.js
```

Other integrations are future work:

- dash.js
- Shaka Player
- Video.js
- native Safari HLS
- WebRTC getStats

---

## Event model

Player events use a versioned wire format.

Example:

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

Important concepts:

- `schema_version`
- `project_id`
- `session_id`
- `client_timestamp`
- `server_received_at`
- `sequence_number`
- SDK name and version

The backend must handle schema evolution, time skew, rate limits, and invalid event batches explicitly.

---

## Metrics

Prometheus is used for aggregate metrics only.

Examples:

```text
mtrace_playback_events_total
mtrace_invalid_events_total
mtrace_rate_limited_events_total
mtrace_dropped_events_total
mtrace_active_sessions
```

High-cardinality values such as `session_id`, `user_agent`, or `segment_url` must not be used as Prometheus labels.

Per-session debugging should be modeled as traces or stored in a suitable event/session store.

---

## OpenTelemetry strategy

m-trace is OpenTelemetry-native by design.

That means:

- use existing OTel semantic conventions where possible
- define media-specific attributes only where needed
- avoid vendor-specific telemetry formats
- keep session data trace-compatible
- keep Prometheus focused on aggregates
- prepare future correlation across ingest, origin, and player

A future player-session trace may look like this:

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

## Local development goal

The intended developer experience is:

```bash
git clone https://github.com/<owner>/m-trace.git
cd m-trace
make dev
```

Expected local services:

| Service          | Purpose                         |
| ---------------- | ------------------------------- |
| API              | event ingestion and session API |
| Dashboard        | web UI and `/demo` player route |
| MediaMTX         | local media server              |
| FFmpeg generator | test stream                     |
| Prometheus       | aggregate metrics               |
| Grafana          | optional dashboards             |
| OTel Collector   | optional telemetry pipeline     |

This setup is not implemented yet.  
It is the target for the first MVP.

---

## Backend technology spike

Before implementing the backend API, m-trace will run a short technology spike.

Goal:

```text
Decide between Go and Micronaut for apps/api.
```

Spike branches:

```text
spike/go-api
spike/micronaut-api
```

Outcome:

```text
docs/adr/0001-backend-stack.md
```

The selected stack becomes the foundation for `apps/api`.

---

## Roadmap

### v0.1.0 — OTel-native local demo

- MediaMTX local setup
- hls.js demo route
- player event ingestion
- basic session view
- basic event view
- Prometheus aggregate metrics
- OpenTelemetry-compatible event model

### v0.2.0 — Publishable player SDK

- npm package
- stable public API
- hls.js adapter tests
- event schema compatibility tests
- batching and sampling
- documented browser support

### v0.3.0 — Stream analyzer

- HLS manifest parsing
- segment duration checks
- target duration checks
- standalone CLI foundation

### v0.4.0 — Advanced trace correlation

- player session traces
- optional Tempo integration
- session timeline view
- sampling strategy

### v0.5.0 — Multi-protocol lab

- DASH example
- SRS example
- extended MediaMTX examples

### v0.6.0 — SRT health view

- SRT metrics
- RTT, packet loss, retransmissions
- link health dashboard
- SRT troubleshooting docs

---

## Browser support

MVP browser support is intentionally narrow.

| Environment                     | MVP status          |
| ------------------------------- | ------------------- |
| Chrome Desktop, current stable  | supported           |
| Firefox Desktop, current stable | supported           |
| Safari Desktop, current stable  | limited             |
| Chromium-based browsers         | best effort         |
| iOS Safari                      | not required in MVP |
| Android Chrome                  | not required in MVP |
| Smart-TV browsers               | out of scope        |
| Embedded WebViews               | out of scope        |

The MVP integration path is hls.js.  
Native Safari HLS introspection is not a v0.1.0 goal.

---

## Security and privacy

m-trace should be safe by default for self-hosted environments.

MVP principles:

- no secrets in the repository
- no cookie-based telemetry ingestion
- SDK requests use `credentials: "omit"` by default
- allowed origins are configured per project
- project tokens are treated as low-criticality browser tokens
- rate limits are required
- IP addresses should not be stored unnecessarily
- user-agent data should be reducible or anonymized
- GDPR-friendly operation must be possible

---

## What m-trace is not

m-trace is not:

- a commercial QoE analytics replacement
- an advertising analytics system
- a DRM analytics platform
- a CDN optimizer
- a full multi-tenant SaaS platform
- a replacement for MediaMTX, FFmpeg, Grafana, or Prometheus

m-trace is a technical observability and diagnostics project for media streaming workflows.

---

## Current status

The project is in the planning and spike phase.

Current documents:

```text
docs/spike/0001-backend-stack.md
docs/adr/0001-backend-stack.md
```

The first major implementation decision is the backend stack.

---

## License

License is not finalized yet.

Recommended options:

- Apache-2.0
- MIT

Apache-2.0 is preferred if long-term open-source adoption and patent clarity are important.

---

## Contributing

Contributions are not open yet because the repository is still in the initial planning phase.

Planned contribution areas:

- player SDK
- hls.js telemetry
- backend API
- MediaMTX examples
- OpenTelemetry modeling
- Prometheus/Grafana dashboards
- HLS/DASH analyzer
- SRT metrics

---

## Name

`m-trace` means:

```text
Media Trace
```

The project goal is simple:

```text
Trace media streams from ingest to player.
```
