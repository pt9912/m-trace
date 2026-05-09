# @npm9912/stream-analyzer

Stream-Manifest-Analyzer für die m-trace-Toolchain. Lädt HLS- und
DASH-Manifeste, klassifiziert sie und liefert ein stabiles JSON-
Ergebnis mit Findings — direkt aus der Bibliothek, aus der m-trace-
API oder aus der CLI (`pnpm m-trace check <url>`).

Status: HLS produktiv seit `0.3.0` (Master + Media + URL-Loader);
DASH-MPD produktiv seit `0.9.0` Tranche 3 (RAK-58 / NF-12) für VOD-
und einfache Live-MPDs.

## Installation

```bash
pnpm add @npm9912/stream-analyzer
```

## Schnellstart

Der Public-Entry-Point dispatcht intern auf HLS oder DASH anhand des
Body-Anfangs (`#EXTM3U` → HLS, `<?xml`/`<MPD` → DASH). Aufrufer
müssen das Format nicht angeben:

```ts
// Generischer Name (ab 0.9.0): analyzeManifest
import { analyzeManifest } from "@npm9912/stream-analyzer";

const result = await analyzeManifest({ kind: "text", text: manifest });
if (result.status === "ok") {
  console.log(result.analyzerKind, result.playlistType, result.findings);
} else {
  console.error(result.code, result.message);
}
```

`analyzeHlsManifest` bleibt als Backward-Kompat-Alias erhalten:

```ts
import { analyzeHlsManifest } from "@npm9912/stream-analyzer";
// funktional identisch zu analyzeManifest, dispatcht ebenfalls
// HLS/DASH; der Name spiegelt nur die historische 0.3.0-Public-API.
const result = await analyzeHlsManifest({ kind: "text", text: manifest });
```

## DASH-Eingabeform (ab `0.9.0`)

Das Result-Schema unterscheidet HLS- und DASH-Pfad über das
Diskriminator-Paar `analyzerKind` + `playlistType`:

| Pfad | `analyzerKind` | `playlistType`                    | `details`                 |
| --- | --- | --- | --- |
| HLS Master  | `"hls"`  | `"master"`                                       | `MasterPlaylistDetails`   |
| HLS Media   | `"hls"`  | `"media"`                                        | `MediaPlaylistDetails`    |
| HLS unklar  | `"hls"`  | `"unknown"`                                      | `null`                    |
| DASH        | `"dash"` | `"dash"`                                         | `DashManifestDetails`     |

`DashManifestDetails` trägt `profiles` / `type` (`static` / `dynamic`)
/ `live` (Boolean) / `mediaPresentationDuration` /
`minimumUpdatePeriod` / `availabilityStartTime` / `periodCount` plus
`adaptationSets[]` mit Mindest-Feldern pro `Representation`:
`bandwidth`, `width`/`height`, `codecs`, `mimeType`, `frameRate`,
`audioSamplingRate` (bei Audio-Renditions). `summary.itemCount`
zählt die Gesamtzahl Representations über alle Periods/AdaptationSets.

Beispiel (DASH-VOD-Result, gekürzt — vollständige Form in
[`spec/contract-fixtures/analyzer/success-dash-vod.json`](../../spec/contract-fixtures/analyzer/success-dash-vod.json)):

```json
{
  "status": "ok",
  "analyzerKind": "dash",
  "playlistType": "dash",
  "summary": { "itemCount": 3 },
  "details": {
    "profiles": "urn:mpeg:dash:profile:isoff-on-demand:2011",
    "type": "static",
    "live": false,
    "mediaPresentationDuration": "PT10M0S",
    "periodCount": 1,
    "adaptationSets": [
      {
        "id": "1",
        "mimeType": "video/mp4",
        "contentType": "video",
        "representations": [
          {
            "id": "v1",
            "bandwidth": 1280000,
            "width": 1280,
            "height": 720,
            "frameRate": "30",
            "codecs": "avc1.4d401e",
            "mimeType": "video/mp4"
          }
        ]
      }
    ]
  }
}
```

## CLI-Dispatcher

`pnpm m-trace check <url-or-file>` dispatcht automatisch auf HLS
oder DASH; das Format wird am Body-Anfang erkannt, nicht an der
Datei-Endung oder am Content-Type. Der Loader nimmt
`application/vnd.apple.mpegurl`, `application/x-mpegurl`,
`application/dash+xml`, `application/xml` und `text/xml`
gleichermaßen an (plus `text/plain` als Fallback).

```bash
# HLS via URL
pnpm m-trace check https://cdn.example.test/manifest.m3u8

# DASH via URL
pnpm m-trace check https://cdn.example.test/manifest.mpd

# Lokale Datei (HLS oder DASH)
pnpm m-trace check ./fixtures/master.m3u8
pnpm m-trace check ./fixtures/vod.mpd
```

Der CLI-Code selbst entscheidet nichts — die Dispatch-Logik lebt im
gemeinsamen Detector im `analyzeManifest`-Pfad. Eingaben, die weder
HLS noch DASH sind (HTML-Bodies, JSON, leerer Text), werden mit
`code: "manifest_not_supported"` abgewiesen (HTTP 422 in der API,
Exit 1 in der CLI). `manifest_not_hls` bleibt nur erhalten, wenn der
Detector den Input als HLS klassifiziert hat, der HLS-Parser ihn
dann aber selbst ablehnt (defektes `#EXTM3U`-Manifest).

## Scope

- ✅ HLS Master- und Media-Playlist (RAK-22..RAK-26 seit `0.3.0`).
- ✅ HLS via URL mit SSRF-Schutz, Größenlimit, Redirects (RAK-27 / RAK-28).
- ✅ DASH-MPD VOD und einfache Live-MPDs (RAK-58 / NF-12 seit `0.9.0`).
- ✅ DASH via URL — Loader generalisiert, SSRF-Schutz unverändert.
- 🟡 CMAF-Analyse im Stream-Analyzer-Scope (NF-13, RAK-60..RAK-64) —
  in `0.10.0` als additives `details.cmaf`-Signalmodell unter den
  bestehenden HLS-/DASH-Detail-Objekten; manifestbasierte HLS-/DASH-
  Signale plus begrenzte binäre CMAF-Konformitätsprüfung ausgewählter
  Init-/Media-Segmente (Brand-Allowlist `cmfc`/`cmf2`/`cmfs`/`cmff`,
  Defaults `maxSegmentBytes=2_000_000`, `maxBinarySegments=6`). **Kein
  neuer `analyzerKind`** — `details.cmaf` lebt unter
  `MasterPlaylistDetails.cmaf?` / `MediaPlaylistDetails.cmaf?` /
  `DashManifestDetails.cmaf?`.
- ⬜ DASH SegmentTemplate-Edge-Cases (`$Time$`-Variablen,
  `availabilityStartTime`-Drift) — Folge-Plan (Out of Scope laut
  `plan-0.9.0.md` §0.3).
- ⬜ Low-Latency-CMAF (`#EXT-X-PART`, chunked CMAF), vollständige
  Segmentset-Abdeckung, Codec-Decoding und Player-SDK-CMAF-Support —
  bewusst Folge-Scope nach `0.10.0`.

Vollständige Doku: [`docs/user/stream-analyzer.md`](../../docs/user/stream-analyzer.md).
