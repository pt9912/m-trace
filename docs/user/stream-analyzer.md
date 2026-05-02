# Stream Analyzer

`@npm9912/stream-analyzer` ist die HLS-Manifestanalyse der m-trace-Toolchain.
Das Paket liefert eine Bibliotheks-API für Backend-Integration (`apps/api`),
eine CLI (ab Tranche 7) und ein stabiles JSON-Ergebnisformat.

Bezug: [`spec/lastenheft.md`](../../spec/lastenheft.md) §7.7 (RAK-22..RAK-28,
F-68..F-81), [`docs/planning/done/plan-0.3.0.md`](../planning/done/plan-0.3.0.md),
[`spec/architecture.md`](../../spec/architecture.md) §5/§8 (Hexagon-Port).

## 1. Status (0.3.0 — veröffentlicht)

- ✅ Public API, Result-/Fehlerschema, Versionssynchronizität, Build-Pipeline
  und Coverage-Gate ≥ 90 % stehen.
- ✅ Manifest-Klassifikator: erkennt Master- und Media-Playlists anhand der
  Tags, lehnt Nicht-HLS und leere Manifeste mit `manifest_not_hls` ab,
  markiert ambige Mischformen als Master-Variante mit Warning-Finding.
- ✅ URL-Loader: HTTP/HTTPS, Timeout, Größenlimit, manuelles Redirect-
  Handling und SSRF-Schutzregeln (siehe §6).
- ✅ Master-Detail-Auswertung: Variants (`#EXT-X-STREAM-INF`) mit
  Bandbreite/Resolution/Codecs/Frame-Rate/Group-Refs, Renditions
  (`#EXT-X-MEDIA`) mit Typ/GroupId/Name/Lang/URI/Flags, Group-Cross-Check,
  optionale Base-URL-Auflösung als `resolvedUri`.
- ✅ Media-Detail-Auswertung: Segmente aus `#EXTINF`, Aggregate (Anzahl,
  Min/Max/Mittel/Total), TARGETDURATION-Verletzung, Outlier-Erkennung,
  Live-/VOD-Klassifikation und 3×-Latenzschätzung — siehe §7.
- ✅ JSON-Ergebnisformat: `AnalysisResult` als diskriminierte Union per
  `playlistType`, `analyzerKind: "hls"` als Erweiterungspfad für
  DASH/CMAF, deterministische Serialisierung, Stabilitätsregel als
  operativer Vertrag — siehe §4.
- ✅ API-Anbindung: `POST /api/analyze` reicht den Aufruf an den
  internen `analyzer-service` (Node-HTTP-Wrapper) weiter; Go-API
  bleibt distroless-static. Vollständig in `docker-compose.yml`
  verdrahtet, Smoke-Test über `make smoke-analyzer` — siehe §5.
- ✅ CLI `pnpm m-trace check <url-or-file>`: stdout-JSON, Exit-Codes
  0/1/2, Datei- und URL-Input, SSRF-Schutz aus dem Loader greift
  unverändert — siehe §9.

Tranche 5 sperrt das JSON-Format. Konsumenten erkennen Erfolg/Fehler an
`status`, schalten auf `playlistType` zur Auswahl der Detail-Form und
filtern bei Bedarf weiter über `analyzerKind` (heute nur `"hls"`).

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
  `AnalysisInputMetadata`, `AnalysisResult` (Union aus
  `MasterAnalysisResult | MediaAnalysisResult | UnknownAnalysisResult`,
  diskriminiert per `playlistType`), `AnalysisSummary`, `AnalyzerKind`,
  `BaseAnalysisResult`, `PlaylistType`, `AnalysisErrorCode`,
  `AnalysisErrorResult`, `MasterPlaylistDetails`, `MasterRendition`,
  `MasterVariant`, `MediaPlaylistDetails`, `MediaSegment`,
  `MediaSegmentSummary`.

Konsumenten brauchen keine Casts:

```ts
const result = await analyzeHlsManifest({ kind: "url", url });
if (result.status === "error") {
  console.error(result.code, result.details);
  return;
}
if (result.playlistType === "master") {
  // result.details: MasterPlaylistDetails (TypeScript narrowed)
  console.log(result.details.variants.length, "variants");
} else if (result.playlistType === "media") {
  // result.details: MediaPlaylistDetails
  console.log("live:", result.details.live, "segments:", result.details.segments.length);
} else {
  // result.playlistType === "unknown" → details: null
}
```

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
  analyzerKind: "hls",
  input: { source: "text" | "url", url?: string, baseUrl?: string },
  playlistType: "master" | "media" | "unknown",
  summary: { itemCount: number },
  findings: Array<{ code: string, level: "info" | "warning" | "error", message: string }>,
  // details ist diskriminiert per playlistType:
  details: MasterPlaylistDetails | MediaPlaylistDetails | null
}
```

Beispiel (Master-Playlist):

```json
{
  "status": "ok",
  "analyzerVersion": "0.3.0",
  "analyzerKind": "hls",
  "input": { "source": "text", "baseUrl": "https://cdn.example.test/" },
  "playlistType": "master",
  "summary": { "itemCount": 3 },
  "findings": [],
  "details": {
    "variants": [
      {
        "bandwidth": 1280000,
        "resolution": { "width": 720, "height": 480 },
        "codecs": ["avc1.4d401e", "mp4a.40.2"],
        "audio": "aud-en",
        "uri": "video/720p/main.m3u8",
        "resolvedUri": "https://cdn.example.test/video/720p/main.m3u8"
      }
    ],
    "renditions": [
      {
        "type": "AUDIO",
        "groupId": "aud-en",
        "name": "English",
        "language": "en",
        "default": true,
        "autoselect": true,
        "uri": "audio/en/main.m3u8",
        "resolvedUri": "https://cdn.example.test/audio/en/main.m3u8"
      }
    ]
  }
}
```

Beispiel (Live-Media-Playlist):

```json
{
  "status": "ok",
  "analyzerVersion": "0.3.0",
  "analyzerKind": "hls",
  "input": {
    "source": "url",
    "url": "https://cdn.example.test/live/manifest.m3u8",
    "baseUrl": "https://cdn.example.test/live/manifest.m3u8"
  },
  "playlistType": "media",
  "summary": { "itemCount": 4 },
  "findings": [],
  "details": {
    "targetDuration": 4,
    "mediaSequence": 8423,
    "endList": false,
    "live": true,
    "liveLatencyEstimateSeconds": 12,
    "segments": [
      { "uri": "seg-8423.ts", "duration": 3.84, "sequenceNumber": 8423,
        "resolvedUri": "https://cdn.example.test/live/seg-8423.ts" }
    ],
    "summary": {
      "count": 4,
      "averageDuration": 3.86,
      "minDuration": 3.84,
      "maxDuration": 3.92,
      "totalDuration": 15.44
    }
  }
}
```

### 2.3 Fehler-Ergebnis

```ts
{
  status: "error",
  analyzerVersion: "0.3.0",
  analyzerKind: "hls",
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
  "analyzerKind": "hls",
  "code": "fetch_blocked",
  "message": "Aufgelöste IP-Adresse verletzt SSRF-Sperrliste: ip_blocked.",
  "details": { "hop": 0, "host": "internal.example.test", "address": "10.0.0.5", "family": 4 }
}
```

`details.hop` ist ein 0-basierter Zähler über die Redirect-Kette
(0 = ursprünglicher Request, 1 = erste Weiterleitung); er taucht in
allen `fetch_*`-Fehlern auf und ist hilfreich, um zu erkennen, in
welchem Hop ein SSRF-Block bzw. ein Statuscode-Problem auftrat.

## 3. Scope

| Bereich       | 0.3.0   | Bemerkung                                                     |
| ------------- | ------- | ------------------------------------------------------------- |
| HLS Master    | ✅       | Variants/Renditions, Group-Cross-Check, Base-URL-Auflösung.   |
| HLS Media     | ✅       | Segmente, Toleranzregel, Live/VOD, 3×-Latenz.                 |
| HLS via URL   | ✅       | Timeout, Größenlimit, SSRF-Schutz (siehe §6).                 |
| DASH/CMAF     | ❌       | Out of scope — F-73 als zusätzlicher `analyzerKind` möglich.   |
| SRT           | ❌       | Eigener Bereich (`0.6.0`).                                    |

## 4. Stabilitätsregel

Das Result-Schema ist additiv erweiterbar. Konsumenten dürfen sich auf
die folgenden Garantien verlassen, solange `analyzerVersion` Major und
Minor unverändert bleibt:

**Erlaubte additive Änderungen** (kein Major-Bump):

- Neue optionale Felder im Erfolgs- oder Fehler-Result.
- Neue optionale Felder in `details.*`-Sub-Strukturen.
- Neue Werte für `playlistType` (z. B. wenn HLS-Spec einen weiteren
  Typ einführt).
- Neue Werte für `analyzerKind` (z. B. `"dash"`, `"cmaf"` als
  zusätzliche Union-Member).
- Neue Finding-Codes oder Finding-Levels (Konsumenten dürfen
  Unbekannte ignorieren oder als Info behandeln).
- Neue `AnalysisErrorCode`-Werte.

**Breaking Changes** (verlangen Major-Bump + Eintrag in
`CHANGELOG.md` + Update von `docs/user/stream-analyzer.md` und
`docs/planning/done/plan-0.3.0.md`):

- Felder löschen oder umbenennen.
- Den Typ eines bestehenden Felds ändern (`number → string`,
  Optional zu Pflicht etc.).
- Bedeutung eines bestehenden Wertes ändern (z. B. `live` plötzlich
  `true` für VOD).
- Discriminator-Felder ändern (`status`, `playlistType`,
  `analyzerKind`).

**Sicherheitskritische Optionen**: `FetchOptions.allowPrivateNetworks`
(Default `false`) lockert die SSRF-IP-Sperrlisten. Library-Konsumenten
können das Flag explizit aktivieren, sollten es aber **nicht** in
Produktion verwenden. Die m-trace-API selbst nutzt das Flag nur,
wenn der Operator `ANALYZER_ALLOW_PRIVATE_NETWORKS` auf dem
analyzer-service-Container setzt; Aufrufer der API können es **nicht**
über den Request-Body anfordern (siehe §6).

**Diskriminatoren**:

- `result.status` trennt Erfolg (`"ok"`) und Fehler (`"error"`).
- `result.playlistType` (nur bei `status === "ok"`) trennt
  `MasterPlaylistDetails | MediaPlaylistDetails | null`.
- `result.analyzerKind` ist heute immer `"hls"`; künftige Werte
  zeigen Konsumenten an, dass sie das Result mit einem anderen
  Detail-Schema interpretieren müssen.

**Exhaustive Switches**: Konsumenten, die per `switch`-Anweisung
über `analyzerKind`, `playlistType`, `code` (`AnalysisErrorCode`)
oder Finding-Codes verzweigen, sollten einen `default`-Branch
behalten. Neue Werte werden additiv ergänzt — ein TypeScript-
Konsument ohne Default-Fall bricht beim Upgrade auf eine spätere
Minor-Version den eigenen Build (TS bemängelt nicht-erschöpfte
Cases). Eine Default-Behandlung als „Info ignorieren" oder
„unbekannt loggen" ist forward-kompatibel.

`analyzerVersion` aus `package.json` wird in jedem Result
mitgeliefert (Erfolg und Fehler). Sie reflektiert die Bake-Zeit
des Analyzer-Pakets, nicht die Laufzeit. Operativ aussagekräftig
ist deshalb der Vergleich gegen eine **erwartete** Version aus
Konfiguration oder Service-Discovery — nicht gegen den eigenen
`STREAM_ANALYZER_VERSION`-Import (der ist zwangsläufig identisch).

### 4.1 Serialisierungsgarantien

Diese Eigenschaften sind als Tests in
`tests/result-stability.test.ts` festgenagelt:

- **Deterministisch innerhalb eines Prozesses**: zwei Aufrufe mit
  identischer Eingabe liefern byte-identische
  `JSON.stringify(result)`-Strings — keine Map-Iterations-Drift,
  keine Zeitstempel im Result. Cross-Process-Determinismus ist nicht
  garantiert (V8 kann z. B. Hash-Randomization für nicht-string
  Keys nutzen); in der Praxis sind aber alle Result-Objekte aus
  Object-Literals mit fixer Schlüsselreihenfolge gebaut, sodass
  cross-Process-Stabilität de facto gilt.
- **Round-Trip-stabil**: `JSON.parse(JSON.stringify(result))` ist
  deep-equal zum Original. Damit kann das Result über Prozess- oder
  Service-Grenzen geschickt werden, ohne dass strukturierte
  Information verlorengeht.
- **Kein `undefined` im Output**: optionale Felder werden weggelassen,
  nicht als `undefined`-Property gesetzt. JSON-Konsumenten sehen das
  Feld entweder oder es fehlt — nie den `undefined`-Wert, der von
  `JSON.stringify` stillschweigend entfernt würde.
- **Nur finite Zahlen**: keine `NaN`, kein `Infinity` im Output.
  Eingabewerte, die so etwas erzeugen würden (z. B. unparseable
  EXTINF-Dauer), werden als Findings gemeldet und der Wert auf einen
  finiten Default normalisiert.

## 5. Backend-Anbindung

`apps/api` ruft den Analyzer über den Driven-Port
`hexagon/port/driven.StreamAnalyzer.AnalyzeManifest(ctx, request) (result, error)`
auf. Der produktive Adapter (`HTTPStreamAnalyzer`) postet das
Manifest-Eingabe-Schema des stream-analyzer-Pakets gegen den internen
`analyzer-service` (Node-HTTP-Wrapper unter `apps/analyzer-service`)
und mappt das `AnalyzeOutput`-JSON auf `domain.StreamAnalysisResult`
zurück. So bleibt das Go-API-Image distroless-static, ohne
Node-Runtime einzubetten.

`docker-compose.yml` startet den Service als `analyzer-service` und
setzt `ANALYZER_BASE_URL=http://analyzer-service:7000` an `apps/api`.
Ohne diese Env-Variable greift `apps/api` auf den Noop-Slot zurück
und meldet den Zustand im Startup-Log.

**API-Endpunkt**: `POST /api/analyze` (vollständig dokumentiert in
[`spec/backend-api-contract.md`](../../spec/backend-api-contract.md)
§3.6). Request- und Response-Schema spiegeln die Public API des
Pakets; Fehler werden auf eine Problem-Shape gemappt:

| HTTP | `code`                  | Anlass                                                                |
| ---- | ----------------------- | --------------------------------------------------------------------- |
| 400  | `invalid_request`       | API-Eingabe fehlerhaft (Pflichtfelder, Kind unbekannt).               |
| 400  | `invalid_json`          | Body kein gültiges JSON.                                              |
| 415  | `unsupported_media_type`| Content-Type nicht `application/json`.                                |
| 413  | `payload_too_large`     | Body über 1 MiB.                                                       |
| 400  | `invalid_input`         | Analyzer hat den Manifest-Input als formal ungültig zurückgewiesen.    |
| 400  | `fetch_blocked`         | SSRF-Schutz hat die URL abgelehnt (privat/loopback/Credentials).       |
| 422  | `manifest_not_hls`      | Geladenes Manifest ist kein HLS-Inhalt.                                |
| 502  | `fetch_failed`          | Analyzer konnte die URL nicht laden (Netzwerk/Status/Content-Type).    |
| 502  | `manifest_too_large`    | Manifest übersteigt das Loader-Größenlimit.                            |
| 502  | `internal_error`        | Unerwarteter Fehler im Analyzer-Stack.                                 |
| 502  | `analyzer_unavailable`  | Transportfehler API↔analyzer-service (kein Domain-Fehler).             |

Ein lokaler End-to-End-Smoke (`make smoke-analyzer`) startet den
Stack, prüft `/health` an Service und API, sendet einen Master-
Manifest-Text gegen `/api/analyze` und verifiziert zusätzlich, dass
ein RFC1918-URL-Input vom SSRF-Schutz korrekt mit 400 `fetch_blocked`
abgelehnt wird (kein 502 — der Analyzer hat den Aufruf bewusst
zurückgewiesen, nicht der Service ist ausgefallen).

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

### `allowPrivateNetworks` (opt-in)

Standardmäßig blockt der Loader die in der Tabelle gelisteten privaten
und Reservierungs-Bereiche. `FetchOptions.allowPrivateNetworks: true`
lockert **ausschließlich** die IPv4-/IPv6-Sperrlisten — Schema-Whitelist,
Credentials-Block und Größen-/Redirect-Regeln bleiben aktiv.

Anwendungsfälle:

- Compose-/Lab-Setups, in denen interne Streams (z. B. mediamtx auf
  einem RFC1918-Hostnamen) analysiert werden sollen.
- `apps/analyzer-service` liest die Env-Variable
  `ANALYZER_ALLOW_PRIVATE_NETWORKS=true|1|yes|on` und reicht das Flag
  pro Aufruf an den Loader weiter. Default in `apps/analyzer-service`
  ist aus; **das Compose-Lab-Setup** (`docker-compose.yml`) setzt das
  Flag explizit auf `true`, damit der Analyzer den eingebauten
  MediaMTX-Stream (`http://mediamtx:8888/teststream/index.m3u8` →
  RFC1918-IP der Docker-Bridge) erreichen kann. Produktions-
  Deployments setzen die Variable nicht.

**Service-Schalter ist ausschließlich die Env-Variable**: ein
Aufrufer kann das Flag nicht über den Request-Body setzen. Die
analyzer-service-Whitelist erlaubt im `fetch`-Sub-Block nur
`timeoutMs`, `maxBytes`, `maxRedirects`; `allowPrivateNetworks`
fällt heraus und wird nicht an den Loader durchgereicht. Das ist
absichtlich und in `apps/analyzer-service/tests/server.test.ts`
gepinnt.

Wer das Flag setzt, übernimmt explizit das Risiko: ein böswillig
gewähltes URL-Ziel kann dann auch interne Services (Metadata-
Endpoints, Admin-UIs, Datenbank-Ports) treffen. Ein zusätzlich vor-
geschalteter Egress-Filter bleibt dringend empfohlen.

### DNS-Rebinding-Entscheidung

Der Loader löst den Host genau **einmal** auf, prüft jeden zurückgegebenen
Eintrag gegen die Sperrlisten und übergibt die URL anschließend an den
Runtime-Adapter, der sie regulär via `fetch` zustellt. Ein zweiter
DNS-Lookup zwischen Validierung und TCP-Connect ist auf Anwendungsebene
nicht ausgeschlossen — eine sichere Egress-Topologie verlangt zusätzlich
eine Netzwerk-/Firewall-Schicht, die direkt gegen IP-Bereiche filtert.
Diese Architekturgrenze ist bewusst, dokumentiert und in Tests gepinnt
(`tests/loader-fetch.test.ts` „DNS-Rebinding-Entscheidung").

## 7. Media-Playlist-Auswertung

Tranche 4 setzt Segmente aus `#EXTINF`, dazu Aggregat-Statistiken und
Konformitätsprüfungen.

### 7.1 Segmentdaten

Pro Segment werden `uri`, `duration` (Sekunden, Float), optionaler
`title` aus `#EXTINF:duration,title`, `sequenceNumber` und — bei
gesetzter Base-URL — `resolvedUri` ausgegeben. Die Sequenznummer
startet bei `mediaSequence` (aus `#EXT-X-MEDIA-SEQUENCE`, sonst 0)
und steigt um 1 je Segment.

`details.summary` enthält `count`, `averageDuration`, `minDuration`,
`maxDuration` und `totalDuration` über alle Segmente — Werte in
Sekunden, gerechnet auf den geparsten Float-Dauern.

### 7.2 Toleranzregel und Findings

| Finding-Code                         | Level   | Bedeutung                                                                                       |
| ------------------------------------ | ------- | ----------------------------------------------------------------------------------------------- |
| `media_missing_targetduration`       | error   | RFC 8216 §4.3.3.1 macht das Tag verpflichtend; ohne es bleibt der Manifest-Konformitätscheck offen. |
| `media_malformed_targetduration`     | error   | TARGETDURATION ist nicht parseable; weitere Auswertung läuft trotzdem.                          |
| `media_malformed_mediasequence`      | warning | MEDIA-SEQUENCE nicht parseable; Fallback `mediaSequence = 0`.                                   |
| `segment_duration_exceeds_target`    | error   | `round(duration) > TARGETDURATION` — Spec-Verstoß (RFC 8216 §4.3.3.1). Deckt Upper-Drift ab.    |
| `segment_duration_outlier`           | warning | Segmentdauer ist < 50 % des Ankers. Anker = TARGETDURATION (Apple-HLS-Authoring-Guide); fehlt das Tag, Mean-Fallback. Letztes VOD-Segment ausgenommen. Lower-Bound-Check: zu lange Segmente sind über `segment_duration_exceeds_target` abgedeckt. |
| `media_encryption_present`           | info    | EXT-X-KEY mit aktiver Methode vorhanden; Schlüssel-/Decryption-Pfade werden nicht validiert.    |
| `media_init_segment_present`         | info    | EXT-X-MAP (fMP4-Init-Segment) vorhanden; Init-Segment-Konsistenz wird nicht geprüft.            |
| `media_discontinuity_present`        | info    | EXT-X-DISCONTINUITY vorhanden; Timeline-Continuity wird nicht ausgewertet.                       |
| `media_program_date_time_present`    | info    | EXT-X-PROGRAM-DATE-TIME vorhanden; Wall-Clock-Annotationen werden nicht ausgewertet.            |
| `segment_malformed_extinf`           | warning | EXTINF-Dauer nicht parseable; Segment wird mit `duration: 0` aufgenommen.                       |
| `segment_missing_uri`                | error   | EXTINF ohne folgende URI-Zeile.                                                                  |
| `segment_unexpected_uri`             | warning | Manifest-Zeile, die wie URI aussieht, ohne vorhergehendes EXTINF.                               |
| `segment_malformed_uri`              | warning | URI konnte nicht gegen die Base-URL aufgelöst werden.                                            |

### 7.3 Live-/VOD-Erkennung und Latenzschätzung

`details.endList` reflektiert `#EXT-X-ENDLIST`. `details.live = !endList`.
Wenn `live === true` und `targetDuration` bekannt ist, liefert der
Analyzer eine einfache Latenzschätzung nach Apples HLS-Authoring-
Empfehlung („3×-Regel"):

```
liveLatencyEstimateSeconds = 3 × targetDuration
```

Das ist eine **Authoring-orientierte Schätzung**, kein Mess- oder
Ende-zu-Ende-Wert. Reale Latenz hängt von Encoder, Player-Buffer,
CDN und Client-Distanz ab. Für VOD-Playlists ist das Feld undefiniert.

`details.playlistType` spiegelt `#EXT-X-PLAYLIST-TYPE` (`VOD` oder
`EVENT`), falls gesetzt; das Feld ist informativ und greift nicht in
die Live/VOD-Klassifikation ein (`endList` ist die maßgebliche Quelle).

## 8. Lokale Entwicklung

```bash
# Tests
pnpm --filter @npm9912/stream-analyzer run test

# Coverage (Schwelle 90 % auf src/**)
pnpm --filter @npm9912/stream-analyzer run test:coverage

# Lint (tsc + Boundary-Check + Public-API-Snapshot)
pnpm --filter @npm9912/stream-analyzer run lint

# Build (ESM + CJS + d.ts inkl. CLI-Bundle)
pnpm --filter @npm9912/stream-analyzer run build
```

Root-Aggregat: `make test`, `make lint`, `make coverage-gate`, `make build`
beziehen das Paket über `pnpm -r --if-present` automatisch ein.

## 9. CLI `m-trace check`

Der Lastenheft-Aufruf `pnpm m-trace check <url-or-file>` analysiert ein
HLS-Manifest und gibt das vollständige `AnalysisResult`-JSON auf stdout
aus. URL-Inputs nutzen denselben Loader-Pfad wie die Bibliothek (siehe
§6 für SSRF-Regeln); Datei-Inputs werden direkt eingelesen und als
`kind: "text"` mit `baseUrl: "file://..."` an den Analyzer gegeben.

```bash
# URL gegen öffentlichen Stream
pnpm m-trace check https://cdn.example.test/manifest.m3u8

# Lokale Datei
pnpm m-trace check ./fixtures/master.m3u8

# Hilfe und Version
pnpm m-trace --help
pnpm m-trace --version
```

Empfehlung: pnpm-Output unterdrücken mit `pnpm --silent m-trace …`,
damit nur das Analyzer-JSON auf stdout landet — sinnvoll, wenn man
`pnpm m-trace check … | jq` oder `... > result.json` nutzt.

### 9.1 Exit-Codes

| Code | Bedeutung                                                        |
| ---- | ---------------------------------------------------------------- |
| 0    | Analyse erfolgreich (`status: "ok"`); JSON liegt auf stdout.     |
| 1    | Analyse-Fehler (`status: "error"` auf stdout) **oder** I/O-Fehler beim Lesen der Datei (Diagnose auf stderr). |
| 2    | Aufrufargument-/Usage-Fehler (Hilfe auf stderr).                 |

### 9.2 Smoke-Test

`make smoke-cli` baut das Paket und exerziert sieben Pfade:
`--help` (über die pnpm-Skript-Form), Master-Datei (Exit 0 + JSON),
Nicht-HLS-Datei (Exit 1 + `manifest_not_hls`), fehlende Datei (Exit 1
mit stderr-Hinweis), no-args (Exit 2), URL-Input gegen eine
RFC1918-Adresse (Exit 1 + `fetch_blocked` — exerciert den echten
Loader-Pfad inklusive SSRF-Schutz) und `--help` über `pnpm exec
m-trace` (Bin-Symlink + Shebang). Der Aufruf spiegelt das
DoD-Smoke-Kriterium aus plan-0.3.0 §8 Tranche 7.
