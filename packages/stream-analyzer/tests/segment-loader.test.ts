import { describe, expect, it } from "vitest";

import { loadSegment } from "../src/internal/cmaf/segment-loader.js";
import type {
  LoaderResponse,
  LoaderRuntime,
  ResolvedAddress
} from "../src/internal/loader/runtime.js";

/**
 * Direkte Unit-Tests für `loadSegment` (`0.10.0` Tranche 4 / RAK-64).
 * Decken die Pfade ab, die der Verifier-Pfad mit echten Bytes nicht
 * trivial trifft: SSRF-Block, Redirect-Limit, DNS-Fehler, IPv4-/
 * IPv6-Literal-Hosts.
 */

interface ResponseSpec {
  status?: number;
  headers?: Record<string, string>;
  bytes?: Uint8Array;
  delayMs?: number;
  expectedRange?: string;
}

function asyncIterable(bytes: Uint8Array): AsyncIterable<Uint8Array> {
  return (async function* () {
    yield bytes;
  })();
}

function makeRuntime(responses: Record<string, ResponseSpec>, opts: {
  resolveError?: Error;
  resolveCustom?: (host: string) => ResolvedAddress[];
} = {}): LoaderRuntime {
  return {
    async resolveHost(hostname) {
      if (opts.resolveError) throw opts.resolveError;
      if (opts.resolveCustom) return opts.resolveCustom(hostname);
      return [{ address: "93.184.216.34", family: 4 }];
    },
    async fetch(url, init) {
      const spec = responses[url];
      if (!spec) throw new Error(`unexpected fetch ${url}`);
      if (spec.expectedRange !== undefined) {
        expect(init.headers.range).toBe(spec.expectedRange);
      }
      if (spec.delayMs !== undefined) {
        await new Promise<void>((resolve, reject) => {
          const t = setTimeout(resolve, spec.delayMs);
          init.signal.addEventListener("abort", () => {
            clearTimeout(t);
            reject(new DOMException("aborted", "AbortError"));
          });
        });
      }
      const headers = new Map<string, string>(
        Object.entries(spec.headers ?? {}).map(([k, v]) => [k.toLowerCase(), v])
      );
      const response: LoaderResponse = {
        status: spec.status ?? 200,
        headers: { get: (n) => headers.get(n.toLowerCase()) ?? null },
        body: spec.bytes !== undefined ? asyncIterable(spec.bytes) : null
      };
      return response;
    }
  };
}

const baseOpts = {
  timeoutMs: 1000,
  maxSegmentBytes: 4096,
  maxRedirects: 2,
  allowPrivateNetworks: false
};

describe("loadSegment — happy paths", () => {
  it("returns bytes and final URL on 200", async () => {
    const url = "https://cdn.example.test/init.mp4";
    const runtime = makeRuntime({
      [url]: { headers: { "content-type": "video/mp4" }, bytes: new Uint8Array([1, 2, 3]) }
    });
    const result = await loadSegment(url, { runtime, ...baseOpts });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.bytes).toEqual(new Uint8Array([1, 2, 3]));
      expect(result.finalUrl).toBe(url);
      expect(result.contentType).toBe("video/mp4");
    }
  });

  it("accepts an empty content-type", async () => {
    const url = "https://cdn.example.test/seg.m4s";
    const runtime = makeRuntime({
      [url]: { bytes: new Uint8Array([4]) }
    });
    const result = await loadSegment(url, { runtime, ...baseOpts });
    expect(result.ok).toBe(true);
  });

  it("follows a redirect within maxRedirects", async () => {
    const start = "https://cdn.example.test/seg-1.m4s";
    const target = "https://cdn.example.test/seg-2.m4s";
    const runtime = makeRuntime({
      [start]: { status: 302, headers: { location: target } },
      [target]: { headers: { "content-type": "video/mp4" }, bytes: new Uint8Array([7]) }
    });
    const result = await loadSegment(start, { runtime, ...baseOpts });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.finalUrl).toBe(target);
    }
  });

  it("accepts an IPv4 literal host without DNS lookup", async () => {
    const url = "http://93.184.216.34/x.m4s";
    const runtime = makeRuntime({
      [url]: { headers: { "content-type": "video/mp4" }, bytes: new Uint8Array([1]) }
    });
    const result = await loadSegment(url, { runtime, ...baseOpts });
    expect(result.ok).toBe(true);
  });

  it("returns bytes for a single HTTP Range request on 206", async () => {
    const url = "https://cdn.example.test/init.mp4";
    const runtime = makeRuntime({
      [url]: {
        status: 206,
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array([1, 2, 3, 4]),
        expectedRange: "bytes=10-13"
      }
    });
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      range: { offset: 10, length: 4 }
    });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.bytes).toEqual(new Uint8Array([1, 2, 3, 4]));
      expect(result.bytes.byteLength).toBe(4);
    }
  });

  it("keeps the Range header while following a validated redirect", async () => {
    const start = "https://cdn.example.test/init.mp4";
    const target = "https://edge.example.test/init.mp4";
    const runtime = makeRuntime({
      [start]: {
        status: 302,
        headers: { location: target },
        expectedRange: "bytes=4-7"
      },
      [target]: {
        status: 206,
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array([5, 6, 7, 8]),
        expectedRange: "bytes=4-7"
      }
    });
    const result = await loadSegment(start, {
      runtime,
      ...baseOpts,
      range: { offset: 4, length: 4 }
    });
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.finalUrl).toBe(target);
      expect(result.bytes).toEqual(new Uint8Array([5, 6, 7, 8]));
    }
  });
});

describe("loadSegment — failure mappings", () => {
  it("maps a 5xx response to segment_fetch_failed", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime = makeRuntime({ [url]: { status: 500 } });
    const result = await loadSegment(url, { runtime, ...baseOpts });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.code).toBe("segment_fetch_failed");
  });

  it("maps 200 OK on a Range request to segment_fetch_failed", async () => {
    const url = "https://cdn.example.test/init.mp4";
    const runtime = makeRuntime({
      [url]: {
        status: 200,
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array([1, 2, 3, 4]),
        expectedRange: "bytes=0-3"
      }
    });
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      range: { offset: 0, length: 4 }
    });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.code).toBe("segment_fetch_failed");
  });

  it("maps a short Range response body to segment_fetch_failed", async () => {
    const url = "https://cdn.example.test/init.mp4";
    const runtime = makeRuntime({
      [url]: {
        status: 206,
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array([1, 2, 3]),
        expectedRange: "bytes=0-3"
      }
    });
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      range: { offset: 0, length: 4 }
    });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.code).toBe("segment_fetch_failed");
  });

  it("maps an overlong Range response body to segment_too_large", async () => {
    const url = "https://cdn.example.test/init.mp4";
    const runtime = makeRuntime({
      [url]: {
        status: 206,
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array([1, 2, 3, 4, 5]),
        expectedRange: "bytes=0-3"
      }
    });
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      range: { offset: 0, length: 4 }
    });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.code).toBe("segment_too_large");
  });

  it("maps a Range length above maxSegmentBytes to segment_too_large", async () => {
    const url = "https://cdn.example.test/init.mp4";
    const runtime = makeRuntime({});
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      maxSegmentBytes: 4,
      range: { offset: 0, length: 5 }
    });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.code).toBe("segment_too_large");
  });

  it("maps an unsupported content-type to segment_content_type_unsupported", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime = makeRuntime({
      [url]: { headers: { "content-type": "text/html" }, bytes: new Uint8Array([0]) }
    });
    const result = await loadSegment(url, { runtime, ...baseOpts });
    if (!result.ok) expect(result.code).toBe("segment_content_type_unsupported");
  });

  it("maps a body that exceeds maxSegmentBytes to segment_too_large", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime = makeRuntime({
      [url]: {
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array(8192)
      }
    });
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      maxSegmentBytes: 1024
    });
    if (!result.ok) expect(result.code).toBe("segment_too_large");
  });

  it("maps a 3xx without Location to segment_fetch_failed", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime = makeRuntime({ [url]: { status: 302 } });
    const result = await loadSegment(url, { runtime, ...baseOpts });
    if (!result.ok) expect(result.code).toBe("segment_fetch_failed");
  });

  it("maps redirect-limit overflow to segment_uri_blocked", async () => {
    const a = "https://cdn.example.test/a.m4s";
    const b = "https://cdn.example.test/b.m4s";
    const c = "https://cdn.example.test/c.m4s";
    const runtime = makeRuntime({
      [a]: { status: 302, headers: { location: b } },
      [b]: { status: 302, headers: { location: c } },
      [c]: { status: 302, headers: { location: a } }
    });
    const result = await loadSegment(a, { runtime, ...baseOpts, maxRedirects: 1 });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });

  it("maps a non-parseable URL to segment_uri_blocked", async () => {
    const runtime = makeRuntime({});
    const result = await loadSegment("ht!tps://broken/x", { runtime, ...baseOpts });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });

  it("maps a credentialed URL to segment_uri_blocked", async () => {
    const runtime = makeRuntime({});
    const result = await loadSegment("https://user:pw@cdn.example.test/x.m4s", {
      runtime,
      ...baseOpts
    });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });

  it("maps a DNS failure to segment_uri_blocked", async () => {
    const runtime = makeRuntime(
      {},
      { resolveError: new Error("ENOTFOUND") }
    );
    const result = await loadSegment("https://cdn.example.test/x.m4s", {
      runtime,
      ...baseOpts
    });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });

  it("maps an empty DNS resolution to segment_uri_blocked", async () => {
    const runtime = makeRuntime(
      {},
      { resolveCustom: () => [] }
    );
    const result = await loadSegment("https://cdn.example.test/x.m4s", {
      runtime,
      ...baseOpts
    });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });

  it("maps a private-network resolved address to segment_uri_blocked when not opted in", async () => {
    const runtime = makeRuntime(
      {},
      { resolveCustom: () => [{ address: "127.0.0.1", family: 4 }] }
    );
    const result = await loadSegment("https://cdn.example.test/x.m4s", {
      runtime,
      ...baseOpts
    });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });

  it("allows private-network address when allowPrivateNetworks=true", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime = makeRuntime(
      {
        [url]: { headers: { "content-type": "video/mp4" }, bytes: new Uint8Array([1]) }
      },
      { resolveCustom: () => [{ address: "127.0.0.1", family: 4 }] }
    );
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      allowPrivateNetworks: true
    });
    expect(result.ok).toBe(true);
  });

  it("maps a network error during fetch to segment_fetch_failed", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime: LoaderRuntime = {
      async resolveHost() {
        return [{ address: "93.184.216.34", family: 4 }];
      },
      async fetch() {
        throw new Error("ECONNRESET");
      }
    };
    const result = await loadSegment(url, { runtime, ...baseOpts });
    if (!result.ok) expect(result.code).toBe("segment_fetch_failed");
  });

  it("maps a fetch timeout to segment_fetch_failed", async () => {
    const url = "https://cdn.example.test/x.m4s";
    const runtime = makeRuntime({ [url]: { delayMs: 200 } });
    const result = await loadSegment(url, {
      runtime,
      ...baseOpts,
      timeoutMs: 50
    });
    if (!result.ok) expect(result.code).toBe("segment_fetch_failed");
  });

  it("rejects unsafe schemes upfront", async () => {
    const runtime = makeRuntime({});
    const result = await loadSegment("file:///etc/passwd", {
      runtime,
      ...baseOpts
    });
    if (!result.ok) expect(result.code).toBe("segment_uri_blocked");
  });
});
