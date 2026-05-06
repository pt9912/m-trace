/**
 * WebRTC-Adapter-Fehlercode-Taxonomie für `0.8.0` Tranche 2.
 *
 * Bezug: `docs/planning/done/plan-0.8.0.md` §3 Tranche 2 DoD —
 * "WebRTC-Fehlercode-Taxonomie ist im Contract dokumentiert und wird
 * vom Adapter genutzt".
 *
 * Reservierter Meta-Key: `webrtc.error_code`. Der Adapter setzt diesen
 * Key auf einem `playback_error`-Event nur, wenn der Fehler dem WebRTC-
 * Pfad zuzuordnen ist; freie Strings sind verboten und werden via
 * {@link normalizeWebRtcErrorCode} auf den Fallback `peer_connection_failed`
 * abgebildet, damit das Dashboard-/Releasing-Abnahme-Surface
 * maschinenlesbar bleibt.
 *
 * **Tranche-2-Scope**: Liste und Normalisierung sind SDK-intern. Der
 * Wire-Vertrag (formal in `contracts/event-schema.json` /
 * `spec/backend-api-contract.md`) wird in Tranche 3 zusammen mit dem
 * `webrtc.*`-Meta-Namespace produktiv gepinnt.
 */

/**
 * Liste der erlaubten WebRTC-Fehlercodes. Reihenfolge entspricht dem
 * typischen Lebenszyklus eines WHEP-Handshakes.
 */
export const WEBRTC_ERROR_CODES = [
  "whep_signaling_failed",
  "whep_sdp_invalid",
  "webrtc_no_tracks",
  "peer_connection_failed",
  "webrtc_destroyed_before_connected"
] as const;

export type WebRtcErrorCode = (typeof WEBRTC_ERROR_CODES)[number];

/** Reservierter Meta-Key für WebRTC-Fehlercodes auf `playback_error`-Events. */
export const WEBRTC_ERROR_CODE_META_KEY = "webrtc.error_code";

/**
 * Normalisiert einen Eingabewert auf die `WebRtcErrorCode`-Allowlist.
 * Unbekannte Strings, leere Strings und Nicht-Strings werden auf
 * `peer_connection_failed` abgebildet — das ist der "generische"
 * Fallback für unerwartete WebRTC-Probleme. So kann der Adapter mit
 * Sicherheit einen gültigen Code in `playback_error.meta` setzen,
 * ohne dass freie Strings durchschlagen.
 */
export function normalizeWebRtcErrorCode(value: unknown): WebRtcErrorCode {
  if (isWebRtcErrorCode(value)) {
    return value;
  }
  return "peer_connection_failed";
}

/** Type-Guard ohne Fallback — gibt `false` für unbekannte Codes. */
export function isWebRtcErrorCode(value: unknown): value is WebRtcErrorCode {
  return typeof value === "string" && WEBRTC_ERROR_CODES.includes(value as WebRtcErrorCode);
}
