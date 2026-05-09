import type { CmafBinaryOptions } from "../../types/input.js";
import type {
  CmafBinaryVerification,
  CmafBoxAnchor,
  CmafFailure,
  CmafFailureCode,
  CmafLimits,
  CmafSegmentCheck,
  CmafSignalSummary
} from "../../types/result.js";
import {
  validateCmafInit,
  validateCmafMediaFragment,
  type IsoBmffParseError
} from "./iso-bmff.js";
import type { LoaderRuntime } from "../loader/runtime.js";
import { loadSegment, type SegmentLoadResult } from "./segment-loader.js";
import type {
  DashCmafMetadata,
  DashCmafRepresentationEntry,
  DashCmafSegmentRef
} from "../parsers/dash.js";
import type {
  HlsExtXMapMeta,
  HlsFirstMediaSegmentMeta,
  HlsMediaCmafMetadata
} from "../parsers/cmaf-hls.js";

/**
 * Orchestrator für die binäre CMAF-Konformitätsprüfung
 * (`0.10.0` Tranche 4, NF-13 / RAK-64).
 *
 * Nimmt das interne CMAF-Manifest-Metadaten-Objekt vom HLS-Media-
 * oder DASH-Parser entgegen, wendet `cmaf.binary`-Optionen an,
 * lädt Init- und Media-Segmente über den bounded Segment-Loader,
 * parst Boxes mit dem ISO-BMFF-Reader und baut ein
 * `CmafBinaryVerification`-Objekt mit:
 *
 *  - `status` — Aggregation `failed > skipped > passed`.
 *  - `segmentsChecked[]` — Pflichtprüfungs-Einträge mit Einzel-
 *    status und Failure-Code.
 *  - `boxes[]` — eindeutige Box-Anker (kein Re-Parsing).
 *  - `failures[]` — strukturierte Skip-/Failure-Begründungen.
 *  - `limits` — `requiredSegmentChecks`, `plannedSegmentFetches`
 *    plus die durchgereichten Loader-Grenzen.
 *
 * Failure-Code-Präzedenz aus Plan §3 Tranche 1:
 *   1. `binary_disabled`
 *   2. `segment_reference_missing` / `dash_template_unresolved` /
 *      `hls_map_byterange_unsupported` / `hls_media_byterange_unsupported`
 *   3. `not_planned_due_to_limit`
 *   4. `segment_base_url_missing` / `segment_uri_blocked`
 *   5. `segment_fetch_failed` / `segment_content_type_unsupported` /
 *      `segment_too_large`
 *   6. `cmaf_box_validation_failed` / `invalid_box_structure`
 */

export interface BinaryVerifyOptions {
  readonly cmafBinary: Required<CmafBinaryOptions>;
  readonly timeoutMs: number;
  readonly maxRedirects: number;
  readonly allowPrivateNetworks: boolean;
  readonly runtime: LoaderRuntime;
  /**
   * Trust-Anker für Segment-Auflösung im Text-Pfad (oder finale
   * Manifest-URL im URL-Pfad). Wenn `undefined`, kann der Verifier
   * relative Segment-URIs nicht auflösen — Pflichtprüfungen
   * bekommen `segment_base_url_missing`.
   */
  readonly baseUrl: string | undefined;
}

/**
 * Eingabe für die HLS-Media-Pfad-Verifikation. Tranche 4 prüft pro
 * Media-Playlist genau Init-Segment + erstes fMP4-Media-Segment;
 * fehlende Referenzen werden mit
 * `segment_reference_missing` gemeldet.
 */
export interface HlsBinaryVerifyInput {
  readonly source: "hls";
  readonly mediaCmaf: HlsMediaCmafMetadata | undefined;
}

/**
 * Eingabe für den DASH-Pfad: pro AdaptationSet mit CMAF-Signal eine
 * Representation mit Init+/Media-Referenz.
 */
export interface DashBinaryVerifyInput {
  readonly source: "dash";
  readonly dashCmaf: DashCmafMetadata | undefined;
}

export type BinaryVerifyInput = HlsBinaryVerifyInput | DashBinaryVerifyInput;

export async function verifyBinaryCmaf(
  input: BinaryVerifyInput,
  options: BinaryVerifyOptions
): Promise<CmafBinaryVerification> {
  const plan = buildPlan(input);
  if (!options.cmafBinary.enabled) {
    return buildDisabledVerification(plan, options);
  }
  return executePlan(plan, options);
}

interface PlannedCheck {
  readonly kind: "init" | "media";
  readonly source: "hls" | "dash";
  readonly manifestAnchor: string;
  /** Skip-Code, der unabhängig vom Loader gilt (manifestseitige Lücke). */
  readonly preFailureCode?: CmafFailureCode;
  readonly preFailureMessage?: string;
  /** Roh-URI bzw. Template (für Anchors / Audit). */
  readonly uri?: string;
  /** Aufgelöste sichere HTTP(S)-URL für den Loader, wenn vorhanden. */
  readonly resolvedUrl?: string;
}

interface VerificationPlan {
  readonly checks: readonly PlannedCheck[];
  readonly requiredSegmentChecks: number;
}

function buildPlan(input: BinaryVerifyInput): VerificationPlan {
  if (input.source === "hls") {
    return buildHlsPlan(input.mediaCmaf);
  }
  return buildDashPlan(input.dashCmaf);
}

function buildHlsPlan(meta: HlsMediaCmafMetadata | undefined): VerificationPlan {
  const checks: PlannedCheck[] = [];
  // HLS: genau ein Init- und ein Media-Pflichtcheck pro Media-Playlist,
  // sofern Manifest-Signale vorliegen. Fehlt eine Referenz, bleibt
  // sie als skipped sichtbar — kein "stiller Erfolg".
  if (meta === undefined) {
    return { checks: [], requiredSegmentChecks: 0 };
  }
  checks.push(buildHlsInitCheck(meta.initSegment));
  checks.push(buildHlsMediaCheck(meta.firstMediaSegment));
  return { checks, requiredSegmentChecks: checks.length };
}

function buildHlsInitCheck(init: HlsExtXMapMeta | undefined): PlannedCheck {
  if (init === undefined) {
    return {
      kind: "init",
      source: "hls",
      manifestAnchor: "media:ext_x_map",
      preFailureCode: "segment_reference_missing",
      preFailureMessage: "HLS-Media-Playlist ohne EXT-X-MAP — kein Init-Segment ableitbar."
    };
  }
  if (init.byterange !== undefined) {
    return {
      kind: "init",
      source: "hls",
      manifestAnchor: init.manifestAnchor,
      uri: init.uri,
      preFailureCode: "hls_map_byterange_unsupported",
      preFailureMessage: "EXT-X-MAP mit BYTERANGE wird in 0.10.0 nicht per HTTP Range geladen."
    };
  }
  return {
    kind: "init",
    source: "hls",
    manifestAnchor: init.manifestAnchor,
    uri: init.uri,
    ...(init.resolvedUri !== undefined ? { resolvedUrl: init.resolvedUri } : {})
  };
}

function buildHlsMediaCheck(media: HlsFirstMediaSegmentMeta | undefined): PlannedCheck {
  if (media === undefined) {
    return {
      kind: "media",
      source: "hls",
      manifestAnchor: "media:first_fmp4_segment",
      preFailureCode: "segment_reference_missing",
      preFailureMessage: "HLS-Media-Playlist ohne fMP4-Media-Segment ableitbar."
    };
  }
  if (media.byterange !== undefined) {
    return {
      kind: "media",
      source: "hls",
      manifestAnchor: media.manifestAnchor,
      uri: media.uri,
      preFailureCode: "hls_media_byterange_unsupported",
      preFailureMessage: "#EXT-X-BYTERANGE auf erstem fMP4-Media-Segment wird in 0.10.0 nicht per HTTP Range geladen."
    };
  }
  return {
    kind: "media",
    source: "hls",
    manifestAnchor: media.manifestAnchor,
    uri: media.uri,
    ...(media.resolvedUri !== undefined ? { resolvedUrl: media.resolvedUri } : {})
  };
}

function buildDashPlan(meta: DashCmafMetadata | undefined): VerificationPlan {
  if (meta === undefined || meta.representations.length === 0) {
    // Wenn nur ein MP4-MIME-Signal vorliegt (kein cmafEntry), wird
    // ein einzelner Pflichtcheck "init" mit segment_reference_missing
    // erzeugt, damit binary nicht weggelassen wird.
    return {
      checks: [
        {
          kind: "init",
          source: "dash",
          manifestAnchor: "dash:no_segment_reference",
          preFailureCode: "segment_reference_missing",
          preFailureMessage: "DASH-CMAF-Signal ohne ableitbare Init-/Media-Segment-Referenz."
        }
      ],
      requiredSegmentChecks: 1
    };
  }
  const checks: PlannedCheck[] = [];
  for (const rep of meta.representations) {
    checks.push(buildDashRefCheck(rep, "init", rep.init));
    checks.push(buildDashRefCheck(rep, "media", rep.media));
  }
  return { checks, requiredSegmentChecks: checks.length };
}

function buildDashRefCheck(
  rep: DashCmafRepresentationEntry,
  kind: "init" | "media",
  ref: DashCmafSegmentRef | undefined
): PlannedCheck {
  const anchorBase = rep.manifestAnchor;
  if (ref === undefined) {
    return {
      kind,
      source: "dash",
      manifestAnchor: `${anchorBase}/${kind === "init" ? "Initialization" : "Media"}`,
      preFailureCode: "segment_reference_missing",
      preFailureMessage: `Representation ohne ${kind === "init" ? "Init" : "Media"}-Referenz.`
    };
  }
  const refAnchor = `${anchorBase}/${ref.source === "segment_template" ? "SegmentTemplate" : "SegmentList"}@${ref.attribute}`;
  if (ref.templateUnresolved) {
    return {
      kind,
      source: "dash",
      manifestAnchor: refAnchor,
      uri: ref.rawTemplate,
      preFailureCode: "dash_template_unresolved",
      preFailureMessage: `DASH-Template-Variable nicht im 0.10.0-Scope auflösbar: ${ref.rawTemplate}`
    };
  }
  if (rep.baseUrlBlocked) {
    return {
      kind,
      source: "dash",
      manifestAnchor: refAnchor,
      uri: ref.rawTemplate,
      preFailureCode: "segment_uri_blocked",
      preFailureMessage: "BaseURL-Chain in dieser Ebene führt nur unsichere/Nicht-HTTP(S)-Werte."
    };
  }
  if (ref.resolvedUri === undefined) {
    return {
      kind,
      source: "dash",
      manifestAnchor: refAnchor,
      uri: ref.rawTemplate,
      preFailureCode: "segment_base_url_missing",
      preFailureMessage: "Keine sichere HTTP(S)-baseUrl als Trust-Anker für Segment-Fetch vorhanden."
    };
  }
  return {
    kind,
    source: "dash",
    manifestAnchor: refAnchor,
    uri: ref.rawTemplate,
    resolvedUrl: ref.resolvedUri
  };
}

function buildDisabledVerification(
  plan: VerificationPlan,
  options: BinaryVerifyOptions
): CmafBinaryVerification {
  const checks: CmafSegmentCheck[] = plan.checks.map((c) => ({
    kind: c.kind,
    source: c.source,
    manifestAnchor: c.manifestAnchor,
    ...(c.uri !== undefined ? { uri: c.uri } : {}),
    status: "skipped" as const,
    failureCode: "binary_disabled" as const,
    message: "Binary-Prüfung per cmaf.binary.enabled=false deaktiviert."
  }));
  const failures: CmafFailure[] = checks.length > 0
    ? [
        {
          code: "binary_disabled",
          level: "info",
          message: "Binary-Prüfung deaktiviert."
        }
      ]
    : [];
  // Wenn der Plan-Pfad „kein Pflichtcheck" liefert (z. B. HLS-Master),
  // wird trotzdem ein binary-Objekt erzeugt — dort dürfen aber keine
  // segmentsChecked entstehen.
  return {
    status: "skipped",
    segmentsChecked: checks,
    boxes: [],
    failures,
    limits: buildLimits(plan, 0, options)
  };
}

async function executePlan(
  plan: VerificationPlan,
  options: BinaryVerifyOptions
): Promise<CmafBinaryVerification> {
  const checks: CmafSegmentCheck[] = [];
  const failures: CmafFailure[] = [];
  const boxes: CmafBoxAnchor[] = [];
  const cap = options.cmafBinary.maxBinarySegments;
  let plannedFetches = 0;
  for (const planned of plan.checks) {
    const result = await runCheck(planned, plannedFetches, cap, options, boxes, failures);
    checks.push(result.check);
    if (result.fetched) plannedFetches += 1;
  }
  return {
    status: aggregateStatus(checks, failures),
    segmentsChecked: checks,
    boxes,
    failures,
    limits: buildLimits(plan, plannedFetches, options)
  };
}

interface CheckRunResult {
  readonly check: CmafSegmentCheck;
  readonly fetched: boolean;
}

async function runCheck(
  planned: PlannedCheck,
  plannedFetchesSoFar: number,
  cap: number,
  options: BinaryVerifyOptions,
  boxes: CmafBoxAnchor[],
  failures: CmafFailure[]
): Promise<CheckRunResult> {
  if (planned.preFailureCode !== undefined) {
    const message = planned.preFailureMessage ?? "Pflichtprüfung manifestseitig nicht ableitbar.";
    return { check: skip(planned, planned.preFailureCode, message, failures), fetched: false };
  }
  if (plannedFetchesSoFar >= cap) {
    return {
      check: skip(
        planned,
        "not_planned_due_to_limit",
        `Pflichtprüfung über maxBinarySegments=${cap} hinaus nicht geplant.`,
        failures
      ),
      fetched: false
    };
  }
  if (planned.resolvedUrl === undefined) {
    return {
      check: skip(
        planned,
        "segment_base_url_missing",
        "Keine sichere HTTP(S)-baseUrl als Trust-Anker für Segment-Fetch vorhanden.",
        failures
      ),
      fetched: false
    };
  }
  const fetchResult = await loadSegment(planned.resolvedUrl, {
    runtime: options.runtime,
    timeoutMs: options.timeoutMs,
    maxSegmentBytes: options.cmafBinary.maxSegmentBytes,
    maxRedirects: options.maxRedirects,
    allowPrivateNetworks: options.allowPrivateNetworks
  });
  if (!fetchResult.ok) {
    return {
      check: skip(planned, fetchResult.code, fetchResult.message, failures),
      fetched: true
    };
  }
  return {
    check: validateBytes(planned, fetchResult, boxes, failures),
    fetched: true
  };
}

function validateBytes(
  planned: PlannedCheck,
  fetchResult: Extract<SegmentLoadResult, { ok: true }>,
  boxes: CmafBoxAnchor[],
  failures: CmafFailure[]
): CmafSegmentCheck {
  const validation =
    planned.kind === "init"
      ? validateCmafInit(fetchResult.bytes)
      : validateCmafMediaFragment(fetchResult.bytes);
  const checkBoxes = mapAnchors(planned, validation.anchors);
  for (const a of checkBoxes) boxes.push(a);
  if (validation.ok) {
    return {
      kind: planned.kind,
      source: planned.source,
      manifestAnchor: planned.manifestAnchor,
      ...(planned.uri !== undefined ? { uri: planned.uri } : {}),
      resolvedUrl: fetchResult.finalUrl,
      status: "passed",
      contentType: fetchResult.contentType,
      bytesRead: fetchResult.bytes.byteLength,
      boxes: checkBoxes
    };
  }
  const code = pickValidationCode(validation.errors);
  for (const err of validation.errors) {
    failures.push({
      code: err.code,
      level: "error",
      message: err.message,
      manifestAnchor: planned.manifestAnchor,
      segmentAnchor: planned.kind === "init" ? "segment:init" : "segment:media[0]",
      ...(err.boxPath !== undefined ? { boxPath: err.boxPath } : {})
    });
  }
  return {
    kind: planned.kind,
    source: planned.source,
    manifestAnchor: planned.manifestAnchor,
    ...(planned.uri !== undefined ? { uri: planned.uri } : {}),
    resolvedUrl: fetchResult.finalUrl,
    status: "failed",
    failureCode: code,
    message: validation.errors.map((e) => e.message).join(" | "),
    contentType: fetchResult.contentType,
    bytesRead: fetchResult.bytes.byteLength,
    boxes: checkBoxes
  };
}

function pickValidationCode(errors: readonly IsoBmffParseError[]): CmafFailureCode {
  for (const err of errors) {
    if (err.code === "invalid_box_structure") return "invalid_box_structure";
  }
  return "cmaf_box_validation_failed";
}

function mapAnchors(
  planned: PlannedCheck,
  anchors: ReadonlyArray<{ readonly path: string; readonly type: string; readonly offset: number; readonly size: number }>
): CmafBoxAnchor[] {
  const segmentAnchor = planned.kind === "init" ? "segment:init" : "segment:media[0]";
  return anchors.map((a) => ({
    segmentAnchor,
    path: `${segmentAnchor}:${a.path}`,
    type: a.type,
    offset: a.offset,
    size: a.size
  }));
}

function skip(
  planned: PlannedCheck,
  code: CmafFailureCode,
  message: string,
  failures: CmafFailure[]
): CmafSegmentCheck {
  failures.push({
    code,
    level: code === "binary_disabled" ? "info" : "warning",
    message,
    manifestAnchor: planned.manifestAnchor
  });
  return {
    kind: planned.kind,
    source: planned.source,
    manifestAnchor: planned.manifestAnchor,
    ...(planned.uri !== undefined ? { uri: planned.uri } : {}),
    status: "skipped",
    failureCode: code,
    message
  };
}

function aggregateStatus(
  checks: readonly CmafSegmentCheck[],
  failures: readonly CmafFailure[]
): "passed" | "failed" | "skipped" {
  // Tranche-1-Regel: Fehler vor Skip vor Pass.
  // - failed sobald ein Pflichtcheck fehlgeschlagen ist.
  // - passed nur, wenn alle Pflichtchecks bestanden wurden.
  // - skipped sonst.
  if (checks.length === 0 && failures.length === 0) return "skipped";
  if (checks.some((c) => c.status === "failed")) return "failed";
  if (checks.length > 0 && checks.every((c) => c.status === "passed")) return "passed";
  return "skipped";
}

function buildLimits(
  plan: VerificationPlan,
  plannedFetches: number,
  options: BinaryVerifyOptions
): CmafLimits {
  return {
    maxSegmentBytes: options.cmafBinary.maxSegmentBytes,
    maxBinarySegments: options.cmafBinary.maxBinarySegments,
    timeoutMs: options.timeoutMs,
    maxRedirects: options.maxRedirects,
    requiredSegmentChecks: plan.requiredSegmentChecks,
    plannedSegmentFetches: plannedFetches
  };
}

/**
 * Hebt die Summary-Confidence auf `binary`, wenn der Binary-Status
 * `passed` ist — analog der Plan §3-Regel:
 * Summary-`confidence:"binary"` entsteht nur, wenn
 * `binary.status:"passed"`.
 */
export function withBinaryConfidenceUpgrade(
  summary: CmafSignalSummary,
  binary: CmafBinaryVerification
): CmafSignalSummary {
  if (binary.status !== "passed") {
    return { ...summary, binary };
  }
  return { ...summary, confidence: "binary", binary };
}
