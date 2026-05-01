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
   * Typspezifische Detail-Strukturen (Master-Playlist-Varianten,
   * Media-Playlist-Segmente, …). Tranche 3/4 stabilisieren das Shape;
   * `null` markiert „kein typspezifisches Detail geliefert".
   */
  readonly details: Readonly<Record<string, unknown>> | null;
}
