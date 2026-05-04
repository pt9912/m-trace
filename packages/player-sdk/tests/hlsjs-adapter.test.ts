import { describe, expect, it } from "vitest";
import { attachHlsJs } from "../src/adapters/hlsjs/adapter";
import type { PlayerTracker } from "../src/core/tracker";
import type { EventDraft } from "../src/types/events";

class FakeHls {
  private readonly listeners = new Map<string, Set<(...args: unknown[]) => void>>();

  on(event: string, callback: (...args: unknown[]) => void): void {
    this.listeners.set(event, this.listeners.get(event) ?? new Set());
    this.listeners.get(event)?.add(callback);
  }

  off(event: string, callback: (...args: unknown[]) => void): void {
    this.listeners.get(event)?.delete(callback);
  }

  emit(event: string, ...args: unknown[]): void {
    for (const callback of this.listeners.get(event) ?? []) {
      callback(...args);
    }
  }
}

class FakeVideo {
  private readonly listeners = new Map<string, Set<EventListener>>();

  addEventListener(event: string, callback: EventListener): void {
    this.listeners.set(event, this.listeners.get(event) ?? new Set());
    this.listeners.get(event)?.add(callback);
  }

  removeEventListener(event: string, callback: EventListener): void {
    this.listeners.get(event)?.delete(callback);
  }

  emit(event: string): void {
    for (const callback of this.listeners.get(event) ?? []) {
      callback(new Event(event));
    }
  }
}

class RecordingTracker implements PlayerTracker {
  readonly sessionId = "session-1";
  readonly events: EventDraft[] = [];

  track(event: EventDraft): void {
    this.events.push(event);
  }

  addBoundary(): void {
    // hls.js-Adapter ruft addBoundary nicht auf; no-op reicht.
  }

  async flush(): Promise<void> {
    return undefined;
  }

  async destroy(): Promise<void> {
    return undefined;
  }
}

function setup(): { hls: FakeHls; video: FakeVideo; tracker: RecordingTracker } {
  const hls = new FakeHls();
  const video = new FakeVideo();
  const tracker = new RecordingTracker();
  attachHlsJs(video as unknown as HTMLVideoElement, hls as never, tracker);
  return { hls, video, tracker };
}

describe("attachHlsJs", () => {
  it("maps hls.js manifest, fragment, level and error events", () => {
    const { hls, tracker } = setup();

    hls.emit("hlsManifestLoaded");
    hls.emit("hlsFragLoaded", { frag: { sn: 0, level: 0, type: "main", cc: 0 } });
    hls.emit("hlsLevelSwitched");
    hls.emit("hlsError");

    expect(tracker.events.map((event) => event.eventName)).toEqual([
      "manifest_loaded",
      "segment_loaded",
      "bitrate_switch",
      "playback_error"
    ]);
  });

  it("maps startup and rebuffer video events", () => {
    const { video, tracker } = setup();

    video.emit("loadeddata");
    video.emit("waiting");
    video.emit("playing");

    expect(tracker.events.map((event) => event.eventName)).toEqual([
      "playback_started",
      "startup_time_measured",
      "rebuffer_started",
      "rebuffer_ended"
    ]);
    expect(tracker.events.at(-1)?.meta).toMatchObject({
      rebuffer_count: 1
    });
  });

  it("does not start duplicate rebuffer spans", () => {
    const { video, tracker } = setup();

    video.emit("waiting");
    video.emit("waiting");
    video.emit("playing");

    expect(tracker.events.map((event) => event.eventName)).toEqual(["rebuffer_started", "rebuffer_ended"]);
  });

  it("maps startup from the first playing event", () => {
    const { video, tracker } = setup();

    video.emit("playing");
    video.emit("playing");

    expect(tracker.events.map((event) => event.eventName)).toEqual([
      "playback_started",
      "startup_time_measured"
    ]);
  });

  it("removes all listeners on destroy", () => {
    const hls = new FakeHls();
    const video = new FakeVideo();
    const tracker = new RecordingTracker();
    const adapter = attachHlsJs(video as unknown as HTMLVideoElement, hls as never, tracker);

    adapter.destroy();
    hls.emit("hlsManifestLoaded");
    hls.emit("hlsLevelLoaded");
    hls.emit("hlsFragLoaded", { frag: { sn: 0 } });
    hls.emit("hlsLevelSwitched");
    hls.emit("hlsError");
    video.emit("loadeddata");
    video.emit("waiting");
    video.emit("playing");

    expect(tracker.events).toEqual([]);
  });

  // --- §4.6 mapping tests ----------------------------------------

  it("attaches network.kind/network.detail_status meta to manifest_loaded", () => {
    const { hls, tracker } = setup();

    hls.emit("hlsManifestLoaded", { url: "https://cdn.example.test/master.m3u8" });

    expect(tracker.events).toHaveLength(1);
    expect(tracker.events[0]?.meta).toMatchObject({
      "network.kind": "manifest",
      "network.detail_status": "available",
      "network.redacted_url": "https://cdn.example.test/master.m3u8"
    });
  });

  it("emits a fresh manifest_loaded for each LEVEL_LOADED (live reloads)", () => {
    const { hls, tracker } = setup();

    hls.emit("hlsLevelLoaded", { level: 0, details: { url: "https://cdn.example.test/level0.m3u8" } });
    hls.emit("hlsLevelLoaded", { level: 0, details: { url: "https://cdn.example.test/level0.m3u8" } });
    hls.emit("hlsLevelLoaded", { level: 1, details: { url: "https://cdn.example.test/level1.m3u8" } });

    const manifestEvents = tracker.events.filter((e) => e.eventName === "manifest_loaded");
    expect(manifestEvents).toHaveLength(3);
    expect(manifestEvents.every((e) => e.meta?.["network.kind"] === "manifest")).toBe(true);
  });

  it("dedups segment_loaded for fragment retries (same sn/cc/type/level)", () => {
    const { hls, tracker } = setup();
    const frag = { frag: { sn: 42, cc: 0, type: "main", level: 0, url: "https://cdn.example.test/seg/42.ts" } };

    // hls.js emits FRAG_LOADED only on success; nested players or
    // doubled listeners may still fire it twice for the same frag.
    hls.emit("hlsFragLoaded", frag);
    hls.emit("hlsFragLoaded", frag);

    expect(tracker.events.filter((e) => e.eventName === "segment_loaded")).toHaveLength(1);
  });

  it("does not dedup distinct fragments", () => {
    const { hls, tracker } = setup();

    hls.emit("hlsFragLoaded", { frag: { sn: 0, cc: 0, type: "main", level: 0 } });
    hls.emit("hlsFragLoaded", { frag: { sn: 1, cc: 0, type: "main", level: 0 } });
    hls.emit("hlsFragLoaded", { frag: { sn: 0, cc: 1, type: "main", level: 0 } });
    hls.emit("hlsFragLoaded", { frag: { sn: 0, cc: 0, type: "audio", level: 0 } });
    hls.emit("hlsFragLoaded", { frag: { sn: 0, cc: 0, type: "main", level: 1 } });

    expect(tracker.events.filter((e) => e.eventName === "segment_loaded")).toHaveLength(5);
  });

  it("marks init segments with is_init=true", () => {
    const { hls, tracker } = setup();

    hls.emit("hlsFragLoaded", { frag: { sn: "initSegment", cc: 0, type: "main", level: 0 } });
    hls.emit("hlsFragLoaded", { frag: { sn: 0, cc: 0, type: "main", level: 0 } });

    const segments = tracker.events.filter((e) => e.eventName === "segment_loaded");
    expect(segments).toHaveLength(2);
    expect(segments[0]?.meta?.is_init).toBe(true);
    expect(segments[1]?.meta?.is_init).toBeUndefined();
  });

  it("redacts signed segment URLs in network.redacted_url", () => {
    const { hls, tracker } = setup();
    const signedUrl =
      "https://cdn.example.test/v1/" +
      "a".repeat(32) +
      "/seg/0001.ts?token=abc&sig=xyz#frag";

    hls.emit("hlsFragLoaded", {
      frag: { sn: 0, cc: 0, type: "main", level: 0, url: signedUrl }
    });

    const meta = tracker.events.find((e) => e.eventName === "segment_loaded")?.meta;
    expect(meta?.["network.redacted_url"]).toBe(
      "https://cdn.example.test/v1/:redacted/seg/0001.ts"
    );
  });

  it("falls back to the redacted sentinel for unparsable URLs", () => {
    const { hls, tracker } = setup();
    hls.emit("hlsFragLoaded", { frag: { sn: 0, cc: 0, type: "main", level: 0, url: "not-a-url" } });
    const meta = tracker.events.find((e) => e.eventName === "segment_loaded")?.meta;
    expect(meta?.["network.redacted_url"]).toBe(":redacted");
  });

  it("emits segment_loaded even when fragment payload is missing", () => {
    // hls.js version drift safety: if the FRAG_LOADED payload comes
    // without a fragment object, the SDK should still surface the
    // event with the bare network.kind/detail_status meta.
    const { hls, tracker } = setup();
    hls.emit("hlsFragLoaded", {});
    const meta = tracker.events.find((e) => e.eventName === "segment_loaded")?.meta;
    expect(meta?.["network.kind"]).toBe("segment");
    expect(meta?.["network.redacted_url"]).toBeUndefined();
  });
});
