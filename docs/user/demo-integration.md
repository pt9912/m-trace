# Demo-Integration

> **Status**: `0.2.0`-Arbeitsstand. Dieses Dokument beschreibt die
> Dashboard-Route `/demo` als lokale Beispielintegration für
> `@npm9912/player-sdk`, hls.js und die m-trace API.

## Zweck

`/demo` ist die Referenzintegration für den lokalen MVP-Stack. Die Route lädt
den MediaMTX-HLS-Teststream, bindet hls.js an ein `<video>`-Element und sendet
Playback-Events über das Player-SDK an `POST /api/playback-events`.

## Laufzeitpfad

```text
stream-generator
  -> mediamtx HLS
  -> apps/dashboard /demo
  -> hls.js
  -> @npm9912/player-sdk
  -> apps/api /api/playback-events
  -> Dashboard Sessions/Events
```

## Lokaler Start

```bash
make dev
```

Danach ist die Demo unter `http://localhost:5173/demo` erreichbar.

Für einen reproduzierbaren Session-Namen kann die Route mit Query-Parametern
gestartet werden:

```text
http://localhost:5173/demo?session_id=demo-docs-1&autostart=1
```

| Parameter | Bedeutung |
|---|---|
| `session_id` | optionale Session-ID; ohne Wert generiert das SDK eine Session-ID |
| `autostart=1` | startet Playback beim Laden der Route |

## SDK-Konfiguration der Demo

Die Route nutzt diese SDK-Konfiguration:

```ts
const tracker = createTracker({
  endpoint: collectorEndpoint,
  token: "demo-token",
  projectId: "demo",
  sessionId: new URLSearchParams(window.location.search).get("session_id") ?? undefined,
  batchSize: 5,
  flushIntervalMs: 2000
});
```

| Feld | Demo-Wert |
|---|---|
| API-URL | `PUBLIC_PLAYER_COLLECTOR_ENDPOINT` oder `http://localhost:8080/api/playback-events` |
| Token | `demo-token` |
| Project | `demo` |
| HLS-URL | `PUBLIC_HLS_URL` oder `http://localhost:8888/teststream/index.m3u8` |
| Batch-Größe | `5` |
| Flush-Intervall | `2000 ms` |

## hls.js-Anbindung

Wenn hls.js im Browser unterstützt wird, nutzt die Demo diesen Pfad:

```ts
const hls = new Hls();
hls.loadSource(hlsUrl);
hls.attachMedia(video);
const adapter = attachHlsJs(video, hls, tracker);
```

Der Adapter meldet hls.js- und Video-Events an den Tracker. Beim Stoppen oder
Unmount der Route entfernt `adapter.destroy()` die Listener; anschließend
schließt `tracker.destroy()` die Session mit `session_ended`.

Wenn hls.js nicht unterstützt wird, fällt die Route auf native HLS-Wiedergabe
zurück, sofern der Browser `application/vnd.apple.mpegurl` abspielen kann.
Dieser Safari-Pfad ist bewusst als eingeschränkt dokumentiert, weil native HLS
weniger Ereignistiefe bietet.

## Erwartete Events

Im hls.js-Pfad sind abhängig vom Playback-Verlauf insbesondere diese Events
zu erwarten:

| Event | Auslöser |
|---|---|
| `manifest_loaded` | hls.js meldet geladenes Manifest |
| `segment_loaded` | hls.js meldet geladenes Fragment |
| `bitrate_switch` | hls.js meldet Level-Wechsel |
| `playback_started` | Video startet bzw. erste Daten sind verfügbar |
| `startup_time_measured` | Startup-Zeit wurde gemessen |
| `rebuffer_started` | Video feuert `waiting` |
| `rebuffer_ended` | Video spielt nach Rebuffer weiter |
| `playback_error` | hls.js meldet Fehler |
| `session_ended` | Demo wird gestoppt oder Route wird verlassen |

## Verifikation

```bash
make browser-e2e
```

Das Browser-E2E-Gate startet API, MediaMTX, FFmpeg-Teststream und Dashboard im
Container und prüft die Demo-Route in Chromium und Firefox.

Manuell kann die erzeugte Session über das Dashboard geprüft werden:

1. `http://localhost:5173/demo?session_id=demo-docs-1&autostart=1` öffnen.
2. Nach einigen Sekunden `Stop` klicken.
3. `http://localhost:5173/sessions/demo-docs-1` öffnen.
4. Prüfen, dass Playback-Events und `session_ended` sichtbar sind.

Alternativ über die API:

```bash
curl http://localhost:8080/api/stream-sessions/demo-docs-1/events
```
