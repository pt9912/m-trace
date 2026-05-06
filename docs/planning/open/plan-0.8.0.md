# Implementation Plan — `0.8.0` (Player-SDK-WebRTC-Adapter)

> **Status**: ⬜ geplant (Plan-Skelett, liegt unter
> `docs/planning/open/`). Vorgänger `v0.7.0` ist released
> (Tag `v0.7.0` auf `11a3368`, CI-Run 25418640334 grün; Plan archiviert
> in [`done/plan-0.7.0.md`](../done/plan-0.7.0.md)). Tranche 0
> aktiviert die Phase, sobald der zugehörige Lastenheft-Patch
> `1.1.10` fertig ist (siehe §0.2). Plan wandert dann atomar nach
> `docs/planning/in-progress/`.
>
> **Lastenheft-Status**: `1.1.9` führt RAK-51 als „Kann"
> (`@npm9912/player-sdk` exposed einen optionalen WebRTC-Adapter-
> Pfad ohne Vermischung mit `hls.js`; Public-API bleibt
> abwärtskompatibel). `0.8.0` zieht RAK-51 verbindlich aus dem
> deferred-Status; der Lastenheft-Patch `1.1.10` (siehe §0.2)
> hebt RAK-51 auf „Muss/Soll" und ergänzt 2–3 neue RAK für die
> Sub-Items.
>
> **Bezug**: [Lastenheft `1.1.9`](../../../spec/lastenheft.md) §7.6
> F-62 (Player-Adapter-Folgeoptionen), §8.3 NF-14 (Erweiterbarkeit),
> §12.1 MVP-24, §13.9 RAK-51 (Player-SDK-WebRTC-Adapter, Kann);
> [`done/plan-0.7.0.md`](../done/plan-0.7.0.md) §7 (Folge-Anker
> RAK-51); [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
> §3.2 (WebRTC-Aggregat-Allowlist) und §3.5 (Future-Telemetry-Notiz —
> wird in dieser Phase produktiv aktiviert);
> [`examples/webrtc/`](../../../examples/webrtc/) (Lab-Compose ab
> `0.7.0` Tranche 1).
>
> **Nachfolger**: offen — kein `plan-0.9.0.md` vorbereitet.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`done/plan-0.1.0.md`](../done/plan-0.1.0.md) §0:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Lastenheft-Inkonsistenz oder offene
  Entscheidung.
- 🟡 in Arbeit.

Scope-Grenze: dieser Plan liefert einen **produktiven Player-SDK-
WebRTC-Adapter** als Folge zur Lab-Erweiterung aus `0.7.0`. Er
aktiviert WebRTC-Telemetrie auf der bounded Allowlist aus
`spec/telemetry-model.md` §3.5 und schließt R-12 (WebRTC-`getStats()`-
Schema-Drift) als Spec-/Adapter-Review-Gate ab. Er liefert keinen
neuen Media-Server, keinen Cross-Browser-Auto-Failover und keine
Multi-Tenant-Erweiterung der Cardinality-Regeln.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- `0.7.0` ist released (Tag `v0.7.0` auf `11a3368`); Lab-Compose
  unter [`examples/webrtc/`](../../../examples/webrtc/),
  `make smoke-webrtc-prep` opt-in grün, `spec/telemetry-model.md`
  §3.5 als Future-Telemetry-Notiz vorhanden.
- Lastenheft-Patch `1.1.10` (siehe §0.2) ist akzeptiert; RAK-51
  steht auf „Muss/Soll" und neue RAK-52..RAK-55 sind im Lastenheft
  §13.10 (oder analog) verankert.
- Browser-Support-Matrix-Strategie für WebRTC ist bestätigt:
  Chromium 120+ und Firefox 120+ sind Pflicht; Safari ist
  Best-Effort (analog zur RAK-19-Linie aus `plan-0.2.0.md`).

### 0.2 Lastenheft-Patch `1.1.10` (Vorschlag)

Der Patch hebt RAK-51 verbindlich aus dem Kann-Status und ergänzt
4 neue RAK für die Sub-Items, die `0.8.0` ausliefert. Die genauen
Wortlaute sind beim Tranche-0-Closeout in `spec/lastenheft.md` §13.10
(neuer Block) und im Wartungslog (analog `done/plan-0.1.0.md`
Tranche 0c) festzuhalten:

| RAK | Priorität | Inhalt (Vorschlag für `spec/lastenheft.md` §13.10) |
| --- | --------- | -------------------------------------------------- |
| RAK-51 | Muss | `@npm9912/player-sdk` exposed einen produktiven WebRTC-Adapter-Pfad ohne Vermischung mit `hls.js`; Public-API bleibt abwärtskompatibel. **Hochstufung von 1.1.9 §13.9 „Kann" auf „Muss".** |
| RAK-52 | Muss | Public-API für Adapter-Auswahl (`Hls`/`WebRTC`) ist dokumentiert; hls.js-Pfad bleibt Default und unverändert; opt-in pro Player-Instanz. Pack-Smoke und Browser-Support-Matrix erweitert. |
| RAK-53 | Soll | Produktive WebRTC-Telemetrie auf bounded Allowlist aus `spec/telemetry-model.md` §3.2 (`connection_state`, `ice_state`, `dtls_state`); `mtrace_webrtc_*`-Counter im API-Ingress; `scripts/smoke-observability.sh` spiegelt die WebRTC-Forbidden-Liste aus §3.1. |
| RAK-54 | Soll | `getStats()`-Sammlung im SDK aktiv; Muss-/Soll-Felder pro `RTCStatsType`-Gruppe aus §3.5.2 werden geliefert; Schema-Drift-Strategie aus §3.5.3 ist im Adapter-Code umgesetzt. R-12 wird ab diesem Punkt release-blockierend. |
| RAK-55 | Kann | Browser-E2E-Smoke (Playwright) für den WebRTC-Adapter-Pfad gegen das `examples/webrtc/`-Lab; opt-in im CI-Workflow. |

Begründung der Hochstufung: das `0.7.0`-Lab hat alle Vorbedingungen
(Compose, WHIP/WHEP-Endpoints, Telemetrie-Spec) geliefert; ohne
produktiven Adapter bleibt der WebRTC-Lab-Pfad ohne Operator-Use-Case
im Dashboard.

### 0.3 Out-of-Scope-Klauseln (durchgängig)

- Kein neuer Media-Server. `0.8.0` nutzt das `examples/webrtc/`-Lab
  (MediaMTX 1.x mit WHIP/WHEP) als einzigen produktiven Test-Pfad.
- Kein Cross-Browser-Auto-Failover. Falls eine Browser-Version
  ein Muss-Feld aus §3.5.2 nicht liefert, bleibt die Aggregat-
  Metrik leer (siehe §3.5.3 Fallback); SDK-Konsumenten können
  zwischen Adaptern manuell wechseln.
- Keine Multi-Tenant-Erweiterung der Cardinality-Regeln. Die
  WebRTC-Aggregat-Labels bleiben auf `connection_state`/`ice_state`/
  `dtls_state` beschränkt; eine `project_id`-Allowlist bleibt
  Folge-ADR (Roadmap §4 CORS-Preflight Variante A).
- Kein Player-SDK-Major-Bump. Public-API bleibt abwärtskompatibel
  (RAK-51 / RAK-52); WebRTC-Adapter ist additiv und opt-in pro
  Player-Instanz.
- Keine produktive Verteilung des WebRTC-Adapters über CDN. Der
  Pack-Smoke aus `0.2.0` Tranche 6 wird erweitert, aber `npm
  publish` bleibt manueller Operator-Schritt analog zu `0.2.0`.
- Kein TLS-/Public-Internet-Setup. Lab-Pfad bleibt localhost
  (analog `0.7.0` Out-of-Scope §0.3).

### 0.4 Sequenzierung und harte Gates

1. Tranche 1 (Public-API-Spec) liefert die Adapter-Auswahl-Surface
   und die hls.js-Trennung. Tranche 2 darf keine impliziten WebRTC-
   Defaults im hls.js-Pfad voraussetzen.
2. Tranche 3 (Telemetrie-Aktivierung) ist der Punkt, an dem R-12
   (WebRTC-`getStats()`-Schema-Drift) release-blockierend wird;
   ab hier muss `scripts/smoke-observability.sh` die WebRTC-
   Forbidden-Liste aus `spec/telemetry-model.md` §3.1 spiegeln und
   die Allowlist-Labels aus §3.2 als bounded prüfen.
3. Tranche 4 (Compat-Tests) erweitert die Browser-Support-Matrix
   aus `0.2.0` Tranche 5; ein Pack-Smoke-Bug im WebRTC-Adapter ist
   release-blockierend.
4. Tranche 5 (Closeout) verschiebt diesen Plan nach `done/`,
   bumpt die Versionen 0.7.0 → 0.8.0 in allen package.json/Test-
   Fixtures (analog `0.7.0` Tranche 5; `contracts/sdk-compat.json`
   ist mit dabei) und setzt den Tag `v0.8.0`.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung (`open/` → `in-progress/`) + Lastenheft-Patch `1.1.10` (RAK-51 hochziehen, RAK-52..RAK-55 ergänzen) + ggf. Toolchain-Hardening | ⬜ |
| 1 | Public-API-Spec für Adapter-Auswahl in `@npm9912/player-sdk` (RAK-51 / RAK-52) | ⬜ |
| 2 | WebRTC-Adapter-Implementation gegen WHEP-Pfad aus `examples/webrtc/` | ⬜ |
| 3 | Produktive WebRTC-Telemetrie aktivieren (Allowlist aus §3.2/§3.5; `mtrace_webrtc_*`-Counter; `smoke-observability` spiegelt §3.1; R-12 release-blockierend) | ⬜ |
| 4 | Compat-Tests + Browser-Support-Matrix-Erweiterung; Pack-Smoke; SDK-Performance-Budget verifizieren | ⬜ |
| 5 | Release-Doku, RAK-Verifikationsmatrix und Closeout (Versions-Bump 0.7.0 → 0.8.0, Plan nach `done/`, Tag `v0.8.0`) | ⬜ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Lastenheft-Patch

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche analog
`done/plan-0.7.0.md` §1a.

DoD:

- [ ] Plan-Skelett von `docs/planning/open/plan-0.8.0.md` nach
  `docs/planning/in-progress/plan-0.8.0.md` verschoben (Status
  `⬜ → 🟡`); Cross-Refs in `roadmap.md` §1.2/§3, README,
  `done/plan-0.7.0.md` §7 (falls aktiv) und ggf. `examples/webrtc/`
  README nachgezogen.
- [ ] Lastenheft-Patch `1.1.10` schreiben: RAK-51 von „Kann" auf
  „Muss" hochgezogen; neuer Block §13.10 mit RAK-52..RAK-55 (siehe
  §0.2 Vorschlag). Patch ist als Eintrag im jeweiligen
  `done/`-Plan-Wartungslog (analog `plan-0.1.0.md` Tranche 0c)
  dokumentiert.
- [ ] Optional: Toolchain-Bump prüfen (Go-/Node-/golangci-lint-/
  pnpm-Linien gegen aktuelle non-EOL-Versionen). Wenn Bump nötig
  ist, eigener DoD-Block analog `plan-0.7.0.md` Tranche 0 mit
  Race-/Schema-Stage in `make gates`. Wenn nicht, dokumentieren
  warum nicht.

---

## 2. Tranche 1 — Public-API-Spec

Bezug: Lastenheft `1.1.9` §7.6 F-62 (Player-Adapter-Folgeoptionen);
`docs/planning/done/plan-0.2.0.md` §1 (Public-API-Stand `0.2.0`);
`packages/player-sdk/src/`.

Ziel: Eine Adapter-Auswahl-API ist als Public-API-Vertrag
spezifiziert (z. B. `createPlayer({ kind: "hls" | "webrtc", … })`).
hls.js-Pfad bleibt Default und unverändert; WebRTC ist opt-in pro
Player-Instanz.

DoD:

- [ ] Public-API-Erweiterung in `packages/player-sdk/src/` ist als
  TypeScript-Vertrag (Types + JSDoc) ausgegliedert; hls.js-Pfad
  ist explizit als Default markiert.
- [ ] `packages/player-sdk/README.md` (oder eigenständiges
  `docs/sdk-webrtc.md`) dokumentiert Adapter-Auswahl, opt-in-Form
  und Browser-Anforderungen.
- [ ] Versions-Strategie für die Adapter-Auswahl ist im
  `contracts/sdk-compat.json` widergespiegelt (z. B. Feld-
  Erweiterung), damit Schema-Drift-Tests den neuen Pfad
  erkennen.
- [ ] Migrations-/Compat-Hinweis: bestehender hls.js-only Code
  (z. B. `apps/dashboard` `/demo`-Route) bleibt unverändert
  funktionsfähig; entsprechende Tests und der Pack-Smoke aus
  `0.2.0` bleiben grün.

---

## 3. Tranche 2 — WebRTC-Adapter-Implementation

Bezug: Tranche 1; `examples/webrtc/` (WHEP-Read-Pfad).

Ziel: Ein funktionierender WebRTC-Adapter, der den WHEP-Pfad
(`http://localhost:8892/webrtc-test/whep` aus dem Lab) konsumiert
und die Player-Events analog zum hls.js-Adapter in den Player-SDK-
Event-Stream einspeist.

DoD:

- [ ] WebRTC-Adapter-Klasse/Funktion in `packages/player-sdk/src/`
  ist implementiert; baut eine `RTCPeerConnection`, signalisiert
  via WHEP, und mappt MediaTracks auf das Player-Surface (das
  Demo nutzt `<video>`-Element analog zum hls.js-Pfad).
- [ ] Player-Event-Stream (`playback_started`, `playback_error`,
  `rebuffer_started`, `manifest_loaded`-Pendant für WebRTC, …)
  ist konsistent zu §1.3 des Telemetrie-Modells; ggf. spec-Patch
  für WebRTC-spezifische Event-Felder (z. B.
  `peer_connection_state_changed`).
- [ ] `apps/dashboard` bekommt eine `/demo-webrtc`-Route (oder ein
  Toggle in `/demo`), die den WebRTC-Adapter gegen das
  `examples/webrtc/`-Lab demonstriert. Bestehende `/demo`-Route
  (hls.js) bleibt unverändert.
- [ ] Lokal verifiziert: `make dev` plus `mtrace-webrtc`-Stack +
  Demo-Route zeigt Test-Pattern + Sinuston, Player-Events
  erscheinen in der Dashboard-Session-Timeline.

---

## 4. Tranche 3 — Produktive WebRTC-Telemetrie

Bezug: `spec/telemetry-model.md` §3.2/§3.5;
`docs/planning/open/risks-backlog.md` R-12;
`scripts/smoke-observability.sh`.

Ziel: WebRTC-Telemetrie ist auf der bounded Allowlist aktiv;
`mtrace_webrtc_*`-Counter im API-Ingress; R-12 ist ab dieser
Tranche release-blockierend.

DoD:

- [ ] WebRTC-Adapter sammelt `getStats()`-Reports gemäß §3.5.2
  Muss-/Soll-Felder; nur die in §3.2 dokumentierten bounded
  Aggregat-Labels werden in das Wire-Format eingespeist;
  Per-Identifier-Felder (§3.1 Forbidden) werden hart gefiltert.
- [ ] API-Ingress erkennt WebRTC-Aggregat-Labels und exportiert
  `mtrace_webrtc_*`-Counter (z. B.
  `mtrace_webrtc_connection_state_total{connection_state}`,
  `mtrace_webrtc_ice_state_total{ice_state}`,
  `mtrace_webrtc_dtls_state_total{dtls_state}`); Labelset bleibt
  auf §3.2 + `instance`/`job` beschränkt.
- [ ] `scripts/smoke-observability.sh` spiegelt die WebRTC-
  Forbidden-Liste aus `spec/telemetry-model.md` §3.1 und prüft
  die WebRTC-Counter auf bounded Cardinality (RAK-9-Stil).
- [ ] R-12 wird im Risiken-Backlog von „Triggerschwelle nicht
  ausgelöst" auf „release-blockierend, sobald Browser-Major-Bump
  Schema ändert" angehoben; ein Browser-Drift-Smoke-Plan ist im
  Plan-/Releasing-Doku verankert.
- [ ] Schema-Drift-Strategie aus §3.5.3 ist im Adapter-Code
  umgesetzt: fehlt ein Muss-Feld, wird die Aggregat-Metrik nicht
  emittiert (kein `unknown`-Surrogat); fehlt ein Soll-Feld,
  läuft die übrige Metrik weiter.

---

## 5. Tranche 4 — Compat-Tests + Browser-Support-Matrix

Bezug: `done/plan-0.2.0.md` §5 (Browser-Support-Matrix); Tranche 1–3.

Ziel: Adapter-Auswahl ist gegen Chromium 120+ und Firefox 120+
nachweislich kompatibel; Pack-Smoke aus `0.2.0` ist erweitert;
SDK-Performance-Budget ist eingehalten.

DoD:

- [ ] Browser-E2E-Smoke (Playwright) für den WebRTC-Adapter-Pfad
  ist als opt-in im `make browser-e2e`-Workflow ergänzt; Lab-
  abhängig (benötigt `mtrace-webrtc`-Compose hochgefahren).
- [ ] Pack-Smoke (`packages/player-sdk/scripts/pack-smoke.mjs`)
  prüft, dass der WebRTC-Adapter im Tarball verfügbar und gegen
  die Public-API testbar ist; `expectedVersion` bleibt durch
  Versions-Bump in Tranche 5 abgedeckt.
- [ ] Browser-Support-Matrix in `packages/player-sdk/README.md`
  ist um WebRTC-spezifische Hinweise erweitert (Chromium 120+,
  Firefox 120+, Safari Best-Effort; `getStats()`-Verfügbarkeit
  pro Engine analog §3.5.3).
- [ ] `make sdk-performance-smoke` (oder Pendant) verifiziert,
  dass der WebRTC-Adapter das Performance-Budget aus RAK-18
  einhält; Adapter-Code wird gegen die Player-SDK-Bundle-Größen-
  Grenze gemessen.

---

## 6. Tranche 5 — Release-Doku, RAK-Matrix und Closeout

Bezug: RAK-51 (hochgezogen), RAK-52..RAK-55 (neu);
`docs/user/releasing.md`; `README.md`; `roadmap.md`.

Ziel: `0.8.0` ist auffindbar dokumentiert, Versions-Bump
durchgezogen, Tag `v0.8.0` gesetzt.

DoD:

- [ ] `README.md` und `docs/user/local-development.md` erwähnen
  den WebRTC-Adapter-Pfad und verweisen auf
  `packages/player-sdk/README.md` plus `apps/dashboard/`
  Demo-Route.
- [ ] `docs/user/releasing.md` bekommt einen `0.8.0`-Block
  (manuelle Browser-Handcheck-Prüfung gegen `mtrace-webrtc` plus
  optionaler Browser-E2E-Smoke).
- [ ] RAK-Verifikationsmatrix §6.1 (siehe unten) ist mit Commit-
  Verweisen ausgefüllt; deferred Kann-Anforderungen sind als
  `deferred / Folgeplan` markiert.
- [ ] Versions-Bump 0.7.0 → 0.8.0 in allen package.json (root,
  apps, packages) plus `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts`, `packages/player-sdk/
  scripts/pack-smoke.mjs`, `contracts/sdk-compat.json`
  (`sdk_version`) und allen Test-Fixtures, die SDK-/Analyzer-
  Versions-Strings hartkodieren (analog `0.7.0` Tranche 5;
  Bulk-Fix per `xargs sed -i 's/"0\.7\.0"/"0.8.0"/g'` über die
  Test-Files).
- [ ] CHANGELOG: [Unreleased]-Block in `[0.8.0] - YYYY-MM-DD`
  umgewandelt; neuer leerer [Unreleased]-Block obenauf.
- [ ] `./scripts/verify-doc-refs.sh` (`make docs-check`) grün vor
  Closeout-Commit; `make gates` grün.
- [ ] `plan-0.8.0.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben; alle relativen Cross-Refs
  angepasst (analog `0.7.0` Closeout in Commit `11a3368`); Roadmap
  §3 zeigt `0.8.0` ✅.
- [ ] Tag `v0.8.0` annotiert; Push opt-in (User-Bestätigung);
  GitHub-Release mit CHANGELOG-Block als Notes-Body.

### 6.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-51 | Muss | Public-API in `packages/player-sdk/src/` exposed produktiven WebRTC-Adapter; hls.js-Pfad unverändert. | [ ] |
| RAK-52 | Muss | Adapter-Auswahl dokumentiert in `packages/player-sdk/README.md`; Pack-Smoke + Browser-Support-Matrix erweitert. | [ ] |
| RAK-53 | Soll | `mtrace_webrtc_*`-Counter exportiert; `scripts/smoke-observability.sh` spiegelt §3.1 WebRTC-Forbidden + §3.2 Allowlist. | [ ] |
| RAK-54 | Soll | `getStats()`-Sammlung im SDK aktiv; Schema-Drift-Strategie aus §3.5.3 im Adapter-Code; R-12 release-blockierend. | [ ] |
| RAK-55 | Kann | Browser-E2E-Smoke (Playwright) gegen `examples/webrtc/`-Lab; opt-in. | [ ] |

---

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen (analog `done/plan-0.7.0.md` §7).
- Lastenheft-Patch `1.1.10` (siehe §0.2) ist Vorgänger-Gate für
  Tranche 1; vor Tranche-0-Closeout darf RAK-52..RAK-55 nicht in
  DoD-Items referenziert werden, weil sie noch nicht im Lastenheft
  stehen.
- Wenn ein `0.8.0`-Item in einer Folge-Phase neu bewertet wird
  (z. B. „Cross-Browser-Auto-Failover" doch nötig), entweder
  Folgeplan eröffnen oder hier als Wartungs-Eintrag vermerken.
- R-12 (WebRTC-`getStats()`-Schema-Drift) wechselt mit Tranche 3
  von „Triggerschwelle nicht ausgelöst" auf „release-blockierend";
  Risiken-Backlog-Eintrag muss im selben Commit nachgezogen
  werden.
