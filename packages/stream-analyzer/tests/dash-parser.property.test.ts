import { describe, expect, it } from "vitest";
import * as fc from "fast-check";

import { analyzeManifest } from "../src/index.js";
import { detectManifestKind } from "../src/internal/parsers/detect.js";

// plan-0.9.5 §4 Tranche 3 (extra-gates.md §3.5) — Property-Tests
// für den DASH-Manifest-Parser (`0.9.0` Tranche 3, RAK-58 / NF-12).
// Pinnt:
//
//   - Detector klassifiziert jeden Body, der mit `<?xml` oder `<MPD`
//     beginnt, als DASH (egal was danach steht).
//   - DASH-Manifeste mit komplettem MPD/Period/AdaptationSet/
//     Representation-Skelett liefern `analyzerKind: "dash"`,
//     `playlistType: "dash"`, deterministische `details.type`
//     (`static`/`dynamic`) und `summary.itemCount` ≥ 0.

describe("DASH parser property tests (RAK-Wave-2)", () => {
  it("detector classifies any <?xml-prefixed body as dash", () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 0, maxLength: 256 }),
        (suffix) => {
          const text = `<?xml version="1.0"?>${suffix}`;
          const result = detectManifestKind(text);
          expect(result.kind).toBe("dash");
        }
      ),
      { numRuns: 100 }
    );
  });

  it("detector classifies any <MPD-prefixed body as dash", () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 0, maxLength: 256 }),
        (suffix) => {
          const text = `<MPD${suffix}`;
          const result = detectManifestKind(text);
          expect(result.kind).toBe("dash");
        }
      ),
      { numRuns: 100 }
    );
  });

  it("any well-formed MPD with Period/AdaptationSet/Representation produces a deterministic dash result", async () => {
    // Lehre aus Tranche 3b (siehe docs/dev/fuzzing.md §3): kein
    // `.filter(...)` auf `fc.string`, weil fast-check 4.4 bei vielen
    // Discards in eine CPU-Schleife läuft. Stattdessen deterministische
    // Generators mit fixer Alphabet-Quelle.
    const idChar = fc.constantFrom(
      "a", "b", "c", "d", "e", "f", "0", "1", "2", "3", "_", "-"
    );
    const arbId = fc
      .array(idChar, { minLength: 1, maxLength: 16 })
      .map((cs) => cs.join(""));
    const arbAdaptationSet = fc.record({
      contentType: fc.constantFrom("video", "audio", "text"),
      mimeType: fc.constantFrom("video/mp4", "audio/mp4", "application/mp4"),
      reps: fc.array(
        fc.record({
          id: arbId,
          bandwidth: fc.integer({ min: 64_000, max: 10_000_000 }),
        }),
        { minLength: 0, maxLength: 5 }
      ),
    });

    await fc.assert(
      fc.asyncProperty(
        fc.constantFrom<"static" | "dynamic">("static", "dynamic"),
        fc.array(arbAdaptationSet, { minLength: 1, maxLength: 4 }),
        async (mpdType, sets) => {
          const adaptationSets = sets
            .map((set, sIdx) => {
              const reps = set.reps
                .map(
                  (rep) =>
                    `<Representation id="${rep.id}" bandwidth="${rep.bandwidth}" codecs="avc1.4d401e"/>`
                )
                .join("");
              return `<AdaptationSet id="${sIdx}" contentType="${set.contentType}" mimeType="${set.mimeType}">${reps}</AdaptationSet>`;
            })
            .join("");
          const text =
            `<?xml version="1.0"?>` +
            `<MPD type="${mpdType}"><Period>${adaptationSets}</Period></MPD>`;

          const result = await analyzeManifest({ kind: "text", text });
          expect(result.status).toBe("ok");
          if (result.status === "ok" && result.playlistType === "dash") {
            expect(result.analyzerKind).toBe("dash");
            expect(result.details.type).toBe(mpdType);
            expect(result.details.live).toBe(mpdType === "dynamic");
            expect(result.summary.itemCount).toBeGreaterThanOrEqual(0);
            const totalReps = sets.reduce((acc, s) => acc + s.reps.length, 0);
            expect(result.summary.itemCount).toBe(totalReps);
          } else if (result.status === "ok") {
            throw new Error(
              `expected playlistType "dash", got ${result.playlistType}`
            );
          }
        }
      ),
      { numRuns: 50 }
    );
  });
});
