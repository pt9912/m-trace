import { readFileSync } from "node:fs";
import path from "node:path";
import { performance } from "node:perf_hooks";
import { fileURLToPath, pathToFileURL } from "node:url";
import { gzipSync } from "node:zlib";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const packageDir = path.resolve(scriptDir, "..");
const distEntry = path.join(packageDir, "dist", "index.js");
const maxGzipBytes = 30 * 1024;
const maxEventProcessingMs = 5;
const maxBodyBytes = 256 * 1024;

const distCode = readFileSync(distEntry);
const gzipBytes = gzipSync(distCode).length;
assert(gzipBytes < maxGzipBytes, `SDK ESM bundle gzip size ${gzipBytes} bytes must stay below ${maxGzipBytes} bytes`);

const { HttpTransport, createTracker } = await import(pathToFileURL(distEntry).href);

class RecordingTransport {
  batches = [];

  async send(batch) {
    this.batches.push(batch);
  }
}

const hotPathTransport = new RecordingTransport();
const hotPathTracker = createTracker({
  endpoint: "http://localhost:8080/api/playback-events",
  token: "demo-token",
  projectId: "demo",
  batchSize: 100,
  flushIntervalMs: 0,
  transport: hotPathTransport
});

const eventCount = 90;
const startedAt = performance.now();
for (let index = 0; index < eventCount; index += 1) {
  hotPathTracker.track({ eventName: "segment_loaded", meta: { duration_ms: index } });
}
const elapsedMs = performance.now() - startedAt;
const perEventMs = elapsedMs / eventCount;
assert(
  perEventMs < maxEventProcessingMs,
  `event processing ${perEventMs.toFixed(3)} ms/event must stay below ${maxEventProcessingMs} ms/event`
);
assert(hotPathTransport.batches.length === 0, "track() must not synchronously send network batches in the hot path");

await hotPathTracker.flush();
assert(hotPathTransport.batches.length === 1, "flush() must send the queued hot-path batch");
assert(batchBytes(hotPathTransport.batches[0]) <= maxBodyBytes, "flushed batch must stay within the API body limit");

const queueTransport = new RecordingTransport();
const queueTracker = createTracker({
  endpoint: "http://localhost:8080/api/playback-events",
  token: "demo-token",
  projectId: "demo",
  batchSize: 10,
  flushIntervalMs: 0,
  maxQueueEvents: 5,
  transport: queueTransport
});
for (let index = 0; index < 9; index += 1) {
  queueTracker.track({ eventName: "segment_loaded" });
}
await queueTracker.flush();
assert(queueTransport.batches[0]?.events.length === 5, "queue limit must bound locally buffered playback events");

const retrySleeps = [];
const retryTransport = new HttpTransport("http://localhost:8080/api/playback-events", "demo-token", {
  fetchFn: makeRetryFetch(),
  baseDelayMs: 7,
  maxAttempts: 2,
  sleep: async (ms) => {
    retrySleeps.push(ms);
  }
});
await retryTransport.send({ schema_version: "1.0", events: [] });
assert(retrySleeps.length === 1 && retrySleeps[0] === 7, "transport retry boundary must use bounded backoff");

console.log(
  `player-sdk performance smoke ok: gzip=${gzipBytes} bytes, event=${perEventMs.toFixed(3)} ms/event, retries=${retrySleeps.length}`
);

function makeRetryFetch() {
  let calls = 0;
  return async () => {
    calls += 1;
    return new Response(null, { status: calls === 1 ? 503 : 204 });
  };
}

function batchBytes(batch) {
  return new TextEncoder().encode(JSON.stringify(batch)).length;
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}
