# Analyzer-Wire-Format Contract Fixtures

Geteilte Wahrheit für das JSON-Wire-Format zwischen
`apps/analyzer-service` (TypeScript-Producer) und
`apps/api/adapters/driven/streamanalyzer` (Go-Consumer). Beide Seiten
testen gegen dieselben Dateien:

- `success-master.json` — Erfolgsfall mit `playlistType: "master"`,
  einem Variant und einer Rendition. Pinnt das volle Envelope-Schema
  inklusive `analyzerKind`, `findings`-Form und `details`-Struktur.
- `success-dash-vod.json` — Erfolgsfall mit `analyzerKind: "dash"` /
  `playlistType: "dash"` (VOD-MPD, `type=static`, on-demand-Profile;
  ab `0.9.0` Tranche 3, RAK-58). Pinnt die DASH-Variante des
  Envelopes mit `details.profiles` / `type` / `live` / `periodCount`
  / `adaptationSets`-Hierarchie inklusive Mindest-Felder pro
  `Representation` (`bandwidth`, `width`/`height`, `codecs`,
  `mimeType`).
- `success-dash-live.json` — Erfolgsfall mit `type: "dynamic"` /
  `live: true` (Live-MPD mit `minimumUpdatePeriod` und
  `availabilityStartTime`); pinnt die Live-spezifischen
  Detail-Felder.
- `error-fetch-blocked.json` — Fehlerfall mit `status: "error"`,
  `code: "fetch_blocked"`. Pinnt die Error-Envelope-Form.

## Tests

- TypeScript: `packages/stream-analyzer/tests/contract.test.ts` —
  speist eine bekannte Manifest-Eingabe in `analyzeHlsManifest`,
  serialisiert das Result und vergleicht byte-equal gegen
  `success-master.json`. Jede TS-Output-Drift bricht den Test.
- Go: `apps/api/adapters/driven/streamanalyzer/contract_test.go` —
  liest beide Dateien per `go:embed`, parst sie via
  `parseSuccessResponse` / `parseDomainError`, und prüft die
  resultierenden Domain-Strukturen feldgenau. Jede Drift, die das
  Go-Decoding bricht, fällt auf.

## Updates

Wenn das Format absichtlich erweitert wird:

1. Code-Änderung committen.
2. TS-Test zeigt den Diff — die Fixture mit `vitest -u` (oder
   manuell) aktualisieren.
3. Go-Test gegen die neue Fixture prüfen.
4. Drift in einem Pflichtfeld (z. B. neuer `analyzerKind`-Wert)
   bedingt synchrone Anpassung beider Seiten — das ist der ganze
   Sinn dieser Fixtures.

## CMAF-Fixture-Inventar (Plan `0.10.0`, Tranche 0)

Der `0.10.0`-Plan sieht folgende Fixtures vor; sie werden nicht
alle in Tranche 0 angelegt, sondern in der Tranche, die das
Code-Pendant liefert (Tranche 2 für HLS, Tranche 3 für DASH,
Tranche 4 für binäre Box-Prüfung). Die Inventartabelle ist
Vertrag, damit `details.cmaf`-Erweiterungen reproduzierbar
geprüft werden.

| Datei (geplant) | Tranche | Was die Fixture pinnt |
| --- | --- | --- |
| `success-hls-cmaf-vod.json` | 2 | HLS Media-Playlist mit `EXT-X-MAP` und `.m4s`-Segmenten; `details.cmaf` mit starkem manifestbasiertem Signal und `binary.status:"passed"`. |
| `success-hls-ts-negative.json` | 2 | HLS Media-Playlist mit `.ts`-Segmenten ohne `EXT-X-MAP` als Negativ-/Regression-Pfad; kein `details.cmaf`. |
| `success-hls-master-codecs-only.json` | 2 | HLS Master-Playlist mit `CODECS` und TS-basierten Variants; daraus darf **kein** `details.cmaf` entstehen. |
| `success-hls-map-byterange.json` | 2 | `EXT-X-MAP` mit `BYTERANGE`; Manifest-Signal sichtbar, `binary.status:"skipped"` mit `hls_map_byterange_unsupported`. |
| `success-hls-media-byterange.json` | 2 | `#EXT-X-BYTERANGE` vor erstem fMP4-Media-Segment; `binary.status:"skipped"` mit `hls_media_byterange_unsupported`. |
| `success-dash-mp4-mime-only.json` | 3 | DASH-MPD nur mit `video/mp4`-/`audio/mp4`-MIME ohne Initialization-/Media-Referenzen; `details.cmaf.confidence:"inferred"` und `binary.status:"skipped"` mit `segment_reference_missing`. |
| `success-dash-cmaf-vod.json` | 3 | DASH-MPD mit `SegmentTemplate@initialization` plus fMP4-Segmentmuster als starker manifestbasierter Pfad; `binary.status:"passed"`. |
| `success-dash-no-cmaf-signals.json` | 3 | DASH-MPD ohne MP4-MIME, ohne Initialization, ohne fMP4-URI-Muster (Negativ-/Regression-Pfad); kein `details.cmaf`. |
| `success-dash-baseurl-inheritance.json` | 3 | `BaseURL`-Vererbung über `MPD`/`Period`/`AdaptationSet`/`Representation`-Ebene plus mehrperiodige Manifestanker mit fehlenden IDs (Index-Anker). |
| `success-dash-segmentlist.json` | 3 | `SegmentList` mit `Initialization@sourceURL` und `SegmentURL@media` als alternativer Auflösungspfad. |
| `error-cmaf-binary-validation.json` | 4 | CMAF-Init-Segment ohne kompatibles `ftyp` oder Media-Segment ohne `mdat`; `binary.status:"failed"` mit `cmaf_box_validation_failed`. |
| `error-cmaf-invalid-box-structure.json` | 4 | Ungültige Box-Größe oder überlappende Boxen; `binary.status:"failed"` mit `invalid_box_structure`. |
| `success-cmaf-skipped-too-large.json` | 4 | Segment über `maxSegmentBytes`; `binary.status:"skipped"` mit `segment_too_large`. |
| `success-cmaf-skipped-content-type.json` | 4 | Segment-Content-Type nicht MP4-/Byte-kompatibel; `binary.status:"skipped"` mit `segment_content_type_unsupported`. |
| `success-cmaf-skipped-binary-disabled.json` | 4 | Caller setzt `cmaf.binary.enabled:false`; `binary.status:"skipped"` mit `binary_disabled` (kein weggelassenes `binary`-Objekt). |
| `success-cmaf-skipped-not-planned.json` | 4 | Mehr verpflichtende Init-/Media-Prüfungen als `maxBinarySegments`; überzählige Checks tragen `not_planned_due_to_limit`. |

Bestehende Fixtures, die in Tranche 3 additiv erweitert werden:

- `success-dash-vod.json` — bekommt zusätzlich `details.cmaf` mit
  `confidence:"inferred"` und `binary.status:"skipped"` mit
  `segment_reference_missing`, weil das Fixture nur MP4-MIME ohne
  Initialization-Referenzen hat. Bewusst nicht byte-kompatibel zum
  `0.9.x`-Stand.
- `success-dash-live.json` — analog, falls die Live-MPD
  CMAF-relevante Signale trägt; sonst bleibt sie unverändert.

Bestehende Fixtures, die unverändert bleiben:

- `success-master.json` (HLS Master ohne `EXT-X-MAP`) — kein `cmaf`-Signal.
- `error-fetch-blocked.json` — kein `details`-Block.

Binäre Segment-Bytes liegen nicht als JSON-Fixture vor; sie werden
für Tranche 4 als deterministische Test-Builder im
`packages/stream-analyzer/tests/`-Pfad erzeugt (minimales CMAF-
Init mit `ftyp`+`moov` und minimales Media-Fragment mit
`styp`+`moof`+`traf`+`tfdt`+`mdat` plus zugehörige Negativ-Builder
für fehlende/inkompatible Boxen und ungültige Boxgrößen).

Pflicht-Brand-Allowlist (`0.10.0`):

- Init-`ftyp.major_brand` oder ein Eintrag aus
  `compatible_brands`: `cmfc` oder `cmf2`.
- Media-`styp.major_brand` oder ein Eintrag aus
  `compatible_brands`: `cmfs`, `cmff`, `cmfc` oder `cmf2`.

`cmf1` und neuere Structural-Brand-Profile bleiben Folge-Scope, bis
sie in Projekt-Doku, Fixtures und Kompatibilitätsaussage explizit
aufgenommen werden.
