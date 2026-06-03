import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisResult,
  AnalysisSummary
} from "../../types/result.js";
import { classifyHlsManifest } from "./classify.js";
import type { HlsMediaCmafMetadata } from "./cmaf-hls.js";
import { parseMasterPlaylist } from "./master.js";
import { parseMediaPlaylist } from "./media.js";

/**
 * Output des HLS-Pfads — `result` ist die Public-`AnalysisResult`-Form,
 * `cmafMeta` trägt die internen Tranche-4-Eingabedaten für die binäre
 * Konformitätsprüfung im Media-Playlist-Pfad. `cmafMeta` ist nur bei
 * Media-Playlists mit CMAF-Signalen gesetzt; Master-/Unknown-Pfad
 * tragen es nicht.
 */
export interface HlsAnalyzeOutput {
  readonly result: AnalysisResult;
  readonly cmafMeta?: HlsMediaCmafMetadata;
}

/**
 * Setzt die Tranche-Resultate zusammen: hat klassifiziert,
 * wertet Master-Playlists detailliert aus, ergänzt
 * Media-Playlists, zieht alles in einen diskriminierten
 * Union-Typ. `details_pending` bleibt als Info-Marker für
 * `playlistType === "unknown"` erhalten — dort hat der Analyzer
 * bewusst keinen Detail-Pfad.
 */
export function analyzeHlsManifestText(
  text: string,
  inputMeta: AnalysisInputMetadata,
  analyzerVersion: string
): HlsAnalyzeOutput {
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

  const base = {
    status: "ok" as const,
    analyzerVersion,
    analyzerKind: "hls" as const,
    input: inputMeta
  };

  if (classification.playlistType === "master") {
    const result = parseMasterPlaylist(text, inputMeta.baseUrl);
    findings.push(...result.findings);
    const summary: AnalysisSummary = {
      itemCount: result.details.variants.length + result.details.renditions.length
    };
    return {
      result: {
        ...base,
        summary,
        findings,
        playlistType: "master",
        details: result.details
      }
    };
  }

  if (classification.playlistType === "media") {
    const result = parseMediaPlaylist(text, inputMeta.baseUrl);
    findings.push(...result.findings);
    const summary: AnalysisSummary = { itemCount: result.details.segments.length };
    return {
      result: {
        ...base,
        summary,
        findings,
        playlistType: "media",
        details: result.details
      },
      ...(result.cmafMeta !== undefined ? { cmafMeta: result.cmafMeta } : {})
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
      "stream-analyzer 0.6.0: Manifest ist als HLS erkannt, aber weder als Master noch als Media klassifiziert."
  });

  return {
    result: {
      ...base,
      summary: { itemCount: 0 },
      findings,
      playlistType: "unknown",
      details: null
    }
  };
}
