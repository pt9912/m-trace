import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

import { analyzeHlsManifest, AnalysisError, STREAM_ANALYZER_VERSION } from "../src/index.js";
import type {
  AnalysisResult,
  AnalysisErrorResult,
  MasterAnalysisResult,
  MediaAnalysisResult
} from "../src/index.js";

const fixturesDir = join(dirname(fileURLToPath(import.meta.url)), "fixtures");
const fixture = (name: string): string => readFileSync(join(fixturesDir, name), "utf8");

describe("AnalysisResult — discriminated union narrowing", () => {
  it("narrows details to MasterPlaylistDetails when playlistType === 'master'", async () => {
    const result = (await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8") })) as AnalysisResult;
    expect(result.status).toBe("ok");
    if (result.status !== "ok") return;
    expect(result.playlistType).toBe("master");
    if (result.playlistType !== "master") return;
    // Without any cast, TypeScript knows details is MasterPlaylistDetails:
    const variants = result.details.variants;
    expect(variants.length).toBeGreaterThan(0);
    expect(variants[0].bandwidth).toBeGreaterThan(0);
  });

  it("narrows details to MediaPlaylistDetails when playlistType === 'media'", async () => {
    const result = (await analyzeHlsManifest({ kind: "text", text: fixture("media.m3u8") })) as AnalysisResult;
    expect(result.status).toBe("ok");
    if (result.status !== "ok") return;
    expect(result.playlistType).toBe("media");
    if (result.playlistType !== "media") return;
    // No cast needed: details is MediaPlaylistDetails.
    expect(result.details.segments).toBeInstanceOf(Array);
    expect(result.details.endList).toBeTypeOf("boolean");
  });

  it("narrows details to null when playlistType === 'unknown'", async () => {
    const result = (await analyzeHlsManifest({ kind: "text", text: "#EXTM3U\n" })) as AnalysisResult;
    expect(result.status).toBe("ok");
    if (result.status !== "ok") return;
    if (result.playlistType !== "unknown") return;
    expect(result.details).toBeNull();
  });
});

describe("AnalysisResult — envelope shape (Tranche 5)", () => {
  it("carries analyzerKind=hls on every success result", async () => {
    const master = (await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8") })) as MasterAnalysisResult;
    expect(master.analyzerKind).toBe("hls");

    const media = (await analyzeHlsManifest({ kind: "text", text: fixture("media.m3u8") })) as MediaAnalysisResult;
    expect(media.analyzerKind).toBe("hls");

    const unknown = (await analyzeHlsManifest({ kind: "text", text: "#EXTM3U\n" })) as AnalysisResult;
    expect(unknown.analyzerKind).toBe("hls");
  });

  it("carries analyzerKind=hls on every error result", async () => {
    const result = (await analyzeHlsManifest({
      kind: "url",
      url: "ftp://example.test/manifest"
    })) as AnalysisErrorResult;
    expect(result.status).toBe("error");
    expect(result.analyzerKind).toBe("hls");
  });

  it("uses status as the discriminator between ok and error", async () => {
    const ok = await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8") });
    expect(ok.status).toBe("ok");
    const err = await analyzeHlsManifest({ kind: "text", text: "" });
    expect(err.status).toBe("error");
  });

  it("locks analyzerVersion to package.json", async () => {
    const result = await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8") });
    expect(result.analyzerVersion).toBe(STREAM_ANALYZER_VERSION);
  });
});

describe("AnalysisResult — JSON serialization stability", () => {
  it("is deterministic across repeated calls (no Map iteration leakage, no Date.now drift)", async () => {
    const text = fixture("master.m3u8");
    const baseUrl = "https://cdn.example.test/";
    const a = await analyzeHlsManifest({ kind: "text", text, baseUrl });
    const b = await analyzeHlsManifest({ kind: "text", text, baseUrl });
    expect(JSON.stringify(a)).toBe(JSON.stringify(b));
  });

  it("survives a JSON round-trip without losing data", async () => {
    const result = await analyzeHlsManifest({
      kind: "text",
      text: fixture("media.m3u8"),
      baseUrl: "https://cdn.example.test/stream/manifest.m3u8"
    });
    const reparsed = JSON.parse(JSON.stringify(result));
    expect(reparsed).toEqual(result);
  });

  it("never carries `undefined` properties (would break JSON consumers)", async () => {
    const cases = [
      await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8"), baseUrl: "https://cdn.example.test/" }),
      await analyzeHlsManifest({ kind: "text", text: fixture("media.m3u8") }),
      await analyzeHlsManifest({ kind: "text", text: fixture("media-live.m3u8") }),
      await analyzeHlsManifest({ kind: "text", text: "#EXTM3U\n" }),
      await analyzeHlsManifest({ kind: "url", url: "ftp://blocked.test/m" })
    ];
    for (const result of cases) {
      assertNoUndefined(result, "$");
    }
  });

  it("only contains finite numbers (no NaN/Infinity)", async () => {
    const cases = [
      await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8") }),
      await analyzeHlsManifest({ kind: "text", text: fixture("media.m3u8") }),
      await analyzeHlsManifest({ kind: "text", text: fixture("media-live.m3u8") }),
      await analyzeHlsManifest({ kind: "text", text: fixture("media-target-violation.m3u8") }),
      await analyzeHlsManifest({ kind: "text", text: fixture("media-outlier.m3u8") })
    ];
    for (const result of cases) {
      assertOnlyFiniteNumbers(result, "$");
    }
  });
});

describe("AnalysisErrorResult — independence from success shape", () => {
  it("does not carry success-only fields", async () => {
    const err = (await analyzeHlsManifest({ kind: "text", text: "" })) as AnalysisErrorResult;
    expect(err.status).toBe("error");
    expect(err).not.toHaveProperty("playlistType");
    expect(err).not.toHaveProperty("summary");
    expect(err).not.toHaveProperty("findings");
    expect(err).not.toHaveProperty("input");
  });

  it("AnalysisError.details is undefined when not provided to the constructor", () => {
    const e = new AnalysisError("internal_error", "x");
    expect(e.details).toBeUndefined();
  });
});

function assertNoUndefined(value: unknown, path: string): void {
  if (value === null || value === undefined) {
    if (value === undefined) {
      throw new Error(`Found undefined at ${path}`);
    }
    return;
  }
  if (Array.isArray(value)) {
    for (let i = 0; i < value.length; i++) {
      assertNoUndefined(value[i], `${path}[${i}]`);
    }
    return;
  }
  if (typeof value === "object") {
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      if (v === undefined) {
        throw new Error(`Found undefined at ${path}.${k}`);
      }
      assertNoUndefined(v, `${path}.${k}`);
    }
  }
}

function assertOnlyFiniteNumbers(value: unknown, path: string): void {
  if (typeof value === "number" && !Number.isFinite(value)) {
    throw new Error(`Found non-finite number ${value} at ${path}`);
  }
  if (Array.isArray(value)) {
    for (let i = 0; i < value.length; i++) {
      assertOnlyFiniteNumbers(value[i], `${path}[${i}]`);
    }
    return;
  }
  if (value !== null && typeof value === "object") {
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      assertOnlyFiniteNumbers(v, `${path}.${k}`);
    }
  }
}
