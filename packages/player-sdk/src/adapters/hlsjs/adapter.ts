import type Hls from "hls.js";
import { SessionMetrics } from "../../core/session-metrics";
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
  const metrics = new SessionMetrics(startedAt);

  const onManifest = () => tracker.track({ eventName: "manifest_loaded" });
  const onFragmentLoaded = () => tracker.track({ eventName: "segment_loaded" });
  const onLevelSwitched = () => tracker.track({ eventName: "bitrate_switch" });
  const onError = () => tracker.track({ eventName: "playback_error" });
  const onWaiting = () => {
    if (metrics.startRebuffer(performance.now())) {
      tracker.track({ eventName: "rebuffer_started" });
    }
  };
  const onPlaying = () => {
    const rebuffer = metrics.endRebuffer(performance.now());
    if (rebuffer) {
      tracker.track({
        eventName: "rebuffer_ended",
        meta: {
          duration_ms: rebuffer.durationMs,
          rebuffer_count: rebuffer.rebufferCount,
          total_rebuffer_duration_ms: rebuffer.totalRebufferDurationMs
        }
      });
      return;
    }

    reportStartup();
  };
  const onLoadedData = () => reportStartup();

  const reportStartup = () => {
    const startupTimeMs = metrics.completeStartup(performance.now());
    if (startupTimeMs !== undefined) {
      tracker.track({ eventName: "playback_started" });
      tracker.track({ eventName: "startup_time_measured", meta: { duration_ms: startupTimeMs } });
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
