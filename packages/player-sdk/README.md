# @pt9912/player-sdk

Browser SDK for sending m-trace playback telemetry from an `HTMLVideoElement`
or an hls.js player to `POST /api/playback-events`.

## Install

```bash
pnpm add @pt9912/player-sdk hls.js
```

`hls.js` is a peer dependency because applications usually already own the
player version.

## Basic Usage

```ts
import Hls from "hls.js";
import { attachHlsJs, createTracker } from "@pt9912/player-sdk";

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

const adapter = attachHlsJs(video, hls, tracker);

window.addEventListener("pagehide", () => {
  adapter.destroy();
  void tracker.destroy();
});
```

## Public API

- `createTracker(config)` creates a `PlayerTracker`.
- `MTracePlayerTracker` is the concrete tracker implementation.
- `HttpTransport` sends batches to the m-trace API.
- `attachHlsJs(video, hls, tracker)` wires video and hls.js events into a tracker (default playback path).
- `attachWebRtc(video, options, tracker)` is the additive WebRTC/WHEP adapter; opt-in per player instance. Public-API contract is shipped in `0.8.0` Tranche 1; the WHEP handshake implementation lands in Tranche 2.
- `createSessionId()` creates a browser-safe random session id.
- `SessionMetrics` tracks startup and rebuffer measurements.

Type exports cover the wire payload and configuration surface:
`PlayerSDKConfig`, `Transport`, `PlaybackEventBatch`, `PlaybackEvent`,
`PlaybackEventName`, `EventDraft`, `EventMeta`, `SDKInfo`, `HlsJsAdapter`,
`WebRtcAdapter`, `WebRtcAdapterOptions`, `PlayerTracker`,
`SessionMetricsSnapshot`, and `RebufferMeasurement`.

Deep imports from `src/` or `dist/` are not public API. Import from
`@pt9912/player-sdk`.

### Adapter selection (hls.js vs. WebRTC)

The SDK ships two playback adapters that share the same `PlayerTracker`
event surface. Adapter selection is the consumer's choice — pick one
per player instance:

- **hls.js** (`attachHlsJs`) is the default and unchanged. Use it for
  HLS manifests; it is the path exercised by `apps/dashboard`'s
  `/demo` route.
- **WebRTC/WHEP** (`attachWebRtc`) is additive and opt-in. It targets
  a WHEP endpoint such as the `examples/webrtc/` lab compose
  (`http://localhost:8892/webrtc-test/whep`). Browser support: Chromium
  120+ and Firefox 120+ are required; Safari is best-effort.

Adapter selection is purely an SDK concern in `0.8.0` Tranche 1: the
wire format, `contracts/sdk-compat.json` and the API ingress are
untouched. Tranche 3 introduces a reserved `webrtc.*` meta namespace
in the wire schema for production WebRTC telemetry; until then,
events emitted via either adapter share the same shape.

#### WebRTC error codes (Tranche 2)

`attachWebRtc` emits a single `playback_error` event per failure with
the reserved meta key `webrtc.error_code`. The SDK normalises any
unknown value to `peer_connection_failed`; consumers can therefore
match the code list directly without sanitising free strings:

| Code | Emitted when |
|------|--------------|
| `whep_signaling_failed` | WHEP `POST` returned non-2xx, or the request itself errored before the server replied. |
| `whep_sdp_invalid` | The WHEP response body did not parse as an SDP answer (no `v=` prefix, or `createOffer()` returned an empty SDP). |
| `webrtc_no_tracks` | Handshake completed but the `RTCPeerConnection` exposed neither audio nor video transceivers. |
| `peer_connection_failed` | `connectionstatechange` reached `failed`/`disconnected`/`closed` before or after `connected`. Generic fallback for unmapped errors. |
| `webrtc_destroyed_before_connected` | Consumer called `adapter.destroy()` before the peer connection reported `connected`. |

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
| `maxQueueEvents` | no | Local queue cap before normal playback events are dropped. Defaults to 1000. |
| `transport` | no | Custom transport implementing `send(batch)`. |

The canonical source for these option semantics, value ranges and defaults
is [`spec/telemetry-model.md`](../../spec/telemetry-model.md) §4.4. The
table above only documents the SDK call-site shape; the contract between
SDK configuration and backend (max 100 events per batch, 256 KiB request
body, drop policy, time-skew) lives in `telemetry-model.md`.

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
[`spec/telemetry-model.md`](../../spec/telemetry-model.md).

### hls.js Mapping (network events)

`attachHlsJs` maps native hls.js callbacks onto `manifest_loaded` and
`segment_loaded` events. The mapping is kept narrow and explicit so
retries and redirects do not generate semantic duplicates. Reference:
`docs/planning/done/plan-0.4.0.md` §4.6,
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §1.4.

| m-trace event | hls.js source | Trigger | Dedup (per session) |
|---|---|---|---|
| `manifest_loaded` (initial) | `MANIFEST_LOADED` | The master/media playlist is loaded the first time. Fires once per session under nominal conditions. | None — every callback emits a fresh event. |
| `manifest_loaded` (reload) | `LEVEL_LOADED` | Live media-playlist reloads, ABR-driven level switches that re-fetch the playlist, master refresh. Each callback emits a fresh event. | None — periodic reloads must remain visible in the timeline; suppressing duplicates would hide live-refresh patterns. |
| `segment_loaded` | `FRAG_LOADED` (success only) | Each successful fragment fetch, including init segments. hls.js does not emit FRAG_LOADED for failed retries (those go via FRAG_LOAD_ERROR / FRAG_LOAD_EMERGENCY_ABORTED), so retries do not duplicate. | `(level, type, sn, cc, isInit)` — `sn === "initSegment"` toggles `isInit`. Doubled listeners or nested player setups that re-fire FRAG_LOADED for the same fragment are dropped. |

The segment dedup key is derived exclusively from hls.js-native
fragment identifiers (`sn`, `cc`, `type`, `level`, init-segment
marker) plus the SDK session context (`project_id`, `session_id`,
sequence). Redacted URLs are persisted as diagnostics under
`meta.network.redacted_url` and **must not** be used as dedup keys —
signed URLs change on every refresh even though the underlying
fragment identity stays stable.

Each emitted event carries the reserved meta keys from
`spec/telemetry-model.md` §1.4:

- `network.kind` — `"manifest"` or `"segment"`.
- `network.detail_status` — `"available"` when timing or URL data is
  usable after redaction; `"network_detail_unavailable"` when the
  browser, CORS, Resource Timing, Service Worker, native HLS or a CDN
  redirect blocks the detail. Documented degradation only.
- `network.unavailable_reason` — set only when `detail_status` is
  `"network_detail_unavailable"`. Reason enum from
  `contracts/event-schema.json#network_unavailable_reasons`.
- `network.redacted_url` — already-redacted URL representative
  (scheme + host + non-token path segments only). The SDK redacts
  fragments, queries, userinfo, signed query parameters, and
  token-like path segments before sending; the API rejects raw URLs
  in this key with `422`.

### URL redaction at the SDK boundary

The SDK applies the `spec/telemetry-model.md` §1.4 redaction matrix
**before** the URL leaves the browser. Tokens never reach the
collector, regardless of the browser's `connect-src` policy:

- query string and fragment are dropped;
- userinfo (`user:pass@`) is dropped;
- token-like path segments (≥ 24 chars and ≥ 80 % `[A-Za-z0-9_-]`,
  even hex ≥ 32, JWT-like three base64url blocks) are replaced with
  `:redacted`;
- prefer the network event's redacted URL over `meta.url` /
  `meta.segment_url` / `meta.manifest_url` — those generic keys are
  redacted defensively too, but `network.redacted_url` is the
  documented surface.

The tracker keeps batches within the API limits of 100 events and 256 KiB
request body size. If a single event cannot fit into one request body, it is
dropped during `flush()` instead of sending a payload the API must reject.

Sampling is event-based for normal playback events. Sampled-out events do not
consume `sequence_number`; `session_ended` bypasses sampling so `destroy()` can
close the session reliably.

**Timeline completeness limit for `sampleRate < 1`** (decision
`docs/planning/done/plan-0.4.0.md` §8.3, variant (b)): full
timeline acceptance and all E2E smokes run with `sampleRate = 1`. With
`sampleRate < 1` the timeline cannot be proven complete without a new
session-/batch-scoped sampling metadata signal — sampled-out events do
not consume a `sequence_number`, so the server cannot tell a sampling
gap apart from a genuine loss. As of `0.4.0`, sampled sessions are
flagged exclusively through documented configuration and operator
notes, not through server-side gap detection. A future tranche may
introduce a durable sampling metadata signal in the read response
(schema migration, read endpoint extension, dashboard marker); that
follow-up becomes release-blocking the moment the first
production-or-lab session with `sampleRate < 1` requires completeness
guarantees.

## Trace correlation

The SDK can propagate a W3C `traceparent` header per batch-send through
the optional `traceparent` provider callback. With no provider the SDK
sends no header and the server generates a root span. The full server
contract — valid-header acceptance, invalid-header parse-error
behavior, single-span-per-batch model, `trace_id` vs `correlation_id`
separation — is normatively documented in
[`spec/telemetry-model.md`](../../spec/telemetry-model.md) §2.5 and
[`spec/player-sdk.md`](../../spec/player-sdk.md) "Trace-Korrelation".

## Browser Build

The package ships ESM, CJS and IIFE builds. The stable browser entry is the
`browser` field in `package.json`:

```html
<script src="/node_modules/@pt9912/player-sdk/dist/index.global.js"></script>
<script>
  const tracker = MTracePlayerSDK.createTracker({
    endpoint: "http://localhost:8080/api/playback-events",
    token: "demo-token",
    projectId: "demo"
  });
</script>
```

## Error Behavior

`HttpTransport` retries network errors, request timeouts, `5xx` responses and
`429` responses up to three attempts by default. `429` respects `Retry-After`
as a cooldown before the next send; without that header it uses exponential
backoff. Non-transient `4xx` responses, including `413`, are not retried.
Applications can provide a custom `transport` to integrate different buffering
behavior.

## Custom and OTel-style Transports

The default bundle has no OpenTelemetry dependency. Applications that already
own an OTel pipeline can inject an opt-in transport through `transport`:

```ts
import type { PlaybackEventBatch, Transport } from "@pt9912/player-sdk";

class OTelLikeTransport implements Transport {
  async send(batch: PlaybackEventBatch): Promise<void> {
    void batch;
  }
}
```

The stable integration point is `Transport.send(batch)`.

## Performance and Browser Support

The performance budget — first set in `0.2.0`, unchanged through `0.8.0` — is:

| Metric | Budget |
|---|---:|
| Bundle size | < 30 KiB gzip without hls.js (incl. additive WebRTC adapter) |
| Event processing | < 5 ms per event in the normal path |
| Hot path | no synchronous network calls |
| Playback safety | telemetry failures must not abort playback |

Run the reproducible smoke with:

```bash
pnpm --filter @pt9912/player-sdk run performance:smoke
```

The general browser matrix is maintained in
[`spec/browser-support.md`](../../spec/browser-support.md).

### WebRTC adapter browser matrix (`0.8.0`)

| Browser | Status | Notes |
|---|---|---|
| Chromium 120+ | Required | `getStats()` shape matches `spec/telemetry-model.md` §3.5.2; `connection_state`, `ice_state`, `dtls_state` are stable Muss-Felder. |
| Firefox 120+ | Required | WHEP handshake, `connection_state`, `ice_state` and RTP/candidate-pair stats are required. Some Playwright/Firefox builds do not expose `RTCStatsType.transport`; per §3.5.3 the adapter drops the `dtls_state` aggregate instead of emitting an `unknown` surrogate. |
| Safari 17+ | Best-effort | `RTCDtlsTransport.dtlsState` and parts of `inbound-rtp` may be missing in older Safari majors. Per the Schema-Drift-Strategy (`spec/telemetry-model.md` §3.5.3) the adapter drops the sample silently rather than emitting an `unknown` surrogate; the WHEP handshake itself remains testable. |
| Other (mobile, embedded) | Out of scope | The lab compose ships only HTTP/WHEP signaling; mobile WebViews and SDK-only consumers without a `<video>` element are out of scope for the production telemetry path in `0.8.0`. |

CI policy (Tranche 4):

- **Release-blocking:** Vitest unit tests, `check-public-api.mjs`
  snapshot, `pack-smoke.mjs` (ESM + CJS + IIFE entries plus
  `dist/index.d.ts`), `performance-smoke.mjs`. These run in
  `make gates` and `make sdk-performance-smoke`.
- **Opt-in / lab-dependent:** Browser-E2E
  (`tests/e2e/dashboard-demo-webrtc.spec.ts`) needs a running
  `mtrace-webrtc` lab compose to exercise the happy path; without
  the lab the test still validates the error path
  (`whep_signaling_failed`) end-to-end through the API session
  detail. Set `MTRACE_WEBRTC_LAB=1` to flip the assertion to
  `playback_started`.
