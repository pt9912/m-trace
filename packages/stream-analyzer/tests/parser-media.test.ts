import { describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

import { parseMediaPlaylist } from "../src/internal/parsers/media.js";

const fixturesDir = join(dirname(fileURLToPath(import.meta.url)), "fixtures");
const fixture = (name: string): string => readFileSync(join(fixturesDir, name), "utf8");

describe("parseMediaPlaylist — VOD", () => {
  it("extracts segments, target duration, sequence and ENDLIST", () => {
    const result = parseMediaPlaylist(fixture("media.m3u8"), undefined);
    expect(result.details.targetDuration).toBe(6);
    expect(result.details.mediaSequence).toBe(0);
    expect(result.details.playlistType).toBe("VOD");
    expect(result.details.endList).toBe(true);
    expect(result.details.live).toBe(false);
    expect(result.details.liveLatencyEstimateSeconds).toBeUndefined();
    expect(result.details.segments).toHaveLength(4);
    const first = result.details.segments[0];
    expect(first.uri).toBe("seg-0.ts");
    expect(first.duration).toBeCloseTo(5.967);
    expect(first.sequenceNumber).toBe(0);
    const last = result.details.segments[3];
    expect(last.duration).toBeCloseTo(3.5);
    expect(last.sequenceNumber).toBe(3);
  });

  it("computes summary statistics", () => {
    const result = parseMediaPlaylist(fixture("media.m3u8"), undefined);
    const s = result.details.summary;
    expect(s.count).toBe(4);
    expect(s.totalDuration).toBeCloseTo(5.967 * 3 + 3.5, 3);
    expect(s.minDuration).toBeCloseTo(3.5);
    expect(s.maxDuration).toBeCloseTo(5.967);
    expect(s.averageDuration).toBeCloseTo((5.967 * 3 + 3.5) / 4, 3);
  });

  it("does not flag a short last VOD segment as outlier", () => {
    const result = parseMediaPlaylist(fixture("media.m3u8"), undefined);
    expect(result.findings.some((f) => f.code === "segment_duration_outlier")).toBe(false);
  });
});

describe("parseMediaPlaylist — Live", () => {
  it("treats absence of ENDLIST as live and computes a 3× target latency estimate", () => {
    const result = parseMediaPlaylist(fixture("media-live.m3u8"), undefined);
    expect(result.details.endList).toBe(false);
    expect(result.details.live).toBe(true);
    expect(result.details.targetDuration).toBe(4);
    expect(result.details.liveLatencyEstimateSeconds).toBe(12);
    expect(result.details.mediaSequence).toBe(8423);
    expect(result.details.segments[0].sequenceNumber).toBe(8423);
    expect(result.details.segments[3].sequenceNumber).toBe(8426);
  });

  it("does not exempt the last live segment from outlier checks", () => {
    const live = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXTINF:6.000,",
      "a.ts",
      "#EXTINF:6.000,",
      "b.ts",
      "#EXTINF:0.100,",
      "c.ts"
    ].join("\n");
    const result = parseMediaPlaylist(live, undefined);
    expect(result.details.live).toBe(true);
    expect(result.findings.some((f) => f.code === "segment_duration_outlier")).toBe(true);
  });
});

describe("parseMediaPlaylist — findings and tolerances", () => {
  it("flags target-duration violations as errors", () => {
    const result = parseMediaPlaylist(fixture("media-target-violation.m3u8"), undefined);
    const violation = result.findings.find((f) => f.code === "segment_duration_exceeds_target");
    expect(violation).toBeDefined();
    expect(violation?.level).toBe("error");
  });

  it("flags very-short non-last segments as outliers", () => {
    const result = parseMediaPlaylist(fixture("media-outlier.m3u8"), undefined);
    const outlier = result.findings.find((f) => f.code === "segment_duration_outlier");
    expect(outlier).toBeDefined();
    expect(outlier?.level).toBe("warning");
  });

  it("emits media_missing_targetduration when the tag is absent", () => {
    const noTd = ["#EXTM3U", "#EXTINF:5.0,", "a.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(noTd, undefined);
    expect(result.findings.some((f) => f.code === "media_missing_targetduration" && f.level === "error")).toBe(true);
    expect(result.details.targetDuration).toBeUndefined();
    expect(result.details.liveLatencyEstimateSeconds).toBeUndefined();
  });

  it("flags malformed EXT-X-TARGETDURATION but continues parsing", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:six", "#EXTINF:5.0,", "a.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "media_malformed_targetduration" && f.level === "error")).toBe(true);
    expect(result.details.segments).toHaveLength(1);
  });

  it("flags malformed EXTINF duration but still records the segment", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "#EXTINF:abc,", "a.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "segment_malformed_extinf" && f.level === "warning")).toBe(true);
    expect(result.details.segments).toHaveLength(1);
    expect(result.details.segments[0].duration).toBe(0);
  });

  it("flags EXTINF without a following URI as missing-uri error", () => {
    const bad = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXTINF:6.0,",
      "#EXTINF:6.0,",
      "b.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "segment_missing_uri" && f.level === "error")).toBe(true);
    expect(result.details.segments).toHaveLength(1);
    expect(result.details.segments[0].uri).toBe("b.ts");
  });

  it("flags trailing EXTINF with no URI as missing-uri error", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "#EXTINF:6.0,", "a.ts", "#EXTINF:6.0,"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "segment_missing_uri" && f.level === "error")).toBe(true);
  });

  it("flags stray URI lines without a preceding EXTINF", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "stray.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "segment_unexpected_uri" && f.level === "warning")).toBe(true);
    expect(result.details.segments).toHaveLength(0);
  });

  it("captures the optional EXTINF title", () => {
    const titled = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", '#EXTINF:6.0,Episode 1', "a.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(titled, undefined);
    expect(result.details.segments[0].title).toBe("Episode 1");
  });

  it("falls back to mediaSequence=0 when MEDIA-SEQUENCE is malformed", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "#EXT-X-MEDIA-SEQUENCE:abc", "#EXTINF:6.0,", "a.ts"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "media_malformed_mediasequence" && f.level === "warning")).toBe(true);
    expect(result.details.mediaSequence).toBe(0);
    expect(result.details.segments[0].sequenceNumber).toBe(0);
  });
});

describe("parseMediaPlaylist — base URL resolution", () => {
  it("resolves segment URIs against the base URL when supplied", () => {
    const result = parseMediaPlaylist(
      fixture("media.m3u8"),
      "https://cdn.example.test/stream/manifest.m3u8"
    );
    expect(result.details.segments[0].uri).toBe("seg-0.ts");
    expect(result.details.segments[0].resolvedUri).toBe("https://cdn.example.test/stream/seg-0.ts");
  });
});
