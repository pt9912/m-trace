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

  async flush(): Promise<void> {
    return undefined;
  }

  async destroy(): Promise<void> {
    return undefined;
  }
}

describe("attachHlsJs", () => {
  it("maps hls.js manifest, fragment, level and error events", () => {
    const hls = new FakeHls();
    const video = new FakeVideo();
    const tracker = new RecordingTracker();

    attachHlsJs(video as unknown as HTMLVideoElement, hls as never, tracker);

    hls.emit("hlsManifestLoaded");
    hls.emit("hlsFragLoaded");
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
    const hls = new FakeHls();
    const video = new FakeVideo();
    const tracker = new RecordingTracker();

    attachHlsJs(video as unknown as HTMLVideoElement, hls as never, tracker);

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

  it("removes all listeners on destroy", () => {
    const hls = new FakeHls();
    const video = new FakeVideo();
    const tracker = new RecordingTracker();
    const adapter = attachHlsJs(video as unknown as HTMLVideoElement, hls as never, tracker);

    adapter.destroy();
    hls.emit("hlsManifestLoaded");
    hls.emit("hlsFragLoaded");
    hls.emit("hlsLevelSwitched");
    hls.emit("hlsError");
    video.emit("loadeddata");
    video.emit("waiting");
    video.emit("playing");

    expect(tracker.events).toEqual([]);
  });
});
