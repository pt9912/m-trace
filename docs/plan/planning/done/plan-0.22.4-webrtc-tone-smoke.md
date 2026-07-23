# Implementation Plan — `0.22.4` (WebRTC-Ton-Smoke)

> **Status**: ✅ abgeschlossen am 2026-06-16 als **Test-/CI-only-Fix** —
> kein SDK-/Wire-/Runtime-Touch. Per derselben Regel wie
> [`plan-0.22.3-webrtc-drift.md`](./plan-0.22.3-webrtc-drift.md) §4
> rollt der Smoke in die nächste versionstragende Lieferung; ein
> eigener `v0.22.4`-Tag ist optional (Maintainer-Entscheidung am
> Release-Gate), nicht durch diesen Smoke erzwungen.
>
> **Vorgänger**: `0.22.3` Security-/CI-Sammel-Patch
> ([`CHANGELOG.md`](../../../../CHANGELOG.md) `[0.22.3]`).
>
> **Auslöser**: Review der manuellen Release-Abnahmen in
> [`docs/user/releasing.md`](../../../user/releasing.md) §2.3 — der Schritt
> „bestätigen, dass ein 1-kHz-Sinuston hörbar abspielt" war als rein
> menschliche Hörprüfung definiert. Die *enge* Aussage „ein sauberer
> 1-kHz-Ton liegt an" ist maschinell präziser prüfbar als per Ohr;
> der WebRTC-`getStats()`-Drift-Smoke (RAK-56) prüft nur
> `bytesReceived>0` (Medien fließen), nicht die Tonqualität.

## 0. Scope

In Scope:

- Dependency-freier Einzel-Ton-Detektor `scripts/check-tone.mjs`
  (Goertzel-Einzel-Bin-DFT): liest Mono-`f32le`-PCM von stdin, prüft
  per Energie-Anteil (Parseval) des Ziel-Bands an der Gesamtenergie,
  ob ein sauberer Ton an der Zielfrequenz dominiert.
- Lab-Smoke `scripts/smoke-webrtc-tone.sh` + `make smoke-webrtc-tone`:
  fährt das `mtrace-webrtc`-Lab hoch, zieht den RTSP-Egress
  (`rtsp://mediamtx:8554/webrtc-test`, Opus → Mono-48k-`f32le`) per
  ffmpeg im Lab-Netz und pipet das PCM an den Detektor.
- Nightly-Integration in
  [`.github/workflows/webrtc-drift.yml`](../../../../.github/workflows/webrtc-drift.yml)
  als zusätzlicher, nicht-blockierender Schritt im selben Job.
- Doku: Gate-Eintrag in
  [`extra-gates.md`](../extra-gates.md) und Verweis aus
  [`releasing.md`](../../../user/releasing.md) §2.3.

Nicht in Scope:

- Browser-Decode-Pfad / Web-Audio-Capture in Playwright: bewusst
  vermieden (headless-WebRTC-Audio-Capture ist die brittle Variante).
  Der RTSP-Egress-Pull validiert die Pipeline bis MediaMTX; der
  Browser-Pfad bleibt über den getStats-Drift-Smoke (`bytesReceived`)
  abgedeckt.
- Die **perzeptuelle** Operator-Abnahme („klingt/sieht das ganze Demo
  im echten Browser richtig"). Der Ton-Smoke deckt nur den eng
  definierten „1-kHz-Ton vorhanden und dominant"-Teil ab; die
  holistische Sicht-/Hörabnahme bleibt manuell (releasing.md §2.3).
- PR-blockierender Gate: das WebRTC-Lab ist unter Last flaky (siehe
  `plan-0.22.3-webrtc-drift` §2), daher opt-in lokal + Nightly statt
  `make gates`.
- ADR / neue RAK / Lastenheft-Patch: keine Architektur-Entscheidung,
  keine neue User-Surface; verfeinert die bestehende
  WebRTC-Verifikations-Familie (RAK-51..RAK-56). Patch-Releases führen
  keine RAK-Matrix (releasing.md §3.1).

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Ergebnis |
| --- | --- | --- |
| 1 | Detektor + deterministischer Selbsttest | `check-tone.mjs` (Goertzel + Energie-Anteil). Fünf synthetische ffmpeg-Fälle: reiner 1 kHz **PASS** (Anteil 0.500 = theoretische ½), 440 Hz **FAIL** (0.000), Stille **FAIL**, Rauschen **FAIL** (0.000), 1 kHz + Rauschen **PASS** (0.452). |
| 2 | Lab-Smoke + Make-Target | `smoke-webrtc-tone.sh` (ffmpeg-RTSP-Pull im Lab-Netz → Detektor), `make smoke-webrtc-tone`, opt-in. Live gegen `mtrace-webrtc` verifiziert. |
| 3 | Nightly + Doku | Schritt in `webrtc-drift.yml` (nicht-blockierend), Gate-Eintrag in `extra-gates.md`, Verweis aus `releasing.md` §2.3. |

## 2. Methodik — warum Goertzel + Energie-Anteil

`check-tone.mjs` rechnet die DFT-Bin-Leistung `|X_k|²` per Goertzel —
für eine *bekannte* Zielfrequenz genügt der eine Bin, keine FFT-Lib.
Diskriminierend ist nicht ein Off-Target-Verhältnis (eine erste
Variante lieferte für 440 Hz fälschlich PASS, weil bei ganzzahligen
Frequenzen über 1 s alle Off-Bins ~0 sind und das Verhältnis nur
Float-Rauschen vergleicht), sondern der **Energie-Anteil** des
Ziel-Bands an der Gesamtenergie (Parseval): ein reiner, auf dem Bin
liegender Ton konzentriert ~½ der Energie im `+f`-Bin (½ im
konjugierten `-f`-Bin), Anteil ~0.5; abwesende Frequenz oder
Breitbandrauschen ergibt ~0. Ein schmales ±`band-hz`-Band fängt
Leckage bei nicht-exakter Bin-Lage (reale, Opus-/resample-behaftete
Töne) ab. Verdict: `rms >= min-rms` UND `energy-fraction >= min-fraction`
(Default 0.2 — ~2,5× Marge unter dem Reinton-Wert).

## 3. DoD

- [x] `scripts/check-tone.mjs` — Goertzel + Energie-Anteil,
  dependency-frei, stdin-`f32le`.
- [x] Deterministischer Selbsttest (fünf synthetische ffmpeg-Signale,
  ohne WebRTC-Lab): Reinton/Mischton PASS, falsche Frequenz/Stille/
  Rauschen FAIL.
- [x] `scripts/smoke-webrtc-tone.sh` + `make smoke-webrtc-tone`
  (auto-up/down `mtrace-webrtc`, `SMOKE_WEBRTC_AUTOSTART=0` für
  laufenden Stack), live gegen das Lab verifiziert.
- [x] Nightly-Schritt in `webrtc-drift.yml` (nicht-blockierend,
  `continue-on-error`).
- [x] Gate-Eintrag in `extra-gates.md` (§2-Matrix + §3-Abschnitt) und
  Verweis aus `releasing.md` §2.3.
- [ ] Bestätigung im nächsten `webrtc-drift.yml`-Nightly: Ton-Schritt
  läuft grün neben dem Drift-Schritt.

## 4. Abgrenzung zur manuellen Abnahme

Der Smoke ersetzt **nicht** die menschliche Schlussabnahme. Ein grüner
FFT-Check beweist „1-kHz-Ton vorhanden und dominant" — nicht „das Demo
verhält sich im echten Browser richtig" oder „die Audiowiedergabe ist
subjektiv sauber". releasing.md §2.3 behält die perzeptuelle Abnahme;
der Smoke nimmt nur den eng definierten, automatisierbaren Teil ab und
liefert regelmäßige Nightly-Regressionserkennung.
