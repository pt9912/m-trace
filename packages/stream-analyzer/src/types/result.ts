import type { AnalysisErrorResult } from "./error.js";
import type { AnalysisFinding } from "./finding.js";

/**
 * Grobe Klassifikation der erkannten Manifestform.
 *
 * - HLS-Pfad: `master` / `media` / `unknown` (Fallback für HLS-
 *   Manifeste ohne klare Master-/Media-Tag-Distinktion).
 * - DASH-Pfad (ab `0.9.0` Tranche 3): `dash` als zweiter Wert,
 *   parallel zu HLS. DASH liefert kein Master/Media-Subtyping aus
 *   `analyzerKind`-Sicht — `playlistType: "dash"` ist die DASH-
 *   spezifische Klassifikation.
 *
 * Konsumenten, die exhaustiv über diesen Typ schalten, sollten einen
 * `default`-Branch behalten — neue Werte werden additiv ergänzt und
 * brechen sonst den Konsumenten-Build (siehe `docs/user/stream-
 * analyzer.md` §4).
 */
export type PlaylistType = "master" | "media" | "unknown" | "dash";

export interface AnalysisInputMetadata {
  /** Spiegelt die ursprüngliche Eingabeform; "url" markiert geladene Manifeste. */
  readonly source: "text" | "url";
  /** Quell-URL bei `source === "url"`, sonst `undefined`. */
  readonly url?: string;
  /** Aufgelöste Base-URL für relative URIs, falls bekannt. */
  readonly baseUrl?: string;
}

export interface AnalysisSummary {
  /**
   * Anzahl der erkannten Manifest-Kindelemente. Tranche 3 füllt das
   * für Master Playlists (Variants/Renditions); Tranche 4 für Media
   * Playlists (Segmente). Bis dahin bleibt der Wert 0.
   */
  readonly itemCount: number;
}

/**
 * Kennzeichnet, welcher Analyzer das Ergebnis erzeugt hat. Aktuelle
 * Werte: `"hls"` (seit `0.3.0`) und `"dash"` (ab `0.9.0` Tranche 3,
 * NF-12 / RAK-58); weitere Manifestformate (CMAF — F-73) werden
 * additiv ergänzt, ohne den Envelope zu brechen.
 *
 * Konsumenten unterscheiden HLS-Variants über `playlistType`
 * (`"master"`/`"media"`/`"unknown"`); DASH liefert `playlistType:
 * "dash"` als einzige Variante (DASH hat keine analoge Master/Media-
 * Trennung in der Manifest-Form selbst — die Period/AdaptationSet/
 * Representation-Hierarchie ist immer in einem MPD geliefert).
 */
export type AnalyzerKind = "hls" | "dash";

/**
 * Gemeinsame Felder aller Erfolgs-Result-Varianten. Konsumenten
 * sollten direkt das Union-Result `AnalysisResult` verwenden, damit
 * TypeScript via `playlistType` typgenau auf `details` schließt.
 */
export interface BaseAnalysisResult {
  readonly status: "ok";
  /** Aus `packages/stream-analyzer/package.json#version` abgeleitet. */
  readonly analyzerVersion: string;
  readonly analyzerKind: AnalyzerKind;
  readonly input: AnalysisInputMetadata;
  readonly summary: AnalysisSummary;
  readonly findings: readonly AnalysisFinding[];
}

export interface MasterAnalysisResult extends BaseAnalysisResult {
  readonly analyzerKind: "hls";
  readonly playlistType: "master";
  readonly details: MasterPlaylistDetails;
}

export interface MediaAnalysisResult extends BaseAnalysisResult {
  readonly analyzerKind: "hls";
  readonly playlistType: "media";
  readonly details: MediaPlaylistDetails;
}

/**
 * HLS-Manifest, das als Manifest erkannt, aber weder als Master noch
 * als Media klassifiziert wurde. `details: null` ist HLS-spezifisch;
 * der DASH-Pfad hat keine analoge „unbekannt"-Variante, weil ein
 * DASH-MPD entweder als MPD geparsed werden kann oder vom Detector
 * gar nicht erst als DASH eingestuft wird (→ `manifest_not_supported`).
 */
export interface UnknownAnalysisResult extends BaseAnalysisResult {
  readonly analyzerKind: "hls";
  readonly playlistType: "unknown";
  readonly details: null;
}

/**
 * Erfolgreiches DASH-MPD-Result (`0.9.0` Tranche 3, RAK-58 / NF-12).
 * Diskriminator-Paar `analyzerKind: "dash"` + `playlistType: "dash"`;
 * `details` trägt die geparsten Period/AdaptationSet/Representation-
 * Strukturen (siehe `DashManifestDetails`).
 *
 * Forward-Compat-Hinweis: wenn DASH-Live-MPDs später feinere
 * Sub-Klassifikationen brauchen (z. B. `playlistType: "dash-live"` /
 * `"dash-vod"`), bekommen sie eigene Varianten; der Live-Status
 * wandert in `details.live` als Boolean (analog HLS-`live` in
 * `MediaPlaylistDetails`).
 */
export interface DashAnalysisResult extends BaseAnalysisResult {
  readonly analyzerKind: "dash";
  readonly playlistType: "dash";
  readonly details: DashManifestDetails;
}

/**
 * Erfolgsergebnis eines Analyseaufrufs. Diskriminiert über
 * `playlistType`: TypeScript narrowed `details` automatisch auf den
 * passenden Typ (kein Cast notwendig).
 *
 * Stabilitätsregel (plan-0.3.0 §6): additive Änderungen sind erlaubt
 * (neue Felder, neue PlaylistType-Werte, neue analyzerKind-Werte,
 * neue Finding-Codes). Breaking Changes (Felder löschen/umbenennen/
 * umtypisieren, finite Wertedomänen einschränken) erfordern eine
 * Major-Version, einen Eintrag in `CHANGELOG.md` und ein Update von
 * `docs/user/stream-analyzer.md` und `docs/planning/done/plan-0.3.0.md`.
 */
export type AnalysisResult =
  | MasterAnalysisResult
  | MediaAnalysisResult
  | UnknownAnalysisResult
  | DashAnalysisResult;

/**
 * Vollständiger Rückgabetyp von `analyzeHlsManifest`. Trennt Erfolg
 * (`status === "ok"`) und Fehler (`status === "error"`) statisch.
 * Konsumenten sollten direkt diesen Typ verwenden, statt die Union
 * lokal aus `AnalysisResult | AnalysisErrorResult` zusammenzusetzen.
 */
export type AnalyzeOutput = AnalysisResult | AnalysisErrorResult;

/**
 * Ein Variant aus `#EXT-X-STREAM-INF`. Pflichtfeld ist `bandwidth`;
 * fehlt `BANDWIDTH`, wird der Eintrag dennoch aufgenommen
 * (`bandwidth: 0`) und mit einem Error-Finding markiert. Optionale
 * Felder fehlen, wenn das Tag sie nicht setzt.
 */
export interface MasterVariant {
  readonly bandwidth: number;
  readonly averageBandwidth?: number;
  readonly resolution?: { readonly width: number; readonly height: number };
  readonly codecs?: readonly string[];
  readonly frameRate?: number;
  readonly audio?: string;
  readonly video?: string;
  readonly subtitles?: string;
  readonly closedCaptions?: string;
  /** URI exakt wie im Manifest geliefert (relativ oder absolut). */
  readonly uri: string;
  /** Absolute URI nach Auflösung gegen die Base-URL, falls vorhanden. */
  readonly resolvedUri?: string;
}

/**
 * Eine Rendition aus `#EXT-X-MEDIA`. Pflichtfelder sind `type`,
 * `groupId`, `name`; alles andere optional, weil je nach Typ
 * unterschiedlich relevant.
 */
export interface MasterRendition {
  readonly type: string;
  readonly groupId: string;
  readonly name: string;
  readonly language?: string;
  readonly uri?: string;
  readonly resolvedUri?: string;
  readonly default?: boolean;
  readonly autoselect?: boolean;
  readonly forced?: boolean;
  readonly channels?: string;
}

export interface MasterPlaylistDetails {
  readonly variants: readonly MasterVariant[];
  readonly renditions: readonly MasterRendition[];
}

/**
 * Ein Segment aus `#EXTINF` plus folgender URI-Zeile.
 */
export interface MediaSegment {
  /** URI exakt wie im Manifest (Whitespace getrimmt). */
  readonly uri: string;
  /** Absolute URI nach Auflösung gegen die Base-URL, falls vorhanden. */
  readonly resolvedUri?: string;
  /** Dauer in Sekunden. */
  readonly duration: number;
  /** Optionaler Titel aus `#EXTINF:duration,title`. */
  readonly title?: string;
  /**
   * HLS-Sequenznummer. Erstes Segment startet bei `mediaSequence`,
   * jedes weitere +1. Fehlt `#EXT-X-MEDIA-SEQUENCE`, beginnt die
   * Zählung bei 0.
   */
  readonly sequenceNumber: number;
}

/**
 * Aggregat-Statistiken über alle Segmente. `count === 0` markiert
 * eine Media-Playlist ohne ausgewertete Segmente; in dem Fall sind
 * `min`/`max`/`average`/`total` 0.
 */
export interface MediaSegmentSummary {
  readonly count: number;
  readonly averageDuration: number;
  readonly minDuration: number;
  readonly maxDuration: number;
  readonly totalDuration: number;
}

/**
 * Detail-Sektion einer HLS Media Playlist (RFC 8216 §4.3.3).
 *
 * `live === !endList`. `liveLatencyEstimateSeconds` ist die einfache
 * 3×-Target-Duration-Schätzung nach Apples HLS-Authoring-Empfehlung
 * (siehe `docs/user/stream-analyzer.md` §7); für VOD-Playlists
 * undefiniert.
 */
export interface MediaPlaylistDetails {
  readonly targetDuration?: number;
  readonly mediaSequence: number;
  /** Wert von `#EXT-X-PLAYLIST-TYPE` (`VOD` oder `EVENT`), falls gesetzt. */
  readonly playlistType?: string;
  readonly endList: boolean;
  readonly live: boolean;
  readonly liveLatencyEstimateSeconds?: number;
  readonly segments: readonly MediaSegment[];
  readonly summary: MediaSegmentSummary;
}

/**
 * Eine Representation aus einem DASH-MPD-`AdaptationSet`. Pflichtfeld
 * ist `id` (XML-Attribut von `<Representation>`); fehlt es, wird die
 * Representation dennoch aufgenommen (`id: ""`) und mit einem
 * `dash_representation_missing_id`-Finding markiert. `bandwidth` ist
 * laut MPEG-DASH-Spec (ISO/IEC 23009-1) Pflicht; fehlt es, ist der
 * Wert `0` und es kommt ein Error-Finding.
 */
export interface DashRepresentation {
  readonly id: string;
  readonly bandwidth: number;
  readonly width?: number;
  readonly height?: number;
  readonly frameRate?: string;
  readonly codecs?: string;
  readonly mimeType?: string;
  readonly audioSamplingRate?: string;
}

/**
 * Eine `AdaptationSet`-Gruppe aus einer DASH-Period. `mimeType` und
 * `codecs` werden auf AdaptationSet-Ebene erfasst, falls dort gesetzt;
 * fallback-fähig nach Representation-Ebene (siehe `dash.ts`).
 */
export interface DashAdaptationSet {
  readonly id?: string;
  readonly mimeType?: string;
  readonly codecs?: string;
  readonly contentType?: string;
  readonly lang?: string;
  readonly representations: readonly DashRepresentation[];
}

/**
 * Detail-Sektion eines DASH-MPD (`0.9.0` Tranche 3, RAK-58). Mindest-
 * Felder pro `details.adaptationSets[]`-Eintrag sind `mimeType`,
 * `codecs`, `bandwidth`, `width`/`height` (letztere zwei optional —
 * Audio-only-Streams haben keine Auflösung); `summary.itemCount`
 * im äußeren Result zählt die Gesamtzahl der Representations über
 * alle Periods/AdaptationSets.
 *
 * Live vs. VOD ist über `live: boolean` markiert (geleitet aus
 * `MPD@type` — `dynamic` → live, `static` → VOD; Default `static`).
 * `availabilityStartTime`-/Segment-Template-Edge-Cases (z. B.
 * `$Time$`-Variablen) sind aus dem Plan-Scope explizit ausgenommen
 * (siehe `plan-0.9.0.md` §0.3).
 */
export interface DashManifestDetails {
  readonly profiles?: string;
  readonly type: "static" | "dynamic";
  readonly live: boolean;
  readonly mediaPresentationDuration?: string;
  readonly minimumUpdatePeriod?: string;
  readonly availabilityStartTime?: string;
  readonly periodCount: number;
  readonly adaptationSets: readonly DashAdaptationSet[];
}
