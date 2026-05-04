import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";
import { consumeFrames, startSseClient, type AppendedFrame } from "../src/lib/sse-client";
import { sseConnection } from "../src/lib/status";

beforeEach(() => {
  sseConnection.set({ state: "not_yet_connected", changedAt: null, detail: null });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("consumeFrames edge cases", () => {
  it("handles a field-only line without colon", () => {
    // EventSource-Spec: line ohne `:` wird komplett als field-name
    // mit value=""behandelt. Ein "id"-only line setzt id="".
    const got = consumeFrames("id\nevent: x\ndata: y\n\n");
    expect(got.frames).toHaveLength(1);
    expect(got.frames[0]?.id).toBe("");
    expect(got.frames[0]?.event).toBe("x");
  });

  it("trims an exact `: ` prefix only once", () => {
    const got = consumeFrames("data:  leading-space\n\n");
    // Spec: nur EIN Space wird getrimmt; weitere bleiben Teil des Werts.
    expect(got.frames[0]?.data).toBe(" leading-space");
  });

  it("preserves CR-LF line endings", () => {
    const got = consumeFrames("event: x\r\ndata: y\r\n\n");
    expect(got.frames).toHaveLength(1);
    expect(got.frames[0]?.event).toBe("x");
    expect(got.frames[0]?.data).toBe("y");
  });
});

describe("consumeFrames (WHATWG SSE-Parser)", () => {
  it("returns empty when buffer has no terminator", () => {
    const got = consumeFrames("event: foo\ndata: 1");
    expect(got.frames).toEqual([]);
    expect(got.remainder).toBe("event: foo\ndata: 1");
  });

  it("dispatches a single frame on \\n\\n", () => {
    const got = consumeFrames("id: 7\nevent: event_appended\ndata: {\"a\":1}\n\n");
    expect(got.frames).toEqual([{ id: "7", event: "event_appended", data: '{"a":1}' }]);
    expect(got.remainder).toBe("");
  });

  it("ignores comment lines starting with colon", () => {
    const got = consumeFrames(": heartbeat\n\nid: 8\nevent: x\ndata: y\n\n");
    expect(got.frames).toHaveLength(1);
    expect(got.frames[0]?.id).toBe("8");
  });

  it("strips a single space after the colon and concatenates multi-line data", () => {
    const got = consumeFrames("event: x\ndata: line1\ndata: line2\n\n");
    expect(got.frames[0]?.data).toBe("line1\nline2");
  });

  it("keeps an unfinished tail in remainder", () => {
    const got = consumeFrames("event: a\ndata: 1\n\nevent: b\ndata: 2");
    expect(got.frames).toHaveLength(1);
    expect(got.remainder).toBe("event: b\ndata: 2");
  });
});

describe("startSseClient", () => {
  function streamFromChunks(chunks: string[]): ReadableStream<Uint8Array> {
    const encoder = new TextEncoder();
    let i = 0;
    return new ReadableStream({
      pull(controller) {
        if (i < chunks.length) {
          controller.enqueue(encoder.encode(chunks[i++]));
        } else {
          controller.close();
        }
      }
    });
  }

  it("dispatches event_appended frames to the onAppended callback", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(streamFromChunks([
        "id: 1\nevent: event_appended\ndata: {\"project_id\":\"demo\",\"session_id\":\"s\",\"ingest_sequence\":1,\"event_name\":\"manifest_loaded\"}\n\n"
      ]), { status: 200 })
    );
    const seen: AppendedFrame[] = [];
    let connectedSeen = false;
    const off = sseConnection.subscribe((s) => {
      if (s.state === "connected") connectedSeen = true;
    });
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: (f) => seen.push(f),
      fetchFn: fetchMock,
      schedule: () => () => undefined // suppress reconnect for this test
    });
    await flushAsyncTicks();
    expect(seen).toHaveLength(1);
    expect(seen[0].ingest_sequence).toBe(1);
    // `connected` ist transient während der Reader-Loop läuft;
    // die Subscriber-Aufzeichnung pinnt, dass der State mindestens
    // einmal sichtbar war.
    expect(connectedSeen).toBe(true);
    client.disconnect();
    expect(get(sseConnection).state).toBe("disabled");
    off();
  });

  it("attaches the X-MTrace-Token header to the fetch", async () => {
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () => new Response(streamFromChunks([]), { status: 200 })
    );
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule: () => () => undefined
    });
    await flushAsyncTicks();
    expect(fetchMock).toHaveBeenCalledOnce();
    const init = fetchMock.mock.calls[0][1] ?? {};
    expect((init.headers as Record<string, string>)["X-MTrace-Token"]).toBe("demo-token");
    client.disconnect();
  });

  // boundedSchedule fires nur die ersten N callbacks; danach no-op.
  // Damit lassen sich Reconnect-Loops synchron testen, ohne in eine
  // unendliche Rekursion zu laufen.
  function boundedSchedule(maxCalls: number) {
    let count = 0;
    return (cb: () => void) => {
      if (count < maxCalls) {
        count += 1;
        cb();
      }
      return () => undefined;
    };
  }

  it("sends Last-Event-ID on reconnect after a frame", async () => {
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () =>
        new Response(
          streamFromChunks([
            "id: 5\nevent: event_appended\ndata: {\"project_id\":\"demo\",\"session_id\":\"s\",\"ingest_sequence\":5,\"event_name\":\"x\"}\n\n"
          ]),
          { status: 200 }
        )
    );
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule: boundedSchedule(2)
    });
    for (let i = 0; i < 5; i += 1) await flushAsyncTicks();
    // Mindestens zwei fetch-Calls (initial + reconnect); zweiter Call
    // trägt Last-Event-ID = "5".
    expect(fetchMock.mock.calls.length).toBeGreaterThanOrEqual(2);
    const reconnectInit = fetchMock.mock.calls[1][1] ?? {};
    expect((reconnectInit.headers as Record<string, string>)["Last-Event-ID"]).toBe("5");
    client.disconnect();
  });

  it("falls back to polling after persistent SSE failure", async () => {
    const fetchMock = vi.fn(async () => {
      throw new Error("network unreachable");
    });
    const onPollingTick = vi.fn(async () => undefined);
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      onPollingTick,
      fetchFn: fetchMock,
      schedule: boundedSchedule(4)
    });
    // Drei Reconnect-Versuche scheitern → Polling-Fallback aktiv.
    for (let i = 0; i < 8; i += 1) await flushAsyncTicks();
    expect(get(sseConnection).state).toBe("polling_fallback");
    expect(onPollingTick).toHaveBeenCalled();
    client.disconnect();
  });

  it("treats non-2xx responses as connection errors", async () => {
    const fetchMock = vi.fn(async () => new Response("nope", { status: 401 }));
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule: boundedSchedule(4)
    });
    for (let i = 0; i < 8; i += 1) await flushAsyncTicks();
    // 401 zählt als Connection-Error → nach 3 Versuchen
    // polling_fallback.
    expect(get(sseConnection).state).toBe("polling_fallback");
    client.disconnect();
  });

  it("ignores backfill_truncated frames without an onTruncated callback", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(streamFromChunks([
        "event: backfill_truncated\ndata: {\"oldest_ingest_sequence\":99}\n\n"
      ]), { status: 200 })
    );
    const onAppended = vi.fn();
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended,
      // onTruncated bewusst undefined → Frame wird silent gedroppt
      fetchFn: fetchMock,
      schedule: () => () => undefined
    });
    await flushAsyncTicks();
    expect(onAppended).not.toHaveBeenCalled();
    client.disconnect();
  });

  it("handles backfill_truncated frames via onTruncated", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(streamFromChunks([
        "event: backfill_truncated\ndata: {\"oldest_ingest_sequence\":42}\n\n"
      ]), { status: 200 })
    );
    const onTruncated = vi.fn();
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      onTruncated,
      fetchFn: fetchMock
    });
    await flushAsyncTicks();
    expect(onTruncated).toHaveBeenCalledWith({ oldest_ingest_sequence: 42 });
    client.disconnect();
  });

  it("ignores frames with malformed JSON data", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(streamFromChunks([
        // event_appended mit unparsable JSON → kein onAppended-Call
        "event: event_appended\ndata: {not-json\n\n"
      ]), { status: 200 })
    );
    const onAppended = vi.fn();
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended,
      fetchFn: fetchMock,
      schedule: () => () => undefined
    });
    await flushAsyncTicks();
    expect(onAppended).not.toHaveBeenCalled();
    client.disconnect();
  });

  it("treats response without a body as a connection error", async () => {
    // Synthetischer Fall: Response.ok=true aber body=null. Im
    // Real-World selten, aber defensiv gepinnt.
    const fetchMock = vi.fn(async () => {
      const r = new Response(null, { status: 200 });
      Object.defineProperty(r, "body", { value: null });
      return r;
    });
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule: boundedSchedule(0)
    });
    await flushAsyncTicks();
    // Nach 1 Versuch ohne body → connecting-Pfad mit detail
    expect(get(sseConnection).state).toBe("connecting");
    client.disconnect();
  });

  it("polling tick reschedules itself after each call", async () => {
    const fetchMock = vi.fn(async () => {
      throw new Error("offline");
    });
    let tickCalls = 0;
    const onPollingTick = vi.fn(async () => {
      tickCalls += 1;
    });
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      onPollingTick,
      fetchFn: fetchMock,
      schedule: boundedSchedule(8)
    });
    for (let i = 0; i < 12; i += 1) await flushAsyncTicks();
    // Mehrfach-Tick-Aufrufe pinnen den re-schedule-Pfad in
    // pollingTick (Zeile 158-160 im Code).
    expect(tickCalls).toBeGreaterThanOrEqual(2);
    client.disconnect();
  });

  it("disconnect cancels a pending reconnect timer (F1)", async () => {
    // Server liefert 500 → Reconnect wird mit Backoff scheduled.
    // schedule speichert den Timer-Callback; disconnect() muss den
    // noch nicht gefeuerten Timer kassieren und auch ein spaeteres
    // Callback-Feuern darf keinen zweiten Fetch starten.
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>(
      async () => new Response("nope", { status: 500 })
    );
    const cancelCalls: string[] = [];
    const callbacks: Array<() => void> = [];
    let scheduleId = 0;
    const schedule = (cb: () => void) => {
      const id = `t${++scheduleId}`;
      callbacks.push(cb);
      return () => {
        cancelCalls.push(id);
      };
    };
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule
    });
    await flushAsyncTicks();
    // Initialer fetch ist passiert, ein Reconnect-Timer ist scheduled.
    expect(scheduleId).toBeGreaterThanOrEqual(1);
    expect(cancelCalls).toEqual([]);
    client.disconnect();
    // disconnect() muss die zuletzt gespeicherte Cancel-Handle gerufen
    // haben (mindestens eine cancel-call landet im Array).
    expect(cancelCalls.length).toBeGreaterThanOrEqual(1);
    callbacks[0]?.();
    await flushAsyncTicks();
    expect(fetchMock).toHaveBeenCalledOnce();
  });

  it("fired reconnect timer is no longer cancelled on later disconnect", async () => {
    const fetchMock = vi
      .fn<(url: string, init?: RequestInit) => Promise<Response>>()
      .mockResolvedValueOnce(new Response("nope", { status: 500 }))
      .mockResolvedValueOnce(
        new Response(
          new ReadableStream<Uint8Array>({
            pull() {
              return new Promise(() => undefined);
            }
          }),
          { status: 200 }
        )
      );
    const callbacks: Array<() => void> = [];
    const cancelCalls: string[] = [];
    let scheduleId = 0;
    const schedule = (cb: () => void) => {
      const id = `t${++scheduleId}`;
      callbacks.push(cb);
      return () => {
        cancelCalls.push(id);
      };
    };
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule
    });
    await flushAsyncTicks();
    expect(fetchMock).toHaveBeenCalledOnce();
    callbacks[0]?.();
    await flushAsyncTicks();
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(cancelCalls).toEqual([]);

    client.disconnect();
    expect(cancelCalls).toEqual([]);
  });

  it("disconnect aborts the in-flight fetch", async () => {
    const abortHandlers: AbortSignal[] = [];
    // Stream, der sich "infinite" verhält — pull liefert nichts
    // zurück und wartet auf manuellen Abort. Damit ist der
    // AbortController noch live, wenn `disconnect()` greift.
    const fetchMock = vi.fn(async (_url: unknown, init?: RequestInit) => {
      if (init?.signal) abortHandlers.push(init.signal);
      return new Response(
        new ReadableStream<Uint8Array>({
          pull() {
            return new Promise(() => undefined); // never resolves
          }
        }),
        { status: 200 }
      );
    });
    const client = startSseClient({
      url: "/api/stream-sessions/stream",
      token: "demo-token",
      onAppended: () => undefined,
      fetchFn: fetchMock,
      schedule: () => () => undefined
    });
    await flushAsyncTicks();
    client.disconnect();
    expect(abortHandlers[0]?.aborted).toBe(true);
  });
});

async function flushAsyncTicks(): Promise<void> {
  // Drei aufeinanderfolgende Microtask-Yields reichen, um die
  // promise-chains in startSseClient durchzuspülen.
  for (let i = 0; i < 3; i += 1) {
    await Promise.resolve();
  }
  await new Promise((r) => setTimeout(r, 0));
}
