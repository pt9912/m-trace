import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const playerMocks = vi.hoisted(() => ({
  hlsSupported: true,
  hlsDestroy: vi.fn(),
  loadSource: vi.fn(),
  attachMedia: vi.fn(),
  createTracker: vi.fn(),
  trackerDestroy: vi.fn(),
  attachHlsJs: vi.fn(),
  adapterDestroy: vi.fn()
}));

vi.mock("@npm9912/player-sdk", () => ({
  createTracker: playerMocks.createTracker,
  attachHlsJs: playerMocks.attachHlsJs
}));

vi.mock("hls.js", () => ({
  default: class HlsMock {
    static isSupported() {
      return playerMocks.hlsSupported;
    }

    loadSource(url: string) {
      playerMocks.loadSource(url);
    }

    attachMedia(video: HTMLVideoElement) {
      playerMocks.attachMedia(video);
    }

    destroy() {
      playerMocks.hlsDestroy();
    }
  }
}));

beforeEach(() => {
  playerMocks.hlsSupported = true;
  playerMocks.hlsDestroy.mockClear();
  playerMocks.loadSource.mockClear();
  playerMocks.attachMedia.mockClear();
  playerMocks.trackerDestroy.mockClear();
  playerMocks.adapterDestroy.mockClear();
  playerMocks.createTracker.mockReset();
  playerMocks.attachHlsJs.mockReset();
  playerMocks.createTracker.mockReturnValue({ sessionId: "session-1", track: vi.fn(), flush: vi.fn(), destroy: playerMocks.trackerDestroy });
  playerMocks.attachHlsJs.mockReturnValue({ destroy: playerMocks.adapterDestroy });
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => undefined);
  vi.spyOn(HTMLMediaElement.prototype, "load").mockImplementation(() => undefined);
  window.history.replaceState({}, "", "/demo?session_id=test-session");
});

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("demo player page", () => {
  it("starts hls.js playback and stops resources", async () => {
    const { default: DemoPage } = await import("../src/routes/demo/+page.svelte");

    render(DemoPage);
    await fireEvent.click(screen.getByRole("button", { name: "Start" }));

    expect(await screen.findByText("Status: hls.js attached")).toBeTruthy();
    expect(playerMocks.createTracker).toHaveBeenCalledWith(
      expect.objectContaining({
        endpoint: "http://localhost:8080/api/playback-events",
        projectId: "demo",
        sessionId: "test-session"
      })
    );
    expect(playerMocks.loadSource).toHaveBeenCalledWith("http://localhost:8888/teststream/index.m3u8");
    expect(playerMocks.attachHlsJs).toHaveBeenCalled();

    await fireEvent.click(screen.getByRole("button", { name: "Stop" }));

    expect(playerMocks.adapterDestroy).toHaveBeenCalled();
    expect(playerMocks.hlsDestroy).toHaveBeenCalled();
    expect(playerMocks.trackerDestroy).toHaveBeenCalled();
  });

  it("falls back to native HLS when hls.js is unsupported", async () => {
    playerMocks.hlsSupported = false;
    vi.spyOn(HTMLVideoElement.prototype, "canPlayType").mockReturnValue("maybe");
    const { default: DemoPage } = await import("../src/routes/demo/+page.svelte");

    render(DemoPage);
    await fireEvent.click(screen.getByRole("button", { name: "Start" }));

    expect(await screen.findByText("Status: native HLS")).toBeTruthy();
    expect(playerMocks.attachHlsJs).not.toHaveBeenCalled();
  });

  it("reports unsupported HLS when neither path is available", async () => {
    playerMocks.hlsSupported = false;
    vi.spyOn(HTMLVideoElement.prototype, "canPlayType").mockReturnValue("");
    const { default: DemoPage } = await import("../src/routes/demo/+page.svelte");

    render(DemoPage);
    await fireEvent.click(screen.getByRole("button", { name: "Start" }));

    expect(await screen.findByText("Status: HLS unsupported")).toBeTruthy();
  });

  it("autostarts playback from the query string", async () => {
    window.history.replaceState({}, "", "/demo?autostart=1");
    const { default: DemoPage } = await import("../src/routes/demo/+page.svelte");

    render(DemoPage);

    expect(await screen.findByText("Status: hls.js attached")).toBeTruthy();
    expect(playerMocks.loadSource).toHaveBeenCalled();
  });
});
