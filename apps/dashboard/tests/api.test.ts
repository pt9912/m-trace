import { afterEach, describe, expect, it, vi } from "vitest";
import { formatTime, getHealth, getSession, isErrorEvent, listSessions } from "../src/lib/api";
import type { PlaybackEvent } from "../src/lib/api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("dashboard API client", () => {
  it("requests sessions with an explicit limit", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ sessions: [] }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(listSessions(25)).resolves.toEqual({ sessions: [] });

    expect(fetchMock).toHaveBeenCalledWith("/api/stream-sessions?limit=25", {
      headers: { Accept: "application/json" },
      cache: "no-store"
    });
  });

  it("encodes session ids for detail requests", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        session: {
          session_id: "session/1",
          project_id: "demo",
          state: "active",
          started_at: "2026-04-30T00:00:00.000Z",
          last_event_at: "2026-04-30T00:00:00.000Z",
          event_count: 0
        },
        events: []
      })
    );
    vi.stubGlobal("fetch", fetchMock);

    await getSession("session/1", 10);

    expect(fetchMock).toHaveBeenCalledWith("/api/stream-sessions/session%2F1?events_limit=10", {
      headers: { Accept: "application/json" },
      cache: "no-store"
    });
  });

  it("appends events_cursor when paginating session detail", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ session: {}, events: [] }));
    vi.stubGlobal("fetch", fetchMock);
    await getSession("session-1", 50, "opaque-cursor");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/stream-sessions/session-1?events_limit=50&events_cursor=opaque-cursor",
      { headers: { Accept: "application/json" }, cache: "no-store" }
    );
  });

  it("maps health responses and unreachable APIs", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("ok", { status: 200 })));
    await expect(getHealth()).resolves.toEqual({ ok: true, status: 200, text: "ok" });

    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      })
    );
    await expect(getHealth()).resolves.toEqual({ ok: false, status: 0, text: "offline" });
  });

  it("rejects JSON requests with non-2xx responses", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response("nope", { status: 500 })));

    await expect(listSessions()).rejects.toThrow("/api/stream-sessions?limit=100 returned 500");
  });

  it("records network errors in the read-error store (§5 H3)", async () => {
    const { lastReadError, clearLastReadError } = await import("../src/lib/status");
    clearLastReadError();
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      })
    );
    await expect(listSessions()).rejects.toThrow("offline");
    const { get } = await import("svelte/store");
    const rec = get(lastReadError);
    expect(rec?.message).toBe("offline");
    expect(rec?.source).toContain("/api/stream-sessions");
    clearLastReadError();
  });

  it("records non-Error throws via getJSON catch path", async () => {
    const { lastReadError, clearLastReadError } = await import("../src/lib/status");
    clearLastReadError();
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        // eslint-disable-next-line @typescript-eslint/only-throw-error
        throw "string-instead-of-error";
      })
    );
    await expect(listSessions()).rejects.toBe("string-instead-of-error");
    const { get } = await import("svelte/store");
    expect(get(lastReadError)?.message).toBe("string-instead-of-error");
    clearLastReadError();
  });

  it("classifies error and warning events", () => {
    expect(isErrorEvent(event("playback_error"))).toBe(true);
    expect(isErrorEvent(event("buffer_warning"))).toBe(true);
    expect(isErrorEvent(event("segment_loaded"))).toBe(false);
  });

  it("formats absent timestamps as n/a", () => {
    expect(formatTime(undefined)).toBe("n/a");
  });

  it("formats present timestamps", () => {
    expect(formatTime("2026-04-30T12:34:56.000Z")).not.toBe("n/a");
  });
});

function jsonResponse(payload: unknown): Response {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" }
  });
}

function event(eventName: string): PlaybackEvent {
  return {
    event_name: eventName,
    project_id: "demo",
    session_id: "session-1",
    client_timestamp: "2026-04-30T00:00:00.000Z",
    server_received_at: "2026-04-30T00:00:00.000Z",
    ingest_sequence: 1,
    sdk: { name: "@npm9912/player-sdk", version: "0.2.0" }
  };
}
