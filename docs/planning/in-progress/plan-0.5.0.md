# Implementation Plan — `0.5.0` (Multi-Protocol Lab)

> **Status**: 🟡 Scope-Cut. `0.4.0` ist released (Tag `v0.4.0` auf
> `9e4fdb3`, CI-Run 25359933129 grün, Plan in
> `docs/planning/done/plan-0.4.0.md`). `plan-0.5.0.md` ist nach
> `docs/planning/in-progress/` verschoben und tritt als aktive Phase in
> Tranche 0 ein.
>
> **Bezug**: [Lastenheft `1.1.8`](../../../spec/lastenheft.md) §7.1
> (Repo-Struktur, `examples/`), §7.6 (Player-Adapter-Folgeoptionen), §7.8
> (lokales Streaming-Lab), §8.3 (Erweiterbarkeit), §10.5 (Infrastruktur),
> §13.7 (RAK-36..RAK-40); [Roadmap](./roadmap.md) §3/§4;
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
| 0 | Vorgänger-Gate und Scope-Festlegung | ⬜ |
| 1 | Example-Struktur und Lab-Konventionen | ⬜ |
| 2 | MediaMTX-Beispiel erweitern (RAK-36) | ⬜ |
| 3 | SRT-Beispiel als Lab-Szenario (RAK-37) | ⬜ |
| 4 | DASH-Beispiel und Analyzer-Grenze (RAK-38) | ⬜ |
| 5 | WebRTC vorbereitet, nicht produktiv (RAK-39) | ⬜ |
| 6 | Dokumentation, Smokes und Release-Gates (RAK-40) | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate und Scope-Festlegung

Bezug: Roadmap §1/§3; Lastenheft §13.7; `plan-0.4.0.md` §9.5.

Ziel: Vor jeder Implementierung ist klar, dass `0.5.0` auf dem
veröffentlichten `0.4.0`-Stand aufsetzt und nur Lab-Beispiele, nicht
neue Betriebsmetriken, liefert.

DoD:

- [ ] `plan-0.4.0.md` ist vollständig abgeschlossen, nach
  `docs/planning/done/` verschoben und mit finalen Commit-Hashes
  versehen.
- [ ] Annotierter Release-Tag `v0.4.0` existiert und zeigt auf den
  finalen Release-Commit.
- [ ] GitHub Actions `Build` ist für den `v0.4.0`-Commit grün.
- [ ] `docs/planning/in-progress/roadmap.md` markiert `0.5.0` als
  aktive Phase und verweist auf dieses Dokument.
- [ ] Historische `0.4.0`-Statusquellen sind konsistent: archivierter
  Plan, Roadmap und README enthalten keine widersprüchlichen Restmarker
  oder alten Scope-Cut-Texte mehr.
- [ ] Scope-Entscheidung ist dokumentiert: SRT in `0.5.0` bedeutet
  Beispiel/Smoke, **nicht** SRT-Health-View, SRT-Metrikimport oder
  CGO-Binding in `apps/api`.
- [ ] WebRTC in `0.5.0` ist als vorbereiteter Beispielplatz
  festgelegt; produktive `getStats()`-Erfassung bleibt out of scope.
- [ ] Bestehende `0.4.0`-Folgerisiken R-2, R-5, R-7 und R-10 sind gemäß
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

- [ ] `examples/` ist angelegt und enthält je Protokoll-Beispiel ein
  eigenes Unterverzeichnis mit `README.md`.
- [ ] Jede Beispiel-README folgt derselben Mindeststruktur: Zweck,
  Voraussetzungen, Start, Verifikation, Stop/Reset, Troubleshooting,
  bekannte Grenzen.
- [ ] Jedes Beispiel benennt Zweck, Startbefehl, erwartete Ports/URLs,
  Abbruch-/Reset-Pfad und bekannte Grenzen.
- [ ] Compose-Erweiterungen liegen entweder als klar benannte Override-
  Dateien unter `examples/<name>/compose.yaml` oder als Profile im
  Root-Compose; die Entscheidung ist einheitlich dokumentiert.
- [ ] Für eigene Example-Compose-Dateien ist der Projektname im
  Startbefehl dokumentiert, damit Volumes/Container nicht versehentlich
  mit dem Core-Lab kollidieren.
- [ ] `make dev` bleibt das Core-Lab und startet keine zusätzlichen
  optionalen Beispiele.
- [ ] Neue Smoke-Targets sind opt-in (`make smoke-mediamtx`,
  `make smoke-srt`, `make smoke-dash`) und hängen nicht an `make gates`,
  solange sie zusätzliche Streaming-Images oder lange Medien-Starts
  benötigen.
- [ ] Smoke-Skripte sind unter `scripts/` abgelegt, failen mit klaren
  Fehlermeldungen und räumen keine fremden Volumes/Container auf.
- [ ] `scripts/verify-doc-refs.sh` bleibt grün; neue Doku-Links werden
  vom bestehenden Docs-Gate erfasst.

---

## 3. Tranche 2 — MediaMTX-Beispiel erweitern (RAK-36)

Bezug: RAK-36; Lastenheft §7.8 F-82..F-84; Architektur §8.2;
`services/media-server/mediamtx.yml`.

Ziel: MediaMTX ist nicht nur impliziter Core-Service, sondern als
eigenes Beispiel nachvollziehbar: RTSP/RTMP-Ingest, HLS-Ausspielung,
Status/API-URL und Analyzer-/Dashboard-Bezug sind dokumentiert und
smoke-testbar.

DoD:

- [ ] `examples/mediamtx/README.md` beschreibt den bestehenden
  MediaMTX-Pfad aus Root-Compose, inklusive `mediamtx`, `stream-generator`,
  HLS-URL und API/Status-Port.
- [ ] `make smoke-mediamtx` ist als dünner Wrapper dokumentiert und nutzt
  ein Skript, das den bestehenden Core-Lab-Pfad validiert.
- [ ] MediaMTX-Konfiguration ist so dokumentiert, dass die aktiven
  Protokolle und Ports (`RTSP`, `HLS`, API/Status; optional `RTMP`, falls
  im Beispiel aktiviert) aus einem frischen Clone nachvollziehbar sind.
- [ ] Falls ein separates MediaMTX-Override nötig ist, kollidiert es
  nicht mit dem Core-Lab und nutzt eigene Pfade/Stream-Namen.
- [ ] Ein Smoke-Pfad prüft mindestens: MediaMTX erreichbar,
  Teststream published, HLS-Manifest unter der dokumentierten URL
  erreichbar.
- [ ] Der Smoke nutzt bounded Waits mit Diagnoseausgabe aus MediaMTX
  API/Status oder Container-Logs; ein noch startender Teststream führt
  nicht zu einem flakigen Sofort-Fehler.
- [ ] Dashboard-Demo und `POST /api/analyze` können die dokumentierte
  HLS-URL weiterhin nutzen; private-Netzwerk-Analyzer-Flag bleibt auf das
  lokale Lab beschränkt.
- [ ] `docs/user/local-development.md` verweist auf das neue
  MediaMTX-Beispiel, ohne die Core-Schnellstart-Anleitung aufzublähen.

---

## 4. Tranche 3 — SRT-Beispiel als Lab-Szenario (RAK-37)

Bezug: RAK-37; Lastenheft §4.3, §12.3 MVP-38; Roadmap §4
SRT-Binding-Stack; Risiken-Backlog R-2.

Ziel: Es gibt ein lokal startbares SRT-Beispiel, das Contribution-
Ingest als Szenario zeigt. `0.5.0` erfasst dabei noch keine SRT-Health-
Metriken und importiert keine SRT-Bindings in die Go-API.

DoD:

- [ ] `examples/srt/README.md` beschreibt SRT-Sender, Media-Server-
  Ziel, Ausspielungs-URL und erwarteten Erfolg.
- [ ] Das Beispiel nutzt bevorzugt vorhandene Container-Images
  (MediaMTX + FFmpeg) und vermeidet neue API-Runtime-Abhängigkeiten.
- [ ] Das Beispiel pinnt explizit SRT-Port, Stream-Name und Container-
  Richtung (Publisher → MediaMTX → HLS/anderer Ausspielungspfad), statt
  sich auf implizite MediaMTX-Defaults zu verlassen.
- [ ] Ein Startpfad ist dokumentiert, z. B. ein Compose-Profil oder
  `docker compose -f examples/srt/compose.yaml up --build`.
- [ ] `make smoke-srt` startet nur die für das SRT-Beispiel nötigen
  Dienste und beendet sich deterministisch mit Erfolg oder Diagnose.
- [ ] Der Smoke-Pfad prüft mindestens, dass ein SRT-Publisher verbinden
  kann und daraus eine abspiel- oder analysierbare Ausspielung entsteht.
- [ ] Der Smoke validiert nicht nur offene Ports, sondern ruft das
  erzeugte Manifest oder eine vergleichbare Media-Ausspielung tatsächlich
  ab.
- [ ] Keine SRT-Verbindungsmetriken werden als erfüllt behauptet; RAK-41
  bis RAK-46 bleiben explizit `0.6.0`-Scope.
- [ ] R-2 bleibt unverändert oder wird präzisiert, falls das Beispiel
  eine neue CGO-/Runtime-Entscheidung für `0.6.0` sichtbar macht.

---

## 5. Tranche 4 — DASH-Beispiel und Analyzer-Grenze (RAK-38)

Bezug: RAK-38; Lastenheft §7.6 F-58, §7.7 F-73, §12.3 MVP-37;
`packages/stream-analyzer` DASH/CMAF-Erweiterungsgrenze.

Ziel: Ein DASH-Beispiel ist lokal nachvollziehbar. Dabei wird klar
getrennt zwischen **DASH-Ausspielungsbeispiel** (`0.5.0`-Pflicht) und
vollständiger DASH-Manifestanalyse (`MVP-37`, Kann-Folge-Scope).

DoD:

- [ ] `examples/dash/README.md` beschreibt einen lokalen DASH-Stream
  inklusive MPD-URL, Startpfad und erwarteter Player-/CLI-Nutzung.
- [ ] Das Beispiel erzeugt oder liefert deterministische DASH-Artefakte,
  die ohne externe Netzwerkabhängigkeit im lokalen Lab nutzbar sind.
- [ ] Bevorzugter Pfad ist ein kleiner lokaler DASH-Origin im
  Example-Compose: FFmpeg erzeugt oder ein statischer HTTP-Server liefert
  MPD/Segmente aus einem Example-Verzeichnis. Externe CDNs sind nicht
  Teil des Smoke.
- [ ] Wenn `dash.js` im Dashboard-Demo-Pfad verwendet wird, bleibt der
  bestehende HLS-Demo-Pfad unverändert und die neue Adapter-Logik ist
  optional.
- [ ] `packages/player-sdk` bleibt abwärtskompatibel; ein möglicher
  dash.js-Adapter darf die hls.js-Public-API nicht brechen.
- [ ] `POST /api/analyze` bleibt in `0.5.0` entweder klar HLS-only oder
  erhält einen ausdrücklich geplanten additiven DASH-Pfad. Falls keine
  DASH-Analyse geliefert wird, dokumentieren API- und Analyzer-Doku diese
  Grenze unmissverständlich.
- [ ] `make smoke-dash` prüft mindestens, dass die dokumentierte MPD
  erreichbar ist, mindestens ein referenziertes Segment lokal abrufbar
  ist und der Beispielpfad ohne Internet läuft.

---

## 6. Tranche 5 — WebRTC vorbereitet, nicht produktiv (RAK-39)

Bezug: RAK-39; Lastenheft §7.6 F-62, §12.1 MVP-24, §8.3 NF-14.

Ziel: WebRTC ist als späteres Lab-Szenario vorbereitet, ohne das MVP mit
WebRTC-Monitoring, Signaling-Server oder `getStats()`-Normalisierung zu
belasten.

DoD:

- [ ] `examples/webrtc/README.md` existiert und beschreibt den
  geplanten lokalen WebRTC-Pfad, Mindestdienste, offene Entscheidungen
  und Nicht-Ziele.
- [ ] Wenn MediaMTX-WebRTC genutzt werden soll, sind notwendige Ports,
  NAT-/ICE-Hinweise und Browser-Grenzen dokumentiert.
- [ ] Die README enthält eine klare Entscheidungsmarke, ob `0.5.0`
  lediglich Port-/Konfigurationsvorbereitung liefert oder zusätzlich
  einen lokalen Browser-Handcheck beschreibt.
- [ ] Es gibt keine produktive Behauptung für WebRTC-Metriken in
  Dashboard, API oder Telemetry-Model.
- [ ] Ein optionaler Smoke ist entweder implementiert oder bewusst als
  Folge-Item dokumentiert, weil headless WebRTC in CI zusätzliche
  Browser-/Netzwerkstabilitätsrisiken hat.
- [ ] Falls ein Smoke implementiert wird, ist er als
  `make smoke-webrtc-prep` gekennzeichnet und prüft nur die
  Vorbereitungsgrenze, nicht Playback-Qualität oder `getStats()`.
- [ ] Roadmap/Folge-Plan nennt, was aus der Vorbereitung später echte
  WebRTC-Metrikunterstützung macht.

---

## 7. Tranche 6 — Dokumentation, Smokes und Release-Gates (RAK-40)

Bezug: RAK-40; `docs/user/local-development.md`; `README.md`;
`docs/user/releasing.md`; bestehende Gates im `Makefile`.

Ziel: Alle Beispiele sind für Nutzer auffindbar, startbar und mit
klaren Grenzen dokumentiert. Release-Akzeptanz prüft RAK-36..RAK-40
explizit, ohne die normalen Qualitäts-Gates zu schwächen.

DoD:

- [ ] `README.md` Release-Übersicht und Schnellstart verweisen auf die
  Multi-Protocol-Beispiele.
- [ ] `docs/user/local-development.md` enthält eine knappe
  Multi-Protocol-Lab-Sektion mit Start-/Stop-/Reset-Hinweisen.
- [ ] Optionales neues `docs/user/multi-protocol-lab.md` bündelt
  MediaMTX, SRT, DASH und WebRTC, falls `local-development.md` sonst zu
  groß wird.
- [ ] Beispiel-READMEs verweisen auf den gebündelten User-Guide, während
  der User-Guide auf die konkreten Example-Verzeichnisse zurückverweist.
- [ ] `Makefile` listet die Pflicht-Smoke-Targets `smoke-mediamtx`,
  `smoke-srt` und `smoke-dash` im `help`-Text; ein optionaler
  `smoke-webrtc-prep` wird nur gelistet, falls er implementiert wird.
- [ ] `docs/user/releasing.md` nennt die zusätzlichen manuellen oder
  automatisierten `0.5.0`-Release-Smokes.
- [ ] RAK-Verifikationsmatrix ist im Plan ergänzt; zum Release-Closeout
  sind alle Muss-Kriterien entweder `[x]` oder explizit `[!]` mit
  Lastenheft-Patch-Pfad.
- [ ] `make gates` ist grün.
- [ ] Pflicht-Smokes für MediaMTX, SRT und DASH sind lokal grün und mit
  Befehl/Datum in §7.2 dokumentiert; WebRTC bleibt nur dann ein Smoke-
  Nachweis, wenn `smoke-webrtc-prep` implementiert wird.
- [ ] Versionen, `CHANGELOG.md`, Planstatus, Roadmap und Release-Notes
  sind für `0.5.0` aktualisiert.
- [ ] `plan-0.5.0.md` ist nach Abschluss nach `docs/planning/done/`
  verschoben; Roadmap verweist danach auf den finalen Pfad.

### 7.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-36 | Muss | `examples/mediamtx/`, MediaMTX-Smoke prüft Erreichbarkeit, published Teststream und HLS-Manifest, Doku-Link | [ ] |
| RAK-37 | Muss | `examples/srt/`, SRT-Smoke prüft Publisher-Verbindung und abspiel- oder analysierbare Ausspielung | [ ] |
| RAK-38 | Muss | `examples/dash/`, DASH-MPD lokal erreichbar und dokumentierter Beispielpfad läuft ohne Internet | [ ] |
| RAK-39 | Soll | `examples/webrtc/README.md` und ggf. optionaler Smoke oder bewusst dokumentiertes Folge-Item | [ ] |
| RAK-40 | Muss | README/User-Doku/Releasing-Doku vollständig, Docs-Gate grün | [ ] |

### 7.2 Release-Closeout-Protokoll

Dieser Abschnitt wird während Tranche 6 gefüllt, damit Smoke-Nachweise
nicht ad hoc in Commit-Bodies oder Release-Notes verschwinden.

| Prüfung | Befehl / Nachweis | Datum | Ergebnis | Notiz |
| ------- | ----------------- | ----- | -------- | ----- |
| `make gates` | — | — | [ ] | — |
| MediaMTX-Smoke | — | — | [ ] | RAK-36 |
| SRT-Smoke | — | — | [ ] | RAK-37 |
| DASH-Smoke | — | — | [ ] | RAK-38 |
| WebRTC-Vorbereitung | — | — | [ ] | RAK-39 |
| Docs-Gate | `make docs-check` oder Teil von `make gates` | — | [ ] | RAK-40 |
