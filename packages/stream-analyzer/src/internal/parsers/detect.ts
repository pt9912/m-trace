import type { AnalyzerKind } from "../../types/result.js";

/**
 * Klassifiziert einen Manifest-Text-Body als HLS, DASH oder
 * unsupported (`plan-0.9.0` §4 Tranche 3, RAK-58). Der Detector
 * entscheidet, welcher Parser den Text bekommt; eine eindeutige
 * Klassifikation hier ist Voraussetzung dafür, dass `manifest_not_hls`
 * (HLS-spezifisch) und `manifest_not_supported` (Detector-Sammelfehler)
 * sauber getrennt bleiben.
 *
 * Erkennungsregeln (in Priorität):
 *  1. **DASH** — der Body beginnt (nach Leerzeilen/optionalem BOM)
 *     mit `<?xml` oder `<MPD`. `<?xml` ist die XML-Deklaration nach
 *     XML 1.0 §2.8; `<MPD` ist das ISO/IEC 23009-1 §5.3.1 Wurzel-
 *     Element. Optionale Content-Type-Heuristik aus dem Loader
 *     (`application/dash+xml`) ist redundant zur Header-Heuristik.
 *  2. **HLS** — die erste nicht-leere Zeile ist `#EXTM3U` (oder
 *     beginnt mit `#EXTM3U`, falls Trailing-Whitespace).
 *  3. **Unsupported** — alles andere (HTML, JSON, leerer Body,
 *     Binärdaten).
 *
 * Der Detector entscheidet **ohne** vollen Parse: er liest nur den
 * Anfang des Bodies. Ein `<MPD>` ohne XML-Deklaration und ohne
 * Schließtag wird als DASH klassifiziert; der Parser wirft dann
 * `internal_error`, wenn der MPD-Body strukturell defekt ist.
 */
export interface DetectResult {
  readonly kind: AnalyzerKind | "unsupported";
  /** Erste nicht-leere Zeile (für Diagnose-Findings, max. 80 Zeichen). */
  readonly firstLine: string;
}

const HLS_HEADER = "#EXTM3U";

export function detectManifestKind(text: string): DetectResult {
  // BOM (UTF-8 EF BB BF wird in JS bereits als U+FEFF dekodiert).
  let body = text;
  if (body.charCodeAt(0) === 0xfeff) {
    body = body.slice(1);
  }

  // DASH-Detector: prüft den Body-Anfang nach Leerzeichen/Zeilen-
  // umbrüchen. `<?xml` und `<MPD` sind beide eindeutige DASH-Marker.
  const trimmedStart = body.replace(/^[\s\r\n]+/, "");
  if (trimmedStart.startsWith("<?xml") || trimmedStart.startsWith("<MPD")) {
    return { kind: "dash", firstLine: firstNonEmptyLine(body) };
  }

  // HLS-Detector: erste nicht-leere Zeile.
  const firstLine = firstNonEmptyLine(body);
  if (firstLine === HLS_HEADER || firstLine.startsWith(`${HLS_HEADER} `)) {
    return { kind: "hls", firstLine };
  }
  if (firstLine === "") {
    // Leerer Body — weder HLS noch DASH.
    return { kind: "unsupported", firstLine: "" };
  }
  return { kind: "unsupported", firstLine };
}

function firstNonEmptyLine(text: string): string {
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim();
    if (line.length > 0) {
      return line.slice(0, 80);
    }
  }
  return "";
}
