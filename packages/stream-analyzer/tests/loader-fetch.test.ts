import { describe, expect, it } from "vitest";

import type { LoaderRuntime, LoaderResponse, ResolvedAddress } from "../src/internal/loader/runtime.js";
import { loadHlsManifest } from "../src/internal/loader/fetch.js";
import { AnalysisError } from "../src/index.js";
import { analyzeWithRuntime } from "../src/analyze.js";

interface ResponseSpec {
  status: number;
  headers?: Record<string, string>;
  body?: string;
  /** Setze auf "abort" damit die fetch-Antwort die signal bemerkt. */
  delayMs?: number;
}

function makeRuntime(opts: {
  resolve?: ResolvedAddress[] | ((host: string) => ResolvedAddress[]);
  responses: Record<string, ResponseSpec>;
  resolveError?: Error;
}): LoaderRuntime {
  const calls: string[] = [];
  const runtime: LoaderRuntime & { calls: string[] } = {
    calls,
    async resolveHost(hostname) {
      if (opts.resolveError) throw opts.resolveError;
      const resolved = typeof opts.resolve === "function"
        ? opts.resolve(hostname)
        : opts.resolve ?? [{ address: "93.184.216.34", family: 4 }];
      return resolved;
    },
    async fetch(url, init) {
      calls.push(url);
      const spec = opts.responses[url];
      if (!spec) throw new Error(`unexpected fetch ${url}`);
      if (spec.delayMs && spec.delayMs > 0) {
        await new Promise<void>((resolve, reject) => {
          const timer = setTimeout(resolve, spec.delayMs);
          init.signal.addEventListener("abort", () => {
            clearTimeout(timer);
            reject(new DOMException("aborted", "AbortError"));
          });
        });
      }
      const headersMap = new Map<string, string>(
        Object.entries(spec.headers ?? {}).map(([k, v]) => [k.toLowerCase(), v])
      );
      const response: LoaderResponse = {
        status: spec.status,
        headers: { get: (n) => headersMap.get(n.toLowerCase()) ?? null },
        body: spec.body !== undefined ? toAsyncIterable(spec.body) : null
      };
      return response;
    }
  };
  return runtime;
}

function toAsyncIterable(text: string): AsyncIterable<Uint8Array> {
  const encoded = new TextEncoder().encode(text);
  return (async function* () {
    yield encoded;
  })();
}

function loadOpts(
  runtime: LoaderRuntime,
  overrides: Partial<{ timeoutMs: number; maxBytes: number; maxRedirects: number; allowPrivateNetworks: boolean }> = {}
) {
  return {
    runtime,
    timeoutMs: 1000,
    maxBytes: 1_000_000,
    maxRedirects: 3,
    allowPrivateNetworks: false,
    ...overrides
  };
}

describe("loadHlsManifest — happy path", () => {
  it("returns body text and final URL on 200", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://example.test/manifest.m3u8": {
          status: 200,
          headers: { "content-type": "application/vnd.apple.mpegurl" },
          body: "#EXTM3U\n"
        }
      }
    });
    const result = await loadHlsManifest("https://example.test/manifest.m3u8", loadOpts(runtime));
    expect(result.text).toBe("#EXTM3U\n");
    expect(result.finalUrl).toBe("https://example.test/manifest.m3u8");
  });
});

describe("loadHlsManifest — SSRF runtime path", () => {
  it("blocks scheme via fetch-stage validation", async () => {
    const runtime = makeRuntime({ responses: {} });
    await expect(loadHlsManifest("ftp://example.test/x", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
  });

  it("blocks DNS results in private ranges", async () => {
    const runtime = makeRuntime({
      resolve: [{ address: "10.0.0.5", family: 4 }],
      responses: {}
    });
    await expect(loadHlsManifest("https://example.test/x", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
  });

  it("blocks IPv4 literal hostnames without consulting DNS", async () => {
    let resolveCalled = false;
    const runtime: LoaderRuntime = {
      async resolveHost() {
        resolveCalled = true;
        return [];
      },
      async fetch() {
        throw new Error("should not reach");
      }
    };
    await expect(loadHlsManifest("https://127.0.0.1/m.m3u8", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
    expect(resolveCalled).toBe(false);
  });

  it("blocks IPv6 literal hostnames in private ranges", async () => {
    let resolveCalled = false;
    const runtime: LoaderRuntime = {
      async resolveHost() {
        resolveCalled = true;
        return [];
      },
      async fetch() {
        throw new Error("should not reach");
      }
    };
    await expect(loadHlsManifest("https://[fc00::1]/m.m3u8", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
    expect(resolveCalled).toBe(false);
  });

  it("allows private IPv4 results when allowPrivateNetworks=true", async () => {
    const runtime: LoaderRuntime = {
      async resolveHost() {
        return [{ address: "10.0.0.5", family: 4 }];
      },
      async fetch() {
        return {
          status: 200,
          headers: { get: (n) => (n.toLowerCase() === "content-type" ? "application/vnd.apple.mpegurl" : null) },
          body: (async function* () {
            yield new TextEncoder().encode("#EXTM3U\n");
          })()
        };
      }
    };
    const result = await loadHlsManifest(
      "https://internal.test/m.m3u8",
      loadOpts(runtime, { allowPrivateNetworks: true })
    );
    expect(result.text).toBe("#EXTM3U\n");
  });

  it("allowPrivateNetworks does not relax credentials/scheme/redirect rules", async () => {
    const runtime = makeRuntime({ responses: {} });
    // Credentials in URL bleibt geblockt.
    await expect(
      loadHlsManifest("https://user:pass@example.test/m.m3u8", loadOpts(runtime, { allowPrivateNetworks: true }))
    ).rejects.toMatchObject({ code: "fetch_blocked" });
    // ftp-Schema bleibt geblockt.
    await expect(
      loadHlsManifest("ftp://example.test/m.m3u8", loadOpts(runtime, { allowPrivateNetworks: true }))
    ).rejects.toMatchObject({ code: "fetch_blocked" });
  });

  it("blocks IPv6 loopback literal", async () => {
    const runtime: LoaderRuntime = {
      async resolveHost() {
        throw new Error("should not reach");
      },
      async fetch() {
        throw new Error("should not reach");
      }
    };
    await expect(loadHlsManifest("https://[::1]/m.m3u8", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
  });

  it("allows public IPv6 literal hostnames", async () => {
    const runtime: LoaderRuntime = {
      async resolveHost() {
        throw new Error("should not reach");
      },
      async fetch() {
        return {
          status: 200,
          headers: { get: (n) => (n.toLowerCase() === "content-type" ? "application/vnd.apple.mpegurl" : null) },
          body: (async function* () {
            yield new TextEncoder().encode("#EXTM3U\n");
          })()
        };
      }
    };
    const result = await loadHlsManifest("https://[2001:4860:4860::8888]/m.m3u8", loadOpts(runtime));
    expect(result.text).toBe("#EXTM3U\n");
  });

  it("blocks if any resolved address is private (mixed dual-stack response)", async () => {
    const runtime = makeRuntime({
      resolve: [
        { address: "1.1.1.1", family: 4 },
        { address: "::1", family: 6 }
      ],
      responses: {}
    });
    await expect(loadHlsManifest("https://example.test/x", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
  });

  it("blocks credentials in the URL", async () => {
    const runtime = makeRuntime({ responses: {} });
    await expect(
      loadHlsManifest("https://user:pass@example.test/x", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_blocked" });
  });

  it("turns DNS lookup failures into fetch_blocked", async () => {
    const runtime = makeRuntime({
      resolve: [],
      resolveError: new Error("ENOTFOUND"),
      responses: {}
    });
    await expect(loadHlsManifest("https://example.test/x", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
  });

  it("returns fetch_blocked when DNS yields zero entries", async () => {
    const runtime = makeRuntime({
      resolve: [],
      responses: {}
    });
    await expect(loadHlsManifest("https://example.test/x", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_blocked"
    });
  });
});

describe("loadHlsManifest — redirects", () => {
  it("follows up to maxRedirects", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://a.test/m.m3u8": {
          status: 302,
          headers: { location: "https://b.test/m.m3u8" }
        },
        "https://b.test/m.m3u8": {
          status: 200,
          headers: { "content-type": "application/x-mpegurl" },
          body: "#EXTM3U\nseg.ts\n"
        }
      }
    });
    const result = await loadHlsManifest("https://a.test/m.m3u8", loadOpts(runtime, { maxRedirects: 1 }));
    expect(result.text).toBe("#EXTM3U\nseg.ts\n");
    expect(result.finalUrl).toBe("https://b.test/m.m3u8");
  });

  it("rejects when redirect limit is exceeded", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://a.test/m.m3u8": { status: 302, headers: { location: "https://b.test/m.m3u8" } },
        "https://b.test/m.m3u8": { status: 302, headers: { location: "https://c.test/m.m3u8" } },
        "https://c.test/m.m3u8": { status: 302, headers: { location: "https://d.test/m.m3u8" } }
      }
    });
    await expect(
      loadHlsManifest("https://a.test/m.m3u8", loadOpts(runtime, { maxRedirects: 1 }))
    ).rejects.toMatchObject({ code: "fetch_blocked" });
  });

  it("rejects redirects that point to a private IP", async () => {
    const runtime = makeRuntime({
      resolve: (host) =>
        host === "internal.test"
          ? [{ address: "10.0.0.1", family: 4 }]
          : [{ address: "1.1.1.1", family: 4 }],
      responses: {
        "https://public.test/m.m3u8": {
          status: 301,
          headers: { location: "https://internal.test/secrets" }
        }
      }
    });
    await expect(
      loadHlsManifest("https://public.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_blocked" });
  });

  it("rejects redirect responses that lack a Location header", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://a.test/m.m3u8": { status: 302, headers: {} }
      }
    });
    await expect(loadHlsManifest("https://a.test/m.m3u8", loadOpts(runtime))).rejects.toMatchObject({
      code: "fetch_failed"
    });
  });

  it("enforces size limit on the post-redirect body", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://a.test/m.m3u8": { status: 302, headers: { location: "https://b.test/m.m3u8" } },
        "https://b.test/m.m3u8": {
          status: 200,
          headers: { "content-type": "application/vnd.apple.mpegurl" },
          body: "#EXTM3U\n" + "x".repeat(2_000)
        }
      }
    });
    await expect(
      loadHlsManifest("https://a.test/m.m3u8", loadOpts(runtime, { maxBytes: 100 }))
    ).rejects.toMatchObject({ code: "manifest_too_large" });
  });
});

describe("loadHlsManifest — fetch failures", () => {
  it("maps non-2xx statuses to fetch_failed", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://example.test/m.m3u8": { status: 404, headers: {}, body: "not found" }
      }
    });
    await expect(
      loadHlsManifest("https://example.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_failed" });
  });

  it("rejects unsupported content-types", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://example.test/m.m3u8": {
          status: 200,
          headers: { "content-type": "text/html" },
          body: "<html>"
        }
      }
    });
    await expect(
      loadHlsManifest("https://example.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_failed" });
  });

  it("accepts a missing content-type as text fallback", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://example.test/m.m3u8": { status: 200, headers: {}, body: "#EXTM3U\n" }
      }
    });
    const result = await loadHlsManifest(
      "https://example.test/m.m3u8",
      loadOpts(runtime)
    );
    expect(result.text).toBe("#EXTM3U\n");
  });

  it("turns timeouts into fetch_failed", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://slow.test/m.m3u8": {
          status: 200,
          headers: { "content-type": "application/vnd.apple.mpegurl" },
          body: "#EXTM3U\n",
          delayMs: 100
        }
      }
    });
    await expect(
      loadHlsManifest("https://slow.test/m.m3u8", loadOpts(runtime, { timeoutMs: 10 }))
    ).rejects.toMatchObject({ code: "fetch_failed" });
  });

  it("aborts slow body streams on timeout, not just header phase", async () => {
    // Slow-Loris-Stub: Header sind sofort da, der Body trickelt
    // jedoch über `setTimeout`-Chunks. Der Timer muss den Body
    // abbrechen, sonst hängt der Loader bis maxBytes (Slow-Loris).
    const runtime: LoaderRuntime = {
      async resolveHost() {
        return [{ address: "1.1.1.1", family: 4 }];
      },
      async fetch(_url, init) {
        async function* slowBody() {
          for (let i = 0; i < 1000; i++) {
            await new Promise<void>((resolve, reject) => {
              const t = setTimeout(resolve, 50);
              init.signal.addEventListener(
                "abort",
                () => {
                  clearTimeout(t);
                  reject(new DOMException("aborted", "AbortError"));
                },
                { once: true }
              );
            });
            yield new TextEncoder().encode("x");
          }
        }
        return {
          status: 200,
          headers: { get: (n) => (n.toLowerCase() === "content-type" ? "application/vnd.apple.mpegurl" : null) },
          body: slowBody()
        };
      }
    };
    const start = Date.now();
    await expect(
      loadHlsManifest("https://slow.test/m.m3u8", loadOpts(runtime, { timeoutMs: 50, maxBytes: 100_000 }))
    ).rejects.toMatchObject({ code: "fetch_failed", message: expect.stringContaining("Timeout") });
    expect(Date.now() - start).toBeLessThan(2_000); // nicht maxBytes-gebunden
  });

  it("wraps generic fetch errors as fetch_failed", async () => {
    const runtime: LoaderRuntime = {
      resolveHost: async () => [{ address: "1.1.1.1", family: 4 }],
      fetch: async () => {
        throw new Error("ECONNRESET");
      }
    };
    await expect(
      loadHlsManifest("https://example.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_failed" });
  });

  it("describes non-Error throw values via String()", async () => {
    const runtime: LoaderRuntime = {
      resolveHost: async () => [{ address: "1.1.1.1", family: 4 }],
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      fetch: async () => {
        throw "raw string boom";
      }
    };
    await expect(
      loadHlsManifest("https://example.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_failed", message: expect.stringContaining("raw string boom") });
  });

  it("turns non-Error DNS rejections into fetch_blocked with String() reason", async () => {
    const runtime: LoaderRuntime = {
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      resolveHost: async () => {
        throw "non-error dns failure";
      },
      fetch: async () => ({ status: 200, headers: { get: () => null }, body: null })
    };
    await expect(
      loadHlsManifest("https://example.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_blocked", message: expect.stringContaining("non-error dns failure") });
  });

  it("rejects unparseable URLs with invalid_input", async () => {
    const runtime = makeRuntime({ responses: {} });
    await expect(loadHlsManifest("not a url", loadOpts(runtime))).rejects.toBeInstanceOf(AnalysisError);
  });
});

describe("analyzeWithRuntime — URL pipeline", () => {
  it("integrates loader output into AnalysisResult", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://cdn.example.test/manifest.m3u8": {
          status: 200,
          headers: { "content-type": "application/vnd.apple.mpegurl" },
          body: "#EXTM3U\n#EXT-X-TARGETDURATION:6\n#EXTINF:6.0,\nseg.ts\n"
        }
      }
    });
    const result = await analyzeWithRuntime(
      { kind: "url", url: "https://cdn.example.test/manifest.m3u8" },
      {},
      runtime
    );
    expect(result.status).toBe("ok");
    if (result.status !== "ok") return;
    expect(result.playlistType).toBe("media");
    expect(result.input).toEqual({
      source: "url",
      url: "https://cdn.example.test/manifest.m3u8",
      baseUrl: "https://cdn.example.test/manifest.m3u8"
    });
  });

  it("propagates SSRF rejections through the public surface", async () => {
    const runtime = makeRuntime({
      resolve: [{ address: "127.0.0.1", family: 4 }],
      responses: {}
    });
    const result = await analyzeWithRuntime(
      { kind: "url", url: "https://example.test/manifest.m3u8" },
      {},
      runtime
    );
    expect(result.status).toBe("error");
    if (result.status !== "error") return;
    expect(result.code).toBe("fetch_blocked");
  });

  it("respects fetch options (maxBytes)", async () => {
    const runtime = makeRuntime({
      responses: {
        "https://cdn.example.test/manifest.m3u8": {
          status: 200,
          headers: { "content-type": "application/vnd.apple.mpegurl" },
          body: "#EXTM3U\n" + "x".repeat(500)
        }
      }
    });
    const result = await analyzeWithRuntime(
      { kind: "url", url: "https://cdn.example.test/manifest.m3u8" },
      { fetch: { maxBytes: 100 } },
      runtime
    );
    expect(result.status).toBe("error");
    if (result.status !== "error") return;
    expect(result.code).toBe("manifest_too_large");
  });
});

describe("DNS-Rebinding-Entscheidung (Dokumentationspunkt)", () => {
  it("dokumentiert: Loader prüft alle Adressen einmal beim Lookup", async () => {
    // Diese DoD-Position aus plan-0.3.0 §3 verlangt eine dokumentierte
    // DNS-Rebinding-Entscheidung. Der Loader nimmt die Lookup-Antwort als
    // Set, prüft jeden Eintrag gegen die Sperrliste und delegiert die
    // eigentliche Verbindung an die globale `fetch`-Implementierung — ein
    // perfekter Schutz gegen Rebinding zwischen Lookup und TCP-Connect ist
    // in 0.3.0 nicht garantiert. Die Doku in `docs/user/stream-analyzer.md`
    // §6 hält die Entscheidung fest.
    const calls: string[] = [];
    const runtime: LoaderRuntime = {
      async resolveHost(host) {
        calls.push(`resolve:${host}`);
        return [
          { address: "1.1.1.1", family: 4 },
          { address: "10.0.0.5", family: 4 } // sollte den Bail-out auslösen
        ];
      },
      async fetch() {
        calls.push("fetch");
        throw new Error("should not reach");
      }
    };
    await expect(
      loadHlsManifest("https://example.test/m.m3u8", loadOpts(runtime))
    ).rejects.toMatchObject({ code: "fetch_blocked" });
    expect(calls).toEqual(["resolve:example.test"]);
  });
});
