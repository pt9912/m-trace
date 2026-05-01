import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

import { analyzeHlsManifest, AnalysisError, STREAM_ANALYZER_VERSION } from "../src/index.js";
import type { AnalysisErrorResult, AnalysisResult } from "../src/index.js";

const fixturesDir = join(dirname(fileURLToPath(import.meta.url)), "fixtures");
const fixture = (name: string): string => readFileSync(join(fixturesDir, name), "utf8");

describe("analyzeHlsManifest — Tranche 2 contract", () => {
  it("classifies and parses a master playlist", async () => {
    const result = (await analyzeHlsManifest({ kind: "text", text: fixture("master.m3u8") })) as AnalysisResult;

    expect(result.status).toBe("ok");
    expect(result.analyzerVersion).toBe(STREAM_ANALYZER_VERSION);
    expect(result.input).toEqual({ source: "text" });
    expect(result.playlistType).toBe("master");
    expect(result.findings.some((f) => f.code === "playlist_ambiguous")).toBe(false);
    expect(result.details).not.toBeNull();
    const details = result.details as { variants: unknown[]; renditions: unknown[] };
    expect(details.variants).toHaveLength(2);
    expect(details.renditions).toHaveLength(1);
    expect(result.summary).toEqual({ itemCount: 3 });
  });

  it("snapshots the master fixture result for stability", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: fixture("master.m3u8"),
      baseUrl: "https://cdn.example.test/"
    })) as AnalysisResult;
    expect(result.playlistType).toBe("master");
    expect(result.details).toMatchSnapshot();
  });

  it("snapshots the media fixture result for stability", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: fixture("media.m3u8"),
      baseUrl: "https://cdn.example.test/stream/manifest.m3u8"
    })) as AnalysisResult;
    expect(result.playlistType).toBe("media");
    expect(result.details).toMatchSnapshot();
  });

  it("classifies a media playlist", async () => {
    const result = (await analyzeHlsManifest({ kind: "text", text: fixture("media.m3u8") })) as AnalysisResult;

    expect(result.status).toBe("ok");
    expect(result.playlistType).toBe("media");
    expect(result.findings.some((f) => f.code === "playlist_ambiguous")).toBe(false);
  });

  it("flags ambiguous playlists with master as primary type", async () => {
    const result = (await analyzeHlsManifest({ kind: "text", text: fixture("ambiguous.m3u8") })) as AnalysisResult;

    expect(result.status).toBe("ok");
    expect(result.playlistType).toBe("master");
    expect(result.findings.find((f) => f.code === "playlist_ambiguous")).toMatchObject({ level: "warning" });
  });

  it("rejects non-HLS content with manifest_not_hls", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: fixture("not-hls.txt")
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("manifest_not_hls");
    expect(result.details).toMatchObject({ firstLine: expect.stringContaining("<html>") });
  });

  it("rejects empty manifests with manifest_not_hls", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: fixture("empty.m3u8")
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("manifest_not_hls");
  });

  it("rejects whitespace-only manifests with manifest_not_hls", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: "\n\n   \n\t\n"
    })) as AnalysisErrorResult;

    expect(result.status).toBe("error");
    expect(result.code).toBe("manifest_not_hls");
    expect(result.message).toContain("leer");
  });

  it("classifies malformed HLS files but still emits findings", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: fixture("malformed.m3u8")
    })) as AnalysisResult;

    expect(result.status).toBe("ok");
    expect(result.playlistType).toBe("media");
    // Tranche 4: malformed.m3u8 hat ungültige EXTINF-Dauer und keinen
    // EXT-X-ENDLIST → erwartet mindestens segment_malformed_extinf
    // und media_missing_targetduration.
    expect(result.findings.some((f) => f.code === "segment_malformed_extinf" && f.level === "warning")).toBe(true);
    expect(result.findings.some((f) => f.code === "media_missing_targetduration" && f.level === "error")).toBe(true);
  });

  it("preserves baseUrl in input metadata when supplied", async () => {
    const result = (await analyzeHlsManifest({
      kind: "text",
      text: "#EXTM3U\n",
      baseUrl: "https://cdn.example.test/"
    })) as AnalysisResult;

    expect(result.status).toBe("ok");
    expect(result.input).toEqual({ source: "text", baseUrl: "https://cdn.example.test/" });
    expect(result.playlistType).toBe("unknown");
    expect(result.findings.some((f) => f.code === "playlist_type_unknown")).toBe(true);
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

  it("returns a stable top-level shape (Tranche 5 envelope incl. analyzerKind)", async () => {
    const result = await analyzeHlsManifest({ kind: "text", text: "#EXTM3U\n" });

    const okKeys = Object.keys(result).sort();
    expect(okKeys).toEqual(
      [
        "analyzerKind",
        "analyzerVersion",
        "details",
        "findings",
        "input",
        "playlistType",
        "status",
        "summary"
      ].sort()
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
