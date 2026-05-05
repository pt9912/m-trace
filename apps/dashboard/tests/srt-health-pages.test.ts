import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { SrtHealthDetailResponse, SrtHealthListResponse } from "../src/lib/api";

const routeState = vi.hoisted<{ params: { stream_id?: string } }>(() => ({
  params: { stream_id: "srt-test" }
}));

const apiState = vi.hoisted<{
  list: SrtHealthListResponse;
  detail: SrtHealthDetailResponse;
  listError: Error | undefined;
  detailError: Error | undefined;
}>(() => ({
  list: { items: [] },
  detail: { stream_id: "srt-test", items: [] },
  listError: undefined,
  detailError: undefined
}));

vi.mock("$app/stores", () => ({
  page: {
    subscribe: (run: (value: { params: { stream_id?: string } }) => void) => {
      run({ params: routeState.params });
      return () => undefined;
    }
  }
}));

vi.mock("$lib/api", async () => {
  const actual = await vi.importActual<typeof import("../src/lib/api")>("../src/lib/api");
  return {
    ...actual,
    listSrtHealth: vi.fn(async () => {
      if (apiState.listError) {
        throw apiState.listError;
      }
      return apiState.list;
    }),
    getSrtHealthDetail: vi.fn(async () => {
      if (apiState.detailError) {
        throw apiState.detailError;
      }
      return apiState.detail;
    })
  };
});

beforeEach(() => {
  routeState.params = { stream_id: "srt-test" };
  apiState.list = { items: [] };
  apiState.detail = { stream_id: "srt-test", items: [] };
  apiState.listError = undefined;
  apiState.detailError = undefined;
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("SRT health list page", () => {
  it("renders the empty-state hint when no streams reported", async () => {
    const { default: SrtHealthPage } = await import("../src/routes/srt-health/+page.svelte");
    render(SrtHealthPage);

    expect(
      await screen.findByText(/Collector may be disabled/i, undefined, { timeout: 1000 })
    ).toBeTruthy();
  });

  it("renders one row per stream with the four required metrics", async () => {
    apiState.list = {
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "healthy",
          source_status: "ok",
          source_error_code: "none",
          connection_state: "connected",
          metrics: {
            rtt_ms: 0.231,
            packet_loss_total: 7,
            retransmissions_total: 3,
            available_bandwidth_bps: 4_352_217_617
          },
          derived: {},
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T08:48:01.000Z",
            ingested_at: "2026-05-05T08:48:01.250Z",
            sample_age_ms: 250,
            stale_after_ms: 15000
          }
        }
      ]
    };
    const { default: SrtHealthPage } = await import("../src/routes/srt-health/+page.svelte");
    render(SrtHealthPage);

    expect(await screen.findByRole("link", { name: "srt-test" })).toBeTruthy();
    expect(screen.getByText("0.23 ms")).toBeTruthy();
    expect(screen.getByText("7")).toBeTruthy();
    expect(screen.getByText("3")).toBeTruthy();
    expect(screen.getByText("4352.218 Mbit/s")).toBeTruthy();
    expect(screen.getByText("healthy")).toBeTruthy();
  });

  it("renders the source-status hint and error-code column for non-OK items", async () => {
    apiState.list = {
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "unknown",
          source_status: "stale",
          source_error_code: "stale_sample",
          connection_state: "connected",
          metrics: {
            rtt_ms: 0.5,
            packet_loss_total: 0,
            retransmissions_total: 0,
            available_bandwidth_bps: 4_000_000_000
          },
          derived: {},
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T08:48:01.000Z",
            ingested_at: "2026-05-05T08:48:01.000Z",
            sample_age_ms: 30_000,
            stale_after_ms: 15_000
          }
        }
      ]
    };
    const { default: SrtHealthPage } = await import("../src/routes/srt-health/+page.svelte");
    render(SrtHealthPage);

    expect(await screen.findByText(/source: stale/i, undefined, { timeout: 1000 })).toBeTruthy();
    expect(screen.getByText("stale_sample")).toBeTruthy();
  });

  it("polls listSrtHealth every 5s after mount", async () => {
    vi.useFakeTimers();
    const apiModule = await import("../src/lib/api");
    const listSpy = vi.mocked(apiModule.listSrtHealth);
    listSpy.mockClear();

    const { default: SrtHealthPage } = await import("../src/routes/srt-health/+page.svelte");
    render(SrtHealthPage);
    // Initial refresh microtask + setInterval-Tick anstoßen.
    await vi.advanceTimersByTimeAsync(5_500);
    expect(listSpy.mock.calls.length).toBeGreaterThanOrEqual(2);
    vi.useRealTimers();
  });

  it("renders an error message when the API call fails", async () => {
    apiState.listError = new Error("api down");
    const { default: SrtHealthPage } = await import("../src/routes/srt-health/+page.svelte");
    render(SrtHealthPage);

    expect(await screen.findByText("api down")).toBeTruthy();
  });

  it("shows the stale label when sample age exceeds stale_after_ms", async () => {
    apiState.list = {
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "healthy",
          source_status: "ok",
          source_error_code: "none",
          connection_state: "connected",
          metrics: {
            rtt_ms: 0.5,
            packet_loss_total: 0,
            retransmissions_total: 0,
            available_bandwidth_bps: 4_000_000_000
          },
          derived: {},
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T08:48:01.000Z",
            ingested_at: "2026-05-05T08:48:01.000Z",
            sample_age_ms: 30_000,
            stale_after_ms: 15_000
          }
        }
      ]
    };
    const { default: SrtHealthPage } = await import("../src/routes/srt-health/+page.svelte");
    render(SrtHealthPage);

    expect(await screen.findByText(/healthy \(stale\)/i)).toBeTruthy();
  });
});

describe("SRT health detail page", () => {
  it("renders 'No persisted health samples' when the API returns 404", async () => {
    apiState.detailError = new Error("/api/srt/health/missing returned 404");
    routeState.params = { stream_id: "missing" };
    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);

    expect(await screen.findByText(/has no persisted health samples/i, undefined, { timeout: 1000 })).toBeTruthy();
  });

  it("renders the current sample plus history rows", async () => {
    const sample = {
      stream_id: "srt-test",
      connection_id: "c1",
      health_state: "degraded" as const,
      source_status: "ok" as const,
      source_error_code: "none" as const,
      connection_state: "connected" as const,
      metrics: {
        rtt_ms: 150,
        packet_loss_total: 5,
        retransmissions_total: 2,
        available_bandwidth_bps: 4_000_000_000,
        required_bandwidth_bps: 1_500_000,
        throughput_bps: 1_200_000
      },
      derived: { bandwidth_headroom_factor: 2666.67 },
      freshness: {
        source_observed_at: null,
        source_sequence: "1",
        collected_at: "2026-05-05T12:00:00.000Z",
        ingested_at: "2026-05-05T12:00:00.500Z",
        sample_age_ms: 500,
        stale_after_ms: 15_000
      }
    };
    apiState.detail = {
      stream_id: "srt-test",
      items: [sample, { ...sample, freshness: { ...sample.freshness, ingested_at: "2026-05-05T11:59:55.000Z" } }]
    };
    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);

    expect(await screen.findByText("Current", undefined, { timeout: 1000 })).toBeTruthy();
    expect(screen.getByText("History")).toBeTruthy();
    expect(screen.getByText("150.00 ms")).toBeTruthy();
    // 4000.000 Mbit/s erscheint in Current + History (zwei History-
    // Einträge mit selbem Sample-Wert) — getAllByText um die Mehrdeutig-
    // keit zu zeigen, statt sie zu erzwingen.
    expect(screen.getAllByText("4000.000 Mbit/s").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("1.500 Mbit/s")).toBeTruthy();
    expect(screen.getByText(/×2666\.67/)).toBeTruthy();
  });

  it("shows error message on non-404 failures", async () => {
    apiState.detailError = new Error("server boom");
    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);

    expect(await screen.findByText("server boom")).toBeTruthy();
  });

  it("renders the source_observed_at timestamp when the source provides one", async () => {
    apiState.detail = {
      stream_id: "srt-test",
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "healthy",
          source_status: "ok",
          source_error_code: "none",
          connection_state: "connected",
          metrics: {
            rtt_ms: 1,
            packet_loss_total: 0,
            retransmissions_total: 0,
            available_bandwidth_bps: 4_000_000_000
          },
          derived: {},
          freshness: {
            source_observed_at: "2026-05-05T11:59:59.000Z",
            source_sequence: "1",
            collected_at: "2026-05-05T12:00:00.000Z",
            ingested_at: "2026-05-05T12:00:00.000Z",
            sample_age_ms: 100,
            stale_after_ms: 15_000
          }
        }
      ]
    };
    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);

    expect(await screen.findByText("Source observed at", undefined, { timeout: 1000 })).toBeTruthy();
    // formatTime renders the timestamp (mocked or real); not "not provided".
    expect(screen.queryByText(/not provided by source/i)).toBeNull();
  });

  it("renders the stale pill when source_status is stale", async () => {
    apiState.detail = {
      stream_id: "srt-test",
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "unknown",
          source_status: "stale",
          source_error_code: "stale_sample",
          connection_state: "connected",
          metrics: {
            rtt_ms: 1,
            packet_loss_total: 0,
            retransmissions_total: 0,
            available_bandwidth_bps: 4_000_000_000
          },
          derived: {},
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T12:00:00.000Z",
            ingested_at: "2026-05-05T12:00:00.000Z",
            sample_age_ms: 30_000,
            stale_after_ms: 15_000
          }
        }
      ]
    };
    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);

    expect(await screen.findByText("Current", undefined, { timeout: 1000 })).toBeTruthy();
    expect(screen.getByText(/stale_sample/)).toBeTruthy();
  });

  it("does not call getSrtHealthDetail when stream_id route param is missing", async () => {
    routeState.params = { stream_id: undefined };
    const apiModule = await import("../src/lib/api");
    const detailSpy = vi.mocked(apiModule.getSrtHealthDetail);
    detailSpy.mockClear();

    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);
    await Promise.resolve();
    await Promise.resolve();
    expect(detailSpy).not.toHaveBeenCalled();
  });

  it("polls getSrtHealthDetail every 5s after mount", async () => {
    vi.useFakeTimers();
    apiState.detail = {
      stream_id: "srt-test",
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "healthy",
          source_status: "ok",
          source_error_code: "none",
          connection_state: "connected",
          metrics: {
            rtt_ms: 1,
            packet_loss_total: 0,
            retransmissions_total: 0,
            available_bandwidth_bps: 4_000_000_000
          },
          derived: {},
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T12:00:00.000Z",
            ingested_at: "2026-05-05T12:00:00.000Z",
            sample_age_ms: 100,
            stale_after_ms: 15_000
          }
        }
      ]
    };
    const apiModule = await import("../src/lib/api");
    const detailSpy = vi.mocked(apiModule.getSrtHealthDetail);
    detailSpy.mockClear();

    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);
    await vi.advanceTimersByTimeAsync(5_500);
    expect(detailSpy.mock.calls.length).toBeGreaterThanOrEqual(2);
    vi.useRealTimers();
  });

  it("triggers a refresh when the Refresh button is clicked", async () => {
    apiState.detail = {
      stream_id: "srt-test",
      items: [
        {
          stream_id: "srt-test",
          connection_id: "c1",
          health_state: "healthy",
          source_status: "ok",
          source_error_code: "none",
          connection_state: "connected",
          metrics: {
            rtt_ms: 1,
            packet_loss_total: 0,
            retransmissions_total: 0,
            available_bandwidth_bps: 4_000_000_000
          },
          derived: {},
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T12:00:00.000Z",
            ingested_at: "2026-05-05T12:00:00.000Z",
            sample_age_ms: 100,
            stale_after_ms: 15_000
          }
        }
      ]
    };
    const apiModule = await import("../src/lib/api");
    const detailSpy = vi.mocked(apiModule.getSrtHealthDetail);
    const callsBefore = detailSpy.mock.calls.length;
    const { default: DetailPage } = await import("../src/routes/srt-health/[stream_id]/+page.svelte");
    render(DetailPage);
    await screen.findByText("Current");

    const refreshButtons = screen.getAllByRole("button", { name: "Refresh" });
    await fireEvent.click(refreshButtons[0]);
    await Promise.resolve();
    await Promise.resolve();

    expect(detailSpy.mock.calls.length).toBeGreaterThan(callsBefore + 1);
  });
});
