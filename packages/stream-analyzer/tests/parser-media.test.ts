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

  it("flags EXTINF with empty duration as malformed (Number(\"\") === 0 trap)", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "#EXTINF:,foo", "a.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "segment_malformed_extinf" && f.level === "warning")).toBe(true);
    expect(result.details.segments[0].duration).toBe(0);
  });

  it("flags EXTINF without comma as malformed", () => {
    // Tatsächlich liefert RFC 8216 EXTINF immer mit Komma, aber ein
    // bloßes "#EXTINF:6" ist parseable (Komma optional, Title leer).
    // Hier prüfen wir, dass das nicht crasht und die Dauer korrekt
    // geparst wird.
    const ok = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "#EXTINF:6", "a.ts", "#EXT-X-ENDLIST"].join("\n");
    const result = parseMediaPlaylist(ok, undefined);
    expect(result.details.segments).toHaveLength(1);
    expect(result.details.segments[0].duration).toBe(6);
    expect(result.details.segments[0].title).toBeUndefined();
    expect(result.findings.some((f) => f.code === "segment_malformed_extinf")).toBe(false);
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

  it("preserves commas inside EXTINF title (only first comma is a delimiter)", () => {
    const titled = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXTINF:6.0,Episode 1, Part 2",
      "a.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(titled, undefined);
    expect(result.details.segments[0].title).toBe("Episode 1, Part 2");
  });

  it("falls back to mediaSequence=0 when MEDIA-SEQUENCE is malformed", () => {
    const bad = ["#EXTM3U", "#EXT-X-TARGETDURATION:6", "#EXT-X-MEDIA-SEQUENCE:abc", "#EXTINF:6.0,", "a.ts"].join("\n");
    const result = parseMediaPlaylist(bad, undefined);
    expect(result.findings.some((f) => f.code === "media_malformed_mediasequence" && f.level === "warning")).toBe(true);
    expect(result.details.mediaSequence).toBe(0);
    expect(result.details.segments[0].sequenceNumber).toBe(0);
  });
});

describe("parseMediaPlaylist — unsupported feature info findings", () => {
  it("emits media_encryption_present for EXT-X-KEY with active method", () => {
    const enc = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      '#EXT-X-KEY:METHOD=AES-128,URI="key.php?id=42"',
      "#EXTINF:6.0,",
      "a.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(enc, undefined);
    const finding = result.findings.find((f) => f.code === "media_encryption_present");
    expect(finding).toBeDefined();
    expect(finding?.level).toBe("info");
  });

  it("does not flag METHOD=NONE as encryption present", () => {
    const noEnc = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXT-X-KEY:METHOD=NONE",
      "#EXTINF:6.0,",
      "a.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(noEnc, undefined);
    expect(result.findings.some((f) => f.code === "media_encryption_present")).toBe(false);
  });

  it("emits media_init_segment_present for EXT-X-MAP", () => {
    const fmp4 = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:6.0,",
      "a.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(fmp4, undefined);
    expect(result.findings.some((f) => f.code === "media_init_segment_present" && f.level === "info")).toBe(true);
  });

  it("emits media_discontinuity_present for EXT-X-DISCONTINUITY", () => {
    const disc = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXTINF:6.0,",
      "a.ts",
      "#EXT-X-DISCONTINUITY",
      "#EXTINF:6.0,",
      "b.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(disc, undefined);
    expect(result.findings.some((f) => f.code === "media_discontinuity_present" && f.level === "info")).toBe(true);
  });

  it("emits media_program_date_time_present for EXT-X-PROGRAM-DATE-TIME", () => {
    const pdt = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXT-X-PROGRAM-DATE-TIME:2026-05-01T09:00:00Z",
      "#EXTINF:6.0,",
      "a.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(pdt, undefined);
    expect(result.findings.some((f) => f.code === "media_program_date_time_present" && f.level === "info")).toBe(true);
  });

  it("emits each feature finding only once even if the tag appears multiple times", () => {
    const multi = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      "#EXT-X-DISCONTINUITY",
      "#EXTINF:6.0,",
      "a.ts",
      "#EXT-X-DISCONTINUITY",
      "#EXTINF:6.0,",
      "b.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = parseMediaPlaylist(multi, undefined);
    const count = result.findings.filter((f) => f.code === "media_discontinuity_present").length;
    expect(count).toBe(1);
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
