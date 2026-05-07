# Implementation Plan — `0.9.0` (Drift-Smoke + SRS-Lab + DASH-Analyse)

> **Status**: 🟡 in Arbeit (Plan-Skelett am 2026-05-07 von
> `docs/planning/open/` nach `docs/planning/in-progress/` gezogen,
> Tranche 0 abgeschlossen — Plan-Aktivierung, Lastenheft-Patch
> `1.1.11` mit RAK-56..RAK-59 und MVP-37-Hochstufung sowie
> Toolchain-Bump-Check ohne Bump-Bedarf). Vorgänger `v0.8.5` ist released
> (Tag `v0.8.5` auf `ce05e3b`, GitHub-Release veröffentlicht; Plan
> archiviert in [`done/plan-0.8.5.md`](../done/plan-0.8.5.md)). `0.8.0`
> (Player-SDK-WebRTC-Adapter) bleibt auf Tag `v0.8.0` (`8df263a`,
> Release-Gate-Fix-Closeout); Plan archiviert in
> [`done/plan-0.8.0.md`](../done/plan-0.8.0.md). Lastenheft-Patch
> `1.1.11` wird im Rahmen von Tranche 0b ausgeliefert (siehe §0.2).
>
> **Lastenheft-Status**: `1.1.10` §13.10 ist abgeschlossen. `0.9.0`
> bündelt drei eigenständige Liefergegenstände (Browser-Drift-Smoke
> für R-12, SRS-Lab für MVP-36, DASH-Analyse für MVP-37 / NF-12).
> Lastenheft-Patch `1.1.11` (siehe §0.2) ergänzt einen neuen Block
> §13.11 mit RAK-56..RAK-59 und hebt MVP-37 (DASH-Analyse, Kann)
> entsprechend NF-12 (DASH-Analyse, Muss) auf „Muss".
>
> **Bezug**: [Lastenheft `1.1.10`](../../../spec/lastenheft.md) §8.x
> NF-12 (DASH-Analyse, Muss), §12.3 MVP-36 (SRS-Beispiel, Kann),
> §12.3 MVP-37 (DASH-Analyse, Kann);
> [`done/plan-0.8.0.md`](../done/plan-0.8.0.md) §4 Tranche 3 (R-12
> wurde dort release-blockierend angehoben);
> [`done/plan-0.8.5.md`](../done/plan-0.8.5.md) (Quality-Gates Wave 1,
> Vorgänger-Patch);
> [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
> §3.5.3 (WebRTC-Schema-Drift-Strategie);
> [`docs/planning/open/risks-backlog.md`](../open/risks-backlog.md) R-12;
> [`packages/stream-analyzer/`](../../../packages/stream-analyzer/)
> (HLS-Stand `0.3.0`, RAK-22..RAK-28).
>
> **Nachfolger**: offen — kein `plan-0.10.0.md` vorbereitet.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`done/plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

Scope-Grenze: dieser Plan bündelt drei thematisch getrennte, aber
für eine Solo-Phase einzeln zu kleine Liefergegenstände. Jede der
drei Themen-Tranchen liefert ihre eigene RAK-/MVP-Gruppe; sie sind
**unabhängig** und müssen nicht in der genannten Reihenfolge
erledigt werden, solange Tranche 0 (Plan-Aktivierung + Lastenheft-
Patch) und Tranche 5 (Closeout) den Rahmen bilden.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- `0.8.0` ist released (Tag `v0.8.0` auf dem Release-Gate-Fix nach
  `8df263a`); produktive
  WebRTC-Telemetrie ist live mit `mtrace_webrtc_*`-Countern und
  release-blockierendem R-12.
- Lastenheft-Patch `1.1.11` (siehe §0.2) ist akzeptiert; RAK-56..
  RAK-59 sind im Lastenheft §13.11 (oder analog) verankert; MVP-37
  ist auf „Muss" hochgezogen.
- Toolchain ist non-EOL: Go-/Node-/golangci-lint-Linien aus `0.7.0`
  Tranche 0 (Commits `ccf68b1` + `8bfad21`) sind weiterhin aktuell.
  Bei Bedarf eigene Toolchain-Hardening-Sub-Tranche analog
  `0.7.0`.

### 0.2 Lastenheft-Patch `1.1.11` (ausgeliefert)

Der Patch ergänzt vier neue RAK in einem neuen §13.11-Block und
zieht MVP-37 entsprechend NF-12 auf „Muss". Lieferstand mit
Tranche-0b-Commit (Header-Bump `1.1.10` → `1.1.11`, neuer §13.11,
§12.3 MVP-37-Patch-Note, Patch-Log §4a.14 in `done/plan-0.1.0.md`,
roadmap.md §2 Schritt 42):

| RAK | Priorität | Inhalt (Vorschlag für `spec/lastenheft.md` §13.11) |
| --- | --------- | -------------------------------------------------- |
| RAK-56 | Soll | Browser-Drift-Smoke (Playwright) probt `getStats()` aus echten Browser-Versionen gegen das `examples/webrtc/`-Lab und vergleicht die Reports gegen die Allowlist aus `spec/telemetry-model.md` §1.4 / §3.5.2. Treffer eines unbekannten Enum-Werts oder fehlender Muss-Felder bricht den Smoke; opt-in `make smoke-webrtc-stats-drift`, Nightly-CI-Job. Schließt R-12 als „release-blockierend, automatisiert detektiert". |
| RAK-57 | Kann | SRS-Lab-Beispiel `examples/srs/` (Project `mtrace-srs`, analog `examples/srt/`/`examples/dash/`/`examples/webrtc/`): Compose-Stack mit `ossrs/srs:5`-Image, FFmpeg-Publisher, opt-in `make smoke-srs` (endpoint-/compose-only). Hebt MVP-36 von „Kann" auf eingelöst, ohne MVP-Priorität zu ändern. |
| RAK-58 | Muss | DASH-Manifest-Analyse im `@npm9912/stream-analyzer`: Auto-Detection von DASH-MPD-Eingaben (XML-Header), Parse von `AdaptationSet`/`Representation`/`SegmentTemplate`, JSON-Result-Schema mit `analyzerKind: "dash"` analog HLS aus `0.3.0`. Hebt MVP-37 (Kann) auf „Muss" entsprechend NF-12. |
| RAK-59 | Kann | DASH-CLI-Pfad: `pnpm m-trace check <url-or-file.mpd>` dispatcht automatisch auf DASH und liefert dasselbe JSON-Result wie der Library-Pfad. CLI-Smoke (`make smoke-cli`) erweitert. |

Begründung des Bündels:

- RAK-56 ist die natürliche Folge der `0.8.0`-Tranche-3-Sequenzierung
  (R-12 release-blockierend, aber bisher nur durch manuellen Drift-
  Review abgesichert) — alleine wäre er zu klein für eine eigene
  Phase.
- RAK-57 / SRS-Lab ist ein direkter Analog-Schritt zu `examples/srt/`,
  `examples/dash/` und `examples/webrtc/`; das Operator-Surface ist
  bekannt.
- RAK-58 / DASH-Analyse ist die offene NF-12-Pflicht aus dem
  Stammvertrag des Stream-Analyzers (`0.3.0` lieferte HLS-only,
  NF-12 verlangt DASH).

### 0.3 Out-of-Scope-Klauseln (durchgängig)

- Kein produktiver WebRTC-Adapter-Pfad-Bruch. Der Drift-Smoke
  prüft nur die `getStats()`-Allowlist; eine Schema-Migration
  (z. B. neue `webrtc.*`-Keys) ist Folge-Plan, nicht `0.9.0`.
- Kein DASH-Player-SDK-Adapter (`attachDash` o. ä.). Player-SDK
  bleibt auf `attachHlsJs`/`attachWebRtc`; DASH-Analyse ist
  Analyzer-/CLI-Pfad, nicht Player-Pfad.
- Kein produktiver SRS-Telemetriepfad (`mtrace_srs_*`-Counter).
  SRS-Lab ist endpoint-/compose-only analog `examples/srt/` —
  kein Lastenheft-/Wire-Vertrag.
- Kein DASH-Live-/Low-Latency-Spezialfall. RAK-58 deckt VOD-MPD
  und einfache Live-MPD; segment-template-bezogene Edge-Cases
  (z. B. `$Time$`-Variablen, `availabilityStartTime`-Drift) sind
  als Out-of-Scope dokumentiert und Folge-Plan.
- Keine Multi-Tenant-Erweiterungen (Postgres MVP-40, K8s MVP-42,
  ClickHouse MVP-41). Diese Themen brauchen eigene Phase.
- Keine Quality-Gates (govulncheck, Benchmark-Smoke, Fuzzing,
  Mutation Testing, Generated-Artifact-Drift). Diese sind in
  [`plan-0.8.5.md`](../done/plan-0.8.5.md) (Wave 1, vor `0.9.0`) und
  [`plan-0.9.5.md`](../open/plan-0.9.5.md) (Wave 2, nach `0.9.0`)
  konkretisiert; Master-Backlog steht in
  [`extra-gates.md`](../open/extra-gates.md).

### 0.4 Sequenzierung und harte Gates

1. Tranche 0 (Plan-Aktivierung + Lastenheft-Patch) ist Pflicht
   vor jeder anderen Tranche.
2. Tranche 1 (Drift-Smoke), Tranche 2 (SRS-Lab) und Tranche 3
   (DASH-Analyse) sind **unabhängig** — Reihenfolge richtet sich
   nach Operator-Präferenz. Default-Empfehlung: Drift-Smoke zuerst
   (kleinste, schließt R-12-Daueraufgabe), dann DASH-Analyse
   (größter Liefergegenstand), dann SRS-Lab (kleinster Operator-
   Use-Case).
3. Tranche 4 (Compat-Tests + Doku) erst nach Tranche 1+2+3.
4. Tranche 5 (Closeout) verschiebt diesen Plan nach `done/`,
   bumpt die Versionen 0.8.0 → 0.9.0 (analog `0.8.0` Tranche 5,
   inkl. hartkodierter Tarball-Pfad in `pack:smoke`) und setzt
   den Tag `v0.9.0`.

### 0.5 Implementierungsleitplanken

**Drift-Smoke (Tranche 1)**: Bevorzugte Form ist eine Playwright-
Spec, die das `mtrace-webrtc`-Lab als Stack hochfährt, in echten
Browser-Versionen (Chromium und Firefox via Playwright-Default,
Safari/WebKit opt-in) `attachWebRtc` ausführt und nach Handshake
direkt `pc.getStats()` aufruft. Die Spec validiert, dass alle in
`spec/telemetry-model.md` §3.5.2 als Muss markierten Felder
existieren und alle Enum-Werte in der §1.4-Allowlist liegen.

**SRS-Lab (Tranche 2)**: Bevorzugte Form ist `examples/srs/` mit
eigenständigem `compose.yaml` (Project `mtrace-srs`), `ossrs/srs:5`-
Image gepinnt, FFmpeg-Publisher analog `examples/srt/ffmpeg-srt-loop.sh`,
opt-in `make smoke-srs` (endpoint-/compose-only, kein
Playback-/Telemetrie-Anspruch).

**DASH-Analyse (Tranche 3)**: Bevorzugte Form erweitert
`packages/stream-analyzer/src/` um einen DASH-Detector (XML-Header-
Sniffing) und einen MPD-Parser; das JSON-Result-Schema bekommt
`analyzerKind: "dash"` als zweiten Wert (HLS bleibt unverändert).
`createCLI`-Dispatcher detektiert Eingabetyp aus `Content-Type`
oder Datei-Endung. Der gemeinsame Manifest-Loader wird dabei von
HLS-spezifischen Namen/Fehlermeldungen auf HLS+DASH generalisiert,
damit `application/dash+xml` nicht vor dem Parser geblockt wird.
Analyzer-Wire-Vertrag (`spec/contract-fixtures/analyzer/` plus
Go-Testdata-Kopien) wird um zwei DASH-Beispiele erweitert.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung (`open/` → `in-progress/`) + Lastenheft-Patch `1.1.11` (RAK-56..RAK-59 + MVP-37-Hochstufung) + ggf. Toolchain-Hardening | ✅ |
| 1 | Browser-Drift-Smoke für WebRTC-`getStats()` (RAK-56) | ✅ |
| 2 | SRS-Lab `examples/srs/` (RAK-57, MVP-36) | ✅ |
| 3 | DASH-Manifest-Analyse im `@npm9912/stream-analyzer` (RAK-58/RAK-59, MVP-37, NF-12) | ✅ |
| 4 | Compat-Tests + Browser-Support-Matrix-Pflege; Pack-Smoke + CLI-Smoke erweitert | ⬜ |
| 5 | Release-Doku, RAK-Verifikationsmatrix und Closeout (Versions-Bump 0.8.0 → 0.9.0, Plan nach `done/`, Tag `v0.9.0`) | ⬜ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Lastenheft-Patch

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche analog
`done/plan-0.8.0.md` §1a.

DoD:

- [x] Plan-Skelett von `docs/planning/open/plan-0.9.0.md` nach
  `docs/planning/in-progress/plan-0.9.0.md` verschoben (Status
  `⬜ → 🟡`); Cross-Refs in `roadmap.md` §1.2/§3 nachgezogen
  (Tranche-0a-Commit).
- [x] Lastenheft-Patch `1.1.11` schreiben: §13.11 neu mit RAK-56..
  RAK-59; §12.3 MVP-37 von „Kann" auf „Muss" hochgezogen entsprechend
  NF-12 (Hinweis: §12.3 historisch beibehalten mit Patch-Note).
  Patch-Eintrag als §4a.14 in `done/plan-0.1.0.md` Tranche 0c
  (Tranche-0b-Commit).
- [x] Toolchain-Bump-Check: keine Anpassung nötig. Go (`1.26`),
  golangci-lint (`v2.12.1-alpine`), Node (`22-trixie-slim`, seit
  `0.8.5` Tranche 1) und pnpm (`>=10 <11`) sind seit `0.7.0`
  Tranche 0 (Commits `ccf68b1` + `8bfad21`) und `0.8.5` Tranche 1
  (Image-Hardening, Commits `927555a` + `388491e`) aktuell und
  non-EOL. Race-Detector-Stage (`make api-race`) ist seit `0.7.0`
  in `make gates` enthalten; Generated-Drift-Gate (`make
  generated-drift-check`) seit `0.8.5` ebenso. Keine `0.9.0`-
  spezifischen neuen Tools — der DASH-Parser in Tranche 3 nutzt
  ausschließlich Workspace-interne TypeScript-Dependencies, und
  der Drift-Smoke in Tranche 1 läuft auf dem bestehenden
  Playwright-Container aus `0.4.0` (Tranche-0c-Commit).

---

## 2. Tranche 1 — Browser-Drift-Smoke (R-12)

Bezug: `risks-backlog.md` R-12; `spec/telemetry-model.md` §1.4
(webrtc.*-Allowlist) + §3.5.2/§3.5.3; `tests/e2e/`.

Ziel: Ein automatisierter Smoke, der den `getStats()`-Schema-Drift
in echten Browser-Versionen frühzeitig erkennt. Schließt R-12
operativ — der Drift-Review-Gate ist nicht mehr manuelle Pflicht
vor jedem Release-Tag, sondern auto-detektiert.

DoD:

- [x] `tests/e2e/webrtc-stats-drift.spec.ts` (neu, Playwright):
  öffnet im Page-Context (eigene `RTCPeerConnection`, kein
  Adapter-Hook nötig — Plan §0.5 gibt beide Pfade frei) eine
  WHEP-Verbindung gegen `http://localhost:8892/webrtc-test/whep`
  mit recvonly video+audio Transceivers; nach `connectionState=
  connected` ruft die Spec `pc.getStats()` auf und sammelt alle
  Reports. Die Spec ist via `MTRACE_WEBRTC_STATS_DRIFT=1` opt-in,
  damit `make browser-e2e` (anderer Stack, kein `mtrace-webrtc`-
  Lab) sie nicht versehentlich auslöst (Tranche-1.1-Commit).
- [x] Spec validiert für jede `RTCStatsType`-Gruppe aus §3.5.2,
  dass alle Muss-Felder existieren (peer-connection.connectionState,
  transport.dtlsState, candidate-pair.state, inbound-rtp.
  packetsLost+bytesReceived, outbound-rtp.bytesSent — letzteres
  legitim leer bei recvonly); Drift bricht den Smoke mit klarer
  Fehlermeldung („Browser X dropped field Z from RTCStatsType.foo
  (id=…)"). Soll-Felder werden über `console.log` als
  `[drift-soll]` geloggt, brechen den Smoke aber nicht
  (Tranche-1.1-Commit).
- [x] Spec validiert, dass `pc.connectionState` ∈ §1.4
  `connection_state`-Allowlist, `pc.iceConnectionState` ∈
  `ice_state`-Allowlist und alle `transport.dtlsState`-Werte ∈
  `dtls_state`-Allowlist liegen; unbekannter Enum-Wert → Smoke-Fail
  (Tranche-1.1-Commit).
- [x] `make smoke-webrtc-stats-drift`-Target opt-in (nicht in
  `make gates`); Help-Eintrag analog `smoke-webrtc-prep`. Default-
  Browser sind `chromium,firefox` aus dem Playwright-Bundle;
  `MTRACE_WEBRTC_DRIFT_BROWSERS=chromium,firefox,webkit` toggelt
  Safari/WebKit opt-in. Skript `scripts/smoke-webrtc-stats-drift.sh`
  fährt den `mtrace-webrtc`-Stack via `docker compose -p mtrace-
  webrtc up -d --build` hoch, delegiert die Endpoint-Probe an
  `scripts/smoke-webrtc-prep.sh` (`SMOKE_WEBRTC_AUTOSTART=0`-Modus
  hält den Stack offen) und ruft anschließend
  `pnpm exec playwright test tests/e2e/webrtc-stats-drift.spec.ts
  --project=$browser` für jeden Browser. Cleanup räumt nur den
  `mtrace-webrtc`-Project-Namen ab (Tranche-1.2-Commit).
- [x] CI-Workflow `.github/workflows/webrtc-drift.yml` (neu, Nightly
  via `schedule: cron '30 3 * * *'` plus `workflow_dispatch`):
  Setup-Steps wie `build.yml` (Checkout, pnpm 10.18.0, Node 22 aus
  `.nvmrc`, `pnpm install --frozen-lockfile`); installiert die
  Playwright-Browser explizit via
  `pnpm exec playwright install --with-deps chromium firefox`;
  führt `make smoke-webrtc-stats-drift`. Bei Failure wird (opt-in
  über das Repository-Secret `DRIFT_AUTO_ISSUE=1`, gemappt auf
  job-level `env.DRIFT_AUTO_ISSUE`) ein Issue mit Title, Workflow-
  Run-URL, Playwright-Stand und Reaktions-Pfad erstellt;
  `permissions: issues: write` ist auf Workflow-Ebene gesetzt
  (Tranche-1.3-Commit).
- [x] R-12 im `risks-backlog.md` von „release-blockierend ab
  nächstem Browser-Major-Bump" auf „automatisiert detektiert, Drift
  bricht den Drift-Smoke" angehoben; Manuell-Review entfällt;
  Reaktions-Pfad bleibt dokumentiert (Allowlist-Update + Spec-Patch
  + lokaler Smoke). Release-Pfad in `docs/user/releasing.md` neue
  §2.4 referenziert den Drift-Smoke und nennt die Cron-Zeit + den
  `MTRACE_WEBRTC_DRIFT_BROWSERS`-Toggle für WebKit/Safari; die
  Smoke-Liste in §2 listet `make smoke-webrtc-stats-drift` als
  `0.9.0`-Smoke (Tranche-1.4-Commit).

---

## 3. Tranche 2 — SRS-Lab `examples/srs/` (MVP-36 / RAK-57)

Bezug: `examples/README.md` (Multi-Protocol-Lab-Konvention,
`plan-0.5.0.md` §0.1); MVP-36; `examples/srt/`/`examples/dash/`/
`examples/webrtc/` als Vorlage.

Ziel: Ein eigenständiger SRS-Lab-Pfad analog zu den anderen
Multi-Protocol-Beispielen. Kein produktiver Telemetriepfad; opt-in
Smoke prüft Compose-Stack-Boot und Endpoint-Statussatz.

DoD:

- [x] `examples/srs/compose.yaml` (neu): SRS-Container
  (`ossrs/srs:5` gepinnt auf Major-Tag) mit RTMP-Listener (1935),
  HTTP-FLV (8088), HTTP-API (1985); FFmpeg-Publisher
  (`jrottenberg/ffmpeg:8.1-ubuntu2404`) pushed RTMP-Stream über
  das Compose-interne Netzwerk an `rtmp://srs:1935/live/srs-test`.
  Project-Name `mtrace-srs`. Eigene minimale `examples/srs/srs.conf`
  aktiviert HTTP-API auf `1985`, HTTP-Server auf `8088` und
  `http_remux` für `[vhost]/[app]/[stream].flv` (Tranche-2-Commit).
- [x] Host-Port-Schnitt kollisionsfrei zu Core-Lab/`mtrace-srt`/
  `mtrace-dash`/`mtrace-webrtc`: `1935/tcp` (RTMP) + `8088/tcp`
  (HTTP-FLV) + `1985/tcp` (HTTP-API). `docs/user/local-development.md`
  §2.7 Beispiele-Tabelle und Port-Quickref um `mtrace-srs`-Zeile
  erweitert; Beispiele-Spalte zusätzlich um den `make
  smoke-webrtc-stats-drift`-Eintrag aus Tranche 1 ergänzt
  (Tranche-2-Commit).
- [x] `examples/srs/README.md` auf 7-Punkt-Standard analog
  `examples/srt/`/`examples/dash/`/`examples/webrtc/`: Zweck,
  Voraussetzungen, Start, Verifikation, Stop/Reset, Troubleshooting,
  Bekannte Grenzen; verlinkt auf `examples/README.md`-Konvention
  (Project-Name-Pflicht). Markiert MVP-36 als „eingelöst, MVP-
  Priorität bleibt Kann"; nennt explizit Out-of-Scope
  (`mtrace_srs_*`-Counter, Player-SDK-Hookup, HLS-/DASH-/WebRTC-
  Output) (Tranche-2-Commit).
- [x] `make smoke-srs` (neu) startet `mtrace-srs`-Stack via
  `docker compose up -d --build`, prüft endpoint-/compose-only:
  (1) SRS-HTTP-API antwortet 200 auf `/api/v1/streams/`,
  (2) Stream `live/srs-test` ist registriert mit
  `publish.active=true`, (3) HTTP-FLV-Egress
  `http://localhost:8088/live/srs-test.flv` liefert 200 plus
  FLV-Magic-Header (`FLV`-Bytes). Skript
  `scripts/smoke-srs.sh` mit `SMOKE_SRS_AUTOSTART=0`-Modus für
  manuelle Aufrufe; Cleanup auf `mtrace-srs`-Project beschränkt.
  Opt-in (NICHT in `make gates`); Help-Eintrag analog `smoke-srt`/
  `smoke-dash` (Tranche-2-Commit).
- [x] `examples/README.md` Smoke-Tabelle und Beispiele-Tabelle
  um SRS-Eintrag erweitert (Tranche `—` weil außerhalb der
  `0.5.0`-Tranchen-Numerik; Status verweist auf `0.9.0` Tranche 2)
  (Tranche-2-Commit).

---

## 4. Tranche 3 — DASH-Manifest-Analyse (MVP-37 / NF-12 / RAK-58/59)

Bezug: Lastenheft §8.x NF-12 (DASH-Analyse, Muss); §12.3 MVP-37;
`done/plan-0.3.0.md` (Stream-Analyzer HLS-Stand);
`packages/stream-analyzer/src/`;
`spec/contract-fixtures/analyzer/`.

Ziel: Der `@npm9912/stream-analyzer` versteht DASH-MPD-Eingaben
zusätzlich zu HLS-Manifesten. Das JSON-Result-Schema bekommt
`analyzerKind: "dash"` als zweiten Wert; HLS-Pfad bleibt
unverändert. CLI dispatcht automatisch.

DoD:

- [x] DASH-Detector in
  `packages/stream-analyzer/src/internal/parsers/detect.ts`:
  XML-Header-Sniffing (`<?xml`/`<MPD`) plus optionaler BOM-Strip;
  liefert `"dash"`, `"hls"` oder `"unsupported"` plus erste
  nicht-leere Zeile (max. 80 Zeichen) für Diagnose-Findings
  (Tranche-3a-Commit).
- [x] Manifest-Loader von HLS-only auf HLS+DASH generalisiert
  (`packages/stream-analyzer/src/internal/loader/fetch.ts`,
  Funktion `loadManifest`): Content-Type-Allowlist um
  `application/dash+xml` / `application/xml` / `text/xml`
  erweitert, `Accept`-Header listet alle drei DASH-Typen vor
  `text/plain;q=0.9`. Fehlertext `Content-Type "<X>" ist kein
  unterstütztes Manifest-Format (HLS/DASH)` statt der
  HLS-spezifischen Variante; bestehende SSRF-/Größen-/Redirect-
  Regeln unverändert (Tranche-3a-Commit).
- [x] Fehlercode-Strategie festgelegt und umgesetzt:
  `manifest_not_hls` bleibt nur für den HLS-Parser-/HLS-Kompat-
  Pfad erhalten (HLS-Detector hat klassifiziert, HLS-Parser hat
  abgelehnt); `manifest_not_supported` als additiver Public-Code
  für Eingaben ohne HLS-/DASH-Marker in
  `packages/stream-analyzer/src/types/error.ts`,
  `docs/user/stream-analyzer.md` §2.3,
  `apps/api/hexagon/domain/stream_analysis.go`
  (`StreamAnalysisManifestNotSupported`-Konstante), HTTP-Status-
  Mapping (`domainHTTPStatus` → 422 für beide), API-Metrik-
  Allowlist (`normalizeAnalyzeCode`) und CLI/API-Tests
  (Tranche-3a/3c-Commits).
- [x] MPD-Parser
  (`packages/stream-analyzer/src/internal/parsers/dash.ts`)
  parst `MPD/Period/AdaptationSet/Representation`-Hierarchie
  regex-basiert (keine externe XML-Dependency). Mindest-Felder im
  Result: `playlistType: "dash"`, `summary.itemCount` (Anzahl
  Representations über alle Periods/AdaptationSets),
  `details.adaptationSets[].representations[]` mit `bandwidth`
  (Pflicht laut MPEG-DASH §5.3.5; fehlend → Error-Finding),
  `width`/`height` (optional, Audio-Streams haben sie nicht),
  `codecs`, `mimeType` (mit Inheritance vom AdaptationSet-Level).
  `details.type` aus `MPD@type` (`static`/`dynamic`, Default
  `static`); `details.live = type === "dynamic"`. Out-of-Scope:
  SegmentTemplate-`$Time$`-Variablen, `availabilityStartTime`-
  Drift (Plan §0.3) (Tranche-3a-Commit).
- [x] `analyzerKind: "dash"` ist in `spec/contract-fixtures/
  analyzer/` mit zwei neuen Beispielen verankert:
  `success-dash-vod.json` (VOD-MPD, `type=static`, on-demand-
  Profil, 2 video + 1 audio Representation, itemCount=3) und
  `success-dash-live.json` (Live-MPD, `type=dynamic`, `live=true`,
  `minimumUpdatePeriod=PT2S`, `availabilityStartTime`, 1 video
  Representation, itemCount=1). `spec/contract-fixtures/analyzer/
  README.md` listet beide; `packages/stream-analyzer/tests/
  contract.test.ts` pinnt jede Fixture mit byte-equal-Test gegen
  einen synthetischen MPD-Source-String; `make sync-contract-
  fixtures` kopiert die zwei Files synchron als
  `contract-success-dash-{vod,live}.json` nach `apps/api/.../
  testdata/`; `make generated-drift-check` validiert die Kopien
  (Tranche-3b-Commit). Kein Update von
  `contracts/event-schema.json` (Playback-Event-Meta-Vertrag,
  nicht Analyzer-Result).
- [x] HLS-Pfad bleibt unverändert: bestehende
  `contract-success-master.json` und alle `0.3.0`-Tests bleiben
  grün; DASH-Pfad ist additiv. Drei `analyze.test.ts`-Tests
  (Whitespace-only / leeres Manifest / HTML-Body), die zuvor
  `manifest_not_hls` erwarteten, sind auf `manifest_not_supported`
  aktualisiert — der Detector klassifiziert diese Inputs jetzt vor
  dem HLS-Parser ab (Tranche-3a-Commit).
- [x] CLI-Pfad: `pnpm m-trace check <url-or-file.mpd>` detektiert
  MPD über den gemeinsamen Detector und liefert DASH-Result; CLI-
  Code selbst entscheidet nichts. Tests in
  `packages/stream-analyzer/tests/cli.test.ts` decken DASH-File-
  Pfad, DASH-URL-Pfad und `manifest_not_supported`-Fehlerpfad
  parallel zu den HLS-Tests (Tranche-3a-Commit).
- [x] `make smoke-cli` erweitert (`scripts/smoke-cli.sh`): neuer
  Block 3 prüft `m-trace check <vod.mpd>` → `analyzerKind=dash` /
  `playlistType=dash` plus mindestens ein `details.adaptationSets[]`-
  Eintrag; vorheriger Block 3 (HTML-Body) auf
  `manifest_not_supported` umgestellt; bestehende HLS-Master-/
  SSRF-/IO-Smoke-Pfade unverändert. Live verifiziert
  (Tranche-3d-Commit).
- [x] `apps/api`-Adapter
  (`adapters/driven/streamanalyzer/http.go`): HTTP-Adapter
  übernimmt `analyzerKind` aus dem Analyzer-Result ins Domain-
  Modell (`StreamAnalysisResult.AnalyzerKind` als neuer
  `AnalyzerKind`-Type-Domain-Field, plus `mapAnalyzerKind`-Helper);
  Driving-HTTP (`analyze.go`) gibt `analysis.analyzerKind` aus
  `result.AnalyzerKind` aus statt der HLS-Konstante.
  `playlistType: "dash"` als additiver Domain-/Wire-Wert
  durchgereicht (`PlaylistTypeDash` in `domain/stream_analysis.go`,
  `mapPlaylistType`-Erweiterung um `case "dash"`); `unknown`-Pfad
  unverändert. Adapter-Tests in `contract_test.go` decken VOD- und
  Live-Fixture mit `AnalyzerKindDASH`/`PlaylistTypeDash`-Assertions
  ab; HLS-Tests bleiben grün (Tranche-3c-Commit).

---

## 5. Tranche 4 — Compat-Tests + Doku-Pflege

Bezug: `done/plan-0.8.0.md` §5 (Tranche-4-Vorlage); `packages/
player-sdk/README.md` Browser-Support-Matrix.

Ziel: Pack-Smoke und CLI-Smoke spiegeln die neuen Liefergegenstände;
Browser-Support-Matrix-Pflege; CI-Policy bleibt explizit.

DoD:

- [ ] Pack-Smoke (`packages/stream-analyzer/scripts/`?): falls
  Stream-Analyzer ein eigenes Pack-Smoke hat, prüft er den
  DASH-Pfad in ESM/CJS. Wenn nicht, Folge-DoD analog
  `0.8.0` Tranche 4 für Player-SDK — kein Stream-Analyzer-
  Pack-Smoke in `0.9.0` Pflicht.
- [ ] `packages/stream-analyzer/README.md` (oder neuer Abschnitt)
  dokumentiert DASH-Eingabeform und CLI-Dispatcher-Logik.
- [ ] `examples/README.md` listet `smoke-srs` konsistent.
- [ ] `docs/user/local-development.md` §2.7 Port-Quickref mit
  `mtrace-srs`-Eintrag.
- [ ] `docs/user/releasing.md` neue §2.4 für `0.9.0`-Smokes
  (Drift-Smoke + DASH-Smoke + SRS-Smoke) als opt-in im Release-
  Pfad analog `smoke-srt-health`/`smoke-webrtc-prep`.

---

## 6. Tranche 5 — Release-Doku, RAK-Matrix und Closeout

Bezug: RAK-56..RAK-59; `docs/user/releasing.md`; `README.md`;
`roadmap.md`.

Ziel: `0.9.0` ist auffindbar dokumentiert, Versions-Bump
durchgezogen, Tag `v0.9.0` gesetzt.

DoD:

- [ ] `README.md` Status-Block auf „`0.9.0` released" und
  Verweise auf `examples/srs/` plus DASH-Analyzer-Pfad.
- [ ] `docs/user/releasing.md` neue §2.4 mit manuellen `0.9.0`-
  Prüfungen (DASH-CLI-Probe, SRS-Lab-Boot, Drift-Smoke-Trigger).
- [ ] RAK-Verifikationsmatrix §6.1 (siehe unten) ist mit Commit-
  Verweisen ausgefüllt.
- [ ] Versions-Bump 0.8.0 → 0.9.0 in allen package.json (root,
  apps, packages) plus `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts`, `packages/player-sdk/
  scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version` und allen Test-
  Fixtures (analog `0.8.0` Tranche 5; der hartkodierte Tarball-
  Pfad in `packages/player-sdk/package.json` Script `pack:smoke`
  ist ausdrücklich mitzuprüfen). Zusätzlich alle hartkodierten
  Analyzer-/API-/Dashboard-Test-Erwartungen nachziehen, insbesondere
  `packages/stream-analyzer/tests/version.test.ts`,
  `packages/stream-analyzer/tests/cli.test.ts`,
  `apps/analyzer-service/tests/server.test.ts`,
  `apps/api/adapters/driven/streamanalyzer/*_test.go`,
  `apps/api/adapters/driving/http/*analyze*_test.go` und weitere
  `sdk.version`-/`analyzerVersion`-Fixtures, soweit sie den
  Release-Stand statt historische Kompatibilitätsfälle pinnen.
- [ ] CHANGELOG: [Unreleased]-Block in `[0.9.0] - YYYY-MM-DD`
  umgewandelt; neuer leerer [Unreleased]-Block obenauf.
- [ ] `./scripts/verify-doc-refs.sh` (`make docs-check`) grün
  vor Closeout-Commit; `make gates` grün.
- [ ] `plan-0.9.0.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben (`git mv`); alle relativen
  Cross-Refs angepasst (analog `0.8.0` Closeout plus Release-Gate-
  Fix); Roadmap §3 zeigt `0.9.0` ✅.
- [ ] Tag `v0.9.0` annotiert; Push opt-in (User-Bestätigung);
  GitHub-Release mit CHANGELOG-`[0.9.0]`-Block als Notes-Body.

### 6.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-56 | Soll | `tests/e2e/webrtc-stats-drift.spec.ts` plus `make smoke-webrtc-stats-drift`; Nightly-CI-Job; R-12 im Risiken-Backlog auf „automatisiert detektiert" angehoben. | [ ] |
| RAK-57 | Kann | `examples/srs/compose.yaml` (Project `mtrace-srs`) plus `make smoke-srs`; `examples/srs/README.md` 7-Punkt-Standard; Port-Quickref nachgezogen. | [ ] |
| RAK-58 | Muss | `@npm9912/stream-analyzer` versteht DASH-MPD; `analyzerKind: "dash"` mit Analyzer-Contract-Fixtures, Go-Testdata-Sync und API-Durchreichung; HLS-Pfad unverändert. | [ ] |
| RAK-59 | Kann | `pnpm m-trace check <file.mpd>` dispatcht auf DASH und liefert valides Result; `make smoke-cli` erweitert. | [ ] |

---

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen (analog `done/plan-0.8.0.md` §7).
- Lastenheft-Patch `1.1.11` (siehe §0.2) ist Vorgänger-Gate für
  Tranchen 1–4; bis Tranche 0 abgeschlossen ist, sind RAK-56..
  RAK-59 nur geplante RAK aus dem Patch-Vorschlag.
- Wenn ein `0.9.0`-Item in einer Folge-Phase neu bewertet wird
  (z. B. ein Soll-Pfad doch als Muss eingelöst werden muss),
  entweder Folgeplan eröffnen oder hier als Wartungs-Eintrag
  vermerken.
- R-12 wechselt mit Tranche 1 von „release-blockierend ab
  nächstem Browser-Major-Bump" auf „automatisiert detektiert,
  Drift bricht den Drift-Smoke"; Risiken-Backlog-Eintrag muss im
  selben Commit wie die Drift-Smoke-Implementation nachgezogen
  werden.
- Wenn die drei Themen unterschiedliches Tempo haben, ist es
  zulässig, Tranche 1/2/3 in mehreren Sub-Releases (`0.9.0`,
  `0.9.1`, `0.9.2`) auszuliefern statt in einem einzigen `0.9.0`-
  Release; in dem Fall wird der Plan vor Tranche 5 in mehrere
  Plan-Dateien umstrukturiert.
