import { describe, expect, it, vi } from "vitest";
import { createTracker } from "../src/core/tracker";
import type { PlaybackEventBatch } from "../src/types/events";
import type { Transport } from "../src/types/config";

class MemoryTransport {
  batches: PlaybackEventBatch[] = [];

  async send(batch: PlaybackEventBatch): Promise<void> {
    this.batches.push(batch);
  }
}

describe("MTracePlayerTracker", () => {
  it("batches events and emits the 1.0 wire format", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      sessionId: "session-1",
      batchSize: 2,
      flushIntervalMs: 0,
      transport
    });

    tracker.track({ eventName: "manifest_loaded", timestamp: new Date("2026-04-30T00:00:00.000Z") });
    expect(transport.batches).toHaveLength(0);

    tracker.track({
      eventName: "segment_loaded",
      timestamp: new Date("2026-04-30T00:00:01.000Z"),
      meta: { duration_ms: 120 }
    });
    await tracker.flush();

    expect(transport.batches).toEqual([
      {
        schema_version: "1.0",
        events: [
          {
            event_name: "manifest_loaded",
            project_id: "demo",
            session_id: "session-1",
            client_timestamp: "2026-04-30T00:00:00.000Z",
            sequence_number: 1,
            sdk: { name: "@npm9912/player-sdk", version: "0.5.0" }
          },
          {
            event_name: "segment_loaded",
            project_id: "demo",
            session_id: "session-1",
            client_timestamp: "2026-04-30T00:00:01.000Z",
            sequence_number: 2,
            sdk: { name: "@npm9912/player-sdk", version: "0.5.0" },
            meta: { duration_ms: 120 }
          }
        ]
      }
    ]);
  });

  it("honors sampleRate=0", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 1,
      flushIntervalMs: 0,
      sampleRate: 0,
      transport
    });

    tracker.track({ eventName: "manifest_loaded" });
    await tracker.flush();

    expect(transport.batches).toHaveLength(0);
  });

  it("uses event-level sampling without consuming sequence numbers for sampled-out events", async () => {
    const random = vi.spyOn(Math, "random").mockReturnValueOnce(0.9).mockReturnValueOnce(0.1);
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      sessionId: "session-1",
      flushIntervalMs: 0,
      sampleRate: 0.5,
      transport
    });

    tracker.track({ eventName: "manifest_loaded" });
    tracker.track({ eventName: "segment_loaded" });
    await tracker.flush();

    expect(transport.batches[0]?.events).toHaveLength(1);
    expect(transport.batches[0]?.events[0]).toMatchObject({
      event_name: "segment_loaded",
      sequence_number: 1
    });
    random.mockRestore();
  });

  it("flushes queued events on destroy", async () => {
    vi.useFakeTimers();
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 10,
      flushIntervalMs: 1000,
      transport
    });

    tracker.track({ eventName: "playback_error" });
    await tracker.destroy();

    expect(transport.batches).toHaveLength(1);
    expect(transport.batches[0]?.events.map((event) => event.event_name)).toEqual(["playback_error", "session_ended"]);
    vi.useRealTimers();
  });

  it("accepts an opt-in OTel-style transport through the stable transport port", async () => {
    const exportedBatches: PlaybackEventBatch[] = [];
    const otelLikeTransport: Transport = {
      async send(batch) {
        exportedBatches.push(batch);
      }
    };
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      sessionId: "session-otel",
      flushIntervalMs: 0,
      transport: otelLikeTransport
    });

    tracker.track({ eventName: "metrics_sampled", meta: { duration_ms: 12 } });
    await tracker.flush();

    expect(exportedBatches[0]?.events[0]).toMatchObject({
      event_name: "metrics_sampled",
      session_id: "session-otel",
      meta: { duration_ms: 12 }
    });
  });

  it("splits local queues into API-compatible batches", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 250,
      flushIntervalMs: 0,
      transport
    });

    for (let i = 0; i < 101; i += 1) {
      tracker.track({ eventName: "segment_loaded" });
    }
    await tracker.flush();

    expect(transport.batches.map((batch) => batch.events.length)).toEqual([100, 1]);
  });

  it("splits batches before the API payload body limit", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 10,
      flushIntervalMs: 0,
      transport
    });

    const payload = "x".repeat(170 * 1024);
    tracker.track({ eventName: "segment_loaded", meta: { payload } });
    tracker.track({ eventName: "segment_loaded", meta: { payload } });
    await tracker.flush();

    expect(transport.batches.map((batch) => batch.events.length)).toEqual([1, 1]);
    for (const sentBatch of transport.batches) {
      expect(batchBytes(sentBatch)).toBeLessThanOrEqual(256 * 1024);
    }
  });

  it("drops single events that cannot fit into an API payload", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 10,
      flushIntervalMs: 0,
      transport
    });

    tracker.track({ eventName: "segment_loaded", meta: { payload: "x".repeat(270 * 1024) } });
    await tracker.flush();

    expect(transport.batches).toHaveLength(0);
  });

  it("applies the local queue limit to sampled playback events", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 10,
      flushIntervalMs: 0,
      maxQueueEvents: 2,
      transport
    });

    tracker.track({ eventName: "manifest_loaded" });
    tracker.track({ eventName: "segment_loaded" });
    tracker.track({ eventName: "playback_started" });
    await tracker.flush();

    expect(transport.batches[0]?.events.map((event) => event.event_name)).toEqual(["segment_loaded", "playback_started"]);
  });

  it("emits session_ended only once on destroy", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 10,
      flushIntervalMs: 0,
      transport
    });

    await tracker.destroy();
    await tracker.destroy();

    expect(transport.batches).toHaveLength(1);
    expect(transport.batches[0]?.events.map((event) => event.event_name)).toEqual(["session_ended"]);
  });

  it("emits session_ended even when sampling drops playback events", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      batchSize: 10,
      flushIntervalMs: 0,
      sampleRate: 0,
      transport
    });

    tracker.track({ eventName: "manifest_loaded" });
    await tracker.destroy();

    expect(transport.batches[0]?.events.map((event) => event.event_name)).toEqual(["session_ended"]);
  });
});

function batchBytes(batch: PlaybackEventBatch): number {
  return new TextEncoder().encode(JSON.stringify(batch)).length;
}
