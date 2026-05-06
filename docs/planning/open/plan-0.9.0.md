# Implementation Plan — `0.9.0` (Drift-Smoke + SRS-Lab + DASH-Analyse)

> **Status**: ⬜ geplant (Plan-Skelett, liegt unter
> `docs/planning/open/`). Vorgänger `v0.8.0` ist released
> (Tag `v0.8.0` auf dem Release-Gate-Fix nach `8df263a`; Plan
> archiviert in [`done/plan-0.8.0.md`](../done/plan-0.8.0.md)).
> Tranche 0 aktiviert die Phase, sobald der zugehörige Lastenheft-
> Patch `1.1.11` fertig ist (siehe §0.2). Plan wandert dann atomar
> nach `docs/planning/in-progress/`.
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
> [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
> §3.5.3 (WebRTC-Schema-Drift-Strategie);
> [`docs/planning/open/risks-backlog.md`](./risks-backlog.md) R-12;
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

### 0.2 Lastenheft-Patch `1.1.11` (Vorschlag)

Der Patch ergänzt vier neue RAK in einem neuen §13.11-Block und
zieht MVP-37 entsprechend NF-12 auf „Muss". Genaue Wortlaute werden
beim Tranche-0-Closeout in `spec/lastenheft.md` und im Wartungslog
(neuer §4a.14 in `done/plan-0.1.0.md`) festgehalten:

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
  [`plan-0.8.5.md`](../in-progress/plan-0.8.5.md) (Wave 1, vor `0.9.0`) und
  [`plan-0.9.5.md`](./plan-0.9.5.md) (Wave 2, nach `0.9.0`)
  konkretisiert; Master-Backlog steht in
  [`extra-gates.md`](./extra-gates.md).

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
| 0 | Plan-Aktivierung (`open/` → `in-progress/`) + Lastenheft-Patch `1.1.11` (RAK-56..RAK-59 + MVP-37-Hochstufung) + ggf. Toolchain-Hardening | ⬜ |
| 1 | Browser-Drift-Smoke für WebRTC-`getStats()` (RAK-56) | ⬜ |
| 2 | SRS-Lab `examples/srs/` (RAK-57, MVP-36) | ⬜ |
| 3 | DASH-Manifest-Analyse im `@npm9912/stream-analyzer` (RAK-58/RAK-59, MVP-37, NF-12) | ⬜ |
| 4 | Compat-Tests + Browser-Support-Matrix-Pflege; Pack-Smoke + CLI-Smoke erweitert | ⬜ |
| 5 | Release-Doku, RAK-Verifikationsmatrix und Closeout (Versions-Bump 0.8.0 → 0.9.0, Plan nach `done/`, Tag `v0.9.0`) | ⬜ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Lastenheft-Patch

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche analog
`done/plan-0.8.0.md` §1a.

DoD:

- [ ] Plan-Skelett von `docs/planning/open/plan-0.9.0.md` nach
  `docs/planning/in-progress/plan-0.9.0.md` verschoben (Status
  `⬜ → 🟡`); Cross-Refs in `roadmap.md` §1.2/§3 nachgezogen.
- [ ] Lastenheft-Patch `1.1.11` schreiben: §13.11 neu mit RAK-56..
  RAK-59; §12.3 MVP-37 von „Kann" auf „Muss" hochgezogen entsprechend
  NF-12 (Hinweis: §12.3 historisch beibehalten mit Patch-Note).
  Patch-Eintrag als §4a.14 in `done/plan-0.1.0.md` Tranche 0c.
- [ ] Toolchain-Bump-Check (Go/Node/pnpm/golangci-lint). Wenn kein
  Bump nötig, dokumentieren warum nicht.

---

## 2. Tranche 1 — Browser-Drift-Smoke (R-12)

Bezug: `risks-backlog.md` R-12; `spec/telemetry-model.md` §1.4
(webrtc.*-Allowlist) + §3.5.2/§3.5.3; `tests/e2e/`.

Ziel: Ein automatisierter Smoke, der den `getStats()`-Schema-Drift
in echten Browser-Versionen frühzeitig erkennt. Schließt R-12
operativ — der Drift-Review-Gate ist nicht mehr manuelle Pflicht
vor jedem Release-Tag, sondern auto-detektiert.

DoD:

- [ ] `tests/e2e/webrtc-stats-drift.spec.ts` (neu, Playwright):
  rendert `/demo-webrtc?autostart=1` mit aktivem `mtrace-webrtc`-
  Stack; nach erfolgreichem Handshake ruft die Spec `pc.getStats()`
  auf (entweder direkt im Page-Context oder über einen Hook-Test-
  Endpoint im Adapter) und sammelt alle Reports.
- [ ] Spec validiert für jede `RTCStatsType`-Gruppe aus §3.5.2,
  dass alle Muss-Felder existieren; Drift bricht den Smoke mit
  klarer Fehlermeldung („Browser X.Y dropped field Z from
  RTCStatsType.foo"). Soll-Felder werden geloggt aber nicht
  release-blockierend geprüft.
- [ ] Spec validiert, dass alle gefundenen `connectionState`/
  `iceConnectionState`/`dtlsState`-Werte in der §1.4-Allowlist
  liegen. Unbekannter Enum-Wert → Smoke-Fail.
- [ ] `make smoke-webrtc-stats-drift`-Target opt-in (nicht in
  `make gates`); Help-Eintrag analog `smoke-webrtc-prep`. Default-
  Browser sind Chromium und Firefox aus dem Playwright-Bundle;
  `MTRACE_WEBRTC_DRIFT_BROWSERS=chromium,firefox,webkit` toggelt
  Safari/WebKit opt-in.
- [ ] CI-Workflow `.github/workflows/webrtc-drift.yml` (neu,
  Nightly via `schedule: cron`): startet `mtrace-webrtc`-Stack,
  führt den Smoke aus, eröffnet bei Failure automatisch ein
  Issue mit Browser-Version und Drift-Befund (gh-cli oder
  Action). Issue-Auto-Erstellung ist opt-in über
  `secrets.DRIFT_AUTO_ISSUE`.
- [ ] R-12 im `risks-backlog.md` von „release-blockierend ab
  nächstem Browser-Major-Bump" auf „automatisiert detektiert,
  Drift bricht den Drift-Smoke" angehoben; Release-Pfad in
  `releasing.md` referenziert den Drift-Smoke.

---

## 3. Tranche 2 — SRS-Lab `examples/srs/` (MVP-36 / RAK-57)

Bezug: `examples/README.md` (Multi-Protocol-Lab-Konvention,
`plan-0.5.0.md` §0.1); MVP-36; `examples/srt/`/`examples/dash/`/
`examples/webrtc/` als Vorlage.

Ziel: Ein eigenständiger SRS-Lab-Pfad analog zu den anderen
Multi-Protocol-Beispielen. Kein produktiver Telemetriepfad; opt-in
Smoke prüft Compose-Stack-Boot und Endpoint-Statussatz.

DoD:

- [ ] `examples/srs/compose.yaml` (neu): SRS-Container
  (`ossrs/srs:5` gepinnt) mit RTMP-Listener (1935), HTTP-FLV
  (8088), HTTP-API (1985); FFmpeg-Publisher pushed RTMP-Stream
  via lokalem Compose-Netzwerk. Project-Name `mtrace-srs`.
- [ ] Host-Port-Schnitt kollisionsfrei zu Core-Lab/`mtrace-srt`/
  `mtrace-dash`/`mtrace-webrtc`: voraussichtlich `1935/tcp` (RTMP)
  + `8088/tcp` (FLV) + `1985/tcp` (API). Aktualisierung von
  `docs/user/local-development.md` §2.7 Port-Quickref.
- [ ] `examples/srs/README.md` auf 7-Punkt-Standard mit Start/
  Verifikation/Stop/Troubleshooting/Bekannte Grenzen; verlinkt
  auf `examples/README.md` Konvention.
- [ ] `make smoke-srs` (neu) startet `mtrace-srs`-Stack, prüft
  HTTP-API-Erreichbarkeit + FFmpeg-Stream-Registrierung; opt-in
  (nicht in `make gates`); Cleanup auf `mtrace-srs`-Project
  beschränkt.
- [ ] `examples/README.md` Smoke-Tabelle und Beispiele-Tabelle
  um SRS-Eintrag erweitert.

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

- [ ] DASH-Detector in `packages/stream-analyzer/src/`: XML-Header-
  Sniffing (`<?xml`/`<MPD`) plus optionale Content-Type-Heuristik
  (`application/dash+xml`). Gibt `"dash"` oder `"hls"` zurück.
- [ ] Manifest-Loader von HLS-only auf HLS+DASH generalisiert:
  Content-Type-Allowlist und `Accept`-Header enthalten
  `application/dash+xml`; Funktions-/Fehlertexte sprechen von
  Manifest statt „HLS-Manifest"; URL-Tests decken DASH-Content-Type
  und weiterhin geblockte Nicht-Manifest-Typen ab.
- [ ] Fehlercode-Strategie festgelegt und umgesetzt:
  `manifest_not_hls` bleibt nur für den HLS-Parser-/HLS-Kompat-
  Pfad erhalten; für Eingaben, die weder HLS noch DASH sind, kommt
  ein additiver Public-Code (z. B. `manifest_not_supported`) in
  `packages/stream-analyzer/src/types/error.ts`,
  `docs/user/stream-analyzer.md`, `apps/api/hexagon/domain/
  stream_analysis.go`, HTTP-Status-Mapping, API-Metrik-Allowlist
  und CLI/API-Tests. Fehlermeldungen dürfen nicht mehr behaupten,
  eine DASH-MPD sei „kein HLS-Manifest".
- [ ] MPD-Parser parst `MPD/Period/AdaptationSet/Representation/
  SegmentTemplate`-Hierarchie. Mindest-Felder im Result:
  `playlistType` (`"dash"`), `summary.itemCount` (Anzahl
  Representations), `details.adaptationSets` (Array mit
  `mimeType`, `codecs`, `bandwidth`, `width`/`height`).
- [ ] `analyzerKind: "dash"` ist in `spec/contract-fixtures/
  analyzer/` mit zwei neuen Beispielen verankert: ein VOD-MPD
  und ein einfaches Live-MPD. `spec/contract-fixtures/analyzer/
  README.md`, `packages/stream-analyzer/tests/contract.test.ts`,
  `apps/api/adapters/driven/streamanalyzer/testdata/` und
  `apps/api/adapters/driven/streamanalyzer/contract_test.go`
  werden synchron erweitert; `make sync-contract-fixtures` kopiert
  auch die neuen DASH-Fixtures. Kein Update von
  `contracts/event-schema.json`: diese Datei gehört zum Playback-
  Event-Meta-Vertrag, nicht zum Analyzer-Result.
- [ ] HLS-Pfad bleibt unverändert: bestehende
  `contract-success-master.json` und alle `0.3.0`-Tests bleiben
  grün; DASH-Pfad ist additiv.
- [ ] CLI-Pfad (`packages/stream-analyzer/src/cli/`): `pnpm
  m-trace check <url-or-file.mpd>` detektiert MPD und liefert
  DASH-Result. Tests in `packages/stream-analyzer/tests/cli.test.ts`
  decken HLS- und DASH-Pfad parallel.
- [ ] `make smoke-cli` erweitert: zusätzlich zu HLS-Smoke wird
  ein DASH-MPD-Beispiel geprüft.
- [ ] `apps/api`-Adapter (`adapters/driven/streamanalyzer/`):
  HTTP-Adapter übernimmt `analyzerKind` aus dem Analyzer-Result ins
  Domain-Modell; Driving-HTTP gibt `analysis.analyzerKind: "dash"`
  statt der bisherigen HLS-Konstante aus. `playlistType: "dash"`
  wird als additiver Domain-/Wire-Wert durchgereicht, nicht auf
  `unknown` normalisiert. Tests decken Adapter-Parsing,
  `/api/analyze`-Response und HLS-Backward-Compat ab.

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
