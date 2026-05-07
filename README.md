# m-trace

**OpenTelemetry-native Observability für Live-Media-Streaming.**

m-trace ist ein selbst-gehosteter Observability- und Diagnose-Stack für Live-Media-Workflows.  
Er hilft, Media-Streams von der Ingest-Seite bis zum Player nachzuverfolgen, indem er Player-Telemetrie, Stream-Sessions, Infrastruktursignale, Prometheus-Metriken und ein OpenTelemetry-kompatibles Eventmodell zusammenführt.

> Status: `0.9.5` released — Quality-Gates Wave 2 (Patch-Release-Konvention `0.X.Y`): vier Tranche-Lieferungen ohne User-Surface aus [`docs/planning/done/plan-0.9.5.md`](docs/planning/done/plan-0.9.5.md) — Benchmark-Smoke (PR-Pfad, opt-in bis Beobachtungsphase abgeschlossen; `make benchmark-smoke`), Nightly-`benchstat`-Regressionen mit Quarantäne-Mechanik (`.github/workflows/benchmark.yml`), selektives Fuzzing (sechs Go-Targets) plus TS-Property-Tests via `fast-check` (`make fuzz-check`, `.github/workflows/fuzz.yml`), Mutation-Testing als nicht-blockierender Nightly-Report mit gremlins (Go) und StrykerJS (TS) (`make mutation-report`, `.github/workflows/mutation.yml`). Erstfund über `FuzzMapMediaMtxItem` ([`mapping.go`](apps/api/adapters/driven/srt/mediamtxclient/mapping.go) — `mbpsLinkCapacity=-1` leakte als negativer `AvailableBandwidthBPS`). Kein Lastenheft-Patch (Quality-Gates, keine User-Surface). Vorgänger `v0.9.1` (Drift-Smoke-Robustheit) und `v0.9.0` (Drift-Smoke + SRS-Lab + DASH-Manifest-Analyse, RAK-56..RAK-59 / Lastenheft `1.1.11` §13.11) bleiben unverändert.

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

## Lieferstand und Roadmap

- **Aktueller Lieferstand pro Release**: [`CHANGELOG.md`](CHANGELOG.md).
- **Aktive Phase und nächste Schritte**: Sektion „Roadmap" weiter
  unten plus [`docs/planning/in-progress/`](docs/planning/in-progress/).
- **Was m-trace bewusst nicht ist**: Sektion „Was m-trace nicht ist"
  weiter unten.

---

## Architekturprinzipien

Die aktuelle Architektur ist in [spec/architecture.md](spec/architecture.md) beschrieben.

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
        "name": "@npm9912/player-sdk",
        "version": "0.2.0"
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

Prometheus wird ausschließlich für Aggregat-Metriken genutzt. Die
drei Backends teilen die Verantwortung wie folgt (kanonische 3-Spalten-
Tabelle: [`spec/telemetry-model.md`](spec/telemetry-model.md) §3.3):

| Backend | Rolle | Cardinality |
|---|---|---|
| **Prometheus** | Aggregat-Metriken (Counter, Rates) | bounded — Forbidden-Liste aus [`spec/telemetry-model.md`](spec/telemetry-model.md) §3.1 release-blockierend |
| **SQLite** (ADR-0002) | Per-Session-Historie inkl. `session_id`, `correlation_id`, `trace_id`, redacted URLs | unbeschränkt |
| **OTel/Tempo** | Per-Request-Trace-Spans (sample-basiert) | nicht im Cardinality-Vertrag |

Beispiele für Prometheus-Counter (alle label-frei):

```text
mtrace_playback_events_total
mtrace_invalid_events_total
mtrace_rate_limited_events_total
mtrace_dropped_events_total
mtrace_active_sessions
mtrace_api_batches_received
```

Hochkardinale Werte wie `session_id`, `correlation_id`, `trace_id`,
`user_agent`, `segment_url`, `client_ip` oder Token-/Credential-Felder
dürfen **nicht** als Prometheus-Labels verwendet werden — die
vollständige Forbidden-Liste plus Suffix-Regeln (`*_url`, `*_uri`,
`*_token`, `*_secret`) steht in
[`spec/telemetry-model.md`](spec/telemetry-model.md) §3.1.

Per-Session-Debugging läuft über die durable SQLite-Persistenz und
optional über Tempo-Spans (`make dev-tempo`) — niemals über Prometheus.

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

## Lokale Entwicklung

Das lokale Core-Lab startet Backend-API, Dashboard, MediaMTX und einen FFmpeg-Teststream:

```bash
git clone https://github.com/pt9912/m-trace.git
cd m-trace
make dev
```

Smoke-Test:

```bash
make smoke
```

SDK- und Demo-Dokumentation:

- [spec/player-sdk.md](spec/player-sdk.md) beschreibt Installation, Public API, Transport, Performance-Budget und Browser-Build.
- [docs/user/demo-integration.md](docs/user/demo-integration.md) beschreibt die Dashboard-Route `/demo` als lokale hls.js-/Player-SDK-Integration.
- [spec/browser-support.md](spec/browser-support.md) dokumentiert die Browser-Support-Matrix.

Lokaler SDK-/Demo-Pfad:

```bash
pnpm --filter @npm9912/player-sdk run pack:smoke
make dev
# dann http://localhost:5173/demo?session_id=readme-demo&autostart=1 öffnen
```

Optionaler Observability-Stack mit Prometheus, Grafana und OTel-Collector:

```bash
make dev-observability
make smoke-observability
```

Browser-End-to-End-Test im Playwright-Container:

```bash
make browser-e2e
```

- Dashboard: <http://localhost:5173>
- API: <http://localhost:8080/api/health>
- HLS-Teststream: <http://localhost:8888/teststream/index.m3u8>
- Prometheus: <http://localhost:9090> (`make dev-observability`)
- Grafana: <http://localhost:3000> (`admin`/`admin`, `make dev-observability`)
- OTel Collector: OTLP `localhost:4317`/`4318`, Health <http://localhost:13133>

Details stehen in [docs/user/local-development.md](docs/user/local-development.md).

## Roadmap

Status pro Release, aktive Phase, nächste Schritte und Folge-ADRs
stehen kanonisch in
[`docs/planning/in-progress/roadmap.md`](docs/planning/in-progress/roadmap.md).
Pro-Release-Lieferstand mit Bullet-Listen siehe
[`CHANGELOG.md`](CHANGELOG.md).

---

## Browser-Support

Der MVP-Browser-Support ist bewusst eng gefasst.

| Umgebung                         | MVP-Status                |
| -------------------------------- | ------------------------- |
| Chrome Desktop, aktuelle Stable  | unterstützt               |
| Firefox Desktop, aktuelle Stable | unterstützt               |
| Safari Desktop, aktuelle Stable  | eingeschränkt             |
| Chromium-basierte Browser        | best effort               |
| iOS Safari                       | im MVP nicht erforderlich |
| Android Chrome                   | im MVP nicht erforderlich |
| Smart-TV-Browser                 | nicht im Scope            |
| Embedded WebViews                | nicht im Scope            |

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
- ein Production-Ready-Kubernetes-Deployment
- ein Ersatz für MediaMTX, FFmpeg, Grafana, Prometheus oder
  Production-Grade-Storage-Backends wie Mimir oder ClickHouse

m-trace ist ein technisches Observability- und Diagnose-Projekt für Media-Streaming-Workflows.

---

## Aktueller Stand

Das Projekt steht bei `0.9.5` released — Quality-Gates Wave 2
Patch-Release nach `0.9.0`/`0.9.1` (Patch-Release-Konvention
`0.X.Y`, siehe [`docs/user/releasing.md`](docs/user/releasing.md)
§3.1). Inhalt: vier statistisch- bzw. langlaufende Quality-Gates
aus [`docs/planning/open/extra-gates.md`](docs/planning/open/extra-gates.md)
in einem Patch-Release ausgeliefert, Plan-File in
[`done/plan-0.9.5.md`](docs/planning/done/plan-0.9.5.md). Kein
Lastenheft-Patch (Quality-Gates, keine User-Surface).

**Tranche 1 — Benchmark-Smoke**: Go-Bench-Suite in `apps/api`
für vier Hot-Paths (RegisterPlaybackEventBatch typical+max,
EventRepository AppendBatch, SessionsService ListSessions,
Cursor encode/decode), TS-Bench-Suite
`packages/stream-analyzer/benchmarks/analyzer.bench.ts` für
sieben Hot-Paths (HLS Master/Media, DASH-MPD VOD/Live, Detector,
SSRF-URL-Klassifizierung). Single-Source-Budgets in
[`docs/perf/budgets.md`](docs/perf/budgets.md). Wrapper
`make benchmark-smoke` plus Validator
`scripts/check-bench-budgets.mjs`. Beobachtungs-Nightly
[`.github/workflows/benchmark-observation.yml`](.github/workflows/benchmark-observation.yml)
(Cron `30 2 * * *` UTC, `continue-on-error: true`); PR-
Blockierung folgt nach N=3..5 grünen Beobachtungsläufen.

**Tranche 2 — Nightly-`benchstat`-Regressionen**:
[`.github/workflows/benchmark.yml`](.github/workflows/benchmark.yml)
(Cron `0 4 * * *` UTC) führt 10× `go test -bench=.` aus,
vergleicht via `benchstat` gegen Baseline aus orphan-Branch
`benchmark-baseline`. Schwelle +15 % auf p<0.05; Auto-Issue mit
benchstat-Diff-Block. Quarantäne-Mechanik via
`// bench:quarantine YYYY-MM-DD reason: <text>` (max. 30 Tage),
Validator `scripts/check-bench-quarantines.mjs`.

**Tranche 3 — Selektives Fuzzing + Property Tests**: sechs
Go-Fuzz-Targets in vier Packages (Cursor encode/decode, wireBatch
Decode, Reserved-Event-Meta, Unavailable-Reason, MediaMTX-Item-
Mapping); drei TS-Property-Test-Suites via `fast-check@4.4.0`
(HLS- und DASH-Parser, URL-Redaction). `make fuzz-check`-Target
mit `FUZZTIME`-Override (Default 30 s) plus Nightly
[`.github/workflows/fuzz.yml`](.github/workflows/fuzz.yml)
(Cron `0 5 * * *` UTC, 5 min/Target, Auto-Issue mit Crash-Repo-
Pfad). Erstfund: `FuzzMapMediaMtxItem` zeigte
`mbpsLinkCapacity=-1` → `AvailableBandwidthBPS=-1_000_000` in
[`mapping.go`](apps/api/adapters/driven/srt/mediamtxclient/mapping.go).
Operator-Doku in [`docs/dev/fuzzing.md`](docs/dev/fuzzing.md).

**Tranche 4 — Mutation Testing (Nightly-Report)**: Pilot-Module
`apps/api/hexagon/application/event_meta_validation.go` (gremlins
statt unmaintainted go-mutesting) und
`packages/player-sdk/src/adapters/webrtc/sampling.ts` (StrykerJS
+ vitest-runner). `make mutation-report` als Wrapper. Nightly
[`.github/workflows/mutation.yml`](.github/workflows/mutation.yml)
(Cron `0 6 * * *` UTC, beide Jobs `continue-on-error: true`).
Score-Schwelle und Übergangs-Pfad zur PR-Blockierung in
[`docs/dev/mutation-testing.md`](docs/dev/mutation-testing.md).

Lieferstand `0.9.0` und früher bleibt unverändert:

**Tranche 1 — Browser-Drift-Smoke (RAK-56)**: automatisiert R-12 ab.
`tests/e2e/webrtc-stats-drift.spec.ts` (Playwright) öffnet im Page-
Context eine eigene `RTCPeerConnection` gegen das `mtrace-webrtc`-
Lab, ruft `pc.getStats()` und vergleicht gegen die Muss-Felder pro
`RTCStatsType`-Gruppe aus
[`spec/telemetry-model.md`](spec/telemetry-model.md) §3.5.2 plus die
Enum-Allowlists aus §1.4. Nightly-Workflow
[`.github/workflows/webrtc-drift.yml`](.github/workflows/webrtc-drift.yml)
führt `make smoke-webrtc-stats-drift` gegen Chromium und Firefox aus
dem Playwright-Bundle aus; bei Schema-Drift bricht der Smoke und
optional erstellt der Workflow ein Issue (opt-in über
`secrets.DRIFT_AUTO_ISSUE=1`). R-12 wandert von „release-blockierend"
auf „automatisiert detektiert".

**Tranche 2 — SRS-Lab (RAK-57 / MVP-36)**:
[`examples/srs/`](examples/srs/) als fünftes Multi-Protocol-Beispiel
analog `examples/srt/`/`examples/dash/`/`examples/webrtc/`. Eigener
Compose-Project `mtrace-srs` mit `ossrs/srs:5` plus FFmpeg-RTMP-
Publisher; Host-Ports 1935 (RTMP) / 1985 (HTTP-API) / 8088 (HTTP-FLV)
kollisionsfrei zu Core-Lab und anderen Stacks. Opt-in `make smoke-srs`
prüft endpoint-/compose-only HTTP-API + FFmpeg-Stream-Registrierung +
FLV-Magic-Header. Kein produktiver Telemetriepfad.

**Tranche 3 — DASH-Manifest-Analyse (RAK-58 / RAK-59 / NF-12 /
MVP-37)**: `@npm9912/stream-analyzer` versteht DASH-MPD-Eingaben
zusätzlich zu HLS-Manifesten. Auto-Detection am Body-Anfang
(`<?xml`/`<MPD` → DASH; `#EXTM3U` → HLS); Manifest-Loader
generalisiert auf HLS+DASH (Content-Type-Allowlist um
`application/dash+xml`); regex-basierter MPD-Parser ohne externe
XML-Dependency deckt MPD/Period/AdaptationSet/Representation für
VOD- und einfache Live-MPDs ab. JSON-Result-Schema bekommt
`analyzerKind:"dash"` / `playlistType:"dash"` als zweite Variante;
HLS-Pfad bleibt unverändert (additiv). Neuer Public-Code
`manifest_not_supported` für Eingaben weder HLS noch DASH (HTTP
422); `manifest_not_hls` bleibt HLS-Parser-spezifisch.
`pnpm m-trace check <file.mpd>` dispatcht automatisch; `make
smoke-cli` validiert den Pfad live.

Lastenheft-Patch `1.1.11` ergänzt §13.11 mit RAK-56..RAK-59 und
zieht MVP-37 entsprechend NF-12 von „Kann" auf „Muss" hoch
(Patch-Log §4a.14 in
[`docs/planning/done/plan-0.1.0.md`](docs/planning/done/plan-0.1.0.md)).
Operator-Verifikation in
[`docs/user/releasing.md`](docs/user/releasing.md) §2.4 (drei Sub-
Blöcke: Drift-Smoke / SRS-Lab-Boot / DASH-CLI-Probe).

Quality-Gates Wave 1 aus `0.8.5` (Security-Gates + Generated-
Artifact-Drift), Player-SDK-WebRTC-Adapter aus `0.8.0`, SRT-Health-
View aus `0.6.0` und Multi-Protokoll-Lab aus `0.5.0` bleiben
unverändert. Tranchen 0–5 in
[`docs/planning/done/plan-0.9.0.md`](docs/planning/done/plan-0.9.0.md)
archiviert. Nächste Phase: `plan-0.9.5.md` (Quality-Gates Wave 2 —
Benchmark-Smoke, Nightly-benchstat, Fuzzing, Mutation Testing) liegt
unter [`docs/planning/open/`](docs/planning/open/).
Archivierte Plan-Dateien:
[`docs/planning/done/plan-0.9.0.md`](docs/planning/done/plan-0.9.0.md),
[`docs/planning/done/plan-0.8.5.md`](docs/planning/done/plan-0.8.5.md),
[`docs/planning/done/plan-0.8.0.md`](docs/planning/done/plan-0.8.0.md),
[`docs/planning/done/plan-0.7.0.md`](docs/planning/done/plan-0.7.0.md),
[`docs/planning/done/plan-0.6.0.md`](docs/planning/done/plan-0.6.0.md),
[`docs/planning/done/plan-0.5.0.md`](docs/planning/done/plan-0.5.0.md),
[`docs/planning/done/plan-0.4.0.md`](docs/planning/done/plan-0.4.0.md).

Leitende Dokumente:

- [spec/lastenheft.md](spec/lastenheft.md) — Anforderungen (verbindlich, 1.1.9)
- [docs/planning/in-progress/roadmap.md](docs/planning/in-progress/roadmap.md) — Status, Folge-ADRs, offene Entscheidungen
- [docs/adr/0001-backend-stack.md](docs/adr/0001-backend-stack.md) — Backend-Entscheidung (Accepted: Go)
- [docs/user/releasing.md](docs/user/releasing.md) — Release-Prozess
- [docs/user/quality.md](docs/user/quality.md) — Qualitätsrichtlinien

Nächste Schritte stehen in [docs/planning/in-progress/roadmap.md](docs/planning/in-progress/roadmap.md) §2.

---

## Lizenz

[MIT License](LICENSE).

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
