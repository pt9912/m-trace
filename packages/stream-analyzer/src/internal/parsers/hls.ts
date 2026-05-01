import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisResult,
  AnalysisSummary,
  PlaylistType
} from "../../types/result.js";

/**
 * Tranche-1-Stub: liefert eine stabile, additiv erweiterbare
 * Result-Struktur, ohne den eigentlichen Parser zu implementieren.
 * Das tatsächliche HLS-Parsing wandert in Tranche 2/3/4 hierher;
 * bis dahin meldet der Stub ein `not_implemented`-Finding, damit
 * Konsumenten den Implementierungsstand sofort erkennen.
 *
 * Der Parser bekommt den Manifesttext und die Input-Metadaten getrennt:
 * Tranche 2 kann URL-Manifeste laden und an dieselbe Parser-Funktion
 * weitergeben, ohne dass die Result-Struktur zwischen Text- und
 * URL-Fall divergiert.
 */
export function analyzeHlsManifestText(
  _text: string,
  inputMeta: AnalysisInputMetadata,
  analyzerVersion: string
): AnalysisResult {
  const playlistType: PlaylistType = "unknown";
  const summary: AnalysisSummary = { itemCount: 0 };
  const findings: AnalysisFinding[] = [
    {
      code: "not_implemented",
      level: "info",
      message:
        "stream-analyzer 0.3.0 Tranche 1: HLS-Parser ist noch nicht angeschlossen; Result-Schema ist stabil."
    }
  ];

  return {
    status: "ok",
    analyzerVersion,
    input: inputMeta,
    playlistType,
    summary,
    findings,
    details: null
  };
}
