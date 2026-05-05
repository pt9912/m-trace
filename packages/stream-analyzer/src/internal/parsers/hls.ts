import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisResult,
  AnalysisSummary,
  BaseAnalysisResult
} from "../../types/result.js";
import { classifyHlsManifest } from "./classify.js";
import { parseMasterPlaylist } from "./master.js";
import { parseMediaPlaylist } from "./media.js";

/**
 * Setzt die Tranche-Resultate zusammen: Tranche 2 hat klassifiziert,
 * Tranche 3 wertet Master-Playlists detailliert aus, Tranche 4 ergänzt
 * Media-Playlists, Tranche 5 zieht alles in einen diskriminierten
 * Union-Typ. `details_pending` bleibt als Info-Marker für
 * `playlistType === "unknown"` erhalten — dort hat der Analyzer
 * bewusst keinen Detail-Pfad.
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

  const base: Omit<BaseAnalysisResult, "summary" | "findings"> = {
    status: "ok",
    analyzerVersion,
    analyzerKind: "hls",
    input: inputMeta
  };

  if (classification.playlistType === "master") {
    const result = parseMasterPlaylist(text, inputMeta.baseUrl);
    findings.push(...result.findings);
    const summary: AnalysisSummary = {
      itemCount: result.details.variants.length + result.details.renditions.length
    };
    return {
      ...base,
      summary,
      findings,
      playlistType: "master",
      details: result.details
    };
  }

  if (classification.playlistType === "media") {
    const result = parseMediaPlaylist(text, inputMeta.baseUrl);
    findings.push(...result.findings);
    const summary: AnalysisSummary = { itemCount: result.details.segments.length };
    return {
      ...base,
      summary,
      findings,
      playlistType: "media",
      details: result.details
    };
  }

  findings.push({
    code: "playlist_type_unknown",
    level: "warning",
    message:
      "Manifest beginnt mit #EXTM3U, enthält aber weder Master- noch Media-Tags. Inhalt wird als unklassifiziert gemeldet."
  });
  findings.push({
    code: "details_pending",
    level: "info",
    message:
      "stream-analyzer 0.4.0: Manifest ist als HLS erkannt, aber weder als Master noch als Media klassifiziert."
  });

  return {
    ...base,
    summary: { itemCount: 0 },
    findings,
    playlistType: "unknown",
    details: null
  };
}
