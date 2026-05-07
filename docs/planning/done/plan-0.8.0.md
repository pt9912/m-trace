# Implementation Plan — `0.8.0` (Player-SDK-WebRTC-Adapter)

> **Status**: ✅ released am 2026-05-06 (Tranchen 0–5 abgeschlossen;
> Tag `v0.8.0` setzt den Release-Gate-Fix nach dem Closeout). Liegt unter
> `docs/planning/done/`. Vorgänger `v0.7.0` (Tag auf `11a3368`,
> CI-Run 25418640334 grün) bleibt unverändert. RAK-51..RAK-55
> erfüllt; Lastenheft-Patch `1.1.10` (Commit `17a4507`, §13.10) ist
> der zugehörige Vertrag.
>
> **Lastenheft-Status**: ausgeliefert in `1.1.10` — §13.10 enthält
> RAK-51 (hochgestuft auf „Muss") plus RAK-52..RAK-55 für die
> Sub-Items (Public-API + hls.js-Trennung, produktive WebRTC-
> Telemetrie auf bounded Allowlist, `getStats()`-Sammlung mit
> Schema-Drift-Strategie, opt-in Browser-E2E). §13.9 (Block für
> `0.7.0`) bleibt historisch unverändert mit einem Hinweis auf
> §13.10. Lastenheft-Patch ist als §4a.13 in
> [`plan-0.1.0.md`](./plan-0.1.0.md) Tranche 0c
> dokumentiert.
>
> **Bezug**: [Lastenheft `1.1.9`](../../../spec/lastenheft.md) §7.6
> F-62 (Player-Adapter-Folgeoptionen), §8.3 NF-14 (Erweiterbarkeit),
> §12.1 MVP-24, §13.9 RAK-51 (Player-SDK-WebRTC-Adapter, Kann);
> [`plan-0.7.0.md`](./plan-0.7.0.md) §7 (Folge-Anker
> RAK-51); [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
> §3.2 (WebRTC-Aggregat-Allowlist) und §3.5 (Future-Telemetry-Notiz —
> wird in dieser Phase produktiv aktiviert);
> [`examples/webrtc/`](../../../examples/webrtc/) (Lab-Compose ab
> `0.7.0` Tranche 1).
>
> **Nachfolger**: [`plan-0.8.5.md`](./plan-0.8.5.md) (Quality-Gates
> Wave 1, Patch-Release) ausgeliefert; [`plan-0.9.0.md`](./plan-0.9.0.md)
> released am 2026-05-07 (Drift-Smoke + SRS-Lab + DASH-Manifest-Analyse).

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand analog
[`plan-0.1.0.md`](./plan-0.1.0.md) §0:

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
- Lastenheft-Patch `1.1.10` ist ausgeliefert; RAK-51 steht in
  §13.10 auf „Muss" und neue RAK-52..RAK-55 sind dort verankert
  (siehe §0.2).
- Browser-Support-Matrix-Strategie für WebRTC ist bestätigt:
  Chromium 120+ und Firefox 120+ sind Pflicht; Safari ist
  Best-Effort (analog zur RAK-19-Linie aus `plan-0.2.0.md`).

### 0.2 Lastenheft-Patch `1.1.10` (ausgeliefert)

Der Patch hebt RAK-51 verbindlich aus dem Kann-Status und ergänzt
vier neue RAK für die Sub-Items, die `0.8.0` ausliefert. Patch ist
als §4a.13 in [`plan-0.1.0.md`](./plan-0.1.0.md) Tranche 0c
dokumentiert. Inhalt für die Tranchen-DoD unten:

| RAK | Priorität | Inhalt (`spec/lastenheft.md` §13.10) |
| --- | --------- | ------------------------------------ |
| RAK-51 | Muss | `@npm9912/player-sdk` exposed einen produktiven WebRTC-Adapter-Pfad ohne Vermischung mit `hls.js`; Public-API bleibt abwärtskompatibel. **Hochstufung von 1.1.9 §13.9 „Kann" auf „Muss".** |
| RAK-52 | Muss | Public-API für Adapter-Auswahl (z. B. `attachHlsJs(...)` / `attachWebRtc(...)`) ist dokumentiert; hls.js-Pfad bleibt Default und unverändert; opt-in pro Player-Instanz. Pack-Smoke und Browser-Support-Matrix erweitert. |
| RAK-53 | Soll | Produktive WebRTC-Telemetrie auf bounded Allowlist aus `spec/telemetry-model.md` §3.2 (`connection_state`, `ice_state`, `dtls_state`); `mtrace_webrtc_*`-Counter im API-Ingress; `scripts/smoke-observability.sh` spiegelt die WebRTC-Forbidden-Liste aus §3.1. |
| RAK-54 | Soll | `getStats()`-Sammlung im SDK aktiv; Muss-/Soll-Felder pro `RTCStatsType`-Gruppe aus §3.5.2 werden geliefert. Die Muss-Felder sind per Contract plus Metrik-/Read-Pfad nachgewiesen; Schema-Drift-Strategie aus §3.5.3 ist im Adapter-Code umgesetzt. R-12 wird ab diesem Punkt release-blockierend. |
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
   ab hier müssen `contracts/event-schema.json`,
   `contracts/sdk-compat.json`, `spec/backend-api-contract.md` und
   `spec/telemetry-model.md` den produktiven WebRTC-Wire-/Ingress-/
   Metrikvertrag normativ pinnen. `scripts/smoke-observability.sh`
   muss die WebRTC-Forbidden-Liste aus `spec/telemetry-model.md`
   §3.1 spiegeln und die Allowlist-Labels aus §3.2 als bounded
   prüfen.
3. Tranche 4 (Compat-Tests) erweitert die Browser-Support-Matrix
   aus `0.2.0` Tranche 5; ein Pack-Smoke-Bug im WebRTC-Adapter ist
   release-blockierend.
4. Tranche 5 (Closeout) verschiebt diesen Plan nach `done/`,
   bumpt die Versionen 0.7.0 → 0.8.0 in allen package.json/Test-
   Fixtures (analog `0.7.0` Tranche 5; `contracts/sdk-compat.json`
   ist mit dabei) und setzt den Tag `v0.8.0`.

### 0.5 Ziel-Schnitt und Implementierungsleitplanken

Die bevorzugte SDK-Form folgt der bestehenden Paketoberfläche:
`createTracker(...)` bleibt der Telemetrie-Anker, `attachHlsJs(...)`
bleibt der hls.js-Adapter, und WebRTC kommt additiv als eigener
Attach-Pfad hinzu (z. B. `attachWebRtc(video, options, tracker)` plus
`WebRtcAdapter`/`WebRtcAdapterOptions`). Eine neue übergreifende
`createPlayer(...)`-Abstraktion ist nur zulässig, wenn Tranche 1
nachweist, dass sie die bestehende `attachHlsJs`-Surface nicht
entwertet und keine Migration für bestehende Konsumenten erzwingt.

Public-Exports sind Teil des Vertrags: `packages/player-sdk/src/index.ts`
und `packages/player-sdk/scripts/public-api.snapshot.txt` müssen im
gleichen Commit aktualisiert werden. Deep Imports aus `src/` oder
`dist/` bleiben weiterhin nicht public.

Für den lokalen WebRTC-Pfad gilt WHEP als einziger Signalisierungsweg:
SDK erzeugt SDP Offer, POSTet gegen den WHEP-Endpoint aus
`examples/webrtc/`, verarbeitet die SDP Answer, hängt empfangene
MediaTracks an das übergebene `<video>`-Element und räumt PeerConnection
plus WHEP-Resource bei `destroy()` auf. Trickle-ICE, TURN, Auth am
WHEP-Endpoint und Public-Internet-Betrieb bleiben außerhalb dieses Plans.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung (`open/` → `in-progress/`) + Lastenheft-Patch `1.1.10` (RAK-51 hochziehen, RAK-52..RAK-55 ergänzen) + ggf. Toolchain-Hardening | ✅ |
| 1 | Public-API-Spec für Adapter-Auswahl in `@npm9912/player-sdk` (RAK-51 / RAK-52) | ✅ |
| 2 | WebRTC-Adapter-Implementation gegen WHEP-Pfad aus `examples/webrtc/` | ✅ |
| 3 | Produktive WebRTC-Telemetrie aktivieren (Allowlist aus §3.2/§3.5; `mtrace_webrtc_*`-Counter; `smoke-observability` spiegelt §3.1; R-12 release-blockierend) | ✅ |
| 4 | Compat-Tests + Browser-Support-Matrix-Erweiterung; Pack-Smoke; SDK-Performance-Budget verifizieren | ✅ |
| 5 | Release-Doku, RAK-Verifikationsmatrix und Closeout (Versions-Bump 0.7.0 → 0.8.0, Plan nach `done/`, Tag `v0.8.0`) | ✅ |

---

## 1a. Tranche 0 — Plan-Aktivierung + Lastenheft-Patch

Bezug: keine RAK direkt; Wartungs-/Hygiene-Tranche analog
`done/plan-0.7.0.md` §1a.

DoD:

- [x] Plan-Skelett von `docs/planning/open/plan-0.8.0.md` nach
  `docs/planning/in-progress/plan-0.8.0.md` verschoben (Status
  `⬜ → 🟡`); Cross-Refs in `roadmap.md` §1.2/§3 nachgezogen.
  README zeigt seit `0.7.0`-Closeout „nächste Phase offen" — der
  Hinweis bleibt bis zum `0.8.0`-Release-Closeout, weil die Phase
  „in Arbeit" und nicht „released" ist. `done/plan-0.7.0.md` §7
  ist statisch released und wird nicht angefasst (`done/`-Pläne
  sind historisch). `examples/webrtc/` README betrifft nur den
  Lab-Pfad aus `0.7.0` und ist nicht `0.8.0`-spezifisch.
- [x] Lastenheft-Patch `1.1.10` ist ausgeliefert (Commit `17a4507`):
  RAK-51 in §13.10 von „Kann" auf „Muss" hochgezogen; RAK-52..
  RAK-55 als Sub-Items ergänzt; §13.9 RAK-51-Zeile bekommt Hinweis
  auf §13.10. Patch-Log als §4a.13 in
  [`plan-0.1.0.md`](./plan-0.1.0.md) Tranche 0c
  dokumentiert.
- [x] Toolchain-Bump-Check: keine Anpassung nötig. Go (`1.26`),
  golangci-lint (`v2.12.1`), Node (`22 LTS`), pnpm sind seit
  `0.7.0` Tranche 0 (Commits `ccf68b1` + `8bfad21`) aktuell und
  non-EOL. Race-Detector-Stage ist in `make gates` enthalten. Ein
  erneuter Bump wäre Aufwand ohne fachlichen Nutzen — bei
  konkreter EOL-Schwelle oder Security-Advisory wäre er
  nachzuholen.

---

## 2. Tranche 1 — Public-API-Spec

Bezug: Lastenheft `1.1.9` §7.6 F-62 (Player-Adapter-Folgeoptionen);
`docs/planning/done/plan-0.2.0.md` §1 (Public-API-Stand `0.2.0`);
`packages/player-sdk/src/`.

Ziel: Eine Adapter-Auswahl-API ist als Public-API-Vertrag
spezifiziert. Standard ist ein additiver Attach-Pfad analog
`attachHlsJs(...)`; eine neue `createPlayer(...)`-Facade ist nur nach
expliziter Tranche-1-Entscheidung zulässig. hls.js-Pfad bleibt Default
und unverändert; WebRTC ist opt-in pro Player-Instanz.

DoD:

- [x] Public-API-Erweiterung in `packages/player-sdk/src/adapters/
  webrtc/adapter.ts` ist als TypeScript-Vertrag mit JSDoc
  ausgegliedert: `WebRtcAdapter` (Lifecycle mit `destroy()`),
  `WebRtcAdapterOptions` (`whepUrl` Pflicht; `peerConnectionConfig`
  und `signal` optional) und `attachWebRtc(video, options, tracker)`-
  Funktion. hls.js-Pfad bleibt Default — README §Public-API listet
  ihn explizit als „default playback path", WebRTC als „additive
  and opt-in".
- [x] `packages/player-sdk/scripts/public-api.snapshot.txt`
  bewusst aktualisiert (zwei neue `export`-Zeilen); `check-public-
  api.mjs` bleibt das Gate gegen versehentliche Export-Änderungen.
- [x] Contract-Entscheidung: **rein SDK-intern**. Tranche 1 macht
  keinen Patch in `contracts/event-schema.json`,
  `contracts/sdk-compat.json` oder `spec/backend-api-contract.md`.
  Begründung: Adapter-Auswahl ist Sache des SDK-Konsumenten (opt-in
  pro Player-Instanz); Wire-Format und API-Ingress bleiben in
  Tranche 1 unverändert. Tranche 3 erweitert das Wire-Schema um
  den reservierten `webrtc.*`-Meta-Namespace, wenn die produktive
  Telemetrie-Aktivierung das verlangt. Die Entscheidung ist im
  Adapter-JSDoc dokumentiert.
- [x] `packages/player-sdk/README.md` §Public-API erweitert:
  `attachWebRtc`/`WebRtcAdapter`/`WebRtcAdapterOptions` als
  Public-Exports gelistet; neuer Abschnitt „Adapter selection
  (hls.js vs. WebRTC)" dokumentiert opt-in-Form, Browser-
  Anforderungen (Chromium 120+, Firefox 120+, Safari best-effort)
  und die Tranche-1/3-Trennung der Wire-Schema-Wirkung.
- [x] Versions-Strategie für die Adapter-Auswahl: bewusst **nicht**
  in `contracts/sdk-compat.json` widergespiegelt. Begründung: das
  Wire-Schema bleibt in Tranche 1 unverändert (siehe Contract-
  Entscheidung); ein Feld dort jetzt einzuführen, das in Tranche 3
  inhaltlich ersetzt würde, wäre Doku-Drift. Tranche 3 erweitert
  `sdk-compat.json` zusammen mit dem `webrtc.*`-Meta-Namespace.
- [x] Migrations-/Compat-Hinweis: bestehender hls.js-Pfad
  unverändert. `apps/dashboard` `/demo`-Route, `tests/hlsjs-
  adapter.test.ts`, Pack-Smoke aus `0.2.0` und alle anderen
  `tests/`-Pfade bleiben grün — es wurden ausschließlich neue
  Files (`adapters/webrtc/adapter.ts`, `tests/webrtc-adapter.test.ts`)
  und additive Index-Exports angelegt.
- [x] Testform: `tests/webrtc-adapter.test.ts` pinnt vier
  Eigenschaften: (1) `attachWebRtc` ist als Funktion exportiert,
  (2) wirft deterministisch `not implemented (plan-0.8.0
  Tranche 2)` (Schutz gegen versehentlichen Produktiv-Einsatz vor
  Tranche 2), (3) `WebRtcAdapterOptions` mit allen optionalen
  Feldern (`peerConnectionConfig`, `signal`) als Type-Vertrag,
  (4) `WebRtcAdapter` exposed `destroy()`-Surface. Public-API-
  Snapshot-Test ist der vorhandene `check-public-api.mjs`-Pfad
  (geprüft via `make ts-lint`). Dashboard-Mocks bleiben hls.js-only
  unverändert.

---

## 3. Tranche 2 — WebRTC-Adapter-Implementation

Bezug: Tranche 1; `examples/webrtc/` (WHEP-Read-Pfad).

Ziel: Ein funktionierender WebRTC-Adapter, der den WHEP-Pfad
(`http://localhost:8892/webrtc-test/whep` aus dem Lab) konsumiert
und die Player-Events analog zum hls.js-Adapter in den Player-SDK-
Event-Stream einspeist.

DoD:

- [x] WebRTC-Adapter-Implementation in
  `packages/player-sdk/src/adapters/webrtc/adapter.ts`: baut eine
  `RTCPeerConnection` mit `addTransceiver("video"|"audio",
  recvonly)`, signalisiert via WHEP (POST `application/sdp`), parst
  die SDP-Answer und montiert die empfangenen Tracks an das
  übergebene `<video>`-Element. `destroy()` schließt PeerConnection,
  abortet den WHEP-`fetch` und stoppt alle Tracks (idempotent).
  Browser-API-Abhängigkeit ist über `deps.PeerConnection` /
  `deps.fetch` testbar entkoppelt.
- [x] Player-Event-Stream konsistent: `playback_started` mit
  `webrtc.connection_state=connected` bei erfolgreichem Handshake;
  `playback_error` für jeden Fehlerpfad. **Kein** synthetisches
  `manifest_loaded`-Event — der WebRTC-Pfad nutzt
  `playback_started` als Handshake-Erfolgs-Signal. Wire-Format
  bleibt unverändert (Tranche-1-Contract-Entscheidung); Tranche 3
  zieht den `webrtc.*`-Namespace produktiv.
- [x] WebRTC-Fehlercode-Taxonomie in
  `packages/player-sdk/src/adapters/webrtc/error-codes.ts`:
  Mindestcodes (`whep_signaling_failed`, `whep_sdp_invalid`,
  `webrtc_no_tracks`, `peer_connection_failed`,
  `webrtc_destroyed_before_connected`) als `WebRtcErrorCode`-Type-
  Union plus `WEBRTC_ERROR_CODES`-Allowlist. Reservierter Meta-Key:
  `webrtc.error_code` (`WEBRTC_ERROR_CODE_META_KEY`). Codes stehen
  immer in `playback_error.meta` unter diesem Key — maschinenlesbar
  für Dashboard/Releasing.
- [x] Reservierter Meta-Key validiert: `normalizeWebRtcErrorCode()`
  bildet jeden Nicht-Allowlist-Wert auf den Fallback
  `peer_connection_failed` ab (auch `undefined`/`null`/Numbers/
  unbekannte Strings); freie Strings können nicht durch das Surface
  durchschlagen. Tests decken gültige Codes, ungültigen Code,
  fehlenden Code und die `isWebRtcErrorCode`-Type-Guard-Variante ab.
- [x] WHEP-Fehlerpfade abgenommen: nicht-2xx Signalisierung
  (`whep_signaling_failed`), Antwort ohne SDP-Header
  (`whep_sdp_invalid`), `connectionstatechange=failed`
  (`peer_connection_failed`), `destroy()` vor Handshake-Abschluss
  (`webrtc_destroyed_before_connected`). Alle vier Pfade sind in
  `tests/webrtc-adapter.test.ts` mit gemocktem
  `RTCPeerConnection` + `fetch` verifiziert; `destroy()` ist
  idempotent.
- [x] `apps/dashboard` `/demo-webrtc`-Route in
  `src/routes/demo-webrtc/+page.svelte` (analog `/demo`-Route);
  liest `PUBLIC_WHEP_URL` aus `.env` (Default
  `http://localhost:8892/webrtc-test/whep` für das Lab).
  Bestehende `/demo`-Route (hls.js) bleibt unverändert; SDK-
  Imports parallel.
- [x] Lokal verifiziert: SDK-Tests (`pnpm --filter
  @npm9912/player-sdk run test`) 91/91 grün; `make gates` grün
  (Adapter, Demo-Route + alle anderen 0.7.0-Pfade unverändert).
  Browser-Live-Handcheck (`make dev` + `mtrace-webrtc`-Stack +
  Browser auf <http://localhost:5173/demo-webrtc?autostart=1>) ist
  Operator-Pflicht-Schritt im `0.8.0`-Closeout (Tranche 5).

---

## 4. Tranche 3 — Produktive WebRTC-Telemetrie

Bezug: `spec/telemetry-model.md` §3.2/§3.5;
`docs/planning/open/risks-backlog.md` R-12;
`scripts/smoke-observability.sh`.

Ziel: WebRTC-Telemetrie ist auf der bounded Allowlist aktiv;
`mtrace_webrtc_*`-Counter im API-Ingress; R-12 ist ab dieser
Tranche release-blockierend.

DoD:

- [x] WebRTC-Adapter sammelt `getStats()`-Reports gemäß §3.5.2-Muss-
  Feldern: `packages/player-sdk/src/adapters/webrtc/sampling.ts`
  iteriert das `RTCStatsReport` und extrahiert `transport.dtlsState`,
  `candidate-pair.state` (mit nominated/selected-Bevorzugung) sowie
  `inbound-rtp`/`outbound-rtp`-Counter (`packetsLost`,
  `bytesReceived`, `bytesSent`). Per-Identifier-Felder gehen nicht
  ins Wire-Format; nur die in §1.4 / §3.2 freigegebenen bounded
  Aggregat-Labels werden gesendet (Phase A, Commit `9e0d561`).
- [x] WebRTC-Wire-Vertrag normativ gepinnt: `spec/telemetry-model.md`
  §1.4 (webrtc.*-Tabelle), `spec/backend-api-contract.md` §3.4a,
  `contracts/event-schema.json` (`reserved_meta_keys` plus
  `reserved_meta_namespace_webrtc`-Block), `contracts/sdk-compat.json`
  (`reserved_meta_namespaces`-Liste) — alle vier in Phase A
  zusammen mit dem SDK-Sampling committet.
- [x] API-Validierung hart: `apps/api/hexagon/application/event_meta_validation.go`
  prüft jeden `webrtc.*`-Key gegen die Allowlist. Unbekannte Keys,
  falsche Typen, ungültige Enum-Werte, negative Counter-Werte,
  Pattern-Verletzung bei `peer_connection_run_id` und Forbidden-
  Identifier (`webrtc.track_id`, `webrtc.candidate_pair_id`,
  `webrtc.ssrc`, `webrtc.user_agent`, weitere) liefern
  `domain.ErrInvalidEvent` → HTTP 422; die `WebRTCSample`-Counter
  werden nicht aufgerufen. Tests in
  `event_meta_validation_internal_test.go` decken alle Fälle aus
  dem Plan-DoD plus weitere Edge-Cases ab.
- [x] Reservierter Meta-Namespace dokumentiert in
  `contracts/event-schema.json#reserved_meta_namespace_webrtc` und
  `spec/telemetry-model.md` §1.4. Strikte Allowlist nur für
  `webrtc.*`-Keys; nicht-reservierte Meta-Keys (z. B. `future_marker`,
  `experimental`) bleiben Forward-Compatibility-konform akzeptiert
  (Test `TestValidateReservedEventMeta_ForwardCompatibility` ist
  unverändert grün).
- [x] `mtrace_webrtc_*`-Counter exportiert:
  `apps/api/adapters/driven/metrics/webrtc_metrics.go` registriert
  drei State-CounterVec (`connection_state_total{connection_state}`,
  `ice_state_total{ice_state}`, `dtls_state_total{dtls_state}`)
  plus drei label-freie Delta-Counter (`packets_lost_total`,
  `bytes_received_total`, `bytes_sent_total`); Wertebereiche aus
  §3.2 sind hartcodiert in der Validation, Smoke prüft Allowlist
  defense-in-depth.
- [x] §3.5.2-Muss-Felder als Counter abgenommen: `packetsLost`,
  `bytesReceived`, `bytesSent` werden pro PeerConnection summiert
  (Adapter-Side) und serverseitig deltadiffenziert. Andere
  Soll-Felder (`jitter`, `roundTripTime`, …) bleiben für eine
  spätere Histogram-Tranche; `spec/telemetry-model.md` §3.5.2
  markiert sie weiterhin als „bei Verfügbarkeit".
- [x] Counter-Semantik dokumentiert in `spec/telemetry-model.md`
  §3.5.1 und implementiert in `webrtc_metrics.go`: State-Counter
  zählen Samples; Delta-Counter aus `(project_id, session_id,
  peer_connection_run_id, metric)`-State-Map mit Sample-ID-Idempotenz
  (Duplicate/Retry-Samples mit `sample_id ≤ last_sample_id`
  inkrementieren keinen Counter); Reconnect mit neuer
  `peer_connection_run_id` startet eine eigene Baseline; negative
  Deltas bleiben aus dem Counter. In-memory-State, kein durable
  Persist — Baseline-Reset nach API-Restart ist explizites
  Verhalten (§3.5.1 Punkt 2). Tests in
  `webrtc_metrics_test.go` decken Baseline, positive Delta,
  negativer Delta, Duplicate, Reconnect, State-Counter-Independence
  und No-Forbidden-Labels.
- [x] `scripts/smoke-observability.sh` erweitert:
  WebRTC-Forbidden-Liste (peer_connection_run_id, ssrc, track_id,
  candidate_pair_id, transport_id, local/remote_candidate_id,
  mime_type) plus vier Self-Tests; eine zweite WebRTC-Allowlist-
  Sektion prüft (sobald `mtrace_webrtc_*`-Series da sind), dass die
  State-Counter nur ihr State-Label tragen und Byte-/Loss-Counter
  fachlich label-frei sind.
- [x] R-12 angehoben: `risks-backlog.md` Status-Header auf
  „release-blockierend ab nächstem Browser-Major-Bump" gesetzt;
  Eintrag verweist auf den ausgelieferten produktiven Pfad
  (Adapter-Sammlung + Wire-Validation + `mtrace_webrtc_*`); konkrete
  Drift-Smoke-Spezifikation bleibt Folgeplan.
- [x] Schema-Drift-Strategie aus §3.5.3 im Adapter-Code:
  `collectAggregate()` in `sampling.ts` gibt `null` zurück, wenn
  ein Muss-Feld (`dtlsState`/`iceState`/connection-State, oder
  weder inbound noch outbound RTP-Daten) fehlt — kein
  `unknown`-Surrogat. Soll-Felder sind in dieser Tranche nicht
  emittiert (würden von API als unbekannte `webrtc.*`-Keys 422
  abweisen); ein Folgeplan kann sie additiv aufnehmen.

---

## 5. Tranche 4 — Compat-Tests + Browser-Support-Matrix

Bezug: `done/plan-0.2.0.md` §5 (Browser-Support-Matrix); Tranche 1–3.

Ziel: Adapter-Auswahl ist gegen Chromium 120+ und Firefox 120+
nachweislich kompatibel; Pack-Smoke aus `0.2.0` ist erweitert;
SDK-Performance-Budget ist eingehalten.

DoD:

- [x] Browser-E2E-Smoke
  `tests/e2e/dashboard-demo-webrtc.spec.ts` (neu, Playwright):
  rendert `/demo-webrtc?autostart=1`, pollt das Session-Detail
  über `GET /api/stream-sessions/{id}` und verifiziert mindestens
  ein Event mit `webrtc.peer_connection_run_id` in der Meta. Im
  Default-Mode (CI ohne `mtrace-webrtc`) erwartet der Test ein
  `playback_error` mit `webrtc.error_code`; ENV
  `MTRACE_WEBRTC_LAB=1` flippt die Assertion auf `playback_started`.
  Lab-abhängig im Sinne des Happy-Path; Fehlerpfad ist auch ohne
  Lab E2E-getestet.
- [x] Pack-Smoke (`packages/player-sdk/scripts/pack-smoke.mjs`)
  erweitert: prüft im ESM-, CJS- und IIFE-Entry-Pfad, dass
  `attachWebRtc` als Funktion exportiert ist; zusätzlich validiert
  der Smoke, dass `dist/index.d.ts` die TypeScript-Symbole
  `attachWebRtc`, `WebRtcAdapter` und `WebRtcAdapterOptions`
  deklariert (Public-API-Type-Vertrag).
- [x] Browser-Support-Matrix in `packages/player-sdk/README.md`
  erweitert: dedizierte WebRTC-Adapter-Tabelle mit Chromium 120+
  (Required), Firefox 120+ (Required), Safari 17+ (Best-effort,
  Schema-Drift-Hinweis) und Out-of-Scope-Klausel für mobile
  WebViews. Verlinkt auf `spec/telemetry-model.md` §3.5.2/§3.5.3.
- [x] `make sdk-performance-smoke` verifiziert das Bundle-Size-
  Budget (< 30 KiB gzip ESM, additiv inkl. WebRTC-Adapter) und
  importiert `attachWebRtc` aus dem Bundle (Smoke-Assert: ist als
  Funktion verfügbar). Der Adapter sprengt das Budget nicht; CI-
  Pipeline misst gegen den unveränderten Schwellwert aus `0.2.0`.
- [x] CI-Policy in `packages/player-sdk/README.md` §Performance and
  Browser Support dokumentiert: SDK-Unit-Tests, Public-API-Snapshot
  (`check-public-api.mjs`), Pack-Smoke (ESM+CJS+IIFE+`dist/index.d.ts`)
  und Performance-Smoke sind release-blockierend (laufen in
  `make gates` und `make sdk-performance-smoke`); Browser-E2E
  bleibt opt-in/lab-abhängig.

---

## 6. Tranche 5 — Release-Doku, RAK-Matrix und Closeout

Bezug: RAK-51 (hochgezogen), RAK-52..RAK-55 (neu);
`docs/user/releasing.md`; `README.md`; `roadmap.md`.

Ziel: `0.8.0` ist auffindbar dokumentiert, Versions-Bump
durchgezogen, Tag `v0.8.0` gesetzt.

DoD:

- [x] `README.md` Status-Block auf „`0.8.0` released — Player-SDK-
  WebRTC-Adapter und produktive WebRTC-Telemetrie" umgestellt;
  `docs/user/local-development.md` §2.7 verweist auf
  `apps/dashboard` `/demo-webrtc`-Route plus
  `packages/player-sdk/README.md` für die Adapter-Auswahl.
- [x] `docs/user/releasing.md` neue §2.3 für `0.8.0` mit dem
  manuellen Browser-Handcheck-Skript: `make dev` plus
  `mtrace-webrtc`-Compose hochfahren, `/demo-webrtc?autostart=1`
  öffnen, prüfen dass das Video spielt, `chrome://webrtc-internals`
  bzw. `about:webrtc` zeigt `connection_state=connected`,
  `ice_state` in `connected|completed`, `dtls_state=connected`,
  und `GET /api/metrics` listet `mtrace_webrtc_*`-Counter ohne
  Forbidden-Labels. Browser-E2E-opt-in via
  `MTRACE_WEBRTC_LAB=1 make browser-e2e`.
- [x] RAK-Verifikationsmatrix §6.1 vollständig mit Commit-Hashes
  ausgefüllt; RAK-55 (Kann) ist mit dem opt-in Browser-E2E
  abgenommen.
- [x] Versions-Bump 0.7.0 → 0.8.0 durchgezogen: 5 package.json
  (root + 2 apps + 2 packages), `apps/api/cmd/api/main.go`
  `serviceVersion`, `packages/player-sdk/src/version.ts`,
  `packages/player-sdk/scripts/pack-smoke.mjs` `expectedVersion`,
  `contracts/sdk-compat.json` `sdk_version`, alle Test-Fixtures
  mit hartkodierten SDK-/Analyzer-Versions-Strings (Bulk-Fix per
  `xargs sed -i 's/"0\.7\.0"/"0.8.0"/g'` über die `_test.go`/
  `.test.ts`-Files plus `spec/contract-fixtures/analyzer/*.json`
  und `apps/api/.../testdata/contract-*.json`). Hartkodierter
  Tarball-Pfad in `packages/player-sdk/package.json` Script
  `pack:smoke` explizit nachgezogen.
- [x] CHANGELOG: [Unreleased]-Block in `[0.8.0] - 2026-05-06`
  umgewandelt mit Release-Header; neuer leerer [Unreleased]-Block
  obenauf vorbereitet.
- [x] `./scripts/verify-doc-refs.sh` (`make docs-check`) grün vor
  Closeout-Commit; `make gates` grün im Release-Gate-Fix nach dem
  Closeout.
- [x] `plan-0.8.0.md` von `docs/planning/in-progress/` nach
  `docs/planning/done/` verschoben (`git mv` im Closeout-Commit);
  alle relativen Cross-Refs angepasst (analog `0.7.0` Closeout in
  Commit `11a3368`); Roadmap §1.2/§3 zeigen `0.8.0` ✅.
- [x] Tag `v0.8.0` annotiert; Push opt-in (User-Bestätigung im
  Closeout-/Release-Gate-Workflow); GitHub-Release mit CHANGELOG-
  `[0.8.0]`-Block als Notes-Body.

### 6.1 RAK-Verifikationsmatrix

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-51 | Muss | `packages/player-sdk/src/adapters/webrtc/adapter.ts` exposed produktiven WebRTC-Adapter (`attachWebRtc`/`WebRtcAdapter`/`WebRtcAdapterOptions`); hls.js-Pfad unverändert (`tests/hlsjs-adapter.test.ts` grün). Public-API-Snapshot in `scripts/public-api.snapshot.txt` aktualisiert. Commits `f9e17c2` (Tranche 1, Vertrag), `d11fc7f` (Tranche 2, WHEP-Implementation). | ✅ |
| RAK-52 | Muss | Adapter-Auswahl dokumentiert in `packages/player-sdk/README.md` §Public-API + neuer Adapter-selection-Block; Pack-Smoke prüft `attachWebRtc` in ESM/CJS/IIFE plus `dist/index.d.ts` für `WebRtcAdapter`/`WebRtcAdapterOptions`-Types; Browser-Support-Matrix mit Chromium 120+ Required, Firefox 120+ Required, Safari 17+ Best-effort. Commit `c49c81a` (Tranche 4). | ✅ |
| RAK-53 | Soll | `apps/api/adapters/driven/metrics/webrtc_metrics.go` exportiert sechs `mtrace_webrtc_*`-Counter (drei State-Counter mit bounded Labels + drei label-freie Delta-Counter); `scripts/smoke-observability.sh` spiegelt §3.1 WebRTC-Forbidden (peer_connection_run_id, ssrc, track_id, …) und prüft die §3.2-Allowlist gegen aktive `mtrace_webrtc_*`-Series. Commit `55f965b` (Tranche 3 Phase B). | ✅ |
| RAK-54 | Soll | `packages/player-sdk/src/adapters/webrtc/sampling.ts` sammelt `getStats()`-Reports und extrahiert §3.5.2-Muss-Felder (`connectionState`, `dtlsState`, ICE-State-Aggregat, `packetsLost`, `bytesReceived`, `bytesSent`); Wire-Vertrag in `spec/telemetry-model.md` §1.4/§3.5 + `contracts/event-schema.json` `reserved_meta_keys`-Block; Schema-Drift-Strategie aus §3.5.3 als „Muss-Feld fehlt → kein Sample" implementiert (kein unknown-Surrogat). R-12 ist im `risks-backlog.md` auf release-blockierend angehoben. Commits `9e0d561` (Tranche 3 Phase A) + `55f965b` (Phase B). | ✅ |
| RAK-55 | Kann | `tests/e2e/dashboard-demo-webrtc.spec.ts` (Playwright) prüft `/demo-webrtc?autostart=1` gegen die Session-Timeline; Default-Mode validiert Error-Pfad (`whep_signaling_failed`), `MTRACE_WEBRTC_LAB=1` flippt auf Happy-Path. Commit `c49c81a` (Tranche 4). | ✅ |

---

## 7. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen (analog `done/plan-0.7.0.md` §7).
- Lastenheft-Patch `1.1.10` (siehe §0.2) ist Vorgänger-Gate für
  Tranche 1. Bis Tranche 0 abgeschlossen ist, sind RAK-52..RAK-55
  in diesem Dokument nur geplante RAK aus dem Patch-Vorschlag; beim
  Tranche-0-Closeout müssen alle Referenzen auf die finalen
  Lastenheft-IDs und Prioritäten abgeglichen werden.
- Wenn ein `0.8.0`-Item in einer Folge-Phase neu bewertet wird
  (z. B. „Cross-Browser-Auto-Failover" doch nötig), entweder
  Folgeplan eröffnen oder hier als Wartungs-Eintrag vermerken.
- R-12 (WebRTC-`getStats()`-Schema-Drift) wechselt mit Tranche 3
  von „Triggerschwelle nicht ausgelöst" auf „release-blockierend";
  Risiken-Backlog-Eintrag muss im selben Commit nachgezogen
  werden.
