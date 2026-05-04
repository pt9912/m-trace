import type { PlayerSDKConfig, Transport } from "../types/config";
import type { BoundaryDraft, EventDraft, PlaybackEvent, PlaybackEventBatch, SessionBoundary } from "../types/events";
import { HttpTransport } from "../transport/http";
import { EVENT_SCHEMA_VERSION, PLAYER_SDK_NAME, PLAYER_SDK_VERSION } from "../version";
import { createSessionId } from "./session";

const sdk = { name: PLAYER_SDK_NAME, version: PLAYER_SDK_VERSION };
const maxBatchEvents = 100;
const maxBatchBodyBytes = 256 * 1024;
const maxBatchBoundaries = 20;
const defaultMaxQueueEvents = 1000;

export interface PlayerTracker {
  readonly sessionId: string;
  track(event: EventDraft): void;
  /**
   * Reichts einen `session_boundaries[]`-Eintrag in den nächsten
   * Batch ein, der mindestens ein Event derselben Session enthält
   * (plan-0.4.0 §4.4 / §4.6). Maximal 20 Boundaries pro Batch:
   * Boundaries jenseits des Caps werden mit drop-oldest verworfen.
   * Der Tracker setzt `kind="network_signal_absent"`, `project_id`,
   * `session_id` und `client_timestamp` automatisch; Caller liefern
   * nur `networkKind`, `adapter`, `reason` und optional einen
   * `timestamp`.
   */
  addBoundary(boundary: BoundaryDraft): void;
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
  private readonly boundaryQueue: SessionBoundary[] = [];
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

  addBoundary(boundary: BoundaryDraft): void {
    if (this.destroyed) {
      return;
    }
    const entry: SessionBoundary = {
      kind: "network_signal_absent",
      project_id: this.projectId,
      session_id: this.sessionId,
      network_kind: boundary.networkKind,
      adapter: boundary.adapter,
      reason: boundary.reason,
      client_timestamp: (boundary.timestamp ?? new Date()).toISOString()
    };
    this.boundaryQueue.push(entry);
    while (this.boundaryQueue.length > maxBatchBoundaries) {
      // Drop-oldest: beim Cap-Überlauf verlieren wir die ältesten
      // Boundaries — neuere Capability-Signale haben Vorrang in der
      // Diagnose. Backend lehnt mehr als 20 pro Batch sowieso ab.
      this.boundaryQueue.shift();
    }
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

    // plan-0.4.0 §4.4: Boundaries dürfen nur mit einem Batch
    // gehen, der mindestens ein Event für dieselbe Session
    // enthält. Da der Tracker single-session ist, reicht es,
    // das `events.length > 0`-Predicate zu prüfen — Backend
    // enforced den Partition-Match zusätzlich. Wir hängen die
    // Boundaries an den **ersten** Batch dran, der durch diesen
    // Flush-Call rausgeht; weitere Batch-Splits durch
    // Body-Size-Limit gehen ohne Boundaries.
    //
    // Snapshot am Flush-Start (statt pro Loop-Iteration), damit
    // ein `addBoundary`-Aufruf während `await transport.send` der
    // ersten Iteration nicht in den zweiten Batch gerät und das
    // "nur erster Batch trägt Boundaries"-Invariant verletzt.
    // Spätere addBoundary-Aufrufe warten auf den nächsten flush-
    // Cycle.
    const boundariesForThisFlush = this.boundaryQueue.splice(0, this.boundaryQueue.length);
    let pendingBoundaries: SessionBoundary[] = boundariesForThisFlush;

    while (this.queue.length > 0) {
      const events = this.drainNextBatch(pendingBoundaries);
      if (events.length === 0) {
        return;
      }
      const batch: PlaybackEventBatch = {
        schema_version: EVENT_SCHEMA_VERSION,
        events
      };
      if (pendingBoundaries.length > 0) {
        batch.session_boundaries = pendingBoundaries;
        pendingBoundaries = [];
      }
      await this.transport.send(batch);
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

  // drainNextBatch sammelt Events bis zum Body-Size- oder
  // MaxBatch-Limit. `reservedBoundaries` zählt vorab in den Body-
  // Size-Budget, damit ein 256 KiB-grenzwertiger Event-Stream die
  // Boundaries nicht verdrängt; bei Folge-Batches wird ein leerer
  // Slice übergeben (Boundaries waren bereits am ersten Batch).
  private drainNextBatch(reservedBoundaries: SessionBoundary[]): PlaybackEvent[] {
    const events: PlaybackEvent[] = [];

    while (this.queue.length > 0 && events.length < maxBatchEvents) {
      const event = this.queue.shift();
      if (event === undefined) {
        break;
      }

      if (batchSizeBytes([...events, event], reservedBoundaries) <= maxBatchBodyBytes) {
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

function batchSizeBytes(events: PlaybackEvent[], boundaries: SessionBoundary[] = []): number {
  const payload: PlaybackEventBatch = {
    schema_version: EVENT_SCHEMA_VERSION,
    events
  };
  if (boundaries.length > 0) {
    payload.session_boundaries = boundaries;
  }
  return new TextEncoder().encode(JSON.stringify(payload)).length;
}
