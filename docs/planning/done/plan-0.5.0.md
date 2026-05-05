# Implementation Plan — `0.5.0` (Multi-Protocol Lab)

> **Status**: ✅ released. Alle Tranchen 0–6 abgeschlossen. Tag
> `v0.5.0` auf `a56dc0b`; GitHub-Actions-Run 25364250989 in 4m37s grün;
> Release-Notes auf https://github.com/pt9912/m-trace/releases/tag/v0.5.0.
> Plan ist nach `docs/planning/done/` archiviert. Roadmap-Fokus auf
> `0.6.0` (SRT Health View, RAK-41..RAK-46) verschoben.
>
> **Bezug**: [Lastenheft `1.1.8`](../../../spec/lastenheft.md) §7.1
> (Repo-Struktur, `examples/`), §7.6 (Player-Adapter-Folgeoptionen), §7.8
> (lokales Streaming-Lab), §8.3 (Erweiterbarkeit), §10.5 (Infrastruktur),
> §13.7 (RAK-36..RAK-40); [Roadmap](../in-progress/roadmap.md) §3/§4;
> [Architektur](../../../spec/architecture.md) §8.2; [Local Development](../../user/local-development.md);
> [Stream Analyzer](../../user/stream-analyzer.md); [Risiken-Backlog](../open/risks-backlog.md).
>
> **Vorgänger-Gate (Stand zum `0.5.0`-Start)**:
>
> - [`plan-0.4.0.md`](../done/plan-0.4.0.md) ist vollständig (`[x]`)
>   und nach `docs/planning/done/` verschoben.
> - `v0.4.0` ist veröffentlicht; GitHub Actions `Build` ist für den
>   Release-Commit grün.
> - Roadmap §1/§2/§3 ist auf `0.5.0` als nächstes aktives Release
>   umgestellt.
> - Historische Statusquellen aus `0.4.0` enthalten keine widersprüchlichen
>   Restmarker mehr (z. B. `🟡 in Arbeit` in archivierten Abschlusszeilen
>   oder alte `0.4.0`-Scope-Cut-Texte in der Roadmap).
> - `0.4.0`-Folgerisiken, die Lab-Start oder Smoke-Gates beeinflussen,
>   sind im Risiken-Backlog bewertet.
>
> **Nachfolger**: `plan-0.6.0.md` (SRT Health View).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene Entscheidung.
- 🟡 in Arbeit.

Neue Lastenheft-Patches während `0.5.0` landen weiterhin zentral in
`plan-0.1.0.md` Tranche 0c, weil sie projektweit gelten.

Scope-Grenze: `0.5.0` liefert **Beispiele und verifizierbare lokale
Lab-Pfade**. Es liefert keine produktive Ingest-Verwaltung, keine
SRT-Health-Metriken und kein WebRTC-Monitoring. Diese Themen bleiben
Folge-Scope (`0.6.0` bzw. spätere Releases).

### 0.1 Vorab-Entscheidungen für den Scope-Cut

Diese Entscheidungen verhindern, dass `0.5.0` während der Umsetzung in
mehrere Richtungen wächst:

| Thema | Entscheidung für `0.5.0` | Nicht-Ziel |
| ----- | ------------------------ | ---------- |
| Compose-Form | Beispiele bevorzugen eigene `examples/<name>/compose.yaml`-Dateien. Root-Compose bleibt Core-Lab. | Root-Compose mit vielen optionalen Beispielprofilen überfrachten. |
| Smoke-Targets | Root-`Makefile` bekommt opt-in Targets `smoke-mediamtx`, `smoke-srt`, `smoke-dash`; WebRTC höchstens `smoke-webrtc-prep`, falls stabil. | Neue Streaming-Smokes in `make gates` erzwingen. |
| Doku-Form | Kurze README je Beispiel plus gebündelter User-Guide `docs/user/multi-protocol-lab.md`, falls die Inhalte über reine Verweise hinausgehen. | Lange Protokoll-Tutorials in `README.md`. |
| DASH | `0.5.0` liefert ein lokal erreichbares MPD-/DASH-Ausspielungsbeispiel. | Vollständige DASH-Manifestanalyse oder dash.js-Adapter als Release-Pflicht. |
| SRT | `0.5.0` liefert SRT-Publish-/Ausspielungs-Smoke. | SRT-Health-Metriken, CGO-Bindings oder API-Runtime-Integration. |
| WebRTC | Dokumentierter Vorbereitungspfad mit Ports/Grenzen; Smoke nur, wenn headless stabil genug ist. | WebRTC-Observability, Signaling-Service oder `getStats()`-Normalisierung. |

### 0.2 Risiko-Triage aus `0.4.0`

Diese Tabelle schneidet die offenen Folgepunkte aus
[`risks-backlog.md`](../open/risks-backlog.md) für den `0.5.0`-Scope. Sie
verhindert, dass Lab-Beispiele versehentlich Backend- oder Dashboard-
Schulden aus `0.4.0` miterledigen müssen.

| Risiko | Entscheidung für `0.5.0` | Trigger / Nachweis |
| ------ | ------------------------ | ------------------ |
| R-2 CGO/SRT-Bindings | Beobachten. `0.5.0` nutzt SRT nur über Lab-Container; keine CGO- oder API-Runtime-Entscheidung. | Wenn das SRT-Beispiel ein Binding in `apps/api` erzwingen würde, wird R-2 präzisiert und der Umfang auf `0.6.0` verschoben. |
| R-5 Time-Skew-Persistenz | Nicht Teil des Multi-Protocol-Lab. Deferred, solange Lab-Smokes keine skew-betroffenen Operator-Flows erzeugen. | Aktivieren nur bei Backlog-Trigger: ≥ 5 relevante `mtrace.time.skew_warning=true`-Spans außerhalb von Synthetik-Tests oder konkreter Operator-Report. |
| R-7 Session-List-N+1 | Nicht als `0.5.0`-Pflicht einplanen; bei Smoke-Arbeit nur beobachten. | Falls neue Lab-Smokes `GET /api/stream-sessions` p95 ≥ 200 ms reproduzierbar machen, wird ein Bulk-Read-Port als additive Tranche ergänzt. |
| R-10 Sampling-Vollständigkeit | Deferred. `0.5.0`-Lab-Smokes bleiben bei vollständigen Testdaten (`sampleRate = 1`). | Wieder öffnen, sobald ein Lab- oder Produktionspfad `sampleRate < 1` semantisch vollständig nachweisen muss. |

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Vorgänger-Gate und Scope-Festlegung | ✅ |
| 1 | Example-Struktur und Lab-Konventionen | ✅ |
| 2 | MediaMTX-Beispiel erweitern (RAK-36) | ✅ |
| 3 | SRT-Beispiel als Lab-Szenario (RAK-37) | ✅ |
| 4 | DASH-Beispiel und Analyzer-Grenze (RAK-38) | ✅ |
| 5 | WebRTC vorbereitet, nicht produktiv (RAK-39) | ✅ |
| 6 | Dokumentation, Smokes und Release-Gates (RAK-40) | ✅ |

---

## 1a. Tranche 0 — Vorgänger-Gate und Scope-Festlegung

Bezug: Roadmap §1/§3; Lastenheft §13.7; `plan-0.4.0.md` §9.5.

Ziel: Vor jeder Implementierung ist klar, dass `0.5.0` auf dem
veröffentlichten `0.4.0`-Stand aufsetzt und nur Lab-Beispiele, nicht
neue Betriebsmetriken, liefert.

DoD:

- [x] `plan-0.4.0.md` ist vollständig abgeschlossen, nach
  `docs/planning/done/` verschoben und mit finalen Commit-Hashes
  versehen (`da883af` plus Sub-Tranchen-Hashes in §9-Matrix).
- [x] Annotierter Release-Tag `v0.4.0` existiert und zeigt auf den
  finalen Release-Commit (`9e4fdb3`).
- [x] GitHub Actions `Build` ist für den `v0.4.0`-Commit grün
  (Run 25359933129 in 4m59s).
- [x] `docs/planning/in-progress/roadmap.md` markiert `0.5.0` als
  aktive Phase und verweist auf dieses Dokument: Status-Header auf
  „`0.5.0` (Multi-Protokoll-Lab, RAK-36..RAK-40) in Arbeit, Tranche 0";
  §1.2 als „Aktive `0.5.0`-Phase" mit Tranchen-Übersicht 0–6 und
  Verweis auf [`plan-0.5.0.md`](./plan-0.5.0.md).
- [x] Historische `0.4.0`-Statusquellen sind konsistent: archivierter
  Plan §9.6-Matrix-Zeile auf ✅ (`da883af`); README "Aktueller Stand"
  und v0.4.0-Roadmap-Block tragen "released" statt "in Arbeit"; Roadmap
  §1.1-Zeile `0.4.0` ✅, §3-Release-Übersicht ✅, Schritt 36 ✅; Audit
  per `grep -rn "0.4.0.*in Arbeit\|0.4.0-Scope-Cut"` zeigt keine
  Treffer außerhalb der Doku-Konvention-Legende.
- [x] Scope-Entscheidung ist dokumentiert: SRT in `0.5.0` bedeutet
  Beispiel/Smoke, **nicht** SRT-Health-View, SRT-Metrikimport oder
  CGO-Binding in `apps/api` (siehe §0.1 SRT-Zeile, Tranche 3 §4
  und §0.2 R-2-Zeile).
- [x] WebRTC in `0.5.0` ist als vorbereiteter Beispielplatz
  festgelegt; produktive `getStats()`-Erfassung bleibt out of scope
  (siehe §0.1 WebRTC-Zeile und Tranche 5 §6).
- [x] Bestehende `0.4.0`-Folgerisiken R-2, R-5, R-7 und R-10 sind gemäß
  §0.2 triagiert; falls Lab-Smokes einen Trigger auslösen, wird
  `risks-backlog.md` aktualisiert oder eine additive Tranche ergänzt.

---

## 2. Tranche 1 — Example-Struktur und Lab-Konventionen

Bezug: Lastenheft §7.1 F-5, Zielstruktur `examples/`; §7.8; §10.5.

Ziel: Beispiele liegen nicht verstreut in Compose-Kommentaren, sondern
haben eine stabile Struktur, kurze READMEs und wiederholbare Start- oder
Smoke-Pfade. Das bestehende Core-Lab bleibt weiterhin per `make dev`
unverändert startbar.

Vorgeschlagene Zielstruktur:

```text
examples/
├── mediamtx/
├── srt/
├── dash/
└── webrtc/
```

DoD:

- [x] `examples/` ist angelegt und enthält je Protokoll-Beispiel ein
  eigenes Unterverzeichnis mit `README.md` (`examples/{mediamtx,srt,
  dash,webrtc}/README.md` plus Konventions-Index `examples/README.md`).
- [x] Jede Beispiel-README folgt derselben Mindeststruktur: Zweck,
  Voraussetzungen, Start, Verifikation, Stop/Reset, Troubleshooting,
  bekannte Grenzen — verbindlich dokumentiert in `examples/README.md`
  Sektion „README-Mindeststruktur" und in allen vier Skelett-READMEs
  konsistent verwendet.
- [x] Jedes Beispiel benennt Zweck, Startbefehl, erwartete Ports/URLs,
  Abbruch-/Reset-Pfad und bekannte Grenzen — wo Inhalte erst durch
  Tranche 2–5 produktiv werden, ist das per `_Liefert Tranche X._`-
  Marker explizit benannt; Project-Name-Konvention `mtrace-<name>` ist
  in allen Skeletten gepinnt.
- [x] Compose-Erweiterungen liegen als eigene Override-Dateien unter
  `examples/<name>/compose.yaml` (Beschluss `plan-0.5.0.md` §0.1
  „Compose-Form"; Konvention dokumentiert in `examples/README.md`
  Sektion „Compose-Form").
- [x] Für eigene Example-Compose-Dateien ist der Projektname im
  Startbefehl dokumentiert (`docker compose -p mtrace-<name> -f
  examples/<name>/compose.yaml up …`); Konvention zentral in
  `examples/README.md` und in jeder Skelett-README am Start-Befehl
  belegt.
- [x] `make dev` bleibt das Core-Lab und startet keine zusätzlichen
  optionalen Beispiele — Tranche 1 fügt keine `make dev`-Diffs ein,
  und die Konvention in `examples/README.md` schreibt fest, dass
  Beispiele über eigene Compose-Projekte starten.
- [x] Neue Smoke-Targets sind opt-in (`make smoke-mediamtx`,
  `make smoke-srt`, `make smoke-dash`, ggf. `make smoke-webrtc-prep`)
  und hängen nicht an `make gates` — Konvention dokumentiert in
  `examples/README.md` Sektion „Smoke-Targets". Tranche 1 ergänzt
  noch keine Targets im Makefile; sie kommen pro Tranche 2–5 mit
  ihrem Compose-/Smoke-Inhalt.
- [x] Smoke-Skripte sind unter `scripts/` abgelegt, failen mit klaren
  Fehlermeldungen und räumen keine fremden Volumes/Container auf —
  Konvention dokumentiert in `examples/README.md` Sektion „Smoke-
  Skript-Konvention" (set -euo pipefail; `[smoke-<name>]`-Präfix;
  Compose-Down nur für `mtrace-<name>`-Project-Name).
- [x] `scripts/verify-doc-refs.sh` bleibt grün; neue Doku-Links werden
  vom bestehenden Docs-Gate erfasst — `make gates` (Stage `docs-check`
  ruft `verify-doc-refs.sh`) ist nach Tranche 1 grün.

---

## 3. Tranche 2 — MediaMTX-Beispiel erweitern (RAK-36)

Bezug: RAK-36; Lastenheft §7.8 F-82..F-84; Architektur §8.2;
`services/media-server/mediamtx.yml`.

Ziel: MediaMTX ist nicht nur impliziter Core-Service, sondern als
eigenes Beispiel nachvollziehbar: RTSP/RTMP-Ingest, HLS-Ausspielung,
Status/API-URL und Analyzer-/Dashboard-Bezug sind dokumentiert und
smoke-testbar.

DoD:

- [x] `examples/mediamtx/README.md` beschreibt den bestehenden
  MediaMTX-Pfad aus Root-Compose, inklusive `mediamtx`, `stream-generator`,
  HLS-URL und API/Status-Port — als Core-Lab-Beispiel-Variante (siehe
  `examples/README.md` Sektion „Compose-Form"); kein eigener Stack,
  Start ist `make dev`.
- [x] `make smoke-mediamtx` ist als dünner Wrapper dokumentiert und nutzt
  ein Skript (`scripts/smoke-mediamtx.sh`), das den bestehenden
  Core-Lab-Pfad validiert; Makefile-Target ist opt-in (nicht in
  `make gates`).
- [x] MediaMTX-Konfiguration ist so dokumentiert, dass die aktiven
  Protokolle und Ports (RTSP `8554`, HLS `8888`, API/Status `9997`)
  aus einem frischen Clone nachvollziehbar sind — siehe
  `examples/mediamtx/README.md` Sektionen „Voraussetzungen" und
  „Verifikation" plus Verweis auf `services/media-server/mediamtx.yml`.
- [x] Falls ein separates MediaMTX-Override nötig ist, kollidiert es
  nicht mit dem Core-Lab und nutzt eigene Pfade/Stream-Namen — Beispiel
  nutzt das Core-Lab direkt; Konvention für eigenständige Overrides
  (Project-Name `mtrace-<name>`) bleibt in `examples/README.md`
  dokumentiert für die folgenden Tranchen 3–5.
- [x] Ein Smoke-Pfad prüft mindestens: MediaMTX erreichbar,
  Teststream published, HLS-Manifest unter der dokumentierten URL
  erreichbar — `scripts/smoke-mediamtx.sh` prüft funktional über HLS:
  (1) Manifest unter `http://localhost:8888/teststream/index.m3u8`
  liefert 200, (2) Body beginnt mit `#EXTM3U`, (3) Body enthält
  Media-Referenzen (`.m3u8` Substreams oder `.ts`/`.m4s`/`.aac`-Segmente).
  HLS-200 + sinnvolles Manifest impliziert, dass MediaMTX läuft und
  der teststream erfolgreich publishing ist.
- [x] Der Smoke nutzt bounded Waits mit Diagnoseausgabe aus MediaMTX
  API/Status oder Container-Logs; ein noch startender Teststream führt
  nicht zu einem flakigen Sofort-Fehler — Default 30s `WAIT_SECONDS`-
  Polling auf HLS-200; bei Timeout schreibt der Smoke konkrete Hints
  (`docker compose logs stream-generator|mediamtx | tail -20`) auf
  stderr.
- [x] Dashboard-Demo und `POST /api/analyze` können die dokumentierte
  HLS-URL weiterhin nutzen; private-Netzwerk-Analyzer-Flag bleibt auf
  das lokale Lab beschränkt — Tranche 2 ändert nichts am bestehenden
  Compose-Wiring (`PUBLIC_HLS_URL` in `docker-compose.yml`,
  `ANALYZER_ALLOW_PRIVATE_NETWORKS=true` für Lab-URLs); Beispiel-README
  pinnt das in „Bekannte Grenzen".
- [x] `docs/user/local-development.md` verweist auf das neue
  MediaMTX-Beispiel, ohne die Core-Schnellstart-Anleitung aufzublähen
  — neue §2.7 mit knappem Verweisblock auf `examples/mediamtx/README.md`,
  plus Hinweis auf die SRT-/DASH-/WebRTC-Skelette für Folge-Tranchen.

Beobachtung MediaMTX 1.14+: Control-API auf `:9997` ist standardmäßig
Auth-pflichtig (HTTP 401 ohne Credentials). Der Smoke prüft daher
absichtlich nicht den Config-API-Endpoint, sondern den HLS-Pfad —
das ist die funktional aussagekräftige Zielmetrik. README dokumentiert
das in „Troubleshooting" als bekannten Lab-Default; produktive API-
Auth ist Folge-Scope (siehe `services/media-server/mediamtx.yml`
`authInternalUsers`-Hinweis).

---

## 4. Tranche 3 — SRT-Beispiel als Lab-Szenario (RAK-37)

Bezug: RAK-37; Lastenheft §4.3, §12.3 MVP-38; Roadmap §4
SRT-Binding-Stack; Risiken-Backlog R-2.

Ziel: Es gibt ein lokal startbares SRT-Beispiel, das Contribution-
Ingest als Szenario zeigt. `0.5.0` erfasst dabei noch keine SRT-Health-
Metriken und importiert keine SRT-Bindings in die Go-API.

DoD:

- [x] `examples/srt/README.md` beschreibt SRT-Sender (FFmpeg-Loop-
  Container `srt-publisher`), Media-Server-Ziel (MediaMTX-Container
  `mediamtx`), Ausspielungs-URL (`http://localhost:8889/srt-test/index.m3u8`)
  und erwarteten Erfolg (`make smoke-srt` grün nach 10–25 s).
- [x] Das Beispiel nutzt bevorzugt vorhandene Container-Images:
  `bluenviron/mediamtx:1` (gleiche Major-Version wie das Core-Lab in
  `docker-compose.yml`) und `jrottenberg/ffmpeg:8.1-ubuntu2404`
  (gleiche Image-Linie wie `services/stream-generator`). Keine neuen
  API-Runtime-Abhängigkeiten — `apps/api` bleibt
  `distroless-static` ohne CGO.
- [x] Das Beispiel pinnt explizit SRT-Port (`8890/udp`), Stream-Name
  (`srt-test`, absichtlich anders als `teststream` aus dem Core-Lab,
  damit Pfade nicht kollidieren) und Container-Richtung (Publisher
  → MediaMTX → HLS) statt sich auf implizite MediaMTX-Defaults zu
  verlassen — siehe `examples/srt/mediamtx.yml` und
  `examples/srt/compose.yaml`.
- [x] Ein Startpfad ist dokumentiert: `docker compose -p mtrace-srt
  -f examples/srt/compose.yaml up -d --build` in
  `examples/srt/README.md` Sektion „Start"; gleiche Befehlszeile
  liegt im Smoke-Skript für die Auto-Start-Variante.
- [x] `make smoke-srt` startet nur die für das SRT-Beispiel nötigen
  Dienste (Project `mtrace-srt`, eigenes Compose) und beendet sich
  deterministisch mit Erfolg oder Diagnose — Stack-Up + Smoke + Stack-
  Down per `trap cleanup EXIT`; bei Timeout konkrete Diagnose-Hints
  (`docker compose logs srt-publisher|mediamtx`) auf stderr.
- [x] Der Smoke-Pfad prüft, dass ein SRT-Publisher verbinden kann und
  daraus eine abspiel- oder analysierbare Ausspielung entsteht — wenn
  das HLS-Manifest auf `:8889/srt-test/index.m3u8` mit `200`
  antwortet **und** Media-Referenzen enthält, ist sowohl der SRT-
  Ingress als auch der HLS-Egress nachweislich funktional.
- [x] Der Smoke validiert nicht nur offene Ports, sondern ruft das
  erzeugte Manifest tatsächlich ab und prüft den Body
  (`#EXTM3U`-Header plus Media-Referenzen `.m3u8`/`.ts`/`.m4s`/
  `.aac`).
- [x] Keine SRT-Verbindungsmetriken werden als erfüllt behauptet;
  RAK-41..RAK-46 bleiben explizit `0.6.0`-Scope —
  `examples/srt/README.md` „Bekannte Grenzen" pinnt das.
- [x] R-2 bleibt unverändert: das Beispiel triggert keine neue CGO-/
  Runtime-Entscheidung für `0.6.0`, weil SRT ausschließlich im
  MediaMTX-Container läuft — keine `apps/api`-Code-Änderung nötig.
  Die R-2-Triage in §0.2 bleibt damit gültig.

Live-verifiziert:
- `make smoke-srt` → "[smoke-srt] all checks passed" mit Stack-Up
  in Project `mtrace-srt`, HLS-200 nach < 45 s, sauberer Compose-Down
  am Ende.

---

## 5. Tranche 4 — DASH-Beispiel und Analyzer-Grenze (RAK-38)

Bezug: RAK-38; Lastenheft §7.6 F-58, §7.7 F-73, §12.3 MVP-37;
`packages/stream-analyzer` DASH/CMAF-Erweiterungsgrenze.

Ziel: Ein DASH-Beispiel ist lokal nachvollziehbar. Dabei wird klar
getrennt zwischen **DASH-Ausspielungsbeispiel** (`0.5.0`-Pflicht) und
vollständiger DASH-Manifestanalyse (`MVP-37`, Kann-Folge-Scope).

DoD:

- [x] `examples/dash/README.md` beschreibt einen lokalen DASH-Stream
  inklusive MPD-URL (`http://localhost:8891/manifest.mpd`), Startpfad
  (`docker compose -p mtrace-dash -f examples/dash/compose.yaml up -d --build`)
  und erwarteter Player-Nutzung (Shaka-Player- bzw. dash.js-Reference-
  Player-Verlinkung als manueller Test, plus `curl`-Schnellprüfung).
- [x] Das Beispiel erzeugt oder liefert deterministische DASH-Artefakte
  ohne externe Netzwerkabhängigkeit: FFmpeg-`testsrc2`-Generator
  schreibt einen synthetischen Stream als CMAF/fMP4-Segmentkette in
  ein shared Volume; nginx servisiert das Volume statisch — kein CDN,
  kein Internet-Zugriff im laufenden Stack.
- [x] Bevorzugter Pfad ist ein kleiner lokaler DASH-Origin im
  Example-Compose: zwei Container im Compose
  (`dash-generator` mit `jrottenberg/ffmpeg:8.1-ubuntu2404` plus
  `dash-server` mit `nginx:1-alpine`). MPD und Segmente liegen im
  shared `dash-output`-Volume; `examples/dash/nginx.conf` konfiguriert
  CORS und MIME-Types für DASH (`application/dash+xml`,
  `video/iso.segment`).
- [x] Wenn `dash.js` im Dashboard-Demo-Pfad verwendet wird, bleibt der
  bestehende HLS-Demo-Pfad unverändert — `0.5.0` liefert **keinen**
  DASH-Demo-Player im `apps/dashboard`-Code; der existierende
  `hls.js`-Pfad in `apps/dashboard/src/routes/demo/` ist unverändert.
  README pinnt das in „Bekannte Grenzen".
- [x] `packages/player-sdk` bleibt abwärtskompatibel: `0.5.0` liefert
  **keinen** dash.js-Adapter; `attachHlsJs`, `createTracker`,
  Public-API und Wire-Format bleiben unverändert. Kein neuer Test,
  kein neuer Public-API-Snapshot-Eintrag.
- [x] `POST /api/analyze` bleibt in `0.5.0` klar HLS-only — keine
  zusätzliche DASH-Analyse-Route, keine Schemaerweiterung in
  `contracts/sdk-compat.json` oder `contracts/event-schema.json`.
  README dokumentiert die Grenze: wer aus dem Beispiel heraus eine
  MPD an `POST /api/analyze` schickt, bekommt einen `not_hls`-Fehler
  — erwartet in `0.5.0`.
- [x] `make smoke-dash` prüft, dass die dokumentierte MPD erreichbar
  ist (HTTP 200 + `<MPD`-Body), mindestens ein referenziertes Segment
  lokal abrufbar ist (`init-stream0.m4s` HEAD-200) und der
  Beispielpfad ohne Internet läuft (FFmpeg-Generator nutzt nur
  lavfi-Quellen, nginx serviert nur das lokale Volume).

Live-verifiziert:
- `make smoke-dash` → "[smoke-dash] all checks passed" mit Stack-Up
  in Project `mtrace-dash`, MPD-200 nach < 45 s, Init-Segment
  HEAD-200, sauberer Compose-Down inkl. `--volumes` am Ende.

---

## 6. Tranche 5 — WebRTC vorbereitet, nicht produktiv (RAK-39)

Bezug: RAK-39; Lastenheft §7.6 F-62, §12.1 MVP-24, §8.3 NF-14.

Ziel: WebRTC ist als späteres Lab-Szenario vorbereitet, ohne das MVP mit
WebRTC-Monitoring, Signaling-Server oder `getStats()`-Normalisierung zu
belasten.

DoD:

- [x] `examples/webrtc/README.md` existiert und beschreibt den
  geplanten lokalen WebRTC-Pfad (MediaMTX-WHIP-/WHEP),
  Mindestdienste (MediaMTX, optional `coturn`-STUN/TURN), offene
  Entscheidungen (Port-Schnitt gegen `examples/srt/`) und Nicht-
  Ziele (kein Signaling-Server, keine `getStats()`-Erfassung,
  keine Dashboard-Sichtbarkeit).
- [x] MediaMTX-WebRTC: notwendige Ports (geplant `8889/tcp` für
  WHIP/WHEP, kollidiert in `0.5.0` mit dem SRT-Beispiel — Folge-
  Tranche muss Port-Mapping pro Beispiel neu schneiden), NAT-/ICE-
  Hinweise (STUN/TURN nur nötig, falls Lab-Pfad nicht-localhost
  abdecken soll) und Browser-Grenzen (Chromium/Firefox supported,
  Safari documented limitation) sind in `examples/webrtc/README.md`
  dokumentiert.
- [x] Klare Entscheidungsmarke: `0.5.0` liefert **ausschließlich
  Port-/Konfigurationsvorbereitung als Doku**, **keinen** Browser-
  Handcheck und **kein** `compose.yaml` — README-Header und „Start"-
  Sektion pinnen das explizit.
- [x] Keine produktive Behauptung für WebRTC-Metriken in Dashboard,
  API oder Telemetry-Model: `examples/webrtc/README.md` „Bekannte
  Grenzen" zählt explizit auf, dass kein Signaling-Server in
  `apps/api`, keine `getStats()`-Sammlung im `@npm9912/player-sdk`,
  keine WebRTC-Aggregat-Metriken in Prometheus und kein Dashboard-
  Hook geliefert werden.
- [x] Optionaler Smoke ist bewusst als Folge-Item dokumentiert,
  **nicht** in `0.5.0` implementiert — Begründung in der
  README-Sektion „Verifikation": headless-WebRTC ist instabil in
  CI, `getStats()` browser-spezifisch, Lab-Wert ohne Smoke begrenzt.
- [x] Falls ein Smoke implementiert wird, ist er als
  `make smoke-webrtc-prep` gekennzeichnet — Target-Name ist
  reserviert in [`examples/README.md`](../../../examples/README.md)
  Sektion „Smoke-Targets" mit Status `⬜ — nur falls headless stabil`,
  Prüfumfang explizit auf Vorbereitungsgrenze begrenzt (kein
  Playback-Qualitäts- oder `getStats()`-Anspruch).
- [x] Roadmap/Folge-Plan nennt, was aus der Vorbereitung später
  echte WebRTC-Metrikunterstützung macht — `examples/webrtc/README.md`
  „Bekannte Grenzen" benennt die vier strukturellen Out-of-Scope-
  Klauseln (kein Signaling, kein `getStats()`, keine Aggregat-
  Metriken, kein Dashboard-Hook) und verweist auf
  [`docs/planning/open/plan-0.7.0.md`](../open/plan-0.7.0.md) als
  Folge-Plan. Detaillierte Folge-Schritte (Lab-Compose, README-
  Konkretisierung, `smoke-webrtc-prep`-Target, WebRTC-Telemetrie-
  Bewertung) gehören in `plan-0.7.0.md`, nicht in die Beispiel-
  README — Konsequenz aus T6c-Review. `plan-0.7.0.md` §0.2 schlägt
  zusätzlich die Lastenheft-Erweiterung um RAK-47..RAK-N als
  Pflicht-Vorarbeit vor.

---

## 7. Tranche 6 — Dokumentation, Smokes und Release-Gates (RAK-40)

Bezug: RAK-40; `docs/user/local-development.md`; `README.md`;
`docs/user/releasing.md`; bestehende Gates im `Makefile`.

Ziel: Alle Beispiele sind für Nutzer auffindbar, startbar und mit
klaren Grenzen dokumentiert. Release-Akzeptanz prüft RAK-36..RAK-40
explizit, ohne die normalen Qualitäts-Gates zu schwächen.

DoD:

- [x] `README.md` Release-Übersicht und Schnellstart verweisen auf die
  Multi-Protocol-Beispiele — neuer „Lieferstand und Roadmap"-Block
  (ersetzt den alten 100-Zeilen-`MVP-Scope`-Block durch Verweise auf
  CHANGELOG, Roadmap und „Was m-trace nicht ist"); v0.5.0-Roadmap-
  Sektion mit konkreten Bullets; „Aktueller Stand"-Block referenziert
  `examples/`.
- [x] `docs/user/local-development.md` enthält eine knappe
  Multi-Protocol-Lab-Sektion mit Start-/Stop-/Reset-Hinweisen — §2.7
  hat Quickref-Tabelle (alle vier Beispiele mit Variante/Start/Smoke),
  parallel-Stack-Port-Schnitt-Tabelle und Reset-Anleitung pro Project-
  Name.
- [x] Optionales neues `docs/user/multi-protocol-lab.md` ist **nicht**
  nötig — `local-development.md` §2.7 ist mit Quickref-Tabellen und
  Cross-Verweisen auf die jeweiligen `examples/<name>/README.md`
  ausreichend kompakt; der User-Guide bleibt zentral in
  `local-development.md`.
- [x] Beispiel-READMEs verweisen auf den gebündelten User-Guide
  (`docs/user/local-development.md` §2.7) im Header-Bezugs-Block —
  alle vier `examples/<name>/README.md` haben „Quickref aller Multi-
  Protocol-Lab-Beispiele:"-Zeile; `examples/README.md` umgekehrt mit
  Verweis auf §2.7 am Tabellen-Fuß.
- [x] `Makefile` listet `smoke-mediamtx`, `smoke-srt` und `smoke-dash`
  im `help`-Text (Tranchen 2/3/4 ergänzt); `smoke-webrtc-prep` ist
  bewusst nicht gelistet, weil in `0.5.0` nicht implementiert
  (Begründung in `examples/webrtc/README.md` „Verifikation").
- [x] `docs/user/releasing.md` nennt die drei `0.5.0`-Release-Smokes
  als opt-in-Block in §2 Verifikation; CI-Beschreibung dokumentiert,
  dass `smoke-mediamtx`/`smoke-srt`/`smoke-dash` lokal vor dem Tag
  laufen (analog `smoke-analyzer`/`smoke-observability`/`browser-e2e`).
- [x] RAK-Verifikationsmatrix ist im Plan ergänzt (§7.1); alle fünf
  RAKs (RAK-36..RAK-40) sind `[x]` mit konkretem Nachweis-Pfad zu
  `examples/`-Verzeichnis und Smoke-Target.
- [x] `make gates` ist grün — letzte Verifikation am Ende von Tranche 6
  in §7.2 mit Datum gepinnt.
- [x] Pflicht-Smokes für MediaMTX, SRT und DASH sind lokal grün und mit
  Befehl/Datum in §7.2 dokumentiert; WebRTC bleibt explizit kein Smoke-
  Nachweis, weil `smoke-webrtc-prep` in `0.5.0` nicht implementiert ist
  (Doku-only Vorbereitungspfad gemäß §6).
- [x] Versionen, `CHANGELOG.md`, Planstatus, Roadmap und Release-Notes
  sind für `0.5.0` aktualisiert (T6b `f25e40e`, T6c CHANGELOG-Datierung
  `a56dc0b`). Alle 21+ versionsführenden Stellen (5 `package.json`,
  `apps/api/cmd/api/main.go` `serviceVersion`, `PLAYER_SDK_VERSION`,
  `contracts/sdk-compat.json`, `pack-smoke.mjs` `expectedVersion`,
  `pack:smoke`-Tarballname, `stream-analyzer/.../parsers/hls.ts`-
  Fehlertext, alle Test-/Contract-Fixtures, 3 Doku-Stellen in
  `apps/analyzer-service/README.md` und `packages/player-sdk/README.md`)
  auf `0.5.0`. Schema-Versionen unverändert (`wire_schema_version=1.0`).
  CHANGELOG-Block `[0.5.0] - 2026-05-05`. GitHub-Release auf
  https://github.com/pt9912/m-trace/releases/tag/v0.5.0 mit Auszug
  als Notes.
- [x] `plan-0.5.0.md` ist nach Abschluss nach `docs/planning/done/`
  verschoben; Roadmap verweist danach auf den finalen Pfad (T6c).
  `git mv` zum `done/`-Pfad; interner Roadmap-Verweis im Bezugs-
  Block auf `../in-progress/roadmap.md` korrigiert (war vor `git mv`
  noch `./roadmap.md`); Cross-Verweise in
  `README.md`, `examples/{README,mediamtx,srt,dash,webrtc}/README.md`
  und `docs/planning/open/plan-0.6.0.md` auf `done/plan-0.5.0.md`
  nachgezogen.

### 7.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-36 | Muss | [`examples/mediamtx/`](../../../examples/mediamtx/) als Core-Lab-Beispiel; `make smoke-mediamtx` prüft HLS-Erreichbarkeit, `#EXTM3U`-Body und Media-Referenzen; Doku-Verweis aus [`docs/user/local-development.md`](../../user/local-development.md) §2.7 | [x] |
| RAK-37 | Muss | [`examples/srt/`](../../../examples/srt/) eigenes Compose `mtrace-srt`; `make smoke-srt` prüft Publisher-Verbindung end-to-end (FFmpeg → SRT → MediaMTX → HLS) und HLS-Manifest-Body | [x] |
| RAK-38 | Muss | [`examples/dash/`](../../../examples/dash/) eigenes Compose `mtrace-dash`; `make smoke-dash` prüft MPD lokal erreichbar (HTTP 200, `<MPD`-Body), Init-Segment HEAD-200; Pfad nutzt nur lokale `lavfi`-Quellen, kein Internet | [x] |
| RAK-39 | Soll | [`examples/webrtc/README.md`](../../../examples/webrtc/README.md) als Doku-only Vorbereitungspfad; bewusst dokumentiertes Folge-Item (kein Smoke in `0.5.0`); `smoke-webrtc-prep`-Target-Name reserviert für spätere Tranche | [x] |
| RAK-40 | Muss | [`README.md`](../../../README.md) v0.5.0-Block + „Lieferstand und Roadmap"-Sektion; [`docs/user/local-development.md`](../../user/local-development.md) §2.7 mit Quickref-Tabelle; [`docs/user/releasing.md`](../../user/releasing.md) listet die drei `0.5.0`-Smokes; Beispiel-READMEs verweisen zurück auf §2.7; Docs-Gate (`make docs-check`) grün | [x] |

### 7.2 Release-Closeout-Protokoll

Dieser Abschnitt wird während Tranche 6 gefüllt, damit Smoke-Nachweise
nicht ad hoc in Commit-Bodies oder Release-Notes verschwinden.

| Prüfung | Befehl / Nachweis | Datum | Ergebnis | Notiz |
| ------- | ----------------- | ----- | -------- | ----- |
| `make gates` | Coverage 97.61%/91.25%/100%/100%, arch-check, schema-validate, docs-check (mehrfach grün im Tranche-1–6-Verlauf) | 2026-05-05 | [x] | letzte Verifikation am Ende §7-Closeout |
| MediaMTX-Smoke | `make smoke-mediamtx` (Core-Lab via `make dev` vorausgesetzt) | 2026-05-05 | [x] | RAK-36; HLS 200, `#EXTM3U`-Body, Media-Referenzen |
| SRT-Smoke | `make smoke-srt` (Auto-Start `mtrace-srt`) | 2026-05-05 | [x] | RAK-37; HLS auf 8889 nach FFmpeg→SRT→MediaMTX |
| DASH-Smoke | `make smoke-dash` (Auto-Start `mtrace-dash`) | 2026-05-05 | [x] | RAK-38; MPD + Init-Segment HEAD auf 8891 |
| WebRTC-Vorbereitung | Doku-Review `examples/webrtc/README.md` | 2026-05-05 | [x] | RAK-39; kein Smoke in `0.5.0` (Begründung in README + Plan §6) |
| Docs-Gate | `make docs-check` (Teil von `make gates`) | 2026-05-05 | [x] | RAK-40; alle Cross-Verweise valide |
