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
| 1 | Lab-Compose `examples/webrtc/compose.yaml` mit MediaMTX-WHIP/-WHEP plus optionalem `coturn` | ⬜ |
| 2 | README-Konkretisierung — Operator-Befehle, Port-Schnitt, Browser-Handcheck | ⬜ |
| 3 | `make smoke-webrtc-prep`-Target mit reservierter Vorbereitungs-Verifikation | ⬜ |
| 4 | WebRTC-Telemetrie-Bewertung — bounded Allowlist, `getStats()`-Subset, Schema-Drift-Strategie | ⬜ |
| 5 | Release-Doku, RAK-Matrix und Closeout | ⬜ |

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
- [ ] Media-Server-Image ist gepinnt (kein floating `latest`) und die
  README nennt die getestete WHIP-/WHEP-Pfadform des Images. Wenn sich
  MediaMTX-Endpunkte zwischen Versionen ändern, blockiert das Tranche 1
  bis zur Doku-/Smoke-Anpassung.
- [ ] Host-Port-Schnitt ist aufgelöst — der `0.5.0`-Skelett-Stand
  notiert einen geplanten Konflikt mit `examples/srt/` auf `8889/tcp`;
  Tranche 1 entscheidet, ob WebRTC einen anderen Host-Port bekommt
  oder SRT-Beispiel umschnitten wird.
- [ ] ICE-/Media-Port-Schnitt ist explizit entschieden: für MediaMTX-
  WebRTC im Container ist der UDP-Pfad (typisch `8189/udp`) inklusive
  Docker-Port-Mapping dokumentiert, oder ein bewusst getesteter TCP-/
  TURN-Fallback ist als RAK-50-Handcheck-Grenze beschrieben. Ein grüner
  WHIP-/WHEP-HTTP-Endpoint allein reicht nicht als Browser-Pfad-
  Nachweis.
- [ ] Host-Port-Schnitt ist in `examples/README.md` dokumentiert, falls
  die Project-Konvention oder Parallelbetriebshinweise angepasst werden
  müssen.
- [ ] FFmpeg-basierte Teststream-Mechanik ist entschieden und
  dokumentiert (F-84 Muss). GStreamer, MediaMTX-Testpublisher oder ein
  dedizierter Browser-Sender dürfen ergänzende Lab-/Handcheck-Pfade
  sein, ersetzen aber nicht den FFmpeg-Nachweis. Wenn ein Publisher nicht
  zuverlässig headless läuft, bleibt er manueller RAK-50-Handcheck und
  nicht Teil des Muss-Smokes.
- [ ] WHEP-Readiness hängt nicht von einem echten Browser ab: Der
  Stack stellt einen stabil prüfbaren Endpoint bereit, dessen
  erwarteter Status/Fehlercode dokumentiert ist (z. B. `OPTIONS`,
  `405`, `415`, `404 path missing` vs. `200` je nach Serververtrag).
- [ ] `examples/webrtc/README.md` „Start"-Sektion zeigt den
  `docker compose -p mtrace-webrtc -f examples/webrtc/compose.yaml
  up -d --build`-Befehl mit konkreten Ports.
- [ ] Stop/Reset ist auf das `mtrace-webrtc`-Projekt begrenzt und
  entfernt keine Core-/SRT-/DASH-Volumes.

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
- [ ] README benennt den exakten WHIP-/WHEP-URL-Satz inklusive Stream-
  Namen, Host-Port und Container-Port.
- [ ] „Verifikation" beschreibt einen Browser-Handcheck (manuell)
  und/oder den `make smoke-webrtc-prep`-Pfad, sobald Tranche 3
  liefert.
- [ ] „Stop / Reset" und „Troubleshooting" sind mit echten
  Fehlerbildern gefüllt (ICE-Negotiation, Codec-Negotiation,
  Headless-Restriktionen).
- [ ] „Bekannte Grenzen" wird auf das reduziert, was nach den
  Tranchen 1–3 nicht im Scope ist (typisch: produktive Telemetrie,
  Multi-Tenant, TLS-Setup).
- [ ] README erklärt klar, dass `make smoke-webrtc-prep` keine
  Playback-Qualität, keine ICE-Erfolgsquote und keine `getStats()`-
  Stabilität beweist.

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

- [ ] `scripts/smoke-webrtc-prep.sh` existiert, `set -euo pipefail`,
  `[smoke-webrtc-prep]`-Stderr-Präfix, Cleanup nur für
  `mtrace-webrtc`-Project-Name (Konvention aus
  `examples/README.md`).
- [ ] Smoke prüft endpoint-/compose-only: Compose-Stack ist
  hochgefahren, WebRTC-WHIP/-WHEP-Endpoint antwortet mit dem in
  Tranche 1 dokumentierten erwarteten Status/Fehlercode, ggf.
  STUN/TURN-Container läuft. Kein Browser, kein Playback,
  kein `getStats()`.
- [ ] Smoke unterscheidet Fehlerklassen: Compose nicht gestartet,
  Endpoint nicht erreichbar, falscher Status/Serververtrag, Port-
  Konflikt, optionaler TURN-Container down.
- [ ] Smoke nutzt bounded Waits und gibt bei Fehlern relevante
  Container-Logs oder Compose-Service-Status aus.
- [ ] `Makefile` listet `smoke-webrtc-prep` im `help`-Text und
  `.PHONY`; opt-in (nicht in `make gates`).
- [ ] Optional / nicht release-blockierend: Headless-Browser-
  Erweiterung (Browser-Handcheck-Automatisierung, getStats()-Smoke)
  ist separat bewertet — wenn reproduzierbar grün, ergänzt sich
  ein zusätzliches Target, sonst bleibt der Browser-Pfad bei
  Tranche 2 (manueller Handcheck). Diese Bewertung kann das
  Muss-Target `smoke-webrtc-prep` nicht aus dem Scope kippen.

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

- [ ] `spec/telemetry-model.md` §3.2 ist um eine WebRTC-Sub-Tabelle
  erweitert: erlaubte bounded Aggregat-Labels (z. B.
  `connection_state`, `ice_state`, `dtls_state`) mit festem
  Wertebereich.
- [ ] Unbounded WebRTC-/`getStats()`-Identifier bleiben als Prometheus-
  Labels verboten. Dazu zählen mindestens `source_id`, `id`/Report-ID,
  `peer_connection_id`, `track_id`, `transport_id`,
  `candidate_pair_id`, `local_candidate_id`, `remote_candidate_id`,
  `candidate_id`, `ssrc`, ICE-User-Fragmente, DTLS-/Zertifikats-
  Fingerprints, IP-Adresse, URL, Codec-String und Browser-User-Agent.
  Falls diese Werte im `getStats()`-Subset vorkommen, sind sie nur als
  Event-/Debug-Daten für einen späteren Read-Pfad zulässig; Prometheus
  bekommt ausschließlich die in Item 1 explizit erlaubten bounded
  Aggregat-Labels.
- [ ] `spec/telemetry-model.md` §3.2 dokumentiert das
  `getStats()`-Subset ausschließlich als Future-Telemetry-Notiz für
  bounded Aggregation: Report-Gruppen, Muss-/Soll-Felder und die
  Schema-Drift-Strategie zwischen Chromium-/Firefox-/Safari-Versionen.
  Diese Notiz ist kein Player-SDK-/Adapter-Public-API-Vertrag.
- [ ] Das `getStats()`-Subset benennt Fallback-Verhalten bei fehlenden
  Browser-Feldern. Eine Browser-Version ohne einzelnes Soll-Feld darf
  keinen späteren vollständigen Telemetriepfad blockieren, solange
  Muss-Felder stabil bleiben.
- [ ] Risiken-Backlog erweitert den Schema-Drift als
  **Spec-/Adapter-Review-Gate** und setzt dabei atomar die nächste freie
  Kennung (Stand Planerstellung: `R-11`, da `R-1`..`R-10` vergeben
  sind): bei Browser-Major-Version X mit `getStats()`-Schema-Änderung
  wird die WebRTC-Allowlist plus Future-Telemetry-Notiz in
  `spec/telemetry-model.md` §3.2 reviewed; konkrete Smoke-/Contract-
  Test-Updates sind erst dann release-blockierend, wenn ein produktiver
  WebRTC-Telemetrie-Pfad existiert (RAK-51 / Folge-Plan).
  Vor diesem Punkt ist `smoke-webrtc-prep` (endpoint-/compose-only) vom
  Schema-Drift nicht betroffen.
- [ ] Tranche 4 dokumentiert ausdrücklich, dass RAK-49 nur die
  Telemetrie-Spezifikation vorbereitet. Sie aktiviert keine produktive
  `getStats()`-Sammlung und keine Prometheus-Metrik in `0.7.0`, sofern
  nicht ein separater Codepfad geplant und abgenommen wird.
- [ ] `0.7.0` dokumentiert als negative Cardinality-Prüfung: im Release-
  Scope existiert kein produktiver `mtrace_webrtc_*`-Counter und kein
  WebRTC-Prometheus-Exportpfad. Die Erweiterung von
  `scripts/smoke-observability.sh` auf WebRTC-Allowlist-Labels ist ein
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
