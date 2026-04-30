import { describe, expect, it } from "vitest";
import { createTracker } from "../src/core/tracker";
import { EVENT_SCHEMA_VERSION, PLAYER_SDK_NAME, PLAYER_SDK_VERSION } from "../src/version";
import type { PlaybackEventBatch } from "../src/types/events";

class MemoryTransport {
  batches: PlaybackEventBatch[] = [];

  async send(batch: PlaybackEventBatch): Promise<void> {
    this.batches.push(batch);
  }
}

describe("SDK compatibility contract", () => {
  it("emits SDK metadata and schema version from exported constants", async () => {
    const transport = new MemoryTransport();
    const tracker = createTracker({
      endpoint: "http://localhost:8080/api/playback-events",
      token: "demo-token",
      projectId: "demo",
      sessionId: "session-compat",
      flushIntervalMs: 0,
      transport
    });

    tracker.track({ eventName: "manifest_loaded", timestamp: new Date("2026-04-30T00:00:00.000Z") });
    await tracker.flush();

    expect(transport.batches[0]?.schema_version).toBe(EVENT_SCHEMA_VERSION);
    expect(transport.batches[0]?.events[0]?.sdk).toEqual({
      name: PLAYER_SDK_NAME,
      version: PLAYER_SDK_VERSION
    });
  });
});
