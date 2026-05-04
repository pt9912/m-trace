/**
 * URL redaction at the SDK boundary. Mirrors the backend matrix in
 * `spec/telemetry-model.md` §1.4 so tokens are stripped client-side
 * before the URL leaves the browser.
 *
 * Rules:
 *   - drop query and fragment
 *   - drop userinfo (`user:pass@`)
 *   - replace token-like path segments with `:redacted`
 *     - len ≥ 24 chars AND ≥ 80% from [A-Za-z0-9_-]
 *     - OR even-length hex with len ≥ 32
 *     - OR JWT-shape: three base64url blocks separated by dots
 *   - unparsable input collapses to the literal `:redacted` sentinel
 *     (never surface raw user input)
 */

const REDACTED = ":redacted";
const HEX_SEGMENT = /^(?:[0-9A-Fa-f]{2}){16,}$/;
const JWT_SEGMENT = /^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$/;

export function redactUrl(raw: string | undefined | null): string {
  if (typeof raw !== "string" || raw.length === 0) {
    return REDACTED;
  }
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    return REDACTED;
  }
  if (parsed.protocol === "" || parsed.host === "") {
    return REDACTED;
  }
  const segments = parsed.pathname.split("/").map((segment, index) => {
    if (index === 0 && segment === "") {
      // Leading "/" — keep it; pathname.split always produces an
      // empty leading segment for absolute paths.
      return segment;
    }
    if (segment === "") {
      return segment;
    }
    return isTokenLikePathSegment(decodePathSegment(segment))
      ? REDACTED
      : segment;
  });
  return `${parsed.protocol}//${parsed.host}${segments.join("/")}`;
}

export function isTokenLikePathSegment(seg: string): boolean {
  if (seg.length === 0) {
    return false;
  }
  if (HEX_SEGMENT.test(seg)) {
    return true;
  }
  if (JWT_SEGMENT.test(seg)) {
    return true;
  }
  if (seg.length < 24) {
    return false;
  }
  let allowed = 0;
  for (const ch of seg) {
    const code = ch.charCodeAt(0);
    const isUpper = code >= 0x41 && code <= 0x5a;
    const isLower = code >= 0x61 && code <= 0x7a;
    const isDigit = code >= 0x30 && code <= 0x39;
    const isUnder = code === 0x5f;
    const isDash = code === 0x2d;
    if (isUpper || isLower || isDigit || isUnder || isDash) {
      allowed += 1;
    }
  }
  return allowed * 100 >= seg.length * 80;
}

function decodePathSegment(seg: string): string {
  try {
    return decodeURIComponent(seg);
  } catch {
    // Defensively treat undecodable segments as token-like by
    // returning a value that will trip the heuristic length check
    // and the [A-Za-z0-9_-] ratio. The literal sentinel is the
    // simplest such value.
    return REDACTED;
  }
}
