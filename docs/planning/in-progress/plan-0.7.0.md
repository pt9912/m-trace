# Implementation Plan — `0.7.0` (WebRTC-Lab-Erweiterung)

> **Status**: 🟡 in Arbeit (Tranche 0 aktiviert am 2026-05-05; Vorgänger
> `v0.6.0` ist released, Tag `d08a89f`). Liegt unter
> `docs/planning/in-progress/`. Sammelt die Folge-Schritte aus
> `0.5.0` Tranche 5 und macht aus dem WebRTC-Vorbereitungspfad eine
> produktive Lab-Erweiterung.
>
> **Lastenheft-Status**: ausgeliefert in `1.1.9` — §13.9 enthält
> RAK-47..RAK-51 für die produktive WebRTC-Lab-Erweiterung; RAK-39
> (in `0.5.0`-Block §13.7) deckt unverändert den Vorbereitungspfad
> ab. Lastenheft-Patch ist als §4a.12 in
> [`docs/planning/done/plan-0.1.0.md`](../done/plan-0.1.0.md)
> Tranche 0c dokumentiert.
>
> **Bezug**: [Lastenheft `1.1.9`](../../../spec/lastenheft.md) §7.6
> F-62 (Player-Adapter-Folgeoptionen), §8.3 NF-14 (Erweiterbarkeit),
> §12.1 MVP-24, §13.7 RAK-39 (Vorbereitungspfad), §13.9 RAK-47..RAK-51
> (Lab-Erweiterung); [`done/plan-0.5.0.md`](../done/plan-0.5.0.md)
> §6 (Tranche 5);
> [`examples/webrtc/README.md`](../../../examples/webrtc/README.md)
> (aktueller Doku-only Stand);
> [`plan-0.6.0.md`](../done/plan-0.6.0.md) (vorhergehende Phase).
> Beim Verschieben von `0.6.0` nach `docs/planning/done/` muss dieser
> Link im selben Commit auf `../done/plan-0.6.0.md` umgestellt werden.
>
> **Nachfolger**: `plan-0.8.0.md` (offen).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`done/plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene Entscheidung.
- 🟡 in Arbeit.

Scope-Grenze: dieser Plan beschreibt eine **produktive WebRTC-Lab-
Erweiterung** als Folge-Pfad zum Vorbereitungs-Skelett aus `0.5.0`.
Er liefert einen lokal startbaren WHIP-/WHEP-Lab-Pfad und einen
opt-in Smoke für die Vorbereitungsgrenze. Er liefert keine produktive
WebRTC-Telemetrie im Ingest-Pfad und keinen Player-SDK-WebRTC-Adapter.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- `0.5.0` ist released (Tag `v0.5.0` auf `a56dc0b`); WebRTC-Skelett
  unter [`examples/webrtc/`](../../../examples/webrtc/) existiert.
- `0.6.0` (SRT Health View) ist released; siehe
  [`plan-0.6.0.md`](../done/plan-0.6.0.md). Wenn der
  Vorgängerplan im Release-Closeout bereits nach `done/` verschoben ist,
  wird dieser Link beim Aktivieren von `0.7.0` atomar auf
  `../done/plan-0.6.0.md` korrigiert.
- Lastenheft `1.1.9` ist um RAK-47..RAK-51 für `0.7.0` erweitert (§13.9 + §4a.12 in `done/plan-0.1.0.md`); siehe §0.2.

`smoke-webrtc-prep` (RAK-48 Muss) ist endpoint-/compose-only und
benötigt **keinen** headless-Browser-Pfad — Headless-Browser-
Stabilität ist daher kein Vorgänger-Gate. Eine separate
Browser-Stabilitätsbewertung wird in Tranche 3 ausschließlich für
optionale Erweiterungen (Browser-Handcheck, getStats()-Sammlung)
geführt und kann das Muss-Target nicht aus dem Scope kippen.

### 0.2 Lastenheft-Erweiterung (ausgeliefert in `1.1.9`)

`spec/lastenheft.md` `1.1.9` enthält in §13.9 die RAK-Pflicht-/Soll-/
Kann-Aussagen für `0.7.0`. Patch ist als §4a.12 in
[`done/plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c
dokumentiert. Inhalt für die Tranchen-DoD unten:

| RAK | Priorität | Inhalt (`spec/lastenheft.md` §13.9) |
| --- | --------- | ----------------------------------- |
| RAK-47 | Muss | WebRTC-Lab-Setup mit lokalem WHIP-/WHEP-Endpoint und Compose-Stack `mtrace-webrtc`. |
| RAK-48 | Muss | `make smoke-webrtc-prep`-Target prüft Vorbereitungsgrenze (Endpoints antworten, kein Playback-/`getStats()`-Anspruch). |
| RAK-49 | Soll | `getStats()`-Subset ist als bounded Allowlist in `spec/telemetry-model.md` §3.2 dokumentiert; Schema-Drift-Strategie zwischen Browser-Versionen ist definiert. |
| RAK-50 | Kann | Browser-Handcheck ist in `examples/webrtc/README.md` als manueller Verifikationspfad dokumentiert. |
| RAK-51 | Kann | Player-SDK exposed einen optionalen WebRTC-Adapter-Pfad ohne Vermischung mit `hls.js`. **Nicht Teil dieses Plans** — Lastenheft-Anker für eine Folge-Tranche/-Release; DoD wird in einem späteren Plan formuliert. |

Die DoD-Items in §1–§6 unten referenzieren RAK-47..RAK-50
verbindlich; RAK-51 bleibt bewusst ohne Tranche und ist im
[`plan-0.7.0.md`](./plan-0.7.0.md)-Wartungsblock §7 als Folge-
Punkt vermerkt.

### 0.3 Out-of-Scope-Klauseln (durchgängig)

- Keine produktive `getStats()`-Sammlung im `apps/api`-Ingress, solange
  Schema-Drift zwischen Browser-Versionen (Chromium/Firefox/Safari)
  nicht durch eigene Tranche bewertet ist.
- Keine WebRTC-Aggregat-Metriken in Prometheus, bis bounded Allowlist-
  Labels in [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
  §3.2 erweitert sind.
- Keine Vermischung mit dem `hls.js`-Demo-Pfad in `apps/dashboard`;
  WebRTC bekommt eine eigene Demo-Route oder bleibt im Lab-Beispiel.
- Kein eigener Signaling-Service in `apps/api`. `0.7.0` nutzt einen
  Lab-Media-Server mit WHIP-/WHEP-Endpunkten; falls ein eigener
  Signaling-Pfad nötig wird, ist das ein Folgeplan.
- Kein TLS-/Public-Internet-/NAT-Produktionssetup. localhost ist der
  Pflichtpfad; STUN/TURN ist optional und darf RAK-47/RAK-48 nicht
  blockieren.

### 0.4 Sequenzierung und harte Gates

1. Tranche 1 entscheidet Media-Server-Image, Endpoint-Pfade, Ports und
   Publisher-Mechanik. Tranche 2/3 dürfen keine impliziten WebRTC-
   Defaults voraussetzen.
2. Tranche 3 ist endpoint-/compose-only. Ein Browser- oder Playback-
   Handcheck kann RAK-50 unterstützen, darf aber RAK-48 nicht ersetzen
   und darf `make smoke-webrtc-prep` nicht flakig machen.
3. Tranche 4 ist Spec-Arbeit für spätere produktive Telemetrie in
   `spec/telemetry-model.md`. Sie darf keine `mtrace_webrtc_*`-Metrik in
   Code oder Prometheus behaupten und keine Player-SDK-/Adapter-Public-
   API spezifizieren, solange kein produktiver WebRTC-Telemetriepfad
   existiert.
4. RAK-51 bleibt Folge-Scope. Jede Änderung an `packages/player-sdk`
   braucht eine eigene Public-API-Tranche mit hls.js-Kompatibilitäts-
   nachweis.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung (open/ → in-progress/) + Toolchain-Hardening (Go 1.22 → 1.26, golangci-lint v1.62 → v2.12.1, neue `race`-Stage in gates) | ✅ |
| 1 | Lab-Compose `examples/webrtc/compose.yaml` mit MediaMTX-WHIP/-WHEP plus FFmpeg-RTSP-Publisher | ✅ |
| 2 | README-Konkretisierung — Operator-Befehle, Port-Schnitt, Browser-Handcheck | ✅ |
| 3 | `make smoke-webrtc-prep`-Target mit reservierter Vorbereitungs-Verifikation | ✅ |
| 4 | WebRTC-Telemetrie-Bewertung — bounded Allowlist, `getStats()`-Subset, Schema-Drift-Strategie | ✅ |
| 5 | Release-Doku, RAK-Matrix und Closeout | ⬜ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Toolchain-Hardening

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche, die `0.7.0`
auf eine non-EOL Toolchain stellt und `make gates` um eine
Race-Detector-Stage erweitert. Sammelt drei Lieferungen, die
**vor** Tranche 1 (echtem Lab-Compose-Code) ausgeführt sind, damit
die nachfolgenden Tranchen auf einer aktuellen Build-/Test-Basis
aufsetzen.

DoD:

- [x] Plan-Skelett von `docs/planning/open/plan-0.7.0.md` nach
  `docs/planning/in-progress/plan-0.7.0.md` verschoben; Status
  `⬜ geplant` → `🟡 in Arbeit`. Cross-Refs in Roadmap §0/§1/§3,
  README, examples/webrtc/README.md, plan-0.1.0.md Tranche 0c und
  plan-0.5.0.md Tranche 6 nachgezogen. Sed-Drift-Artefakte in
  plan-0.1.0/plan-0.5.0 (`docs/planni../in-progress/...`)
  bereinigt. Commit `3156ef4`.
- [x] Go-Toolchain-Bump auf aktuelle non-EOL-Linie:
  `apps/api/go.mod` `go 1.22.7` (EOL Februar 2025) → `go 1.26.0`
  (Release 2026-02-10); `apps/api/Dockerfile` `golang:1.22` →
  `golang:1.26` für deps/test/coverage/build-Stages und
  `apps/api/Makefile` arch-check; `apps/api/README.md`-Hinweis.
  golangci-lint von `v1.62-alpine` (Sep 2024, Go 1.23) auf
  `v2.12.1-alpine` (Mai 2026, Go 1.26.2) gehoben — v1.62
  verweigert das Linting mit „Go language version (go1.23) used
  to build golangci-lint is lower than the targeted Go version
  (1.26.0)". `.golangci.yml` über `golangci-lint migrate` auf
  v2-Schema gezogen (`disable-all: true` → `default: none`,
  `gomodguard` → `gomodguard_v2`, `run.timeout` entfällt).
  staticcheck QF1006 (lift-condition-into-loop) in
  `srt_health_collector_test.go` zwei Stellen umgesetzt.
  Commit `ccf68b1`.
- [x] Race-Detector-Stage in `apps/api/Dockerfile` neu
  (`FROM deps AS race` mit `CGO_ENABLED=1 go test -race ./...`);
  `make race`-Target in `apps/api/Makefile` mit
  `--no-cache-filter race`, damit ein einmal grüner Race-Lauf
  keinen späteren flaky Race über den Layer-Cache verdeckt;
  `make api-race` im root Makefile. **In `make gates`
  aufgenommen** (`api-test` durch `api-race` ersetzt — Race ist
  Superset). Lokaler Lauf ~33 s vs. ~20 s `api-test` (~1.7×); CI-
  Mehraufwand ist tragbar. Erster Lauf hat sofort einen echten
  Race in `mockSrtHealthRepo` gefunden (Test-Helper schreibt aus
  Collector-Goroutine, Test-Body liest `len(repo.appended)` aus
  Test-Goroutine ohne Sync); Mock mit `sync.Mutex` +
  `appendedCount()`-Helper abgesichert. `releasing.md` §2
  Beschreibung des `make gates`-Inhalts nachgezogen.

**Lehre für `0.7.0`-Tranchen 1–5**: Toolchain-Pflege gehört in
eine eigene Tranche 0, nicht in Bug-Fix-Commits. Wenn die
nachfolgenden Tranchen in 0.7.0 neuen Concurrency-Code anlegen
(Lab-Compose-Adapter, Smoke-Helper), greift die Race-Stage
automatisch über `make gates`. Plan-0.6.0 §8.3 Lehre #3 (Adapter-
Felder mit Konfig-Eingang brauchen Tabellen-Test) ist ergänzend
gültig.

---

## 2. Tranche 1 — Lab-Compose

Bezug: Lastenheft §7.6 F-62, §7.8 F-82..F-84;
[`examples/webrtc/README.md`](../../../examples/webrtc/README.md)
„Voraussetzungen (geplant)".

Ziel: `examples/webrtc/compose.yaml` startet einen lokalen WebRTC-
WHIP-/WHEP-Pfad ohne Internet-Abhängigkeit. Project-Name
`mtrace-webrtc` (in `examples/README.md` reserviert).

DoD:

- [x] `examples/webrtc/compose.yaml` definiert MediaMTX mit WHIP-/WHEP-
  Listener (`webrtc: yes`, `webrtcAddress: :8889` intern) plus FFmpeg-
  RTSP-Publisher als zweiten Service. `coturn` ist als Folge-Scope in
  der README dokumentiert (LAN-Pfad), nicht in Tranche 1 enthalten.
- [x] Media-Server-Image gepinnt auf `bluenviron/mediamtx:1` (konsistent
  mit Core-Lab und `examples/srt/`). README nennt die getestete WHIP-/
  WHEP-Pfadform für MediaMTX `1.18.1` (über `:1` aktuell ausgelieferte
  Version): `http://localhost:8892/<stream>/{whip,whep}`. Drift zwischen
  Major-Versionen blockiert die Doku, der Major-Pin verhindert das im
  Lab-Pfad.
- [x] Host-Port-Schnitt entschieden: WebRTC weicht aus, weil
  `examples/srt/` (released in `0.5.0`) Host-Port `8889/tcp` belegt.
  WebRTC-Lab nutzt `8892/tcp` (WHIP/WHEP-HTTP) → MediaMTX-`8889`,
  `9999/tcp` → MediaMTX-API-`9997`. SRT-Beispiel bleibt unverändert.
- [x] ICE-/Media-Port-Schnitt explizit: `8189/udp` Host → MediaMTX-
  `webrtcLocalUDPAddress: :8189`. `webrtcAdditionalHosts:
  [127.0.0.1, mediamtx]` in `mediamtx.yml` advertised lokale + Docker-
  interne ICE-Kandidaten. TCP-/TURN-Fallback ist Folge-Scope; LAN-Pfad
  ist als RAK-50-Handcheck-Grenze in der README dokumentiert.
- [x] Host-Port-Schnitt in `examples/README.md` Tabelle „Beispiele"
  und `docs/user/local-development.md` §2.7 Port-Quickref nachgezogen:
  `mtrace-webrtc | 8892/tcp · 8189/udp · 9999/tcp`.
- [x] FFmpeg-Publisher-Mechanik entschieden: RTSP-Push in
  `examples/webrtc/ffmpeg-rtsp-loop.sh` (analog
  `examples/srt/ffmpeg-srt-loop.sh`); MediaMTX re-published den Stream
  serverintern auf den WHEP-Read-Pfad. Begründung: FFmpeg-`f whip`
  benötigt Build-Zeit-WebRTC-Stack im FFmpeg-Image, der bei
  `jrottenberg/ffmpeg:8.1` nicht zugesichert ist. RTSP-Push erfüllt
  F-84 ohne Sonder-Build und MediaMTX gewährleistet die Re-Publish-
  Brücke. Browser-WHIP-Push bleibt manueller RAK-50-Handcheck.
- [x] WHEP-Readiness browserfrei prüfbar: live verifizierter
  Statussatz in `examples/webrtc/README.md` „Verifikation":
  `OPTIONS /webrtc-test/whep` und `…/whip` → `204` bei aktivem
  Stream, `→ 500` für unbekannten Pfad, `GET/HEAD → 405`,
  `POST` ohne SDP → `400`, fehlender Compose-Stack → Connection
  refused. Tranche 3 baut ihren Smoke darauf auf.
- [x] `examples/webrtc/README.md` „Start"-Sektion zeigt den
  `docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml
  up -d --build`-Befehl plus die Port-Tabelle (8892 WHIP/WHEP, 8189
  ICE, 9999 API).
- [x] Stop/Reset auf `mtrace-webrtc` begrenzt: README zeigt
  `docker compose -p mtrace-webrtc … down [--volumes]`. Greift weder
  Core-Lab- (`mtrace`) noch SRT- (`mtrace-srt`) noch DASH-Volumes
  (`mtrace-dash`) an.

---

## 3. Tranche 2 — README-Konkretisierung

Bezug: Tranche 1; [`examples/README.md`](../../../examples/README.md)
„README-Mindeststruktur".

Ziel: `examples/webrtc/README.md` ist auf den 7-Punkt-Standard der
anderen Beispiele umgestellt — keine Doku-only-Klauseln mehr,
sondern konkrete Operator-Befehle.

DoD:

- [x] „Voraussetzungen" listet Chromium 120+/Firefox 120+, Safari als
  Best-Effort, Docker Engine ≥ 24.0, Compose v2.20, freie Host-Ports
  (8892/8189/9999) und die localhost-Pflichtgrenze. STUN/TURN-Setup
  ist als Folge-Scope vermerkt, nicht erforderlich für den
  localhost-Lab-Pfad.
- [x] README benennt den exakten WHIP-/WHEP-URL-Satz in der Tabelle
  „WHIP-/WHEP-URL-Form": Stream `webrtc-test`, Host-Port 8892,
  Container-Port 8889 (`webrtcAddress`), MediaMTX-Version `1.18.1`
  (über `:1` gepinnt).
- [x] „Verifikation" beschreibt zwei Pfade nebeneinander:
  Endpoint-Probe (browserfrei, Tranche-3-Smoke-Vorlage mit
  Statussatz-Tabelle) und manueller Browser-Handcheck (RAK-50)
  über die MediaMTX-Read-Demo-Seite.
- [x] „Stop / Reset" zeigt `down [--volumes]` ausschließlich für
  `mtrace-webrtc`. „Troubleshooting" mit echten Fehlerbildern:
  Port-Konflikt-Diagnose, `OPTIONS → 500` ohne Stream, Browser
  zeigt schwarzes Bild (ICE-Auswahl), LAN-/STUN-Grenze, FFmpeg-
  Opus-Encoder-Hinweis (`libopus` vs. native `opus`), MediaMTX-
  Auth (`-u any:`).
- [x] „Bekannte Grenzen" reduziert auf den Restscope: kein
  produktiver `apps/api`-Ingress, kein Player-SDK-WebRTC-Adapter
  (RAK-51 deferred), kein TLS/Public-Internet/LAN-NAT, kein
  Dashboard-Hook, Headless-Browser-Smoke optional.
- [x] README erklärt im Abschnitt nach der Statussatz-Tabelle
  wörtlich, dass die Probe **nicht** Playback-Qualität, **nicht**
  ICE-Erfolgsquote und **nicht** `getStats()`-Stabilität nachweist
  — diese drei Aspekte deckt nur der Browser-Handcheck ab.

---

## 4. Tranche 3 — `make smoke-webrtc-prep`-Target

Bezug: Tranche 1; [`examples/README.md`](../../../examples/README.md)
„Smoke-Targets" (Target-Name reserviert).

Ziel: Ein opt-in Smoke-Target, das ausschließlich
**Vorbereitungsgrenzen** prüft — kein Playback-Qualitäts- oder
`getStats()`-Anspruch. Verifikation läuft endpoint-/compose-only
(HTTP-Status, Container-Health) und benötigt **keinen** headless-
Browser — RAK-48 ist damit ohne Browser-Stabilitäts-Vorbedingung
erfüllbar.

DoD:

- [x] `scripts/smoke-webrtc-prep.sh` existiert, `set -euo pipefail`,
  `[smoke-webrtc-prep]`-Stderr-Präfix, `cleanup`-Trap startet/stoppt
  ausschließlich `mtrace-webrtc` (`SMOKE_WEBRTC_AUTOSTART=0` für
  manuellen Aufruf gegen vorhandenen Stack); fremde
  Volumes/Container werden nicht angefasst.
- [x] Smoke prüft endpoint-/compose-only fünf Stufen: (1) MediaMTX-
  API antwortet `200`, (2) Stream-Pfad `webrtc-test` ist
  `ready=true`, (3) `OPTIONS …/whep → 204`, (4) `OPTIONS …/whip →
  204`, (5) Negativ-Probe für unbekannten Pfad antwortet **nicht**
  `204`. Kein Browser, kein Playback, kein `getStats()`.
- [x] Fehlerklassen sind durch eigenständige Meldungen
  unterscheidbar: „MediaMTX-API unreachable" (Compose nicht oben /
  Port-Konflikt 9999), „stream path … not ready" (FFmpeg-Publisher
  liefert keinen Frame), „WHEP/WHIP OPTIONS unexpected status"
  (falscher Listener-Status), „negative probe failed" (Pfad-Filter
  greift nicht). Lokal manuell verifiziert mit
  `SMOKE_WEBRTC_AUTOSTART=0` (Compose down) und mit `STREAM=…`
  auf nicht existierenden Pfad.
- [x] Bounded Waits über `WAIT_SECONDS` (Default 30 s, je 1 s
  Sleep); bei Fehlern werden `docker compose ps` und Tail-Logs
  (`mediamtx`, `webrtc-publisher`) als Diagnose-Hinweise auf
  stderr aufgeführt.
- [x] Root-`Makefile`: `smoke-webrtc-prep` in `.PHONY` aufgenommen,
  im `help`-Text mit Hinweis „endpoint-only" gelistet, eigenes
  Target ruft `bash scripts/smoke-webrtc-prep.sh` auf. Opt-in,
  nicht in `make gates`.
- [x] Optional / nicht release-blockierend: Headless-Browser-
  Erweiterung bleibt bewusst out-of-scope; die README dokumentiert
  diese Grenze in „Bekannte Grenzen". Eine spätere
  Browser-Smoke-Tranche kann sich additiv anhängen, ohne diesen
  Muss-Smoke zu kippen.

---

## 5. Tranche 4 — WebRTC-Telemetrie-Bewertung

Bezug: [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
§3.1/§3.2; Risiken-Backlog; Lastenheft §8.3 NF-14.

Ziel: Klare Telemetrie-Modell-Aussage, was eine spätere produktive
WebRTC-Telemetrie-Anbindung bräuchte — bounded Allowlist-Labels,
`getStats()`-Subset, Schema-Drift-Strategie. Diese Tranche kann **vor**
Tranche 1–3 erledigt werden (Spec-Vorarbeit) oder nach den
Lab-Erfahrungen. Sie erzeugt keinen Player-SDK-/Adapter-Vertrag.

DoD:

- [x] `spec/telemetry-model.md` §3.2 um drei WebRTC-Aggregat-Label-
  Zeilen erweitert: `connection_state` (W3C `RTCPeerConnectionState`-
  Enum), `ice_state` (W3C `RTCIceConnectionState`-Enum), `dtls_state`
  (W3C `RTCDtlsTransportState`-Enum); jeweils klar als „zukünftige
  WebRTC-Aggregate (siehe §3.5; **kein** produktiver Counter in
  `0.7.0`)" gekennzeichnet.
- [x] §3.1 um eine WebRTC-Forbidden-Zeile erweitert: `peer_connection_id`,
  Report-`id`, `track_id`, `transport_id`, `candidate_pair_id`,
  `local_candidate_id`, `remote_candidate_id`, `candidate_id`, `ssrc`,
  ICE-User-Fragmente, DTLS-/Zertifikats-Fingerprints, IP-Adressen,
  URLs, Codec-Strings, Browser-`user_agent`, generisches `source_id`-
  Label aus einem WebRTC-Adapter-Pfad. Verbot ist release-blockierend,
  sobald der erste produktive `mtrace_webrtc_*`-Counter eingeführt
  wird; bis dahin ist die Spiegelung in `scripts/smoke-observability.sh`
  Folge-DoD.
- [x] Neue §3.5 „WebRTC-Telemetrie-Vorbereitung (Future-Telemetry-Notiz)"
  in `spec/telemetry-model.md`: §3.5.2 listet das `getStats()`-Subset
  als Tabelle pro `RTCStatsType` (`peer-connection`, `transport`,
  `candidate-pair`, `inbound-rtp`/`outbound-rtp`) mit Muss-/Soll-Feldern;
  Per-Identifier-Felder werden explizit auf §3.1 verwiesen (nur Read-
  Pfad/Span, niemals Prometheus-Label). Notiz ist explizit kein
  Player-SDK-/Adapter-Public-API-Vertrag.
- [x] §3.5.3 dokumentiert die Schema-Drift-Strategie und das
  Fallback-Verhalten: Muss-Felder sind Pflichtbedingung (sonst Metrik
  leer statt `unknown`-Surrogat); Soll-Felder sind opt-in pro Engine
  (Histogram/Gauge weglassen, übrige Metriken bleiben aktiv); Schema-
  Drift ist Spec-/Adapter-Review-Gate, kein automatischer Release-
  Block; vor produktivem Pfad ist `smoke-webrtc-prep` vom Drift nicht
  betroffen.
- [x] `docs/planning/open/risks-backlog.md` §1.1 um **R-12** erweitert
  (nicht R-11 — der ist seit `0.6.0`-Closeout für SRT-Health-Cursor-
  Pagination vergeben, Plan-Stand „R-11 frei" war zur Plan-Erstellung
  korrekt, ist durch atomare Vergabe inzwischen überholt). R-12 ist
  als Spec-/Adapter-Review-Gate formuliert mit klarer Triggerschwelle
  (Browser-Major-Bump mit Schema-Änderung ODER Beginn produktiver
  WebRTC-Telemetrie-Implementierung). Stand-Datum im Header auf
  2026-05-06 nachgezogen.
- [x] §3.5.4 hält die RAK-49-Klausel wörtlich fest: keine produktive
  `getStats()`-Sammlung, keine `mtrace_webrtc_*`-Counter, kein Player-
  SDK-WebRTC-Adapter-Public-API-Vertrag in `0.7.0`. Diese Out-of-Scope-
  Klauseln sind redundant zu §0.3 dieses Plans und mit den Tranchen 1–3
  konsistent (Lab-Compose, Smoke endpoint-only, README-Folge-Pfad-
  Hinweis).
- [x] §3.5.1 dokumentiert die negative Cardinality-Prüfung explizit:
  im `0.7.0`-Release-Scope existiert kein produktiver `mtrace_webrtc_*`-
  Counter und kein WebRTC-Prometheus-Exportpfad; die Erweiterung von
  `scripts/smoke-observability.sh` auf WebRTC-Allowlist-Labels ist
  Folge-DoD für den ersten Plan, der eine produktive WebRTC-Metrik
  einführt.

---

## 6. Tranche 5 — Release-Doku, RAK-Matrix und Closeout

Bezug: RAK-47..RAK-50; `docs/user/local-development.md`;
`docs/user/releasing.md`; `README.md`.

Ziel: Die WebRTC-Lab-Erweiterung ist auffindbar, klar abgegrenzt und
mit den Release-Gates nachweisbar. RAK-51 bleibt sichtbar deferred.

DoD:

- [ ] `README.md` und `docs/user/local-development.md` verweisen auf
  `examples/webrtc/` mit Start-/Stop-/Smoke-Befehlen.
- [ ] `docs/user/releasing.md` nennt `make smoke-webrtc-prep` als
  zusätzlichen manuellen/opt-in Release-Smoke für `0.7.0`.
- [ ] `examples/README.md` listet `smoke-webrtc-prep` konsistent mit
  den anderen Example-Smokes.
- [ ] RAK-Verifikationsmatrix §6.1 ist vollständig ausgefüllt; bewusst
  verschobene Kann-Anforderungen stehen als `deferred / Folgeplan` statt
  als offene 0.7.0-Checkbox.
- [ ] `./scripts/verify-doc-refs.sh` ist grün.
- [ ] `plan-0.7.0.md` wird beim Release-Closeout nach
  `docs/planning/done/` verschoben und Roadmap §3 zeigt `0.7.0`
  released.

### 6.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-47 | Muss | `examples/webrtc/compose.yaml`, Project `mtrace-webrtc`, lokale WHIP-/WHEP-Endpunkte, HTTP- und ICE-/Media-Portkonflikte entschieden | [ ] |
| RAK-48 | Muss | `make smoke-webrtc-prep` endpoint-/compose-only, opt-in dokumentiert, kein Playback-/`getStats()`-Anspruch | [ ] |
| RAK-49 | Soll | `spec/telemetry-model.md` §3.2 mit WebRTC-`getStats()`-Future-Telemetry-Notiz und Schema-Drift-Strategie | [ ] |
| RAK-50 | Kann | Manueller Browser-Handcheck in `examples/webrtc/README.md` dokumentiert | [ ] |
| RAK-51 | Kann | Deferred; eigener Folgeplan für Player-SDK-WebRTC-Adapter | deferred / Folgeplan |

---

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen.
- Lastenheft-Patch ist erledigt: `spec/lastenheft.md` steht auf
  `1.1.9` mit §13.9-Sektion „Version 0.7.0" und RAK-47..RAK-51;
  Eintrag als §4a.12 in
  [`done/plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c.
  Plan kann nach `docs/planning/in-progress/` ziehen, sobald
  Vorgänger-Gates aus §0.1 grün sind.
- RAK-51 (Player-SDK-WebRTC-Adapter, „Kann") ist bewusst **nicht**
  Teil der Tranchen 1–4 — der Lastenheft-Anker bleibt für eine
  Folge-Tranche/-Release reserviert. Auslöser für eine Folge-
  Planung: Lab-Erfahrungen aus Tranche 1–3 zeigen einen konkreten
  Operator-Use-Case oder ein Folgeplan überführt die `getStats()`-
  Allowlist + Schema-Drift-Strategie aus `spec/telemetry-model.md` in
  einen produktiven Adapter-Pfad ohne Spec-Lücke.
  Das Folge-Dokument benennt dann eine eigene Tranche mit
  Public-API-Definition, hls.js-Pfad-Trennung und Compat-Tests.
- Wenn der WebRTC-Pfad nach Bewertung als „nie produktiv"
  entschieden wird, dieses Dokument als historische Notiz nach
  `docs/planning/done/` ziehen oder löschen — je nachdem, ob die
  Aufzeichnung weiter Wert hat.
