import type { PlayerSDKConfig, Transport } from "../types/config";
import type { EventDraft, PlaybackEvent } from "../types/events";
import { HttpTransport } from "../transport/http";
import { createSessionId } from "./session";

const sdk = { name: "@m-trace/player-sdk", version: "0.1.1-dev" };

export interface PlayerTracker {
  readonly sessionId: string;
  track(event: EventDraft): void;
  flush(): Promise<void>;
  destroy(): Promise<void>;
}

export class MTracePlayerTracker implements PlayerTracker {
  readonly sessionId: string;

  private readonly projectId: string;
  private readonly sampleRate: number;
  private readonly batchSize: number;
  private readonly transport: Transport;
  private readonly queue: PlaybackEvent[] = [];
  private sequence = 0;
  private flushTimer: ReturnType<typeof setInterval> | undefined;

  constructor(config: PlayerSDKConfig) {
    if (!config.endpoint.trim()) {
      throw new Error("endpoint is required");
    }
    if (!config.token.trim()) {
      throw new Error("token is required");
    }
    if (!config.projectId.trim()) {
      throw new Error("projectId is required");
    }

    this.projectId = config.projectId;
    this.sessionId = config.sessionId ?? createSessionId();
    this.sampleRate = clampSampleRate(config.sampleRate ?? 1);
    this.batchSize = Math.max(1, Math.floor(config.batchSize ?? 10));
    this.transport = config.transport ?? new HttpTransport(config.endpoint, config.token);

    const flushIntervalMs = Math.max(0, Math.floor(config.flushIntervalMs ?? 5000));
    if (flushIntervalMs > 0) {
      this.flushTimer = setInterval(() => {
        void this.flush();
      }, flushIntervalMs);
    }
  }

  track(event: EventDraft): void {
    if (Math.random() > this.sampleRate) {
      return;
    }

    this.sequence += 1;
    this.queue.push({
      event_name: event.eventName,
      project_id: this.projectId,
      session_id: this.sessionId,
      client_timestamp: (event.timestamp ?? new Date()).toISOString(),
      sequence_number: this.sequence,
      sdk
    });

    if (this.queue.length >= this.batchSize) {
      void this.flush();
    }
  }

  async flush(): Promise<void> {
    if (this.queue.length === 0) {
      return;
    }

    const events = this.queue.splice(0, this.queue.length);
    await this.transport.send({
      schema_version: "1.0",
      events
    });
  }

  async destroy(): Promise<void> {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = undefined;
    }
    await this.flush();
  }
}

export function createTracker(config: PlayerSDKConfig): PlayerTracker {
  return new MTracePlayerTracker(config);
}

function clampSampleRate(rate: number): number {
  if (!Number.isFinite(rate)) {
    return 1;
  }
  return Math.min(1, Math.max(0, rate));
}
