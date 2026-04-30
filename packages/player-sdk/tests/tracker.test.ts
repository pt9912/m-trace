import { describe, expect, it, vi } from "vitest";
import { createTracker } from "../src/core/tracker";
import type { PlaybackEventBatch } from "../src/types/events";

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

    tracker.track({ eventName: "segment_loaded", timestamp: new Date("2026-04-30T00:00:01.000Z") });
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
            sdk: { name: "@m-trace/player-sdk", version: "0.1.1-dev" }
          },
          {
            event_name: "segment_loaded",
            project_id: "demo",
            session_id: "session-1",
            client_timestamp: "2026-04-30T00:00:01.000Z",
            sequence_number: 2,
            sdk: { name: "@m-trace/player-sdk", version: "0.1.1-dev" }
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
    vi.useRealTimers();
  });
});
