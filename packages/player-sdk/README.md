# @npm9912/player-sdk

Browser SDK for sending m-trace playback telemetry from an `HTMLVideoElement`
or an hls.js player to `POST /api/playback-events`.

## Install

```bash
pnpm add @npm9912/player-sdk hls.js
```

`hls.js` is a peer dependency because applications usually already own the
player version.

## Basic Usage

```ts
import Hls from "hls.js";
import { attachHlsJs, createTracker } from "@npm9912/player-sdk";

const video = document.querySelector("video");
if (!video) {
  throw new Error("video element missing");
}

const tracker = createTracker({
  endpoint: "http://localhost:8080/api/playback-events",
  token: "demo-token",
  projectId: "demo",
  batchSize: 10,
  flushIntervalMs: 5000,
  sampleRate: 1
});

const hls = new Hls();
hls.loadSource("http://localhost:8888/teststream/index.m3u8");
hls.attachMedia(video);

const adapter = attachHlsJs(hls, tracker);

window.addEventListener("pagehide", () => {
  adapter.detach();
  void tracker.destroy();
});
```

## Public API

- `createTracker(config)` creates a `PlayerTracker`.
- `MTracePlayerTracker` is the concrete tracker implementation.
- `HttpTransport` sends batches to the m-trace API.
- `attachHlsJs(hls, tracker)` wires hls.js events into a tracker.
- `createSessionId()` creates a browser-safe random session id.
- `SessionMetrics` tracks startup and rebuffer measurements.

Type exports cover the wire payload and configuration surface:
`PlayerSDKConfig`, `Transport`, `PlaybackEventBatch`, `PlaybackEvent`,
`PlaybackEventName`, `EventDraft`, `EventMeta`, `SDKInfo`, `HlsJsAdapter`,
`PlayerTracker`, `SessionMetricsSnapshot`, and `RebufferMeasurement`.

Deep imports from `src/` or `dist/` are not public API. Import from
`@npm9912/player-sdk`.

## Tracker Lifecycle

`track(event)` queues a playback event. The tracker flushes automatically when
`batchSize` is reached and optionally on `flushIntervalMs`.

`flush()` sends all queued events immediately.

`destroy()` sends one final `session_ended` event, clears timers and flushes the
remaining queue. Calling `destroy()` more than once is safe.

## Configuration

| Option | Required | Description |
|---|---:|---|
| `endpoint` | yes | Full m-trace ingest endpoint URL. |
| `token` | yes | Project token sent as `X-MTrace-Token`. |
| `projectId` | yes | Project id written into each event. |
| `sessionId` | no | Explicit session id; generated when omitted. |
| `batchSize` | no | Events per request. Clamped to the API limit of 100. |
| `flushIntervalMs` | no | Periodic flush interval; `0` disables the timer. |
| `sampleRate` | no | `0..1` event sampling rate. |
| `transport` | no | Custom transport implementing `send(batch)`. |

## Events

The SDK can emit these event names:

- `manifest_loaded`
- `segment_loaded`
- `playback_started`
- `bitrate_switch`
- `rebuffer_started`
- `rebuffer_ended`
- `playback_error`
- `startup_time_measured`
- `metrics_sampled`
- `session_ended`
- `startup_completed`

Each batch uses `schema_version: "1.0"` and includes SDK metadata on every
event. The full wire contract is described in
[`docs/telemetry-model.md`](../../docs/telemetry-model.md).

## Browser Build

The package ships ESM, CJS and IIFE builds. The stable browser entry is the
`browser` field in `package.json`:

```html
<script src="/node_modules/@npm9912/player-sdk/dist/index.global.js"></script>
<script>
  const tracker = MTracePlayerSDK.createTracker({
    endpoint: "http://localhost:8080/api/playback-events",
    token: "demo-token",
    projectId: "demo"
  });
</script>
```

## Error Behavior

`HttpTransport` rejects when the ingest API returns a non-2xx response or when
`fetch` fails. Applications can provide a custom `transport` to integrate
different retry or buffering behavior.
