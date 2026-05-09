import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisSummary,
  CmafSignal,
  CmafSignalSummary,
  DashAdaptationSet,
  DashAnalysisResult,
  DashManifestDetails,
  DashRepresentation
} from "../../types/result.js";
import { AnalysisError } from "../../types/error.js";
import {
  buildDashAnchor,
  isFmp4DashUri,
  isMp4MimeType,
  resolveBaseUrlChain,
  resolveDashTemplate,
  resolveSegmentUri,
  type DashAnchorPath,
  type ResolvedBaseUrl
} from "./cmaf-dash.js";

/**
 * MPD-Parser für DASH-Manifests (`plan-0.9.0` §4 Tranche 3, RAK-58 /
 * NF-12). Der Parser extrahiert die Mindest-Felder aus
 * `MPD/Period/AdaptationSet/Representation/SegmentTemplate`-Hierarchie
 * pro `docs/planning/done/plan-0.9.0.md` §0.5:
 *
 *   - `details.profiles` (aus `MPD@profiles`)
 *   - `details.type` (`MPD@type` ∈ `static` / `dynamic`, Default `static`)
 *   - `details.live` (`type === "dynamic"`)
 *   - `details.adaptationSets[]` mit Representation-Mindest-Feldern
 *     (`mimeType`, `codecs`, `bandwidth`, `width`/`height`)
 *   - `summary.itemCount` = Gesamtzahl Representations
 *
 * Der Parser ist **bewusst regex-basiert**: keine externe XML-
 * Dependency, kein DOMParser-Pfad (Node-CLI muss ohne Browser-
 * APIs laufen). Trade-off: keine Validierung gegen das XSD-Schema,
 * keine Behandlung von `xs:include` / Entity-Expansion / DTD-
 * Tricks. Out-of-Scope laut Plan §0.3 (Live-Edge-Cases wie
 * `$Time$`-Variablen, `availabilityStartTime`-Drift) bleibt out
 * of scope; wenn diese später benötigt werden, ist die Migration
 * auf eine echte XML-Library (z. B. `fast-xml-parser`) ein
 * Folge-Plan-Item.
 */
export interface DashParseOutput {
  readonly details: DashManifestDetails;
  readonly findings: readonly AnalysisFinding[];
  /**
   * Interne CMAF-Metadaten für den Tranche-4-Binary-Pfad. Liegt
   * bewusst neben dem Public-Result, damit Tranche 4 keine
   * MPD-Tags erneut tokenisieren muss. Konsumenten bekommen das
   * Public-Summary über `details.cmaf`.
   */
  readonly cmafMeta?: DashCmafMetadata;
}

/**
 * Pflichtprüfungs-Eintrag pro AdaptationSet mit CMAF-Signal. Liefert
 * Tranche 4 alles, was sie zum Init- und Media-Segment-Fetch braucht.
 */
export interface DashCmafRepresentationEntry {
  readonly manifestAnchor: string;
  readonly periodIdx: number;
  readonly adaptationSetIdx: number;
  readonly representationIdx: number;
  readonly representationId: string;
  readonly bandwidth: number;
  readonly baseUrl?: string;
  readonly baseUrlBlocked: boolean;
  readonly init?: DashCmafSegmentRef;
  readonly media?: DashCmafSegmentRef;
  /**
   * Interne Marker, welche Pflichtreferenz fehlt — Tranche 4 mappt
   * sie auf `segment_reference_missing` bzw.
   * `dash_template_unresolved`.
   */
  readonly missingRefs: readonly DashMissingRef[];
}

export interface DashCmafSegmentRef {
  readonly source: "segment_template" | "segment_list";
  readonly attribute: "initialization" | "media" | "sourceURL";
  readonly rawTemplate: string;
  readonly resolvedUri?: string;
  readonly templateUnresolved: boolean;
}

export type DashMissingRef = "init" | "media" | "init_template_unresolved" | "media_template_unresolved";

export interface DashCmafMetadata {
  readonly representations: readonly DashCmafRepresentationEntry[];
}

export function analyzeDashManifestText(
  text: string,
  inputMeta: AnalysisInputMetadata,
  analyzerVersion: string
): DashAnalysisResult {
  const parsed = parseDashManifest(text, inputMeta.baseUrl);
  const summary: AnalysisSummary = {
    itemCount: parsed.details.adaptationSets.reduce(
      (acc, set) => acc + set.representations.length,
      0
    )
  };
  return {
    status: "ok",
    analyzerVersion,
    analyzerKind: "dash",
    input: inputMeta,
    summary,
    findings: parsed.findings,
    playlistType: "dash",
    details: parsed.details
  };
}

function parseDashManifest(text: string, manifestBaseUrl: string | undefined): DashParseOutput {
  const findings: AnalysisFinding[] = [];

  const mpdMatch = matchTag(text, "MPD");
  if (mpdMatch === null) {
    throw new AnalysisError(
      "internal_error",
      "DASH-Manifest enthält kein <MPD>-Wurzelelement.",
      { firstLine: text.slice(0, 80).trim() }
    );
  }

  const mpdAttrs = parseAttributes(mpdMatch.openTag);
  const typeAttr = mpdAttrs.get("type");
  const type: "static" | "dynamic" =
    typeAttr === "dynamic" ? "dynamic" : "static";
  if (typeAttr !== undefined && typeAttr !== "static" && typeAttr !== "dynamic") {
    findings.push({
      code: "dash_mpd_type_unknown",
      level: "warning",
      message: `MPD@type="${typeAttr}" ist weder "static" noch "dynamic"; der Analyzer interpretiert das Manifest konservativ als VOD ("static").`
    });
  }

  const periods = collectAll(mpdMatch.body, "Period");
  if (periods.length === 0) {
    findings.push({
      code: "dash_period_missing",
      level: "error",
      message: "MPD enthält keine <Period>-Elemente; der DASH-Stream ist nicht abspielbar."
    });
  }

  // BaseURL-Chain: MPD-Ebene gewinnt über `manifestBaseUrl` aus dem
  // URL-Loader, sobald sie sicher ist. Unsichere/blocked Werte auf
  // MPD-Ebene werden in `cmafMeta` für Tranche 4 sichtbar.
  const mpdBase = resolveBaseUrlChain(extractBaseURLs(mpdMatch.body), manifestBaseUrl);

  const adaptationSets: DashAdaptationSet[] = [];
  const cmafEntries: DashCmafRepresentationEntry[] = [];
  const cmafSignals: CmafSignal[] = [];
  collectPeriods(periods, mpdBase, findings, adaptationSets, cmafEntries, cmafSignals);

  if (adaptationSets.length > 0 &&
      adaptationSets.every((set) => set.representations.length === 0)) {
    findings.push({
      code: "dash_representation_missing",
      level: "error",
      message: "Keine <Representation>-Elemente in irgendeiner AdaptationSet gefunden."
    });
  }

  const cmafSummary = buildDashCmafSummary(cmafSignals);
  const details: DashManifestDetails = {
    profiles: mpdAttrs.get("profiles"),
    type,
    live: type === "dynamic",
    mediaPresentationDuration: mpdAttrs.get("mediaPresentationDuration"),
    minimumUpdatePeriod: mpdAttrs.get("minimumUpdatePeriod"),
    availabilityStartTime: mpdAttrs.get("availabilityStartTime"),
    periodCount: periods.length,
    adaptationSets,
    ...(cmafSummary !== undefined ? { cmaf: cmafSummary } : {})
  };

  return {
    details,
    findings,
    ...(cmafEntries.length > 0 ? { cmafMeta: { representations: cmafEntries } } : {})
  };
}

function collectPeriods(
  periods: readonly TagMatch[],
  mpdBase: ResolvedBaseUrl,
  findings: AnalysisFinding[],
  adaptationSets: DashAdaptationSet[],
  cmafEntries: DashCmafRepresentationEntry[],
  cmafSignals: CmafSignal[]
): void {
  for (let pIdx = 0; pIdx < periods.length; pIdx++) {
    const period = periods[pIdx];
    const periodAttrs = parseAttributes(period.openTag);
    const periodBase = resolveBaseUrlChain(extractBaseURLs(period.body), mpdBase.baseUrl);
    const periodAnchor: DashAnchorPath = {
      periodIdx: pIdx,
      ...(periodAttrs.get("id") !== undefined ? { periodId: periodAttrs.get("id") } : {})
    };
    const periodSets = collectAll(period.body, "AdaptationSet");
    for (let asIdx = 0; asIdx < periodSets.length; asIdx++) {
      adaptationSets.push(
        parseAdaptationSet(
          periodSets[asIdx],
          { ...periodAnchor, adaptationSetIdx: asIdx },
          periodBase,
          findings,
          cmafEntries,
          cmafSignals
        )
      );
    }
  }
}

interface SegmentTemplateData {
  readonly initialization?: string;
  readonly media?: string;
  readonly startNumber: number;
}

interface SegmentListData {
  readonly initializationSourceUrl?: string;
  readonly firstMediaUri?: string;
}

function parseAdaptationSet(
  setMatch: TagMatch,
  anchor: DashAnchorPath,
  parentBase: ResolvedBaseUrl,
  findings: AnalysisFinding[],
  cmafEntries: DashCmafRepresentationEntry[],
  cmafSignals: CmafSignal[]
): DashAdaptationSet {
  const setAttrs = parseAttributes(setMatch.openTag);
  const setMimeType = setAttrs.get("mimeType");
  const setCodecs = setAttrs.get("codecs");
  const setContentType = setAttrs.get("contentType");
  const setLang = setAttrs.get("lang");
  const setBase = resolveBaseUrlChain(extractBaseURLs(setMatch.body), parentBase.baseUrl);
  const setTemplate = parseSegmentTemplate(setMatch.body);
  const setList = parseSegmentList(setMatch.body);

  const setAnchor: DashAnchorPath = {
    ...anchor,
    ...(setAttrs.get("id") !== undefined ? { adaptationSetId: setAttrs.get("id") } : {})
  };

  const representations: DashRepresentation[] = [];
  const repMatches = collectAll(setMatch.body, "Representation");
  let chosenRepEntry: DashCmafRepresentationEntry | undefined;
  for (let rIdx = 0; rIdx < repMatches.length; rIdx++) {
    const rep = repMatches[rIdx];
    const result = parseRepresentation(
      rep,
      setMimeType,
      setCodecs,
      { ...setAnchor, representationIdx: rIdx },
      setBase,
      setTemplate,
      setList,
      findings
    );
    representations.push(result.representation);
    if (chosenRepEntry === undefined && result.cmafEntry !== undefined) {
      // Plan T3-DoD-Auswahl: pro AdaptationSet genau eine
      // Representation, bevorzugt erste mit eigener oder geerbter
      // Init+Media-Referenz, sonst die erste mit Signal überhaupt.
      if (result.cmafEntry.init !== undefined && result.cmafEntry.media !== undefined) {
        chosenRepEntry = result.cmafEntry;
      } else if (chosenRepEntry === undefined) {
        chosenRepEntry = result.cmafEntry;
      }
    }
  }
  if (chosenRepEntry !== undefined) {
    cmafEntries.push(chosenRepEntry);
    pushRepresentationSignals(chosenRepEntry, setMimeType, cmafSignals);
  } else if (isMp4MimeType(setMimeType)) {
    // AdaptationSet-Ebene-Signal: MP4-MIME ohne weitere Init/Media-
    // Referenzen ergibt nur ein inferred-Summary.
    cmafSignals.push({
      code: dashMimeSignalCode(setMimeType),
      level: "info",
      manifestAnchor: buildDashAnchor(setAnchor, "@mimeType"),
      confidence: "inferred"
    });
  }

  return {
    id: setAttrs.get("id"),
    mimeType: setMimeType,
    codecs: setCodecs,
    contentType: setContentType,
    lang: setLang,
    representations
  };
}

interface RepresentationParseResult {
  readonly representation: DashRepresentation;
  readonly cmafEntry?: DashCmafRepresentationEntry;
}

function parseRepresentation(
  repMatch: TagMatch,
  inheritedMimeType: string | undefined,
  inheritedCodecs: string | undefined,
  anchor: DashAnchorPath,
  parentBase: ResolvedBaseUrl,
  parentTemplate: SegmentTemplateData | undefined,
  parentList: SegmentListData | undefined,
  findings: AnalysisFinding[]
): RepresentationParseResult {
  const repAttrs = parseAttributes(repMatch.openTag);
  const id = repAttrs.get("id") ?? "";
  const bandwidth = checkRequiredRepresentationAttrs(repAttrs, id, findings);
  const repBase = resolveBaseUrlChain(extractBaseURLs(repMatch.body), parentBase.baseUrl);
  const finalMimeType = repAttrs.get("mimeType") ?? inheritedMimeType;
  const finalCodecs = repAttrs.get("codecs") ?? inheritedCodecs;
  const repAnchor: DashAnchorPath = {
    ...anchor,
    ...(id.length > 0 ? { representationId: id } : {})
  };

  const cmafEntry = buildRepresentationCmafEntry(
    repAnchor,
    id,
    bandwidth,
    finalMimeType,
    repBase,
    mergeSegmentTemplates(parentTemplate, parseSegmentTemplate(repMatch.body)),
    parseSegmentList(repMatch.body) ?? parentList
  );

  return {
    representation: buildRepresentationDetails(repAttrs, id, bandwidth, finalMimeType, finalCodecs),
    ...(cmafEntry !== undefined ? { cmafEntry } : {})
  };
}

function checkRequiredRepresentationAttrs(
  repAttrs: Map<string, string>,
  id: string,
  findings: AnalysisFinding[]
): number {
  if (id === "") {
    findings.push({
      code: "dash_representation_missing_id",
      level: "warning",
      message: "<Representation> ohne id-Attribut; der Eintrag wird mit leerer id aufgenommen."
    });
  }
  const bandwidthRaw = repAttrs.get("bandwidth");
  if (bandwidthRaw === undefined) {
    findings.push({
      code: "dash_representation_missing_bandwidth",
      level: "error",
      message: `<Representation id="${id}"> ohne bandwidth; bandwidth ist laut MPEG-DASH §5.3.5 Pflicht.`
    });
    return 0;
  }
  return parseIntegerAttr(bandwidthRaw);
}

function buildRepresentationDetails(
  repAttrs: Map<string, string>,
  id: string,
  bandwidth: number,
  mimeType: string | undefined,
  codecs: string | undefined
): DashRepresentation {
  const widthRaw = repAttrs.get("width");
  const heightRaw = repAttrs.get("height");
  const width = widthRaw === undefined ? undefined : parseIntegerAttr(widthRaw);
  const height = heightRaw === undefined ? undefined : parseIntegerAttr(heightRaw);
  return {
    id,
    bandwidth,
    width: width !== undefined && Number.isFinite(width) ? width : undefined,
    height: height !== undefined && Number.isFinite(height) ? height : undefined,
    frameRate: repAttrs.get("frameRate"),
    codecs,
    mimeType,
    audioSamplingRate: repAttrs.get("audioSamplingRate")
  };
}

function parseIntegerAttr(value: string): number {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : 0;
}

/**
 * Extrahiert alle direkten `<BaseURL>...</BaseURL>`-Einträge aus
 * einem Element-Body. Reihenfolge bleibt erhalten — die erste
 * sichere `BaseURL` gewinnt laut Plan T3-DoD.
 */
function extractBaseURLs(body: string): string[] {
  const results: string[] = [];
  const pattern = /<BaseURL\b[^>]*?>([\s\S]*?)<\/BaseURL\s*>/g;
  let m: RegExpExecArray | null;
  while ((m = pattern.exec(body)) !== null) {
    results.push(decodeXmlEntities(m[1]).trim());
  }
  return results;
}

function parseSegmentTemplate(body: string): SegmentTemplateData | undefined {
  const matches = collectAll(body, "SegmentTemplate");
  if (matches.length === 0) return undefined;
  const attrs = parseAttributes(matches[0].openTag);
  const startNumberRaw = attrs.get("startNumber");
  const startNumber = startNumberRaw === undefined ? 1 : parseIntegerAttr(startNumberRaw);
  return {
    initialization: attrs.get("initialization"),
    media: attrs.get("media"),
    startNumber: startNumber > 0 ? startNumber : 1
  };
}

function mergeSegmentTemplates(
  parent: SegmentTemplateData | undefined,
  own: SegmentTemplateData | undefined
): SegmentTemplateData | undefined {
  if (parent === undefined && own === undefined) return undefined;
  if (parent === undefined) return own;
  if (own === undefined) return parent;
  return {
    initialization: own.initialization ?? parent.initialization,
    media: own.media ?? parent.media,
    startNumber: own.startNumber !== 1 ? own.startNumber : parent.startNumber
  };
}

function parseSegmentList(body: string): SegmentListData | undefined {
  const matches = collectAll(body, "SegmentList");
  if (matches.length === 0) return undefined;
  const list = matches[0];
  const initMatches = collectAll(list.body, "Initialization");
  const initSourceUrl =
    initMatches.length > 0 ? parseAttributes(initMatches[0].openTag).get("sourceURL") : undefined;
  const segmentMatches = collectAll(list.body, "SegmentURL");
  const firstMediaUri =
    segmentMatches.length > 0 ? parseAttributes(segmentMatches[0].openTag).get("media") : undefined;
  if (initSourceUrl === undefined && firstMediaUri === undefined) return undefined;
  return {
    ...(initSourceUrl !== undefined ? { initializationSourceUrl: initSourceUrl } : {}),
    ...(firstMediaUri !== undefined ? { firstMediaUri } : {})
  };
}

/**
 * Stellt aus einer `DashCmafRepresentationEntry` die Public-Signal-
 * Einträge für `details.cmaf.signals[]` her. Die stärkste Confidence
 * gewinnt: explizite Init- oder Media-Referenz (`segment_template`/
 * `segment_list`) → `manifest`; reines `mimeType`-Indiz oder
 * fMP4-Suffix-Heuristik im Template → `inferred`.
 */
function pushRepresentationSignals(
  entry: DashCmafRepresentationEntry,
  mimeType: string | undefined,
  out: CmafSignal[]
): void {
  if (entry.init !== undefined && !entry.init.templateUnresolved) {
    out.push({
      code:
        entry.init.source === "segment_template"
          ? "dash_segment_template_initialization"
          : "dash_segment_list_initialization",
      level: "info",
      manifestAnchor: `${entry.manifestAnchor}/${entry.init.source === "segment_template" ? "SegmentTemplate" : "SegmentList/Initialization"}@${entry.init.attribute}`,
      confidence: "manifest"
    });
  }
  if (entry.media !== undefined && !entry.media.templateUnresolved) {
    const mediaCode =
      entry.media.source === "segment_template"
        ? "dash_segment_template_media"
        : "dash_segment_list_media";
    out.push({
      code: mediaCode,
      level: "info",
      manifestAnchor: `${entry.manifestAnchor}/${entry.media.source === "segment_template" ? "SegmentTemplate" : "SegmentList/SegmentURL"}@${entry.media.attribute}`,
      confidence: "manifest"
    });
    if (isFmp4DashUri(entry.media.rawTemplate)) {
      out.push({
        code: "dash_segment_extension_fmp4",
        level: "info",
        manifestAnchor: `${entry.manifestAnchor}/${entry.media.source === "segment_template" ? "SegmentTemplate" : "SegmentList/SegmentURL"}@${entry.media.attribute}`,
        confidence: "manifest"
      });
    }
  }
  if (isMp4MimeType(mimeType)) {
    out.push({
      code: dashMimeSignalCode(mimeType),
      level: "info",
      manifestAnchor: `${entry.manifestAnchor}/@mimeType`,
      confidence: entry.init !== undefined || entry.media !== undefined ? "manifest" : "inferred"
    });
  }
}

function dashMimeSignalCode(mimeType: string | undefined): string {
  if (mimeType === undefined) return "dash_mime_mp4";
  const lc = mimeType.toLowerCase();
  if (lc === "video/mp4") return "dash_mime_video_mp4";
  if (lc === "audio/mp4") return "dash_mime_audio_mp4";
  if (lc === "application/mp4") return "dash_mime_application_mp4";
  return "dash_mime_mp4";
}

function buildRepresentationCmafEntry(
  anchor: DashAnchorPath,
  representationId: string,
  bandwidth: number,
  mimeType: string | undefined,
  base: ResolvedBaseUrl,
  template: SegmentTemplateData | undefined,
  list: SegmentListData | undefined
): DashCmafRepresentationEntry | undefined {
  const init = resolveInitRef(anchor, representationId, bandwidth, base, template, list);
  const media = resolveMediaRef(anchor, representationId, bandwidth, base, template, list);
  const hasMimeSignal = isMp4MimeType(mimeType);
  if (init === undefined && media === undefined && !hasMimeSignal) {
    return undefined;
  }
  const missingRefs: DashMissingRef[] = [];
  if (init === undefined) missingRefs.push("init");
  else if (init.templateUnresolved) missingRefs.push("init_template_unresolved");
  if (media === undefined) missingRefs.push("media");
  else if (media.templateUnresolved) missingRefs.push("media_template_unresolved");
  return {
    manifestAnchor: buildDashAnchor(anchor),
    periodIdx: anchor.periodIdx,
    adaptationSetIdx: anchor.adaptationSetIdx ?? 0,
    representationIdx: anchor.representationIdx ?? 0,
    representationId,
    bandwidth,
    ...(base.baseUrl !== undefined ? { baseUrl: base.baseUrl } : {}),
    baseUrlBlocked: base.blocked,
    ...(init !== undefined ? { init } : {}),
    ...(media !== undefined ? { media } : {}),
    missingRefs
  };
}

function resolveInitRef(
  anchor: DashAnchorPath,
  representationId: string,
  bandwidth: number,
  base: ResolvedBaseUrl,
  template: SegmentTemplateData | undefined,
  list: SegmentListData | undefined
): DashCmafSegmentRef | undefined {
  if (template?.initialization !== undefined) {
    return resolveTemplateRef(
      "segment_template",
      "initialization",
      template.initialization,
      template.startNumber,
      representationId,
      bandwidth,
      base.baseUrl
    );
  }
  if (list?.initializationSourceUrl !== undefined) {
    return resolveLiteralRef(
      "segment_list",
      "sourceURL",
      list.initializationSourceUrl,
      base.baseUrl
    );
  }
  // Anchor wird derzeit nicht für die fehlende Referenz benötigt;
  // Tranche 4 baut den Skip-Eintrag aus dem `representationCmafEntry`-
  // Anker direkt.
  void anchor;
  return undefined;
}

function resolveMediaRef(
  anchor: DashAnchorPath,
  representationId: string,
  bandwidth: number,
  base: ResolvedBaseUrl,
  template: SegmentTemplateData | undefined,
  list: SegmentListData | undefined
): DashCmafSegmentRef | undefined {
  if (template?.media !== undefined) {
    return resolveTemplateRef(
      "segment_template",
      "media",
      template.media,
      template.startNumber,
      representationId,
      bandwidth,
      base.baseUrl
    );
  }
  if (list?.firstMediaUri !== undefined) {
    return resolveLiteralRef("segment_list", "media", list.firstMediaUri, base.baseUrl);
  }
  void anchor;
  return undefined;
}

function resolveTemplateRef(
  source: "segment_template",
  attribute: "initialization" | "media",
  rawTemplate: string,
  startNumber: number,
  representationId: string,
  bandwidth: number,
  baseUrl: string | undefined
): DashCmafSegmentRef {
  const expanded = resolveDashTemplate(rawTemplate, {
    representationId,
    bandwidth,
    number: startNumber
  });
  if (expanded === null) {
    return { source, attribute, rawTemplate, templateUnresolved: true };
  }
  const resolvedUri = resolveSegmentUri(expanded, baseUrl);
  return {
    source,
    attribute,
    rawTemplate,
    templateUnresolved: false,
    ...(resolvedUri !== null ? { resolvedUri } : {})
  };
}

function resolveLiteralRef(
  source: "segment_list",
  attribute: "sourceURL" | "media",
  rawTemplate: string,
  baseUrl: string | undefined
): DashCmafSegmentRef {
  const resolvedUri = resolveSegmentUri(rawTemplate, baseUrl);
  return {
    source,
    attribute,
    rawTemplate,
    templateUnresolved: false,
    ...(resolvedUri !== null ? { resolvedUri } : {})
  };
}

/**
 * Aggregiert die pro Representation gesammelten Signale plus
 * eventuelle MP4-MIME-only-Indikationen zu einem
 * `CmafSignalSummary`. Tranche 3 emittiert kein `binary`-Objekt —
 * das setzt Tranche 4 mit dem bounded Loader.
 */
function buildDashCmafSummary(
  signals: readonly CmafSignal[]
): CmafSignalSummary | undefined {
  if (signals.length === 0) return undefined;
  let confidence: CmafSignalSummary["confidence"] = "inferred";
  for (const s of signals) {
    if (s.confidence === "manifest") confidence = "manifest";
  }
  return {
    source: "dash",
    confidence,
    signals: [...signals]
  };
}

interface TagMatch {
  /** Eröffnungs-Tag inklusive Attribute (`<Foo attr="bar">` oder
   * `<Foo attr="bar"/>`). */
  readonly openTag: string;
  /** Inhalt zwischen Eröffnungs- und Schluss-Tag (leerer String bei
   * Self-Closing-Tags). */
  readonly body: string;
}

function matchTag(text: string, name: string): TagMatch | null {
  const matches = collectAll(text, name);
  return matches.length === 0 ? null : matches[0];
}

function collectAll(text: string, name: string): TagMatch[] {
  const results: TagMatch[] = [];
  const escapedName = name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  // Match either self-closing (<Name ... />) or paired (<Name ...>...</Name>).
  // Non-greedy body match; XML allows nested same-name elements only
  // for AdaptationSet/Representation if the schema explicitly nests
  // them — DASH spec doesn't, so non-greedy is safe.
  const pattern = new RegExp(
    `<${escapedName}\\b([^>]*?)\\s*/>|<${escapedName}\\b([^>]*?)>([\\s\\S]*?)</${escapedName}\\s*>`,
    "g"
  );
  let m: RegExpExecArray | null;
  while ((m = pattern.exec(text)) !== null) {
    if (m[0].endsWith("/>")) {
      results.push({
        openTag: `<${name}${m[1] ?? ""}/>`,
        body: ""
      });
    } else {
      results.push({
        openTag: `<${name}${m[2] ?? ""}>`,
        body: m[3] ?? ""
      });
    }
  }
  return results;
}

const ATTR_PATTERN = /([A-Za-z_][A-Za-z0-9_:.-]*)\s*=\s*(?:"([^"]*)"|'([^']*)')/g;

function parseAttributes(openTag: string): Map<string, string> {
  const attrs = new Map<string, string>();
  // Strip the element name and the closing `>` / `/>`.
  const nameEnd = openTag.search(/[\s/>]/);
  if (nameEnd === -1) {
    return attrs;
  }
  const attrText = openTag.slice(nameEnd, openTag.length - 1).replace(/\/$/, "");
  let m: RegExpExecArray | null;
  while ((m = ATTR_PATTERN.exec(attrText)) !== null) {
    const value = m[2] ?? m[3] ?? "";
    attrs.set(m[1], decodeXmlEntities(value));
  }
  return attrs;
}

function decodeXmlEntities(value: string): string {
  return value
    .replaceAll("&amp;", "&")
    .replaceAll("&lt;", "<")
    .replaceAll("&gt;", ">")
    .replaceAll("&quot;", '"')
    .replaceAll("&apos;", "'");
}
