import { describe, expect, it } from "vitest";
import { isTokenLikePathSegment, redactUrl } from "../src/adapters/hlsjs/redact";

describe("redactUrl", () => {
  it("returns the redacted sentinel for null, undefined, or empty input", () => {
    expect(redactUrl(undefined)).toBe(":redacted");
    expect(redactUrl(null)).toBe(":redacted");
    expect(redactUrl("")).toBe(":redacted");
  });

  it("returns the redacted sentinel for unparsable input", () => {
    expect(redactUrl("not-a-url")).toBe(":redacted");
    expect(redactUrl("://no-scheme")).toBe(":redacted");
  });

  it("strips query and fragment", () => {
    expect(redactUrl("https://cdn.example.test/playlists/main.m3u8?token=abc&sig=xy#frag")).toBe(
      "https://cdn.example.test/playlists/main.m3u8"
    );
  });

  it("strips userinfo", () => {
    expect(redactUrl("https://alice:p%40ss@cdn.example.test/seg/0001.ts")).toBe(
      "https://cdn.example.test/seg/0001.ts"
    );
  });

  it("redacts token-like long base64 path segments", () => {
    const longSegment = "a".repeat(32);
    expect(redactUrl(`https://cdn.example.test/${longSegment}/playlist.m3u8`)).toBe(
      "https://cdn.example.test/:redacted/playlist.m3u8"
    );
  });

  it("redacts even-length hex segments ≥ 32", () => {
    const hex = "ab".repeat(16);
    expect(redactUrl(`https://cdn.example.test/${hex}/seg.ts`)).toBe(
      "https://cdn.example.test/:redacted/seg.ts"
    );
  });

  it("redacts JWT-shaped segments (three base64url blocks)", () => {
    const jwt = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.dozjgNryP4J3jVmNHl0w5N";
    expect(redactUrl(`https://cdn.example.test/auth/${jwt}/playlist.m3u8`)).toBe(
      "https://cdn.example.test/auth/:redacted/playlist.m3u8"
    );
  });

  it("decodes percent-encoded JWT segments before applying the heuristic", () => {
    // %2E is the percent-encoded `.` — without decode, the segment
    // would not match the JWT pattern.
    const encoded = "eyJhbGciOiJIUzI1NiJ9%2EeyJzdWIiOiJ1c2VyIn0%2EsigBlockHere";
    expect(redactUrl(`https://cdn.example.test/auth/${encoded}/playlist.m3u8`)).toBe(
      "https://cdn.example.test/auth/:redacted/playlist.m3u8"
    );
  });

  it("leaves short, normal path segments intact", () => {
    expect(redactUrl("https://cdn.example.test/playlists/v1/main.m3u8")).toBe(
      "https://cdn.example.test/playlists/v1/main.m3u8"
    );
  });

  it("preserves the leading slash and empty intermediate segments", () => {
    expect(redactUrl("https://cdn.example.test//double/slash.ts")).toBe(
      "https://cdn.example.test//double/slash.ts"
    );
  });

  it("returns scheme+host when only the host is present", () => {
    // URL constructor normalizes "https://cdn.example.test" to
    // "https://cdn.example.test/" (trailing slash). The redactor
    // preserves that normalization.
    expect(redactUrl("https://cdn.example.test")).toBe("https://cdn.example.test/");
  });

  it("preserves a trailing slash", () => {
    expect(redactUrl("https://cdn.example.test/playlists/")).toBe(
      "https://cdn.example.test/playlists/"
    );
  });
});

describe("isTokenLikePathSegment", () => {
  it("rejects empty and short segments", () => {
    expect(isTokenLikePathSegment("")).toBe(false);
    expect(isTokenLikePathSegment("short")).toBe(false);
    expect(isTokenLikePathSegment("a".repeat(23))).toBe(false);
  });

  it("accepts long base64-shaped segments", () => {
    expect(isTokenLikePathSegment("a".repeat(24))).toBe(true);
    expect(isTokenLikePathSegment("A1_b-".repeat(6))).toBe(true);
  });

  it("rejects segments with too few [A-Za-z0-9_-] chars", () => {
    // 24 chars, half are special — below the 80% threshold.
    expect(isTokenLikePathSegment("a".repeat(12) + "%".repeat(12))).toBe(false);
  });

  it("accepts even-length hex strings ≥ 32", () => {
    expect(isTokenLikePathSegment("0".repeat(32))).toBe(true);
    // Odd-length hex string still trips the long-token heuristic
    // (≥24 chars + 100% from [A-Za-z0-9_-]). Pin both paths so the
    // dual-heuristic stays correct.
    expect(isTokenLikePathSegment("0".repeat(31))).toBe(true);
  });

  it("accepts JWT-shaped strings", () => {
    expect(isTokenLikePathSegment("aaa.bbb.ccc")).toBe(true);
    expect(isTokenLikePathSegment("aaa.bbb")).toBe(false); // only two blocks
  });
});
