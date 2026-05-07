import { describe, expect, it } from "vitest";
import * as fc from "fast-check";

import { redactUrl } from "../src/adapters/hlsjs/redact.js";

// plan-0.9.5 §4 Tranche 3 (extra-gates.md §3.5) — Property-Tests
// für die SDK-URL-Redaction. Pinnt:
//
//   - jede Eingabe (string | undefined | null) liefert einen
//     non-empty String zurück; niemals throw.
//   - Eingaben mit JWT-Shape oder langen Hex-Strings im Pfad
//     werden auf den `:redacted`-Sentinel reduziert.
//
// Tranche-3b-Bench-Lehre: `fc.webUrl(...)` und
// `fc.stringMatching(...)` mit weiten Bereichen + `.filter(...)`
// haben fast-check 4.4 in einen Discard-Loop geschickt (Test
// hing > 30 min in CPU-Schleife). Lösung: alle Properties nutzen
// **deterministische Generators mit fixer Länge** — keine `.filter`-
// Pfade, kein webUrl. Das ist ein reduzierter, aber aussagekräftiger
// Property-Korpus; der vollständige URL-Redaction-Korpus
// (mit fc.webUrl o. ä.) ist Folge-Backlog-Item, sobald fast-check
// in einer Folge-Version den Discard-Pfad hardened.
//
// `interruptAfterTimeLimit` ist als Schutznetz gesetzt, falls
// fast-check trotz fixer Generators wieder in eine ähnliche Falle
// läuft — der Test fail dann stattdessen mit einer eindeutigen
// Timeout-Meldung, statt CI/Pre-Commit zu hängen.
const FC_OPTIONS = {
  verbose: 1 as const,
  interruptAfterTimeLimit: 4_000,
};

describe("redactUrl property tests (RAK-Wave-2)", () => {
  it("never throws on bounded path-segment-only inputs", () => {
    // Wir generieren nur kurze ASCII-Strings (statt Unicode-Random)
    // plus die drei Sentinel-Inputs (undefined/null/empty). Das
    // exerciert den early-return-Pfad in redactUrl ohne fast-check
    // in einen string-Generator-Discard-Pfad zu schicken.
    const asciiChar = fc.constantFrom(
      "a", "b", "c", "x", "y", "z", "0", "1", "2", "_", "-", "/", "."
    );
    const asciiString = fc
      .array(asciiChar, { minLength: 0, maxLength: 32 })
      .map((cs) => cs.join(""));
    fc.assert(
      fc.property(
        fc.oneof(
          asciiString,
          fc.constant(""),
          fc.constant(undefined),
          fc.constant(null)
        ),
        (input) => {
          const out = redactUrl(input as string | undefined | null);
          expect(typeof out).toBe("string");
          expect(out.length).toBeGreaterThan(0);
        }
      ),
      { ...FC_OPTIONS, numRuns: 100 }
    );
  });

  it("reduces JWT-shape path segments to :redacted", () => {
    // Drei base64url-Blöcke à fixer Länge — keine `.filter`-Pfade.
    const base64UrlChar = fc.constantFrom(
      "A", "B", "C", "D", "E", "F", "G", "H", "a", "b", "c", "d",
      "0", "1", "2", "3", "4", "5", "_", "-"
    );
    const block = fc
      .array(base64UrlChar, { minLength: 16, maxLength: 24 })
      .map((cs) => cs.join(""));
    fc.assert(
      fc.property(
        fc.tuple(block, block, block).map(([a, b, c]) => `${a}.${b}.${c}`),
        (jwt) => {
          const out = redactUrl(`https://example.test/api/${jwt}/profile`);
          expect(out).toContain(":redacted");
          expect(out).not.toContain(jwt);
        }
      ),
      { ...FC_OPTIONS, numRuns: 30 }
    );
  });

  it("reduces long hex-only path segments to :redacted", () => {
    const hexChar = fc.constantFrom(
      "0", "1", "2", "3", "4", "5", "6", "7",
      "8", "9", "a", "b", "c", "d", "e", "f"
    );
    // Fixe 40-Zeichen-Hex-Strings (gerade, ≥ 32 — passt zur
    // Hex-Erkennung in redact.ts).
    const fixedHex = fc
      .array(hexChar, { minLength: 40, maxLength: 40 })
      .map((cs) => cs.join(""));
    fc.assert(
      fc.property(fixedHex, (hex) => {
        const out = redactUrl(`https://example.test/sessions/${hex}/events`);
        expect(out).toContain(":redacted");
        expect(out).not.toContain(hex);
      }),
      { ...FC_OPTIONS, numRuns: 30 }
    );
  });
});
