import type Hls from "hls.js";
import type { PlayerTracker } from "../../core/tracker";

interface HlsEventEmitter {
  on(event: string, callback: (...args: unknown[]) => void): void;
  off(event: string, callback: (...args: unknown[]) => void): void;
}

export interface HlsJsAdapter {
  destroy(): void;
}

export function attachHlsJs(video: HTMLVideoElement, hls: Hls, tracker: PlayerTracker): HlsJsAdapter {
  const emitter = hls as unknown as HlsEventEmitter;
  const startedAt = performance.now();
  let rebufferStartedAt: number | undefined;

  const onManifest = () => tracker.track({ eventName: "manifest_loaded" });
  const onFragmentLoaded = () => tracker.track({ eventName: "segment_loaded" });
  const onLevelSwitched = () => tracker.track({ eventName: "bitrate_switch" });
  const onError = () => tracker.track({ eventName: "playback_error" });
  const onWaiting = () => {
    rebufferStartedAt = performance.now();
    tracker.track({ eventName: "rebuffer_started" });
  };
  const onPlaying = () => {
    if (rebufferStartedAt !== undefined) {
      rebufferStartedAt = undefined;
      tracker.track({ eventName: "rebuffer_ended" });
      return;
    }
    tracker.track({ eventName: "startup_completed" });
  };
  const onLoadedData = () => {
    if (performance.now() >= startedAt) {
      tracker.track({ eventName: "startup_completed" });
    }
  };

  emitter.on("hlsManifestLoaded", onManifest);
  emitter.on("hlsFragLoaded", onFragmentLoaded);
  emitter.on("hlsLevelSwitched", onLevelSwitched);
  emitter.on("hlsError", onError);
  video.addEventListener("waiting", onWaiting);
  video.addEventListener("playing", onPlaying);
  video.addEventListener("loadeddata", onLoadedData, { once: true });

  return {
    destroy() {
      emitter.off("hlsManifestLoaded", onManifest);
      emitter.off("hlsFragLoaded", onFragmentLoaded);
      emitter.off("hlsLevelSwitched", onLevelSwitched);
      emitter.off("hlsError", onError);
      video.removeEventListener("waiting", onWaiting);
      video.removeEventListener("playing", onPlaying);
      video.removeEventListener("loadeddata", onLoadedData);
    }
  };
}
