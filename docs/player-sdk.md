# Player-SDK

> **Status**: `0.2.0`-Arbeitsstand. Dieses Dokument beschreibt die
> projektweite Nutzung des Pakets `@npm9912/player-sdk`.

Das Player-SDK erfasst Playback-Events im Browser und sendet sie im
m-trace-Wire-Format an die API. Der aktuelle Einstiegspunkt ist
`packages/player-sdk`; die Paketdokumentation steht zusätzlich in
[`packages/player-sdk/README.md`](../packages/player-sdk/README.md).

## Installation

```bash
pnpm add @npm9912/player-sdk hls.js
```

`hls.js` ist Peer Dependency. Anwendungen kontrollieren dadurch selbst, welche
Player-Version sie einsetzen.

## Minimalbeispiel

```ts
import Hls from "hls.js";
import { attachHlsJs, createTracker } from "@npm9912/player-sdk";

const tracker = createTracker({
  endpoint: "http://localhost:8080/api/playback-events",
  token: "demo-token",
  projectId: "demo"
});

const hls = new Hls();
hls.loadSource("http://localhost:8888/teststream/index.m3u8");
hls.attachMedia(videoElement);

const adapter = attachHlsJs(hls, tracker);

window.addEventListener("pagehide", () => {
  adapter.detach();
  void tracker.destroy();
});
```

## Public API

Importe laufen über den Package-Entry-Point:

```ts
import {
  HttpTransport,
  SessionMetrics,
  attachHlsJs,
  createSessionId,
  createTracker
} from "@npm9912/player-sdk";
```

Öffentliche Typen:

- `PlayerSDKConfig`
- `Transport`
- `PlayerTracker`
- `PlaybackEventBatch`
- `PlaybackEvent`
- `PlaybackEventName`
- `EventDraft`
- `EventMeta`
- `SDKInfo`
- `HlsJsAdapter`

Tiefe Imports aus `src/` oder `dist/` sind keine stabile API.

## Konfiguration

| Option | Pflicht | Bedeutung |
|---|---:|---|
| `endpoint` | ja | Vollständige URL zu `POST /api/playback-events`. |
| `token` | ja | Projekttoken für den Header `X-MTrace-Token`. |
| `projectId` | ja | Projektkennung im Event-Payload. |
| `sessionId` | nein | Explizite Session-ID; sonst generiert das SDK eine ID. |
| `batchSize` | nein | Events pro Request, hart auf 100 begrenzt. |
| `flushIntervalMs` | nein | Automatischer Flush-Timer; `0` deaktiviert ihn. |
| `sampleRate` | nein | Sampling-Rate zwischen `0` und `1`. |
| `transport` | nein | Eigener Transport mit `send(batch)`. |

## Lifecycle

`track()` reiht Events in die lokale Queue ein. `flush()` sendet die Queue
sofort. `destroy()` beendet die Session, erzeugt genau ein `session_ended`
Event, stoppt Timer und flushed die Queue.

## hls.js-Adapter

`attachHlsJs(hls, tracker)` verbindet hls.js-Events mit dem Tracker. Der
Adapter gibt ein Objekt mit `detach()` zurück. `detach()` entfernt Listener,
zerstört aber nicht den Tracker; der aufrufende Code bleibt für
`tracker.destroy()` verantwortlich.

## Wire-Format

Das SDK sendet Batches mit `schema_version: "1.0"`. Jedes Event enthält
`sdk.name` und `sdk.version`. Das vollständige Datenmodell steht in
[`docs/telemetry-model.md`](./telemetry-model.md), der HTTP-Kontrakt in
[`docs/spike/backend-api-contract.md`](./spike/backend-api-contract.md).
Maschinenlesbare Contract-Artefakte sind
[`contracts/event-schema.json`](../contracts/event-schema.json) und
[`contracts/sdk-compat.json`](../contracts/sdk-compat.json).

## Browser-Build

Das npm-Paket enthält ESM, CJS und IIFE. Der stabile Browser-Einstieg steht im
`browser`-Feld der Paket-Metadaten und zeigt auf
`dist/index.global.js`. Der IIFE-Build exportiert `MTracePlayerSDK` auf dem
globalen Objekt.
