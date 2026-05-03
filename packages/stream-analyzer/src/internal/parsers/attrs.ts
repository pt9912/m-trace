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
  const n = input.length;
  let i = 0;
  while (i < n) {
    i = skipSpacesAndTabs(input, i, n);
    if (i >= n) break;

    const keyRead = readKey(input, i, n);
    i = keyRead.next;
    if (keyRead.key.length === 0) {
      i = consumeComma(input, i, n);
      continue;
    }
    if (i >= n || input[i] !== "=") {
      result.set(keyRead.key, "");
      i = consumeComma(input, i, n);
      continue;
    }

    i++; // consume '='
    const valueRead = readValue(input, i, n);
    result.set(keyRead.key, valueRead.value);
    i = consumeComma(input, valueRead.next, n);
  }
  return result;
}

function skipSpacesAndTabs(input: string, start: number, n: number): number {
  let cursor = start;
  while (cursor < n && (input[cursor] === " " || input[cursor] === "\t")) cursor++;
  return cursor;
}

function consumeComma(input: string, i: number, n: number): number {
  return i < n && input[i] === "," ? i + 1 : i;
}

function readKey(input: string, start: number, n: number): { key: string; next: number } {
  let cursor = start;
  while (cursor < n && input[cursor] !== "=" && input[cursor] !== ",") cursor++;
  return { key: input.slice(start, cursor).trim(), next: cursor };
}

function readValue(input: string, i: number, n: number): { value: string; next: number } {
  if (i < n && input[i] === '"') {
    return readQuotedValue(input, i + 1, n);
  }
  return readUnquotedValue(input, i, n);
}

function readQuotedValue(input: string, start: number, n: number): { value: string; next: number } {
  let cursor = start;
  while (cursor < n && input[cursor] !== '"') cursor++;
  const value = input.slice(start, cursor);
  // Schließendes '"' konsumieren, falls vorhanden.
  return { value, next: cursor < n ? cursor + 1 : cursor };
}

function readUnquotedValue(input: string, start: number, n: number): { value: string; next: number } {
  let cursor = start;
  while (cursor < n && input[cursor] !== ",") cursor++;
  // Reale Manifeste haben gelegentlich Whitespace um '='; HLS verbietet
  // das streng, wir tolerieren es durch Trim auf unquoted Werten.
  // Quoted Werte bleiben byte-genau (siehe readQuotedValue).
  return { value: input.slice(start, cursor).trim(), next: cursor };
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
