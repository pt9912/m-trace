import type { AnalysisFinding } from "../../types/finding.js";
import type {
  AnalysisInputMetadata,
  AnalysisSummary,
  DashAdaptationSet,
  DashAnalysisResult,
  DashManifestDetails,
  DashRepresentation
} from "../../types/result.js";
import { AnalysisError } from "../../types/error.js";

/**
 * MPD-Parser für DASH-Manifests (`plan-0.9.0` §4 Tranche 3, RAK-58 /
 * NF-12). Der Parser extrahiert die Mindest-Felder aus
 * `MPD/Period/AdaptationSet/Representation/SegmentTemplate`-Hierarchie
 * pro `docs/planning/in-progress/plan-0.9.0.md` §0.5:
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
}

export function analyzeDashManifestText(
  text: string,
  inputMeta: AnalysisInputMetadata,
  analyzerVersion: string
): DashAnalysisResult {
  const parsed = parseDashManifest(text);
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

function parseDashManifest(text: string): DashParseOutput {
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

  const adaptationSets: DashAdaptationSet[] = [];
  for (const period of periods) {
    const periodSets = collectAll(period.body, "AdaptationSet");
    for (const set of periodSets) {
      adaptationSets.push(parseAdaptationSet(set, findings));
    }
  }

  if (adaptationSets.length > 0 &&
      adaptationSets.every((set) => set.representations.length === 0)) {
    findings.push({
      code: "dash_representation_missing",
      level: "error",
      message: "Keine <Representation>-Elemente in irgendeiner AdaptationSet gefunden."
    });
  }

  const details: DashManifestDetails = {
    profiles: mpdAttrs.get("profiles"),
    type,
    live: type === "dynamic",
    mediaPresentationDuration: mpdAttrs.get("mediaPresentationDuration"),
    minimumUpdatePeriod: mpdAttrs.get("minimumUpdatePeriod"),
    availabilityStartTime: mpdAttrs.get("availabilityStartTime"),
    periodCount: periods.length,
    adaptationSets
  };

  return { details, findings };
}

function parseAdaptationSet(
  setMatch: TagMatch,
  findings: AnalysisFinding[]
): DashAdaptationSet {
  const setAttrs = parseAttributes(setMatch.openTag);
  const setMimeType = setAttrs.get("mimeType");
  const setCodecs = setAttrs.get("codecs");
  const setContentType = setAttrs.get("contentType");
  const setLang = setAttrs.get("lang");

  const representations: DashRepresentation[] = [];
  const repMatches = collectAll(setMatch.body, "Representation");
  for (const rep of repMatches) {
    representations.push(parseRepresentation(rep, setMimeType, setCodecs, findings));
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

function parseRepresentation(
  repMatch: TagMatch,
  inheritedMimeType: string | undefined,
  inheritedCodecs: string | undefined,
  findings: AnalysisFinding[]
): DashRepresentation {
  const repAttrs = parseAttributes(repMatch.openTag);
  const id = repAttrs.get("id") ?? "";
  if (id === "") {
    findings.push({
      code: "dash_representation_missing_id",
      level: "warning",
      message: "<Representation> ohne id-Attribut; der Eintrag wird mit leerer id aufgenommen."
    });
  }
  const bandwidthRaw = repAttrs.get("bandwidth");
  const bandwidth = bandwidthRaw === undefined ? 0 : parseIntegerAttr(bandwidthRaw);
  if (bandwidthRaw === undefined) {
    findings.push({
      code: "dash_representation_missing_bandwidth",
      level: "error",
      message: `<Representation id="${id}"> ohne bandwidth; bandwidth ist laut MPEG-DASH §5.3.5 Pflicht.`
    });
  }
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
    codecs: repAttrs.get("codecs") ?? inheritedCodecs,
    mimeType: repAttrs.get("mimeType") ?? inheritedMimeType,
    audioSamplingRate: repAttrs.get("audioSamplingRate")
  };
}

function parseIntegerAttr(value: string): number {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : 0;
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
