import { describe, expect, it } from "vitest";

import {
  parseIPv4,
  parseIPv6,
  validateResolvedIp,
  validateUrl
} from "../src/internal/loader/ssrf.js";

describe("validateUrl", () => {
  it.each([
    ["http://example.test/manifest.m3u8", true],
    ["https://example.test/manifest.m3u8", true],
    ["ftp://example.test/manifest.m3u8", false],
    ["file:///etc/passwd", false],
    ["gopher://example.test/", false],
    ["javascript:alert(1)", false]
  ])("accepts only http/https schemes (%s → ok=%s)", (raw, ok) => {
    const decision = validateUrl(new URL(raw));
    expect(decision.ok).toBe(ok);
    if (!ok) {
      expect(decision.reason).toBe("scheme_not_allowed");
    }
  });

  it("rejects URLs with credentials", () => {
    const decision = validateUrl(new URL("https://user:pass@example.test/"));
    expect(decision.ok).toBe(false);
    expect(decision.reason).toBe("credentials_in_url");
  });

  it("rejects URLs with username only", () => {
    const decision = validateUrl(new URL("https://user@example.test/"));
    expect(decision.ok).toBe(false);
    expect(decision.reason).toBe("credentials_in_url");
  });
});

describe("parseIPv4", () => {
  it.each([
    ["1.2.3.4", 0x01020304],
    ["0.0.0.0", 0],
    ["255.255.255.255", 0xffffffff],
    ["127.0.0.1", 0x7f000001]
  ])("parses %s", (input, expected) => {
    expect(parseIPv4(input)).toBe(expected);
  });

  it.each(["1.2.3", "1.2.3.4.5", "1.2.3.256", "1.2.3.-1", "1.2.3.04", "abc.def.ghi.jkl", "1..2.3"])(
    "rejects invalid %s",
    (input) => {
      expect(parseIPv4(input)).toBeNull();
    }
  );
});

describe("parseIPv6", () => {
  it.each([
    "::1",
    "::",
    "fe80::1",
    "2001:db8::1",
    "::ffff:192.0.2.1",
    "1:2:3:4:5:6:7:8",
    "1::8"
  ])("parses %s", (input) => {
    expect(parseIPv6(input)).not.toBeNull();
  });

  it.each(["::1::2", "1:2:3:4:5:6:7:8:9", "fe80::1%eth0"])("handles %s", (input) => {
    // Zone IDs must be stripped to a parseable address.
    if (input.includes("::1::")) {
      expect(parseIPv6(input)).toBeNull();
    } else if (input.includes("%")) {
      expect(parseIPv6(input)).not.toBeNull();
    } else {
      expect(parseIPv6(input)).toBeNull();
    }
  });

  it.each(["xyz", "1:2:3", ":::", "1:2:3:4:5:6:7"])("rejects %s", (input) => {
    expect(parseIPv6(input)).toBeNull();
  });

  it("rejects mapped form with malformed embedded IPv4", () => {
    expect(parseIPv6("::ffff:1.2.3")).toBeNull();
    expect(parseIPv6("::ffff:1.2.3.4.5")).toBeNull();
    expect(parseIPv6("1.2.3.4")).toBeNull();
  });
});

describe("validateResolvedIp — IPv4 blocklist", () => {
  it.each([
    ["0.0.0.0", "unspecified"],
    ["10.0.0.1", "private (RFC1918 10/8)"],
    ["100.64.1.1", "CGN"],
    ["127.0.0.1", "loopback"],
    ["169.254.169.254", "AWS link-local metadata endpoint"],
    ["172.16.0.5", "private (RFC1918 172.16/12)"],
    ["172.31.255.254", "private upper bound"],
    ["192.0.2.1", "TEST-NET-1"],
    ["192.168.1.1", "private (RFC1918 192.168/16)"],
    ["198.18.0.1", "benchmarking"],
    ["198.51.100.1", "TEST-NET-2"],
    ["203.0.113.1", "TEST-NET-3"],
    ["224.0.0.1", "multicast"],
    ["240.0.0.1", "reserved"],
    ["255.255.255.255", "broadcast"]
  ])("blocks %s (%s)", (address) => {
    const decision = validateResolvedIp(address, 4);
    expect(decision.ok).toBe(false);
    expect(decision.reason).toBe("ip_blocked");
  });

  it.each([
    "1.1.1.1",
    "8.8.8.8",
    "172.15.255.255", // just below 172.16/12
    "172.32.0.0", // just above 172.16/12
    "100.63.255.255", // just below 100.64/10
    "192.169.0.0",
    "223.255.255.255" // just below multicast
  ])("allows %s", (address) => {
    expect(validateResolvedIp(address, 4).ok).toBe(true);
  });

  it("rejects unparseable IPv4 input", () => {
    const decision = validateResolvedIp("999.999.999.999", 4);
    expect(decision.ok).toBe(false);
    expect(decision.reason).toBe("ip_unparseable");
  });
});

describe("validateResolvedIp — IPv6 blocklist", () => {
  it.each([
    ["::", "unspecified"],
    ["::1", "loopback"],
    ["::ffff:127.0.0.1", "IPv4-mapped"],
    ["fc00::1", "unique local"],
    ["fd12:3456::1", "unique local"],
    ["fe80::1", "link-local"],
    ["ff02::1", "multicast"],
    ["2001:db8::1", "documentation"],
    ["100::1", "discard-only"]
  ])("blocks %s (%s)", (address) => {
    const decision = validateResolvedIp(address, 6);
    expect(decision.ok).toBe(false);
    expect(decision.reason).toBe("ip_blocked");
  });

  it.each(["2001:4860:4860::8888", "2606:4700:4700::1111"])("allows public %s", (address) => {
    expect(validateResolvedIp(address, 6).ok).toBe(true);
  });

  it("rejects unparseable IPv6 input", () => {
    const decision = validateResolvedIp("not-an-ip", 6);
    expect(decision.ok).toBe(false);
    expect(decision.reason).toBe("ip_unparseable");
  });
});
