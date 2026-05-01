# @npm9912/stream-analyzer

HLS Stream Analyzer für die m-trace-Toolchain. Lädt HLS-Manifeste,
klassifiziert Master- vs. Media-Playlist und liefert ein stabiles
JSON-Ergebnis mit Findings — direkt aus der Bibliothek, aus der
m-trace-API oder aus der CLI (`pnpm m-trace check <url>`, ab 0.3.0
Tranche 7).

Status: **0.3.0 Tranche 1 — Skelett**. Public API und Result-Schema
sind festgelegt; Parser-Implementierung kommt in den Folgetranchen.

## Installation

```bash
pnpm add @npm9912/stream-analyzer
```

## Schnellstart

```ts
import { analyzeHlsManifest } from "@npm9912/stream-analyzer";

const result = await analyzeHlsManifest({ kind: "text", text: manifest });
if (result.status === "ok") {
  console.log(result.playlistType, result.findings);
} else {
  console.error(result.code, result.message);
}
```

## Scope

- ✅ HLS Master- und Media-Playlist (RAK-22..RAK-26).
- ⬜ DASH/CMAF — out of scope für 0.3.0 (F-73 Folge-Release).

Vollständige Doku: [`docs/user/stream-analyzer.md`](../../docs/user/stream-analyzer.md).
