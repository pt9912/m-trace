import { describe, expect, it } from "vitest";

import { STREAM_ANALYZER_NAME, STREAM_ANALYZER_VERSION } from "../src/version.js";
import packageJson from "../package.json" with { type: "json" };

describe("stream-analyzer version metadata", () => {
  it("name matches package.json", () => {
    expect(STREAM_ANALYZER_NAME).toBe(packageJson.name);
  });

  it("version matches package.json", () => {
    expect(STREAM_ANALYZER_VERSION).toBe(packageJson.version);
  });

  it("uses the @npm9912 scope", () => {
    expect(STREAM_ANALYZER_NAME.startsWith("@npm9912/")).toBe(true);
  });

  it("targets release 0.6.0", () => {
    expect(STREAM_ANALYZER_VERSION).toBe("0.6.0");
  });
});
