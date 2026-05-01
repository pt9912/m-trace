import type { AnalysisFinding } from "../../types/finding.js";
import type { MasterPlaylistDetails, MasterRendition, MasterVariant } from "../../types/result.js";
import {
  parseAttributeList,
  parseCodecs,
  parseFloatAttr,
  parseIntAttr,
  parseResolution,
  parseYesNo
} from "./attrs.js";

const STREAM_INF_PREFIX = "#EXT-X-STREAM-INF:";
const MEDIA_PREFIX = "#EXT-X-MEDIA:";
const I_FRAME_PREFIX = "#EXT-X-I-FRAME-STREAM-INF:";

const VALID_RENDITION_TYPES = new Set(["AUDIO", "VIDEO", "SUBTITLES", "CLOSED-CAPTIONS"]);

interface PendingStreamInf {
  readonly attrs: Map<string, string>;
  readonly lineNumber: number; // 0-basiert (Tag-Zeile)
}

interface MasterParseResult {
  readonly details: MasterPlaylistDetails;
  readonly findings: readonly AnalysisFinding[];
}

/**
 * Parst eine HLS Master Playlist (RFC 8216 §4.3.4) in eine
 * strukturierte Variant-/Rendition-Liste plus Findings.
 *
 * Tranche-3-Scope laut plan-0.3.0 §4:
 *  - `#EXT-X-STREAM-INF` Variants: BANDWIDTH (Pflicht), AVERAGE-
 *    BANDWIDTH, RESOLUTION, CODECS, FRAME-RATE, AUDIO, VIDEO,
 *    SUBTITLES, CLOSED-CAPTIONS.
 *  - `#EXT-X-MEDIA` Renditions: TYPE, GROUP-ID, NAME, LANGUAGE,
 *    URI, DEFAULT, AUTOSELECT, FORCED, CHANNELS.
 *  - Optionale Felder fehlen ohne Abbruch; offensichtliche
 *    Inkonsistenzen werden als Findings gemeldet.
 *  - Relative URIs bleiben als `uri` erhalten; `resolvedUri` ist
 *    die optionale absolute Form gegen `baseUrl`.
 */
export function parseMasterPlaylist(text: string, baseUrl: string | undefined): MasterParseResult {
  const lines = text.split(/\r?\n/);
  const variants: MasterVariant[] = [];
  const renditions: MasterRendition[] = [];
  const findings: AnalysisFinding[] = [];

  let pending: PendingStreamInf | null = null;

  for (let lineIdx = 0; lineIdx < lines.length; lineIdx++) {
    const line = lines[lineIdx].trim();
    if (line.length === 0) continue;
    if (line === "#EXTM3U") continue;

    if (line.startsWith(STREAM_INF_PREFIX)) {
      if (pending !== null) {
        findings.push(missingUriFinding(pending.lineNumber));
      }
      pending = {
        attrs: parseAttributeList(line.slice(STREAM_INF_PREFIX.length)),
        lineNumber: lineIdx
      };
      continue;
    }

    if (line.startsWith(MEDIA_PREFIX)) {
      const attrs = parseAttributeList(line.slice(MEDIA_PREFIX.length));
      const built = buildRendition(attrs, baseUrl, lineIdx);
      if (built !== null) {
        renditions.push(built.rendition);
        findings.push(...built.findings);
      } else {
        findings.push({
          code: "rendition_missing_required_attr",
          level: "error",
          message: `EXT-X-MEDIA auf Zeile ${lineIdx + 1} hat kein TYPE, GROUP-ID oder NAME.`
        });
      }
      continue;
    }

    if (line.startsWith(I_FRAME_PREFIX)) {
      // I-Frame-Variants haben keine separate URI-Zeile (URI im Tag).
      // Tranche 3 behandelt sie nicht eigenständig; ein Hinweis
      // markiert sie für Tranche-3-Konsumenten.
      findings.push({
        code: "i_frame_variant_skipped",
        level: "info",
        message: `EXT-X-I-FRAME-STREAM-INF auf Zeile ${lineIdx + 1} wird in 0.3.0 nicht ausgewertet (Folge-Tranche).`
      });
      continue;
    }

    if (line.startsWith("#")) continue;

    if (pending !== null) {
      const built = buildVariant(pending.attrs, line, baseUrl, pending.lineNumber);
      variants.push(built.variant);
      findings.push(...built.findings);
      pending = null;
    }
  }

  if (pending !== null) {
    findings.push(missingUriFinding(pending.lineNumber));
  }

  findings.push(...detectDuplicateRenditions(renditions));
  findings.push(...crossReferenceGroups(variants, renditions));

  return { details: { variants, renditions }, findings };
}

function missingUriFinding(lineNumber: number): AnalysisFinding {
  return {
    code: "variant_missing_uri",
    level: "error",
    message: `EXT-X-STREAM-INF auf Zeile ${lineNumber + 1} hat keine darauffolgende URI-Zeile.`
  };
}

interface VariantBuildResult {
  readonly variant: MasterVariant;
  readonly findings: readonly AnalysisFinding[];
}

function buildVariant(
  attrs: Map<string, string>,
  rawUri: string,
  baseUrl: string | undefined,
  lineNumber: number
): VariantBuildResult {
  const findings: AnalysisFinding[] = [];

  const bandwidthValue = attrs.get("BANDWIDTH");
  const bandwidth = parseIntAttr(bandwidthValue);
  if (bandwidth === null) {
    findings.push({
      code: "variant_missing_bandwidth",
      level: "error",
      message: `EXT-X-STREAM-INF auf Zeile ${lineNumber + 1} fehlt das Pflichtattribut BANDWIDTH.`
    });
  }

  const resolutionRaw = attrs.get("RESOLUTION");
  const resolution = parseResolution(resolutionRaw);
  if (resolutionRaw !== undefined && resolution === null) {
    findings.push({
      code: "variant_malformed_resolution",
      level: "warning",
      message: `RESOLUTION="${resolutionRaw}" auf Zeile ${lineNumber + 1} entspricht nicht dem Format WIDTHxHEIGHT.`
    });
  }

  const codecs = parseCodecs(attrs.get("CODECS"));
  const frameRate = parseFloatAttr(attrs.get("FRAME-RATE"));
  const averageBandwidth = parseIntAttr(attrs.get("AVERAGE-BANDWIDTH"));
  const resolvedUri = resolveUri(rawUri, baseUrl);
  if (rawUri.length > 0 && resolvedUri === null && baseUrl !== undefined) {
    findings.push({
      code: "variant_malformed_uri",
      level: "warning",
      message: `Variant-URI "${rawUri}" auf Zeile ${lineNumber + 2} konnte nicht gegen Base-URL aufgelöst werden.`
    });
  }

  const variant: MasterVariant = {
    bandwidth: bandwidth ?? 0,
    ...(averageBandwidth !== null ? { averageBandwidth } : {}),
    ...(resolution !== null ? { resolution } : {}),
    ...(codecs !== null ? { codecs } : {}),
    ...(frameRate !== null ? { frameRate } : {}),
    ...optionalString(attrs, "AUDIO", "audio"),
    ...optionalString(attrs, "VIDEO", "video"),
    ...optionalString(attrs, "SUBTITLES", "subtitles"),
    ...optionalString(attrs, "CLOSED-CAPTIONS", "closedCaptions"),
    uri: rawUri,
    ...(resolvedUri !== null ? { resolvedUri } : {})
  };
  return { variant, findings };
}

interface RenditionBuildResult {
  readonly rendition: MasterRendition;
  readonly findings: readonly AnalysisFinding[];
}

function buildRendition(
  attrs: Map<string, string>,
  baseUrl: string | undefined,
  lineNumber: number
): RenditionBuildResult | null {
  const type = attrs.get("TYPE");
  const groupId = attrs.get("GROUP-ID");
  const name = attrs.get("NAME");
  if (type === undefined || groupId === undefined || name === undefined) {
    return null;
  }
  const findings: AnalysisFinding[] = [];
  if (!VALID_RENDITION_TYPES.has(type)) {
    findings.push({
      code: "rendition_unknown_type",
      level: "warning",
      message: `EXT-X-MEDIA auf Zeile ${lineNumber + 1} verwendet unbekannten TYPE "${type}".`
    });
  }
  const uri = attrs.get("URI");
  // RFC 8216 §4.3.4.2.1: URI ist Pflicht für SUBTITLES; bei AUDIO/VIDEO
  // bedeutet das Fehlen, dass die Rendition in der Variant-Playlist
  // liegt — kein Fehler. Nur SUBTITLES ohne URI ist ein Spec-Verstoß.
  if (type === "SUBTITLES" && uri === undefined) {
    findings.push({
      code: "rendition_missing_uri",
      level: "error",
      message: `EXT-X-MEDIA TYPE=SUBTITLES auf Zeile ${lineNumber + 1} ohne URI; URI ist für Untertitel-Renditions Pflicht.`
    });
  }
  const resolvedUri = uri !== undefined ? resolveUri(uri, baseUrl) : null;
  if (uri !== undefined && resolvedUri === null && baseUrl !== undefined) {
    findings.push({
      code: "rendition_malformed_uri",
      level: "warning",
      message: `Rendition-URI "${uri}" auf Zeile ${lineNumber + 1} konnte nicht gegen Base-URL aufgelöst werden.`
    });
  }
  const rendition: MasterRendition = {
    type,
    groupId,
    name,
    ...optionalString(attrs, "LANGUAGE", "language"),
    ...(uri !== undefined ? { uri } : {}),
    ...(resolvedUri !== null ? { resolvedUri } : {}),
    ...optionalBoolean(attrs, "DEFAULT", "default"),
    ...optionalBoolean(attrs, "AUTOSELECT", "autoselect"),
    ...optionalBoolean(attrs, "FORCED", "forced"),
    ...optionalString(attrs, "CHANNELS", "channels")
  };
  return { rendition, findings };
}

function crossReferenceGroups(
  variants: readonly MasterVariant[],
  renditions: readonly MasterRendition[]
): AnalysisFinding[] {
  const audioGroups = new Set(renditions.filter((r) => r.type === "AUDIO").map((r) => r.groupId));
  const videoGroups = new Set(renditions.filter((r) => r.type === "VIDEO").map((r) => r.groupId));
  const subtitlesGroups = new Set(renditions.filter((r) => r.type === "SUBTITLES").map((r) => r.groupId));
  const ccGroups = new Set(renditions.filter((r) => r.type === "CLOSED-CAPTIONS").map((r) => r.groupId));

  const findings: AnalysisFinding[] = [];
  for (let i = 0; i < variants.length; i++) {
    const v = variants[i];
    findings.push(...checkGroup(i, "audio", v.audio, audioGroups));
    findings.push(...checkGroup(i, "video", v.video, videoGroups));
    findings.push(...checkGroup(i, "subtitles", v.subtitles, subtitlesGroups));
    // RFC 8216 §4.3.4.2.1 erlaubt CLOSED-CAPTIONS=NONE als explizite
    // „keine CC-Gruppe"-Markierung. Das ist kein Group-Reference,
    // sondern ein Sentinel und darf keinen variant_group_undefined-
    // Finding auslösen.
    if (v.closedCaptions !== "NONE") {
      findings.push(...checkGroup(i, "closedCaptions", v.closedCaptions, ccGroups));
    }
  }
  return findings;
}

function checkGroup(
  variantIndex: number,
  field: string,
  groupRef: string | undefined,
  declared: ReadonlySet<string>
): AnalysisFinding[] {
  if (groupRef === undefined) return [];
  if (declared.has(groupRef)) return [];
  return [
    {
      code: "variant_group_undefined",
      level: "warning",
      message: `Variant #${variantIndex + 1} referenziert ${field}="${groupRef}", aber keine passende EXT-X-MEDIA-Gruppe existiert.`
    }
  ];
}

function detectDuplicateRenditions(renditions: readonly MasterRendition[]): AnalysisFinding[] {
  // RFC 8216 §4.3.4.1.1: Renditions mit gleichem TYPE+GROUP-ID müssen
  // unterschiedliche NAME-Werte tragen. Bei Wiederholung melden wir
  // den zweiten (und folgende) Eintrag als Duplikat.
  const seen = new Set<string>();
  const findings: AnalysisFinding[] = [];
  for (const r of renditions) {
    const key = `${r.type}|${r.groupId}|${r.name}`;
    if (seen.has(key)) {
      findings.push({
        code: "rendition_duplicate_group_member",
        level: "warning",
        message: `Mehrere EXT-X-MEDIA-Einträge mit TYPE=${r.type}, GROUP-ID="${r.groupId}", NAME="${r.name}".`
      });
    } else {
      seen.add(key);
    }
  }
  return findings;
}

function optionalString<K extends string>(
  attrs: Map<string, string>,
  attrKey: string,
  outKey: K
): { readonly [P in K]?: string } {
  const value = attrs.get(attrKey);
  if (value === undefined || value.length === 0) return {} as { readonly [P in K]?: string };
  return { [outKey]: value } as { readonly [P in K]?: string };
}

function optionalBoolean<K extends string>(
  attrs: Map<string, string>,
  attrKey: string,
  outKey: K
): { readonly [P in K]?: boolean } {
  const parsed = parseYesNo(attrs.get(attrKey));
  if (parsed === undefined) return {} as { readonly [P in K]?: boolean };
  return { [outKey]: parsed } as { readonly [P in K]?: boolean };
}

function resolveUri(rawUri: string, baseUrl: string | undefined): string | null {
  if (baseUrl === undefined) return null;
  try {
    return new URL(rawUri, baseUrl).toString();
  } catch {
    return null;
  }
}
