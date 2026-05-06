import { afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import {
  attachWebRtc,
  type WebRtcAdapter,
  type WebRtcAdapterOptions
} from "../src/adapters/webrtc/adapter";
import {
  WEBRTC_ERROR_CODES,
  WEBRTC_ERROR_CODE_META_KEY,
  isWebRtcErrorCode,
  normalizeWebRtcErrorCode
} from "../src/adapters/webrtc/error-codes";
import type { PlayerTracker } from "../src/core/tracker";
import type { EventDraft } from "../src/types/events";

// plan-0.8.0 Tranchen 1+2 — Public-API-Vertrag (Tranche 1) und
// Verhaltens-Tests (Tranche 2). Tests laufen ohne echte Browser-
// Signalisierung; `RTCPeerConnection` und `fetch` werden via
// `deps`-Override gemockt.

class StubTracker implements PlayerTracker {
  readonly sessionId = "test-session";
  events: EventDraft[] = [];
  track = vi.fn((event: EventDraft) => {
    this.events.push(event);
  });
  addBoundary = vi.fn();
  flush = vi.fn().mockResolvedValue(undefined);
  destroy = vi.fn().mockResolvedValue(undefined);
}

interface FakePcOptions {
  /** Soll `connectionstatechange` mit `connected` ausgelöst werden? */
  emitConnected?: boolean;
  /** Wenn gesetzt, wird vor `track`-Event ein einzelner Track simuliert. */
  emitTrack?: boolean;
}

function makeFakePeerConnection(opts: FakePcOptions) {
  const listeners = new Map<string, Set<(event: unknown) => void>>();
  const transceivers: unknown[] = [];
  const fakePc = {
    connectionState: "new" as RTCPeerConnectionState,
    addEventListener(event: string, cb: (e: unknown) => void): void {
      if (!listeners.has(event)) {
        listeners.set(event, new Set());
      }
      listeners.get(event)?.add(cb);
    },
    removeEventListener(): void {
      // ignore
    },
    addTransceiver(kind: string, init: { direction: string }): unknown {
      const t = { kind, direction: init.direction };
      transceivers.push(t);
      return t;
    },
    getTransceivers(): unknown[] {
      return transceivers;
    },
    async createOffer(): Promise<RTCSessionDescriptionInit> {
      return { type: "offer", sdp: "v=0\no=- 1 1 IN IP4 0.0.0.0\ns=-\n" };
    },
    async setLocalDescription(): Promise<void> {
      // no-op
    },
    async setRemoteDescription(): Promise<void> {
      // no-op
    },
    close(): void {
      fakePc.connectionState = "closed";
    },
    emit(event: string, payload: unknown): void {
      for (const cb of listeners.get(event) ?? []) {
        cb(payload);
      }
    }
  };
  if (opts.emitTrack) {
    setTimeout(() => {
      const track = { stop: vi.fn() } as unknown as MediaStreamTrack;
      fakePc.emit("track", { track, streams: [] });
    }, 0);
  }
  if (opts.emitConnected) {
    setTimeout(() => {
      fakePc.connectionState = "connected";
      fakePc.emit("connectionstatechange", undefined);
    }, 1);
  }
  return fakePc;
}

function makeFakeFetch(answer: { ok?: boolean; status?: number; sdp?: string }): typeof fetch {
  return vi.fn(async () => {
    return {
      ok: answer.ok ?? true,
      status: answer.status ?? 200,
      async text() {
        // Default-Answer enthält m=video + m=audio, damit happy-path-
        // Tests den Track-Pfad durchlaufen; spezielle Tests übergeben
        // sdp explizit.
        return (
          answer.sdp ??
          "v=0\no=- 1 1 IN IP4 0.0.0.0\ns=-\nm=video 9 UDP/TLS/RTP/SAVPF 96\nm=audio 9 UDP/TLS/RTP/SAVPF 111\n"
        );
      }
    };
  }) as unknown as typeof fetch;
}

const baseOptions: WebRtcAdapterOptions = {
  whepUrl: "http://localhost:8892/webrtc-test/whep"
};

const fakeVideo = {} as HTMLVideoElement;

beforeAll(() => {
  // Test-Polyfill für `MediaStream` — vitest läuft im Node-Modus,
  // dort fehlt das globale `MediaStream`. Adapter nutzt es im
  // `track`-Listener, wenn das Event keine `streams[0]` liefert.
  if (typeof (globalThis as { MediaStream?: unknown }).MediaStream !== "function") {
    (globalThis as { MediaStream?: unknown }).MediaStream = class {
      private readonly tracks: unknown[];
      constructor(initial?: unknown[]) {
        this.tracks = initial ?? [];
      }
      getTracks(): unknown[] {
        return this.tracks;
      }
    };
  }
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("attachWebRtc — Public-API-Vertrag (Tranche 1)", () => {
  it("ist als Funktion exportiert", () => {
    expect(typeof attachWebRtc).toBe("function");
  });

  it("WebRtcAdapter exposed eine destroy()-Surface", () => {
    const fake: WebRtcAdapter = {
      destroy: () => undefined
    };
    expect(typeof fake.destroy).toBe("function");
  });

  it("akzeptiert die optionale RTCConfiguration und das AbortSignal als Type-Vertrag", () => {
    const optionsAll: WebRtcAdapterOptions = {
      whepUrl: "http://localhost:8892/webrtc-test/whep",
      peerConnectionConfig: { iceServers: [] },
      signal: new AbortController().signal
    };
    expect(optionsAll.whepUrl).toBe(baseOptions.whepUrl);
  });

  it("wirft, wenn weder Browser-API noch deps-Override RTCPeerConnection bereitstellen", () => {
    const tracker = new StubTracker();
    expect(() =>
      attachWebRtc(fakeVideo, baseOptions, tracker, {
        PeerConnection: undefined,
        fetch: makeFakeFetch({})
      })
    ).toThrow(/RTCPeerConnection/);
  });
});

describe("attachWebRtc — Happy Path (Tranche 2)", () => {
  it("emittiert playback_started, sobald connectionState=connected ist", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({ emitConnected: true, emitTrack: true });
    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });

    await new Promise((r) => setTimeout(r, 5));

    const started = tracker.events.find((e) => e.eventName === "playback_started");
    expect(started).toBeDefined();
    expect(started?.meta?.["webrtc.connection_state"]).toBe("connected");
    adapter.destroy();
  });

  it("addTransceiver wird genau für video+audio recvonly aufgerufen", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    await new Promise((r) => setTimeout(r, 5));
    expect(fakePc.getTransceivers()).toHaveLength(2);
  });
});

describe("attachWebRtc — Fehler-Pfade (Tranche 2)", () => {
  it("HTTP-Fehler beim WHEP-POST → playback_error mit whep_signaling_failed", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({ ok: false, status: 500 })
    });
    await new Promise((r) => setTimeout(r, 5));

    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("whep_signaling_failed");
  });

  it("Antwort ist keine SDP → whep_sdp_invalid", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({ sdp: "<html>not SDP</html>" })
    });
    await new Promise((r) => setTimeout(r, 5));

    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("whep_sdp_invalid");
  });

  it("connectionstatechange=failed → playback_error mit peer_connection_failed", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    await new Promise((r) => setTimeout(r, 5));
    fakePc.connectionState = "failed";
    fakePc.emit("connectionstatechange", undefined);

    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("peer_connection_failed");
  });

  it("destroy() vor connected → webrtc_destroyed_before_connected", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    // Fetch hängt, damit der Handshake erst nach destroy() abbricht.
    const hangingFetch = vi.fn(
      async () =>
        new Promise<Response>(() => {
          /* nie auflösen */
        })
    ) as unknown as typeof fetch;

    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: hangingFetch
    });
    await new Promise((r) => setTimeout(r, 1));
    adapter.destroy();
    await new Promise((r) => setTimeout(r, 5));

    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("webrtc_destroyed_before_connected");
  });

  it("destroy() ist idempotent", () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    adapter.destroy();
    expect(() => adapter.destroy()).not.toThrow();
  });

  it("destroy() stoppt alle bisher mounted MediaTracks", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({ emitTrack: true });
    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    await new Promise((r) => setTimeout(r, 5));
    // emit zwei zusätzliche Tracks, damit der destroy-Loop mehr als
    // einen track.stop() aufruft.
    const t1Stop = vi.fn();
    const t2Stop = vi.fn();
    const t1 = { stop: t1Stop } as unknown as MediaStreamTrack;
    const t2 = { stop: t2Stop } as unknown as MediaStreamTrack;
    fakePc.emit("track", { track: t1, streams: [{ getTracks: () => [t1] }] });
    fakePc.emit("track", { track: t2, streams: [] });
    adapter.destroy();
    expect(t1Stop).toHaveBeenCalled();
    expect(t2Stop).toHaveBeenCalled();
  });

  it("ignoriert track-Events nach destroy()", () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    adapter.destroy();
    const t = { stop: vi.fn() } as unknown as MediaStreamTrack;
    expect(() => fakePc.emit("track", { track: t, streams: [] })).not.toThrow();
  });

  it("ignoriert connectionstatechange nach destroy()", () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    adapter.destroy();
    const before = tracker.events.length;
    fakePc.connectionState = "failed";
    fakePc.emit("connectionstatechange", undefined);
    expect(tracker.events.length).toBe(before);
  });

  it("SDP Answer ohne m=video/m=audio → webrtc_no_tracks", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({ sdp: "v=0\no=- 1 1 IN IP4 0.0.0.0\ns=-\n" })
    });
    await new Promise((r) => setTimeout(r, 5));
    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("webrtc_no_tracks");
  });

  it("createOffer liefert leere SDP → whep_sdp_invalid", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    fakePc.createOffer = async () => ({ type: "offer", sdp: "" });
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    await new Promise((r) => setTimeout(r, 5));
    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("whep_sdp_invalid");
  });

  it("composeSignals: ein bereits abgebrochenes options.signal abortet sofort", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const ac = new AbortController();
    ac.abort();
    attachWebRtc(
      fakeVideo,
      { whepUrl: baseOptions.whepUrl, signal: ac.signal },
      tracker,
      {
        PeerConnection: function () {
          return fakePc;
        } as unknown as typeof RTCPeerConnection,
        // fetch sollte mit AbortError ablehnen, sobald signal aborted ist.
        fetch: vi.fn(async (_url: unknown, init?: RequestInit) => {
          if (init?.signal?.aborted) {
            throw new DOMException("Aborted", "AbortError");
          }
          return { ok: true, status: 200, async text() { return "v=0\n"; } };
        }) as unknown as typeof fetch
      }
    );
    await new Promise((r) => setTimeout(r, 5));
    const err = tracker.events.find((e) => e.eventName === "playback_error");
    expect(err?.meta?.[WEBRTC_ERROR_CODE_META_KEY]).toBe("whep_signaling_failed");
  });
});

describe("WebRTC-Fehlercode-Taxonomie (Tranche 2)", () => {
  it("normalize unbekannte Strings auf peer_connection_failed", () => {
    expect(normalizeWebRtcErrorCode("not-a-known-code")).toBe("peer_connection_failed");
    expect(normalizeWebRtcErrorCode("")).toBe("peer_connection_failed");
    expect(normalizeWebRtcErrorCode(undefined)).toBe("peer_connection_failed");
    expect(normalizeWebRtcErrorCode(null)).toBe("peer_connection_failed");
    expect(normalizeWebRtcErrorCode(42)).toBe("peer_connection_failed");
  });

  it("normalize alle gültigen Codes liefert sie unverändert zurück", () => {
    for (const code of WEBRTC_ERROR_CODES) {
      expect(normalizeWebRtcErrorCode(code)).toBe(code);
    }
  });

  it("isWebRtcErrorCode unterscheidet gültig und ungültig", () => {
    for (const code of WEBRTC_ERROR_CODES) {
      expect(isWebRtcErrorCode(code)).toBe(true);
    }
    expect(isWebRtcErrorCode("foo")).toBe(false);
    expect(isWebRtcErrorCode(undefined)).toBe(false);
  });
});
