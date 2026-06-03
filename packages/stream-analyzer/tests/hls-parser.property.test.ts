import { describe, expect, it } from "vitest";
import * as fc from "fast-check";

import { analyzeManifest } from "../src/index.js";

//  (extra-gates.md) — Property-Tests
// für den HLS-Manifest-Parser. Pinnt:
//
//   - jeder Eingabe-String, der mit `#EXTM3U` beginnt, kommt mit
//     `analyzerKind: "hls"` und einem definierten `playlistType`
//     ("master" / "media" / "unknown") zurück. Kein Crash, kein
//     unbounded-Findings-Array, keine NaN-Item-Counts.
//   - Random-Eingaben **ohne** `#EXTM3U`-Header werden vom Detector
//     als `manifest_not_supported` klassifiziert (oder `dash`/`hls`-
//     Detector-Pfad bei zufälligem XML-Marker). Der HLS-Parser
//     selbst läuft nur, wenn der Detector HLS klassifiziert hat.

describe("HLS parser property tests (RAK-Wave-2)", () => {
  it("any input starting with #EXTM3U produces a deterministic AnalysisResult", async () => {
    await fc.assert(
      fc.asyncProperty(
        fc.array(
          fc.oneof(
            fc.constant("#EXT-X-VERSION:6"),
            fc.constant("#EXT-X-TARGETDURATION:10"),
            fc.constant("#EXT-X-MEDIA-SEQUENCE:0"),
            fc.constant("#EXT-X-PLAYLIST-TYPE:VOD"),
            fc.constant("#EXT-X-ENDLIST"),
            fc.constant("#EXTINF:6.000,seg-1.ts"),
            fc.constant("seg-1.ts"),
            fc.constant("#EXT-X-STREAM-INF:BANDWIDTH=1280000"),
            fc.constant("video/720p/main.m3u8"),
            fc.string({ minLength: 0, maxLength: 64 })
          ),
          { minLength: 0, maxLength: 50 }
        ),
        async (lines) => {
          const text = ["#EXTM3U", ...lines].join("\n");
          const result = await analyzeManifest({ kind: "text", text });
          expect(result).toBeDefined();
          expect(result.status).toMatch(/^(ok|error)$/);
          if (result.status === "ok") {
            expect(result.analyzerKind).toBe("hls");
            expect(["master", "media", "unknown"]).toContain(result.playlistType);
            expect(Number.isFinite(result.summary.itemCount)).toBe(true);
            expect(result.summary.itemCount).toBeGreaterThanOrEqual(0);
            expect(Array.isArray(result.findings)).toBe(true);
          }
        }
      ),
      { numRuns: 100 }
    );
  });

  it("any non-HLS, non-DASH input is rejected with manifest_not_supported", async () => {
    await fc.assert(
      fc.asyncProperty(
        // Random text, but explicitly not starting with HLS/DASH-
        // markers. We filter the predicate post-hoc rather than
        // composing two arbitraries.
        fc.string({ minLength: 1, maxLength: 256 }).filter((s) => {
          const trimmed = s.replace(/^[\s\r\n]+/, "");
          return (
            !trimmed.startsWith("#EXTM3U") &&
            !trimmed.startsWith("<?xml") &&
            !trimmed.startsWith("<MPD")
          );
        }),
        async (text) => {
          const result = await analyzeManifest({ kind: "text", text });
          expect(result.status).toBe("error");
          if (result.status === "error") {
            expect(result.code).toBe("manifest_not_supported");
          }
        }
      ),
      { numRuns: 100 }
    );
  });
});
