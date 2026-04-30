import { describe, expect, it } from "vitest";
import { SessionMetrics } from "../src/core/session-metrics";

describe("SessionMetrics", () => {
  it("records startup time once", () => {
    const metrics = new SessionMetrics(100);

    expect(metrics.completeStartup(340)).toBe(240);
    expect(metrics.completeStartup(500)).toBeUndefined();
    expect(metrics.snapshot()).toEqual({
      startupTimeMs: 240,
      rebufferCount: 0,
      totalRebufferDurationMs: 0
    });
  });

  it("records rebuffer duration and cumulative totals", () => {
    const metrics = new SessionMetrics(0);

    expect(metrics.startRebuffer(1000)).toBe(true);
    expect(metrics.startRebuffer(1200)).toBe(false);
    expect(metrics.endRebuffer(1450)).toEqual({
      durationMs: 450,
      rebufferCount: 1,
      totalRebufferDurationMs: 450
    });
    expect(metrics.endRebuffer(1500)).toBeUndefined();

    expect(metrics.startRebuffer(2000)).toBe(true);
    expect(metrics.endRebuffer(2250)).toEqual({
      durationMs: 250,
      rebufferCount: 2,
      totalRebufferDurationMs: 700
    });
  });

  it("clamps negative durations to zero", () => {
    const metrics = new SessionMetrics(500);

    expect(metrics.completeStartup(400)).toBe(0);
  });
});
