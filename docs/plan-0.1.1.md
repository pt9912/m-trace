# Implementation Plan — `0.1.1` (Player-SDK + Dashboard)

> **Status**: 🟡 in Arbeit. Beginnt nach Abschluss von `0.1.0` (Backend Core + Demo-Lab).
> **Bezug**: [Lastenheft `1.1.6`](./lastenheft.md) §13.2 (RAK-2, RAK-5, RAK-7), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).
> **Vorgänger-Gate (Stand zum `0.1.1`-Start, nicht zum heutigen Zeitpunkt)**: [`plan-0.1.0.md`](./plan-0.1.0.md) muss bis zum Start dieser Plan-Doku in folgendem Zustand sein:
>
> - Tranche 0 (Pre-MVP-Vorbereitung): `[x]` — bereits heute erfüllt.
> - Tranchen 0a (Architektur- und Plan-Doku) und 0b (Spike-Code-Korrekturen): alle DoD-Items `[x]`. Heute laufende `[ ]`-Items werden im Verlauf der `0.1.0`-Implementierung geschlossen — sie sind Stand des Plans, nicht des Gates.
> - §5.1–§5.4 (Tranche 1 — Backend-Erweiterung, Compose-Lab Core, RAKs, Übergreifende DoD): alle DoD-Items `[x]`, insbesondere CI-Pflicht-Item aus §5.4.
> - **Tranche 0c (Lastenheft-Patches)**: konstruktionsbedingt fortlaufend offen; das Gate verlangt nur, dass alle bis zum `0.1.1`-Start eingetragenen §4a.x-Items entweder `[x]` oder explizit als nicht-blockierend markiert sind — nicht den Abschluss der Tranche selbst.
>
> Konsequenz: solange `0.1.0` nicht released ist, hat dieses Plan-Dokument Status ⬜ in Tranchen-Übersicht und Roadmap §3. Implementierungsarbeit an `0.1.1`-Items beginnt erst, wenn das Gate erfüllt ist.  
> **Nachfolger**: [`plan-0.1.2.md`](./plan-0.1.2.md) (Observability-Stack).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog [`plan-0.1.0.md`](./plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz (siehe `roadmap.md` §7.1).
- 🟡 in Arbeit.

Tranchen 0/0a/0b/0c werden in `plan-0.1.0.md` gepflegt — neue Lastenheft-Patches in der `0.1.1`-Phase landen ebenfalls dort, in einem neuen §4a-Eintrag, weil Patches projektweit gelten.

---

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
|---|---|---|
| 0 | Vorgänger-Gate-Verifikation | ✅ |
| 1 | Player-SDK unter `packages/player-sdk` | 🟡 |
| 2 | Dashboard unter `apps/dashboard` | ⬜ |
| 3 | Compose-Lab-Erweiterung um den `dashboard`-Service | ⬜ |
| 4 | Release-Akzeptanzkriterien `0.1.1` | ⬜ |

---

## 1a. Tranche 0 — Vorgänger-Gate-Verifikation

Konvertiert die narrative Vorgänger-Gate-Beschreibung aus §0 in prüfbare DoD-Items. Gate ist in zwei Kategorien geteilt: **harte Voraussetzungen** (alle `[x]`) und **weiche Voraussetzungen** (offen erlaubt, wenn explizit als nicht-blockierend markiert). Tranche ist „erfüllt", wenn alle harten und alle blockierenden weichen Items `[x]` sind.

DoD — **harte Voraussetzungen, technisch zwingend** (Pflicht `[x]` vor `0.1.1`-Start; ohne diese kann die `0.1.1`-Implementierung nicht starten):

- [x] `plan-0.1.0.md` §3.5 telemetry-model.md, **Pflicht-Anteile für `0.1.1`** — Wire-Format (F-106..F-115) und Backpressure-/Limit-Regeln (F-118..F-123): das Player-SDK muss das Format und die Limits kennen, um Events korrekt zu senden. Andere `0.1.0`-Bereiche von §3.5 (OTel-Modell, Cardinality, Time-Stempel, Schema-Versionierung) sind weiche Voraussetzungen (`e532e1e`, `51b3812`).
- [x] `plan-0.1.0.md` §4.2 Counter-Scope (InvalidEvents/DroppedEvents-Drops + Tests) `[x]` (`372a6d4`, `9fddfa1`).
- [x] `plan-0.1.0.md` §4.3 Telemetry-Driven-Port + OTel-Counter + Request-Span + autoexport `[x]` (`51b3812`, `46e45ec`).
- [x] `plan-0.1.0.md` §5.1 Backend-Erweiterung `[x]` (insbesondere Stream-Sessions-Endpoints, MVP-16 Persistenz, CORS, Rate-Limit-Dimensionen) (`26a64e2`, `504e4c9`).
- [x] `plan-0.1.0.md` §5.2 Compose-Lab Core `[x]` (`504e4c9`).
- [x] `plan-0.1.0.md` §5.3 Release-Akzeptanzkriterien `0.1.0` (RAK-1/3/4/6/8) `[x]` (`504e4c9`).
- [x] `plan-0.1.0.md` §5.4 Übergreifende DoD `0.1.0` `[x]`, insbesondere CI-Pflicht-Item (`46e45ec`, `95359df`).

DoD — **weiche Voraussetzungen, Dokumentations-/Aufräumarbeiten** (offen erlaubt; Gate **nicht** blockierend, sollten aber bis zum `0.1.0`-Release-Tag geschlossen werden):

- [x] `plan-0.1.0.md` §3.5 telemetry-model.md, **nicht-Pflicht-Anteile für `0.1.1`** — OTel-Modell §2, Cardinality §3, Time-Stempel §5, Schema-Versionierung §6: nur indirekt für SDK-Implementierung relevant (`e532e1e`, `51b3812`).
- [x] `plan-0.1.0.md` §3.6 local-development.md: Developer-Guide; `0.1.1`-Implementierung kann mit dem bestehenden `apps/api`-Setup arbeiten (`2eede43`, `504e4c9`, `35eba88`).
- [x] `plan-0.1.0.md` §4.4 Code-Step-Numbering: Code-Kommentar-Cleanup ohne `0.1.1`-Auswirkung (`dbdcb67`).
- [x] `plan-0.1.0.md` Tranche 0c §4a.x-Items sind bis Patch `1.1.6` geschlossen; keine offenen blockierenden Patch-Items zum `0.1.1`-Start.
    - **blockierend** → muss `[x]` sein (z. B. Lastenheft-Patches, deren Wording die `0.1.1`-Implementierung direkt betrifft), **oder**
    - **nicht-blockierend** → offen erlaubt, mit ausdrücklichem `(nicht-blockierend für 0.1.1)`-Vermerk im jeweiligen §4a.x-Eintrag.
- [x] Vorgänger-Gate-Verifikations-Commit dokumentiert die Einstufung pro offenem Item nachvollziehbar (`35eba88`).

---

## 2. Tranche 1 — Player-SDK (`packages/player-sdk`)

Bezug: MVP-5, MVP-6, F-63..F-67; OE-8 (Paketnamen für npm) wird hier entschieden.

DoD:

**Workspace-Bootstrap** (Repo-weit, gemeinsame Voraussetzung mit Tranche 2 Dashboard):

- [x] Root-`package.json` mit Mono-Repo-Marker und Top-Level-Scripts (`build`, `test`, `lint`, `check`); kein eigener Source-Code-Inhalt (`35eba88`).
- [x] `pnpm-workspace.yaml` mit `apps/*` und `packages/*` als Workspace-Globs (deckt `apps/dashboard` und `packages/player-sdk` ab) (`35eba88`).
- [x] `pnpm-lock.yaml` versioniert; `.npmrc` mit Engine-Strict und `.nvmrc` mit Node-Major-Pinning im Repo-Root (`35eba88`).
- [x] Top-Level-Scripts delegieren via `pnpm -r --if-present run <task>` an die Workspace-Pakete; `make` bleibt als Compose-/Lab-Wrapper unverändert (`35eba88`).
- [x] Architecture §4.1 Mono-Repo-Layout reflektiert die neuen Wurzel-Dateien (Doku-Update gemeinsam mit dem Bootstrap-Commit) (`35eba88`).

**Player-SDK** (`packages/player-sdk/`):

- [x] TypeScript-Package unter `packages/player-sdk/` (`bae4a2a`).
- [x] **MVP-6** Pragmatische SDK-Struktur ohne vollständige Hexagon-Ceremony — leichte Adapter-Struktur laut Lastenheft §9.2, keine `hexagon/`-Pflicht-Aufteilung mit Domain/Application/Port. Verzeichnislayout analog `architecture.md` §4.1: `core/`, `adapters/hlsjs/`, `transport/`, `types/` (`bae4a2a`).
- [x] **F-63**: Anbindung an ein `HTMLVideoElement` über einen klar abgegrenzten Browser-Adapter (`adapters/hlsjs/` initial; weitere Player als spätere Adapter) (`bae4a2a`).
- [x] **F-64**: Erfassung von Playback-Events aus dem hls.js-Stream (Manifest, Segment, Bitrate-Switch, Rebuffer, Error, …) (`bae4a2a`).
- [x] **F-65**: Erfassung einfacher Metriken pro Session (Startup-Time, Rebuffer-Dauer, …) (`cf07fda`).
- [x] **F-66**: Versand der Events via HTTP an `POST /api/playback-events` mit dem Wire-Format aus `docs/telemetry-model.md`. Batching und Sampling konfigurierbar; OpenTelemetry Web SDK bleibt optionaler späterer Transport-Pfad (`bae4a2a`).
- [x] **F-67**: Trennung von Browser-Adapter (`adapters/hlsjs/`) und fachlicher Tracking-Logik (`core/`) — strukturelle Boundary, kein gegenseitiger Zugriff: `core/` darf den Browser-Adapter nicht direkt importieren (`bae4a2a`).
- [x] Browser-Build (ESM + UMD/IIFE) (`bae4a2a`).
- [x] OE-8 entscheiden (Paketname, Scope): `@m-trace/player-sdk` (`bae4a2a`).
- [x] **F-110 origin-Bucket (Backend)**: vor Beginn der Browser-Integrationstests muss `apps/api` den dritten Rate-Limit-Bucket auf der `origin`-Dimension aktiv haben (Vorbereitung optional in `0.1.0` §5.1, verbindliche Aktivierung spätestens hier). Test: ein Browser-Origin mit aufgebrauchtem origin-Budget liefert `429`, auch wenn project_id-Budget noch frei ist (`75e55e7`, `c15d8e1`).
- [ ] Tests: Unit-Tests für Core-Logik (Sampling, Batching, Session-Metriken) vorhanden (`bae4a2a`, `cf07fda`); Integrationstest gegen das `apps/api` aus `0.1.0` (Browser → API End-to-End) auf den im MVP unterstützten Browsern bleibt offen: **Pflicht** Chrome Desktop (aktuelle Stable) und Firefox Desktop (aktuelle Stable); **eingeschränkt** Safari Desktop (Basis-Playback laut Lastenheft §6/§MVP-Browser-Matrix); iOS Safari, Android Chrome, Smart-TV-Browser, Embedded-WebViews bleiben außerhalb des `0.1.1`-Test-Scope.

---

## 3. Tranche 2 — Dashboard (`apps/dashboard`)

Bezug: MVP-3, MVP-4, F-23..F-28, F-35..F-40; RAK-2, RAK-7; OE-4 (Frontend-Styling) wird hier entschieden.

DoD:

- [x] SvelteKit-App-Skelett unter `apps/dashboard/` (TypeScript, pnpm) (`1a6a6c7`).
- [x] Startseite mit Layout (`1a6a6c7`).
- [x] **F-23 + MVP-12** Dashboard-Route `/sessions` zeigt einfache Session-Liste, ruft `GET /api/stream-sessions` auf (`1a6a6c7`).
- [x] **MVP-13 + MVP-14** Dashboard-Route `/sessions/:id` zeigt einfache Event-Anzeige plus eingebaute Session-/Trace-Ansicht (Timeline der zugehörigen Events), ruft `GET /api/stream-sessions/{id}` auf (`1a6a6c7`).
- [x] **F-24** Anzeige aktueller Playback-Metriken — entweder im `/sessions/:id`-Detail oder als globale Übersicht (z. B. Zähler-Card auf der Startseite) (`1a6a6c7`).
- [x] **F-25** Anzeige von Fehlern und Warnungen — entweder dedizierte Route `/errors` oder als Filter über die Event-Liste (`1a6a6c7`).
- [x] **F-26** Anzeige einfacher Stream-Health-Zustände — Active/Stalled/Ended pro Session sichtbar (Backend-Lifecycle aus `0.1.0` §5.1) (`1a6a6c7`).
- [x] **F-27** Anzeige von Backend- und Telemetrie-Status — Health-Indicator basierend auf `GET /api/health`; Telemetry-Status zunächst minimal („wired"/„nicht konfiguriert"; vollständig wenn `0.1.2` Observability-Profil läuft) (`1a6a6c7`).
- [x] **F-28 + F-36** Test-Player-Integration: Dashboard-Route `/demo` mit hls.js + Player-SDK-Referenzintegration. Pfad in der App: `apps/dashboard/src/routes/demo/` (SvelteKit-Konvention, Lastenheft §7.5.3) (`1a6a6c7`).
- [x] **F-35** Live-Übersicht — Startseite zeigt aggregierten Live-Stand (laufende Sessions, Event-Rate, Fehlerzähler) als Landing (`1a6a6c7`).
- [x] **F-37** Playback-Events anzeigen — eine dedizierte Sicht (Route oder Tab) listet eingehende Events mit Filter nach Session und Event-Typ (`1a6a6c7`).
- [x] **F-38** Stream-Sessions-Übersicht — bereits durch F-23/MVP-12 oben abgedeckt (`1a6a6c7`).
- [x] **F-39** API-Status-Anzeige — bereits durch F-27 oben abgedeckt; F-39 verlangt explizite Sichtbarkeit, also mindestens ein UI-Element mit `connected/disconnected` (`1a6a6c7`).
- [x] **System-Status-Ansicht** (Lastenheft §7.4 Mindestansichten Z. 387): dedizierte Route `/status` (oder klar abgegrenzter Bereich) mit Status-Indicator-Block für (a) API (`/api/health`), (b) Media-Server (MediaMTX-HLS-Endpoint), (c) Observability-Komponenten (Prometheus, Grafana, OTel-Collector — bei deaktiviertem observability-Profil als „inaktiv" gekennzeichnet). Konsolidiert F-27 und F-39 zu einer prüfbaren Ansicht (`1a6a6c7`).
- [x] **F-40** Footer- oder Navigations-Links zu Grafana, Prometheus und MediaMTX-API/Status. Ziele werden aus den Compose-Service-URLs abgeleitet: Grafana `http://localhost:3000`, Prometheus `http://localhost:9090`, MediaMTX-API `http://localhost:9997` (HTTP-API/Status; MediaMTX hat keine native Web-UI, der HLS-Endpoint auf Port `8888` ist Stream-Auslieferung, kein Konsolen-Ersatz). Bei deaktiviertem observability-Profil bleiben die Grafana-/Prometheus-Links als „nicht verfügbar" gekennzeichnet (`1a6a6c7`).
- [x] API-Client mit typisierten Anfragen (`1a6a6c7`).
- [x] **API-Origin-Strategie** für beide Endpoint-Klassen aus `plan-0.1.0.md` §5.1 (`1a6a6c7`):
    - **GET-Routen** (Dashboard-API-Client): im **Vite-Dev-Mode** SvelteKit/Vite-Proxy (`/api/*` → `http://localhost:8080`), damit Browser-CORS entfällt; im **Compose-Production-Build** über getrennten Origin mit den `0.1.0`-CORS-Headers für den Dashboard-Lese-Pfad.
    - **POST `/api/playback-events`** (Player-SDK auf der `/demo`-Route): immer Cross-Origin gegen `apps/api`, weil das SDK projektunabhängig konfigurierbar sein soll. Nutzt die `0.1.0`-CORS-Headers für den Player-SDK-Pfad inklusive Variante-B-Origin-Validierung (Preflight gegen globale Allowed-Origins-Union, POST-Validierung Origin↔`project_id`). Vite-Dev-Proxy ist hier **kein** Ersatz, weil das SDK unabhängig vom Dashboard ausgeliefert werden können muss; im Dev-Mode greift dasselbe CORS-Setup.
- [x] **NF-37 CSP-Beispiele** für `connect-src`: `docs/local-development.md` §3 ergänzt einen Mustertext (z. B. `Content-Security-Policy: default-src 'self'; connect-src 'self' http://localhost:8080`) für Dashboard-Auslieferung; `docs/telemetry-model.md` §1 ergänzt SDK-bezogene `connect-src`-Beispiele für Drittanbieter-Embeds (z. B. `connect-src 'self' https://collector.example.com`) (`35eba88`, `bae4a2a`).
- [x] Frontend-Styling: OE-4 entscheiden — eigenes CSS ohne Tailwind/UI-Library (`1a6a6c7`).

---

## 4. Tranche 3 — Compose-Lab-Erweiterung um `dashboard`-Service

Bezug: MVP-7..MVP-9, F-86; RAK-1 (Update auf vier Pflicht-Mindestdienste), RAK-2.

DoD:

- [ ] `apps/dashboard`-Container im Production-Build oder Vite-Dev-Mode.
- [ ] Compose-Stack ergänzt um den `dashboard`-Service (ohne `profiles:`-Direktive — startet per Default; entspricht der vollständigen Pflicht-Mindestdienste-Tabelle aus Lastenheft §7.8 nach Patch `1.0.2`).
- [ ] `make dev` startet jetzt vier Core-Services (`api`, `dashboard`, `mediamtx`, `stream-generator`) — RAK-1 ist bereits in `0.1.0` mit drei Diensten erfüllt (siehe `plan-0.1.0.md` §5.3); `0.1.1` erweitert die Mindestdienste-Liste um `dashboard`, ohne RAK-1 erneut auszulösen.
- [ ] Dashboard erreichbar unter `http://localhost:5173` (oder Compose-equivalent) — **RAK-2** wird hier neu erfüllt.
- [ ] Smoke-Test `0.1.1`: Browser-Player-Demo-Route lädt den MediaMTX-Stream, sendet via Player-SDK Events an die API, Dashboard zeigt die Session live.

---

## 5. Tranche 4 — Release-Akzeptanzkriterien `0.1.1` (Lastenheft §13.2)

DoD:

- [ ] **RAK-2** Dashboard ist erreichbar; `make dev` startet zusätzlich den `dashboard`-Service (Tranche 3).
- [ ] **RAK-5** Player-SDK sendet hls.js-basierte Events (Tranche 1).
- [ ] **RAK-7** Dashboard zeigt empfangene Events und einfache Session-Zusammenhänge (Tranche 2).

### 5.1 Übergreifende DoD `0.1.1` (Lastenheft §18, `0.1.1`-Anteil)

- [ ] CI deckt zusätzlich Player-SDK- und Dashboard-Builds ab (CI als Pflicht-Bestandteil ist bereits in `0.1.0` §5.4 erforderlich; `0.1.1` ergänzt nur Coverage für die neuen Pakete).
- [ ] `CHANGELOG.md` enthält Eintrag für `0.1.1`.
- [ ] README ergänzt um die Player-SDK- und Dashboard-Quickstart-Schritte (RAK-8-Refinement).

---

## 6. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`, Commit-Hash anhängen.
- Neue Findings in `0.1.1`-Phase landen entweder in dieser Datei oder in `risks-backlog.md`.
- Lastenheft-Patches während `0.1.1` werden in `plan-0.1.0.md` Tranche 0c als neue §4a.x-Einträge ergänzt (zentrale Patch-Historie).
- Beim Release-Bump `0.1.1` → `0.1.2`: dieses Dokument als historisch archivieren; Lieferstand wandert dokumentarisch nach `CHANGELOG.md`.
