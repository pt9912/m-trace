export interface RebufferMeasurement {
  durationMs: number;
  rebufferCount: number;
  totalRebufferDurationMs: number;
}

export interface SessionMetricsSnapshot {
  startupTimeMs?: number;
  rebufferCount: number;
  totalRebufferDurationMs: number;
}

export class SessionMetrics {
  private startupTimeMs: number | undefined;
  private activeRebufferStartedAtMs: number | undefined;
  private rebufferCount = 0;
  private totalRebufferDurationMs = 0;

  constructor(private readonly sessionStartedAtMs: number) {}

  completeStartup(nowMs: number): number | undefined {
    if (this.startupTimeMs !== undefined) {
      return undefined;
    }
    this.startupTimeMs = elapsedMs(this.sessionStartedAtMs, nowMs);
    return this.startupTimeMs;
  }

  startRebuffer(nowMs: number): boolean {
    if (this.activeRebufferStartedAtMs !== undefined) {
      return false;
    }
    this.activeRebufferStartedAtMs = nowMs;
    this.rebufferCount += 1;
    return true;
  }

  endRebuffer(nowMs: number): RebufferMeasurement | undefined {
    if (this.activeRebufferStartedAtMs === undefined) {
      return undefined;
    }

    const durationMs = elapsedMs(this.activeRebufferStartedAtMs, nowMs);
    this.activeRebufferStartedAtMs = undefined;
    this.totalRebufferDurationMs += durationMs;

    return {
      durationMs,
      rebufferCount: this.rebufferCount,
      totalRebufferDurationMs: this.totalRebufferDurationMs
    };
  }

  snapshot(): SessionMetricsSnapshot {
    return {
      startupTimeMs: this.startupTimeMs,
      rebufferCount: this.rebufferCount,
      totalRebufferDurationMs: this.totalRebufferDurationMs
    };
  }
}

function elapsedMs(startMs: number, endMs: number): number {
  return Math.max(0, Math.round(endMs - startMs));
}
