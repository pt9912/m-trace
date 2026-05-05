import { afterEach, describe, expect, it, vi } from "vitest";
import {
  formatBandwidthMbps,
  getSrtHealthDetail,
  isSrtSampleStale,
  listSrtHealth,
  type SrtHealthItem
} from "../src/lib/api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("dashboard SRT-Health API client", () => {
  it("calls GET /api/srt/health for the list endpoint", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ items: [] }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(listSrtHealth()).resolves.toEqual({ items: [] });

    expect(fetchMock).toHaveBeenCalledWith("/api/srt/health", {
      headers: { Accept: "application/json" },
      cache: "no-store"
    });
  });

  it("encodes stream id and forwards samples_limit for detail requests", async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ stream_id: "srt/test", items: [] }));
    vi.stubGlobal("fetch", fetchMock);

    await getSrtHealthDetail("srt/test", 25);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/srt/health/srt%2Ftest?samples_limit=25",
      {
        headers: { Accept: "application/json" },
        cache: "no-store"
      }
    );
  });

  it("formats bandwidth in Mbit/s with three decimal places", () => {
    expect(formatBandwidthMbps(4_352_217_617)).toBe("4352.218 Mbit/s");
    expect(formatBandwidthMbps(0)).toBe("0.000 Mbit/s");
  });
});

describe("isSrtSampleStale", () => {
  function baseItem(overrides: Partial<SrtHealthItem> = {}): SrtHealthItem {
    return {
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
        collected_at: "2026-05-05T12:00:00.000Z",
        ingested_at: "2026-05-05T12:00:00.000Z",
        sample_age_ms: 1000,
        stale_after_ms: 15000
      },
      ...overrides
    };
  }

  it("returns true when source_status is stale", () => {
    expect(isSrtSampleStale(baseItem({ source_status: "stale" }))).toBe(true);
  });

  it("returns true when sample age exceeds stale_after_ms", () => {
    expect(
      isSrtSampleStale(
        baseItem({
          freshness: {
            source_observed_at: null,
            source_sequence: "1",
            collected_at: "2026-05-05T12:00:00.000Z",
            ingested_at: "2026-05-05T12:00:00.000Z",
            sample_age_ms: 20_000,
            stale_after_ms: 15_000
          }
        })
      )
    ).toBe(true);
  });

  it("returns false for a fresh, healthy sample", () => {
    expect(isSrtSampleStale(baseItem())).toBe(false);
  });
});

function jsonResponse(payload: unknown): Response {
  return new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" }
  });
}
