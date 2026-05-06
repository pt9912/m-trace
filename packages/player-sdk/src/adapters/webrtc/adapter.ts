import type { PlayerTracker } from "../../core/tracker";
import type { EventMeta } from "../../types/events";
import {
  WEBRTC_ERROR_CODE_META_KEY,
  normalizeWebRtcErrorCode,
  type WebRtcErrorCode
} from "./error-codes";
import { newPeerConnectionRunId, startSampling, type SamplingDeps } from "./sampling";

/**
 * WebRTC-/WHEP-Adapter für `@npm9912/player-sdk`. Implementiert den
 * additiven Adapter-Pfad zu {@link import("../hlsjs/adapter").attachHlsJs
 * attachHlsJs} und liefert Player-Events in den geteilten
 * `PlayerTracker`-Stream.
 *
 * Bezug: `docs/planning/in-progress/plan-0.8.0.md` §0.5
 * (Implementierungsleitplanken — WHEP als einziger Signalisierungsweg)
 * und §3 Tranche 2 DoD.
 *
 * Contract-Entscheidung Tranche 1 (rein SDK-intern) bleibt für Tranche 2
 * gültig: Wire-Format und API-Ingress sind unverändert, der Adapter
 * sendet `playback_started`/`playback_error`-Events analog zum hls.js-
 * Pfad. Die Fehlercode-Taxonomie aus {@link "./error-codes"} ist
 * SDK-intern; Tranche 3 zieht sie zusammen mit dem `webrtc.*`-Meta-
 * Namespace in den Wire-Vertrag.
 */

export interface WebRtcAdapter {
  /** Beendet PeerConnection, entfernt MediaTracks und gibt WHEP-Resource frei. Idempotent. */
  destroy(): void;
}

export interface WebRtcAdapterOptions {
  /** WHEP-Endpoint (`POST <whepUrl>` mit `application/sdp`-Body). Pflicht. */
  whepUrl: string;
  /** Optionale `RTCPeerConnection`-Konfiguration (z. B. STUN-Server). */
  peerConnectionConfig?: RTCConfiguration;
  /** Optionales `AbortSignal` zum Abbruch der WHEP-Signalisierung. */
  signal?: AbortSignal;
  /**
   * `getStats()`-Sampling-Intervall in Millisekunden für
   * `metrics_sampled`-Events (`plan-0.8.0` Tranche 3). Default 1000 ms;
   * `0` deaktiviert das Sampling vollständig.
   */
  samplingIntervalMs?: number;
}

interface AdapterDeps {
  /** Konstruktor für die Browser-`RTCPeerConnection`. Test-Override-fähig. */
  PeerConnection?: typeof RTCPeerConnection;
  /** `fetch`-Implementation. Test-Override-fähig. */
  fetch?: typeof fetch;
  /** Test-Overrides für `setInterval`/`clearInterval` (Sample-Loop). */
  sampling?: SamplingDeps;
  /** Test-Override für die `peer_connection_run_id`-Generierung. */
  newRunId?: () => string;
}

interface AdapterState {
  destroyed: boolean;
  connected: boolean;
  errored: boolean;
}

/**
 * Aktiviert einen WebRTC-/WHEP-Read-Pfad auf dem übergebenen
 * `<video>`-Element. Der Aufrufer ist für DOM-Mounting und -Cleanup
 * verantwortlich; `destroy()` schließt nur die WebRTC-Ressourcen.
 *
 * **Fehlerbehandlung**: jeder Fehler erzeugt genau ein
 * `playback_error`-Event mit dem reservierten Meta-Key
 * `webrtc.error_code` aus der Allowlist in
 * `./error-codes.ts`. Eine durch `destroy()` abgebrochene
 * Verbindung emittiert `webrtc_destroyed_before_connected` nur dann,
 * wenn der Handshake noch nicht fertig war.
 */
export function attachWebRtc(
  video: HTMLVideoElement,
  options: WebRtcAdapterOptions,
  tracker: PlayerTracker,
  deps: AdapterDeps = {}
): WebRtcAdapter {
  const PeerConnection = deps.PeerConnection ?? globalThis.RTCPeerConnection;
  const fetchImpl = deps.fetch ?? globalThis.fetch;
  if (typeof PeerConnection !== "function") {
    throw new Error("WebRTC adapter requires RTCPeerConnection — provide deps.PeerConnection or run in a browser");
  }
  if (typeof fetchImpl !== "function") {
    throw new Error("WebRTC adapter requires fetch — provide deps.fetch or run in a browser");
  }

  const pc = new PeerConnection(options.peerConnectionConfig ?? {});
  const tracks: MediaStreamTrack[] = [];
  const state: AdapterState = { destroyed: false, connected: false, errored: false };
  const localAbort = new AbortController();
  const composedSignal = composeSignals(options.signal, localAbort.signal);
  const runId = (deps.newRunId ?? newPeerConnectionRunId)();

  const reportError = (code: WebRtcErrorCode, detail?: string): void => {
    if (state.errored || state.destroyed) {
      return;
    }
    state.errored = true;
    tracker.track({ eventName: "playback_error", meta: errorMeta(code, runId, detail) });
  };

  pc.addEventListener("track", (event: RTCTrackEvent) => attachTrack(state, video, tracks, event));
  pc.addEventListener("connectionstatechange", () => handleConnectionStateChange(pc, tracker, state, runId, reportError));

  let stopSampling: (() => void) | undefined;
  const samplingIntervalMs = options.samplingIntervalMs ?? 1000;
  if (samplingIntervalMs > 0) {
    stopSampling = startSampling(pc, tracker, runId, {
      ...deps.sampling,
      intervalMs: samplingIntervalMs
    });
  }

  void runWhepHandshake(pc, options.whepUrl, fetchImpl, composedSignal).catch((err: unknown) => {
    if (state.destroyed) {
      return;
    }
    const code: WebRtcErrorCode =
      err instanceof WhepNoTracksError
        ? "webrtc_no_tracks"
        : err instanceof WhepSdpError
          ? "whep_sdp_invalid"
          : "whep_signaling_failed";
    reportError(code, err instanceof Error ? err.message : String(err));
  });

  return {
    destroy(): void {
      if (stopSampling) {
        stopSampling();
        stopSampling = undefined;
      }
      destroyAdapter(state, tracker, tracks, pc, video, localAbort, runId);
    }
  };
}

function errorMeta(code: WebRtcErrorCode, runId: string, detail?: string): EventMeta {
  const meta: EventMeta = {
    [WEBRTC_ERROR_CODE_META_KEY]: normalizeWebRtcErrorCode(code),
    "webrtc.peer_connection_run_id": runId
  };
  if (typeof detail === "string" && detail.length > 0) {
    meta["webrtc.error_detail"] = detail.slice(0, 256);
  }
  return meta;
}

function attachTrack(
  state: AdapterState,
  video: HTMLVideoElement,
  tracks: MediaStreamTrack[],
  event: RTCTrackEvent
): void {
  if (state.destroyed) {
    return;
  }
  const track = event.track;
  tracks.push(track);
  const stream = (event.streams && event.streams[0]) || newStream(track);
  try {
    if (video.srcObject !== stream) {
      video.srcObject = stream;
    }
  } catch {
    // jsdom / Test-Stub liefert ggf. kein srcObject. Tracks bleiben
    // im internen Array für destroy().
  }
}

function handleConnectionStateChange(
  pc: RTCPeerConnection,
  tracker: PlayerTracker,
  state: AdapterState,
  runId: string,
  reportError: (code: WebRtcErrorCode, detail?: string) => void
): void {
  const cs = pc.connectionState;
  if (cs === "connected" && !state.connected) {
    state.connected = true;
    tracker.track({
      eventName: "playback_started",
      meta: {
        "webrtc.connection_state": cs,
        "webrtc.peer_connection_run_id": runId
      }
    });
    return;
  }
  if (cs === "failed" || cs === "disconnected" || cs === "closed") {
    if (state.destroyed) {
      return;
    }
    reportError("peer_connection_failed", cs);
  }
}

function destroyAdapter(
  state: AdapterState,
  tracker: PlayerTracker,
  tracks: MediaStreamTrack[],
  pc: RTCPeerConnection,
  video: HTMLVideoElement,
  localAbort: AbortController,
  runId: string
): void {
  if (state.destroyed) {
    return;
  }
  state.destroyed = true;
  // destroy() vor Handshake-Abschluss: Adapter emittiert den
  // dokumentierten Code synchron.
  if (!state.connected && !state.errored) {
    state.errored = true;
    tracker.track({
      eventName: "playback_error",
      meta: errorMeta("webrtc_destroyed_before_connected", runId)
    });
  }
  localAbort.abort();
  for (const track of tracks) {
    track.stop();
  }
  pc.close();
  if (video.srcObject) {
    video.srcObject = null;
  }
}

class WhepSdpError extends Error {}
class WhepNoTracksError extends Error {}

async function runWhepHandshake(
  pc: RTCPeerConnection,
  whepUrl: string,
  fetchImpl: typeof fetch,
  signal: AbortSignal
): Promise<void> {
  pc.addTransceiver("video", { direction: "recvonly" });
  pc.addTransceiver("audio", { direction: "recvonly" });

  const offer = await pc.createOffer();
  await pc.setLocalDescription(offer);

  const offerSdp = offer.sdp ?? "";
  if (offerSdp.length === 0) {
    throw new WhepSdpError("createOffer() returned empty SDP");
  }

  const response = await fetchImpl(whepUrl, {
    method: "POST",
    headers: { "Content-Type": "application/sdp" },
    body: offerSdp,
    signal
  });
  if (!response.ok) {
    throw new Error(`WHEP signaling failed: HTTP ${String(response.status)}`);
  }
  const answerSdp = await response.text();
  if (!answerSdp.startsWith("v=")) {
    throw new WhepSdpError("WHEP response is not a valid SDP answer");
  }
  if (!answerSdp.includes("m=video") && !answerSdp.includes("m=audio")) {
    throw new WhepNoTracksError("WHEP answer has neither video nor audio media sections");
  }

  await pc.setRemoteDescription({ type: "answer", sdp: answerSdp });
}

function composeSignals(...signals: Array<AbortSignal | undefined>): AbortSignal {
  const controller = new AbortController();
  for (const s of signals) {
    if (!s) {
      continue;
    }
    if (s.aborted) {
      controller.abort();
      break;
    }
    s.addEventListener("abort", () => controller.abort(), { once: true });
  }
  return controller.signal;
}

function newStream(track: MediaStreamTrack): MediaStream {
  if (typeof MediaStream === "function") {
    return new MediaStream([track]);
  }
  // Fallback für Test-/Headless-Umgebungen ohne globalen MediaStream.
  return { getTracks: () => [track] } as unknown as MediaStream;
}
