import { afterEach, beforeEach, describe, expect, it } from "vitest";
import type { AddressInfo } from "node:net";
import type { Server } from "node:http";

import type { AnalyzeOptions, AnalyzeOutput, ManifestInput } from "@npm9912/stream-analyzer";

import { createAnalyzerServer } from "../src/server.js";

interface RunningServer {
  readonly url: string;
  readonly server: Server;
}

type AnalyzeFn = (input: ManifestInput, options?: AnalyzeOptions) => Promise<AnalyzeOutput>;

async function startServer(analyze: AnalyzeFn, allowPrivateNetworks = false): Promise<RunningServer> {
  const server = createAnalyzerServer({ analyze, allowPrivateNetworks });
  await new Promise<void>((resolve) => server.listen(0, "127.0.0.1", () => resolve()));
  const addr = server.address() as AddressInfo;
  return { url: `http://127.0.0.1:${addr.port}`, server };
}

let running: RunningServer | null = null;

beforeEach(() => {
  running = null;
});

afterEach(async () => {
  if (running !== null) {
    await new Promise<void>((resolve) => running!.server.close(() => resolve()));
    running = null;
  }
});

describe("analyzer-service — health", () => {
  it("returns 200 ok on GET /health", async () => {
    running = await startServer(async () => ({} as AnalyzeOutput));
    const res = await fetch(`${running.url}/health`);
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ status: "ok" });
  });
});

describe("analyzer-service — POST /analyze", () => {
  it("forwards the analyzer result on a valid text request", async () => {
    const stubResult = {
      status: "ok",
      analyzerVersion: "0.4.0",
      analyzerKind: "hls",
      input: { source: "text" },
      playlistType: "unknown",
      summary: { itemCount: 0 },
      findings: [],
      details: null
    } as AnalyzeOutput;
    running = await startServer(async (input) => {
      expect(input).toEqual({ kind: "text", text: "#EXTM3U\n" });
      return stubResult;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "text", text: "#EXTM3U\n" })
    });
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual(stubResult);
  });

  it("forwards baseUrl when provided", async () => {
    running = await startServer(async (input) => {
      expect(input).toEqual({ kind: "text", text: "#EXTM3U\n", baseUrl: "https://cdn.test/" });
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "text", text: "#EXTM3U\n", baseUrl: "https://cdn.test/" })
    });
    expect(res.status).toBe(200);
  });

  it("forwards url-input as-is", async () => {
    running = await startServer(async (input) => {
      expect(input).toEqual({ kind: "url", url: "https://example.test/m.m3u8" });
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "url", url: "https://example.test/m.m3u8" })
    });
    expect(res.status).toBe(200);
  });

  it("passes fetch options through and filters the negative redirect value", async () => {
    running = await startServer(async (_input, options) => {
      expect(options).toEqual({ fetch: { timeoutMs: 5000, maxBytes: 1024 } });
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        kind: "url",
        url: "https://example.test/m.m3u8",
        fetch: { timeoutMs: 5000, maxBytes: 1024, maxRedirects: -1 }
      })
    });
    // maxRedirects: -1 is filtered (< 0), only timeoutMs and maxBytes pass through.
    expect(res.status).toBe(200);
  });

  it("accepts maxRedirects: 0 (explicit no-redirects directive)", async () => {
    running = await startServer(async (_input, options) => {
      expect(options).toEqual({ fetch: { maxRedirects: 0 } });
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        kind: "url",
        url: "https://example.test/m.m3u8",
        fetch: { maxRedirects: 0 }
      })
    });
    expect(res.status).toBe(200);
  });

  it("rejects non-JSON content type with 415", async () => {
    running = await startServer(async () => ({} as AnalyzeOutput));
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "text/plain" },
      body: "{}"
    });
    expect(res.status).toBe(415);
    expect(await res.json()).toMatchObject({ status: "error", code: "unsupported_media_type" });
  });

  it("rejects malformed JSON with 400", async () => {
    running = await startServer(async () => ({} as AnalyzeOutput));
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: "{not json"
    });
    expect(res.status).toBe(400);
    expect(await res.json()).toMatchObject({ status: "error", code: "invalid_json" });
  });

  const invalidRequests: Array<{ body: unknown; label: string }> = [
    { body: {}, label: "missing kind" },
    { body: { kind: "text" }, label: "missing text" },
    { body: { kind: "text", text: 42 }, label: "non-string text" },
    { body: { kind: "text", text: "x", baseUrl: 5 }, label: "non-string baseUrl" },
    { body: { kind: "url" }, label: "missing url" },
    { body: { kind: "url", url: "" }, label: "empty url" },
    { body: { kind: "binary" }, label: "unknown kind" }
  ];
  it.each(invalidRequests)("rejects invalid request: $label", async ({ body }) => {
    running = await startServer(async () => ({} as AnalyzeOutput));
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body)
    });
    expect(res.status).toBe(400);
    expect(await res.json()).toMatchObject({ status: "error", code: "invalid_request" });
  });

  it("rejects bodies larger than the request limit with 413", async () => {
    running = await startServer(async () => ({} as AnalyzeOutput));
    const big = "x".repeat(2_000_000);
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "text", text: big })
    });
    expect(res.status).toBe(413);
  });

  it("translates analyzer throws into 500 problem responses", async () => {
    running = await startServer(async () => {
      throw new Error("boom");
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "text", text: "#EXTM3U\n" })
    });
    expect(res.status).toBe(500);
    expect(await res.json()).toMatchObject({ status: "error", code: "internal_error" });
  });

  it("describes non-Error throws via String() in the 500 response", async () => {
    running = await startServer(async () => {
      throw "raw boom";
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "text", text: "#EXTM3U\n" })
    });
    expect(res.status).toBe(500);
    expect(await res.json()).toMatchObject({ status: "error", code: "internal_error", message: expect.stringContaining("raw boom") });
  });

  it("ignores fetch options with invalid types or out-of-range values", async () => {
    running = await startServer(async (_input, options) => {
      // Alle drei Felder werden gefiltert, also kein options.fetch ankommt.
      expect(options).toBeUndefined();
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        kind: "url",
        url: "https://example.test/m.m3u8",
        fetch: { timeoutMs: 0, maxBytes: -10, maxRedirects: "many" }
      })
    });
    expect(res.status).toBe(200);
  });

  it("ignores a non-object fetch field", async () => {
    running = await startServer(async (_input, options) => {
      expect(options).toBeUndefined();
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "url", url: "https://example.test/m.m3u8", fetch: "broken" })
    });
    expect(res.status).toBe(200);
  });
});

describe("analyzer-service — allowPrivateNetworks", () => {
  it("forces fetch.allowPrivateNetworks=true on every analyze call when the flag is set", async () => {
    let observedOptions: AnalyzeOptions | undefined;
    running = await startServer(async (_input, options) => {
      observedOptions = options;
      return { status: "ok" } as unknown as AnalyzeOutput;
    }, true);
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "url", url: "https://example.test/m.m3u8" })
    });
    expect(res.status).toBe(200);
    expect(observedOptions).toEqual({ fetch: { allowPrivateNetworks: true } });
  });

  it("merges service flag with body fetch-options", async () => {
    let observedOptions: AnalyzeOptions | undefined;
    running = await startServer(async (_input, options) => {
      observedOptions = options;
      return { status: "ok" } as unknown as AnalyzeOutput;
    }, true);
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        kind: "url",
        url: "https://example.test/m.m3u8",
        fetch: { timeoutMs: 5000 }
      })
    });
    expect(res.status).toBe(200);
    expect(observedOptions).toEqual({ fetch: { timeoutMs: 5000, allowPrivateNetworks: true } });
  });

  it("does NOT honor a body-set allowPrivateNetworks when env is false", async () => {
    // Defense-in-Depth: parseFetchOptions im Service whitelistet die
    // erlaubten Felder; allowPrivateNetworks ist nicht dabei. Damit
    // kann ein Aufrufer das Flag nicht über den Body bypass-en, wenn
    // der Operator es per Env nicht gesetzt hat. Dieser Test pinnt
    // genau diese Garantie — ein zukünftiges Whitelist-Update muss
    // entweder den Test brechen oder das Verhalten erhalten.
    let observedOptions: AnalyzeOptions | undefined;
    running = await startServer(async (_input, options) => {
      observedOptions = options;
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        kind: "url",
        url: "https://example.test/m.m3u8",
        fetch: { allowPrivateNetworks: true, timeoutMs: 1234 }
      })
    });
    expect(res.status).toBe(200);
    // timeoutMs darf durch (whitelisted), allowPrivateNetworks nicht.
    expect(observedOptions).toEqual({ fetch: { timeoutMs: 1234 } });
  });

  it("ignores the flag when the env did not set it (default)", async () => {
    let observedOptions: AnalyzeOptions | undefined;
    running = await startServer(async (_input, options) => {
      observedOptions = options;
      return { status: "ok" } as unknown as AnalyzeOutput;
    });
    const res = await fetch(`${running.url}/analyze`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ kind: "url", url: "https://example.test/m.m3u8" })
    });
    expect(res.status).toBe(200);
    expect(observedOptions).toBeUndefined();
  });
});

describe("analyzer-service — unknown routes", () => {
  it("returns 404 for unmapped paths", async () => {
    running = await startServer(async () => ({} as AnalyzeOutput));
    const res = await fetch(`${running.url}/anything`);
    expect(res.status).toBe(404);
    expect(await res.json()).toMatchObject({ status: "error", code: "not_found" });
  });
});
