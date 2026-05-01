# Stream Analyzer

`@npm9912/stream-analyzer` ist die HLS-Manifestanalyse der m-trace-Toolchain.
Das Paket liefert eine Bibliotheks-API für Backend-Integration (`apps/api`),
eine CLI (ab Tranche 7) und ein stabiles JSON-Ergebnisformat.

Bezug: [`spec/lastenheft.md`](../../spec/lastenheft.md) §7.7 (RAK-22..RAK-28,
F-68..F-81), [`docs/planning/plan-0.3.0.md`](../planning/plan-0.3.0.md),
[`spec/architecture.md`](../../spec/architecture.md) §5/§8 (Hexagon-Port).

## 1. Status (0.3.0 Tranche 1)

- ✅ Public API, Result-/Fehlerschema, Versionssynchronizität, Build-Pipeline
  und Coverage-Gate ≥ 90 % stehen.
- ⬜ HLS-Parser (Master/Media-Playlist, Segment-Findings) — Tranche 2/3/4.
- ⬜ URL-Loader inkl. SSRF-Schutz — Tranche 2.
- ⬜ Stabilisiertes JSON-Schema mit typspezifischen `details` — Tranche 5.
- ⬜ API-Anbindung über den Driven-Port `StreamAnalyzer.AnalyzeManifest` —
  Tranche 6.
- ⬜ CLI `pnpm m-trace check <url>` — Tranche 7.

Tranche 1 liefert ein lauffähiges Skelett: `analyzeHlsManifest` gibt für
gültige Text-Inputs ein Erfolgsergebnis mit einem `not_implemented`-Finding
zurück; URL-Inputs werden mit Code `fetch_blocked` abgelehnt, bis Tranche 2
die Lade-Politik installiert.

## 2. Public API

```ts
import { analyzeHlsManifest, STREAM_ANALYZER_VERSION } from "@npm9912/stream-analyzer";

const result = await analyzeHlsManifest({ kind: "text", text: manifest });
if (result.status === "ok") {
  console.log(result.playlistType, result.findings);
} else {
  console.error(result.code, result.message);
}
```

Exportierte Symbole (Snapshot in
`packages/stream-analyzer/scripts/public-api.snapshot.txt`):

- `analyzeHlsManifest(input, options?) → Promise<AnalysisResult | AnalysisErrorResult>`
- `AnalysisError` — Fehlerklasse für Adapter; Konsumenten nutzen normalerweise das Result.
- `STREAM_ANALYZER_NAME`, `STREAM_ANALYZER_VERSION` — aus `package.json` abgeleitet.
- Typen: `ManifestInput` (`ManifestTextInput | ManifestUrlInput`),
  `AnalyzeOptions`, `AnalysisFinding`, `FindingLevel`, `AnalysisInputMetadata`,
  `AnalysisResult`, `AnalysisSummary`, `PlaylistType`, `AnalysisErrorCode`,
  `AnalysisErrorResult`.

### 2.1 Eingabeformen

```ts
type ManifestInput =
  | { kind: "text"; text: string; baseUrl?: string }
  | { kind: "url"; url: string };
```

- `text`: Manifestinhalt direkt; optionale `baseUrl` löst relative Variant-/
  Segment-URIs ab Tranche 3 auf.
- `url`: Quelle, die der Analyzer selbst lädt (Tranche 2). Bis dahin liefert
  der Aufruf einen `fetch_blocked`-Fehler.

### 2.2 Erfolgs-Ergebnis

```ts
{
  status: "ok",
  analyzerVersion: "0.3.0",
  input: { source: "text" | "url", url?: string, baseUrl?: string },
  playlistType: "master" | "media" | "unknown",
  summary: { itemCount: number },
  findings: Array<{ code: string, level: "info" | "warning" | "error", message: string }>,
  details: Record<string, unknown> | null
}
```

Beispiel (Tranche-1-Stand):

```json
{
  "status": "ok",
  "analyzerVersion": "0.3.0",
  "input": { "source": "text" },
  "playlistType": "unknown",
  "summary": { "itemCount": 0 },
  "findings": [
    {
      "code": "not_implemented",
      "level": "info",
      "message": "stream-analyzer 0.3.0 Tranche 1: HLS-Parser ist noch nicht angeschlossen; Result-Schema ist stabil."
    }
  ],
  "details": null
}
```

### 2.3 Fehler-Ergebnis

```ts
{
  status: "error",
  analyzerVersion: "0.3.0",
  code: "invalid_input" | "manifest_not_hls" | "fetch_failed" | "fetch_blocked" | "manifest_too_large" | "internal_error",
  message: string,
  details?: Record<string, unknown>
}
```

`status` trennt Erfolg und Fehler statisch — Konsumenten dürfen sich auf das
Diskriminator-Feld verlassen. Beispiel:

```json
{
  "status": "error",
  "analyzerVersion": "0.3.0",
  "code": "fetch_blocked",
  "message": "URL-Laden wird erst in 0.3.0 Tranche 2 freigeschaltet (plan-0.3.0 §3).",
  "details": { "url": "https://example.test/manifest.m3u8" }
}
```

## 3. Scope

| Bereich       | 0.3.0   | Bemerkung                                                     |
| ------------- | ------- | ------------------------------------------------------------- |
| HLS Master    | ✅ Plan | Tranche 3 implementiert Variants/Renditions (RAK-23, F-76).   |
| HLS Media     | ✅ Plan | Tranche 4 implementiert Segmente/Findings (RAK-24/25, F-70..). |
| HLS via URL   | ✅ Plan | Tranche 2 inkl. Timeout, Größenlimit, SSRF-Schutz.            |
| DASH/CMAF     | ❌      | Out of scope — F-73 als eigener Analyzer-Typ in Folge-Release.|
| SRT           | ❌      | Eigener Bereich (`0.6.0`).                                    |

## 4. Stabilitätsregel

Das Result-Schema ist additiv erweiterbar:

- Neue Felder dürfen jederzeit ergänzt werden.
- Bestehende Felder bleiben in Form und Typ stabil.
- Breaking Changes erfordern Eintrag in `CHANGELOG.md` und Update von
  `docs/user/stream-analyzer.md` und `docs/planning/plan-0.3.0.md`.

Die `AnalyzerVersion` aus `package.json` wird in jedem Result mitgeliefert,
damit Konsumenten Schema-Drift erkennen können.

## 5. Backend-Anbindung

Ab Tranche 6 ruft `apps/api` den Analyzer über den Driven-Port
`hexagon/port/driven.StreamAnalyzer.AnalyzeManifest(ctx, request) (result, error)`
auf. Tranche 1 hat den Port bereits um die Zielsignatur erweitert; bis zur
Tranche-6-Verdrahtung trägt `NoopStreamAnalyzer` einen leeren Slot mit
`AnalyzerVersion = "noop"`.

Bevorzugter Integrationsmodus für 0.3.0 ist ein interner Analyzer-HTTP-Service,
damit das distroless-Go-API-Image keinen Node-Runtime mitbringen muss
(plan-0.3.0 §7).

## 6. Lokale Entwicklung

```bash
# Tests
pnpm --filter @npm9912/stream-analyzer run test

# Coverage (Schwelle 90 % auf src/**)
pnpm --filter @npm9912/stream-analyzer run test:coverage

# Lint (tsc + Boundary-Check + Public-API-Snapshot)
pnpm --filter @npm9912/stream-analyzer run lint

# Build (ESM + CJS + d.ts)
pnpm --filter @npm9912/stream-analyzer run build
```

Root-Aggregat: `make test`, `make lint`, `make coverage-gate`, `make build`
beziehen das Paket über `pnpm -r --if-present` automatisch ein.
