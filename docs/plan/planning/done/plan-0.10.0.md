# Implementation Plan — `0.10.0` (CMAF-Analyse / NF-13)

> **Status**: ✅ released — Minor-Release am `2026-05-09`, Tag
> `v0.10.0`. Lastenheft-Patch `1.1.13` mit RAK-60..RAK-64 in §13.12
> verankert den normativ begrenzten Analyzer-Scope. Vorgänger
> `0.9.6` (Lastenheft-Konvergenz; Tag `v0.9.6` auf `ad20228`, Plan
> in [`done/plan-0.9.6.md`](../done/plan-0.9.6.md)).
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
> [`spec/lastenheft.md`](../../../../spec/lastenheft.md) F-73, NF-13,
> RAK-58, RAK-59;
> [`docs/user/stream-analyzer.md`](../../../user/stream-analyzer.md)
> mit der Überschrift „CMAF";
> [`packages/stream-analyzer/`](../../../../packages/stream-analyzer);
> [`apps/api/hexagon/domain/stream_analysis.go`](../../../../apps/api/hexagon/domain/stream_analysis.go).
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
    analysiertem Media-Manifest. HLS-Byte-Range-Segmente werden in
    `0.10.0` nicht per HTTP Range geladen; `EXT-X-MAP` mit
    `BYTERANGE` und `#EXT-X-BYTERANGE` auf dem ersten fMP4-Media-
    Segment führen deterministisch zu `skipped`.
  - DASH: `SegmentTemplate@initialization` oder
    `SegmentList/Initialization@sourceURL` plus erstes ableitbares
    fMP4-Media-Segment je deterministisch ausgewählter
    Representation. Die Pflichtprüfungsmenge entsteht in Manifest-
    Reihenfolge aus jeder `Period` und jedem `AdaptationSet`, das
    mindestens ein CMAF-relevantes Signal trägt. Pro AdaptationSet wird
    genau eine Representation geprüft: bevorzugt die erste
    Representation mit eigener oder geerbter Initialization- und Media-
    Referenz; sonst die erste Representation des Sets. Diese Auswahl
    wird über stabile Manifestanker dokumentiert und bildet die Basis
    für `requiredSegmentChecks`. Für `SegmentList` gilt nur explizit
    referenziertes `SegmentURL@media` als ableitbares Media-Segment;
    fehlt es, wird nicht geraten, sondern deterministisch `skipped`.
  - Byte-Parser für ISO-BMFF-Boxen mit Nachweis mindestens von `ftyp`,
    `moov`, `styp`, `moof`, `traf`, `tfdt` und `mdat`; `sidx` wird
    erkannt, ist aber kein Pflicht-Nachweis. Der normative Brand-Scope
    für `0.10.0` ist bewusst fixiert: CMAF-Header müssen `cmfc` oder
    `cmf2` im `ftyp` tragen; CMAF-Media-Segmente müssen einen `styp`
    mit mindestens einem CMAF-Segment-/Track-Brand aus `cmfs`, `cmff`,
    `cmfc` oder `cmf2` tragen. `cmf1` und neuere Structural-Brand-
    Profile bleiben Folge-Scope, bis sie in Projekt-Doku, Fixtures und
    Kompatibilitätsaussage explizit aufgenommen werden.
  - Separater bounded Binary-Segment-Loader statt Zweckentfremdung des
    bestehenden Manifest-Text-Loaders: Segment-Fetches liefern Bytes
    (`Uint8Array`), akzeptieren MP4-/Byte-Content-Types, nutzen aber
    dieselben SSRF-, Scheme-, Redirect- und Timeout-Regeln wie
    `loadManifest`.
  - Strikte Fetch-/Read-Grenzen: Default `maxSegmentBytes=2_000_000`,
    Default `maxBinarySegments=6`, Timeout/Redirects aus
    `AnalyzeOptions.fetch`, und deterministische `skipped`-/`failed`-
    Ergebnisse bei Limit-Verstößen. Der Default deckt bewusst bis zu
    drei DASH-AdaptationSets mit je Init- und erstem Media-Segment ab;
    größere Pflichtprüfmengen werden nicht als bestanden gewertet,
    sondern bekommen für die überschüssigen Checks
    `not_planned_due_to_limit`, sofern der Aufrufer
    `maxBinarySegments` nicht erhöht.
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
  ein schwaches `confidence:"inferred"`-Summary. Bei Default
  `cmaf.binary.enabled:true` tragen sie zusätzlich
  `details.cmaf.binary.status:"skipped"` mit
  `segment_reference_missing`, weil keine Init-/Media-Referenzen für
  eine binäre Pflichtprüfung vorliegen; nur HLS-Master-Summaries
  bleiben grundsätzlich ohne `binary`-Objekt. Die bestehenden DASH-
  Contract-Fixtures werden absichtlich additiv aktualisiert und sind
  danach nicht byte-kompatibel zum `0.9.x`-Stand. HLS-`unknown` mit
  `details:null` bleibt ohne `cmaf`.
- CLI/API-Durchleitung und Doku für die neuen CMAF-Signale.
- Binäre Segment-Fetches werden für `kind:"url"` gegen die finale
  Manifest-URL und für `kind:"text"` nur bei gesetzter, sicherer
  `http:`-/`https:`-`baseUrl` ausgeführt. Diese `baseUrl` ist im
  Text-Pfad der Trust-Anker für alle Segment-Fetches: relative
  Segment-URIs werden gegen sie aufgelöst; absolute Segment-URIs werden
  nur gefetched, wenn eine sichere HTTP(S)-`baseUrl` vorhanden ist und
  die absolute Segment-URL selbst die Segment-URI-Sicherheitsregeln
  erfüllt. Text-Input ohne `baseUrl` behält manifestbasierte Signale;
  sobald eine manifestseitig vorhandene Init-/Media-Segment-Referenz
  deshalb nicht ladbar ist, bekommt die betroffene Prüfung
  `details.cmaf.binary.status:"skipped"` mit Failure-Code
  `segment_base_url_missing` unabhängig davon, ob die Manifest-Referenz
  relativ oder absolut notiert ist. Fehlt dagegen bereits die
  manifestseitige Init-/Media-Referenz, gewinnt
  `segment_reference_missing` gemäß der Failure-Code-Präzedenz aus
  Tranche 1. Text-Input mit `file:`- oder anderem Nicht-HTTP(S)-
  `baseUrl` wird nicht lokal gelesen; betroffene Segment-Prüfungen
  werden deterministisch als `skipped` mit Failure-Code
  `segment_uri_blocked` berichtet. Ist eine sichere HTTP(S)-`baseUrl`
  vorhanden, blocken unsichere absolute Segment-URIs ebenfalls mit
  `segment_uri_blocked`.

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
| 0 | Plan-Aktivierung + Lastenheft-Patch `1.1.13` + Fixture-Inventar | ✅ |
| 1 | Result-Schema, Public API und Fixture-Vertrag für CMAF-Signale | ✅ |
| 2 | HLS/fMP4-CMAF-Erkennung | ✅ |
| 3 | DASH/CMAF-Erkennung | ✅ |
| 4 | Binäre CMAF-Konformitätsprüfung für Init-/Media-Segmente | ✅ |
| 5 | API-/CLI-Durchleitung, Doku und Smokes | ✅ |
| 6 | Gates, RAK-Verifikationsmatrix, Versions-Bump, Closeout und Tag | ✅ |

---

## 2. Tranche 0 — Aktivierung, Patch und Fixtures

Ziel: NF-13 wird vor Implementierung eindeutig messbar.

DoD:

- [x] Plan von `docs/planning/open/plan-0.10.0.md` nach
  `docs/planning/in-progress/plan-0.10.0.md` verschoben.
- [x] `git status --short` vor erster Änderung dokumentiert: working tree
  clean (Tag `v0.9.6` auf `ad20228`, danach nur README-Politur ohne
  Repo-Drift).
- [x] `spec/lastenheft.md` Header auf `1.1.13` erhöht.
- [x] RAK-60..RAK-64 im Lastenheft ergänzt (neuer §13.12).
- [x] NF-13-Text im Lastenheft von „CMAF-Vollanalyse" auf
  „CMAF-Analyse im Stream-Analyzer-Scope" präzisiert: erfüllt durch
  manifestbasierte Signale plus begrenzte binäre Init-/Media-Segment-
  Prüfung; explizit nicht umfasst sind vollständige Segmentset-
  Abdeckung, Codec-Decoding, Low-Latency-CMAF und Player-Laufzeitpfade.
- [x] [`plan-0.1.0.md`](../done/plan-0.1.0.md) Tranche 0c um
  `4a.16 Patch 1.1.13` ergänzt.
- [x] [`roadmap.md`](../in-progress/roadmap.md) vor erster
  Implementierung auf `0.10.0` als CMAF-Analyse mit manifestbasierten
  Signalen plus begrenzter binärer Konformitätsprüfung korrigiert
  (§1.2 Aktivierung, §2 Schritt 45 angelegt, §3 Release-Tabellenzeile
  ergänzt), damit Folgephasen-Status und Plan-Scope nicht während der
  Umsetzung widersprechen.
- [x] Fixture-Inventar angelegt: tabellarisch in
  [`spec/contract-fixtures/analyzer/README.md`](../../../../spec/contract-fixtures/analyzer/README.md)
  unter „CMAF-Fixture-Inventar (Plan `0.10.0`, Tranche 0)" mit Pflicht-
  Brand-Allowlist, Tranchen-Zuordnung (T2 HLS, T3 DASH, T4 binär) und
  Hinweis, dass `success-dash-vod.json`/`success-dash-live.json`
  additiv und bewusst nicht byte-kompatibel zum `0.9.x`-Stand erweitert
  werden. Detailbullets aus dem Plan-Original:
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
  - HLS-Media-Byte-Range-Fixture mit `#EXT-X-BYTERANGE` vor dem ersten
    fMP4-Media-Segment; das Manifest-Signal bleibt sichtbar, die Media-
    Segment-Prüfung wird mit `hls_media_byterange_unsupported`
    übersprungen.
  - Binäre Positive-Fixtures: minimales CMAF-Init-Segment mit `ftyp`
    (`cmfc` oder `cmf2`) + `moov` und minimales fragmentiertes Media-
    Segment mit `styp` (`cmfs`, `cmff`, `cmfc` oder `cmf2`) + `moof` /
    `traf` / `tfdt` / `mdat`.
  - Binäre Negativ-Fixtures: fehlendes oder inkompatibles `ftyp`,
    `ftyp` nur mit `cmfs`, fehlendes oder inkompatibles `styp`,
    fehlendes `moof`, fehlendes `tfdt`, fehlendes `mdat`, ungültige
    Box-Größe und Segment über dem konfigurierten Größenlimit.

---

## 3. Tranche 1 — Result-Schema und Vertrag

Ziel: CMAF-Signale sind additiv und stabil, bevor Parser-Logik
ausgebaut wird.

DoD:

- [x] `packages/stream-analyzer/src/types/result.ts` um ein
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
    deterministisch auf den stärksten positiven Nachweis. Summary-
    `confidence:"binary"` entsteht nur, wenn
    `details.cmaf.binary.status:"passed"` ist. Einzelne bestandene
    Segmentnachweise dürfen eigene `signals[].confidence:"binary"`
    tragen, erhöhen die Summary-Confidence aber nicht auf `binary`,
    solange der Binary-Gesamtstatus `failed` oder `skipped` ist.
    Fehlgeschlagene oder übersprungene Binary-Prüfungen werden in
    `binary.status`, `segmentsChecked[]` und `failures[]` sichtbar,
    erhöhen aber nicht positiv die Summary-Confidence.
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
    Pflichtprüfungsmenge aus dem Manifest-Scope gebildet; danach cappt
    `maxBinarySegments` die tatsächlichen Fetches. Wegen dieses Caps
    nicht geladene weitere Pflichtprüfungen werden als
    `segmentsChecked[].status:"skipped"` mit Failure-Code
    `not_planned_due_to_limit` und zusätzlich in `limits` auditierbar.
    Sie verhindern `passed`, weil `passed` nur bedeuten darf: alle
    manifestseitig verpflichtenden Prüfungen im definierten Analyzer-
    Scope wurden tatsächlich ausgeführt und bestanden.
    Danach gilt: Gesamtstatus `failed`, sobald irgendeine geplante und
    geladene Pflichtprüfung fehlschlägt; `passed` nur, wenn alle
    Pflichtprüfungen im Manifest-Scope bestanden wurden; `skipped`,
    wenn keine Pflichtprüfung fehlgeschlagen ist, aber mindestens eine
    Pflichtprüfung wegen fehlender, nicht sicher auflösbarer, per Byte-
    Range ausgeschlossener, durch `maxBinarySegments` nicht geplanter
    oder beim Fetch/Read durch Limits blockierter Segment-Referenz nicht
    ausgeführt wurde. `segmentsChecked[]` trägt den jeweiligen
    Einzelstatus, damit gemischte DASH-AdaptationSet-Ergebnisse
    auditierbar bleiben.
    `limits` serialisiert mindestens `maxSegmentBytes`,
    `maxBinarySegments`, `timeoutMs`, `maxRedirects`,
    `requiredSegmentChecks` und `plannedSegmentFetches`, damit
    ausgelassene Prüfungen reproduzierbar bleiben.
    `segmentsChecked[]`-Einträge haben mindestens:
    `kind:"init"|"media"`, `source:"hls"|"dash"`, `manifestAnchor`,
    `uri?`, `resolvedUrl?`, `status:"passed"|"failed"|"skipped"`,
    `failureCode?`, `message?`, `contentType?`, `bytesRead?` und
    `boxes[]` mit Box-Ankern. `boxes[]` auf Binary-Ebene sammelt die
    eindeutigen Box-Nachweise mit mindestens `segmentAnchor`, `path`
    (z. B. `moof/traf/tfdt` oder `mdat`), `type`, `offset` und `size`.
    `failures[]` hat mindestens `code`, `level`, `message`,
    `manifestAnchor?`, `segmentAnchor?` und `boxPath?`. Diese Feldnamen
    sind Contract-Bestandteil und werden in Fixtures gepinnt.
  - `note?: string` darf knapp beschreiben, welcher Anteil nur
    manifestbasiert und welcher Anteil binär verifiziert wurde; Pflicht
    ist diese Klarstellung in Doku und README, nicht in jedem JSON-
    Result.
- [x] `packages/stream-analyzer/src/types/input.ts` ergänzt eine
  additive Optionssektion für die binäre CMAF-Prüfung, z. B.
  `cmaf.binary.enabled`, `cmaf.binary.maxSegmentBytes` und
  `cmaf.binary.maxBinarySegments`. Defaults sind dokumentiert:
  `enabled:true`, `maxSegmentBytes=2_000_000`,
  `maxBinarySegments=6`. Die `maxBinarySegments`-Defaultgrenze deckt
  bewusst höchstens drei DASH-AdaptationSets mit je Init- und Media-
  Prüfung ab; größere Pflichtprüfmengen bleiben auditierbar, verhindern
  aber `binary.status:"passed"`, sofern der Aufrufer das Limit nicht
  erhöht. Diese Limits gelten zusätzlich zu `fetch.maxBytes`, das
  ausschließlich das Manifest-Body-Limit bleibt.
- [x] `AnalyzeOptions.fetch`-Semantik ist in
  `packages/stream-analyzer/src/types/input.ts`,
  `docs/user/stream-analyzer.md` und
  `packages/stream-analyzer/README.md` synchronisiert: Timeout,
  Redirect- und SSRF-Optionen gelten für URL-Manifeste und für binäre
  Segment-Fetches aus Text-Inputs mit sicherer HTTP(S)-`baseUrl`;
  `fetch.maxBytes` bleibt ausschließlich das Manifest-Body-Limit und
  wird nicht als Segment-Byte-Limit verwendet.
- [x] Binary-Status- und Failure-Code-Vertrag ist vor Parser-/Loader-
  Implementierung festgelegt: typisierte Domäne `CmafFailureCode` in
  `packages/stream-analyzer/src/types/result.ts` mit Doku-Block
  Präzedenz (1 Caller-Optionen → 2 Manifest-Scope → 3 Planungs-Cap →
  4 Base-URL-Auflösung → 5 Fetch-Grenzen). Sichtbarkeit in
  Fixtures/Testnamen folgt mit den Tranchen 2/3/4 und ist im
  Fixture-Inventar (`spec/contract-fixtures/analyzer/README.md`,
  T0) bereits namentlich gepinnt:
  - `binary_disabled`: `skipped`, wenn Binary-Prüfung per Option
    deaktiviert ist, aber `details.cmaf` in einem binär prüfbaren
    Detail-Scope vorhanden ist (HLS Media-Playlist oder DASH-MPD; HLS-
    Master-Summaries tragen kein `binary`-Objekt).
  - `segment_base_url_missing`: `skipped`, wenn Text-Input keine sichere
    HTTP(S)-`baseUrl` als Trust-Anker für manifestseitig vorhandene
    Init-/Media-Segment-Referenzen liefert. Das gilt für relative und
    absolute Segment-URIs im Text-Manifest; absolute URLs im Manifest
    ersetzen die fehlende `baseUrl` nicht.
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
  - `hls_media_byterange_unsupported`: `skipped` für HLS-
    Media-Segmente mit `#EXT-X-BYTERANGE`, weil `0.10.0` keine HTTP-
    Range-Requests ausführt und nicht die ganze Ressource als Segment
    fehlinterpretieren darf.
  - `dash_template_unresolved`: `skipped`, wenn DASH-Template-Variablen
    in `0.10.0` nicht deterministisch auflösbar sind.
  - `segment_fetch_failed`: `skipped` bei Segment-Fetch-Timeout oder
    HTTP-/Transportfehler vor erfolgreichem Body-Read.
  - `segment_content_type_unsupported`: `skipped`, wenn der Segment-
    Content-Type nicht MP4-/Byte-kompatibel ist.
  - `segment_too_large`: `skipped`, wenn ein Segment `maxSegmentBytes`
    oder das Body-Read-Limit überschreitet.
  - `not_planned_due_to_limit`: `skipped`, wenn eine manifestseitig
    verpflichtende Init-/Media-Prüfung wegen `maxBinarySegments` nicht
    mehr geplant wird. Der Code wird unmittelbar nach Bildung der
    Pflichtprüfungsmenge vergeben, noch vor Base-URL-/URI-Auflösung, und
    gewinnt für überzählige Checks gegenüber späteren URI- oder Fetch-
    Ursachen.
  - `cmaf_box_validation_failed`: `failed`, wenn ein geladenes Init-/
    Media-Segment fachlich nicht konforme Box-/Brand-Struktur hat.
  - `invalid_box_structure`: `failed`, wenn Box-Größe, Überlappung oder
    Parser-Fortschritt strukturell ungültig sind.

  `failed` ist damit für geladene, aber fachlich nicht konforme oder
  strukturell kaputte Bytes reserviert. `skipped` bedeutet, dass der
  Analyzer im sicheren/bounded Scope keine binäre Konformitätsaussage
  treffen konnte. Die `segmentsChecked[]`-Einträge tragen dieselben
  Codes; der Gesamtstatus aggregiert nach Fehler vor Skip vor Pass.
  Bei mehreren möglichen Skip-Ursachen gilt deterministisch diese
  Präzedenz:
  1. Caller-/Options-Entscheidung: `binary_disabled`.
  2. Manifest-Scope fehlt oder ist nicht ableitbar:
     `segment_reference_missing`, `dash_template_unresolved`,
     `hls_map_byterange_unsupported`,
     `hls_media_byterange_unsupported`.
  3. Planungs-Cap nach gebildeter Pflichtprüfungsmenge:
     `not_planned_due_to_limit`.
  4. Base-URL-/URI-Sicherheitsauflösung:
     `segment_base_url_missing`, `segment_uri_blocked`.
  5. Fetch-/Read-Grenzen nach sicherer Auflösung:
     `segment_fetch_failed`, `segment_content_type_unsupported`,
     `segment_too_large`.
  Damit erzeugt z. B. DASH MP4-MIME-only ohne Initialization-/Media-
  Referenzen auch bei Text-Input ohne `baseUrl`
  `segment_reference_missing`, während ein manifestseitig vorhandenes
  Segment ohne sichere `baseUrl` zu `segment_base_url_missing` führt,
  auch wenn die Segment-URI im Text-Manifest absolut notiert ist.
- [x] Public API exportiert die neuen CMAF-Typen über
  `packages/stream-analyzer/src/index.ts` (`CmafAnalyzeOptions`,
  `CmafBinaryOptions`, `CmafBinaryVerification`, `CmafBoxAnchor`,
  `CmafFailure`, `CmafFailureCode`, `CmafLimits`, `CmafSegmentCheck`,
  `CmafSignal`, `CmafSignalSummary`).
- [x] Options-Wire-Vertrag ist festgelegt: `cmaf.binary.*` ist Public-
  TypeScript-API und wird vom analyzer-service als optionales
  Request-Feld akzeptiert, typ-/range-gefiltert und an
  `analyzeManifest` weitergereicht. `apps/api` nutzt in `0.10.0`
  standardmäßig die Analyzer-Defaults, solange der öffentliche
  `/api/analyze`-Request diese Optionen nicht end-to-end modelliert.
  Wichtig: Der API-Pfad darf einen vorhandenen `cmaf`-/`cmaf.binary`-
  Request-Block nicht still ignorieren. Entweder werden HTTP-Request-
  Schema, Domain-Request, driven Adapter, öffentliche Doku und Tests in
  derselben Tranche erweitert und die Optionen durchgereicht, oder
  `/api/analyze` lehnt Requests mit `cmaf`-/`cmaf.binary`-Block
  explizit mit `400 invalid_request` ab. Stiller Fallback auf Defaults
  ist verboten, weil sonst z. B. caller-seitig gesetztes
  `enabled:false` ignoriert und trotzdem Segment-Fetches ausgelöst
  werden könnten.
- [x] `packages/stream-analyzer/scripts/public-api.snapshot.txt` ist
  synchron aktualisiert; der Stream-Analyzer-Public-API-Check im
  Paket-`lint` bleibt grün. `make generated-drift-check` bleibt in
  T1 unverändert — der Stream-Analyzer-Snapshot ist bereits Teil
  des Paket-Lints (`scripts/check-public-api.mjs` läuft im
  `generated-drift-check`-Block) und braucht keine separate
  Drift-Meldung; Makefile-Erweiterungen für CMAF-Fixture-Sync folgen
  in T2/T3 zusammen mit den Fixture-Updates.
- [x] Bestehende HLS-Result-Fixtures ohne CMAF-Signale bleiben
  byte-kompatibel (in T1 unverändert; T2 verifiziert dies erneut beim
  Ausbau der HLS-CMAF-Erkennung). Bestehende DASH-Result-Fixtures mit
  MP4-MIME werden in T3 zusammen mit dem DASH-Parser-Output additiv
  um `details.cmaf` erweitert; in T1 bleibt die Fixture-Form
  unverändert, weil der Parser noch keine CMAF-Signale emittiert und
  der TS-Contract-Test sonst byte-equal bricht.
- [x] Contract-Fixtures in `spec/contract-fixtures/analyzer/` und
  Go-Testdata-Kopien für API-Adapter ergänzt: Die 16 CMAF-Fixtures aus
  dem Inventarvertrag sind als Spec-Quelle angelegt und werden als
  `contract-*.json` nach
  `apps/api/adapters/driven/streamanalyzer/testdata/` synchronisiert.
- [x] `Makefile`-Fixture-Sync ist synchron erweitert:
  `sync-contract-fixtures` kopiert alle neuen Analyzer-CMAF-Fixtures in
  `apps/api/adapters/driven/streamanalyzer/testdata/`,
  `generated-drift-check` diffed diese neuen Kopien, die
  Drift-Fehlermeldung nennt `make sync-contract-fixtures` als Fix, und
  `spec/contract-fixtures/analyzer/README.md` beschreibt die neuen
  Fixtures. Die kopierte Fixture-Anzahl in der Make-Ausgabe ist auf
  22 Gesamt-Fixtures angepasst.
- [x] Go-Adapter-Kontrakt ist explizit geprüft: weil `cmaf` in
  `details` liegt, reichen `apps/api/adapters/driven/streamanalyzer`
  und `apps/api/adapters/driving/http` die Signale über
  `EncodedDetails`/`details` unverändert durch (Domain-Modell
  `StreamAnalysisResult.EncodedDetails` ist `[]byte` mit JSON-
  Roundtrip; `analyze.go` mappt es auf `json.RawMessage` ohne Feld-
  Filter). `cmaf`-/`cmaf.binary`-Top-Level-Block im Request wird mit
  `400 invalid_request` abgelehnt, damit caller-seitig gesetztes
  `enabled:false` nicht still verworfen wird (Test
  `TestAnalyzeHandler_RejectsCmafOptionsBlock`).
- [x] Backward-Compatibility-Notiz in Stream-Analyzer-README
  (`packages/stream-analyzer/README.md` §Scope): bestehende
  `analyzerKind:"hls"`/`"dash"` bleiben unverändert; CMAF ist ein
  Signal, kein dritter Manifesttyp.
- [x] Bestehende Forward-Compat-Hinweise in
  `docs/user/stream-analyzer.md`, `packages/stream-analyzer/README.md`,
  `packages/stream-analyzer/src/types/result.ts` und
  `apps/api/hexagon/domain/stream_analysis.go` sind synchronisiert:
  CMAF wird in `0.10.0` als Signalmodell beschrieben, nicht als neuer
  `analyzerKind`.

---

## 4. Tranche 2 — HLS/fMP4-CMAF-Erkennung

Ziel: HLS-Manifeste mit CMAF/fMP4-Struktur werden zuverlässig erkannt.

DoD:

- [x] Media-Playlist-Parser erkennt `EXT-X-MAP` als starkes
  CMAF/fMP4-Signal und schreibt es nach
  `MediaPlaylistDetails.cmaf.signals[]` (`hls_ext_x_map` mit
  `confidence:"manifest"`).
- [x] Media-Playlist-Parser extrahiert `EXT-X-MAP` strukturiert:
  `URI`, optional `BYTERANGE`, Manifestanker (`media:line:N`), rohe
  Attributwerte und gegen `baseUrl` aufgelöste URI. Liegt intern auf
  `MediaParseResult.cmafMeta.initSegment` (Tranche-4-Eingabe ohne
  Re-Tokenisierung).
- [x] Media-Playlist-Parser extrahiert `#EXT-X-BYTERANGE` strukturiert
  und bindet den Wert deterministisch an die darauffolgende
  Segment-URI; rohe Byte-Range-Angabe, Length/Offset, Manifestanker
  der Tag-Zeile und Segmentanker liegen intern auf
  `MediaParseResult.cmafMeta.firstMediaSegment.byterange` und in
  `SegmentDraft.byterange*`. Verwaiste BYTERANGE → Finding
  `media_byterange_orphan`; Duplikat → `media_byterange_duplicate`;
  unparsbar → `media_byterange_malformed`.
- [x] `EXT-X-MAP` mit `BYTERANGE` und `#EXT-X-BYTERANGE`-Bindung
  bleiben sichtbar (Manifestsignal `hls_ext_x_map_byterange` und
  strukturierte Daten in `cmafMeta`); der binäre Skip-Pfad mit
  Failure-Codes `hls_map_byterange_unsupported` und
  `hls_media_byterange_unsupported` greift in T4, sobald der
  Binary-Loader läuft (kein Vollressourcen-Download).
- [x] Segment-URI-Muster `.m4s`/`.cmfv`/`.cmfa` werden als
  schwächere manifestbasierte Hinweise erfasst (`isFmp4SegmentUri`
  in `cmaf-hls.ts`; Query-/Fragment-Suffixe werden vor dem Match
  abgeschnitten). Erstes Treffer-Segment wird in `signals[]` mit
  `hls_segment_extension_fmp4` gepinnt; `confidence:"manifest"`,
  wenn parallel ein `EXT-X-MAP`-Signal vorliegt, sonst
  `"inferred"`.
- [x] `EXT-X-INDEPENDENT-SEGMENTS` als zusätzliches Signal
  (`hls_independent_segments`, `confidence:"inferred"`) — nur, wenn
  bereits ein anderes CMAF-Signal vorhanden ist (sonst Falsch-
  Positiv-Risiko bei klassischen TS-Manifesten mit Tag).
  `CODECS`/Map-Kontext erzeugen kein eigenes Signal.
- [x] Master-Playlist-Parser schreibt ein konservatives
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
- [x] Tests decken positive, negative und gemischte HLS-Fälle ab
  (`packages/stream-analyzer/tests/parser-media-cmaf.test.ts`,
  `parser-master-cmaf.test.ts`): EXT-X-MAP mit/ohne BYTERANGE,
  fMP4-Suffix-only inferred, TS-Negativ ohne `cmaf`, Query-/
  Fragment-Suffix-Filter, INDEPENDENT-SEGMENTS-Audit-Signal,
  `media_byterange_orphan`/`malformed`/`duplicate`-Findings.
- [x] HLS-Master-Negativfixture pinnt eine Master-Playlist mit
  `CODECS` und `.m3u8`-Variant-URIs ohne fMP4-Suffix; daraus
  entsteht kein `MasterPlaylistDetails.cmaf`
  (`parser-master-cmaf.test.ts` „does not emit cmaf for a CODECS-
  only TS master").
- [x] Bestehender HLS-Master-/Media-Pfad bleibt grün (232 Bestands-
  Tests + 15 neue CMAF-Tests = 247 Tests grün; Lint inklusive
  Boundary- und Public-API-Snapshot grün).

---

## 5. Tranche 3 — DASH/CMAF-Erkennung

Ziel: DASH-MPDs mit CMAF-kompatiblen Representation-/Segment-
Informationen werden als solche ausgewiesen.

DoD:

- [x] DASH-Parser wertet `mimeType` `video/mp4`, `audio/mp4` und
  `application/mp4` als CMAF-relevante Indizien (Signale
  `dash_mime_video_mp4` / `dash_mime_audio_mp4` /
  `dash_mime_application_mp4`); ohne weitere Init/Media-Referenzen
  gelten sie nur als `confidence:"inferred"`.
- [x] `SegmentTemplate@initialization`, `SegmentTemplate@media`,
  `SegmentList/Initialization@sourceURL`, `SegmentList/SegmentURL@media`
  und `Representation`-Codecs fließen in die Signalbewertung ein
  (Signale `dash_segment_template_initialization` /
  `dash_segment_template_media` / `dash_segment_list_initialization` /
  `dash_segment_list_media` / `dash_segment_extension_fmp4`).
- [x] DASH-Schema und Parser erfassen Initialization-Informationen
  explizit und vererbungsbewusst auf `MPD`/`Period`/`AdaptationSet`/
  `Representation`-Ebene (interne `DashCmafMetadata.representations`
  pro AdaptationSet, plus `details.cmaf.signals[]`-Anker im Public
  Schema). Manifestanker im Format
  `MPD/Period[<id|idx>]/AdaptationSet[<id|idx>]/Representation[<id|idx>]/<source>@<attribute>`;
  fehlende IDs erzeugen Index-Anker (`buildDashAnchor`), Test
  „nutzt Index, wenn Period/AdaptationSet/Representation keine ID
  tragen".
- [x] DASH-Parser-Strategie ist vor Umsetzung festgelegt: **Variante
  A — bestehender regex-basierter Parser wird erweitert; keine neue
  XML-Library-Dependency.** Begründung: (1) Alle in `0.10.0` benötigten
  Strukturen (`BaseURL`-Text, `SegmentTemplate@*`-Attribute,
  `SegmentList/Initialization@sourceURL`, `SegmentURL@media`) sind
  attribut- oder text-basiert und in regulärer XML-Form vorhersagbar;
  (2) der Plan §0.1-Scope schließt `$Time$`-Variablen und
  `SegmentTimeline`-Auflösung explizit aus, also keine komplexe
  Mehrfach-Verschachtelung nötig; (3) keine zusätzliche Audit-/
  CVE-Surface durch externe XML-Parser. Die Vererbung wird **nicht**
  implizit durch verschachtelte Regex-Matches abgebildet, sondern
  explizit als Chain pro Ebene aufgebaut (siehe nächstes DoD-Item).
  Migration auf `fast-xml-parser` bleibt Folge-Plan-Item, wenn echte
  Live-/Edge-Cases (z. B. SegmentTimeline) den Scope erweitern müssen.
- [x] DASH-Template-Auflösungs-Scope ist fixiert (`resolveDashTemplate`
  in `cmaf-dash.ts`): `$RepresentationID$`, `$Bandwidth$`, `$Number$`,
  `$Number%0Nd$`, `$Bandwidth%0Nd$` und `$$` (Literal). `startNumber`
  aus `SegmentTemplate@startNumber`, fehlt → Default `1`. `$Time$`
  und unbekannte Variablen liefern `null`; im internen
  `DashCmafSegmentRef.templateUnresolved=true` und im Public
  `signals[]` wird das `media`-Manifest-Signal in dem Fall
  weggelassen — Tranche 4 mappt das auf `dash_template_unresolved`.
  Manifestbasierte Initialization-Signale bleiben trotzdem sichtbar
  (Test „$Time$ ist nicht aufgelöst und führt zu kein Manifest-Signal
  für media").
- [x] DASH-`BaseURL`-Auflösung deterministisch (`resolveBaseUrlChain`
  in `cmaf-dash.ts`): pro Ebene wird der erste sichere Eintrag in
  Manifest-Reihenfolge gewählt; relative Werte gegen die bereits
  geerbte Base aufgelöst; absolute Werte nur akzeptiert, wenn das
  Schema `http:`/`https:` ist. Wenn alle Kandidaten unsicher sind,
  wird die geerbte Base **nicht** durchgereicht und `blocked=true`
  gesetzt — Tranche 4 mappt das auf `segment_uri_blocked`. Tests
  pinnen Vererbung über alle vier Ebenen, erste-sichere-Wahl,
  `file://`-Block und Sub-Level-Block.
- [x] DASH-Repräsentationsauswahl für den Binary-Pfad: pro
  AdaptationSet wird in Manifest-Reihenfolge die erste Representation
  mit init+media-Referenz gewählt (`chosenRepEntry` in
  `parseAdaptationSet`); fällt das durch (z. B. nur Init geerbt),
  greift die erste Representation mit Signal überhaupt. Anker
  zeigt auf die gewählte Representation (Test
  „erbt Initialization von AdaptationSet auf Representation").
  `requiredSegmentChecks` und Pflichtprüfungs-Anker werden in
  Tranche 4 aus `cmafMeta.representations[].init/media` abgeleitet.
- [x] Confidence-Regeln getestet: MP4-MIME allein → `inferred`;
  Initialization plus fMP4-Segmentmuster → `manifest`. (`describe`
  "DASH-CMAF — Confidence-Regeln (RAK-62)", drei Cases).
- [x] DASH-Tests pinnen drei getrennte Fälle (siehe oben).
- [x] DASH-Tests pinnen Vererbung und URI-Auflösung getrennt
  (`describe` "DASH-CMAF — BaseURL-Vererbung", "SegmentTemplate-
  Vererbung und Override", "SegmentList-Pfad", "Mehrperiodige
  Manifestanker mit Index-Fallback").
- [x] DASH-Live- und VOD-Fixtures behalten bestehende Mindestfelder
  aus RAK-58 — additive `details.cmaf`-Erweiterung in
  `success-dash-vod.json`/`success-dash-live.json` (Go-testdata-
  Kopien via `make sync-contract-fixtures` synchronisiert).
- [x] Tests decken DASH-CMAF positiv, DASH ohne Initialization-Signal
  (MP4-MIME-only) und fehlende Variablen-Auflösung ab; Negativ-Pfad
  ohne MP4-MIME/Initialization/fMP4-URI emittiert kein `cmaf`.

---

## 6. Tranche 4 — Binäre CMAF-Konformitätsprüfung

Ziel: CMAF wird nicht nur als Manifest-Indiz erkannt, sondern im
begrenzten Analyzer-Scope über echte Segment-/Box-Daten geprüft.

DoD:

- [x] ISO-BMFF-Box-Reader implementiert (`internal/cmaf/iso-bmff.ts`):
  - liest 32-bit und `largesize`-Boxgrößen bounds-checkend
    (`size=0`/extends-to-end-of-file ist im 0.10.0-Scope explizit
    `invalid_box_structure`),
  - erkennt ungültige Box-Größen, Boxen die den Buffer überrunnen,
    nicht fortschreitende Boxen und non-ASCII-Types,
  - bricht bei jedem strukturellen Verstoß deterministisch ab (kein
    Recursion-Bomb), liefert Teilergebnis bis zum Fehler,
  - liefert stabile Box-Anker (`segment:init:ftyp`,
    `segment:media[0]:styp`, `segment:media[0]:moof/traf/tfdt`,
    `segment:media[0]:mdat`) für `details.cmaf.binary.boxes[]`.
- [x] Bounded Binary-Segment-Loader implementiert getrennt von
  `loadManifest` (`internal/cmaf/segment-loader.ts`):
  - gibt Bytes zurück, nicht UTF-8-Text,
  - sendet einen MP4-orientierten `Accept`-Header,
  - erlaubt mindestens `video/mp4`, `audio/mp4`, `application/mp4`,
    `video/iso.segment`, `audio/iso.segment`,
    `application/iso.segment`, `application/octet-stream` und leeren
    Content-Type,
  - nutzt dieselbe URL-Validierung, DNS-/SSRF-Prüfung,
    Redirect-Behandlung und Timeout-Mechanik wie der Manifest-Loader,
  - erzwingt `maxSegmentBytes` pro Segment und
    `maxBinarySegments` über den gesamten Analyseaufruf,
  - mappt blockierte, zu große, nicht ladbare oder Content-Type-
    inkompatible Segment-Fetches auf auditierbare
    `segmentsChecked[]`-/`failures[]`-Einträge statt auf einen
    Top-Level-Analysefehler.
- [x] CMAF-Init-Prüfung validiert mindestens:
  - `ftyp` vorhanden,
  - Brand-Policy ist als getestete Allowlist umgesetzt: `ftyp.major_brand`
    oder mindestens ein Eintrag aus `ftyp.compatible_brands` muss `cmfc`
    oder `cmf2` sein. `cmfc` und `cmf2` sind die in `0.10.0`
    unterstützten CMAF-Header-/Track-Structural-Brands. `cmfs` ist ein
    CMAF-Segment-Brand für `styp` und darf im Init-`ftyp` nicht als
    Header-Konformitätsnachweis zählen. Generische MP4-/ISO-BMFF-Brands
    wie `isom`, `iso6`, `mp41` oder `mp42` dürfen zusätzlich vorkommen,
    reichen aber allein nicht für `status:"passed"`. Neuere
    Structural-Brand-Profile wie `cmf1` bleiben in `0.10.0`
    `cmaf_box_validation_failed`, bis der Projekt-Scope sie explizit
    aufnimmt. Fixtures pinnen mindestens diese Fälle: `cmfc` als
    `major_brand`, `cmf2` nur in `compatible_brands`, `cmfs` nur in
    `compatible_brands` als negativer Init-Fall, ausschließlich
    generische Brands, `cmf1` ohne expliziten Support und fehlendes
    `ftyp`,
  - `moov` vorhanden,
  - keine offensichtlich widersprüchliche Top-Level-Box-Struktur.
- [x] CMAF-Media-Fragment-Prüfung validiert mindestens:
  - `styp` vorhanden,
  - `styp.major_brand` oder mindestens ein Eintrag aus
    `styp.compatible_brands` trägt einen in `0.10.0` unterstützten
    CMAF-Media-Brand: `cmfs`, `cmff`, `cmfc` oder `cmf2`. Generische
    MP4-/DASH-Brands reichen allein nicht für `status:"passed"`;
    `cmfl`/chunked CMAF bleibt wegen des Low-Latency-Out-of-Scope
    bewusst Folge-Scope,
  - `moof` vorhanden,
  - mindestens ein `traf` unter `moof`,
  - `tfdt` unter `traf` vorhanden,
  - `mdat` als Media-Data-Box vorhanden, damit `status:"passed"` nicht
    allein auf Fragment-Metadaten ohne Nutzdaten basiert,
  - optionales `sidx` wird erkannt und berichtet, aber nicht als Muss
    gewertet.
- [x] Segment-Resolver lädt nur manifestreferenzierte Init-/Media-
  Segment-URIs (`internal/cmaf/binary-verify.ts` plus Parser-cmafMeta-
  Einschluss). URL-Input löst Segment-URIs gegen die finale Manifest-
  URL auf (`AnalysisInputMetadata.baseUrl` aus `loadManifest.finalUrl`);
  Text-Input nutzt die gesetzte sichere `http:`-/`https:`-`baseUrl`.
  Fehlende `baseUrl` → `segment_base_url_missing`; unsichere
  Auflösung → `segment_uri_blocked`; Block-Vererbung über
  `resolveBaseUrlChain` mit `parentBlocked`-Flag.
- [x] HLS-Binary-Pfad prüft `EXT-X-MAP` plus erstes fMP4-Media-Segment
  (`buildHlsPlan` / `buildHlsInitCheck` / `buildHlsMediaCheck` in
  `binary-verify.ts`). Strukturierte `HlsExtXMapMeta`/
  `HlsFirstMediaSegmentMeta` aus der T2-Parser-Ausgabe werden
  unverändert konsumiert; kein Re-Parsen. `EXT-X-MAP` mit `BYTERANGE`
  → `hls_map_byterange_unsupported`, `#EXT-X-BYTERANGE` auf erstem
  fMP4-Segment → `hls_media_byterange_unsupported` (skipped, kein
  Vollressourcen-Download). Beide Skip-Codes durch Tests gepinnt.
- [x] DASH-Binary-Pfad prüft Initialization plus erstes ableitbares
  fMP4-Media-Segment je deterministisch ausgewählter Representation
  aus Tranche 3. `dash_template_unresolved` für `$Time$`/unbekannte
  Variablen, `segment_uri_blocked` bei Block-Vererbung,
  `segment_base_url_missing` ohne sicheren Trust-Anker,
  `segment_reference_missing` für DASH-MP4-MIME-only ohne Init-/Media-
  Referenzen — alle vier Codes durch Tests gepinnt. Ableitbar sind nur explizite
  `SegmentList/SegmentURL@media`-Referenzen oder Templates aus dem
  Scope von Tranche 3:
  `$RepresentationID$`, `$Bandwidth$`, `$Number$` und
  `$Number%0Nd$` mit `startNumber` bzw. Default `1`. `$Time$`,
  `SegmentTimeline`-abhängige Auflösung oder unbekannte Variablen werden
  nicht geraten, sondern als `skipped` mit Failure-Code
  `dash_template_unresolved` gemeldet. `BaseURL` wird nach der
  erste-sichere-`BaseURL`-Regel aus Tranche 3 aufgelöst; mehrere
  Alternativen werden nicht durchprobiert. Fehlt bei `SegmentList` die
  Media-Referenz vollständig, wird `segment_reference_missing` gemeldet.
- [x] Positive und negative Binär-Fixtures decken Init, Media, fehlende
  Pflichtboxen inklusive `styp` und `mdat`, kaputte Boxgrößen,
  Größenlimit und nicht auflösbare Segment-URIs ab — programmatisch
  in `tests/iso-bmff.test.ts` und `tests/binary-verify.test.ts`
  aufgebaut (deterministische Bytes-Fixtures via
  `makeBox`/`brandPayload`/`concat`-Helpers, keine externen Datei-
  Fixtures, keine Netzwerk-Abhängigkeit).
- [x] Tests pinnen Status-Mapping:
  - `passed` nur bei bestandenen manifestseitig verpflichtenden Init-
    und Media-Prüfungen,
  - `failed` bei geladener, aber nicht konformer Box-Struktur,
  - `skipped` bei fehlender oder aus Sicherheits-/Scope-Gründen nicht
    ladbarer Segmentreferenz,
  - DASH-MP4-MIME-only ohne Initialization-/Media-Referenzen erzeugt
    bei Default `cmaf.binary.enabled:true`
    `details.cmaf.binary.status:"skipped"` mit
    `segment_reference_missing`, nicht ein fehlendes `binary`-Objekt,
  - `binary_disabled`, `segment_base_url_missing`,
    `segment_uri_blocked`, `segment_reference_missing`,
    `hls_map_byterange_unsupported`,
    `hls_media_byterange_unsupported`,
    `dash_template_unresolved`, `segment_fetch_failed`,
    `segment_content_type_unsupported`, `segment_too_large`,
    `not_planned_due_to_limit`, `cmaf_box_validation_failed` und
    `invalid_box_structure` nach der Status-/Failure-Code-Tabelle aus
    Tranche 1,
  - `maxBinarySegments` cappt die Fetch-Menge nach Bildung der
    manifestseitigen Pflichtprüfungsmenge; nicht gefetchte weitere
    Pflichtprüfungen werden als `skipped` mit
    `not_planned_due_to_limit` berichtet, verhindern `passed` und sind
    über `limits`/`note` auditierbar,
  - gemischte DASH-Ergebnisse aggregieren nach der Regel aus Tranche 1:
    jeder Fehler gewinnt vor `skipped`, `skipped` gewinnt vor `passed`.
- [x] Fehler aus binärer Prüfung bleiben Findings oder
  `details.cmaf.binary.failures[]`; sie ändern nicht das bestehende
  `status:"ok"` des Analyse-Results, solange das Manifest selbst
  erfolgreich analysiert wurde.

---

## 7. Tranche 5 — API, CLI und Doku

Ziel: CMAF-Signale sind über alle bestehenden Analyzer-Pfade nutzbar.

DoD:

- [x] `apps/api`-StreamAnalyzer-Adapter reicht `details.cmaf` über
  `EncodedDetails` additiv durch (Test
  `TestHTTPStreamAnalyzer_ContractDashVodCMAFBinarySkipped` in
  `apps/api/adapters/driven/streamanalyzer/contract_test.go`
  decoded das vollständige `cmaf.binary`-Subobjekt aus der
  Spec-Fixture). HTTP-Wire-Nachweis für `analysis.details.cmaf.
  binary.status` im `/api/analyze`-Wrapper läuft über
  `TestAnalyzeHandler_PassesCmafBinaryThroughEncodedDetails`.
- [x] `apps/analyzer-service` akzeptiert `cmaf.binary.{enabled,
  maxSegmentBytes,maxBinarySegments}` und filtert ungültige Werte
  analog zum bestehenden `fetch`-Block (T1-Commit `441c4bb`,
  `parseCmafOptions` in `src/server.ts`); `allowPrivateNetworks`
  bleibt env-only über `ANALYZER_ALLOW_PRIVATE_NETWORKS`.
- [x] `/api/analyze`-Vertrag ist explizit für die Reject-Variante
  entschieden: Requests mit `cmaf`-/`cmaf.binary`-Block werden mit
  `400 invalid_request` abgelehnt (T1-Commit `441c4bb`,
  `analyze.go` Pre-Decode-Check + Test
  `TestAnalyzeHandler_RejectsCmafOptionsBlock`). Begründung und
  Caller-Steuerung über die Library-Optionen in
  `docs/user/stream-analyzer.md` §3.1.
- [x] HTTP-Contract-/Adapter-Tests decken HLS- und DASH-Wire-Pfad
  ab. `TestAnalyzeHandler_PassesCmafBinaryThroughEncodedDetails`
  pinnt die öffentliche `{analysis, session_link}`-Wrapper-Antwort
  mit `analysis.details.cmaf.binary.status`; die existierenden
  Driven-Adapter-Tests pinnen das Decoder-Verhalten gegen die
  Spec-Fixtures (HLS-Master + DASH-VOD inkl. Binary-Subobjekt).
- [x] CLI gibt die neuen Signale unverändert im JSON aus
  (`runCli` schreibt `JSON.stringify(result, null, 2)`; CMAF-Smoke
  in `scripts/smoke-cli.sh` Schritt 8 verifiziert
  `binary.status="passed"` über jq).
- [x] CLI-Opt-in-Schalter `MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS`
  reicht ausschließlich `fetch.allowPrivateNetworks=true` an die
  Library durch; akzeptierte Werte `1`/`true`/`TRUE`/`yes`/`on`
  (case-insensitive, getrimmt). Default off → `fetch_blocked`
  bleibt bei Loopback-URLs sichtbar (Smoke-Schritt 8c). Tests in
  `tests/cli.test.ts` decken Default-Off, alle akzeptierten Werte,
  alle abgelehnten Werte und die Help-Text-Doku.
- [x] `make smoke-cli` um drei CMAF-Probes erweitert: HLS-passed,
  DASH-passed, Loopback-ohne-Opt-in-fetch_blocked. Lokaler
  HTTP-Server via `python3 -u -m http.server 0` aus tmpdir mit
  per `scripts/cmaf-fixture-builder.mjs` deterministisch erzeugten
  Init-/Media-Bytes; Datei-Input mit `file://` ist explizit nicht
  Teil des Smokes, weil der Segment-Loader strikt HTTP(S)-SSRF-
  Regeln nutzt.
- [x] [`docs/user/stream-analyzer.md`](../../../user/stream-analyzer.md)
  §3.1 (CMAF-Binary-Verifikation) und §9.2 (CMAF-Lab-Modus)
  beschreiben Scope, Beispiele, Brand-Allowlist, Defaults, alle 13
  CmafFailureCode-Werte mit Status-Mapping, Caller-Steuerung,
  `/api/analyze`-Vertrag und Opt-in-Env-Variable.
- [x] `packages/stream-analyzer/README.md` synchronisiert (CMAF-
  Block ✅ statt 🟡; Verweis auf §3.1+§9.2 für Details).

---

## 8. Tranche 6 — Gates, Matrix und Release-Closeout

Ziel: `0.10.0` wird als Minor-Release sauber abgeschlossen.

DoD:

- [x] RAK-Verifikationsmatrix vollständig ausgefüllt:

| RAK | Priorität | Nachweis | Status |
| --- | --------- | -------- | ------ |
| RAK-60 | Muss | Lastenheft `1.1.13` Header + §8.3 NF-13 + neuer §13.12 (Commit `b2a8635`); Plan §0.1 Scope-Definition; Folge-Scope explizit ausgegrenzt in §9 + §10. CMAF-Analyse-Scope normativ begrenzt auf manifestbasierte HLS-/DASH-Signale + bounded binäre Init-/Media-Prüfung. | [x] |
| RAK-61 | Muss | `internal/parsers/cmaf-hls.ts` + `media.ts`/`master.ts`-Erweiterung (Commit `14259f9`). Tests: `parser-media-cmaf.test.ts` (11), `parser-master-cmaf.test.ts` (4). Confidence-Regeln gepinnt (EXT-X-MAP `manifest`, fMP4-Suffix-only `inferred`, CODECS-only kein Signal). CLI-Smoke `[smoke-cli] CMAF-HLS probe OK (binary.status=passed)` in `scripts/smoke-cli.sh` Schritt 8a. | [x] |
| RAK-62 | Muss | `internal/parsers/cmaf-dash.ts` + `dash.ts`-Erweiterung (Commit `21354b7`). Tests: `parser-dash-cmaf.test.ts` (11). Drei Confidence-Cases gepinnt (MP4-MIME-only `inferred`, Init+fMP4 `manifest`, kein-Signal kein cmaf); BaseURL-Vererbung über alle vier Ebenen + first-safe; Multi-Period-Index-Anker; Template-Auflösung. Contract-Fixtures `success-dash-vod.json`/`success-dash-live.json` additiv erweitert. CLI-Smoke `[smoke-cli] CMAF-DASH probe OK (binary.status=passed)` in `scripts/smoke-cli.sh` Schritt 8b. | [x] |
| RAK-63 | Muss | API-/CLI-/Doku-Durchleitung (Commits `441c4bb` T1, `c979eb8` T5). HTTP-Wire-Test `TestAnalyzeHandler_PassesCmafBinaryThroughEncodedDetails` pinnt `analysis.details.cmaf.binary.status` im `{analysis, session_link}`-Wrapper; `TestAnalyzeHandler_RejectsCmafOptionsBlock` pinnt 400 invalid_request für `cmaf`-Block; `TestHTTPStreamAnalyzer_ContractDashVodCMAFBinarySkipped` pinnt Driven-Adapter-Wire. CLI-Opt-in `MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS` mit 15 Tests in `tests/cli.test.ts`. Doku in `docs/user/stream-analyzer.md` §3.1+§9.2; README synchronisiert. Contract-Fixtures + Go-testdata via `make sync-contract-fixtures` und Drift-Check. | [x] |
| RAK-64 | Muss | ISO-BMFF-Box-Parser `internal/cmaf/iso-bmff.ts` + bounded Segment-Loader `internal/cmaf/segment-loader.ts` + Verifier-Orchestrator `internal/cmaf/binary-verify.ts` (Commit `1c23e99`). Brand-Allowlist Init `cmfc`/`cmf2` und Media `cmfs`/`cmff`/`cmfc`/`cmf2` durch `iso-bmff.test.ts` (28 Tests) gepinnt; bounded Loader durch `segment-loader.test.ts` (18 Tests); Status-Aggregation und alle 13 `CmafFailureCode`-Werte durch `binary-verify.test.ts` (15 Tests). HTTP-Nachweis: `TestAnalyzeHandler_PassesCmafBinaryThroughEncodedDetails` (Wire-Vertrag); CLI-Nachweis: `make smoke-cli` Schritt 8a/8b/8c (HLS-passed, DASH-passed, Loopback-ohne-Opt-in-fetch_blocked). Confidence-Upgrade auf `binary` nur bei `binary.status:"passed"` getestet. | [x] |

- [x] `make docs-check` grün.
- [x] `make build` grün (über `make gates`).
- [x] `make gates` grün — letzter Run vor Tag log: `gates-t6.log`
  (post-Versions-Bump und post-CHANGELOG).
- [x] `make smoke-cli` grün — alle 11 Schritte inkl. CMAF-HLS-passed,
  CMAF-DASH-passed, Loopback-ohne-Opt-in-fetch_blocked.
- [x] Release-Smokes laut [`docs/user/releasing.md`](../../../user/releasing.md)
  §2 für `0.10.0` gewaved: die CMAF-Funktion lebt vollständig im
  Stream-Analyzer-Pfad (TS-Library + Go-Wire-Adapter + CLI-Smoke),
  ohne neue MediaMTX-/SRT-/SRS-/WebRTC-/DASH-Lab-/Observability-
  Surface oder Compose-Stack-Änderungen. `make smoke-cli` deckt den
  CMAF-end-to-end-Pfad mit Manifest-Loader + Segment-Loader + Box-
  Parser ab; `make smoke-analyzer` ist Bestandteil der
  `make gates`-Pipeline und damit bereits abgesichert. Lab-/Compose-
  Smokes (`make smoke-observability`/`make browser-e2e`/
  `make smoke-mediamtx`/`make smoke-srt`/`make smoke-srt-health`/
  `make smoke-dash`/`make smoke-webrtc-prep`/
  `make smoke-webrtc-stats-drift`/`make smoke-srs`) sind
  unverändert vom `0.9.6`-Stand und betreffen Subsysteme, an
  denen `0.10.0` keine Änderung gemacht hat — analog der
  Closeout-Begründung von `0.9.6` (kein Compose-/Telemetrie-/
  Player-SDK-Wire-Vertrag berührt).
- [x] `make security-gates` grün — Pre-Tag-Lauf log:
  `security-gates-t6.log`.
- [x] Wave-2-Quality-Gates aus
  [`docs/user/releasing.md`](../../../user/releasing.md) §3.1 vor
  dem Tag geprüft: Branch-Coverage Stream-Analyzer 90.4 %
  / API 90.2 % beide >= Threshold; Drift-Check grün
  (`generated-drift-check` in `make gates`); Public-API-Snapshot
  `packages/stream-analyzer/scripts/public-api.snapshot.txt`
  synchron.
- [x] `benchmark.yml`-Nightly-Status: `0.10.0` ändert keine API-
  Hot-Path-/Persistence-Gewichte; CMAF-Pfad ist opt-in und
  nicht im Benchmark-Smoke enthalten. Initiale Beobachtungsphase
  des Wave-2-Benchmark-Smokes bleibt aus `0.9.5`-Closeout aktiv
  (siehe roadmap.md §1.2).
- [x] Kein offenes Crash-Issue mit Label `fuzz` für `0.10.0`-
  Surface (Stream-Analyzer-CMAF-Pfad ist getestet via
  `tests/iso-bmff.test.ts` Negativ-Cases inklusive truncated
  Bytes / falsche Brands / fehlende Pflicht-Boxen / largesize-
  Edge-Cases; bounded Loader rejected ungültige Größen
  deterministisch).
- [x] Mutation-Score-Trend bleibt unverändert; `0.10.0` fügt nur
  additiven Code hinzu, ohne Wave-2-Mutation-Module
  (Cursor-Logik, HLS/DASH-Parser-Mutationen) zu verschlechtern.
- [x] Vollständiger Versions-Bump `0.9.6` → `0.10.0` in allen
  versionsführenden Stellen: 5× `package.json`,
  `apps/api/cmd/api/main.go` `serviceVersion`,
  `packages/player-sdk/src/version.ts`,
  `packages/player-sdk/scripts/pack-smoke.mjs`,
  `contracts/sdk-compat.json`, alle hartkodierten
  Test-Fixture-Versionen sowie die vier
  `spec/contract-fixtures/analyzer/*.json`-Fixtures plus ihre
  Go-testdata-Kopien.
- [x] `CHANGELOG.md` mit `[0.10.0] - 2026-05-09` aktualisiert.
- [x] Roadmap-Status und Release-Übersicht auf `0.10.0` released
  aktualisiert; §1 Phase-Header, §1.2 Folge-Scope, §2 Schritt 45,
  §3 Tabellen-Zeile alle auf `0.10.0` released. Folgephase
  beschreibt nur noch bewusst ausgegrenzte CMAF-Erweiterungen
  (Low-Latency-CMAF, vollständige Segmentset-Abdeckung,
  Codec-Decoding, Player-SDK-CMAF-Support, HTTP-Range-Loader,
  `cmf1`/neuere Brand-Profile).
- [x] Plan nach `docs/planning/done/plan-0.10.0.md` verschoben und
  Status auf ✅ released aktualisiert (Closeout-Commit).
- [x] Annotierter Tag `v0.10.0` erstellt (Closeout-Commit + Tag).

## 9. Nicht-Ziele für Review

Review-Kommentare zu folgenden Themen sollen in Folge-Pläne, nicht in
`0.10.0`:

- Low-Latency-CMAF-Chunks und `#EXT-X-PART`-Analyse.
- Vollständiger Download aller Segmente, CDN-Checks oder Byte-Range-
  Verifikation.
- Player-SDK-CMAF-Playback-Support.
- Neue Storage-, Multi-Tenant- oder Kubernetes-Scope-Erweiterungen.

## 10. Folge-Patch-Trigger (`0.10.1`-Entscheidung)

Beim `0.10.0`-Closeout (Tranche 6) wird **bewusst entschieden**, ob
ein `plan-0.10.1.md` benötigt wird oder ob offene Folgepunkte direkt
in `0.11.0` wandern. `0.10.1` ist nur dann gerechtfertigt, wenn nach
Tag `v0.10.0` einer dieser Trigger eintritt:

- **CVE-/Stdlib-Bump erzwingt Image-Update** (analog `0.8.5`-OTel-
  Bump oder `0.9.6`-Go-Stdlib-Bump): Go-/Node-/Distroless-Basis-Image
  muss aus Sicherheitsgründen kurzfristig hochgezogen werden, ohne
  neuen Minor-Scope.
- **Lastenheft-Audit-Konvergenz nach Release** (analog `0.9.6`-Patch
  nach `0.9.5`): Beim Audit „Ist `NF-13` wirklich erfüllt?" werden
  Lieferstands-Unschärfen sichtbar, die einen Lastenheft-Patch
  brauchen, aber keine neue Produktfunktion.
- **Kleiner CMAF-Bug aus realem Lab-Use** zu klein für einen
  eigenen Minor (z. B. eine Brand-Kombination, die in der Praxis
  vorkommt aber nicht von der `cmfc`/`cmf2`/`cmfs`/`cmff`-Allowlist
  erfasst wird, ohne dass es eine Scope-Erweiterung ist).

Größere Folge-Themen (Low-Latency-CMAF, vollständige Segmentset-
Abdeckung, Player-SDK-CMAF, vollständige `cmf1`-Aufnahme,
Byte-Range-Loader) gehen direkt in `0.11.0`+ und nicht in einen
Patch-Plan. Während der `0.10.0`-Implementierung aufkommende
deferred Tradeoffs gehören als R-N-Eintrag mit Triggerschwelle in
[`docs/planning/in-progress/risks-backlog.md`](../risks-backlog.md),
nicht in einen leeren Patch-Plan-Stub.
