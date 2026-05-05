# Implementation Plan — `0.7.0` (WebRTC-Lab-Erweiterung)

> **Status**: ⬜ geplant. Beginnt erst nach abgeschlossenem
> Release-Artefakt `v0.6.0` (SRT Health View). Liegt unter
> `docs/planning/open/`, bis das Vorgänger-Gate steht — dann nach
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
> [`plan-0.6.0.md`](./plan-0.6.0.md) (vorhergehende Phase).
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
Er beschreibt **nicht**, wann oder in welchem Release das ausgeliefert
wird — das ist eine Lastenheft-/Roadmap-Entscheidung außerhalb dieses
Dokuments.

### 0.1 Vorgänger-Gate

Voraussetzungen, bevor dieser Plan in `in-progress/` gezogen werden
kann:

- `0.5.0` ist released (Tag `v0.5.0` auf `a56dc0b`); WebRTC-Skelett
  unter [`examples/webrtc/`](../../../examples/webrtc/) existiert.
- `0.6.0` (SRT Health View) ist released; siehe
  [`plan-0.6.0.md`](./plan-0.6.0.md).
- Lastenheft `1.1.9` ist um RAK-47..RAK-51 für `0.7.0` erweitert (§13.9 + §4a.12 in `done/plan-0.1.0.md`); siehe §0.2.
- Headless-Browser-WebRTC-Stabilität in CI ist bewertet (siehe
  Tranche 3 unten); falls als zu instabil eingestuft, bleibt der
  Smoke aus dem Scope.

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
| RAK-51 | Kann | Player-SDK exposed einen optionalen WebRTC-Adapter-Pfad ohne Vermischung mit `hls.js`. |

Die DoD-Items in §1–§5 unten referenzieren diese RAKs verbindlich.

### 0.3 Out-of-Scope-Klauseln (durchgängig)

- Keine produktive `getStats()`-Sammlung im `apps/api`-Ingress, solange
  Schema-Drift zwischen Browser-Versionen (Chromium/Firefox/Safari)
  nicht durch eigene Tranche bewertet ist.
- Keine WebRTC-Aggregat-Metriken in Prometheus, bis bounded Allowlist-
  Labels in [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
  §3.2 erweitert sind.
- Keine Vermischung mit dem `hls.js`-Demo-Pfad in `apps/dashboard`;
  WebRTC bekommt eine eigene Demo-Route oder bleibt im Lab-Beispiel.

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 1 | Lab-Compose `examples/webrtc/compose.yaml` mit MediaMTX-WHIP/-WHEP plus optionalem `coturn` | ⬜ |
| 2 | README-Konkretisierung — Operator-Befehle, Port-Schnitt, Browser-Handcheck | ⬜ |
| 3 | `make smoke-webrtc-prep`-Target mit reservierter Vorbereitungs-Verifikation | ⬜ |
| 4 | WebRTC-Telemetrie-Bewertung — bounded Allowlist, `getStats()`-Subset, Schema-Drift-Strategie | ⬜ |

---

## 2. Tranche 1 — Lab-Compose

Bezug: Lastenheft §7.6 F-62, §7.8 F-82..F-84;
[`examples/webrtc/README.md`](../../../examples/webrtc/README.md)
„Voraussetzungen (geplant)".

Ziel: `examples/webrtc/compose.yaml` startet einen lokalen WebRTC-
WHIP-/WHEP-Pfad ohne Internet-Abhängigkeit. Project-Name
`mtrace-webrtc` (in `examples/README.md` reserviert).

DoD:

- [ ] `examples/webrtc/compose.yaml` definiert mindestens MediaMTX
  mit WHIP-/WHEP-Listener; optional `coturn`-Container für nicht-
  localhost-Pfade.
- [ ] Host-Port-Schnitt ist aufgelöst — der `0.5.0`-Skelett-Stand
  notiert einen geplanten Konflikt mit `examples/srt/` auf `8889/tcp`;
  Tranche 1 entscheidet, ob WebRTC einen anderen Host-Port bekommt
  oder SRT-Beispiel umschnitten wird.
- [ ] FFmpeg-Test-Publisher (analog `examples/srt/ffmpeg-srt-loop.sh`)
  schickt einen synthetischen WebRTC-Stream; oder ein dedizierter
  Browser-Sender im Lab-Setup.
- [ ] `examples/webrtc/README.md` „Start"-Sektion zeigt den
  `docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml
  up -d --build`-Befehl mit konkreten Ports.

---

## 3. Tranche 2 — README-Konkretisierung

Bezug: Tranche 1; [`examples/README.md`](../../../examples/README.md)
„README-Mindeststruktur".

Ziel: `examples/webrtc/README.md` ist auf den 7-Punkt-Standard der
anderen Beispiele umgestellt — keine Doku-only-Klauseln mehr,
sondern konkrete Operator-Befehle.

DoD:

- [ ] „Voraussetzungen" listet konkrete Tool-Versionen (Browser,
  ggf. STUN/TURN-Setup, Compose-Version).
- [ ] „Verifikation" beschreibt einen Browser-Handcheck (manuell)
  und/oder den `make smoke-webrtc-prep`-Pfad, sobald Tranche 3
  liefert.
- [ ] „Stop / Reset" und „Troubleshooting" sind mit echten
  Fehlerbildern gefüllt (ICE-Negotiation, Codec-Negotiation,
  Headless-Restriktionen).
- [ ] „Bekannte Grenzen" wird auf das reduziert, was nach den
  Tranchen 1–3 nicht im Scope ist (typisch: produktive Telemetrie,
  Multi-Tenant, TLS-Setup).

---

## 4. Tranche 3 — `make smoke-webrtc-prep`-Target

Bezug: Tranche 1; [`examples/README.md`](../../../examples/README.md)
„Smoke-Targets" (Target-Name reserviert).

Ziel: Ein opt-in Smoke-Target, das ausschließlich
**Vorbereitungsgrenzen** prüft — kein Playback-Qualitäts- oder
`getStats()`-Anspruch.

DoD:

- [ ] `scripts/smoke-webrtc-prep.sh` existiert, `set -euo pipefail`,
  `[smoke-webrtc-prep]`-Stderr-Präfix, Cleanup nur für
  `mtrace-webrtc`-Project-Name (Konvention aus
  `examples/README.md`).
- [ ] Smoke prüft: WebRTC-WHIP/-WHEP-Endpoint antwortet (HTTP-Status,
  kein Playback), Compose-Stack ist hochgefahren, ggf. STUN/TURN-
  Container läuft.
- [ ] Smoke-Stabilität in CI ist bewertet — wenn Headless-Browser-
  Pfad reproduzierbar grün läuft, Target wird ergänzt; sonst bleibt
  Tranche 3 mit `[!]` blockiert und Tranche 4 entscheidet, ob der
  Pfad ganz aus dem Scope fällt.
- [ ] `Makefile` listet `smoke-webrtc-prep` im `help`-Text und
  `.PHONY`; opt-in (nicht in `make gates`).

---

## 5. Tranche 4 — WebRTC-Telemetrie-Bewertung

Bezug: [`spec/telemetry-model.md`](../../../spec/telemetry-model.md)
§3.1/§3.2; Risiken-Backlog; Lastenheft §8.3 NF-14.

Ziel: Klare Spec-Aussage, was eine produktive WebRTC-Telemetrie-
Anbindung bräuchte — bounded Allowlist-Labels, `getStats()`-Subset,
Schema-Drift-Strategie. Diese Tranche kann **vor** Tranche 1–3
erledigt werden (Spec-Vorarbeit) oder nach den Lab-Erfahrungen.

DoD:

- [ ] `spec/telemetry-model.md` §3.2 ist um eine WebRTC-Sub-Tabelle
  erweitert: erlaubte bounded Aggregat-Labels (z. B.
  `connection_state`, `ice_state`, `dtls_state`) mit festem
  Wertebereich.
- [ ] `spec/player-sdk.md` (oder ein neues `spec/webrtc-adapter.md`)
  beschreibt das Subset von `getStats()`-Reports, das produktiv
  gesammelt werden soll, plus die Schema-Drift-Strategie zwischen
  Chromium-/Firefox-/Safari-Versionen.
- [ ] Risiken-Backlog erweitert (R-N) den Schema-Drift als
  release-blockierende Schwelle: „bei Browser-Major-Version X ändert
  sich `getStats()`-Schema, Smoke muss aktualisiert werden".
- [ ] Cardinality-Smoke (`scripts/smoke-observability.sh`) wird auf
  die neuen WebRTC-Allowlist-Labels erweitert — sobald irgendein
  produktiver `mtrace_webrtc_*`-Counter im Code steht.

---

## 6. Wartung

- Beim Auslagern eines `[ ]`-Items in einen Commit: `[ ]` → `[x]`,
  Commit-Hash anhängen.
- Lastenheft-Patch ist erledigt: `spec/lastenheft.md` steht auf
  `1.1.9` mit §13.9-Sektion „Version 0.7.0" und RAK-47..RAK-51;
  Eintrag als §4a.12 in
  [`done/plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c.
  Plan kann nach `docs/planning/in-progress/` ziehen, sobald
  Vorgänger-Gates aus §0.1 grün sind.
- Wenn der WebRTC-Pfad nach Bewertung als „nie produktiv"
  entschieden wird, dieses Dokument als historische Notiz nach
  `docs/planning/done/` ziehen oder löschen — je nachdem, ob die
  Aufzeichnung weiter Wert hat.
