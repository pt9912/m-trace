import type { PlaybackEventBatch } from "./events";

export interface Transport {
  send(batch: PlaybackEventBatch): Promise<void>;
}

export interface PlayerSDKConfig {
  endpoint: string;
  token: string;
  projectId: string;
  sessionId?: string;
  batchSize?: number;
  flushIntervalMs?: number;
  sampleRate?: number;
  maxQueueEvents?: number;
  transport?: Transport;
}
