import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const playerMocks = vi.hoisted(() => ({
  createTracker: vi.fn(),
  trackerDestroy: vi.fn(),
  attachWebRtc: vi.fn(),
  adapterDestroy: vi.fn()
}));

vi.mock("@npm9912/player-sdk", () => ({
  createTracker: playerMocks.createTracker,
  attachWebRtc: playerMocks.attachWebRtc
}));

beforeEach(() => {
  playerMocks.trackerDestroy.mockClear();
  playerMocks.adapterDestroy.mockClear();
  playerMocks.createTracker.mockReset();
  playerMocks.attachWebRtc.mockReset();
  playerMocks.createTracker.mockReturnValue({
    sessionId: "session-1",
    track: vi.fn(),
    flush: vi.fn(),
    destroy: playerMocks.trackerDestroy
  });
  playerMocks.attachWebRtc.mockReturnValue({ destroy: playerMocks.adapterDestroy });
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => undefined);
  window.history.replaceState({}, "", "/demo-webrtc?session_id=test-session");
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("demo-webrtc player page (plan-0.8.0 Tranche 2)", () => {
  it("attached den WebRTC-Adapter mit der Default-WHEP-URL und räumt auf Stop auf", async () => {
    const { default: DemoWebRtcPage } = await import("../src/routes/demo-webrtc/+page.svelte");

    render(DemoWebRtcPage);
    await fireEvent.click(screen.getByRole("button", { name: "Start" }));

    expect(playerMocks.createTracker).toHaveBeenCalledWith(
      expect.objectContaining({
        endpoint: "http://localhost:8080/api/playback-events",
        projectId: "demo",
        sessionId: "test-session"
      })
    );
    expect(playerMocks.attachWebRtc).toHaveBeenCalledWith(
      expect.any(HTMLVideoElement),
      expect.objectContaining({ whepUrl: "http://localhost:8892/webrtc-test/whep" }),
      expect.any(Object)
    );

    await fireEvent.click(screen.getByRole("button", { name: "Stop" }));

    expect(playerMocks.adapterDestroy).toHaveBeenCalled();
    expect(playerMocks.trackerDestroy).toHaveBeenCalled();
  });

  it("zeigt einen attach-failed-Status, wenn attachWebRtc wirft", async () => {
    playerMocks.attachWebRtc.mockImplementation(() => {
      throw new Error("RTCPeerConnection unavailable");
    });
    const { default: DemoWebRtcPage } = await import("../src/routes/demo-webrtc/+page.svelte");

    render(DemoWebRtcPage);
    await fireEvent.click(screen.getByRole("button", { name: "Start" }));

    expect(await screen.findByText(/attach failed: RTCPeerConnection unavailable/)).toBeTruthy();
  });

  it("autostartet den Adapter aus dem Query-String", async () => {
    window.history.replaceState({}, "", "/demo-webrtc?autostart=1");
    const { default: DemoWebRtcPage } = await import("../src/routes/demo-webrtc/+page.svelte");

    render(DemoWebRtcPage);

    expect(playerMocks.attachWebRtc).toHaveBeenCalled();
  });
});
