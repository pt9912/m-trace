import type { CmafSignal, CmafSignalSummary } from "../../types/result.js";

/**
 * HLS-spezifische CMAF-Detection (`0.10.0` Tranche 2, NF-13 / RAK-61).
 * Entscheidet aus den Manifest-Signalen, ob ein `details.cmaf`-Summary
 * emittiert wird, und liefert eine interne, schon strukturierte
 * Sicht auf `EXT-X-MAP` und `#EXT-X-BYTERANGE` für die Tranche-4-
 * Binary-Prüfung. Die internen Felder bleiben absichtlich aus
 * `MediaPlaylistDetails`/`MasterPlaylistDetails` raus — Konsumenten
 * bekommen nur das öffentliche `cmaf`-Summary.
 */

/**
 * `EXT-X-MAP`-Attribute (RFC 8216 §4.3.2.5) strukturiert für den
 * Binary-Pfad. Roh-Attribute bleiben mit dabei, damit Tranche 4
 * keine Re-Tokenisierung der Manifestzeile braucht.
 */
export interface HlsExtXMapMeta {
  readonly uri: string;
  readonly byterange?: HlsByteRange;
  readonly resolvedUri?: string;
  readonly rawAttributes: Readonly<Record<string, string>>;
  /** Stabiler Anker, z. B. `media:line:7` (1-basierte Zeilennummer). */
  readonly manifestAnchor: string;
}

/**
 * Strukturierte Sicht auf `#EXT-X-BYTERANGE` (RFC 8216 §4.3.2.2):
 * `length` ist Pflicht; `offset` optional. Der Roh-String bleibt
 * für Audit-Diagnostik verfügbar.
 */
export interface HlsByteRange {
  readonly length: number;
  readonly offset?: number;
  readonly raw: string;
}

/**
 * Strukturierte Sicht auf das erste fMP4-Media-Segment einer
 * Media-Playlist. Wird benötigt, weil der Tranche-4-Binary-Pfad
 * pro Media-Playlist genau Init-Segment + erstes fMP4-Media-Segment
 * prüft.
 */
export interface HlsFirstMediaSegmentMeta {
  readonly uri: string;
  readonly resolvedUri?: string;
  readonly byterange?: HlsByteRange;
  readonly sequenceNumber: number;
  readonly manifestAnchor: string;
}

/**
 * Interne Metadaten, die der Media-Playlist-Parser zusätzlich zu
 * `MediaPlaylistDetails` zurückliefert. Tranche 4 konsumiert sie als
 * Eingabe für den Binary-Loader.
 */
export interface HlsMediaCmafMetadata {
  readonly initSegment?: HlsExtXMapMeta;
  readonly firstMediaSegment?: HlsFirstMediaSegmentMeta;
}

const FMP4_EXTENSIONS = [".m4s", ".cmfv", ".cmfa"] as const;

/**
 * Erkennt fMP4-Segment-URI-Endungen. Die Allowlist deckt die in
 * `0.10.0` als CMAF-Hinweis gewerteten Erweiterungen ab; Casing
 * wird ignoriert. Query-Strings und Fragmente werden vor dem
 * Suffix-Check entfernt, weil viele Manifeste `?token=...` an die
 * Segment-URI hängen.
 */
export function isFmp4SegmentUri(uri: string): boolean {
  if (uri.length === 0) return false;
  const cleaned = stripUriSuffix(uri).toLowerCase();
  return FMP4_EXTENSIONS.some((ext) => cleaned.endsWith(ext));
}

function stripUriSuffix(uri: string): string {
  const queryIdx = uri.indexOf("?");
  const fragmentIdx = uri.indexOf("#");
  let cut = uri.length;
  if (queryIdx !== -1) cut = Math.min(cut, queryIdx);
  if (fragmentIdx !== -1) cut = Math.min(cut, fragmentIdx);
  return uri.slice(0, cut);
}

/**
 * Parst eine `#EXT-X-BYTERANGE`-Payload (`<length>[@<offset>]`)
 * laut RFC 8216 §4.3.2.2. Liefert `null` bei syntaktisch ungültigen
 * Werten.
 */
export function parseByteRangePayload(payload: string): HlsByteRange | null {
  const trimmed = payload.trim();
  if (trimmed.length === 0) return null;
  const atIdx = trimmed.indexOf("@");
  const lengthStr = atIdx === -1 ? trimmed : trimmed.slice(0, atIdx);
  const offsetStr = atIdx === -1 ? null : trimmed.slice(atIdx + 1);
  const length = parsePositiveInteger(lengthStr);
  if (length === null) return null;
  if (offsetStr !== null) {
    const offset = parsePositiveInteger(offsetStr);
    if (offset === null) return null;
    return { length, offset, raw: trimmed };
  }
  return { length, raw: trimmed };
}

function parsePositiveInteger(value: string): number | null {
  const trimmed = value.trim();
  if (!/^\d+$/.test(trimmed)) return null;
  const parsed = Number.parseInt(trimmed, 10);
  if (!Number.isFinite(parsed) || parsed < 0) return null;
  return parsed;
}

/**
 * Manifest-Anker im Format `media:line:<1-basierte Zeile>` für
 * `details.cmaf.signals[].manifestAnchor` und Tranche-4-Box-Anker.
 */
export function mediaAnchor(lineIdx0: number): string {
  return `media:line:${lineIdx0 + 1}`;
}

export function masterAnchor(lineIdx0: number): string {
  return `master:line:${lineIdx0 + 1}`;
}

/**
 * Aggregiert die stärkste Confidence über mehrere Signal-Einträge.
 * Ordnung `binary > manifest > inferred`. Tranche 2 erzeugt nie
 * `binary`-Confidence — die entsteht erst, wenn Tranche 4 das
 * `binary`-Objekt mit `status:"passed"` setzt.
 */
export function aggregateConfidence(
  signals: readonly CmafSignal[]
): CmafSignalSummary["confidence"] {
  let best: CmafSignalSummary["confidence"] = "inferred";
  for (const s of signals) {
    if (s.confidence === "binary") return "binary";
    if (s.confidence === "manifest") best = "manifest";
  }
  return best;
}
