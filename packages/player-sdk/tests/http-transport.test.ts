import { describe, expect, it, vi } from "vitest";
import { HttpTransport } from "../src/transport/http";
import type { PlaybackEventBatch } from "../src/types/events";

const batch: PlaybackEventBatch = {
  schema_version: "1.0",
  events: []
};

type TestFetch = (input: string, init: RequestInit) => Promise<Response>;

describe("HttpTransport", () => {
  it("sends batches with the ingest headers", async () => {
    const fetchFn = vi.fn<TestFetch>(async () => new Response(null, { status: 204 }));
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", { fetchFn });

    await transport.send(batch);

    expect(fetchFn).toHaveBeenCalledTimes(1);
    expect(fetchFn.mock.calls[0]?.[0]).toBe("http://localhost:8080/api/playback-events");
    expect(fetchFn.mock.calls[0]?.[1]).toMatchObject({
      method: "POST",
      credentials: "omit",
      headers: {
        "Content-Type": "application/json",
        "X-MTrace-Token": "demo-token"
      },
      body: JSON.stringify(batch)
    });
  });

  it("retries transient 5xx responses with bounded backoff", async () => {
    const fetchFn = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(new Response(null, { status: 503 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const sleeps: number[] = [];
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn,
      maxAttempts: 3,
      baseDelayMs: 25,
      sleep: async (ms) => {
        sleeps.push(ms);
      }
    });

    await transport.send(batch);

    expect(fetchFn).toHaveBeenCalledTimes(2);
    expect(sleeps).toEqual([25]);
  });

  it("respects Retry-After for 429 before retrying", async () => {
    const order: string[] = [];
    const fetchFn = vi
      .fn<TestFetch>()
      .mockImplementationOnce(async () => {
        order.push("fetch:429");
        return new Response(null, { status: 429, headers: { "Retry-After": "2" } });
      })
      .mockImplementationOnce(async () => {
        order.push("fetch:204");
        return new Response(null, { status: 204 });
      });
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn,
      sleep: async (ms) => {
        order.push(`sleep:${ms}`);
      }
    });

    await transport.send(batch);

    expect(order).toEqual(["fetch:429", "sleep:2000", "fetch:204"]);
  });

  it("uses backoff for absent or invalid Retry-After values", async () => {
    const sleeps: number[] = [];
    const fetchFn = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(new Response(null, { status: 429 }))
      .mockResolvedValueOnce(new Response(null, { status: 429, headers: { "Retry-After": "not-a-date" } }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn,
      baseDelayMs: 10,
      sleep: async (ms) => {
        sleeps.push(ms);
      }
    });

    await transport.send(batch);

    expect(sleeps).toEqual([10, 20]);
  });

  it("uses HTTP-date Retry-After values as cooldowns", async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-04-30T12:00:00.000Z"));
    const sleeps: number[] = [];
    const fetchFn = vi
      .fn<TestFetch>()
      .mockResolvedValueOnce(
        new Response(null, {
          status: 429,
          headers: { "Retry-After": "Thu, 30 Apr 2026 12:00:03 GMT" }
        })
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn,
      sleep: async (ms) => {
        sleeps.push(ms);
      }
    });

    try {
      await transport.send(batch);
    } finally {
      vi.useRealTimers();
    }

    expect(sleeps).toEqual([3000]);
  });

  it("does not retry non-transient 4xx responses or 413 payload rejections", async () => {
    const badRequestFetch = vi.fn<TestFetch>(async () => new Response(null, { status: 400 }));
    const payloadTooLargeFetch = vi.fn<TestFetch>(async () => new Response(null, { status: 413 }));

    await expect(
      new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
        fetchFn: badRequestFetch
      }).send(batch)
    ).rejects.toThrow("400");
    await expect(
      new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
        fetchFn: payloadTooLargeFetch
      }).send(batch)
    ).rejects.toThrow("413");

    expect(badRequestFetch).toHaveBeenCalledTimes(1);
    expect(payloadTooLargeFetch).toHaveBeenCalledTimes(1);
  });

  it("retries network errors and request timeouts", async () => {
    const networkFetch = vi
      .fn<TestFetch>()
      .mockRejectedValueOnce(new TypeError("network failed"))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const timeoutFetch = vi
      .fn<TestFetch>()
      .mockImplementationOnce(
        async (_input: string, init: RequestInit) =>
          new Promise<Response>((_resolve, reject) => {
            init.signal?.addEventListener("abort", () => {
              reject(new Error("aborted"));
            });
          })
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));

    await new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn: networkFetch,
      sleep: async () => undefined
    }).send(batch);
    await new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn: timeoutFetch,
      timeoutMs: 1,
      sleep: async () => undefined
    }).send(batch);

    expect(networkFetch).toHaveBeenCalledTimes(2);
    expect(timeoutFetch).toHaveBeenCalledTimes(2);
  });

  it("throws the final network error after exhausting retries", async () => {
    const fetchFn = vi.fn<TestFetch>(async () => {
      throw new TypeError("network stayed down");
    });
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn,
      maxAttempts: 2,
      sleep: async () => undefined
    });

    await expect(transport.send(batch)).rejects.toThrow("network stayed down");
    expect(fetchFn).toHaveBeenCalledTimes(2);
  });

  it("supports disabled request timeouts", async () => {
    const fetchFn = vi.fn<TestFetch>(async (_input, init) => {
      expect(init.signal).toBeUndefined();
      return new Response(null, { status: 204 });
    });
    const transport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
      fetchFn,
      timeoutMs: 0
    });

    await transport.send(batch);

    expect(fetchFn).toHaveBeenCalledTimes(1);
  });
});
