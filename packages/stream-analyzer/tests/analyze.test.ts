import { describe, expect, it } from "vitest";

import { analyzeHlsManifest, AnalysisError, STREAM_ANALYZER_VERSION } from "../src/index.js";
import type { AnalysisErrorResult, AnalysisResult } from "../src/index.js";

describe("analyzeHlsManifest — Tranche 1 contract", () => {
  it("returns an ok result for a text manifest input", async () => {
    const result = await analyzeHlsManifest({ kind: "text", text: "#EXTM3U\n" });

    expect(result.status).toBe("ok");
    const ok = result as AnalysisResult;
    expect(ok.analyzerVersion).toBe(STREAM_ANALYZER_VERSION);
    expect(ok.input).toEqual({ source: "text" });
    expect(ok.playlistType).toBe("unknown");
    expect(ok.summary).toEqual({ itemCount: 0 });
    expect(ok.findings).toHaveLength(1);
    expect(ok.findings[0]).toMatchObject({ code: "not_implemented", level: "info" });
    expect(ok.details).toBeNull();
  });

  it("preserves baseUrl in input metadata when supplied", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: "#EXTM3U\n",
      baseUrl: "https://cdn.example.test/"
    })) as AnalysisResult;

    expect(result.status).toBe("ok");
    expect(result.input).toEqual({ source: "text", baseUrl: "https://cdn.example.test/" });
  });

  it("returns a structured error result for url input in Tranche 1", async () => {
    const result = (await analyzeHlsManifest({
      kind: "url",
      url: "https://example.test/manifest.m3u8"
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("fetch_blocked");
    expect(result.analyzerVersion).toBe(STREAM_ANALYZER_VERSION);
    expect(result.message).toMatch(/Tranche 2/);
    expect(result.details).toEqual({ url: "https://example.test/manifest.m3u8" });
  });

  it("rejects invalid url input shape", async () => {
    const result = (await analyzeHlsManifest({
      kind: "url",
      url: ""
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("invalid_input");
    expect(result).not.toHaveProperty("details");
  });

  it("rejects invalid text input shape", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      // @ts-expect-error — runtime guard for unsafe callers
      text: 42
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("invalid_input");
    expect(result).not.toHaveProperty("details");
  });

  it("rejects unknown manifest input kind", async () => {
    const result = (await analyzeHlsManifest({
      // @ts-expect-error — runtime guard for unsafe callers
      kind: "binary",
      data: new Uint8Array()
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("invalid_input");
    expect(result).not.toHaveProperty("details");
  });

  it("returns a stable top-level shape", async () => {
    const result = await analyzeHlsManifest({ kind: "text", text: "#EXTM3U\n" });

    const okKeys = Object.keys(result).sort();
    expect(okKeys).toEqual(
      ["analyzerVersion", "details", "findings", "input", "playlistType", "status", "summary"].sort()
    );
  });
});

describe("AnalysisError class", () => {
  it("carries code and details", () => {
    const err = new AnalysisError("invalid_input", "boom", { field: "kind" });
    expect(err.code).toBe("invalid_input");
    expect(err.message).toBe("boom");
    expect(err.details).toEqual({ field: "kind" });
    expect(err.name).toBe("AnalysisError");
    expect(err).toBeInstanceOf(Error);
  });

  it("works without details", () => {
    const err = new AnalysisError("internal_error", "x");
    expect(err.details).toBeUndefined();
  });
});
