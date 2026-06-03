# Implementation Plan — `0.22.3` (WebRTC-Drift-Firefox-Folge-Patch)

> **Status**: in Arbeit; offen.
>
> **Vorgänger**: `0.22.2` ist als Go-Stdlib-Security-Patch released
> und archiviert in [`../done/plan-0.22.2.md`](../done/plan-0.22.2.md).
>
> **Auslöser**: Nightly-`webrtc-drift.yml` Run
> [`26858728018`](https://github.com/pt9912/m-trace/actions/runs/26858728018)
> vom 2026-06-03 ist auf **Firefox** failed (Chromium passt nach wie
> vor). Bei `0.22.2`-Closeout als Beobachtung dokumentiert
> ([plan-0.22.2 §6](../done/plan-0.22.2.md#6-beobachtung-webrtc-drift-firefox-failure-separates-folge-item)),
> jetzt eigener Folge-Plan.
>
> **Assertion-Pfad** (`tests/e2e/webrtc-stats-drift.spec.ts:239`):
> ```ts
> expect(payload.connectionState).toBe("connected");
> ```
> liefert in Firefox `payload.connectionState = "disconnected"`. Vorher
> neun aufeinanderfolgende Nightly-Läufe (2026-05-25 .. 2026-06-02)
> waren in beiden Browsern grün; der Failure-Wechsel auf 2026-06-03
> spricht für einen Firefox-Major- oder Playwright-Bundle-Drift, nicht
> für eine Test-Regression.
>
> **Release-Typ**: Wenn die Ursache ein Sampling-/Allowlist-Fix in
> `packages/player-sdk/src/adapters/webrtc/` ist → Patch-Release `0.22.3`
> (Versionstragend, weil SDK-Artefakt). Wenn rein Test-/Spec-Doku-Fix
> ohne Code-Pfad-Änderung → kein Versions-Bump nötig, dann Roll-up in
> den nächsten Release ohne eigenen Tag.

## 0. Scope

In Scope:

- Lokale Reproduktion gegen `mtrace-webrtc`-Lab mit Firefox aus dem
  Playwright-Bundle (`mcr.microsoft.com/playwright:v1.59.1-noble`):
  ist `connectionState=disconnected` ein dauerhafter Zustand oder ein
  Timing-Artefakt (Firefox braucht länger bis `connected`)?
- Differenzierung zwischen:
  - **Test-Timing-Fix**: Polling/Wait-for-`connected`-Schleife mit
    Bounded-Retry erweitern, bevor die finale Assertion läuft. Aktueller
    Pfad assertet direkt nach dem WHEP-Handshake.
  - **Spec-Allowlist-Update**: §1.4 `connection_state`-Allowlist enthält
    bereits `disconnected`; die strikte `toBe("connected")`-Assertion auf
    Zeile 239 könnte zur Schwellwert-Logik („§1.4-Allowlist UND nicht
    `failed`/`closed` UND Mehrheit der Samples = `connected` über ein
    Beobachtungsfenster") werden.
  - **Real-Sampling-Bug**: Firefox liefert eine getStats-Generation, in
    der das `RTCPeerConnection.connectionState`-Property nach
    erfolgreichem Handshake transient als `disconnected` erscheint —
    dann Sampling-Pfad in
    `packages/player-sdk/src/adapters/webrtc/sampling.ts` anpassen
    (debounce/aggregation), nicht nur den Test.
- Dokumentation des Befunds in
  `spec/telemetry-model.md` §3.5.2/§3.5.3 (Soll- vs. Muss-Spalte) und
  in `docs/dev/webrtc-drift.md` (Operator-Doku, falls vorhanden — sonst
  Verweis aus `extra-gates.md` ergänzen).

Nicht in Scope:

- Safari/WebKit-Pfad (bleibt opt-in pro `MTRACE_WEBRTC_DRIFT_BROWSERS`).
- Browser-Major-Wechsel im Playwright-Bundle (`v1.59.1-noble` ist die
  aktuelle Pinning-Stelle; Bundle-Bump ist eigener Folge-PR).
- Erweiterungen am `webrtc.*`-Wire-Vertrag jenseits des
  `connection_state`-Soll-Verhaltens.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 1 | Lokale Reproduktion + Charakterisierung (Noise vs. konsistent) | Drei aufeinanderfolgende lokale Firefox-Läufe mit Aufzeichnung von `connectionState`-Transitions |
| 2 | Root-Cause-Entscheidung + Fix-Pfad | Eine der drei Optionen oben gewählt mit Begründung |
| 3 | Release-Closeout (falls Versions-Bump nötig) | Versions-Sweep auf `0.22.3`, Plan archiviert, Tag/Release |

## 2. Tranche 1 — Reproduktion + Charakterisierung

DoD:

- [ ] `make smoke-webrtc-stats-drift MTRACE_WEBRTC_DRIFT_BROWSERS=firefox`
  dreimal hintereinander lokal ausgeführt; Ergebnis-Matrix
  (pass/fail × `payload.connectionState`-Wert) dokumentiert.
- [ ] Aufzeichnung der `connectionState`-Transitions über die
  WHEP-Handshake-Phase (Polling vor finaler Assertion oder
  `pc.onconnectionstatechange`-Hook im Test-Setup), um Timing-Charakter
  abzuschätzen.
- [ ] Quervergleich Chromium-Verhalten im selben Run-Stack: liefert
  Chromium konsistent `connected`, ist der Firefox-Drift Browser-
  spezifisch, nicht Lab-spezifisch.
- [ ] Ergebnis in Plan §1 als Tranchen-Übersicht-Spalte eintragen
  (Noise vs. reproduzierbar konsistent).

## 3. Tranche 2 — Root-Cause + Fix

DoD:

- [ ] Eine der drei Fix-Optionen gewählt mit kurzer Begründung:
  - **Test-Timing**: `tests/e2e/webrtc-stats-drift.spec.ts` führt vor
    der `toBe("connected")`-Assertion eine Bounded-Retry-Schleife
    (max. 5 s, 200 ms Intervall) auf
    `pc.connectionState === "connected"`. Reicht, wenn der Firefox-
    Drift ein verlängerter Handshake ist.
  - **Spec/Test-Lockerung**: Assertion auf §1.4-Allowlist-Membership
    plus „nicht `failed`/`closed`" senken; Soll-Erwartung
    `connection_state=connected` bleibt in `§3.5.3` als Soll-, nicht
    Muss-Vertrag. Reicht, wenn Firefox-Drift transient ist.
  - **Sampling-Pfad**: `packages/player-sdk/src/adapters/webrtc/
    sampling.ts` debouncet `connectionState`-Transitions
    (z. B. drei aufeinanderfolgende Samples mit gleichem Wert vor
    Auslieferung). Reicht, wenn Firefox einen echten Flapping-Zustand
    produziert.
- [ ] Implementierung des gewählten Fix-Pfads inkl. lokaler
  Reproduktions-Verifikation (drei grüne Firefox-Läufe in Folge).
- [ ] Falls Sampling-Pfad: `make analyzer-benchmark-smoke` und
  `make ts-test` grün, weil der Pfad in der Player-SDK-Closure
  liegt.
- [ ] Falls Spec-Anpassung: Wire-Vertrag in `spec/telemetry-model.md`
  §1.4 + §3.5.2 + §3.5.3 synchron geschärft (Allowlist bleibt
  unverändert, Soll-/Muss-Spalte präzisiert).
- [ ] Drift-Smoke-Auto-Issue-Konfiguration (`DRIFT_AUTO_ISSUE=1`
  im Workflow) erwägen — bei diesem Lauf war `DRIFT_AUTO_ISSUE`
  leer, deshalb wurde kein Issue eröffnet.

## 4. Tranche 3 — Release-Closeout (falls Versions-Bump nötig)

Versionstragend nur, wenn Player-SDK-Code (`packages/player-sdk/src/`)
oder Wire-Vertrag (`contracts/event-schema.json` /
`spec/telemetry-model.md`) geändert wird. Reiner Test-Fix in
`tests/e2e/` ohne SDK-/Wire-Änderung ist kein Patch-Release-
Auslöser — landet im nächsten regulären Release ohne eigenen Tag.

DoD (falls Versions-Bump):

- [ ] Versions-Sweep `0.22.2` → `0.22.3` nach `releasing.md` §3.1
  (5× `package.json`, `serviceVersion`, `PLAYER_SDK_VERSION`,
  `sdk_version`, 21 Analyzer-Fixtures, 20 testdata, hartkodierte
  Test-Strings).
- [ ] `CHANGELOG.md`: `[0.22.3]`-Block mit Firefox-Drift-Begründung
  und Verweis auf Nightly-Run `26858728018`.
- [ ] `roadmap.md` §1.1/§1.2/§3 nach Bump aktualisiert.
- [ ] `make gates` lokal grün (post-commit, wegen
  `generated-drift-check`-Mechanik aus `0.22.2`).
- [ ] `MTRACE_RELEASE_APPROVED=1 make release-guard VER=0.22.3`.
- [ ] Plan nach `done/plan-0.22.3-webrtc-drift.md` verschoben.
- [ ] Tag `v0.22.3` + GitHub-Release.

## 5. Wave-2-Quality-Gates-Verdict (Tag-Annotation-Vorbereitung)

Vor dem Tag (falls Versions-Bump) — Run-IDs sammeln analog zu
`plan-0.22.2.md` §5. Aktuelle Stand-Aufnahme zum Zeitpunkt des
Plan-Anlegens:

| Gate | letzte Run-ID (2026-06-03) | Status |
| --- | --- | --- |
| Benchmark regression | `26859227837` | success |
| Fuzz nightly | `26859493259` | success |
| Mutation nightly | `26859551494` | success |
| WebRTC drift smoke | `26858728018` | **failure** (dieser Plan adressiert es) |
| Security audit nightly | nach `0.22.2`-Push erwartet | folgt |

## 6. Offene Fragen

- Wann genau wechselte Firefox-`connectionState`-Verhalten? Diff
  zwischen Run `26792928448` (2026-06-02, grün) und
  `26858728018` (2026-06-03, rot) ist nur 24 h auseinander; gleicher
  Playwright-Bundle-Pin (`v1.59.1-noble`). Hypothese: neuere Firefox-
  Build aus dem Playwright-Image (über Image-Digest-Drift trotz
  Tag-Pin) oder MediaMTX-WHEP-Server-Verhalten.
- Sollte `DRIFT_AUTO_ISSUE` per Default im Workflow auf `1` stehen?
  Aktuell wird der Failure nur in der Workflow-Liste sichtbar,
  ohne automatisches GitHub-Issue.
