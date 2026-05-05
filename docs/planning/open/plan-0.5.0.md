# Implementation Plan — `0.5.0` (Multi-Protocol Lab)

> **Status**: ⬜ geplant. Beginnt erst nach abgeschlossenem Release-Artefakt
> `v0.4.0` (Release-Commit, annotierter Tag, Push, GitHub-Release und grüne
> GitHub Actions am Tag-Commit). Dieses Dokument liegt bis zum Scope-Cut unter
> `docs/planning/open/`.
>
> **Bezug**: [Lastenheft `1.1.8`](../../../spec/lastenheft.md) §7.1
> (Repo-Struktur, `examples/`), §7.6 (Player-Adapter-Folgeoptionen), §7.8
> (lokales Streaming-Lab), §8.3 (Erweiterbarkeit), §10.5 (Infrastruktur),
> §13.7 (RAK-36..RAK-40); [Roadmap](../in-progress/roadmap.md) §3/§4;
> [Architektur](../../../spec/architecture.md) §8.2; [Local Development](../../user/local-development.md);
> [Stream Analyzer](../../user/stream-analyzer.md); [Risiken-Backlog](./risks-backlog.md).
>
> **Vorgänger-Gate (Stand zum `0.5.0`-Start)**:
>
> - [`plan-0.4.0.md`](../done/plan-0.4.0.md) ist vollständig (`[x]`)
>   und nach `docs/planning/done/` verschoben.
> - `v0.4.0` ist veröffentlicht; GitHub Actions `Build` ist für den
>   Release-Commit grün.
> - Roadmap §1/§2/§3 ist auf `0.5.0` als nächstes aktives Release
>   umgestellt.
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
- [ ] Scope-Entscheidung ist dokumentiert: SRT in `0.5.0` bedeutet
  Beispiel/Smoke, **nicht** SRT-Health-View, SRT-Metrikimport oder
  CGO-Binding in `apps/api`.
- [ ] WebRTC in `0.5.0` ist als vorbereiteter Beispielplatz
  festgelegt; produktive `getStats()`-Erfassung bleibt out of scope.
- [ ] Bestehende R-2-CGO/SRT-Risiken sind geprüft; falls ein
  SRT-Beispiel zusätzliche Runtime-Risiken erzeugt, wird
  `risks-backlog.md` aktualisiert.

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
- [ ] Jedes Beispiel benennt Zweck, Startbefehl, erwartete Ports/URLs,
  Abbruch-/Reset-Pfad und bekannte Grenzen.
- [ ] Compose-Erweiterungen liegen entweder als klar benannte Override-
  Dateien unter `examples/<name>/compose.yaml` oder als Profile im
  Root-Compose; die Entscheidung ist einheitlich dokumentiert.
- [ ] `make dev` bleibt das Core-Lab und startet keine zusätzlichen
  optionalen Beispiele.
- [ ] Neue Smoke-Targets sind opt-in und hängen nicht an `make gates`,
  solange sie zusätzliche Streaming-Images oder lange Medien-Starts
  benötigen.
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
- [ ] MediaMTX-Konfiguration ist so dokumentiert, dass die aktiven
  Protokolle und Ports (`RTSP`, `HLS`, API/Status; optional `RTMP`, falls
  im Beispiel aktiviert) aus einem frischen Clone nachvollziehbar sind.
- [ ] Falls ein separates MediaMTX-Override nötig ist, kollidiert es
  nicht mit dem Core-Lab und nutzt eigene Pfade/Stream-Namen.
- [ ] Ein Smoke-Pfad prüft mindestens: MediaMTX erreichbar,
  Teststream published, HLS-Manifest unter der dokumentierten URL
  erreichbar.
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
- [ ] Ein Startpfad ist dokumentiert, z. B. ein Compose-Profil oder
  `docker compose -f examples/srt/compose.yaml up --build`.
- [ ] Der Smoke-Pfad prüft mindestens, dass ein SRT-Publisher verbinden
  kann und daraus eine abspiel- oder analysierbare Ausspielung entsteht.
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
- [ ] Wenn `dash.js` im Dashboard-Demo-Pfad verwendet wird, bleibt der
  bestehende HLS-Demo-Pfad unverändert und die neue Adapter-Logik ist
  optional.
- [ ] `packages/player-sdk` bleibt abwärtskompatibel; ein möglicher
  dash.js-Adapter darf die hls.js-Public-API nicht brechen.
- [ ] `POST /api/analyze` bleibt in `0.5.0` entweder klar HLS-only oder
  erhält einen ausdrücklich geplanten additiven DASH-Pfad. Falls keine
  DASH-Analyse geliefert wird, dokumentieren API- und Analyzer-Doku diese
  Grenze unmissverständlich.
- [ ] Smoke-Pfad prüft mindestens, dass die dokumentierte MPD erreichbar
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
- [ ] Es gibt keine produktive Behauptung für WebRTC-Metriken in
  Dashboard, API oder Telemetry-Model.
- [ ] Ein optionaler Smoke ist entweder implementiert oder bewusst als
  Folge-Item dokumentiert, weil headless WebRTC in CI zusätzliche
  Browser-/Netzwerkstabilitätsrisiken hat.
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
- [ ] `Makefile` listet die neuen opt-in Smoke-Targets im `help`-Text,
  falls solche Targets implementiert werden.
- [ ] `docs/user/releasing.md` nennt die zusätzlichen manuellen oder
  automatisierten `0.5.0`-Release-Smokes.
- [ ] RAK-Verifikationsmatrix ist im Plan ergänzt und alle Muss-Kriterien
  sind entweder `[x]` oder explizit `[!]` mit Lastenheft-Patch-Pfad.
- [ ] `make gates` ist grün.
- [ ] Relevante optionale Smokes sind lokal grün und mit Befehl/Datum im
  Abschlussabschnitt dokumentiert.
- [ ] Versionen, `CHANGELOG.md`, Planstatus, Roadmap und Release-Notes
  sind für `0.5.0` aktualisiert.

### 7.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-36 | Muss | `examples/mediamtx/`, MediaMTX-Smoke, Doku-Link | ⬜ |
| RAK-37 | Muss | `examples/srt/`, SRT-Smoke oder dokumentierter lokaler Startnachweis | ⬜ |
| RAK-38 | Muss | `examples/dash/`, DASH-MPD lokal erreichbar, Doku-Link | ⬜ |
| RAK-39 | Soll | `examples/webrtc/README.md` und ggf. optionaler Smoke | ⬜ |
| RAK-40 | Muss | README/User-Doku/Releasing-Doku vollständig, Docs-Gate grün | ⬜ |
