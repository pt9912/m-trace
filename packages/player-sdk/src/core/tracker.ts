import type { PlayerSDKConfig, Transport } from "../types/config";
import type { EventDraft, PlaybackEvent } from "../types/events";
import { HttpTransport } from "../transport/http";
import { EVENT_SCHEMA_VERSION, PLAYER_SDK_NAME, PLAYER_SDK_VERSION } from "../version";
import { createSessionId } from "./session";

const sdk = { name: PLAYER_SDK_NAME, version: PLAYER_SDK_VERSION };
const maxBatchEvents = 100;
const maxBatchBodyBytes = 256 * 1024;
const defaultMaxQueueEvents = 1000;

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
  private readonly maxQueueEvents: number;
  private readonly transport: Transport;
  private readonly queue: PlaybackEvent[] = [];
  private sequence = 0;
  private destroyed = false;
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
    this.batchSize = Math.min(maxBatchEvents, Math.max(1, Math.floor(config.batchSize ?? 10)));
    this.maxQueueEvents = Math.max(1, Math.floor(config.maxQueueEvents ?? defaultMaxQueueEvents));
    this.transport =
      config.transport ??
      new HttpTransport(config.endpoint, config.token, {
        traceparent: config.traceparent
      });

    const flushIntervalMs = Math.max(0, Math.floor(config.flushIntervalMs ?? 5000));
    if (flushIntervalMs > 0) {
      this.flushTimer = setInterval(() => {
        void this.flush();
      }, flushIntervalMs);
    }
  }

  track(event: EventDraft): void {
    if (this.destroyed) {
      return;
    }
    if (this.sampleRate <= 0 || Math.random() >= this.sampleRate) {
      return;
    }
    this.enqueue(event, true, true);
  }

  private enqueue(event: EventDraft, autoFlush: boolean, enforceQueueLimit: boolean): void {
    this.sequence += 1;
    const playbackEvent: PlaybackEvent = {
      event_name: event.eventName,
      project_id: this.projectId,
      session_id: this.sessionId,
      client_timestamp: (event.timestamp ?? new Date()).toISOString(),
      sequence_number: this.sequence,
      sdk
    };
    if (event.meta !== undefined) {
      playbackEvent.meta = event.meta;
    }

    this.queue.push(playbackEvent);
    if (enforceQueueLimit) {
      while (this.queue.length > this.maxQueueEvents) {
        this.queue.shift();
      }
    }

    if (autoFlush && this.queue.length >= this.batchSize) {
      void this.flush();
    }
  }

  async flush(): Promise<void> {
    if (this.queue.length === 0) {
      return;
    }

    while (this.queue.length > 0) {
      const events = this.drainNextBatch();
      if (events.length === 0) {
        return;
      }
      await this.transport.send({
        schema_version: EVENT_SCHEMA_VERSION,
        events
      });
    }
  }

  async destroy(): Promise<void> {
    if (this.destroyed) {
      return;
    }
    this.destroyed = true;
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = undefined;
    }
    this.enqueue({ eventName: "session_ended" }, false, false);
    await this.flush();
  }

  private drainNextBatch(): PlaybackEvent[] {
    const events: PlaybackEvent[] = [];

    while (this.queue.length > 0 && events.length < maxBatchEvents) {
      const event = this.queue.shift();
      if (event === undefined) {
        break;
      }

      if (batchSizeBytes([...events, event]) <= maxBatchBodyBytes) {
        events.push(event);
        continue;
      }

      if (events.length > 0) {
        this.queue.unshift(event);
        break;
      }
    }

    return events;
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

function batchSizeBytes(events: PlaybackEvent[]): number {
  return new TextEncoder().encode(
    JSON.stringify({
      schema_version: EVENT_SCHEMA_VERSION,
      events
    })
  ).length;
}
