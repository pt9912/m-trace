import type { AnalysisFinding } from "../../types/finding.js";
import type {
  CmafSignal,
  CmafSignalSummary,
  MediaPlaylistDetails,
  MediaSegment,
  MediaSegmentSummary
} from "../../types/result.js";
import { parseAttributeList, parseFloatAttr, parseIntAttr } from "./attrs.js";
import {
  aggregateConfidence,
  isFmp4SegmentUri,
  mediaAnchor,
  parseByteRangePayload,
  type HlsByteRange,
  type HlsExtXMapMeta,
  type HlsFirstMediaSegmentMeta,
  type HlsMediaCmafMetadata
} from "./cmaf-hls.js";

const TARGETDURATION_PREFIX = "#EXT-X-TARGETDURATION:";
const MEDIASEQUENCE_PREFIX = "#EXT-X-MEDIA-SEQUENCE:";
const PLAYLISTTYPE_PREFIX = "#EXT-X-PLAYLIST-TYPE:";
const EXTINF_PREFIX = "#EXTINF:";
const ENDLIST_TAG = "#EXT-X-ENDLIST";
const KEY_PREFIX = "#EXT-X-KEY:";
const MAP_PREFIX = "#EXT-X-MAP:";
const BYTERANGE_PREFIX = "#EXT-X-BYTERANGE:";
const INDEPENDENT_SEGMENTS_TAG = "#EXT-X-INDEPENDENT-SEGMENTS";
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
  /**
   * Interne CMAF-Metadaten für den Tranche-4-Binary-Pfad. Liegt nur
   * bewusst neben dem Public-Result, damit Tranche 4 keine Manifest-
   * zeilen erneut tokenisieren muss. Konsumenten bekommen das
   * Public-Summary über `details.cmaf`.
   */
  readonly cmafMeta?: HlsMediaCmafMetadata;
}

interface PendingExtInf {
  readonly duration: number | null;
  readonly title?: string;
  readonly tagLine: number;
  readonly raw: string;
}

interface PendingByteRange {
  readonly value: HlsByteRange;
  readonly tagLine: number;
}

interface SegmentDraft {
  readonly uri: string;
  readonly duration: number;
  readonly title?: string;
  readonly resolvedUri?: string;
  readonly sequenceNumber: number;
  /** Für Toleranzregel: war die Dauer in EXTINF parsbar? */
  readonly durationParseable: boolean;
  /** Stabile 1-basierte Zeile der URI für CMAF-Anker. */
  readonly uriLine: number;
  /**
   * Strukturierter `#EXT-X-BYTERANGE`-Wert, der unmittelbar vor der
   * Segment-URI gesetzt wurde. Tranche 4 nutzt ihn für `skipped`-
   * Mapping mit `hls_media_byterange_unsupported`.
   */
  readonly byterange?: HlsByteRange;
  /** Zeilennummer der `#EXT-X-BYTERANGE`-Zeile (1-basiert), falls vorhanden. */
  readonly byterangeLine?: number;
  /**
   * `true`, wenn die URI ein fMP4-Segment-Suffix (`.m4s`/`.cmfv`/
   * `.cmfa`) trägt. Schwacher CMAF-Hinweis, im Manifest-Anker
   * sichtbar.
   */
  readonly isFmp4Uri: boolean;
}

interface ExtXMapDraft {
  readonly attrs: Map<string, string>;
  readonly tagLine: number;
}

interface ParserState {
  targetDuration: number | undefined;
  mediaSequenceFromTag: number | undefined;
  playlistType: string | undefined;
  endList: boolean;
  pending: PendingExtInf | null;
  pendingByteRange: PendingByteRange | null;
  drafts: SegmentDraft[];
  features: {
    encryption: boolean;
    initSegment: boolean;
    discontinuity: boolean;
    programDateTime: boolean;
    independentSegments: boolean;
  };
  cmaf: {
    extXMap: ExtXMapDraft | null;
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
    pendingByteRange: null,
    drafts: [],
    features: {
      encryption: false,
      initSegment: false,
      discontinuity: false,
      programDateTime: false,
      independentSegments: false
    },
    cmaf: {
      extXMap: null
    }
  };

  for (let lineIdx = 0; lineIdx < lines.length; lineIdx++) {
    processLine(lines[lineIdx].trim(), lineIdx, state, baseUrl, findings);
  }

  finalizeManifest(state, findings);
  const cmafBuild = buildCmaf(state, baseUrl);
  return {
    details: buildDetails(state, cmafBuild?.summary),
    findings,
    ...(cmafBuild?.metadata !== undefined ? { cmafMeta: cmafBuild.metadata } : {})
  };
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
  if (processFeatureTag(line, lineIdx, state, findings)) return;
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
 * Info-Findings. EXT-X-MAP, EXT-X-INDEPENDENT-SEGMENTS und
 * #EXT-X-BYTERANGE werden zusätzlich strukturiert erfasst, damit der
 * Tranche-4-Binary-Pfad keine Manifestzeilen erneut tokenisieren muss
 * (`0.10.0` Tranche 2, NF-13 / RAK-61). */
function processFeatureTag(
  line: string,
  lineIdx: number,
  state: ParserState,
  findings: AnalysisFinding[]
): boolean {
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
    handleExtXMapTag(line, lineIdx, state, findings);
    return true;
  }
  if (line === INDEPENDENT_SEGMENTS_TAG) {
    state.features.independentSegments = true;
    return true;
  }
  if (line.startsWith(BYTERANGE_PREFIX)) {
    handleByteRangeTag(line, lineIdx, state, findings);
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

/**
 * Strukturierte Erfassung von `#EXT-X-MAP` (RFC 8216 §4.3.2.5).
 * Pflicht-Attribut ist `URI`; `BYTERANGE` ist optional. In `0.10.0`
 * werden nur die letzten beiden Werte für den Binary-Pfad verwendet;
 * eventuell vorhandene `KEYFORMAT`/`KEYFORMATVERSIONS` bleiben in
 * `rawAttributes` verfügbar.
 */
function handleExtXMapTag(
  line: string,
  lineIdx: number,
  state: ParserState,
  findings: AnalysisFinding[]
): void {
  const attrs = parseAttributeList(line.slice(MAP_PREFIX.length));
  const uri = attrs.get("URI");
  if (uri === undefined || uri.length === 0) {
    findings.push({
      code: "media_map_missing_uri",
      level: "error",
      message: `EXT-X-MAP auf Zeile ${lineIdx + 1} hat kein URI-Attribut.`
    });
    return;
  }
  state.cmaf.extXMap = { attrs, tagLine: lineIdx };
}

/**
 * Bindet `#EXT-X-BYTERANGE` (RFC 8216 §4.3.2.2) an die nächste
 * Segment-URI. Mehrere BYTERANGE-Tags vor derselben URI sind ein
 * Spec-Verstoß und werden als Warning gemeldet — der zweite Wert
 * gewinnt, weil er der Manifestzeile am nächsten ist.
 */
function handleByteRangeTag(
  line: string,
  lineIdx: number,
  state: ParserState,
  findings: AnalysisFinding[]
): void {
  const payload = line.slice(BYTERANGE_PREFIX.length);
  const parsed = parseByteRangePayload(payload);
  if (parsed === null) {
    findings.push({
      code: "media_byterange_malformed",
      level: "warning",
      message: `EXT-X-BYTERANGE auf Zeile ${lineIdx + 1} ist nicht parseable ("${payload.trim()}").`
    });
    return;
  }
  if (state.pendingByteRange !== null) {
    findings.push({
      code: "media_byterange_duplicate",
      level: "warning",
      message: `EXT-X-BYTERANGE auf Zeile ${lineIdx + 1} überschreibt einen vorherigen Wert (Zeile ${state.pendingByteRange.tagLine + 1}).`
    });
  }
  state.pendingByteRange = { value: parsed, tagLine: lineIdx };
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
    state.pendingByteRange = null;
    return;
  }
  const sequenceBase = state.mediaSequenceFromTag ?? 0;
  state.drafts.push(
    buildSegmentDraft(
      state.pending,
      state.pendingByteRange,
      line,
      baseUrl,
      sequenceBase + state.drafts.length,
      lineIdx,
      findings
    )
  );
  state.pending = null;
  state.pendingByteRange = null;
}

function finalizeManifest(state: ParserState, findings: AnalysisFinding[]): void {
  if (state.pending !== null) {
    findings.push({
      code: "segment_missing_uri",
      level: "error",
      message: `EXTINF auf Zeile ${state.pending.tagLine + 1} hat keine darauffolgende URI-Zeile.`
    });
  }
  if (state.pendingByteRange !== null) {
    findings.push({
      code: "media_byterange_orphan",
      level: "warning",
      message: `EXT-X-BYTERANGE auf Zeile ${state.pendingByteRange.tagLine + 1} hat keine darauffolgende Segment-URI.`
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

function buildDetails(
  state: ParserState,
  cmaf: CmafSignalSummary | undefined
): MediaPlaylistDetails {
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
    summary: buildSummary(segments),
    ...(cmaf !== undefined ? { cmaf } : {})
  };
}

interface CmafBuildResult {
  readonly summary: CmafSignalSummary;
  readonly metadata?: HlsMediaCmafMetadata;
}

/**
 * Baut `details.cmaf`-Summary plus interne Tranche-4-Metadaten aus
 * dem Parser-State (`0.10.0` Tranche 2, NF-13 / RAK-61). Ohne ein
 * starkes (`EXT-X-MAP`) oder schwaches (`.m4s`/`.cmfv`/`.cmfa`-URI)
 * CMAF-Signal wird kein `cmaf` emittiert — Negativ-Fixtures bleiben
 * byte-kompatibel zum `0.9.x`-Stand.
 *
 * Tranche 2 erzeugt **kein** `binary`-Objekt. Status `passed`/
 * `failed`/`skipped` entstehen erst in Tranche 4 mit dem bounded
 * Segment-Loader; bis dahin ist `binary` schlicht abwesend (das
 * Schema lässt das ausdrücklich zu, weil `binary?:` optional ist).
 */
function buildCmaf(
  state: ParserState,
  baseUrl: string | undefined
): CmafBuildResult | undefined {
  const signals: CmafSignal[] = [];

  const initMeta = buildInitSegmentMeta(state.cmaf.extXMap, baseUrl);
  if (initMeta !== undefined) {
    pushInitSignals(signals, initMeta, state.cmaf.extXMap!.tagLine);
  }

  const firstFmp4 = state.drafts.find((d) => d.isFmp4Uri);
  if (firstFmp4 !== undefined) {
    signals.push({
      code: "hls_segment_extension_fmp4",
      level: "info",
      manifestAnchor: mediaAnchor(firstFmp4.uriLine),
      confidence: state.cmaf.extXMap !== null ? "manifest" : "inferred"
    });
  }

  if (state.features.independentSegments && (initMeta !== undefined || firstFmp4 !== undefined)) {
    signals.push({
      code: "hls_independent_segments",
      level: "info",
      manifestAnchor: "media:tag:#EXT-X-INDEPENDENT-SEGMENTS",
      confidence: "inferred"
    });
  }

  if (signals.length === 0) return undefined;

  const summary: CmafSignalSummary = {
    source: "hls",
    confidence: aggregateConfidence(signals),
    signals
  };

  const firstMediaMeta = firstFmp4 !== undefined ? buildFirstMediaMeta(firstFmp4) : undefined;
  const metadata: HlsMediaCmafMetadata = {
    ...(initMeta !== undefined ? { initSegment: initMeta } : {}),
    ...(firstMediaMeta !== undefined ? { firstMediaSegment: firstMediaMeta } : {})
  };

  return {
    summary,
    ...(initMeta !== undefined || firstMediaMeta !== undefined ? { metadata } : {})
  };
}

/**
 * Strukturierte Sicht auf `EXT-X-MAP` für Tranche 4: URI,
 * optionale BYTERANGE, gegen `baseUrl` aufgelöste URI und
 * Roh-Attribute. Liefert `undefined`, wenn der Parser den Tag
 * nicht erfolgreich tokenisiert hat (URI leer / fehlend).
 */
function buildInitSegmentMeta(
  draft: ExtXMapDraft | null,
  baseUrl: string | undefined
): HlsExtXMapMeta | undefined {
  if (draft === null) return undefined;
  const uri = draft.attrs.get("URI") ?? "";
  if (uri.length === 0) return undefined;
  const byterangeRaw = draft.attrs.get("BYTERANGE");
  const byterange = byterangeRaw !== undefined ? parseByteRangePayload(byterangeRaw) : null;
  const resolvedUri = resolveUri(uri, baseUrl);
  const rawAttributes: Record<string, string> = {};
  for (const [k, v] of draft.attrs) rawAttributes[k] = v;
  return {
    uri,
    ...(byterange !== null ? { byterange } : {}),
    ...(resolvedUri !== null ? { resolvedUri } : {}),
    rawAttributes,
    manifestAnchor: mediaAnchor(draft.tagLine)
  };
}

function pushInitSignals(
  signals: CmafSignal[],
  init: HlsExtXMapMeta,
  tagLine: number
): void {
  signals.push({
    code: "hls_ext_x_map",
    level: "info",
    manifestAnchor: mediaAnchor(tagLine),
    confidence: "manifest"
  });
  if (init.byterange !== undefined) {
    signals.push({
      code: "hls_ext_x_map_byterange",
      level: "info",
      manifestAnchor: mediaAnchor(tagLine),
      confidence: "manifest"
    });
  }
}

function buildFirstMediaMeta(draft: SegmentDraft): HlsFirstMediaSegmentMeta {
  return {
    uri: draft.uri,
    ...(draft.resolvedUri !== undefined ? { resolvedUri: draft.resolvedUri } : {}),
    ...(draft.byterange !== undefined ? { byterange: draft.byterange } : {}),
    sequenceNumber: draft.sequenceNumber,
    manifestAnchor: mediaAnchor(draft.uriLine)
  };
}

function parseExtInf(payload: string, lineIdx: number): PendingExtInf {
  // EXTINF:<duration>,<title> — title ist optional.
  const commaIdx = payload.indexOf(",");
  const durationPart = (commaIdx === -1 ? payload : payload.slice(0, commaIdx)).trim();
  const titlePart = commaIdx === -1 ? undefined : payload.slice(commaIdx + 1).trim();
  const duration = parseFloatAttr(durationPart);
  return {
    duration,
    ...(titlePart !== undefined && titlePart.length > 0 ? { title: titlePart } : {}),
    tagLine: lineIdx,
    raw: payload
  };
}

function buildSegmentDraft(
  pending: PendingExtInf,
  pendingByteRange: PendingByteRange | null,
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
    durationParseable,
    uriLine: uriLineIdx,
    ...(pendingByteRange !== null
      ? { byterange: pendingByteRange.value, byterangeLine: pendingByteRange.tagLine }
      : {}),
    isFmp4Uri: isFmp4SegmentUri(uri)
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
