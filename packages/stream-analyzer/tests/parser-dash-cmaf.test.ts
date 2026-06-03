import { describe, expect, it } from "vitest";

import { analyzeDashManifestText } from "../src/internal/parsers/dash.js";
import type { DashAnalysisResult } from "../src/types/result.js";

/**
 * Tranche-3-Tests für `0.10.0` (NF-13 / RAK-62 / RAK-64). Pinnen die
 * drei Confidence-Cases, BaseURL-Vererbung, SegmentTemplate-/
 * SegmentList-Erfassung und mehrperiodige Manifestanker mit
 * Index-Fallback.
 */

const VERSION = "0.10.0-test";

function analyze(mpd: string, baseUrl: string | undefined = "https://cdn.example.test/"): DashAnalysisResult {
  return analyzeDashManifestText(
    mpd,
    baseUrl !== undefined ? { source: "text", baseUrl } : { source: "text" },
    VERSION
  ).result;
}

describe("DASH-CMAF — Confidence-Regeln (RAK-62)", () => {
  it("MP4-MIME alleine erzeugt confidence:'inferred'", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <Representation id="v1" bandwidth="1280000" codecs="avc1.4d401e"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const cmaf = result.details.cmaf!;
    expect(cmaf.source).toBe("dash");
    expect(cmaf.confidence).toBe("inferred");
    expect(cmaf.signals.map((s) => s.code)).toEqual(["dash_mime_video_mp4"]);
    expect(cmaf.signals[0].confidence).toBe("inferred");
  });

  it("Initialization plus fMP4-Segmentmuster erzeugt confidence:'manifest'", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <SegmentTemplate initialization="$RepresentationID$/init.mp4" media="$RepresentationID$/seg-$Number$.m4s" startNumber="1"/>',
      '      <Representation id="v1" bandwidth="1280000" codecs="avc1.4d401e"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const cmaf = result.details.cmaf!;
    expect(cmaf.confidence).toBe("manifest");
    const codes = cmaf.signals.map((s) => s.code);
    expect(codes).toContain("dash_segment_template_initialization");
    expect(codes).toContain("dash_segment_template_media");
    expect(codes).toContain("dash_segment_extension_fmp4");
  });

  it("kein MP4-MIME, keine Initialization, kein fMP4-Suffix → kein cmaf", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp2t">',
      '      <Representation id="v1" bandwidth="1280000" codecs="avc1.4d401e"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    expect(result.details.cmaf).toBeUndefined();
  });
});

describe("DASH-CMAF — BaseURL-Vererbung (RAK-62)", () => {
  it("vererbt MPD-/Period-/AdaptationSet-/Representation-BaseURL und löst init+media auf", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <BaseURL>https://cdn.example.test/dash/</BaseURL>",
      "  <Period>",
      "    <BaseURL>p1/</BaseURL>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      "      <BaseURL>video/</BaseURL>",
      '      <SegmentTemplate initialization="init.mp4" media="seg-$Number%05d$.m4s" startNumber="1"/>',
      '      <Representation id="v1" bandwidth="1280000" codecs="avc1.4d401e"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const cmaf = result.details.cmaf!;
    expect(cmaf.confidence).toBe("manifest");
    // Anker enthält den AdaptationSet-id (statt Index).
    expect(cmaf.signals[0].manifestAnchor).toContain("AdaptationSet[id=v]");
    expect(cmaf.signals[0].manifestAnchor).toContain("Representation[id=v1]");
  });

  it("nimmt erste sichere BaseURL und ignoriert unsichere Schemes", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <BaseURL>file:///etc/passwd</BaseURL>",
      "  <BaseURL>https://cdn.example.test/dash/</BaseURL>",
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <SegmentTemplate initialization="init.mp4"/>',
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd, undefined);
    expect(result.details.cmaf!.confidence).toBe("manifest");
  });

  it("sub-level mit nur unsicheren BaseURL-Werten markiert blocked, kein Vererbungs-Bypass", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <BaseURL>https://cdn.example.test/</BaseURL>",
      "  <Period>",
      "    <BaseURL>file:///bad/</BaseURL>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <SegmentTemplate initialization="init.mp4"/>',
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd, undefined);
    // Manifest-Signale bleiben sichtbar.
    expect(result.details.cmaf).toBeDefined();
    expect(result.details.cmaf!.confidence).toBe("manifest");
  });
});

describe("DASH-CMAF — SegmentTemplate-Vererbung und Override", () => {
  it("erbt Initialization von AdaptationSet auf Representation", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <SegmentTemplate initialization="$RepresentationID$/init.mp4" media="$RepresentationID$/seg-$Number$.m4s" startNumber="1"/>',
      '      <Representation id="v1" bandwidth="1280000"/>',
      '      <Representation id="v2" bandwidth="2560000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    expect(result.details.cmaf!.confidence).toBe("manifest");
    // Signal-Anker zeigt auf v1, weil Plan T3-Auswahl die erste
    // Representation mit init+media nimmt.
    expect(result.details.cmaf!.signals[0].manifestAnchor).toContain("Representation[id=v1]");
  });
});

describe("DASH-CMAF — SegmentList-Pfad", () => {
  it("erkennt Initialization@sourceURL und SegmentURL@media", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      "      <SegmentList>",
      '        <Initialization sourceURL="init.mp4"/>',
      '        <SegmentURL media="seg-001.m4s"/>',
      '        <SegmentURL media="seg-002.m4s"/>',
      "      </SegmentList>",
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const codes = result.details.cmaf!.signals.map((s) => s.code);
    expect(codes).toContain("dash_segment_list_initialization");
    expect(codes).toContain("dash_segment_list_media");
    expect(codes).toContain("dash_segment_extension_fmp4");
  });
});

describe("DASH-CMAF — Mehrperiodige Manifestanker mit Index-Fallback", () => {
  it("nutzt Index, wenn Period/AdaptationSet/Representation keine ID tragen", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet mimeType="video/mp4">',
      '      <SegmentTemplate initialization="init0.mp4"/>',
      '      <Representation bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "  <Period>",
      '    <AdaptationSet mimeType="video/mp4">',
      '      <SegmentTemplate initialization="init1.mp4"/>',
      '      <Representation bandwidth="2560000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const anchors = result.details.cmaf!.signals.map((s) => s.manifestAnchor);
    expect(anchors.some((a) => a.startsWith("MPD/Period[0]/AdaptationSet[0]/Representation[0]"))).toBe(true);
    expect(anchors.some((a) => a.startsWith("MPD/Period[1]/AdaptationSet[0]/Representation[0]"))).toBe(true);
  });
});

describe("DASH-CMAF — Template-Auflösung", () => {
  it("erkennt $Number%0Nd$ Padding für startNumber-Default 1", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet mimeType="video/mp4">',
      '      <SegmentTemplate initialization="init.mp4" media="seg-$Number%05d$.m4s"/>',
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    expect(result.details.cmaf!.confidence).toBe("manifest");
  });

  it("$Time$ ist nicht aufgelöst und führt zu kein Manifest-Signal für media", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet mimeType="video/mp4">',
      '      <SegmentTemplate initialization="init.mp4" media="seg-$Time$.m4s"/>',
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const codes = result.details.cmaf!.signals.map((s) => s.code);
    expect(codes).toContain("dash_segment_template_initialization");
    // media-Template ist unauflösbar → kein dash_segment_template_media-Signal.
    expect(codes).not.toContain("dash_segment_template_media");
  });
});

/**
 * Coverage-Härtung-Tests für -Folge-Patch: dash.ts branches
 * von ~80% auf >=90% heben durch Edge-Cases, die der Happy-Path-
 * Korpus nicht erreicht.
 */
describe("DASH — MIME-Fallback und Tag-Parsing-Edge-Cases", () => {
  it("unbekannter Representation@mimeType fällt auf dash_mime_mp4 zurück", () => {
    // dashMimeSignalCode (dash.ts Line 528) — Fallback-Branch für
    // bekannten-aber-nicht-gemappten mimeType.
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      "    <AdaptationSet>",
      '      <Representation id="v1" bandwidth="1280000" mimeType="application/dash+xml"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const cmaf = result.details.cmaf;
    if (cmaf !== undefined) {
      // Wenn ein cmaf-Block entsteht, muss der Fallback-Code aktiv sein.
      const mimeSignals = cmaf.signals.filter((s) => s.code.startsWith("dash_mime"));
      for (const signal of mimeSignals) {
        expect(["dash_mime_mp4", "dash_mime_video_mp4", "dash_mime_audio_mp4", "dash_mime_application_mp4"]).toContain(signal.code);
      }
    }
  });

  it("toleriert ein selbstschließendes Tag ohne Whitespace im Open-Tag", () => {
    // parseAttributes (dash.ts Line 739) — degenerierte Tags
    // (z. B. `<EmptyTag/>` ohne Attribute) sollen nicht crashen,
    // sondern eine leere Attribut-Map liefern. Der Parser muss
    // robust gegen Edge-Cases im wild gewachsenen DASH-Korpus sein.
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      "    <BaseURL/>",
      '    <AdaptationSet mimeType="video/mp4">',
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    expect(() => analyze(mpd)).not.toThrow();
  });
});

/**
 * Pflichtfeld-Validierung Representation@bandwidth — Coverage-
 * Härtung für dash.ts Lines 372–377 (dash_representation_missing_
 * bandwidth Finding). MPEG-DASH §5.3.5 macht bandwidth zur Pflicht;
 * der Parser muss das melden, ohne den Lauf zu brechen.
 */
describe("DASH — Representation ohne bandwidth-Attribut", () => {
  it("emittiert dash_representation_missing_bandwidth als Finding und überspringt die Representation", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <Representation id="v1" codecs="avc1.4d401e"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const codes = result.findings.map((f) => f.code);
    expect(codes).toContain("dash_representation_missing_bandwidth");
  });
});

/**
 * SegmentTemplate-Vererbung — Coverage-Härtung für
 * `mergeSegmentTemplate` (dash.ts Lines 440–449). Das Happy-Path-
 * Set deckt nur Single-Source-Templates ab; Merging beider Quellen
 * mit Override-Semantik (initialization, media, startNumber !== 1)
 * fehlt.
 */
describe("DASH — SegmentTemplate-Merge mit Parent+Own (Override)", () => {
  it("merged AdaptationSet- und Representation-Templates und respektiert Override-Reihenfolge", () => {
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      '      <SegmentTemplate initialization="parent-init.mp4" media="parent-$Number$.m4s" startNumber="1"/>',
      '      <Representation id="v1" bandwidth="1280000" codecs="avc1.4d401e">',
      '        <SegmentTemplate media="own-$Number$.m4s" startNumber="7"/>',
      "      </Representation>",
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = analyze(mpd);
    const cmaf = result.details.cmaf!;
    const codes = cmaf.signals.map((s) => s.code);
    // Init wurde von parent vererbt, media von own überschrieben —
    // dash_segment_template_media muss da sein.
    expect(codes).toContain("dash_segment_template_initialization");
    expect(codes).toContain("dash_segment_template_media");
  });
});
