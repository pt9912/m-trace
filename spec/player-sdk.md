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

const adapter = attachHlsJs(videoElement, hls, tracker);

window.addEventListener("pagehide", () => {
  adapter.destroy();
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
- `TraceParentProvider`
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
| `maxQueueEvents` | nein | Lokales Queue-Limit für normale Playback-Events; Standard ist 1000. |
| `transport` | nein | Eigener Transport mit `send(batch)`. |
| `traceparent` | nein | Provider-Funktion für den optionalen W3C-`traceparent`-Header pro Batch-Send (siehe „Trace-Korrelation" unten). |

Die kanonische Quelle für Bedeutung, Wertebereich und Defaults der Sampling-/Batch-Parameter ist [`telemetry-model.md`](./telemetry-model.md) §4.4. Diese Tabelle hier listet nur die SDK-Aufruf-Optionen — wer den Vertrag zwischen SDK-Konfiguration und Backend-Vertrag (Batch-Größe ≤ 100, 256 KiB-Body, Drop-Politik, Time-Skew) nachschlagen will, sollte dort beginnen.

## Lifecycle

`track()` reiht Events in die lokale Queue ein. `flush()` sendet die Queue
sofort und splittet Requests nach den API-Grenzen: maximal 100 Events und
maximal 256 KiB Request-Body. Einzelne Events, die allein nicht in einen
API-Request passen, werden beim Flush verworfen statt als sicher abgelehnter
Payload gesendet.

`sampleRate` wirkt eventbasiert auf normale Playback-Events. Gesampelte Events
verbrauchen keine `sequence_number`. `session_ended` umgeht Sampling, damit
`destroy()` die Session verlässlich schließen kann.

**Timeline-Nachweisgrenze für `sampleRate < 1`** (Beschluss `plan-0.4.0.md` §8.3, Variante (b)): Vollständige Timeline-Abnahme und alle E2E-Smokes laufen mit `sampleRate = 1`. Für `sampleRate < 1` ist Vollständigkeit ohne neues session-/batch-skopiertes Sampling-Metadaten-Signal **nicht beweisbar**, weil gesampelte Events keine `sequence_number` verbrauchen — der Server kann eine fehlende `sequence_number`-Lücke nicht automatisch von einem echten Verlust unterscheiden. Konsequenz für `0.4.0`: Sampled-Sessions werden ausschließlich über dokumentierte Konfiguration und Benutzerhinweis als „sampled" markiert, nicht durch serverseitige Lückenerkennung. Eine zukünftige Tranche kann ein durables Sampling-Metadaten-Signal in der Read-Antwort einführen (Schemamigration, Read-Endpoint-Erweiterung, Dashboard-Markierung); diese Erweiterung wird release-blocking, sobald die erste Produktions- oder Lab-Session mit `sampleRate < 1` Vollständigkeitsnachweise erfordert.

`destroy()` beendet die Session, erzeugt genau ein `session_ended` Event,
stoppt Timer und flushed die Queue.

## hls.js-Adapter

`attachHlsJs(video, hls, tracker)` verbindet Video- und hls.js-Events mit dem
Tracker. Der Adapter gibt ein Objekt mit `destroy()` zurück. `destroy()`
entfernt Listener, zerstört aber nicht den Tracker; der aufrufende Code bleibt für
`tracker.destroy()` verantwortlich.

## Transport-Verhalten

`HttpTransport` wiederholt Netzwerkfehler, Timeouts, `5xx` und `429` begrenzt
auf drei Versuche. `429` mit `Retry-After` wird als Cooldown respektiert; ohne
Header gilt der normale Backoff. Nicht-transiente `4xx` und `413 Payload Too
Large` werden nicht erneut gesendet.

## Trace-Korrelation (optional, ab `0.4.0`)

Das SDK kann pro Batch-Send einen W3C-`traceparent`-Header propagieren —
opt-in über `PlayerSDKConfig.traceparent`. Der Wert kommt aus einem
Provider-Callback, den der Konsument bereitstellt; das SDK selbst hält
keinen Tracer. Ohne Provider sendet das SDK keinen Header, der Server
generiert einen Root-Span. Der vollständige Server-Vertrag — Annahme
gültiger Header, Behandlung ungültiger Header (`mtrace.trace.parse_error`),
fehlender Header, Span-Modell pro Batch, `trace_id`-vs-`correlation_id`-
Trennung, OWS-Verhalten am Wire-Layer — steht normativ in
[`spec/telemetry-model.md`](./telemetry-model.md) §2.5.

> **Scope**: Die Header-Propagation ist eine Eigenschaft des Default-
> `HttpTransport`. Wer einen eigenen `Transport` über
> `PlayerSDKConfig.transport` injiziert, ist selbst verantwortlich, den
> `traceparent`-Provider an seinen Transport-Pfad zu koppeln — das SDK
> ruft den Provider nur im eingebauten HTTP-Pfad auf.

```ts
import { trace } from "@opentelemetry/api";
import { createTracker, type TraceParentProvider } from "@npm9912/player-sdk";

const traceparent: TraceParentProvider = () => {
  const span = trace.getActiveSpan();
  if (!span) return undefined;
  const ctx = span.spanContext();
  if (!ctx.traceId || !ctx.spanId) return undefined;
  const flags = ctx.traceFlags.toString(16).padStart(2, "0");
  return `00-${ctx.traceId}-${ctx.spanId}-${flags}`;
};

const tracker = createTracker({
  endpoint: "http://localhost:8080/api/playback-events",
  token: "demo-token",
  projectId: "demo",
  traceparent
});
```

Format des Header-Werts: `00-<trace_id 32 hex>-<parent_id 16 hex>-<flags 2 hex>`
(W3C [Trace Context](https://www.w3.org/TR/trace-context/)). Das SDK
validiert den Wert **nicht**: ein vom Provider gelieferter, nicht-leerer
Müllstring wird unverändert als `traceparent`-Header gesendet — der
Server markiert ihn als Parse-Error (`mtrace.trace.parse_error=true`)
und fällt auf seine eigene Trace-ID zurück.

Der Provider wird **pro Send synchron** aus dem Default-`HttpTransport`
aufgerufen (kein Caching zwischen Sends, kein Promise-Wrapping). Aus
seiner Rückgabe leitet sich genau eine der folgenden Verhaltensweisen
ab:

| Provider-Rückgabe / -Verhalten | `traceparent`-Header | `console.warn` |
|---|---|---|
| nicht-leerer String | gesetzt auf den Rückgabewert (1:1, ohne Validierung) | nein |
| `undefined` | nicht gesetzt (dokumentiertes Opt-out, kein Fehler) | nein |
| `""` (leerer String) | nicht gesetzt (dokumentiertes Opt-out-Sentinel) | nein |
| Non-String-Wert (z. B. `Promise`, `null`, `number`) | nicht gesetzt | einmal pro `HttpTransport`-Instanz |
| Throw / Exception | nicht gesetzt | einmal pro `HttpTransport`-Instanz |

Der `console.warn` läuft genau einmal pro `HttpTransport`-Instanz, damit
Fehlkonfigurationen (etwa ein versehentlich `Promise<string>` liefernder
Provider) sichtbar werden, ohne den Hot Path bei jedem Send mit Logs zu
fluten. Weitere Fehler derselben Instanz bleiben still. Tests können
die Warnung über `HttpTransportOptions.silent` unterdrücken. In allen
Fällen geht der Batch-Send unverändert weiter — Tracing darf den
Event-Pfad nicht sabotieren.

Abwärtskompatibel mit Backends < `0.4.0`: ältere Server ignorieren
unbekannte Header (HTTP-Standard).

## OpenTelemetry-Vorbereitung

RAK-16 ist in `0.2.0` als vorbereiteter Opt-in-Pfad umgesetzt. Das SDK bringt
keine OTel-Abhängigkeit im Default-Bundle mit. Anwendungen können aber einen
eigenen Transport über `PlayerSDKConfig.transport` injizieren:

```ts
import { createTracker, type PlaybackEventBatch, type Transport } from "@npm9912/player-sdk";

class OTelLikeTransport implements Transport {
  async send(batch: PlaybackEventBatch): Promise<void> {
    // Anwendungsspezifische Übersetzung in OTel-Spans, Logs oder Metriken.
    void batch;
  }
}

const tracker = createTracker({
  endpoint: "http://localhost:8080/api/playback-events",
  token: "demo-token",
  projectId: "demo",
  transport: new OTelLikeTransport()
});
```

Der stabile Port ist `Transport.send(batch)`. Ein späterer offizieller
OTel-Transport muss an diesen Port anschließen und darf den HTTP-Transport
nicht als Default-Pfad ersetzen.

## Performance-Budget

Das SDK übernimmt die normativen MVP-Grenzen aus dem Lastenheft:

| Kennzahl | Budget |
|---|---:|
| Bundle-Größe | < 30 KiB gzip ohne hls.js |
| Event-Verarbeitung | < 5 ms pro Event im Normalfall |
| Hot Path | keine synchronen Netzwerkaufrufe |
| Transport | batchingfähig |
| Fehlerverhalten | Telemetriefehler dürfen Playback nicht abbrechen |
| Sampling | konfigurierbar |

Reproduzierbarer Smoke:

```bash
pnpm --filter @npm9912/player-sdk run performance:smoke
```

Der Smoke baut das SDK, prüft die gzip-Größe des ESM-Bundles, misst
synthetische Event-Verarbeitung und verifiziert Queue-/Retry-Grenzen ohne
echtes Netzwerk.

## Browser-Support

Die Browser-Matrix steht in [`browser-support.md`](./browser-support.md).
Für `0.2.0` sind Chrome Desktop und Firefox Desktop `supported`; Safari
Desktop ist als `documented limitation` klassifiziert.

## Wire-Format

Das SDK sendet Batches mit `schema_version: "1.0"`. Jedes Event enthält
`sdk.name` und `sdk.version`. Das vollständige Datenmodell steht in
[`telemetry-model.md`](./telemetry-model.md), der HTTP-Kontrakt in
[`backend-api-contract.md`](./backend-api-contract.md).
Maschinenlesbare Contract-Artefakte sind
[`contracts/event-schema.json`](../contracts/event-schema.json) und
[`contracts/sdk-compat.json`](../contracts/sdk-compat.json).

## Browser-Build

Das npm-Paket enthält ESM, CJS und IIFE. Der stabile Browser-Einstieg steht im
`browser`-Feld der Paket-Metadaten und zeigt auf
`dist/index.global.js`. Der IIFE-Build exportiert `MTracePlayerSDK` auf dem
globalen Objekt.
