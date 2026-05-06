import { describe, expect, it, vi } from "vitest";
import {
  attachWebRtc,
  type WebRtcAdapter,
  type WebRtcAdapterOptions
} from "../src/adapters/webrtc/adapter";
import type { PlayerTracker } from "../src/core/tracker";

// plan-0.8.0 Tranche 1 — Public-API-Vertrag des WebRTC-Adapters.
// Tests pinnen die Surface (Funktion + Types) ohne Browser-
// Signalisierung; Tranche 2 ergänzt Verhaltens-Tests gegen den WHEP-
// Pfad. `attachWebRtc` wirft in Tranche 1 deterministisch — der Test
// hält das fest, damit ein versehentlich vor Tranche 2 produktiv
// gesetzter Adapter sofort auffällt.

class StubTracker implements PlayerTracker {
  readonly sessionId = "test-session";
  track = vi.fn();
  addBoundary = vi.fn();
  flush = vi.fn().mockResolvedValue(undefined);
  destroy = vi.fn().mockResolvedValue(undefined);
}

describe("attachWebRtc — Tranche-1-Public-API-Vertrag", () => {
  it("ist als Funktion exportiert", () => {
    expect(typeof attachWebRtc).toBe("function");
  });

  it("wirft deterministisch, solange Tranche 2 die Implementation nicht geliefert hat", () => {
    const video = {} as unknown as HTMLVideoElement;
    const tracker = new StubTracker();
    const options: WebRtcAdapterOptions = {
      whepUrl: "http://localhost:8892/webrtc-test/whep"
    };

    expect(() => attachWebRtc(video, options, tracker)).toThrow(
      /not implemented \(plan-0.8.0 Tranche 2\)/
    );
    expect(tracker.track).not.toHaveBeenCalled();
  });

  it("akzeptiert die optionale RTCConfiguration und das AbortSignal als Type-Vertrag", () => {
    // Type-Pinning: TypeScript würde bei einer Surface-Änderung einen
    // Compile-Fehler werfen. Der Body muss die Funktion **nicht**
    // aufrufen — die Definition selbst ist der Vertrag.
    const optionsAll: WebRtcAdapterOptions = {
      whepUrl: "http://localhost:8892/webrtc-test/whep",
      peerConnectionConfig: { iceServers: [] },
      signal: new AbortController().signal
    };
    const optionsMinimal: WebRtcAdapterOptions = {
      whepUrl: "http://localhost:8892/webrtc-test/whep"
    };

    expect(optionsAll.whepUrl).toBe(optionsMinimal.whepUrl);
  });

  it("WebRtcAdapter exposed eine destroy()-Surface", () => {
    // Strukturelles Type-Pinning. Wenn der Vertrag um eine weitere
    // Pflicht-Methode erweitert wird, fällt dieser Test in Tranche 2
    // sichtbar an — das ist by design.
    const fake: WebRtcAdapter = {
      destroy: () => undefined
    };
    expect(typeof fake.destroy).toBe("function");
  });
});
