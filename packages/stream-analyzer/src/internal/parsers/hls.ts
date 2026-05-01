import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisResult,
  AnalysisSummary,
  MasterPlaylistDetails
} from "../../types/result.js";
import { classifyHlsManifest } from "./classify.js";
import { parseMasterPlaylist } from "./master.js";

/**
 * Setzt die Tranche-Resultate zusammen: Tranche 2 hat klassifiziert,
 * Tranche 3 wertet Master-Playlists detailliert aus, Tranche 4 ergänzt
 * Media-Playlists. `details_pending` markiert Tranchen, die noch
 * keinen Detail-Parser angeschlossen haben — das ist hilfreich für
 * Konsumenten, die früh integrieren wollen.
 */
export function analyzeHlsManifestText(
  text: string,
  inputMeta: AnalysisInputMetadata,
  analyzerVersion: string
): AnalysisResult {
  const classification = classifyHlsManifest(text);

  const findings: AnalysisFinding[] = [];
  if (classification.ambiguous) {
    findings.push({
      code: "playlist_ambiguous",
      level: "warning",
      message:
        "Manifest enthält sowohl Master- als auch Media-Tags. Tranche 2 klassifiziert es als Master; Tranche 3/4 entscheidet die Detail-Auswertung."
    });
  }
  if (classification.playlistType === "unknown") {
    findings.push({
      code: "playlist_type_unknown",
      level: "warning",
      message:
        "Manifest beginnt mit #EXTM3U, enthält aber weder Master- noch Media-Tags. Inhalt wird als unklassifiziert gemeldet."
    });
  }

  let details: MasterPlaylistDetails | null = null;
  let summary: AnalysisSummary = { itemCount: 0 };
  if (classification.playlistType === "master") {
    const result = parseMasterPlaylist(text, inputMeta.baseUrl);
    details = result.details;
    summary = { itemCount: result.details.variants.length + result.details.renditions.length };
    findings.push(...result.findings);
  } else if (classification.playlistType === "media" || classification.playlistType === "unknown") {
    findings.push({
      code: "details_pending",
      level: "info",
      message:
        "stream-analyzer 0.3.0: Detail-Auswertung für diesen Playlist-Typ folgt in Tranche 4."
    });
  }

  return {
    status: "ok",
    analyzerVersion,
    input: inputMeta,
    playlistType: classification.playlistType,
    summary,
    findings,
    // Cast: AnalysisResult.details bleibt bis Tranche 5 lose typisiert
    // (Diskriminierte Union per playlistType folgt). Hier ist die Form
    // beim Erstellen exakt bekannt; Konsumenten casten gemäß §2 der
    // User-Doku auf MasterPlaylistDetails.
    details: details as Readonly<Record<string, unknown>> | null
  };
}
