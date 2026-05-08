# Implementation Plan — `0.10.0` (CMAF-Analyse / NF-13)

> **Status**: ⬜ open — noch nicht aktiviert. Dieser Plan darf erst nach
> explizitem Move nach `docs/planning/in-progress/` umgesetzt werden.
> Vorgänger ist `0.9.6` (Lastenheft-Konvergenz + Repo-Artefakte);
> Aktivierung erst nach dessen Release-Closeout.
>
> **Release-Typ**: Minor-Release nach `0.9.6` mit Lastenheft-Patch
> `1.1.13`, neuer RAK-Gruppe `RAK-60`..`RAK-64`,
> RAK-Verifikationsmatrix und Tag `v0.10.0`.
>
> **Ziel**: NF-13 (`CMAF-Analyse`, Muss) wird in `0.10.0` vollständig
> für den bewusst begrenzten Analyzer-Scope umgesetzt: additive
> manifestbasierte CMAF-Signal-Analyse plus binäre CMAF-
> Konformitätsprüfung ausgewählter Init-/Media-Segmente. `F-73` bleibt
> der historische Vorbereitungsschritt; NF-13 wird mit diesem Plan nicht
> über einen neuen Manifesttyp, sondern über HLS-/DASH-Details mit
> prüfbarer CMAF-Semantik geschlossen.
>
> **Bezug**:
> [`spec/lastenheft.md`](../../../spec/lastenheft.md) F-73, NF-13,
> RAK-58, RAK-59;
> [`docs/user/stream-analyzer.md`](../../user/stream-analyzer.md)
> mit der Überschrift „CMAF";
> [`packages/stream-analyzer/`](../../../packages/stream-analyzer/);
> [`apps/api/hexagon/domain/stream_analysis.go`](../../../apps/api/hexagon/domain/stream_analysis.go).
>
> **Nachfolger**: offen für Low-Latency-CMAF, vollständige Segmentset-
> Abdeckung, CDN-/Byte-Range-Sonderfälle und Player-SDK-CMAF-Support.

## 0. Konvention

DoD-Checkboxen tracken den Lieferstand:

- `[x]` ausgeliefert mit Commit-Hash.
- `[ ]` offen.
- `[!]` blockiert durch Scope-Entscheidung oder fehlende Fixture.
- 🟡 in Arbeit.

### 0.1 Scope-Definition

`0.10.0` liefert **CMAF-Analyse** als Kombination aus manifestbasierter
Signal-Erkennung und begrenzter binärer CMAF-Konformitätsprüfung.
Keine neue Analyzer-Art entsteht: CMAF bleibt ein Signal- und
Verifikationsmodell innerhalb der bestehenden HLS-/DASH-Analyse.

In Scope:

- HLS/fMP4-CMAF-Erkennung über `EXT-X-MAP`, Segment-URI-Muster und
  manifestbasierte Konsistenzsignale.
- DASH/CMAF-Erkennung über `mimeType` `video/mp4`/`audio/mp4`/
  `application/mp4`, `SegmentTemplate`/`SegmentList`-Initialisierung
  und Representation-Metadaten.
- Begrenzte binäre CMAF-Konformitätsprüfung für ausgewählte,
  manifestreferenzierte Init- und Media-Segmente:
  - HLS: `EXT-X-MAP`-Init-Segment plus erstes fMP4-Media-Segment je
    analysiertem Media-Manifest.
  - DASH: `SegmentTemplate@initialization` oder
    `SegmentList/Initialization@sourceURL` plus erstes ableitbares
    fMP4-Media-Segment je repräsentativem AdaptationSet. Für
    `SegmentList` gilt nur explizit referenziertes
    `SegmentURL@media` als ableitbares Media-Segment; fehlt es, wird
    nicht geraten, sondern deterministisch `skipped`.
  - Byte-Parser für ISO-BMFF-Boxen mit Nachweis mindestens von `ftyp`,
    `moov`, `moof`, `traf` und `tfdt`; `sidx` wird erkannt, ist aber
    kein Pflicht-Nachweis.
  - Separater bounded Binary-Segment-Loader statt Zweckentfremdung des
    bestehenden Manifest-Text-Loaders: Segment-Fetches liefern Bytes
    (`Uint8Array`), akzeptieren MP4-/Byte-Content-Types, nutzen aber
    dieselben SSRF-, Scheme-, Redirect- und Timeout-Regeln wie
    `loadManifest`.
  - Strikte Fetch-/Read-Grenzen: Default `maxSegmentBytes=2_000_000`,
    Default `maxBinarySegments=6`, Timeout/Redirects aus
    `AnalyzeOptions.fetch`, und deterministische `skipped`-/`failed`-
    Ergebnisse bei Limit-Verstößen.
- Additives Result-Schema für `details.cmaf` unter den bestehenden
  HLS- und DASH-Detail-Objekten, ohne bestehende HLS-/DASH-Felder zu
  brechen und ohne neues Top-Level-Feld im Analyzer/API-Envelope.
  `details.cmaf` ist optional und wird nur ausgegeben, wenn mindestens
  ein CMAF-Signal vorliegt; Negativ-/Regression-Fixtures ohne CMAF-
  Signale behalten ihre bisherige Detail-Form ohne `cmaf`. Jedes
  einzelne Signal trägt eine Confidence (`binary`, `manifest` oder
  `inferred`). Das `cmaf`-Objekt bedeutet „CMAF-Signale oder
  CMAF-Verifikation vorhanden"; eine Konformitätsaussage darf nur aus
  `details.cmaf.binary.status:"passed"` abgeleitet werden. Deshalb
  bekommt das Schema kein boolesches `present:true`-Feld. DASH-
  Resultate mit nur `video/mp4`/`audio/mp4`/`application/mp4` bekommen
  ein schwaches `confidence:"inferred"`-Summary; die bestehenden DASH-
  Contract-Fixtures werden absichtlich additiv aktualisiert und sind
  danach nicht byte-kompatibel zum `0.9.x`-Stand. HLS-`unknown` mit
  `details:null` bleibt ohne `cmaf`.
- CLI/API-Durchleitung und Doku für die neuen CMAF-Signale.
- Binäre Segment-Fetches werden für `kind:"url"` gegen die finale
  Manifest-URL und für `kind:"text"` nur bei gesetzter, sicherer
  `http:`-/`https:`-`baseUrl` ausgeführt. Text-Input ohne `baseUrl`
  behält manifestbasierte Signale, bekommt aber
  `details.cmaf.binary.status:"skipped"` mit Failure-Code
  `segment_base_url_missing`. Text-Input mit `file:`- oder anderem
  Nicht-HTTP(S)-`baseUrl` wird nicht lokal gelesen; betroffene Segment-
  Prüfungen werden deterministisch als `skipped` mit Failure-Code
  `segment_uri_blocked` berichtet.

Out of scope:

- Keine vollständige ISO-BMFF-/MP4-Dateivalidierung außerhalb der für
  CMAF nötigen Box-/Fragment-Indizien.
- Kein Download aller Segmente eines Streams und keine CDN-/Origin-
  Qualitätsprüfung.
- Keine Audio-/Video-Codec-Bitstream-Validierung und kein Decoding.
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

Der Patch ergänzt `spec/lastenheft.md` mit RAK-60..RAK-64 und
markiert NF-13 als in `0.10.0` für den Stream-Analyzer-Scope erfüllt:
manifestbasierte Signale plus begrenzte binäre CMAF-
Konformitätsprüfung. Folge-Scope bleibt nur für Low-Latency-CMAF,
vollständige Segmentset-Abdeckung und Player-/CDN-Sonderfälle offen.
Der bisherige Begriff „CMAF-Vollanalyse" wird dabei normativ
präzisiert: Vollständig heißt in `0.10.0` vollständig für den
Analyzer-Scope aus diesem Plan, nicht vollständige Prüfung aller
Segmente, Codecs, Byte-Ranges oder Player-Laufzeitpfade.

| RAK | Priorität | Inhalt |
| --- | --------- | ------ |
| RAK-60 | Muss | CMAF-Scope ist normativ begrenzt: manifestbasierte Signalanalyse plus begrenzte binäre Prüfung ausgewählter HLS-/DASH-Init- und Media-Segmente; das Lastenheft präzisiert „CMAF-Vollanalyse" als vollständige Erfüllung dieses Analyzer-Scopes, nicht als vollständige Segmentset-/Codec-/Player-Prüfung. |
| RAK-61 | Muss | HLS-CMAF-Signale: `EXT-X-MAP`, fMP4-Segmentmuster und relevante Tags erzeugen stabile `cmaf`-Signals mit Confidence-Semantik im Analyseergebnis. |
| RAK-62 | Muss | DASH-CMAF-Signale: MPD-`mimeType`, `codecs`, `SegmentTemplate`/`SegmentList` und Initialization-Informationen erzeugen stabile `cmaf`-Signals mit Confidence-Semantik; MP4-MIME allein gilt nur als Indiz, nicht als CMAF-Konformitätsnachweis. |
| RAK-63 | Muss | CLI, API-Adapter, Contract-Fixtures und User-Doku führen CMAF-Signale additiv durch; bestehende HLS-/DASH-Smokes bleiben unverändert grün. |
| RAK-64 | Muss | Binäre CMAF-Konformitätsprüfung: ISO-BMFF-Box-Parser validiert ausgewählte Init-/Media-Segmente bounded und meldet `details.cmaf.binary.status` mit nachvollziehbaren Box-/Segment-Nachweisen. |

## 1. Tranchen-Übersicht

| Tranche | Inhalt | Status |
| ------- | ------ | ------ |
| 0 | Plan-Aktivierung + Lastenheft-Patch `1.1.13` + Fixture-Inventar | ⬜ |
| 1 | Result-Schema, Public API und Fixture-Vertrag für CMAF-Signale | ⬜ |
| 2 | HLS/fMP4-CMAF-Erkennung | ⬜ |
| 3 | DASH/CMAF-Erkennung | ⬜ |
| 4 | Binäre CMAF-Konformitätsprüfung für Init-/Media-Segmente | ⬜ |
| 5 | API-/CLI-Durchleitung, Doku und Smokes | ⬜ |
| 6 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ⬜ |

---

## 2. Tranche 0 — Aktivierung, Patch und Fixtures

Ziel: NF-13 wird vor Implementierung eindeutig messbar.

DoD:

- [ ] Plan von `docs/planning/open/plan-0.10.0.md` nach
  `docs/planning/in-progress/plan-0.10.0.md` verschoben.
- [ ] `git status --short` vor erster Änderung dokumentiert.
- [ ] `spec/lastenheft.md` Header auf `1.1.13` erhöht.
- [ ] RAK-60..RAK-64 im Lastenheft ergänzt.
- [ ] NF-13-Text im Lastenheft von „CMAF-Vollanalyse" auf
  „CMAF-Analyse im Stream-Analyzer-Scope" präzisiert: erfüllt durch
  manifestbasierte Signale plus begrenzte binäre Init-/Media-Segment-
  Prüfung; explizit nicht umfasst sind vollständige Segmentset-
  Abdeckung, Codec-Decoding, Low-Latency-CMAF und Player-Laufzeitpfade.
- [ ] [`plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c um
  `4a.16 Patch 1.1.13` ergänzt.
- [ ] [`roadmap.md`](../in-progress/roadmap.md) vor erster
  Implementierung auf `0.10.0` als CMAF-Analyse mit manifestbasierten
  Signalen plus begrenzter binärer Konformitätsprüfung korrigiert,
  damit Folgephasen-Status und Plan-Scope nicht während der Umsetzung
  widersprechen.
- [ ] Fixture-Inventar angelegt:
  - HLS CMAF VOD mit `EXT-X-MAP` und `.m4s`-Segmenten.
  - HLS TS als Negativ-/Regression-Pfad.
  - DASH MP4-MIME-only als schwacher/inferred Pfad; bestehende DASH-
    Contract-Fixtures mit `video/mp4`/`audio/mp4` werden dafür additiv
    um `details.cmaf` erweitert und verlieren bewusst ihre
    Byte-Kompatibilität zum `0.9.x`-Stand.
  - DASH CMAF VOD mit `SegmentTemplate@initialization` plus fMP4-
    Segmentmuster als starker manifestbasierter Pfad.
  - DASH ohne CMAF-Signale als Negativ-/Regression-Pfad, z. B. ohne
    MP4-MIME, ohne Initialization und ohne fMP4-URI-Muster.
  - DASH-Vererbungsfixtures mit `BaseURL` und `SegmentTemplate` auf
    `MPD`-, `Period`-, `AdaptationSet`- und `Representation`-Ebene;
    mindestens ein mehrperiodiges Fixture pinnt stabile Index-/ID-
    Manifestanker.
  - HLS-`EXT-X-MAP`-Fixtures mit `URI`, relativer URI, absoluter URI
    und `BYTERANGE`; `BYTERANGE` wird in `0.10.0` erkannt und
    dokumentiert, aber nicht binär geladen, sondern als `skipped`
    berichtet.
  - Binäre Positive-Fixtures: minimales CMAF-Init-Segment mit `ftyp` +
    `moov` und minimales fragmentiertes Media-Segment mit `moof` /
    `traf` / `tfdt`.
  - Binäre Negativ-Fixtures: fehlendes oder inkompatibles `ftyp`,
    fehlendes `moof`, fehlendes `tfdt`, ungültige Box-Größe und Segment
    über dem konfigurierten Größenlimit.

---

## 3. Tranche 1 — Result-Schema und Vertrag

Ziel: CMAF-Signale sind additiv und stabil, bevor Parser-Logik
ausgebaut wird.

DoD:

- [ ] `packages/stream-analyzer/src/types/result.ts` um ein
  additives `CmafSignalSummary`-Modell ergänzt, das ausschließlich in
  den bestehenden Detail-Objekten lebt:
  `MasterPlaylistDetails.cmaf?`, `MediaPlaylistDetails.cmaf?` und
  `DashManifestDetails.cmaf?`. `cmaf` wird ausgelassen, wenn keine
  CMAF-Signale vorliegen; es wird nicht als `present:false`- oder
  `present:true`-Platzhalter in bestehenden Details serialisiert. Der
  Analyzer-Envelope bekommt kein Top-Level-`cmaf`;
  `UnknownAnalysisResult.details` bleibt `null`.
  Modellfelder:
  - `source: "hls" | "dash"`; ein `mixed`-Wert wird in `0.10.0` nicht
    eingeführt, weil jedes Summary unter genau einem HLS- oder DASH-
    Detail-Objekt lebt.
  - `confidence: "binary" | "manifest" | "inferred"` als aggregierte
    stärkste Confidence des Summary-Objekts. Die normative Ordnung ist
    `binary` > `manifest` > `inferred`; gemischte Signale aggregieren
    deterministisch auf den stärksten vorhandenen Wert.
  - `signals[]` mit `code`, `level`, `manifestAnchor` und eigener
    `confidence: "binary" | "manifest" | "inferred"`, damit gemischte
    starke und schwache Indizien auditierbar bleiben. `level` nutzt
    dieselbe Wertedomäne wie `AnalysisFinding.level`: `"info" |
    "warning" | "error"`.
  - `binary?: CmafBinaryVerification` mit `status: "passed" |
    "failed" | "skipped"`, `segmentsChecked[]`, `boxes[]`,
    `failures[]`, `limits` und `note?`. `status:"passed"` ist die
    einzige Stelle, aus der Doku und Konsumenten eine binäre CMAF-
    Konformitätsaussage für den geprüften Scope ableiten dürfen.
    `status:"skipped"` ist zulässig, wenn keine sicher ladbare Init-/
    Media-Segment-URI vorliegt; manifestbasierte Signale bleiben dann
    sichtbar, aber nicht konformitätsbeweisend.
    Wenn `cmaf.binary.enabled:false` gesetzt ist und ein `cmaf`-Summary
    in einem binär prüfbaren Detail-Scope ausgegeben wird (HLS Media-
    Playlist oder DASH-MPD), wird `binary` nicht weggelassen, sondern
    als `status:"skipped"` mit Failure-Code `binary_disabled`
    serialisiert; so bleibt für Konsumenten sichtbar, dass keine binäre
    Aussage versucht wurde. HLS-Master-Summaries sind in `0.10.0` nicht
    binär prüfbar, weil referenzierte Media-Playlists nicht nachgeladen
    werden; sie tragen auch bei deaktivierter Binary-Prüfung kein
    `binary`-Objekt.
    Aggregation ist deterministisch: Zuerst wird die geplante
    Pflichtprüfungsmenge aus Manifest-Scope und Limits gebildet
    (`maxBinarySegments` cappt die Menge vor dem Fetch; wegen dieses
    Caps nicht ausgewählte weitere Representations/Segmente werden als
    `not_planned_due_to_limit` in `note`/`limits` auditierbar, erzeugen
    aber keinen `skipped`-Einzelstatus und verhindern kein `passed`).
    Danach gilt: Gesamtstatus `failed`, sobald irgendeine geplante und
    geladene Pflichtprüfung fehlschlägt; `passed` nur, wenn alle
    geplanten Pflichtprüfungen für den gewählten, gecappten Scope
    bestanden wurden; `skipped`, wenn keine Pflichtprüfung
    fehlgeschlagen ist, aber mindestens eine geplante Pflichtprüfung
    wegen fehlender, nicht sicher auflösbarer, per Byte-Range
    ausgeschlossener oder beim Fetch/Read durch Limits blockierter
    Segment-Referenz nicht ausgeführt wurde. `segmentsChecked[]` trägt
    den jeweiligen Einzelstatus, damit gemischte DASH-AdaptationSet-
    Ergebnisse auditierbar bleiben.
    `limits` serialisiert mindestens `maxSegmentBytes`,
    `maxBinarySegments`, `timeoutMs` und `maxRedirects`, damit
    ausgelassene Prüfungen reproduzierbar bleiben.
  - `note?: string` darf knapp beschreiben, welcher Anteil nur
    manifestbasiert und welcher Anteil binär verifiziert wurde; Pflicht
    ist diese Klarstellung in Doku und README, nicht in jedem JSON-
    Result.
- [ ] `packages/stream-analyzer/src/types/input.ts` ergänzt eine
  additive Optionssektion für die binäre CMAF-Prüfung, z. B.
  `cmaf.binary.enabled`, `cmaf.binary.maxSegmentBytes` und
  `cmaf.binary.maxBinarySegments`. Defaults sind dokumentiert:
  `enabled:true`, `maxSegmentBytes=2_000_000`,
  `maxBinarySegments=6`. Diese Limits gelten zusätzlich zu
  `fetch.maxBytes`, das ausschließlich das Manifest-Body-Limit bleibt.
- [ ] Binary-Status- und Failure-Code-Vertrag ist vor Parser-/Loader-
  Implementierung festgelegt und in Fixtures/Testnamen sichtbar:
  - `binary_disabled`: `skipped`, wenn Binary-Prüfung per Option
    deaktiviert ist, aber `details.cmaf` in einem binär prüfbaren
    Detail-Scope vorhanden ist (HLS Media-Playlist oder DASH-MPD; HLS-
    Master-Summaries tragen kein `binary`-Objekt).
  - `segment_base_url_missing`: `skipped`, wenn Text-Input keine sichere
    `baseUrl` für relative Segmente liefert.
  - `segment_uri_blocked`: `skipped`, wenn eine relative/absolute
    Segment-URI nicht sicher auflösbar ist, die `baseUrl` oder Segment-
    URL kein `http:`-/`https:`-Scheme nutzt, oder Scheme, Credentials,
    SSRF- oder Redirect-Regeln verletzt.
  - `segment_reference_missing`: `skipped`, wenn für eine im Binary-
    Scope verpflichtende Init- oder Media-Prüfung keine Manifest-
    Referenz vorhanden ist, z. B. HLS ohne `EXT-X-MAP` im sonst
    binär prüfbaren fMP4-Pfad oder DASH `SegmentList` ohne
    `SegmentURL@media`.
  - `hls_map_byterange_unsupported`: `skipped` für HLS `EXT-X-MAP` mit
    `BYTERANGE`.
  - `dash_template_unresolved`: `skipped`, wenn DASH-Template-Variablen
    in `0.10.0` nicht deterministisch auflösbar sind.
  - `segment_fetch_failed`: `skipped` bei Segment-Fetch-Timeout oder
    HTTP-/Transportfehler vor erfolgreichem Body-Read.
  - `segment_content_type_unsupported`: `skipped`, wenn der Segment-
    Content-Type nicht MP4-/Byte-kompatibel ist.
  - `segment_too_large`: `skipped`, wenn ein Segment `maxSegmentBytes`
    oder das Body-Read-Limit überschreitet.
  - `cmaf_box_validation_failed`: `failed`, wenn ein geladenes Init-/
    Media-Segment fachlich nicht konforme Box-/Brand-Struktur hat.
  - `invalid_box_structure`: `failed`, wenn Box-Größe, Überlappung oder
    Parser-Fortschritt strukturell ungültig sind.

  `failed` ist damit für geladene, aber fachlich nicht konforme oder
  strukturell kaputte Bytes reserviert. `skipped` bedeutet, dass der
  Analyzer im sicheren/bounded Scope keine binäre Konformitätsaussage
  treffen konnte. Die `segmentsChecked[]`-Einträge tragen dieselben
  Codes; der Gesamtstatus aggregiert nach Fehler vor Skip vor Pass.
- [ ] Public API exportiert die neuen CMAF-Typen über
  `packages/stream-analyzer/src/index.ts`.
- [ ] Options-Wire-Vertrag ist festgelegt: `cmaf.binary.*` ist Public-
  TypeScript-API und wird vom analyzer-service als optionales
  Request-Feld akzeptiert, typ-/range-gefiltert und an
  `analyzeManifest` weitergereicht. `apps/api` nutzt in `0.10.0`
  standardmäßig die Analyzer-Defaults; wenn `/api/analyze` ebenfalls
  `cmaf.binary.*` entgegennehmen soll, werden HTTP-Request-Schema,
  Domain-Request, driven Adapter, öffentliche Doku und Tests in
  derselben Tranche erweitert. Ohne diese explizite Erweiterung darf
  der API-Pfad nicht behaupten, caller-setzbare CMAF-Binary-Limits
  durchzureichen.
- [ ] `packages/stream-analyzer/scripts/public-api.snapshot.txt` ist
  synchron aktualisiert; der Stream-Analyzer-Public-API-Check im
  Paket-`lint` bleibt grün. Falls `make generated-drift-check` den
  Stream-Analyzer-Snapshot in dieser Tranche zusätzlich als
  Generated-Artefakt aufnehmen soll, werden `Makefile`-Kommentar,
  Prüfkommando und Drift-Meldung im selben Commit erweitert.
- [ ] Bestehende HLS-Result-Fixtures ohne CMAF-Signale bleiben
  byte-kompatibel. Bestehende DASH-Result-Fixtures mit MP4-MIME werden
  mit dokumentierter additiver `details.cmaf`-Erweiterung aktualisiert,
  weil MP4-MIME-only ab `0.10.0` bewusst als
  `confidence:"inferred"`-Signal sichtbar wird.
- [ ] Contract-Fixtures in `spec/contract-fixtures/analyzer/` und
  Go-Testdata-Kopien für API-Adapter ergänzt.
- [ ] `Makefile`-Fixture-Sync ist synchron erweitert:
  `sync-contract-fixtures` kopiert alle neuen Analyzer-CMAF-Fixtures in
  `apps/api/adapters/driven/streamanalyzer/testdata/`,
  `generated-drift-check` diffed diese neuen Kopien, die
  Drift-Fehlermeldung nennt `make sync-contract-fixtures` als Fix, und
  `spec/contract-fixtures/analyzer/README.md` beschreibt die neuen
  Fixtures. Die kopierte Fixture-Anzahl in der Make-Ausgabe wird
  angepasst.
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
- [ ] Media-Playlist-Parser extrahiert `EXT-X-MAP` strukturiert:
  `URI`, optional `BYTERANGE`, Manifestanker, rohe Attributwerte und
  gegen `baseUrl` aufgelöste URI, falls möglich. Diese Daten dürfen
  intern oder additiv in `details` leben, müssen aber vom Binary-Pfad
  ohne erneutes String-Parsing nutzbar sein.
- [ ] `EXT-X-MAP` mit `BYTERANGE` wird in `0.10.0` nicht per HTTP
  Range geladen. Das Manifest-Signal bleibt sichtbar; die binäre Init-
  Prüfung bekommt `status:"skipped"` bzw. einen Segment-Einzelstatus
  `skipped` mit Failure-Code `hls_map_byterange_unsupported`.
- [ ] Segment-URI-Muster `.m4s`/`.cmfv`/`.cmfa` werden als
  schwächere manifestbasierte Hinweise erfasst.
- [ ] `EXT-X-INDEPENDENT-SEGMENTS` und Codec-/Map-Kontext werden als
  zusätzliche Signale dokumentiert, aber nicht allein als CMAF-
  Nachweis gewertet.
- [ ] Master-Playlist-Parser schreibt ein konservatives
  `MasterPlaylistDetails.cmaf`: Variant-URI-Hinweise auf fMP4-/CMAF-
  Media-Playlists oder fMP4-spezifische Variant-Kontexte dürfen nur ein
  Summary mit `confidence:"inferred"` erzeugen. Der Master-Pfad bleibt
  eine Single-Manifest-Analyse und lädt referenzierte Media-Playlists
  nicht nach; wenn eine Variant-URI auf eine Media-Playlist zeigt,
  wird diese erst bei einem separaten Analyzer-Aufruf als eigenes
  Manifest geprüft. `CODECS` allein erzeugt kein CMAF-Signal, weil
  klassische TS-HLS-Master ebenfalls Codecs tragen. Starke
  `EXT-X-MAP`-Signale entstehen erst in Media-Playlists. Das Summary
  darf nicht als bestätigte CMAF-Konformität dokumentiert werden.
  Master-Summaries tragen in `0.10.0` kein `binary`-Objekt, weil der
  Binary-Scope keine referenzierten Media-Playlists nachlädt; binäre
  `passed`-/`failed`-/`skipped`-Status entstehen erst bei separater
  Analyse einer Media-Playlist.
- [ ] Tests decken positive, negative und gemischte HLS-Fälle ab.
- [ ] HLS-Master-Negativfixture pinnt eine Master-Playlist mit
  `CODECS` und TS-basierten Variant-URIs/Media-Playlists; daraus darf
  kein `MasterPlaylistDetails.cmaf` entstehen.
- [ ] Bestehender HLS-Master-/Media-Pfad bleibt grün.

---

## 5. Tranche 3 — DASH/CMAF-Erkennung

Ziel: DASH-MPDs mit CMAF-kompatiblen Representation-/Segment-
Informationen werden als solche ausgewiesen.

DoD:

- [ ] DASH-Parser wertet `mimeType` `video/mp4`, `audio/mp4` und
  `application/mp4` als CMAF-relevante Indizien, aber nicht allein als
  Konformitätsnachweis.
- [ ] `SegmentTemplate@initialization`, `SegmentTemplate@media`,
  `SegmentList/Initialization`, `SegmentList/SegmentURL@media` und
  `Representation`-Codecs fließen in die Signalbewertung ein.
- [ ] DASH-Schema und Parser erfassen Initialization-Informationen
  explizit und vererbungsbewusst mindestens auf
  `MPD`/`Period`/`AdaptationSet`/`Representation`-Ebene:
  `SegmentTemplate@initialization`, `SegmentTemplate@media`,
  `SegmentList/Initialization@sourceURL`,
  `SegmentList/SegmentURL@media` sowie relevante
  `BaseURL`-/URI-Muster. Diese Felder können als interne Parse-
  Metadaten oder additive `details`-Felder umgesetzt werden, müssen
  aber in `DashManifestDetails.cmaf` nachvollziehbare Manifest-Anker
  erzeugen. Für mehrperiodige MPDs muss der Anker eindeutig sein, z. B.
  `MPD/Period[0]/AdaptationSet[id=video]/Representation[id=v1]/...`;
  bei fehlenden IDs werden stabile Index-Anker verwendet. Das konkrete
  Signal-Feld benennt zusätzlich das auslösende Attribut, etwa
  `SegmentTemplate@initialization`.
- [ ] DASH-Parser-Strategie ist vor Umsetzung festgelegt: entweder
  der bestehende regex-basierte Parser wird gezielt um die benötigten
  `BaseURL`-/`SegmentTemplate`-/`SegmentList`-Metadaten erweitert und
  mit Vererbungsfixtures abgesichert, oder Tranche 3 migriert auf eine
  kleine XML-Parser-Abhängigkeit. Die Entscheidung wird im Plan-DoD
  dokumentiert; stille Teilvererbung reicht für RAK-62/RAK-64 nicht.
- [ ] DASH-Template-Auflösungs-Scope ist fixiert und getestet:
  `SegmentTemplate@initialization` und `@media` dürfen in `0.10.0`
  deterministisch nur `$RepresentationID$`, `$Bandwidth$`, `$Number$`
  sowie printf-artige `$Number%0Nd$`-Platzhalter verwenden. `startNumber`
  wird berücksichtigt; fehlt es, gilt DASH-Default `1`. `$Time$`,
  `SegmentTimeline`-abhängige Auflösung und unbekannte Template-
  Variablen sind in `0.10.0` nicht ableitbar und führen im Binary-Pfad
  zu `dash_template_unresolved`. Die manifestbasierten Signale bleiben
  trotzdem sichtbar.
- [ ] Confidence-Regeln sind getestet: MP4-MIME allein erzeugt nur
  `confidence:"inferred"`; Initialization-Informationen plus fMP4-
  Segmentmuster erzeugen ein stärkeres manifestbasiertes Signal.
- [ ] DASH-Tests pinnen drei getrennte Fälle: MP4-MIME-only als
  `confidence:"inferred"`, Initialization plus fMP4-Segmentmuster als
  `confidence:"manifest"` und ein echtes Negativ-Fixture ohne MP4-MIME,
  ohne Initialization und ohne fMP4-URI-Muster.
- [ ] DASH-Tests pinnen Vererbung und URI-Auflösung getrennt:
  `BaseURL` auf MPD-/Period-/AdaptationSet-/Representation-Ebene,
  `SegmentTemplate`-Vererbung und Override-Verhalten, `SegmentList`
  mit `Initialization@sourceURL` und `SegmentURL@media`, sowie
  mehrperiodige stabile Manifestanker bei fehlenden IDs.
- [ ] DASH-Live- und VOD-Fixtures behalten bestehende Mindestfelder
  aus RAK-58.
- [ ] Tests decken DASH-CMAF positiv, DASH ohne Initialization-Signal
  und fehlerhafte MPD-Strukturen ab.

---

## 6. Tranche 4 — Binäre CMAF-Konformitätsprüfung

Ziel: CMAF wird nicht nur als Manifest-Indiz erkannt, sondern im
begrenzten Analyzer-Scope über echte Segment-/Box-Daten geprüft.

DoD:

- [ ] ISO-BMFF-Box-Reader implementiert:
  - liest 32-bit und `largesize`-Boxgrößen bounds-checkend,
  - erkennt ungültige, überlappende oder nicht fortschreitende Boxen,
  - bricht bei konfiguriertem Byte-Limit deterministisch ab,
  - liefert stabile Box-Anker (`segment:init:ftyp`,
    `segment:media[0]:moof/traf/tfdt`) für `details.cmaf.binary`.
- [ ] Bounded Binary-Segment-Loader implementiert getrennt von
  `loadManifest`:
  - gibt Bytes zurück, nicht UTF-8-Text,
  - sendet einen MP4-orientierten `Accept`-Header,
  - erlaubt mindestens `video/mp4`, `audio/mp4`, `application/mp4`,
    `application/octet-stream` und leeren Content-Type,
  - nutzt dieselbe URL-Validierung, DNS-/SSRF-Prüfung,
    Redirect-Behandlung und Timeout-Mechanik wie der Manifest-Loader,
  - erzwingt `maxSegmentBytes` pro Segment und
    `maxBinarySegments` über den gesamten Analyseaufruf,
  - mappt blockierte, zu große, nicht ladbare oder Content-Type-
    inkompatible Segment-Fetches auf auditierbare
    `segmentsChecked[]`-/`failures[]`-Einträge statt auf einen
    Top-Level-Analysefehler.
- [ ] CMAF-Init-Prüfung validiert mindestens:
  - `ftyp` vorhanden,
  - Brand-Policy ist als getestete Allowlist umgesetzt: `cmfc` und
    `cmfs` gelten als CMAF-kompatible Brands; generische MP4-/ISO-
    BMFF-Brands wie `isom`, `iso6`, `mp41` oder `mp42` reichen allein
    nicht für `status:"passed"`,
  - `moov` vorhanden,
  - keine offensichtlich widersprüchliche Top-Level-Box-Struktur.
- [ ] CMAF-Media-Fragment-Prüfung validiert mindestens:
  - `moof` vorhanden,
  - mindestens ein `traf` unter `moof`,
  - `tfdt` unter `traf` vorhanden,
  - optionales `sidx` wird erkannt und berichtet, aber nicht als Muss
    gewertet.
- [ ] Segment-Resolver lädt nur manifestreferenzierte Init-/Media-
  Segment-URIs. URL-Input löst relative Pfade gegen die finale
  Manifest-URL auf; Text-Input löst relative Pfade nur bei gesetzter
  sicherer `http:`-/`https:`-`baseUrl` auf. Ohne `baseUrl`, mit
  Nicht-HTTP(S)-`baseUrl` oder bei unsicherer/nicht auflösbarer URI wird
  nicht gefetched, sondern `skipped` mit eindeutigem Failure-Code
  berichtet.
- [ ] HLS-Binary-Pfad prüft `EXT-X-MAP` plus erstes fMP4-Media-Segment;
  wenn eine der beiden URIs fehlt oder nicht ladbar ist, entsteht
  `details.cmaf.binary.status:"skipped"` oder `"failed"` mit
  nachvollziehbarem Failure-Code, kein stiller Erfolg.
  `EXT-X-MAP`-Attribute werden aus der strukturierten Parser-Ausgabe
  genutzt; erneutes Ad-hoc-Parsen der Manifestzeile im Binary-Pfad ist
  nicht zulässig.
- [ ] DASH-Binary-Pfad prüft Initialization plus erstes ableitbares
  fMP4-Media-Segment je repräsentativem AdaptationSet. Ableitbar sind
  nur explizite `SegmentList/SegmentURL@media`-Referenzen oder
  Templates aus dem Scope von Tranche 3:
  `$RepresentationID$`, `$Bandwidth$`, `$Number$` und
  `$Number%0Nd$` mit `startNumber` bzw. Default `1`. `$Time$`,
  `SegmentTimeline`-abhängige Auflösung oder unbekannte Variablen werden
  nicht geraten, sondern als `skipped` mit Failure-Code
  `dash_template_unresolved` gemeldet. Fehlt bei `SegmentList` die
  Media-Referenz vollständig, wird `segment_reference_missing`
  gemeldet.
- [ ] Positive und negative Binär-Fixtures decken Init, Media, fehlende
  Pflichtboxen, kaputte Boxgrößen, Größenlimit und nicht auflösbare
  Segment-URIs ab.
- [ ] Tests pinnen Status-Mapping:
  - `passed` nur bei bestandener Init- und Media-Prüfung,
  - `failed` bei geladener, aber nicht konformer Box-Struktur,
  - `skipped` bei fehlender oder aus Sicherheits-/Scope-Gründen nicht
    ladbarer Segmentreferenz,
  - `binary_disabled`, `segment_base_url_missing`,
    `segment_uri_blocked`, `segment_reference_missing`,
    `hls_map_byterange_unsupported`,
    `dash_template_unresolved`, `segment_fetch_failed`,
    `segment_content_type_unsupported`, `segment_too_large`,
    `cmaf_box_validation_failed` und `invalid_box_structure` nach der
    Status-/Failure-Code-Tabelle aus Tranche 1,
  - `maxBinarySegments` cappt die geplante Pflichtprüfungsmenge vor dem
    Fetch; nicht geplante weitere Segmente/Representations verhindern
    kein `passed`, müssen aber über `limits`/`note` auditierbar sein,
  - gemischte DASH-Ergebnisse aggregieren nach der Regel aus Tranche 1:
    jeder Fehler gewinnt vor `skipped`, `skipped` gewinnt vor `passed`.
- [ ] Fehler aus binärer Prüfung bleiben Findings oder
  `details.cmaf.binary.failures[]`; sie ändern nicht das bestehende
  `status:"ok"` des Analyse-Results, solange das Manifest selbst
  erfolgreich analysiert wurde.

---

## 7. Tranche 5 — API, CLI und Doku

Ziel: CMAF-Signale sind über alle bestehenden Analyzer-Pfade nutzbar.

DoD:

- [ ] `apps/api`-StreamAnalyzer-Adapter reicht `details.cmaf` im
  bestehenden Domain-Modell über `EncodedDetails` additiv durch; Tests
  prüfen, dass HLS-CMAF, DASH-CMAF und
  `analysis.details.cmaf.binary.status` im HTTP-Result sichtbar
  bleiben.
- [ ] `apps/analyzer-service` akzeptiert optional
  `{ "cmaf": { "binary": { "enabled", "maxSegmentBytes",
  "maxBinarySegments" } } }`, filtert ungültige Werte wie beim
  bestehenden `fetch`-Block, merged die Werte mit Analyzer-Defaults und
  dokumentiert, dass `allowPrivateNetworks` weiterhin ausschließlich
  service-seitig über `ANALYZER_ALLOW_PRIVATE_NETWORKS` gesetzt werden
  kann.
- [ ] `/api/analyze`-Vertrag ist bewusst entschieden und getestet:
  entweder bleibt `cmaf.binary.*` im öffentlichen API-Request
  unsupported und der API-Pfad nutzt Defaults, oder Request-Payload,
  Domain-Modell und driven Adapter leiten die drei Binary-Optionen
  explizit an den analyzer-service durch. Der gewählte Pfad steht in
  `docs/user/stream-analyzer.md` und in den HTTP-Contract-Tests.
- [ ] HTTP-Contract-/Adapter-Tests decken HLS-CMAF und DASH-CMAF ab.
  Mindestens ein Test pinnt die öffentliche `/api/analyze`-Antwort mit
  `{analysis, session_link}`-Wrapper und verifiziert
  `analysis.details.cmaf.binary.status`; die interne driven-Adapter-
  Fixture alleine reicht für RAK-63/RAK-64 nicht als HTTP-Wire-
  Nachweis.
- [ ] CLI gibt die neuen Signale unverändert im JSON aus.
- [ ] CLI bekommt einen dokumentierten, bewusst opt-in Schalter für
  lokale Fixture-/Lab-URLs, z. B. Env
  `MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS=true`, der ausschließlich
  `fetch.allowPrivateNetworks=true` an den Analyzer weitergibt.
  Ohne diesen Schalter bleibt der bestehende SSRF-Default aktiv und der
  vorhandene URL-SSRF-Smoke muss weiterhin `fetch_blocked` liefern.
- [ ] `make smoke-cli` um je eine CMAF-HLS- und CMAF-DASH-Probe mit
  bestandener binärer Prüfung erweitert. Der Smoke nutzt einen lokalen
  Fixture-HTTP-Server, der Manifest, Init-Segment und Media-Segment aus
  `spec/contract-fixtures/analyzer/` ausliefert, und setzt den neuen
  CLI-Opt-in-Schalter für private/loopback URLs nur für diese beiden
  Probes. Datei-Input mit `file://`-`baseUrl` gilt nicht als Ersatz für
  diesen Nachweis, weil der Segment-Loader HTTP(S)-SSRF-Regeln nutzt.
  Externe Netzabhängigkeit oder öffentliche Testmedien sind für den
  Release-Nachweis nicht zulässig.
- [ ] [`docs/user/stream-analyzer.md`](../../user/stream-analyzer.md)
  beschreibt Scope, Beispiele, Grenzen, Binary-Statuswerte und Exit-/
  Fehlerverhalten.
- [ ] `packages/stream-analyzer/README.md` synchronisiert.

---

## 8. Tranche 6 — Gates, Matrix und Release-Closeout

Ziel: `0.10.0` wird als Minor-Release sauber abgeschlossen.

DoD:

- [ ] RAK-Verifikationsmatrix vollständig ausgefüllt:

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-60 | Muss | Scope-Text in Lastenheft und Plan-Scope-Definition; NF-13 als „CMAF-Analyse im Stream-Analyzer-Scope" präzisiert und in diesem Scope durch manifestbasierte plus binäre CMAF-Prüfung geschlossen | [ ] |
| RAK-61 | Muss | HLS-CMAF-Fixtures, Parser-Tests, Confidence-Regeln, CLI-Smoke | [ ] |
| RAK-62 | Muss | DASH-CMAF-Fixtures, Parser-Tests, Confidence-Regeln, API-Contract | [ ] |
| RAK-63 | Muss | API-/CLI-/Doku-Nachweise, Contract-Fixtures | [ ] |
| RAK-64 | Muss | ISO-BMFF-Box-Parser, Binär-Fixtures, Binary-Status-Tests, bounded Segment-Loader, HTTP-/CLI-Nachweis | [ ] |

- [ ] `make docs-check` grün.
- [ ] `make build` grün.
- [ ] `make gates` grün.
- [ ] `make smoke-cli` grün.
- [ ] Release-Smokes laut [`docs/user/releasing.md`](../../user/releasing.md)
  §2 vollständig geprüft oder begründet gewaved:
  `make smoke-analyzer`, `make smoke-observability`, `make browser-e2e`,
  `make smoke-mediamtx`, `make smoke-srt`, `make smoke-srt-health`,
  `make smoke-dash`, `make smoke-webrtc-prep`,
  `make smoke-webrtc-stats-drift` und `make smoke-srs`.
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
- [ ] Roadmap-Status und Release-Übersicht auf `0.10.0` released
  aktualisiert; Folgephase beschreibt nur noch bewusst ausgegrenzte
  CMAF-Erweiterungen wie Low-Latency-CMAF oder vollständige
  Segmentset-Abdeckung, nicht die NF-13-Basiserfüllung.
- [ ] Plan nach `docs/planning/done/plan-0.10.0.md` verschoben und
  Status auf ✅ released aktualisiert.
- [ ] Annotierter Tag `v0.10.0` erstellt.

## 9. Nicht-Ziele für Review

Review-Kommentare zu folgenden Themen sollen in Folge-Pläne, nicht in
`0.10.0`:

- Low-Latency-CMAF-Chunks und `#EXT-X-PART`-Analyse.
- Vollständiger Download aller Segmente, CDN-Checks oder Byte-Range-
  Verifikation.
- Player-SDK-CMAF-Playback-Support.
- Neue Storage-, Multi-Tenant- oder Kubernetes-Scope-Erweiterungen.
