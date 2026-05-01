import type { PlaylistType } from "../../types/result.js";
import { AnalysisError } from "../../types/error.js";

/**
 * Klassifiziert den Manifesttext, ohne ihn vollständig zu parsen.
 * Tranche 2 liefert die Master-/Media-Erkennung; Tranche 3/4
 * extrahieren danach die typspezifischen Inhalte.
 */
export interface ClassifyResult {
  readonly playlistType: PlaylistType;
  /**
   * `true`, wenn sowohl Master- als auch Media-Tags gefunden wurden.
   * Der Aufrufer dokumentiert dies als Finding (plan-0.3.0 §3 DoD
   * "Ambige oder gemischte Playlists").
   */
  readonly ambiguous: boolean;
}

const MASTER_TAGS = new Set(["#EXT-X-STREAM-INF", "#EXT-X-MEDIA", "#EXT-X-I-FRAME-STREAM-INF"]);
const MEDIA_TAGS = new Set([
  "#EXTINF",
  "#EXT-X-TARGETDURATION",
  "#EXT-X-MEDIA-SEQUENCE",
  "#EXT-X-PLAYLIST-TYPE",
  "#EXT-X-ENDLIST"
]);

const HLS_HEADER = "#EXTM3U";

/**
 * Lehnt nicht-HLS-Inhalte mit AnalysisError(manifest_not_hls) ab und
 * gibt ansonsten den erkannten Playlist-Typ zurück.
 *
 * Erkennungsregeln:
 *  - Erste nicht-leere Zeile muss `#EXTM3U` sein, sonst manifest_not_hls.
 *  - Master-Tag und Media-Tag schließen sich nicht aus; beides
 *    gleichzeitig → ambiguous=true und playlistType="master" (Master
 *    referenziert Media; HLS-Spec lässt das implizit zu, wir markieren
 *    es aber als Finding).
 *  - Nur Media-Tag → "media".
 *  - Nur Master-Tag → "master".
 *  - Weder noch → "unknown" (HLS, aber Tags reichen für Klassifikation
 *    nicht aus).
 */
export function classifyHlsManifest(text: string): ClassifyResult {
  const lines = text.split(/\r?\n/);
  let firstNonEmptySeen = false;
  let hasMaster = false;
  let hasMedia = false;

  for (const raw of lines) {
    const line = raw.trim();
    if (line.length === 0) continue;
    if (!firstNonEmptySeen) {
      firstNonEmptySeen = true;
      if (line !== HLS_HEADER) {
        throw new AnalysisError(
          "manifest_not_hls",
          "Manifest beginnt nicht mit #EXTM3U.",
          { firstLine: line.slice(0, 80) }
        );
      }
      continue;
    }
    if (!line.startsWith("#")) continue;

    const tag = extractTag(line);
    if (MASTER_TAGS.has(tag)) {
      hasMaster = true;
    }
    if (MEDIA_TAGS.has(tag)) {
      hasMedia = true;
    }
  }

  if (!firstNonEmptySeen) {
    throw new AnalysisError("manifest_not_hls", "Manifest ist leer.");
  }

  if (hasMaster && hasMedia) {
    return { playlistType: "master", ambiguous: true };
  }
  if (hasMaster) {
    return { playlistType: "master", ambiguous: false };
  }
  if (hasMedia) {
    return { playlistType: "media", ambiguous: false };
  }
  return { playlistType: "unknown", ambiguous: false };
}

function extractTag(line: string): string {
  const colonIndex = line.indexOf(":");
  return colonIndex === -1 ? line : line.slice(0, colonIndex);
}
