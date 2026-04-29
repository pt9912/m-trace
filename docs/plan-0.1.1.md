# Implementation Plan — `0.1.1` (Player-SDK + Dashboard)

> **Status**: ⬜ offen. Beginnt nach Abschluss von `0.1.0` (Backend Core + Demo-Lab).  
> **Bezug**: [Lastenheft `1.1.2`](./lastenheft.md) §13.2 (RAK-2, RAK-5, RAK-7), §18 (MVP-DoD-Anteil); [Roadmap](./roadmap.md) §3; [Architektur (Zielbild)](./architecture.md); [API-Kontrakt](./spike/backend-api-contract.md); [Risiken-Backlog](./risks-backlog.md).  
> **Vorgänger**: [`plan-0.1.0.md`](./plan-0.1.0.md) (Backend Core + Demo-Lab — alle Tranchen 0..0c und §5.1–§5.4 müssen abgeschlossen sein, inklusive Release-Akzeptanzkriterien `0.1.0` (§5.3) und übergreifender DoD `0.1.0` (§5.4); insbesondere CI-Pflicht-Item).  
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
| 1 | Player-SDK unter `packages/player-sdk` | ⬜ |
| 2 | Dashboard unter `apps/dashboard` | ⬜ |
| 3 | Compose-Lab-Erweiterung um den `dashboard`-Service | ⬜ |
| 4 | Release-Akzeptanzkriterien `0.1.1` | ⬜ |

---

## 2. Tranche 1 — Player-SDK (`packages/player-sdk`)

Bezug: MVP-5, F-63..F-67; OE-8 (Paketnamen für npm) wird hier entschieden.

DoD:

- [ ] TypeScript-Package unter `packages/player-sdk/`.
- [ ] **F-63**: Anbindung an ein `HTMLVideoElement` über einen klar abgegrenzten Browser-Adapter (`adapters/hlsjs/` initial; weitere Player als spätere Adapter).
- [ ] **F-64**: Erfassung von Playback-Events aus dem hls.js-Stream (Manifest, Segment, Bitrate-Switch, Rebuffer, Error, …).
- [ ] **F-65**: Erfassung einfacher Metriken pro Session (Startup-Time, Rebuffer-Dauer, …).
- [ ] **F-66**: Versand der Events via HTTP an `POST /api/playback-events` mit dem Wire-Format aus `docs/telemetry-model.md`. Batching und Sampling konfigurierbar; OpenTelemetry Web SDK als optionaler zweiter Transport-Pfad.
- [ ] **F-67**: Trennung von Browser-Adapter (`adapters/hlsjs/`) und fachlicher Tracking-Logik (`core/`) — strukturelle Boundary, kein gegenseitiger Zugriff: `core/` darf den Browser-Adapter nicht direkt importieren.
- [ ] Browser-Build (ESM + UMD).
- [ ] OE-8 entscheiden (Paketname, Scope).
- [ ] Tests: Unit-Tests für Core-Logik (Sampling, Batching), Integration-Test gegen das `apps/api` aus `0.1.0` (Browser → API End-to-End).

---

## 3. Tranche 2 — Dashboard (`apps/dashboard`)

Bezug: MVP-3, F-23..F-28, F-35..F-40; RAK-2, RAK-7; OE-4 (Frontend-Styling) wird hier entschieden.

DoD:

- [ ] SvelteKit-App-Skelett unter `apps/dashboard/` (TypeScript, pnpm).
- [ ] Startseite mit Layout.
- [ ] **F-23 + MVP-12** Dashboard-Route `/sessions` zeigt einfache Session-Liste, ruft `GET /api/stream-sessions` auf.
- [ ] **MVP-13 + MVP-14** Dashboard-Route `/sessions/:id` zeigt einfache Event-Anzeige plus eingebaute Session-/Trace-Ansicht (Timeline der zugehörigen Events), ruft `GET /api/stream-sessions/{id}` auf.
- [ ] **F-24** Anzeige aktueller Playback-Metriken — entweder im `/sessions/:id`-Detail oder als globale Übersicht (z. B. Zähler-Card auf der Startseite).
- [ ] **F-25** Anzeige von Fehlern und Warnungen — entweder dedizierte Route `/errors` oder als Filter über die Event-Liste.
- [ ] **F-26** Anzeige einfacher Stream-Health-Zustände — Active/Stalled/Ended pro Session sichtbar (Backend-Lifecycle aus `0.1.0` §5.1).
- [ ] **F-27** Anzeige von Backend- und Telemetrie-Status — Health-Indicator basierend auf `GET /api/health`; Telemetry-Status zunächst minimal („wired"/„nicht konfiguriert"; vollständig wenn `0.1.2` Observability-Profil läuft).
- [ ] **F-28 + F-36** Test-Player-Integration: Dashboard-Route `/demo` mit hls.js + Player-SDK-Referenzintegration. Pfad in der App: `apps/dashboard/src/routes/demo/` (SvelteKit-Konvention, Lastenheft §7.5.3).
- [ ] **F-35** Live-Übersicht — Startseite zeigt aggregierten Live-Stand (laufende Sessions, Event-Rate, Fehlerzähler) als Landing.
- [ ] **F-37** Playback-Events anzeigen — eine dedizierte Sicht (Route oder Tab) listet eingehende Events mit Filter nach Session und Event-Typ.
- [ ] **F-38** Stream-Sessions-Übersicht — bereits durch F-23/MVP-12 oben abgedeckt.
- [ ] **F-39** API-Status-Anzeige — bereits durch F-27 oben abgedeckt; F-39 verlangt explizite Sichtbarkeit, also mindestens ein UI-Element mit `connected/disconnected`.
- [ ] **System-Status-Ansicht** (Lastenheft §7.4 Mindestansichten Z. 387): dedizierte Route `/status` (oder klar abgegrenzter Bereich) mit Status-Indicator-Block für (a) API (`/api/health`), (b) Media-Server (MediaMTX-HLS-Endpoint), (c) Observability-Komponenten (Prometheus, Grafana, OTel-Collector — bei deaktiviertem observability-Profil als „inaktiv" gekennzeichnet). Konsolidiert F-27 und F-39 zu einer prüfbaren Ansicht.
- [ ] **F-40** Footer- oder Navigations-Links zu Grafana, Prometheus und MediaMTX-Konsole. Ziele werden aus den Compose-Service-URLs abgeleitet (z. B. `http://localhost:3000` Grafana, `http://localhost:9090` Prometheus, `http://localhost:8888` MediaMTX-Web-UI). Bei deaktiviertem observability-Profil bleiben die Grafana-/Prometheus-Links als „nicht verfügbar" gekennzeichnet.
- [ ] API-Client mit typisierten Anfragen.
- [ ] Frontend-Styling: OE-4 entscheiden (eigenes CSS / Tailwind / UI-Library).

---

## 4. Tranche 3 — Compose-Lab-Erweiterung um `dashboard`-Service

Bezug: MVP-7..MVP-9, F-86; RAK-1 (Update auf vier Pflicht-Mindestdienste), RAK-2.

DoD:

- [ ] `apps/dashboard`-Container im Production-Build oder Vite-Dev-Mode.
- [ ] Compose-Stack ergänzt um den `dashboard`-Service (ohne `profiles:`-Direktive — startet per Default; entspricht der vollständigen Pflicht-Mindestdienste-Tabelle aus Lastenheft §7.8 nach Patch `1.0.2`).
- [ ] `make dev` startet jetzt vier Core-Services (`api`, `dashboard`, `mediamtx`, `stream-generator`) und erfüllt damit RAK-1 vollständig.
- [ ] Dashboard erreichbar unter `http://localhost:5173` (oder Compose-equivalent) — RAK-2 erfüllt.
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
