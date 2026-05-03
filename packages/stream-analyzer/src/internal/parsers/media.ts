import type { AnalysisFinding } from "../../types/finding.js";
import type {
  MediaPlaylistDetails,
  MediaSegment,
  MediaSegmentSummary
} from "../../types/result.js";
import { parseFloatAttr, parseIntAttr } from "./attrs.js";

const TARGETDURATION_PREFIX = "#EXT-X-TARGETDURATION:";
const MEDIASEQUENCE_PREFIX = "#EXT-X-MEDIA-SEQUENCE:";
const PLAYLISTTYPE_PREFIX = "#EXT-X-PLAYLIST-TYPE:";
const EXTINF_PREFIX = "#EXTINF:";
const ENDLIST_TAG = "#EXT-X-ENDLIST";
const KEY_PREFIX = "#EXT-X-KEY:";
const MAP_PREFIX = "#EXT-X-MAP:";
const DISCONTINUITY_TAG = "#EXT-X-DISCONTINUITY";
const PROGRAM_DATE_TIME_PREFIX = "#EXT-X-PROGRAM-DATE-TIME:";

/**
 * Toleranzgrenze für `segment_duration_outlier`. Segmentdauern, die
 * unter `OUTLIER_LOWER_FRACTION × Mittel` fallen, gelten als Ausreißer
 * — ausgenommen das letzte Segment einer VOD-Playlist (laut Apple-
 * HLS-Authoring-Spec ist es üblich kürzer). Der Grenzwert ist im Doc
 * §7 fixiert.
 */
const OUTLIER_LOWER_FRACTION = 0.5;

/**
 * Apple-HLS-Authoring-Empfehlung: Live-Latenz ≈ 3 × TARGETDURATION.
 * Dokumentiert in `docs/user/stream-analyzer.md` §7.
 */
const LIVE_LATENCY_TARGET_MULTIPLIER = 3;

interface MediaParseResult {
  readonly details: MediaPlaylistDetails;
  readonly findings: readonly AnalysisFinding[];
}

interface PendingExtInf {
  readonly duration: number | null;
  readonly title?: string;
  readonly tagLine: number;
  readonly raw: string;
}

interface SegmentDraft {
  readonly uri: string;
  readonly duration: number;
  readonly title?: string;
  readonly resolvedUri?: string;
  readonly sequenceNumber: number;
  /** Für Toleranzregel: war die Dauer in EXTINF parsbar? */
  readonly durationParseable: boolean;
}

interface ParserState {
  targetDuration: number | undefined;
  mediaSequenceFromTag: number | undefined;
  playlistType: string | undefined;
  endList: boolean;
  pending: PendingExtInf | null;
  drafts: SegmentDraft[];
  features: {
    encryption: boolean;
    initSegment: boolean;
    discontinuity: boolean;
    programDateTime: boolean;
  };
}

export function parseMediaPlaylist(text: string, baseUrl: string | undefined): MediaParseResult {
  const lines = text.split(/\r?\n/);
  const findings: AnalysisFinding[] = [];
  const state: ParserState = {
    targetDuration: undefined,
    mediaSequenceFromTag: undefined,
    playlistType: undefined,
    endList: false,
    pending: null,
    drafts: [],
    features: {
      encryption: false,
      initSegment: false,
      discontinuity: false,
      programDateTime: false
    }
  };

  for (let lineIdx = 0; lineIdx < lines.length; lineIdx++) {
    processLine(lines[lineIdx].trim(), lineIdx, state, baseUrl, findings);
  }

  finalizeManifest(state, findings);
  return { details: buildDetails(state), findings };
}

/** Verarbeitet eine einzelne (getrimmte) Manifest-Zeile. Tag-Branches
 * mutieren `state`; Non-Tag-Lines werden als Segment-URI im Kontext der
 * letzten EXTINF interpretiert (oder als Stray-Finding). */
function processLine(
  line: string,
  lineIdx: number,
  state: ParserState,
  baseUrl: string | undefined,
  findings: AnalysisFinding[]
): void {
  if (line.length === 0 || line === "#EXTM3U") return;
  if (processGlobalTag(line, lineIdx, state, findings)) return;
  if (processFeatureTag(line, state)) return;
  if (line.startsWith(EXTINF_PREFIX)) {
    handleExtInfTag(line, lineIdx, state, findings);
    return;
  }
  if (line.startsWith("#")) return;
  handleSegmentUri(line, lineIdx, state, baseUrl, findings);
}

/** Top-Level-Tags: ENDLIST, TARGETDURATION, MEDIA-SEQUENCE, PLAYLIST-TYPE.
 * Liefert true, wenn die Zeile als globales Tag erkannt und konsumiert
 * wurde — sonst false (Caller probiert die nächste Tag-Klasse). */
function processGlobalTag(
  line: string,
  lineIdx: number,
  state: ParserState,
  findings: AnalysisFinding[]
): boolean {
  if (line === ENDLIST_TAG) {
    state.endList = true;
    return true;
  }
  if (line.startsWith(TARGETDURATION_PREFIX)) {
    const parsed = parseIntAttr(line.slice(TARGETDURATION_PREFIX.length).trim());
    if (parsed === null) {
      findings.push({
        code: "media_malformed_targetduration",
        level: "error",
        message: `EXT-X-TARGETDURATION auf Zeile ${lineIdx + 1} ist nicht parseable.`
      });
    } else {
      state.targetDuration = parsed;
    }
    return true;
  }
  if (line.startsWith(MEDIASEQUENCE_PREFIX)) {
    const parsed = parseIntAttr(line.slice(MEDIASEQUENCE_PREFIX.length).trim());
    if (parsed === null) {
      findings.push({
        code: "media_malformed_mediasequence",
        level: "warning",
        message: `EXT-X-MEDIA-SEQUENCE auf Zeile ${lineIdx + 1} ist nicht parseable; fallback auf 0.`
      });
    } else {
      state.mediaSequenceFromTag = parsed;
    }
    return true;
  }
  if (line.startsWith(PLAYLISTTYPE_PREFIX)) {
    state.playlistType = line.slice(PLAYLISTTYPE_PREFIX.length).trim();
    return true;
  }
  return false;
}

/** Reine Audit-Marker (Encryption, Init-Segment, Discontinuity,
 * Program-Date-Time). Beeinflussen Aggregate nicht; landen am Ende als
 * Info-Findings. */
function processFeatureTag(line: string, state: ParserState): boolean {
  if (line.startsWith(KEY_PREFIX)) {
    // METHOD=NONE deaktiviert eine zuvor gesetzte Verschlüsselung;
    // nur "echte" Methoden zählen als Feature.
    const payload = line.slice(KEY_PREFIX.length);
    if (!/METHOD\s*=\s*NONE/.test(payload)) {
      state.features.encryption = true;
    }
    return true;
  }
  if (line.startsWith(MAP_PREFIX)) {
    state.features.initSegment = true;
    return true;
  }
  if (line === DISCONTINUITY_TAG || line.startsWith(DISCONTINUITY_TAG + ":")) {
    state.features.discontinuity = true;
    return true;
  }
  if (line.startsWith(PROGRAM_DATE_TIME_PREFIX)) {
    state.features.programDateTime = true;
    return true;
  }
  return false;
}

function handleExtInfTag(
  line: string,
  lineIdx: number,
  state: ParserState,
  findings: AnalysisFinding[]
): void {
  if (state.pending !== null) {
    findings.push({
      code: "segment_missing_uri",
      level: "error",
      message: `EXTINF auf Zeile ${state.pending.tagLine + 1} hat keine darauffolgende URI-Zeile.`
    });
  }
  state.pending = parseExtInf(line.slice(EXTINF_PREFIX.length), lineIdx);
}

function handleSegmentUri(
  line: string,
  lineIdx: number,
  state: ParserState,
  baseUrl: string | undefined,
  findings: AnalysisFinding[]
): void {
  if (state.pending === null) {
    // Stray URI line ohne EXTINF — RFC 8216 verbietet das.
    findings.push({
      code: "segment_unexpected_uri",
      level: "warning",
      message: `URI auf Zeile ${lineIdx + 1} ohne vorhergehendes EXTINF.`
    });
    return;
  }
  const sequenceBase = state.mediaSequenceFromTag ?? 0;
  state.drafts.push(
    buildSegmentDraft(state.pending, line, baseUrl, sequenceBase + state.drafts.length, lineIdx, findings)
  );
  state.pending = null;
}

function finalizeManifest(state: ParserState, findings: AnalysisFinding[]): void {
  if (state.pending !== null) {
    findings.push({
      code: "segment_missing_uri",
      level: "error",
      message: `EXTINF auf Zeile ${state.pending.tagLine + 1} hat keine darauffolgende URI-Zeile.`
    });
  }
  if (state.targetDuration === undefined) {
    findings.push({
      code: "media_missing_targetduration",
      level: "error",
      message: "EXT-X-TARGETDURATION fehlt. RFC 8216 §4.3.3.1 macht das Tag verpflichtend."
    });
  }
  findings.push(...checkTargetViolations(state.drafts, state.targetDuration));
  findings.push(...checkDurationOutliers(state.drafts, state.endList, state.targetDuration));
  findings.push(...featureFindings(state.features));
}

function buildDetails(state: ParserState): MediaPlaylistDetails {
  const segments: MediaSegment[] = state.drafts.map((d) => ({
    uri: d.uri,
    duration: d.duration,
    sequenceNumber: d.sequenceNumber,
    ...(d.title !== undefined ? { title: d.title } : {}),
    ...(d.resolvedUri !== undefined ? { resolvedUri: d.resolvedUri } : {})
  }));
  const live = !state.endList;
  return {
    ...(state.targetDuration !== undefined ? { targetDuration: state.targetDuration } : {}),
    mediaSequence: state.mediaSequenceFromTag ?? 0,
    ...(state.playlistType !== undefined ? { playlistType: state.playlistType } : {}),
    endList: state.endList,
    live,
    ...(live && state.targetDuration !== undefined
      ? { liveLatencyEstimateSeconds: state.targetDuration * LIVE_LATENCY_TARGET_MULTIPLIER }
      : {}),
    segments,
    summary: buildSummary(segments)
  };
}

function parseExtInf(payload: string, lineIdx: number): PendingExtInf {
  // EXTINF:<duration>,<title> — title ist optional.
  const commaIdx = payload.indexOf(",");
  const durationPart = (commaIdx === -1 ? payload : payload.slice(0, commaIdx)).trim();
  const titlePart = commaIdx === -1 ? undefined : payload.slice(commaIdx + 1).trim();
  const duration = parseFloatAttr(durationPart);
  return {
    duration: duration === null ? null : duration,
    ...(titlePart !== undefined && titlePart.length > 0 ? { title: titlePart } : {}),
    tagLine: lineIdx,
    raw: payload
  };
}

function buildSegmentDraft(
  pending: PendingExtInf,
  uri: string,
  baseUrl: string | undefined,
  sequenceNumber: number,
  uriLineIdx: number,
  findings: AnalysisFinding[]
): SegmentDraft {
  let durationParseable = true;
  let duration = pending.duration;
  if (duration === null) {
    durationParseable = false;
    duration = 0;
    findings.push({
      code: "segment_malformed_extinf",
      level: "warning",
      message: `EXTINF auf Zeile ${pending.tagLine + 1} hat keine parsebare Dauer ("${pending.raw}").`
    });
  }
  const resolvedUri = resolveUri(uri, baseUrl);
  if (resolvedUri === null && baseUrl !== undefined) {
    findings.push({
      code: "segment_malformed_uri",
      level: "warning",
      message: `Segment-URI "${uri}" auf Zeile ${uriLineIdx + 1} konnte nicht gegen Base-URL aufgelöst werden.`
    });
  }
  return {
    uri,
    duration,
    ...(pending.title !== undefined ? { title: pending.title } : {}),
    ...(resolvedUri !== null ? { resolvedUri } : {}),
    sequenceNumber,
    durationParseable
  };
}

function checkTargetViolations(
  drafts: readonly SegmentDraft[],
  targetDuration: number | undefined
): AnalysisFinding[] {
  if (targetDuration === undefined) return [];
  const findings: AnalysisFinding[] = [];
  for (const d of drafts) {
    if (!d.durationParseable) continue;
    // RFC 8216 §4.3.3.1: round(duration) MUSS ≤ targetDuration sein.
    const rounded = Math.round(d.duration);
    if (rounded > targetDuration) {
      findings.push({
        code: "segment_duration_exceeds_target",
        level: "error",
        message: `Segment #${d.sequenceNumber} dauert ${d.duration.toFixed(3)} s (gerundet ${rounded} s) und überschreitet TARGETDURATION=${targetDuration}.`
      });
    }
  }
  return findings;
}

function checkDurationOutliers(
  drafts: readonly SegmentDraft[],
  endList: boolean,
  targetDuration: number | undefined
): AnalysisFinding[] {
  const parseable = drafts.filter((d) => d.durationParseable);
  if (parseable.length < 2) return [];
  const total = parseable.reduce((sum, d) => sum + d.duration, 0);
  const average = total / parseable.length;
  // Anker bevorzugt TARGETDURATION (Apple-HLS-Authoring-Guide), weil
  // ein Mean-Anker sich durch Ausreißer selbst absenkt. Fehlt das Tag,
  // greift der Mean-Fallback, damit Manifeste ohne TARGETDURATION
  // trotzdem geprüft werden können (auch wenn dieser Fall ohnehin
  // bereits media_missing_targetduration auslöst).
  const anchor = targetDuration ?? average;
  const anchorLabel = targetDuration !== undefined ? "TARGETDURATION" : "Mittel";
  const lowerBound = anchor * OUTLIER_LOWER_FRACTION;
  const findings: AnalysisFinding[] = [];
  for (let i = 0; i < parseable.length; i++) {
    const d = parseable[i];
    const isLast = i === parseable.length - 1;
    // Bei VOD ist ein kurzes letztes Segment normal (Datei endet auf
    // Sample-Grenze, nicht auf Segmentgrenze). Bei Live ist die Liste
    // sliding und das letzte Segment hat keinen Sonderstatus.
    if (isLast && endList) continue;
    if (d.duration < lowerBound) {
      findings.push({
        code: "segment_duration_outlier",
        level: "warning",
        message: `Segment #${d.sequenceNumber} dauert ${d.duration.toFixed(3)} s, unter ${(OUTLIER_LOWER_FRACTION * 100).toFixed(0)} % des Ankers (${anchorLabel}=${anchor.toFixed(3)} s).`
      });
    }
  }
  return findings;
}

function featureFindings(features: {
  encryption: boolean;
  initSegment: boolean;
  discontinuity: boolean;
  programDateTime: boolean;
}): AnalysisFinding[] {
  // Diese HLS-Features beeinflussen Tranche-4-Aggregate (Anzahl,
  // Dauer, Live/VOD) nicht — sie ändern aber, was der Analyzer NICHT
  // validiert. Eine Info-Finding pro Feature-Klasse signalisiert
  // Konsumenten den Audit-Stand, ohne Warning/Error-Counter zu
  // verfälschen.
  const findings: AnalysisFinding[] = [];
  if (features.encryption) {
    findings.push({
      code: "media_encryption_present",
      level: "info",
      message: "Manifest enthält EXT-X-KEY mit aktiver Methode; Analyzer validiert Schlüssel-/Decryption-Pfade nicht."
    });
  }
  if (features.initSegment) {
    findings.push({
      code: "media_init_segment_present",
      level: "info",
      message: "Manifest enthält EXT-X-MAP (fMP4-Init-Segment); Analyzer prüft Init-Segment-Konsistenz nicht."
    });
  }
  if (features.discontinuity) {
    findings.push({
      code: "media_discontinuity_present",
      level: "info",
      message: "Manifest enthält EXT-X-DISCONTINUITY; Timeline-Continuity wird nicht ausgewertet."
    });
  }
  if (features.programDateTime) {
    findings.push({
      code: "media_program_date_time_present",
      level: "info",
      message: "Manifest enthält EXT-X-PROGRAM-DATE-TIME; Wall-Clock-Annotationen werden nicht ausgewertet."
    });
  }
  return findings;
}

function buildSummary(segments: readonly MediaSegment[]): MediaSegmentSummary {
  if (segments.length === 0) {
    return { count: 0, averageDuration: 0, minDuration: 0, maxDuration: 0, totalDuration: 0 };
  }
  let total = 0;
  let min = Number.POSITIVE_INFINITY;
  let max = Number.NEGATIVE_INFINITY;
  for (const s of segments) {
    total += s.duration;
    if (s.duration < min) min = s.duration;
    if (s.duration > max) max = s.duration;
  }
  return {
    count: segments.length,
    averageDuration: total / segments.length,
    minDuration: min,
    maxDuration: max,
    totalDuration: total
  };
}

function resolveUri(rawUri: string, baseUrl: string | undefined): string | null {
  if (baseUrl === undefined) return null;
  try {
    return new URL(rawUri, baseUrl).toString();
  } catch {
    return null;
  }
}
