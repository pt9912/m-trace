import { describe, expect, it } from "vitest";

import { parseMediaPlaylist } from "../src/internal/parsers/media.js";

/**
 * Tranche-2-Tests für `0.10.0` (NF-13 / RAK-61): Media-Playlist-
 * Parser erzeugt `details.cmaf` nur, wenn HLS-Manifestsignale für
 * CMAF/fMP4 vorliegen, und reicht intern strukturierte
 * `EXT-X-MAP`/`#EXT-X-BYTERANGE`-Daten an die Tranche-4-Binary-
 * Prüfung weiter.
 */

describe("parseMediaPlaylist — CMAF detection", () => {
  it("emits cmaf with manifest confidence when EXT-X-MAP is present", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-VERSION:7",
      "#EXT-X-TARGETDURATION:4",
      "#EXT-X-MEDIA-SEQUENCE:0",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:4.0,",
      "seg-0.m4s",
      "#EXTINF:4.0,",
      "seg-1.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const { details } = parseMediaPlaylist(text, "https://cdn.example.test/");
    expect(details.cmaf).toBeDefined();
    const cmaf = details.cmaf!;
    expect(cmaf.source).toBe("hls");
    expect(cmaf.confidence).toBe("manifest");
    expect(cmaf.signals.map((s) => s.code)).toEqual([
      "hls_ext_x_map",
      "hls_segment_extension_fmp4"
    ]);
    expect(cmaf.signals[0].confidence).toBe("manifest");
    expect(cmaf.signals[0].manifestAnchor).toBe("media:line:5");
    // .m4s segment auf Zeile 7 (1-basiert).
    expect(cmaf.signals[1].manifestAnchor).toBe("media:line:7");
    // emittiert kein binary-Objekt — das kommt mit T4.
    expect(cmaf.binary).toBeUndefined();
  });

  it("downgrades fMP4-suffix-only to confidence:'inferred' without EXT-X-MAP", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXTINF:6.0,",
      "seg-0.cmfv",
      "#EXTINF:6.0,",
      "seg-1.cmfv",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const { details } = parseMediaPlaylist(text, undefined);
    expect(details.cmaf).toBeDefined();
    expect(details.cmaf!.confidence).toBe("inferred");
    expect(details.cmaf!.signals.map((s) => s.code)).toEqual([
      "hls_segment_extension_fmp4"
    ]);
    expect(details.cmaf!.signals[0].confidence).toBe("inferred");
  });

  it("emits no cmaf for classic TS playlists", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXTINF:6.0,",
      "seg-0.ts",
      "#EXTINF:6.0,",
      "seg-1.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const { details } = parseMediaPlaylist(text, undefined);
    expect(details.cmaf).toBeUndefined();
  });

  it("ignores .m4s in query strings and keeps cmaf detection deterministic", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      "#EXTINF:4.0,",
      "seg-0.ts?fmt=.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const { details } = parseMediaPlaylist(text, undefined);
    expect(details.cmaf).toBeUndefined();
  });

  it("adds the EXT-X-INDEPENDENT-SEGMENTS audit signal only when other CMAF signals are present", () => {
    const withMap = [
      "#EXTM3U",
      "#EXT-X-INDEPENDENT-SEGMENTS",
      "#EXT-X-TARGETDURATION:4",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:4.0,",
      "seg-0.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const codes = parseMediaPlaylist(withMap, undefined).details.cmaf!.signals.map((s) => s.code);
    expect(codes).toContain("hls_independent_segments");

    const tsOnly = [
      "#EXTM3U",
      "#EXT-X-INDEPENDENT-SEGMENTS",
      "#EXT-X-TARGETDURATION:4",
      "#EXTINF:4.0,",
      "seg-0.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    expect(parseMediaPlaylist(tsOnly, undefined).details.cmaf).toBeUndefined();
  });
});

describe("parseMediaPlaylist — structured EXT-X-MAP / BYTERANGE", () => {
  it("captures EXT-X-MAP attributes for the binary path", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      '#EXT-X-MAP:URI="init.mp4",BYTERANGE="1024@0"',
      "#EXTINF:4.0,",
      "seg-0.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(text, "https://cdn.example.test/");
    expect(result.cmafMeta).toBeDefined();
    const init = result.cmafMeta!.initSegment!;
    expect(init.uri).toBe("init.mp4");
    expect(init.resolvedUri).toBe("https://cdn.example.test/init.mp4");
    expect(init.byterange).toEqual({ length: 1024, offset: 0, raw: "1024@0" });
    expect(init.rawAttributes.URI).toBe("init.mp4");
    expect(init.rawAttributes.BYTERANGE).toBe("1024@0");
    expect(init.manifestAnchor).toBe("media:line:3");
    // Beide Manifestsignale sichtbar im Public-Summary.
    const codes = result.details.cmaf!.signals.map((s) => s.code);
    expect(codes).toContain("hls_ext_x_map");
    expect(codes).toContain("hls_ext_x_map_byterange");
  });

  it("binds #EXT-X-BYTERANGE to the next segment URI", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:4.0,",
      "#EXT-X-BYTERANGE:2048@4096",
      "seg-0.m4s",
      "#EXTINF:4.0,",
      "seg-1.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(text, "https://cdn.example.test/");
    const first = result.cmafMeta!.firstMediaSegment!;
    expect(first.uri).toBe("seg-0.m4s");
    expect(first.byterange).toEqual({ length: 2048, offset: 4096, raw: "2048@4096" });
    expect(first.sequenceNumber).toBe(0);
    expect(first.manifestAnchor).toBe("media:line:6");
  });

  it("flags malformed #EXT-X-BYTERANGE without dropping the segment", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      "#EXTINF:4.0,",
      "#EXT-X-BYTERANGE:not-a-number",
      "seg-0.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(text, undefined);
    expect(result.findings.some((f) => f.code === "media_byterange_malformed")).toBe(true);
    expect(result.details.segments).toHaveLength(1);
    expect(result.details.segments[0].uri).toBe("seg-0.ts");
  });

  it("warns when #EXT-X-BYTERANGE has no following URI", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      "#EXT-X-BYTERANGE:1024@0",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(text, undefined);
    expect(result.findings.some((f) => f.code === "media_byterange_orphan")).toBe(true);
  });

  it("emits a finding when EXT-X-MAP has no URI and skips the cmaf init metadata", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      '#EXT-X-MAP:BYTERANGE="1024@0"',
      "#EXTINF:4.0,",
      "seg-0.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(text, undefined);
    expect(result.findings.some((f) => f.code === "media_map_missing_uri")).toBe(true);
    // signals[] enthält trotzdem den fMP4-URI-Hinweis (inferred), weil
    // das Suffix-Signal unabhängig vom EXT-X-MAP-Parsing erkannt wird.
    expect(result.details.cmaf!.confidence).toBe("inferred");
    expect(result.cmafMeta?.initSegment).toBeUndefined();
  });
});

describe("parseMediaPlaylist — feature parity", () => {
  it("does not regress the existing media_init_segment_present info finding", () => {
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:4.0,",
      "seg-0.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(text, undefined);
    expect(result.findings.some((f) => f.code === "media_init_segment_present")).toBe(true);
  });
});
