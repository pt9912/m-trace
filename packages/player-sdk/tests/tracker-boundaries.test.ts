import { describe, expect, it } from "vitest";
import { createTracker } from "../src/core/tracker";
import type { PlaybackEventBatch, SessionBoundary } from "../src/types/events";

class MemoryTransport {
  batches: PlaybackEventBatch[] = [];

  async send(batch: PlaybackEventBatch): Promise<void> {
    this.batches.push(batch);
  }
}

function newTracker(transport: MemoryTransport) {
  return createTracker({
    endpoint: "http://localhost:8080/api/playback-events",
    token: "demo-token",
    projectId: "demo",
    sessionId: "session-bnd",
    flushIntervalMs: 0,
    transport
  });
}

describe("session_boundaries[] send path", () => {
  it("flushes boundaries together with the first event batch", async () => {
    const transport = new MemoryTransport();
    const tracker = newTracker(transport);

    tracker.addBoundary({
      networkKind: "segment",
      adapter: "native_hls",
      reason: "native_hls_unavailable"
    });
    tracker.track({ eventName: "manifest_loaded" });
    await tracker.flush();

    expect(transport.batches).toHaveLength(1);
    const batch = transport.batches[0];
    expect(batch?.events).toHaveLength(1);
    expect(batch?.session_boundaries).toBeDefined();
    expect(batch?.session_boundaries).toHaveLength(1);
    const entry = batch?.session_boundaries?.[0] as SessionBoundary;
    expect(entry.kind).toBe("network_signal_absent");
    expect(entry.project_id).toBe("demo");
    expect(entry.session_id).toBe("session-bnd");
    expect(entry.network_kind).toBe("segment");
    expect(entry.adapter).toBe("native_hls");
    expect(entry.reason).toBe("native_hls_unavailable");
    expect(typeof entry.client_timestamp).toBe("string");
  });

  it("omits session_boundaries when no boundary was added", async () => {
    const transport = new MemoryTransport();
    const tracker = newTracker(transport);

    tracker.track({ eventName: "manifest_loaded" });
    await tracker.flush();

    expect(transport.batches[0]?.session_boundaries).toBeUndefined();
  });

  it("caps the boundary queue at 20 with drop-oldest", async () => {
    const transport = new MemoryTransport();
    const tracker = newTracker(transport);

    for (let i = 0; i < 25; i += 1) {
      tracker.addBoundary({
        networkKind: "segment",
        adapter: "hls.js",
        reason: `reason_${i}` // not enum-valid; SDK does not enforce — backend would 422
      });
    }
    tracker.track({ eventName: "manifest_loaded" });
    await tracker.flush();

    const sent = transport.batches[0]?.session_boundaries ?? [];
    expect(sent).toHaveLength(20);
    // drop-oldest: the first five (reason_0..reason_4) should be gone.
    const reasons = sent.map((b) => b.reason);
    expect(reasons).not.toContain("reason_0");
    expect(reasons).not.toContain("reason_4");
    expect(reasons).toContain("reason_5");
    expect(reasons).toContain("reason_24");
  });

  it("only sends boundaries on the first batch within a flush cycle", async () => {
    // Eigene Tracker-Instanz mit großem batchSize und ohne
    // Auto-Flush-Timer, damit der Test deterministisch alle Events
    // in einem flush()-Aufruf draint und die Body-Size-Grenze
    // mehrere Batches erzwingt.
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      sessionId: "session-bnd",
      batchSize: 100,
      flushIntervalMs: 0,
      transport
    });

    // Boundary first, dann large-meta-Events, sodass beim Flush das
    // Body-Size-Limit den Batch in ≥2 Stücke teilt — die Boundary
    // muss am ersten der zwei Batches hängen.
    tracker.addBoundary({
      networkKind: "manifest",
      adapter: "hls.js",
      reason: "cors_timing_blocked"
    });
    const largeMeta = { padding: "x".repeat(3000) };
    for (let i = 0; i < 100; i += 1) {
      tracker.track({ eventName: "metrics_sampled", meta: largeMeta });
    }
    await tracker.flush();

    expect(transport.batches.length).toBeGreaterThanOrEqual(2);
    const [first, ...rest] = transport.batches;
    expect(first?.session_boundaries).toBeDefined();
    expect(first?.session_boundaries).toHaveLength(1);
    for (const batch of rest) {
      expect(batch.session_boundaries).toBeUndefined();
    }
  });

  it("does not send boundaries when the queue is event-less", async () => {
    // Boundary-only flushes are not allowed by the backend; the
    // tracker holds boundaries in its queue until the next flush
    // that drains an event. flush() with empty event queue is a
    // no-op for boundaries too.
    const transport = new MemoryTransport();
    const tracker = newTracker(transport);

    tracker.addBoundary({
      networkKind: "segment",
      adapter: "native_hls",
      reason: "native_hls_unavailable"
    });
    await tracker.flush();

    expect(transport.batches).toHaveLength(0);

    // Once an event arrives, the boundary goes out with it.
    tracker.track({ eventName: "manifest_loaded" });
    await tracker.flush();
    expect(transport.batches).toHaveLength(1);
    expect(transport.batches[0]?.session_boundaries).toHaveLength(1);
  });

  it("respects an explicit boundary timestamp", async () => {
    const transport = new MemoryTransport();
    const tracker = newTracker(transport);
    const ts = new Date("2026-04-28T12:00:00.000Z");

    tracker.addBoundary({
      networkKind: "manifest",
      adapter: "hls.js",
      reason: "cors_timing_blocked",
      timestamp: ts
    });
    tracker.track({ eventName: "manifest_loaded" });
    await tracker.flush();

    expect(transport.batches[0]?.session_boundaries?.[0]?.client_timestamp).toBe(
      "2026-04-28T12:00:00.000Z"
    );
  });

  it("ignores addBoundary calls after destroy", async () => {
    const transport = new MemoryTransport();
    const tracker = newTracker(transport);

    await tracker.destroy();
    tracker.addBoundary({
      networkKind: "segment",
      adapter: "native_hls",
      reason: "native_hls_unavailable"
    });

    // destroy() flushes session_ended once; no boundary should ride
    // along.
    expect(transport.batches[0]?.session_boundaries).toBeUndefined();
  });
});
