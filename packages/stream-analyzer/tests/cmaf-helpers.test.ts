import { describe, expect, it } from "vitest";

import {
  aggregateConfidence,
  isFmp4SegmentUri,
  masterAnchor,
  mediaAnchor,
  parseByteRangePayload
} from "../src/internal/parsers/cmaf-hls.js";
import {
  buildDashAnchor,
  isFmp4DashUri,
  isMp4MimeType,
  resolveBaseUrlChain,
  resolveDashTemplate,
  resolveSegmentUri
} from "../src/internal/parsers/cmaf-dash.js";

/**
 * Branch-coverage-Targeted-Tests für die CMAF-Helper-Module.
 * Decken Edge-Cases ab, die der Public-Pfad (Manifest-Tests) nicht
 * isoliert auslöst — Plan RAK-62 / RAK-64.
 */

describe("cmaf-hls — aggregateConfidence", () => {
  it("returns inferred when no signal carries higher confidence", () => {
    expect(
      aggregateConfidence([
        { code: "x", level: "info", manifestAnchor: "a", confidence: "inferred" }
      ])
    ).toBe("inferred");
  });

  it("upgrades to manifest when at least one signal carries it", () => {
    expect(
      aggregateConfidence([
        { code: "x", level: "info", manifestAnchor: "a", confidence: "inferred" },
        { code: "y", level: "info", manifestAnchor: "b", confidence: "manifest" }
      ])
    ).toBe("manifest");
  });

  it("short-circuits on binary even with later manifest entries", () => {
    expect(
      aggregateConfidence([
        { code: "x", level: "info", manifestAnchor: "a", confidence: "binary" },
        { code: "y", level: "info", manifestAnchor: "b", confidence: "manifest" }
      ])
    ).toBe("binary");
  });

  it("returns inferred for an empty array", () => {
    expect(aggregateConfidence([])).toBe("inferred");
  });
});

describe("cmaf-hls / cmaf-dash — anchor + URI helpers", () => {
  it("mediaAnchor / masterAnchor format 1-based line numbers", () => {
    expect(mediaAnchor(0)).toBe("media:line:1");
    expect(masterAnchor(4)).toBe("master:line:5");
  });

  it("isFmp4SegmentUri ignores empty input", () => {
    expect(isFmp4SegmentUri("")).toBe(false);
  });

  it("isFmp4DashUri ignores empty input", () => {
    expect(isFmp4DashUri("")).toBe(false);
  });

  it("isFmp4DashUri matches .mp4 suffix as well", () => {
    expect(isFmp4DashUri("seg-001.mp4")).toBe(true);
    expect(isFmp4DashUri("seg-001.ts")).toBe(false);
  });

  it("isMp4MimeType accepts the three CMAF-relevant MP4 MIME values", () => {
    expect(isMp4MimeType("video/mp4")).toBe(true);
    expect(isMp4MimeType("audio/mp4")).toBe(true);
    expect(isMp4MimeType("application/mp4")).toBe(true);
    expect(isMp4MimeType("VIDEO/MP4")).toBe(true);
    expect(isMp4MimeType(undefined)).toBe(false);
    expect(isMp4MimeType("video/mp2t")).toBe(false);
  });

  it("buildDashAnchor uses index when id is missing on each level", () => {
    expect(
      buildDashAnchor({
        periodIdx: 0,
        adaptationSetIdx: 1,
        representationIdx: 2
      })
    ).toBe("MPD/Period[0]/AdaptationSet[1]/Representation[2]");
  });

  it("buildDashAnchor stops at the deepest provided level", () => {
    expect(buildDashAnchor({ periodIdx: 0 })).toBe("MPD/Period[0]");
  });

  it("buildDashAnchor appends an attribute when supplied", () => {
    expect(
      buildDashAnchor({ periodIdx: 0 }, "@profiles")
    ).toBe("MPD/Period[0]/@profiles");
  });
});

describe("cmaf-hls — parseByteRangePayload", () => {
  it("returns length-only when no @offset is present", () => {
    expect(parseByteRangePayload("1024")).toEqual({ length: 1024, raw: "1024" });
  });

  it("rejects malformed offset", () => {
    expect(parseByteRangePayload("1024@xyz")).toBeNull();
  });

  it("returns null for empty payload", () => {
    expect(parseByteRangePayload("")).toBeNull();
  });

  it("returns null for non-numeric length", () => {
    expect(parseByteRangePayload("abc@0")).toBeNull();
  });
});

describe("cmaf-dash — resolveBaseUrlChain", () => {
  it("returns parent when no candidates", () => {
    expect(resolveBaseUrlChain([], "https://cdn.example.test/")).toEqual({
      baseUrl: "https://cdn.example.test/",
      blocked: false
    });
  });

  it("skips empty trimmed candidates and returns blocked when only empties remain", () => {
    expect(resolveBaseUrlChain([" ", "\t"], undefined)).toEqual({
      baseUrl: undefined,
      blocked: true
    });
  });

  it("accepts the first safe absolute URL", () => {
    const result = resolveBaseUrlChain(
      ["https://safe.example.test/dash/"],
      undefined
    );
    expect(result.blocked).toBe(false);
    expect(result.baseUrl).toBe("https://safe.example.test/dash/");
  });

  it("falls through to the second candidate when the first is malformed", () => {
    const result = resolveBaseUrlChain(
      ["http://%%bogus%%/", "https://ok.example.test/"],
      undefined
    );
    expect(result.baseUrl).toBe("https://ok.example.test/");
  });

  it("rejects non-http(s) absolute schemes", () => {
    expect(resolveBaseUrlChain(["ftp://bad.example.test/"], undefined)).toEqual({
      baseUrl: undefined,
      blocked: true
    });
  });

  it("rejects relative candidates without a parent base", () => {
    expect(resolveBaseUrlChain(["dash/"], undefined)).toEqual({
      baseUrl: undefined,
      blocked: true
    });
  });

  it("resolves a relative candidate against the parent base", () => {
    const result = resolveBaseUrlChain(["dash/"], "https://cdn.example.test/");
    expect(result.blocked).toBe(false);
    expect(result.baseUrl).toBe("https://cdn.example.test/dash/");
  });
});

describe("cmaf-dash — resolveSegmentUri", () => {
  it("returns null for an empty URI", () => {
    expect(resolveSegmentUri("", "https://cdn.example.test/")).toBeNull();
  });

  it("rejects non-http(s) absolute schemes", () => {
    expect(resolveSegmentUri("ftp://x/seg.mp4", "https://cdn.example.test/")).toBeNull();
  });

  it("rejects malformed absolute URLs", () => {
    expect(resolveSegmentUri("http://%%/seg.mp4", undefined)).toBeNull();
  });

  it("accepts a safe absolute https URI without base", () => {
    expect(resolveSegmentUri("https://cdn.example.test/seg.m4s", undefined)).toBe(
      "https://cdn.example.test/seg.m4s"
    );
  });

  it("returns null for a relative URI without baseUrl", () => {
    expect(resolveSegmentUri("seg.m4s", undefined)).toBeNull();
  });

  it("rejects relative URI when baseUrl produces an unsafe scheme", () => {
    expect(resolveSegmentUri("seg.m4s", "file:///root/")).toBeNull();
  });
});

describe("cmaf-dash — resolveDashTemplate", () => {
  const ctx = { representationId: "v1", bandwidth: 1280000, number: 7 };

  it("supports $$ literal escape for $", () => {
    expect(resolveDashTemplate("a$$b", ctx)).toBe("a$b");
  });

  it("supports zero-padded $Number%0Nd$", () => {
    expect(resolveDashTemplate("seg-$Number%05d$.m4s", ctx)).toBe("seg-00007.m4s");
  });

  it("supports zero-padded $Bandwidth%0Nd$", () => {
    expect(resolveDashTemplate("$Bandwidth%010d$.mp4", ctx)).toBe("0001280000.mp4");
  });

  it("returns null on unbalanced $", () => {
    expect(resolveDashTemplate("seg-$Number", ctx)).toBeNull();
  });

  it("returns null for unknown variables", () => {
    expect(resolveDashTemplate("seg-$Time$.m4s", ctx)).toBeNull();
  });

  it("returns null for printf width that resolves to zero", () => {
    // $Number%00d$ matches the regex (0\d+ → "00") but width=0 ≤ 0
    // → returns null (consistent with the "unknown variable" branch).
    expect(resolveDashTemplate("seg-$Number%00d$.m4s", ctx)).toBeNull();
  });
});
