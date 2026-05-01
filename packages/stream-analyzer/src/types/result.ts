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
 * Kennzeichnet, welcher Analyzer das Ergebnis erzeugt hat. Heute nur
 * `"hls"`; weitere Manifestformate (DASH, CMAF — F-73) werden additiv
 * als zusätzliche Werte ergänzt, ohne den Envelope zu brechen.
 */
export type AnalyzerKind = "hls";

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
  readonly playlistType: "master";
  readonly details: MasterPlaylistDetails;
}

export interface MediaAnalysisResult extends BaseAnalysisResult {
  readonly playlistType: "media";
  readonly details: MediaPlaylistDetails;
}

export interface UnknownAnalysisResult extends BaseAnalysisResult {
  readonly playlistType: "unknown";
  readonly details: null;
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
 * `docs/user/stream-analyzer.md` und `docs/planning/plan-0.3.0.md`.
 */
export type AnalysisResult = MasterAnalysisResult | MediaAnalysisResult | UnknownAnalysisResult;

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
