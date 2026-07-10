import { describe, expect, it } from "vitest";

import { AnalysisError, analyzeHlsManifest } from "../src/index.js";
import type { AnalysisResult } from "../src/index.js";
import { detectManifestKind } from "../src/internal/parsers/detect.js";
import { classifyHlsManifest } from "../src/internal/parsers/classify.js";
import { parseAttributeList } from "../src/internal/parsers/attrs.js";

const BASE_URL = "https://cdn.example.test/stream/manifest.m3u8";

async function findingCodes(text: string): Promise<string[]> {
  const result = (await analyzeHlsManifest({ kind: "text", text, baseUrl: BASE_URL })) as AnalysisResult;
  return result.findings.map((f) => f.code);
}

// Gezielte Edge-Case-Abdeckung reiner Parser-Funktionen (Branch-Coverage-
// Härtung). Trifft Branches, die die Feature-Tests nicht anlaufen.

const BOM = String.fromCharCode(0xfeff);

describe("detectManifestKind — edge cases", () => {
  it("strips a leading UTF-8 BOM before detecting HLS", () => {
    const result = detectManifestKind(`${BOM}#EXTM3U\n#EXT-X-VERSION:3`);
    expect(result.kind).toBe("hls");
  });

  it("strips a leading BOM before detecting DASH", () => {
    const result = detectManifestKind(`${BOM}<?xml version="1.0"?>\n<MPD></MPD>`);
    expect(result.kind).toBe("dash");
  });
});

describe("classifyHlsManifest — edge cases", () => {
  it("throws manifest_not_hls when the first non-empty line is not #EXTM3U", () => {
    let caught: unknown;
    try {
      classifyHlsManifest("not a playlist\n#EXTINF:4.0,");
    } catch (error) {
      caught = error;
    }
    expect(caught).toBeInstanceOf(AnalysisError);
    expect((caught as AnalysisError).code).toBe("manifest_not_hls");
  });

  it("throws manifest_not_hls for a whitespace-only body", () => {
    let caught: unknown;
    try {
      classifyHlsManifest("   \n\n  \t\n");
    } catch (error) {
      caught = error;
    }
    expect(caught).toBeInstanceOf(AnalysisError);
    expect((caught as AnalysisError).code).toBe("manifest_not_hls");
  });
});

describe("parseAttributeList — edge cases", () => {
  it("skips an empty key from a leading comma without emitting a blank entry", () => {
    const result = parseAttributeList(",A=1");
    expect(result.get("A")).toBe("1");
    expect(result.has("")).toBe(false);
  });

  it("skips an empty key from a leading equals sign", () => {
    const result = parseAttributeList("=x,B=2");
    expect(result.get("B")).toBe("2");
    expect(result.has("")).toBe(false);
  });
});

describe("HLS malformed-URI findings (resolveUri null branch)", () => {
  it("emits variant_malformed_uri for an unresolvable STREAM-INF URI", async () => {
    const codes = await findingCodes(["#EXTM3U", "#EXT-X-STREAM-INF:BANDWIDTH=1280000", "http://"].join("\n"));
    expect(codes).toContain("variant_malformed_uri");
  });

  it("emits rendition_malformed_uri for an unresolvable EXT-X-MEDIA URI", async () => {
    const codes = await findingCodes(
      [
        "#EXTM3U",
        '#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="a",NAME="en",URI="http://"',
        '#EXT-X-STREAM-INF:BANDWIDTH=1280000,AUDIO="a"',
        "v1.m3u8"
      ].join("\n")
    );
    expect(codes).toContain("rendition_malformed_uri");
  });

  it("emits segment_malformed_uri for an unresolvable media segment URI", async () => {
    const codes = await findingCodes(
      [
        "#EXTM3U",
        "#EXT-X-VERSION:3",
        "#EXT-X-TARGETDURATION:4",
        "#EXTINF:4.0,",
        "http://",
        "#EXT-X-ENDLIST"
      ].join("\n")
    );
    expect(codes).toContain("segment_malformed_uri");
  });
});
