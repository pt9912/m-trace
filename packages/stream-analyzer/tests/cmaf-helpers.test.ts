import { describe, expect, it } from "vitest";

import { parseByteRangePayload } from "../src/internal/parsers/cmaf-hls.js";
import {
  resolveBaseUrlChain,
  resolveDashTemplate,
  resolveSegmentUri
} from "../src/internal/parsers/cmaf-dash.js";

/**
 * Branch-coverage-Targeted-Tests für die CMAF-Helper-Module.
 * Decken Edge-Cases ab, die der Public-Pfad (Manifest-Tests) nicht
 * isoliert auslöst — Plan `0.10.0` Tranche 3 / RAK-62 / RAK-64.
 */

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
