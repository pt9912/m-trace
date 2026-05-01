import { describe, expect, it } from "vitest";

import {
  parseAttributeList,
  parseCodecs,
  parseFloatAttr,
  parseIntAttr,
  parseResolution,
  parseYesNo
} from "../src/internal/parsers/attrs.js";

describe("parseAttributeList", () => {
  it("parses simple key=value pairs", () => {
    const result = parseAttributeList("BANDWIDTH=1280000,RESOLUTION=1280x720");
    expect(result.get("BANDWIDTH")).toBe("1280000");
    expect(result.get("RESOLUTION")).toBe("1280x720");
  });

  it("preserves commas inside quoted values", () => {
    const result = parseAttributeList('CODECS="avc1.4d401e,mp4a.40.2",RESOLUTION=1280x720');
    expect(result.get("CODECS")).toBe("avc1.4d401e,mp4a.40.2");
    expect(result.get("RESOLUTION")).toBe("1280x720");
  });

  it("handles empty values and trailing commas", () => {
    const result = parseAttributeList("KEY=,KEY2=value,");
    expect(result.get("KEY")).toBe("");
    expect(result.get("KEY2")).toBe("value");
  });

  it("handles keys without value as empty string", () => {
    const result = parseAttributeList("FORCED,DEFAULT=YES");
    expect(result.get("FORCED")).toBe("");
    expect(result.get("DEFAULT")).toBe("YES");
  });

  it("trims leading whitespace before keys", () => {
    const result = parseAttributeList("  KEY1=a,   KEY2=b");
    expect(result.get("KEY1")).toBe("a");
    expect(result.get("KEY2")).toBe("b");
  });

  it("trims whitespace around unquoted values (real-world tolerance)", () => {
    const result = parseAttributeList("KEY = value ,KEY2= other");
    expect(result.get("KEY")).toBe("value");
    expect(result.get("KEY2")).toBe("other");
  });

  it("preserves whitespace inside quoted values byte-for-byte", () => {
    const result = parseAttributeList('NAME="  spaced  "');
    expect(result.get("NAME")).toBe("  spaced  ");
  });

  it("returns an empty map for empty input", () => {
    expect(parseAttributeList("").size).toBe(0);
  });

  it("uses the last value for duplicate keys", () => {
    const result = parseAttributeList("K=a,K=b");
    expect(result.get("K")).toBe("b");
  });

  it("reads values with hex prefix as-is", () => {
    const result = parseAttributeList("IV=0x1234ABCD");
    expect(result.get("IV")).toBe("0x1234ABCD");
  });
});

describe("parseYesNo", () => {
  it("maps YES/NO to booleans", () => {
    expect(parseYesNo("YES")).toBe(true);
    expect(parseYesNo("NO")).toBe(false);
  });
  it("returns undefined for unknown or missing", () => {
    expect(parseYesNo(undefined)).toBeUndefined();
    expect(parseYesNo("maybe")).toBeUndefined();
    expect(parseYesNo("")).toBeUndefined();
  });
});

describe("parseResolution", () => {
  it("parses WIDTHxHEIGHT", () => {
    expect(parseResolution("1280x720")).toEqual({ width: 1280, height: 720 });
  });
  it("returns null for malformed input", () => {
    expect(parseResolution("1280X720")).toBeNull();
    expect(parseResolution("x720")).toBeNull();
    expect(parseResolution("12.0x720")).toBeNull();
    expect(parseResolution(undefined)).toBeNull();
  });
});

describe("parseCodecs", () => {
  it("splits comma-separated codec lists", () => {
    expect(parseCodecs("avc1.4d401e,mp4a.40.2")).toEqual(["avc1.4d401e", "mp4a.40.2"]);
  });
  it("trims spaces around codecs", () => {
    expect(parseCodecs(" avc1 , mp4a ")).toEqual(["avc1", "mp4a"]);
  });
  it("returns null for missing input", () => {
    expect(parseCodecs(undefined)).toBeNull();
  });
  it("returns empty array for empty input", () => {
    expect(parseCodecs("")).toEqual([]);
  });
});

describe("parseIntAttr", () => {
  it("parses decimal integers", () => {
    expect(parseIntAttr("1280000")).toBe(1280000);
  });
  it("rejects non-decimal", () => {
    expect(parseIntAttr("12.5")).toBeNull();
    expect(parseIntAttr("0x10")).toBeNull();
    expect(parseIntAttr("abc")).toBeNull();
    expect(parseIntAttr(undefined)).toBeNull();
  });
});

describe("parseFloatAttr", () => {
  it("parses floats", () => {
    expect(parseFloatAttr("29.97")).toBe(29.97);
    expect(parseFloatAttr("30")).toBe(30);
  });
  it("returns null for NaN or missing", () => {
    expect(parseFloatAttr("abc")).toBeNull();
    expect(parseFloatAttr(undefined)).toBeNull();
  });
  it("rejects empty and whitespace-only inputs (would be Number(\"\") === 0)", () => {
    expect(parseFloatAttr("")).toBeNull();
    expect(parseFloatAttr("   ")).toBeNull();
    expect(parseFloatAttr("\t")).toBeNull();
  });
  it("rejects Infinity and negative values", () => {
    expect(parseFloatAttr("Infinity")).toBeNull();
    expect(parseFloatAttr("-Infinity")).toBeNull();
    expect(parseFloatAttr("-1.5")).toBeNull();
    expect(parseFloatAttr("-0.001")).toBeNull();
  });
});
