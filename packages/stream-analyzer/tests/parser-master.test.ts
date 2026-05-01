import { describe, expect, it } from "vitest";

import { parseMasterPlaylist } from "../src/internal/parsers/master.js";

describe("parseMasterPlaylist — variants", () => {
  it("extracts BANDWIDTH, RESOLUTION, CODECS, FRAME-RATE and URI", () => {
    const result = parseMasterPlaylist(
      [
        "#EXTM3U",
        '#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720,CODECS="avc1.4d401e,mp4a.40.2",FRAME-RATE=29.97',
        "video/720p/main.m3u8",
        ""
      ].join("\n"),
      undefined
    );
    expect(result.details.variants).toHaveLength(1);
    const v = result.details.variants[0];
    expect(v.bandwidth).toBe(1280000);
    expect(v.resolution).toEqual({ width: 1280, height: 720 });
    expect(v.codecs).toEqual(["avc1.4d401e", "mp4a.40.2"]);
    expect(v.frameRate).toBeCloseTo(29.97);
    expect(v.uri).toBe("video/720p/main.m3u8");
    expect(v.resolvedUri).toBeUndefined();
  });

  it("resolves relative URIs against the base URL", () => {
    const result = parseMasterPlaylist(
      [
        "#EXTM3U",
        "#EXT-X-STREAM-INF:BANDWIDTH=1000",
        "v/main.m3u8"
      ].join("\n"),
      "https://cdn.example.test/manifest.m3u8"
    );
    expect(result.details.variants[0].uri).toBe("v/main.m3u8");
    expect(result.details.variants[0].resolvedUri).toBe("https://cdn.example.test/v/main.m3u8");
  });

  it("flags missing BANDWIDTH but still records the variant", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", "#EXT-X-STREAM-INF:RESOLUTION=640x360", "v/main.m3u8"].join("\n"),
      undefined
    );
    expect(result.details.variants).toHaveLength(1);
    expect(result.details.variants[0].bandwidth).toBe(0);
    expect(result.findings.some((f) => f.code === "variant_missing_bandwidth" && f.level === "error")).toBe(true);
  });

  it("flags STREAM-INF without a following URI line", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", "#EXT-X-STREAM-INF:BANDWIDTH=1000", "#EXT-X-STREAM-INF:BANDWIDTH=2000", "v2/main.m3u8"].join("\n"),
      undefined
    );
    expect(result.details.variants).toHaveLength(1);
    expect(result.findings.some((f) => f.code === "variant_missing_uri" && f.level === "error")).toBe(true);
  });

  it("flags STREAM-INF as the very last (unterminated) tag", () => {
    const result = parseMasterPlaylist(["#EXTM3U", "#EXT-X-STREAM-INF:BANDWIDTH=1000"].join("\n"), undefined);
    expect(result.details.variants).toHaveLength(0);
    expect(result.findings.some((f) => f.code === "variant_missing_uri" && f.level === "error")).toBe(true);
  });

  it("warns about malformed RESOLUTION but keeps the variant", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", "#EXT-X-STREAM-INF:BANDWIDTH=1000,RESOLUTION=oops", "v/main.m3u8"].join("\n"),
      undefined
    );
    expect(result.details.variants[0].resolution).toBeUndefined();
    expect(result.findings.some((f) => f.code === "variant_malformed_resolution" && f.level === "warning")).toBe(true);
  });
});

describe("parseMasterPlaylist — renditions", () => {
  it("extracts AUDIO renditions with all optional fields", () => {
    const result = parseMasterPlaylist(
      [
        "#EXTM3U",
        '#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aud-en",NAME="English",LANGUAGE="en",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO,CHANNELS="2",URI="audio/en.m3u8"'
      ].join("\n"),
      undefined
    );
    expect(result.details.renditions).toHaveLength(1);
    const r = result.details.renditions[0];
    expect(r.type).toBe("AUDIO");
    expect(r.groupId).toBe("aud-en");
    expect(r.name).toBe("English");
    expect(r.language).toBe("en");
    expect(r.uri).toBe("audio/en.m3u8");
    expect(r.default).toBe(true);
    expect(r.autoselect).toBe(true);
    expect(r.forced).toBe(false);
    expect(r.channels).toBe("2");
  });

  it("emits an error finding when TYPE/GROUP-ID/NAME are missing", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", '#EXT-X-MEDIA:GROUP-ID="g",NAME="n"'].join("\n"),
      undefined
    );
    expect(result.details.renditions).toHaveLength(0);
    expect(result.findings.some((f) => f.code === "rendition_missing_required_attr" && f.level === "error")).toBe(true);
  });

  it("warns about unknown TYPE values", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", '#EXT-X-MEDIA:TYPE=THERMAL,GROUP-ID="g",NAME="n"'].join("\n"),
      undefined
    );
    expect(result.details.renditions).toHaveLength(1);
    expect(result.findings.some((f) => f.code === "rendition_unknown_type")).toBe(true);
  });

  it("warns about AUDIO/VIDEO/SUBTITLES renditions without URI", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", '#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="g",NAME="n"'].join("\n"),
      undefined
    );
    expect(result.details.renditions).toHaveLength(1);
    expect(result.findings.some((f) => f.code === "rendition_missing_uri")).toBe(true);
  });

  it("does not require URI for CLOSED-CAPTIONS", () => {
    const result = parseMasterPlaylist(
      ["#EXTM3U", '#EXT-X-MEDIA:TYPE=CLOSED-CAPTIONS,GROUP-ID="cc",NAME="cc-en",LANGUAGE="en",INSTREAM-ID="CC1"'].join("\n"),
      undefined
    );
    expect(result.findings.some((f) => f.code === "rendition_missing_uri")).toBe(false);
  });
});

describe("parseMasterPlaylist — cross-references", () => {
  it("flags variants that reference an undeclared AUDIO group", () => {
    const result = parseMasterPlaylist(
      [
        "#EXTM3U",
        "#EXT-X-STREAM-INF:BANDWIDTH=1000,AUDIO=\"missing-group\"",
        "v/main.m3u8"
      ].join("\n"),
      undefined
    );
    expect(result.findings.some((f) => f.code === "variant_group_undefined")).toBe(true);
  });

  it("does not flag valid cross-references", () => {
    const result = parseMasterPlaylist(
      [
        "#EXTM3U",
        '#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aud",NAME="n",URI="a.m3u8"',
        '#EXT-X-STREAM-INF:BANDWIDTH=1000,AUDIO="aud"',
        "v/main.m3u8"
      ].join("\n"),
      undefined
    );
    expect(result.findings.some((f) => f.code === "variant_group_undefined")).toBe(false);
  });
});

describe("parseMasterPlaylist — I-frame variants", () => {
  it("emits an info finding and does not record the variant", () => {
    const result = parseMasterPlaylist(
      [
        "#EXTM3U",
        '#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=86000,RESOLUTION=1280x720,URI="iframes.m3u8"'
      ].join("\n"),
      undefined
    );
    expect(result.details.variants).toHaveLength(0);
    expect(result.findings.some((f) => f.code === "i_frame_variant_skipped" && f.level === "info")).toBe(true);
  });
});
