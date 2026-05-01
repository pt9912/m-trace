import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisResult,
  AnalysisSummary
} from "../../types/result.js";
import { classifyHlsManifest } from "./classify.js";

/**
 * Parst den Manifesttext so weit, wie es Tranche 2 verlangt:
 * Header- und Klassifikator-Erkennung. Master-/Media-Detail-Parsing
 * folgt in Tranche 3/4; die Stelle bleibt hier, damit `analyze.ts`
 * für Text- und URL-Inputs denselben Funktionspfad nutzt.
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
  findings.push({
    code: "details_pending",
    level: "info",
    message:
      "stream-analyzer 0.3.0 Tranche 2: Klassifikation abgeschlossen, typspezifische Detail-Auswertung folgt in Tranche 3/4."
  });

  const summary: AnalysisSummary = { itemCount: 0 };

  return {
    status: "ok",
    analyzerVersion,
    input: inputMeta,
    playlistType: classification.playlistType,
    summary,
    findings,
    details: null
  };
}
