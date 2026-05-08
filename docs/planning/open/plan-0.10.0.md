# Implementation Plan — `0.10.0` (CMAF-Analyse / NF-13)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist `0.9.6` (Lastenheft-Konvergenz + Repo-Artefakte);
> Aktivierung erst nach dessen Release-Closeout.
>
> **Release-Typ**: Minor-Release nach `0.9.6` mit Lastenheft-Patch
> `1.1.13`, neuer RAK-Gruppe `RAK-60`..`RAK-63`,
> RAK-Verifikationsmatrix und Tag `v0.10.0`.
>
> **Ziel**: NF-13 (`CMAF-Analyse`, Muss) wird in `0.10.0` bewusst nur
> als additive, manifestbasierte CMAF-Signal-Analyse im Stream Analyzer
> adressiert. `F-73` bleibt der historische Vorbereitungsschritt; NF-13
> wird mit diesem Plan **nicht vollständig geschlossen**, weil keine
> binäre CMAF-Konformitätsprüfung geliefert wird. Der Release macht
> CMAF-Indizien auditierbar und hält die spätere Segment-/Box-Analyse
> als Folge-Muss offen.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) F-73, NF-13,
> RAK-58, RAK-59;
> [`docs/user/stream-analyzer.md`](../../user/stream-analyzer.md)
> mit der Überschrift „CMAF";
> [`packages/stream-analyzer/`](../../../packages/stream-analyzer/);
> [`apps/api/hexagon/domain/stream_analysis.go`](../../../apps/api/hexagon/domain/stream_analysis.go).
>
> **Nachfolger**: Folge-Plan für binäre CMAF-/ISO-BMFF-Segment- und
> Box-Analyse zur vollständigen NF-13-Schließung.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Scope-Entscheidung oder fehlende Fixture.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.10.0` liefert **manifestbasierte CMAF-Signal-Analyse**, nicht
allgemeine ISO-BMFF-/MP4-Binäranalyse und keine vollständige
CMAF-Konformitätsprüfung.

In Scope:

- HLS/fMP4-CMAF-Erkennung über `EXT-X-MAP`, Segment-URI-Muster und
  manifestbasierte Konsistenzsignale.
- DASH/CMAF-Erkennung über `mimeType` `video/mp4`/`audio/mp4`/
  `application/mp4`, `SegmentTemplate`/`SegmentList`-Initialisierung
  und Representation-Metadaten.
- Additives Result-Schema für `details.cmaf` unter den bestehenden
  HLS- und DASH-Detail-Objekten, ohne bestehende HLS-/DASH-Felder zu
  brechen und ohne neues Top-Level-Feld im Analyzer/API-Envelope.
  Jedes Signal trägt eine Confidence (`manifest` oder `inferred`),
  damit manifestbasierte Indizien nicht als binäre
  Konformitätsaussage missverstanden werden. HLS-`unknown` mit
  `details:null` bleibt ohne `cmaf`.
- CLI/API-Durchleitung und Doku für die neuen CMAF-Signale.

Out of scope:

- Kein Download oder Parsing echter `.m4s`-/`.mp4`-Segmente.
- Keine Validierung von MP4-Boxen (`ftyp`, `moov`, `moof`, `traf`,
  `tfdt`, `sidx`) über Byte-Parsing.
- Keine Low-Latency-CMAF-Spezialfälle (`#EXT-X-PART`, chunked CMAF)
  über dokumentierte Folgehinweise hinaus.
- Kein neuer Player-SDK-Adapter und keine Wire-Änderung am Playback-
  Event-Schema.

### 0.2 Vorgänger-Gate

- `0.9.6` ist released; der Plan liegt unter
  `docs/planning/done/plan-0.9.6.md`.
- `0.9.6` hat NF-13 als offene Muss-Vollimplementierung auf diesen
  Plan verwiesen.
- DASH-Analyse aus `0.9.0` ist unverändert grün; HLS- und DASH-
  Contract-Fixtures sind Baseline für additive Änderungen.

### 0.3 Lastenheft-Patch `1.1.13` (Vorschlag)

Der Patch ergänzt `spec/lastenheft.md` mit RAK-60..RAK-63 und
markiert NF-13 als in `0.10.0` nur manifestbasiert teiladressiert.
NF-13 bleibt bis zur späteren binären CMAF-Konformitätsprüfung offen;
der Patch darf die Muss-Anforderung nicht als vollständig erfüllt
markieren.

| RAK | Priorität | Inhalt |
| --- | --------- | ------ |
| RAK-60 | Muss | CMAF-Scope ist normativ begrenzt: manifestbasierte Signalanalyse für HLS/fMP4 und DASH/CMAF; Segment-/MP4-Box-Parsing bleibt Folge-Scope, NF-13 bleibt bis dahin offen und darf nicht als vollständig erfüllt markiert werden. |
| RAK-61 | Muss | HLS-CMAF-Signale: `EXT-X-MAP`, fMP4-Segmentmuster und relevante Tags erzeugen stabile `cmaf`-Signals mit Confidence-Semantik im Analyseergebnis. |
| RAK-62 | Muss | DASH-CMAF-Signale: MPD-`mimeType`, `codecs`, `SegmentTemplate`/`SegmentList` und Initialization-Informationen erzeugen stabile `cmaf`-Signals mit Confidence-Semantik; MP4-MIME allein gilt nur als Indiz, nicht als CMAF-Konformitätsnachweis. |
| RAK-63 | Muss | CLI, API-Adapter, Contract-Fixtures und User-Doku führen CMAF-Signale additiv durch; bestehende HLS-/DASH-Smokes bleiben unverändert grün. |

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung + Lastenheft-Patch `1.1.13` + Fixture-Inventar | ⬜ |
| 1 | Result-Schema, Public API und Fixture-Vertrag für CMAF-Signale | ⬜ |
| 2 | HLS/fMP4-CMAF-Erkennung | ⬜ |
| 3 | DASH/CMAF-Erkennung | ⬜ |
| 4 | API-/CLI-Durchleitung, Doku und Smokes | ⬜ |
| 5 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

---

## 2. Tranche 0 — Aktivierung, Patch und Fixtures

Ziel: NF-13 wird vor Implementierung eindeutig messbar.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.10.0.md` nach
  `docs/planning/in-progress/plan-0.10.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] `spec/lastenheft.md` Header auf `1.1.13` erhöht.
- [ ] RAK-60..RAK-63 im Lastenheft ergänzt.
- [ ] [`plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c um
  `4a.16 Patch 1.1.13` ergänzt.
- [ ] Fixture-Inventar angelegt:
  - HLS CMAF VOD mit `EXT-X-MAP` und `.m4s`-Segmenten.
  - HLS TS als Negativ-/Regression-Pfad.
  - DASH CMAF VOD mit `SegmentTemplate initialization`.
  - DASH ohne CMAF-Signale als Negativ-/Regression-Pfad.

---

## 3. Tranche 1 — Result-Schema und Vertrag

Ziel: CMAF-Signale sind additiv und stabil, bevor Parser-Logik
ausgebaut wird.

DoD:

- [ ] `packages/stream-analyzer/src/types/result.ts` um ein
  additives `CmafSignalSummary`-Modell ergänzt, das ausschließlich in
  den bestehenden Detail-Objekten lebt:
  `MasterPlaylistDetails.cmaf`, `MediaPlaylistDetails.cmaf` und
  `DashManifestDetails.cmaf`. Der Analyzer-Envelope bekommt kein
  Top-Level-`cmaf`; `UnknownAnalysisResult.details` bleibt `null`.
  Modellfelder:
  - `present: boolean`
  - `source: "hls" | "dash" | "mixed"`
  - `confidence: "manifest" | "inferred"`
  - `signals[]` mit Code, Severity und Manifest-Anker.
- [ ] Public API exportiert die neuen CMAF-Typen über
  `packages/stream-analyzer/src/index.ts`.
- [ ] `packages/stream-analyzer/scripts/public-api.snapshot.txt` ist
  synchron aktualisiert; `make generated-drift-check` bleibt grün.
- [ ] Bestehende HLS-/DASH-Result-Fixtures bleiben byte-kompatibel
  oder werden mit dokumentierter additiver Schema-Erweiterung
  aktualisiert.
- [ ] Contract-Fixtures in `spec/contract-fixtures/analyzer/` und
  Go-Testdata-Kopien für API-Adapter ergänzt.
- [ ] Go-Adapter-Kontrakt ist explizit geprüft: weil `cmaf` in
  `details` liegt, reichen `apps/api/adapters/driven/streamanalyzer`
  und `apps/api/adapters/driving/http` die Signale über
  `EncodedDetails`/`details` unverändert durch; kein unbekanntes
  Top-Level-Feld darf still verworfen werden.
- [ ] Backward-Compatibility-Notiz in Stream-Analyzer-README:
  bestehende `analyzerKind:"hls"`/`"dash"` bleiben unverändert;
  CMAF ist ein Signal, kein dritter Manifesttyp.
- [ ] Bestehende Forward-Compat-Hinweise in
  `docs/user/stream-analyzer.md`, `packages/stream-analyzer/README.md`,
  `packages/stream-analyzer/src/types/result.ts` und
  `apps/api/hexagon/domain/stream_analysis.go` sind synchronisiert:
  CMAF wird in `0.10.0` als Signalmodell beschrieben, nicht als neuer
  `analyzerKind`.

---

## 4. Tranche 2 — HLS/fMP4-CMAF-Erkennung

Ziel: HLS-Manifeste mit CMAF/fMP4-Struktur werden zuverlässig erkannt.

DoD:

- [ ] Media-Playlist-Parser erkennt `EXT-X-MAP` als starkes
  CMAF/fMP4-Signal und schreibt es nach
  `MediaPlaylistDetails.cmaf.signals[]`.
- [ ] Segment-URI-Muster `.m4s`/`.cmfv`/`.cmfa` werden als
  schwächere manifestbasierte Hinweise erfasst.
- [ ] `EXT-X-INDEPENDENT-SEGMENTS` und Codec-/Map-Kontext werden als
  zusätzliche Signale dokumentiert, aber nicht allein als CMAF-
  Nachweis gewertet.
- [ ] Master-Playlist-Parser schreibt ein konservatives
  `MasterPlaylistDetails.cmaf`: Variant-URI-/Codec-Hinweise dürfen
  `present` nur mit `confidence:"inferred"` setzen; starke
  `EXT-X-MAP`-Signale entstehen erst in Media-Playlists.
- [ ] Tests decken positive, negative und gemischte HLS-Fälle ab.
- [ ] Bestehender HLS-Master-/Media-Pfad bleibt grün.

---

## 5. Tranche 3 — DASH/CMAF-Erkennung

Ziel: DASH-MPDs mit CMAF-kompatiblen Representation-/Segment-
Informationen werden als solche ausgewiesen.

DoD:

- [ ] DASH-Parser wertet `mimeType` `video/mp4`, `audio/mp4` und
  `application/mp4` als CMAF-relevante Indizien, aber nicht allein als
  Konformitätsnachweis.
- [ ] `SegmentTemplate@initialization`, `SegmentList/Initialization`
  und `Representation`-Codecs fließen in die Signalbewertung ein.
- [ ] DASH-Schema und Parser erfassen Initialization-Informationen
  explizit und vererbungsbewusst mindestens auf
  `MPD`/`Period`/`AdaptationSet`/`Representation`-Ebene:
  `SegmentTemplate@initialization`, `SegmentTemplate@media`,
  `SegmentList/Initialization@sourceURL` sowie relevante
  `BaseURL`-/URI-Muster. Diese Felder können als interne Parse-
  Metadaten oder additive `details`-Felder umgesetzt werden, müssen
  aber in `DashManifestDetails.cmaf` nachvollziehbare Manifest-Anker
  erzeugen.
- [ ] Confidence-Regeln sind getestet: MP4-MIME allein erzeugt nur
  `confidence:"inferred"`; Initialization-Informationen plus fMP4-
  Segmentmuster erzeugen ein stärkeres manifestbasiertes Signal.
- [ ] DASH-Live- und VOD-Fixtures behalten bestehende Mindestfelder
  aus RAK-58.
- [ ] Tests decken DASH-CMAF positiv, DASH ohne Initialization-Signal
  und fehlerhafte MPD-Strukturen ab.

---

## 6. Tranche 4 — API, CLI und Doku

Ziel: CMAF-Signale sind über alle bestehenden Analyzer-Pfade nutzbar.

DoD:

- [ ] `apps/api`-StreamAnalyzer-Adapter reicht `details.cmaf` im
  bestehenden Domain-Modell über `EncodedDetails` additiv durch; Tests
  prüfen, dass HLS-CMAF und DASH-CMAF im HTTP-`analysis.details.cmaf`
  sichtbar bleiben.
- [ ] HTTP-Contract-/Adapter-Tests decken HLS-CMAF und DASH-CMAF ab.
- [ ] CLI gibt die neuen Signale unverändert im JSON aus.
- [ ] `make smoke-cli` um mindestens eine CMAF-HLS- oder CMAF-DASH-
  Probe erweitert.
- [ ] [`docs/user/stream-analyzer.md`](../../user/stream-analyzer.md)
  beschreibt Scope, Beispiele, Grenzen und Exit-/Fehlerverhalten.
- [ ] `packages/stream-analyzer/README.md` synchronisiert.

---

## 7. Tranche 5 — Gates, Matrix und Release-Closeout

Ziel: `0.10.0` wird als Minor-Release sauber abgeschlossen.

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt:

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-60 | Muss | Scope-Text in Lastenheft und Plan-Scope-Definition; Segment-/Box-Parsing ausdrücklich offen; NF-13 nicht vollständig geschlossen | [ ] |
| RAK-61 | Muss | HLS-CMAF-Fixtures, Parser-Tests, Confidence-Regeln, CLI-Smoke | [ ] |
| RAK-62 | Muss | DASH-CMAF-Fixtures, Parser-Tests, Confidence-Regeln, API-Contract | [ ] |
| RAK-63 | Muss | API-/CLI-/Doku-Nachweise, Contract-Fixtures | [ ] |

- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make smoke-cli` grün.
- [ ] `make security-gates` grün oder CI-Job `Security gates` grün
  dokumentiert.
- [ ] Wave-2-Quality-Gates laut
  [`docs/user/releasing.md`](../../user/releasing.md) mit der
  Überschrift „Patch-Release-Konvention" vor dem Tag
  geprüft.
- [ ] Letzter `benchmark.yml`-Nightly ist grün oder die initiale
  Beobachtungsphase ohne Baseline ist dokumentiert.
- [ ] Kein offenes Crash-Issue mit Label `fuzz` aus dem letzten
  `fuzz.yml`-Nightly.
- [ ] Mutation-Score-Trend aus den letzten drei `mutation.yml`-
  Nightly-Artefakten geprüft; Score-Senkung begründet.
- [ ] Vollständiger Versions-Bump `0.9.6` → `0.10.0` in allen
  versionsführenden Stellen analog Release-Konvention.
- [ ] `CHANGELOG.md` mit `[0.10.0] - YYYY-MM-DD` aktualisiert.
- [ ] Roadmap-Status und Release-Übersicht auf `0.10.0` released und
  Folgephase offen
  aktualisiert.
- [ ] Plan nach `docs/planning/done/plan-0.10.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.10.0` erstellt.

## 8. Nicht-Ziele für Review

Review-Kommentare zu folgenden Themen sollen in Folge-Pläne, nicht in
`0.10.0`:

- Binäres MP4-/CMAF-Box-Parsing.
- Low-Latency-CMAF-Chunks und `#EXT-X-PART`-Analyse.
- Segment-Download, CDN-Checks oder Byte-Range-Verifikation.
- Player-SDK-CMAF-Playback-Support.
- Neue Storage-, Multi-Tenant- oder Kubernetes-Scope-Erweiterungen.
