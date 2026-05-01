import type { AnalysisFinding } from "./finding.js";

/**
 * Grobe Klassifikation der erkannten Manifestform. Weitere Werte
 * (z. B. DASH-MPD-Varianten, F-73) sind additiv erlaubt.
 */
export type PlaylistType = "master" | "media" | "unknown";

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
 * Erfolgsergebnis eines Analyseaufrufs. Das Schema bleibt additiv
 * erweiterbar (plan-0.3.0 §6 Tranche 5); typspezifische Details
 * landen in `details` und werden mit Tranche 3/4 ausgefüllt.
 */
export interface AnalysisResult {
  readonly status: "ok";
  readonly analyzerVersion: string;
  readonly input: AnalysisInputMetadata;
  readonly playlistType: PlaylistType;
  readonly summary: AnalysisSummary;
  readonly findings: readonly AnalysisFinding[];
  /**
   * Typspezifische Detail-Strukturen. Bei `playlistType === "master"`
   * ist die Form `MasterPlaylistDetails`; Tranche 4 ergänzt
   * `MediaPlaylistDetails`. Tranche 5 zieht die Diskriminierung in
   * den Typ. Bis dahin: `null` markiert „kein typspezifisches Detail
   * geliefert"; Konsumenten casten nach `playlistType`.
   */
  readonly details: Readonly<Record<string, unknown>> | null;
}

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
