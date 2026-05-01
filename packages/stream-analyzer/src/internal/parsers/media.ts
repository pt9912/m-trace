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

export function parseMediaPlaylist(text: string, baseUrl: string | undefined): MediaParseResult {
  const lines = text.split(/\r?\n/);
  const findings: AnalysisFinding[] = [];

  let targetDuration: number | undefined;
  let mediaSequenceFromTag: number | undefined;
  let playlistType: string | undefined;
  let endList = false;
  let pending: PendingExtInf | null = null;
  const drafts: SegmentDraft[] = [];

  for (let lineIdx = 0; lineIdx < lines.length; lineIdx++) {
    const line = lines[lineIdx].trim();
    if (line.length === 0) continue;
    if (line === "#EXTM3U") continue;
    if (line === ENDLIST_TAG) {
      endList = true;
      continue;
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
        targetDuration = parsed;
      }
      continue;
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
        mediaSequenceFromTag = parsed;
      }
      continue;
    }

    if (line.startsWith(PLAYLISTTYPE_PREFIX)) {
      playlistType = line.slice(PLAYLISTTYPE_PREFIX.length).trim();
      continue;
    }

    if (line.startsWith(EXTINF_PREFIX)) {
      if (pending !== null) {
        findings.push({
          code: "segment_missing_uri",
          level: "error",
          message: `EXTINF auf Zeile ${pending.tagLine + 1} hat keine darauffolgende URI-Zeile.`
        });
      }
      pending = parseExtInf(line.slice(EXTINF_PREFIX.length), lineIdx);
      continue;
    }

    if (line.startsWith("#")) continue;

    if (pending === null) {
      // Stray URI line ohne EXTINF — RFC 8216 verbietet das.
      findings.push({
        code: "segment_unexpected_uri",
        level: "warning",
        message: `URI auf Zeile ${lineIdx + 1} ohne vorhergehendes EXTINF.`
      });
      continue;
    }

    const sequenceBase = mediaSequenceFromTag ?? 0;
    const draft = buildSegmentDraft(pending, line, baseUrl, sequenceBase + drafts.length, lineIdx, findings);
    drafts.push(draft);
    pending = null;
  }

  if (pending !== null) {
    findings.push({
      code: "segment_missing_uri",
      level: "error",
      message: `EXTINF auf Zeile ${pending.tagLine + 1} hat keine darauffolgende URI-Zeile.`
    });
  }

  if (targetDuration === undefined) {
    findings.push({
      code: "media_missing_targetduration",
      level: "error",
      message: "EXT-X-TARGETDURATION fehlt. RFC 8216 §4.3.3.1 macht das Tag verpflichtend."
    });
  }

  findings.push(...checkTargetViolations(drafts, targetDuration));
  findings.push(...checkDurationOutliers(drafts, endList));

  const segments: MediaSegment[] = drafts.map((d) => ({
    uri: d.uri,
    duration: d.duration,
    sequenceNumber: d.sequenceNumber,
    ...(d.title !== undefined ? { title: d.title } : {}),
    ...(d.resolvedUri !== undefined ? { resolvedUri: d.resolvedUri } : {})
  }));

  const summary = buildSummary(segments);
  const live = !endList;
  const details: MediaPlaylistDetails = {
    ...(targetDuration !== undefined ? { targetDuration } : {}),
    mediaSequence: mediaSequenceFromTag ?? 0,
    ...(playlistType !== undefined ? { playlistType } : {}),
    endList,
    live,
    ...(live && targetDuration !== undefined
      ? { liveLatencyEstimateSeconds: targetDuration * LIVE_LATENCY_TARGET_MULTIPLIER }
      : {}),
    segments,
    summary
  };
  return { details, findings };
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
  endList: boolean
): AnalysisFinding[] {
  const parseable = drafts.filter((d) => d.durationParseable);
  if (parseable.length < 2) return [];
  const total = parseable.reduce((sum, d) => sum + d.duration, 0);
  const average = total / parseable.length;
  const lowerBound = average * OUTLIER_LOWER_FRACTION;
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
        message: `Segment #${d.sequenceNumber} dauert ${d.duration.toFixed(3)} s, unter ${(OUTLIER_LOWER_FRACTION * 100).toFixed(0)} % des Mittels (${average.toFixed(3)} s).`
      });
    }
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
