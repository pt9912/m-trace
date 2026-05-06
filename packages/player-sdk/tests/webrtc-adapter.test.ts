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
import { collectAggregate } from "../src/adapters/webrtc/sampling";
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
    iceConnectionState: "new" as RTCIceConnectionState,
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
      headers: {
        get() {
          return null;
        }
      },
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
  whepUrl: "http://localhost:8892/webrtc-test/whep",
  // Tests deaktivieren das Sampling-Intervall standardmäßig — die
  // Sampling-Tests setzen es explizit über deps.sampling.setInterval.
  samplingIntervalMs: 0
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

  it("destroy() gibt eine WHEP-Resource per DELETE frei, wenn der Server Location liefert", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const fetchSpy = vi.fn(async (_url: string, init?: RequestInit) => {
      if (init?.method === "DELETE") {
        return { ok: true, status: 200, headers: { get: () => null }, async text() { return ""; } };
      }
      return {
        ok: true,
        status: 201,
        headers: {
          get(name: string) {
            return name.toLowerCase() === "location" ? "/webrtc-test/whep/session-a" : null;
          }
        },
        async text() {
          return "v=0\no=- 1 1 IN IP4 0.0.0.0\ns=-\nm=video 9 UDP/TLS/RTP/SAVPF 96\n";
        }
      };
    }) as unknown as typeof fetch;

    const adapter = attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: fetchSpy
    });
    await new Promise((r) => setTimeout(r, 5));
    adapter.destroy();
    await new Promise((r) => setTimeout(r, 5));

    expect(fetchSpy).toHaveBeenCalledWith("http://localhost:8892/webrtc-test/whep/session-a", { method: "DELETE" });
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

describe("collectAggregate (Tranche 3)", () => {
  function makeReport(entries: Record<string, unknown>[]): RTCStatsReport {
    const map = new Map<string, unknown>(entries.map((e, i) => [String(i), e]));
    return map;
  }

  const happyEntries: Record<string, unknown>[] = [
    { type: "transport", dtlsState: "connected" },
    { type: "candidate-pair", state: "succeeded" },
    { type: "candidate-pair", state: "succeeded", nominated: true },
    { type: "inbound-rtp", packetsLost: 5, bytesReceived: 12345 },
    { type: "outbound-rtp", bytesSent: 6789 }
  ];

  it("aggregiert dtls/ice/connection plus Counter-Felder", () => {
    const out = collectAggregate(makeReport(happyEntries), "connected");
    expect(out).not.toBeNull();
    expect(out?.connectionState).toBe("connected");
    expect(out?.dtlsState).toBe("connected");
    expect(out?.iceState).toBe("connected");
    expect(out?.packetsLost).toBe(5);
    expect(out?.bytesReceived).toBe(12345);
    expect(out?.bytesSent).toBe(6789);
  });

  it("liefert null, wenn ein Muss-Feld fehlt (dtls)", () => {
    const entries = happyEntries.filter((e) => e.type !== "transport");
    expect(collectAggregate(makeReport(entries), "connected")).toBeNull();
  });

  it("liefert null, wenn weder inbound noch outbound vorhanden sind", () => {
    const entries = happyEntries.filter((e) => e.type !== "inbound-rtp" && e.type !== "outbound-rtp");
    expect(collectAggregate(makeReport(entries), "connected")).toBeNull();
  });

  it("liefert null bei ungültigem connectionState (kein unknown-Surrogat)", () => {
    expect(collectAggregate(makeReport(happyEntries), "garbage-state")).toBeNull();
  });

  it("bevorzugt pc.iceConnectionState gegenüber Candidate-Pair-Fallback", () => {
    const entries = [
      { type: "transport", dtlsState: "connected" },
      { type: "candidate-pair", state: "failed", nominated: true },
      { type: "inbound-rtp", packetsLost: 0, bytesReceived: 1 }
    ];
    const out = collectAggregate(makeReport(entries), "connected", "completed");
    expect(out?.iceState).toBe("completed");
  });

  it("mappt echte RTCStatsIceCandidatePairState-Werte als Fallback", () => {
    const entries = [
      { type: "transport", dtlsState: "connected" },
      { type: "candidate-pair", state: "succeeded", nominated: true },
      { type: "inbound-rtp", packetsLost: 0, bytesReceived: 1 }
    ];
    const out = collectAggregate(makeReport(entries), "connected");
    expect(out?.iceState).toBe("connected");
  });

  it("ignoriert negative Counter-Werte (pin auf nicht-negative Integer)", () => {
    const entries = [
      { type: "transport", dtlsState: "connected" },
      { type: "candidate-pair", state: "succeeded", nominated: true },
      { type: "inbound-rtp", packetsLost: -1, bytesReceived: 100 },
      { type: "outbound-rtp", bytesSent: 50 }
    ];
    const out = collectAggregate(makeReport(entries), "connected");
    expect(out?.packetsLost).toBe(0);
  });
});

describe("attachWebRtc — getStats-Sampling-Loop (Tranche 3)", () => {
  function makePcWithStats(stats: RTCStatsReport, connectionState = "connected") {
    const fakePc = makeFakePeerConnection({});
    fakePc.connectionState = connectionState as RTCPeerConnectionState;
    fakePc.iceConnectionState = connectionState === "connected" ? "connected" : "new";
    (fakePc as unknown as { getStats: () => Promise<RTCStatsReport> }).getStats = async () => stats;
    return fakePc;
  }

  it("emittiert metrics_sampled mit reservierten webrtc.*-Keys", async () => {
    const tracker = new StubTracker();
    const stats = new Map<string, unknown>([
      ["t", { type: "transport", dtlsState: "connected" }],
      ["p", { type: "candidate-pair", state: "connected", nominated: true }],
      ["i", { type: "inbound-rtp", packetsLost: 7, bytesReceived: 1000 }],
      ["o", { type: "outbound-rtp", bytesSent: 500 }]
    ]) as unknown as RTCStatsReport;
    const fakePc = makePcWithStats(stats, "connected");
    let tickFn: (() => void) | undefined;
    const fakeSetInterval = vi.fn((cb: () => void) => {
      tickFn = cb;
      return 1;
    }) as unknown as typeof setInterval;
    const fakeClearInterval = vi.fn() as unknown as typeof clearInterval;

    attachWebRtc(
      fakeVideo,
      { whepUrl: baseOptions.whepUrl, samplingIntervalMs: 100 },
      tracker,
      {
        PeerConnection: function () {
          return fakePc;
        } as unknown as typeof RTCPeerConnection,
        fetch: makeFakeFetch({}),
        sampling: { setInterval: fakeSetInterval, clearInterval: fakeClearInterval },
        newRunId: () => "run-id-test"
      }
    );
    expect(fakeSetInterval).toHaveBeenCalledTimes(1);

    tickFn?.();
    await new Promise((r) => setTimeout(r, 5));

    const sampled = tracker.events.find((e) => e.eventName === "metrics_sampled");
    expect(sampled).toBeDefined();
    expect(sampled?.meta).toEqual({
      "webrtc.peer_connection_run_id": "run-id-test",
      "webrtc.sample_id": 0,
      "webrtc.connection_state": "connected",
      "webrtc.ice_state": "connected",
      "webrtc.dtls_state": "connected",
      "webrtc.packets_lost": 7,
      "webrtc.bytes_received": 1000,
      "webrtc.bytes_sent": 500
    });
  });

  it("emittiert kein metrics_sampled, wenn die Sample-Aggregation null ist", async () => {
    const tracker = new StubTracker();
    const stats = new Map<string, unknown>([
      ["i", { type: "inbound-rtp", packetsLost: 1, bytesReceived: 1 }]
    ]) as unknown as RTCStatsReport;
    const fakePc = makePcWithStats(stats, "connected");
    let tickFn: (() => void) | undefined;
    const fakeSetInterval = vi.fn((cb: () => void) => {
      tickFn = cb;
      return 1;
    }) as unknown as typeof setInterval;
    const fakeClearInterval = vi.fn() as unknown as typeof clearInterval;

    attachWebRtc(
      fakeVideo,
      { whepUrl: baseOptions.whepUrl, samplingIntervalMs: 100 },
      tracker,
      {
        PeerConnection: function () {
          return fakePc;
        } as unknown as typeof RTCPeerConnection,
        fetch: makeFakeFetch({}),
        sampling: { setInterval: fakeSetInterval, clearInterval: fakeClearInterval }
      }
    );
    tickFn?.();
    await new Promise((r) => setTimeout(r, 5));
    expect(tracker.events.find((e) => e.eventName === "metrics_sampled")).toBeUndefined();
  });

  it("destroy() stoppt das Sampling-Intervall", () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const fakeSetInterval = vi.fn(() => 42) as unknown as typeof setInterval;
    const fakeClearInterval = vi.fn() as unknown as typeof clearInterval;

    const adapter = attachWebRtc(
      fakeVideo,
      { whepUrl: baseOptions.whepUrl, samplingIntervalMs: 100 },
      tracker,
      {
        PeerConnection: function () {
          return fakePc;
        } as unknown as typeof RTCPeerConnection,
        fetch: makeFakeFetch({}),
        sampling: { setInterval: fakeSetInterval, clearInterval: fakeClearInterval }
      }
    );
    adapter.destroy();
    expect(fakeClearInterval).toHaveBeenCalled();
  });

  it("samplingIntervalMs=0 deaktiviert das Sampling vollständig", () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    const fakeSetInterval = vi.fn() as unknown as typeof setInterval;

    attachWebRtc(
      fakeVideo,
      { whepUrl: baseOptions.whepUrl, samplingIntervalMs: 0 },
      tracker,
      {
        PeerConnection: function () {
          return fakePc;
        } as unknown as typeof RTCPeerConnection,
        fetch: makeFakeFetch({}),
        sampling: { setInterval: fakeSetInterval }
      }
    );
    expect(fakeSetInterval).not.toHaveBeenCalled();
  });

  it("nimmt erstes valides candidate-pair, wenn keines nominated/selected ist", () => {
    const stats = new Map<string, unknown>([
      ["t", { type: "transport", dtlsState: "connected" }],
      ["p1", { type: "candidate-pair", state: "waiting" }],
      ["p2", { type: "candidate-pair", state: "succeeded" }],
      ["i", { type: "inbound-rtp", packetsLost: 0, bytesReceived: 1 }]
    ]) as unknown as RTCStatsReport;
    const out = collectAggregate(stats, "connecting");
    expect(out?.iceState).toBe("new");
  });

  it("Tick ohne pc.getStats() emittiert kein metrics_sampled", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({});
    // makeFakePeerConnection({}) liefert absichtlich kein getStats —
    // der Tick-Pfad muss den Sample leise verwerfen.
    let tickFn: (() => void) | undefined;
    const fakeSetInterval = vi.fn((cb: () => void) => {
      tickFn = cb;
      return 1;
    }) as unknown as typeof setInterval;
    attachWebRtc(
      fakeVideo,
      { whepUrl: baseOptions.whepUrl, samplingIntervalMs: 100 },
      tracker,
      {
        PeerConnection: function () {
          return fakePc;
        } as unknown as typeof RTCPeerConnection,
        fetch: makeFakeFetch({}),
        sampling: { setInterval: fakeSetInterval, clearInterval: vi.fn() as unknown as typeof clearInterval }
      }
    );
    tickFn?.();
    await new Promise((r) => setTimeout(r, 5));
    expect(tracker.events.find((e) => e.eventName === "metrics_sampled")).toBeUndefined();
  });

  it("nutzt crypto.randomUUID, wenn keine newRunId-deps gegeben ist", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({ emitConnected: true });
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({})
    });
    await new Promise((r) => setTimeout(r, 5));
    const started = tracker.events.find((e) => e.eventName === "playback_started");
    const runId = started?.meta?.["webrtc.peer_connection_run_id"];
    expect(typeof runId).toBe("string");
    expect((runId as string).length).toBeGreaterThan(0);
  });

  it("playback_started enthält peer_connection_run_id", async () => {
    const tracker = new StubTracker();
    const fakePc = makeFakePeerConnection({ emitConnected: true });
    attachWebRtc(fakeVideo, baseOptions, tracker, {
      PeerConnection: function () {
        return fakePc;
      } as unknown as typeof RTCPeerConnection,
      fetch: makeFakeFetch({}),
      newRunId: () => "fixed-run-id"
    });
    await new Promise((r) => setTimeout(r, 5));
    const started = tracker.events.find((e) => e.eventName === "playback_started");
    expect(started?.meta?.["webrtc.peer_connection_run_id"]).toBe("fixed-run-id");
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
