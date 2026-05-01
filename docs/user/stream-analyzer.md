# Stream Analyzer

`@npm9912/stream-analyzer` ist die HLS-Manifestanalyse der m-trace-Toolchain.
Das Paket liefert eine Bibliotheks-API für Backend-Integration (`apps/api`),
eine CLI (ab Tranche 7) und ein stabiles JSON-Ergebnisformat.

Bezug: [`spec/lastenheft.md`](../../spec/lastenheft.md) §7.7 (RAK-22..RAK-28,
F-68..F-81), [`docs/planning/plan-0.3.0.md`](../planning/plan-0.3.0.md),
[`spec/architecture.md`](../../spec/architecture.md) §5/§8 (Hexagon-Port).

## 1. Status (0.3.0 Tranche 2)

- ✅ Public API, Result-/Fehlerschema, Versionssynchronizität, Build-Pipeline
  und Coverage-Gate ≥ 90 % stehen.
- ✅ Manifest-Klassifikator: erkennt Master- und Media-Playlists anhand der
  Tags, lehnt Nicht-HLS und leere Manifeste mit `manifest_not_hls` ab,
  markiert ambige Mischformen als Master-Variante mit Warning-Finding.
- ✅ URL-Loader: HTTP/HTTPS, Timeout, Größenlimit, manuelles Redirect-
  Handling und SSRF-Schutzregeln (siehe §6).
- ⬜ Master-Detail-Auswertung (Variants/Renditions) — Tranche 3.
- ⬜ Media-Detail-Auswertung (Segmente, Findings, Live-Latenz) — Tranche 4.
- ⬜ Stabilisiertes JSON-Schema mit typspezifischen `details` — Tranche 5.
- ⬜ API-Anbindung über den Driven-Port `StreamAnalyzer.AnalyzeManifest` —
  Tranche 6.
- ⬜ CLI `pnpm m-trace check <url>` — Tranche 7.

Tranche 2 liefert die Klassifikation und das robuste Lade-Subsystem; jede
weitere Tranche erweitert die Result-Details additiv.

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
  `AnalyzeOptions`, `FetchOptions`, `AnalysisFinding`, `FindingLevel`,
  `AnalysisInputMetadata`, `AnalysisResult`, `AnalysisSummary`,
  `PlaylistType`, `AnalysisErrorCode`, `AnalysisErrorResult`.

### 2.1 Eingabeformen

```ts
type ManifestInput =
  | { kind: "text"; text: string; baseUrl?: string }
  | { kind: "url"; url: string };
```

- `text`: Manifestinhalt direkt; optionale `baseUrl` löst relative Variant-/
  Segment-URIs ab Tranche 3 auf.
- `url`: Quelle, die der Analyzer selbst lädt. `analyzeHlsManifest` setzt
  `input.baseUrl` automatisch auf die finale URL nach allen Redirects, damit
  Tranche 3/4 relative URIs konsistent auflösen kann.

`AnalyzeOptions.fetch` justiert das URL-Laden; alle Felder optional:

```ts
type FetchOptions = {
  timeoutMs?: number;    // Default: 10_000
  maxBytes?: number;     // Default: 5_000_000
  maxRedirects?: number; // Default: 5
};
```

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

Beispiel (Media-Playlist nach Tranche 2):

```json
{
  "status": "ok",
  "analyzerVersion": "0.3.0",
  "input": {
    "source": "url",
    "url": "https://cdn.example.test/manifest.m3u8",
    "baseUrl": "https://cdn.example.test/manifest.m3u8"
  },
  "playlistType": "media",
  "summary": { "itemCount": 0 },
  "findings": [
    {
      "code": "details_pending",
      "level": "info",
      "message": "stream-analyzer 0.3.0 Tranche 2: Klassifikation abgeschlossen, typspezifische Detail-Auswertung folgt in Tranche 3/4."
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
Diskriminator-Feld verlassen. Beispiel (URL gegen lokale Adresse):

```json
{
  "status": "error",
  "analyzerVersion": "0.3.0",
  "code": "fetch_blocked",
  "message": "Aufgelöste IP-Adresse verletzt SSRF-Sperrliste: ip_blocked.",
  "details": { "host": "internal.example.test", "address": "10.0.0.5", "family": 4 }
}
```

## 3. Scope

| Bereich       | 0.3.0   | Bemerkung                                                     |
| ------------- | ------- | ------------------------------------------------------------- |
| HLS Master    | ⬜ Plan | Tranche 3 implementiert Variants/Renditions (RAK-23, F-76).   |
| HLS Media     | ⬜ Plan | Tranche 4 implementiert Segmente/Findings (RAK-24/25, F-70..). |
| HLS via URL   | ⬜ Plan | Tranche 2 inkl. Timeout, Größenlimit, SSRF-Schutz.            |
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
`AnalyzerVersion = "noop"`. Das Domain-Modell reicht analyzer-spezifische
Detail-Strukturen als vorcodiertes JSON via `EncodedDetails []byte` weiter,
damit `apps/api/hexagon/domain` kein HLS-Detail-Schema vorgibt.

Tranche 6 entscheidet den Integrationsmodus formal (plan-0.3.0 §7);
favorisiert ist ein interner Analyzer-HTTP-Service, damit das distroless-
Go-API-Image keinen Node-Runtime mitbringen muss. Diese Doku-Sektion
beschreibt den Plan, nicht den Tranche-1-Lieferstand.

## 6. URL-Loader und SSRF-Schutz

Tranche 2 liefert den Loader unter `internal/loader/`. Eingabe-URLs gehen
durch eine harte Schutzkette, jeder Eintrag ist getestet:

| Schutzregel             | Verhalten                                                                 |
| ----------------------- | ------------------------------------------------------------------------- |
| Schema                  | Nur `http:` und `https:`; alles andere → `fetch_blocked`.                |
| Credentials             | `https://user:pass@…` und `https://user@…` werden abgelehnt.             |
| Hostname                | Leerer Hostname → `fetch_blocked`.                                       |
| DNS-Auflösung           | Schon ein Lookup-Eintrag in einem Sperrbereich blockt den ganzen Hop.    |
| IPv4-Sperrbereiche      | `0/8`, `10/8`, `100.64/10`, `127/8`, `169.254/16`, `172.16/12`, `192.0/24`, `192.0.2/24`, `192.88.99/24`, `192.168/16`, `198.18/15`, `198.51.100/24`, `203.0.113/24`, `224/4`, `240/4`. |
| IPv6-Sperrbereiche      | `::/128`, `::1/128`, `::ffff:0:0/96`, `64:ff9b::/96`, `100::/64`, `2001:db8::/32`, `fc00::/7`, `fe80::/10`, `ff00::/8`. |
| Timeout                 | `AbortController` schießt jeden Hop nach `timeoutMs` ab → `fetch_failed`. |
| Größenlimit             | Body-Stream wird mitgezählt; `> maxBytes` → `manifest_too_large`. Auch nach Redirect.|
| Redirect-Handling       | `redirect: "manual"`; jeder Hop durchläuft die volle Schutzkette erneut. |
| Redirect-Limit          | `> maxRedirects` Hops → `fetch_blocked`.                                 |
| Status-Codes            | Nicht-2xx → `fetch_failed`.                                              |
| Content-Type            | `application/vnd.apple.mpegurl`, `application/x-mpegurl`, `audio/mpegurl`, `text/plain`. Fehlt der Header, wird als Text-Fallback akzeptiert; alles andere → `fetch_failed`. |

### DNS-Rebinding-Entscheidung

Der Loader löst den Host genau **einmal** auf, prüft jeden zurückgegebenen
Eintrag gegen die Sperrlisten und übergibt die URL anschließend an den
Runtime-Adapter, der sie regulär via `fetch` zustellt. Ein zweiter
DNS-Lookup zwischen Validierung und TCP-Connect ist auf Anwendungsebene
nicht ausgeschlossen — eine sichere Egress-Topologie verlangt zusätzlich
eine Netzwerk-/Firewall-Schicht, die direkt gegen IP-Bereiche filtert.
Diese Architekturgrenze ist bewusst, dokumentiert und in Tests gepinnt
(`tests/loader-fetch.test.ts` „DNS-Rebinding-Entscheidung").

## 7. Lokale Entwicklung

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
