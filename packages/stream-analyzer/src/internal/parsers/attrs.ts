/**
 * Mini-Parser für HLS-Attribut-Listen (RFC 8216 §4.2). Format:
 * `KEY=VALUE,KEY2="quoted, with commas",KEY3=0xHEX`. Wird sowohl
 * vom Master- als auch vom Media-Detail-Parser genutzt.
 *
 * Verhalten:
 *  - Werte in doppelten Anführungszeichen dürfen Kommas enthalten.
 *  - Unquoted Werte enden am nächsten Komma oder Stringende.
 *  - Schlüssel ohne `=` werden als leerer String erfasst und vom
 *    Aufrufer behandelt (z. B. als Finding bei Pflichtattributen).
 *  - Whitespace zwischen Komma und Schlüssel wird verschluckt;
 *    innerhalb eines unquoted Werts bleibt sie erhalten.
 *  - Doppelte Schlüssel überschreiben einander zugunsten des
 *    letzten Auftretens (HLS verbietet das, wir tolerieren es).
 *
 * Der Parser erfindet nichts und konvertiert nicht — Caster
 * (parseInt/parseFloat/Boolean) sind Aufgabe der Aufrufer, weil
 * je Feld unterschiedliche Konventionen gelten (RESOLUTION,
 * FRAME-RATE, DEFAULT, CHANNELS …).
 */
export function parseAttributeList(input: string): Map<string, string> {
  const result = new Map<string, string>();
  let i = 0;
  const n = input.length;
  while (i < n) {
    while (i < n && (input[i] === " " || input[i] === "\t")) i++;
    if (i >= n) break;
    const keyStart = i;
    while (i < n && input[i] !== "=" && input[i] !== ",") i++;
    const key = input.slice(keyStart, i).trim();
    if (key.length === 0) {
      if (i < n && input[i] === ",") i++;
      continue;
    }
    if (i >= n || input[i] !== "=") {
      result.set(key, "");
      if (i < n && input[i] === ",") i++;
      continue;
    }
    i++; // consume '='
    let value: string;
    if (i < n && input[i] === '"') {
      i++; // consume opening '"'
      const valStart = i;
      while (i < n && input[i] !== '"') i++;
      value = input.slice(valStart, i);
      if (i < n) i++; // consume closing '"'
    } else {
      const valStart = i;
      while (i < n && input[i] !== ",") i++;
      // Reale Manifeste haben gelegentlich Whitespace um '='; HLS
      // verbietet das streng, wir tolerieren es durch Trim auf
      // unquoted Werten. Quoted Werte bleiben byte-genau.
      value = input.slice(valStart, i).trim();
    }
    result.set(key, value);
    if (i < n && input[i] === ",") i++;
  }
  return result;
}

/** Liest `KEY=YES|NO` als Boolean. `undefined` für nicht gesetzt. */
export function parseYesNo(value: string | undefined): boolean | undefined {
  if (value === undefined) return undefined;
  if (value === "YES") return true;
  if (value === "NO") return false;
  return undefined;
}

/** Parst `WIDTHxHEIGHT` aus `RESOLUTION`. `null` bei Form-Fehler. */
export function parseResolution(value: string | undefined): { width: number; height: number } | null {
  if (value === undefined) return null;
  const match = /^([0-9]+)x([0-9]+)$/.exec(value);
  if (!match) return null;
  return { width: Number(match[1]), height: Number(match[2]) };
}

/** Parst Komma-getrennte Codecs aus `CODECS="a,b,c"`. */
export function parseCodecs(value: string | undefined): string[] | null {
  if (value === undefined) return null;
  if (value.length === 0) return [];
  return value.split(",").map((c) => c.trim()).filter((c) => c.length > 0);
}

/**
 * Parst einen Float-Wert. `null` bei undefined, leerem/Whitespace-only
 * String, NaN, ±Infinity oder negativen Werten. `Number("")` ist in
 * JavaScript 0 — diese Falle muss der Helper für die HLS-Pflichtfelder
 * (z. B. EXTINF-Dauer) explizit ausschließen, sonst rutschen leere
 * Eingaben als „0" durch.
 */
export function parseFloatAttr(value: string | undefined): number | null {
  if (value === undefined) return null;
  if (value.trim().length === 0) return null;
  const n = Number(value);
  if (!Number.isFinite(n)) return null;
  if (n < 0) return null;
  return n;
}

/** Parst Decimal-Integer; `null` bei NaN. */
export function parseIntAttr(value: string | undefined): number | null {
  if (value === undefined) return null;
  if (!/^[0-9]+$/.test(value)) return null;
  return Number(value);
}
