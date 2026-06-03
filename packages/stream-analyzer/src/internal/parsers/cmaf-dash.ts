/**
 * DASH-spezifische CMAF-Detection (NF-13 / RAK-62
 * / RAK-64). Stellt drei Hilfsfamilien bereit:
 *
 *  1. **Manifestanker** — stabiler XPath-artiger Pfad pro
 *     Period/AdaptationSet/Representation, mit Index-Fallback bei
 *     fehlenden IDs.
 *  2. **BaseURL-Vererbung** — pro Ebene (`MPD`/`Period`/
 *     `AdaptationSet`/`Representation`) wird die erste sichere
 *     HTTP(S)-`BaseURL` in Manifest-Reihenfolge an die nächste
 *     Ebene vererbt; relative Werte werden gegen die geerbte Base
 *     aufgelöst.
 *  3. **SegmentTemplate-Auflösung** — bounded Variable-Resolver für
 *     `$RepresentationID$`, `$Bandwidth$`, `$Number$` und
 *     `$Number%0Nd$` mit `startNumber`/Default `1`. `$Time$` und
 *     `SegmentTimeline`-abhängige Pfade liefern `null` (Tranche 4
 *     mappt das auf `dash_template_unresolved`).
 */

const FMP4_EXTENSIONS = [".m4s", ".cmfv", ".cmfa", ".mp4"] as const;
const SAFE_SCHEMES = new Set(["http:", "https:"]);
const MP4_MIME_TYPES = new Set(["video/mp4", "audio/mp4", "application/mp4"]);

/**
 * Hierarchischer Manifestanker laut Plan §3-DoD-Beispiel
 * `MPD/Period[0]/AdaptationSet[id=video]/Representation[id=v1]`.
 * Bei fehlender ID wird der 0-basierte Index in eckigen Klammern
 * gesetzt; ID-Variante hat Vorrang, weil sie über mehrere Releases
 * stabiler bleibt.
 */
export interface DashAnchorPath {
  readonly periodIdx: number;
  readonly periodId?: string;
  readonly adaptationSetIdx?: number;
  readonly adaptationSetId?: string;
  readonly representationIdx?: number;
  readonly representationId?: string;
}

export function buildDashAnchor(path: DashAnchorPath, attribute?: string): string {
  const parts: string[] = ["MPD"];
  parts.push(formatLevel("Period", path.periodIdx, path.periodId));
  if (path.adaptationSetIdx !== undefined) {
    parts.push(formatLevel("AdaptationSet", path.adaptationSetIdx, path.adaptationSetId));
  }
  if (path.representationIdx !== undefined) {
    parts.push(formatLevel("Representation", path.representationIdx, path.representationId));
  }
  const base = parts.join("/");
  return attribute !== undefined ? `${base}/${attribute}` : base;
}

function formatLevel(name: string, idx: number, id: string | undefined): string {
  if (id !== undefined && id.length > 0) {
    return `${name}[id=${id}]`;
  }
  return `${name}[${idx}]`;
}

export function isMp4MimeType(value: string | undefined): boolean {
  if (value === undefined) return false;
  return MP4_MIME_TYPES.has(value.trim().toLowerCase());
}

/**
 * fMP4-Suffix-Heuristik analog `cmaf-hls.ts` für DASH-Segment-URI-
 * Templates und konkrete Segment-URIs. Query-/Fragment-Suffixe
 * werden vor dem Match abgeschnitten.
 */
export function isFmp4DashUri(uri: string): boolean {
  if (uri.length === 0) return false;
  const cut = stripUriSuffix(uri).toLowerCase();
  return FMP4_EXTENSIONS.some((ext) => cut.endsWith(ext));
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
 * BaseURL-Chain mit explizitem „erste-sichere-Eintrag-gewinnt"-
 * Regelwerk aus dem T3-DoD. `parent` ist die geerbte sichere Base
 * (oder `undefined` ganz oben); `candidates` sind die in Manifest-
 * Reihenfolge erfassten `<BaseURL>`-Texte der aktuellen Ebene.
 *
 * Returns: `{ baseUrl, blocked }`. `baseUrl` ist die neue sichere
 * Base für die nächste Ebene (kann gleich der geerbten sein, wenn
 * die aktuelle Ebene keinen Eintrag hat). `blocked === true`, wenn
 * mindestens ein Kandidat existierte, aber alle gegen die Sicherheits-
 * regeln verstoßen — nutzt das, um zwischen
 * `segment_uri_blocked` und `segment_base_url_missing` zu trennen.
 */
export interface ResolvedBaseUrl {
  readonly baseUrl?: string;
  readonly blocked: boolean;
}

export function resolveBaseUrlChain(
  candidates: readonly string[],
  parent: string | undefined,
  parentBlocked: boolean = false
): ResolvedBaseUrl {
  if (candidates.length === 0) {
    // Keine eigenen Kandidaten → Vererbung von Parent (inklusive
    // Block-Zustand: ein in höherer Ebene gesetzter Block bleibt
    // sichtbar, damit bei der Pflichtprüfung
    // segment_uri_blocked statt segment_base_url_missing meldet).
    return { baseUrl: parent, blocked: parentBlocked };
  }
  for (const raw of candidates) {
    const trimmed = raw.trim();
    if (trimmed.length === 0) continue;
    const resolved = tryResolveBaseUrl(trimmed, parent);
    if (resolved !== null) {
      return { baseUrl: resolved, blocked: false };
    }
  }
  // Alle Kandidaten unsicher / nicht auflösbar — geerbte Base wird
  // **nicht** durchgereicht, weil das Manifest auf dieser Ebene
  // explizit eine andere Base wollte. mappt das Ergebnis
  // auf segment_uri_blocked.
  return { baseUrl: undefined, blocked: true };
}

function tryResolveBaseUrl(value: string, parent: string | undefined): string | null {
  // Absolute URL ohne Schema-Validation rauswerfen.
  try {
    if (/^[a-z][a-z0-9+.-]*:/i.test(value)) {
      const u = new URL(value);
      if (!SAFE_SCHEMES.has(u.protocol)) return null;
      return u.toString();
    }
  } catch {
    return null;
  }
  if (parent === undefined) return null;
  try {
    const u = new URL(value, parent);
    if (!SAFE_SCHEMES.has(u.protocol)) return null;
    return u.toString();
  } catch {
    return null;
  }
}

/**
 * Auflösung einer Segment-URI gegen eine sichere BaseURL nach
 * denselben Regeln wie der Loader. `null` zurück bedeutet: blocked
 * oder unauflösbar (z. B. relative URI ohne BaseURL, oder
 * unsicheres Schema).
 */
export function resolveSegmentUri(uri: string, baseUrl: string | undefined): string | null {
  if (uri.length === 0) return null;
  // Absolute URI mit unsicherem Schema lehnen wir ab, auch wenn
  // BaseURL gesetzt ist.
  if (/^[a-z][a-z0-9+.-]*:/i.test(uri)) {
    try {
      const u = new URL(uri);
      if (!SAFE_SCHEMES.has(u.protocol)) return null;
      return u.toString();
    } catch {
      return null;
    }
  }
  if (baseUrl === undefined) return null;
  try {
    const u = new URL(uri, baseUrl);
    if (!SAFE_SCHEMES.has(u.protocol)) return null;
    return u.toString();
  } catch {
    return null;
  }
}

/**
 * Resolver für `SegmentTemplate@initialization` und `@media` mit
 * dem Scope: `$RepresentationID$`, `$Bandwidth$`,
 * `$Number$` und `$Number%0Nd$`. `$Time$` oder unbekannte
 * Variablen führen zu `null` (Tranche 4 mappt das auf
 * `dash_template_unresolved`).
 */
export interface DashTemplateContext {
  readonly representationId: string;
  readonly bandwidth: number;
  readonly number: number;
}

export function resolveDashTemplate(
  template: string,
  ctx: DashTemplateContext
): string | null {
  let result = "";
  let i = 0;
  while (i < template.length) {
    const ch = template.charAt(i);
    if (ch !== "$") {
      // Literales `$$` als Escape laut DASH-Spec — wird hier nicht
      // erwartet, weil `$$` außerhalb von Variablen auftaucht; ein
      // Doppel-Dollar-Literal vor einer Variable ist im
      // 0.10.0-Scope nicht ableitbar.
      result += ch;
      i++;
      continue;
    }
    const closeIdx = template.indexOf("$", i + 1);
    if (closeIdx === -1) return null; // unbalanciertes $
    const variable = template.slice(i + 1, closeIdx);
    const expanded = expandVariable(variable, ctx);
    if (expanded === null) return null;
    result += expanded;
    i = closeIdx + 1;
  }
  return result;
}

function expandVariable(variable: string, ctx: DashTemplateContext): string | null {
  if (variable === "") return "$"; // `$$` → literal `$`
  if (variable === "RepresentationID") return ctx.representationId;
  if (variable === "Bandwidth") return String(ctx.bandwidth);
  if (variable === "Number") return String(ctx.number);
  // printf-artiges `Number%0Nd` für Zero-Padding.
  const m = variable.match(/^Number%(0\d+)d$/);
  if (m !== null) {
    const widthStr = m[1].slice(1); // ohne führende `0`
    const width = Number.parseInt(widthStr, 10);
    if (!Number.isFinite(width) || width <= 0) return null;
    return String(ctx.number).padStart(width, "0");
  }
  // Bandwidth%0Nd analog DASH-Spec.
  const bm = variable.match(/^Bandwidth%(0\d+)d$/);
  if (bm !== null) {
    const widthStr = bm[1].slice(1);
    const width = Number.parseInt(widthStr, 10);
    if (!Number.isFinite(width) || width <= 0) return null;
    return String(ctx.bandwidth).padStart(width, "0");
  }
  return null;
}
