/**
 * Bounded ISO-BMFF-Box-Parser für die binäre CMAF-Konformitätsprüfung
 * (`0.10.0` Tranche 4, NF-13 / RAK-64).
 *
 * Scope:
 *  - 32-bit-Boxgrößen und `largesize` (Header-Größe `size = 1`,
 *    gefolgt von 64-bit). `size = 0` (Box erstreckt sich bis Datei-
 *    ende) bleibt **out of scope** in `0.10.0`, weil die geprüften
 *    Init-/Media-Segmente bewusst klein sind und ein `size = 0`-
 *    Konstrukt eine sehr seltene Top-Level-Form ist; entsprechende
 *    Boxen werden mit `invalid_box_structure` gemeldet.
 *  - Pflicht-Box-Erkennung für CMAF-Init (`ftyp`+`moov`) und CMAF-
 *    Media-Fragment (`styp`+`moof`/`traf`/`tfdt`+`mdat`).
 *  - Brand-Allowlist:
 *    Init-`ftyp`: `cmfc` oder `cmf2`.
 *    Media-`styp`: `cmfs`, `cmff`, `cmfc` oder `cmf2`.
 *  - `sidx` wird als optional informational erkannt, ist aber kein
 *    Muss für `passed`.
 *
 * Fehlerklassen (entsprechen den `CmafFailureCode`-Werten aus
 * `result.ts`):
 *  - `invalid_box_structure` — Boxgröße/Überlappung/Fortschritt
 *    strukturell ungültig.
 *  - `cmaf_box_validation_failed` — Boxen vorhanden, aber Brand-
 *    Allowlist verletzt oder Pflicht-Box fehlt.
 */

const HEADER_SIZE = 8;
const LARGE_HEADER_SIZE = 16;

const INIT_BRAND_ALLOWLIST: ReadonlySet<string> = new Set(["cmfc", "cmf2"]);
const MEDIA_BRAND_ALLOWLIST: ReadonlySet<string> = new Set([
  "cmfs",
  "cmff",
  "cmfc",
  "cmf2"
]);

export interface IsoBmffBox {
  /** ISO-BMFF-Box-Type, vier ASCII-Zeichen. */
  readonly type: string;
  /** Absolute Byte-Position der Boxgrenze im Segment. */
  readonly offset: number;
  /** Gesamt-Boxgröße inklusive Header. */
  readonly size: number;
  /** Header-Länge: 8 (32-bit) oder 16 (largesize). */
  readonly headerSize: number;
  /**
   * View auf den Payload-Anteil (ohne Header). Tranche 4 nutzt das
   * View, um Children und Brands zu lesen — keine Kopie, kein neuer
   * Buffer.
   */
  readonly payload: Uint8Array;
}

export interface IsoBmffParseError {
  readonly code: "invalid_box_structure" | "cmaf_box_validation_failed";
  readonly message: string;
  readonly boxPath?: string;
}

export interface IsoBmffParseResult {
  readonly boxes: readonly IsoBmffBox[];
  readonly errors: readonly IsoBmffParseError[];
}

/**
 * Parst die Top-Level-Boxen eines Bytes-Buffers. Bricht bei
 * strukturellen Verstößen (Größe > Buffer, Größe < Header,
 * Fortschritt 0) deterministisch ab und sammelt den Fehler in
 * `errors[]` — Caller entscheidet, ob ein Teilergebnis (bis zum
 * Fehler) auswertbar ist.
 */
export function parseTopLevelBoxes(data: Uint8Array): IsoBmffParseResult {
  const boxes: IsoBmffBox[] = [];
  const errors: IsoBmffParseError[] = [];
  let offset = 0;
  while (offset < data.byteLength) {
    const box = readBox(data, offset);
    if ("error" in box) {
      errors.push(box.error);
      break;
    }
    if (box.size <= 0 || box.offset + box.size > data.byteLength) {
      errors.push({
        code: "invalid_box_structure",
        message: `Box "${box.type}" bei Offset ${box.offset} hat Größe ${box.size}, aber nur ${data.byteLength - box.offset} Bytes verbleiben.`,
        boxPath: box.type
      });
      break;
    }
    boxes.push(box);
    offset = box.offset + box.size;
  }
  return { boxes, errors };
}

interface BoxReadError {
  readonly error: IsoBmffParseError;
}

function readBox(data: Uint8Array, offset: number): IsoBmffBox | BoxReadError {
  if (offset + HEADER_SIZE > data.byteLength) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `Box-Header bei Offset ${offset} bricht den Buffer ab (verbleibend ${data.byteLength - offset} < ${HEADER_SIZE}).`
      }
    };
  }
  const headerView = new DataView(data.buffer, data.byteOffset + offset, HEADER_SIZE);
  const sizeField = headerView.getUint32(0, false);
  const type = decodeBoxType(data, offset + 4);
  if (type === null) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `Box-Type bei Offset ${offset} ist nicht ASCII-printable.`
      }
    };
  }
  if (sizeField === 1) {
    return readLargeSizeBox(data, offset, type);
  }
  if (sizeField === 0) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `Box "${type}" bei Offset ${offset} hat size=0 (extends-to-end-of-file); im 0.10.0-Scope nicht unterstützt.`,
        boxPath: type
      }
    };
  }
  if (sizeField < HEADER_SIZE) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `Box "${type}" bei Offset ${offset} hat size=${sizeField}, kleiner als Header (${HEADER_SIZE}).`,
        boxPath: type
      }
    };
  }
  const payload = data.subarray(offset + HEADER_SIZE, offset + sizeField);
  return { type, offset, size: sizeField, headerSize: HEADER_SIZE, payload };
}

function readLargeSizeBox(data: Uint8Array, offset: number, type: string): IsoBmffBox | BoxReadError {
  if (offset + LARGE_HEADER_SIZE > data.byteLength) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `largesize-Box "${type}" bei Offset ${offset} hat keinen vollständigen 16-Byte-Header.`,
        boxPath: type
      }
    };
  }
  const view = new DataView(data.buffer, data.byteOffset + offset + 8, 8);
  const high = view.getUint32(0, false);
  const low = view.getUint32(4, false);
  // largesize > Number.MAX_SAFE_INTEGER → für unsere Zwecke
  // unvereinbar mit dem Segment-Größenlimit (typisch < 2 GiB).
  if (high !== 0) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `largesize-Box "${type}" überschreitet 2^32 Bytes — im 0.10.0-Scope unzulässig groß.`,
        boxPath: type
      }
    };
  }
  const size = low;
  if (size < LARGE_HEADER_SIZE) {
    return {
      error: {
        code: "invalid_box_structure",
        message: `largesize-Box "${type}" hat size=${size}, kleiner als Header (${LARGE_HEADER_SIZE}).`,
        boxPath: type
      }
    };
  }
  const payload = data.subarray(offset + LARGE_HEADER_SIZE, offset + size);
  return { type, offset, size, headerSize: LARGE_HEADER_SIZE, payload };
}

function decodeBoxType(data: Uint8Array, offset: number): string | null {
  let result = "";
  for (let i = 0; i < 4; i++) {
    const code = data[offset + i];
    // ASCII-printable inkl. ' ' (Space, häufig als Padding in Brands).
    if (code < 0x20 || code > 0x7e) return null;
    result += String.fromCharCode(code);
  }
  return result;
}

/**
 * Liest die Children-Box-Liste eines Container-Box-Payloads. Wird
 * z. B. für `moov`, `moof`, `traf` benötigt.
 */
export function parseChildBoxes(payload: Uint8Array): IsoBmffParseResult {
  return parseTopLevelBoxes(payload);
}

export interface BrandsInfo {
  readonly majorBrand: string;
  readonly minorVersion: number;
  readonly compatibleBrands: readonly string[];
}

/**
 * Parst `ftyp`/`styp`-Body laut ISO-BMFF. Liefert `null`, wenn der
 * Payload zu kurz für `major_brand` (4) + `minor_version` (4) ist.
 */
export function parseBrands(payload: Uint8Array): BrandsInfo | null {
  if (payload.byteLength < 8) return null;
  const view = new DataView(payload.buffer, payload.byteOffset, payload.byteLength);
  const majorBrand = decodeBrand(payload, 0);
  if (majorBrand === null) return null;
  const minorVersion = view.getUint32(4, false);
  const compatibleBrands: string[] = [];
  let cursor = 8;
  while (cursor + 4 <= payload.byteLength) {
    const b = decodeBrand(payload, cursor);
    if (b === null) break;
    compatibleBrands.push(b);
    cursor += 4;
  }
  return { majorBrand, minorVersion, compatibleBrands };
}

function decodeBrand(data: Uint8Array, offset: number): string | null {
  let s = "";
  for (let i = 0; i < 4; i++) {
    const code = data[offset + i];
    if (code === undefined) return null;
    if (code < 0x20 || code > 0x7e) return null;
    s += String.fromCharCode(code);
  }
  return s;
}

/**
 * Findet die erste Box mit dem gegebenen Type — top-level. Liefert
 * `undefined`, wenn nicht vorhanden.
 */
export function findBox(boxes: readonly IsoBmffBox[], type: string): IsoBmffBox | undefined {
  return boxes.find((b) => b.type === type);
}

export interface InitValidationResult {
  readonly ok: boolean;
  /** Brand-Information, sofern `ftyp` parsbar war. */
  readonly brands?: BrandsInfo;
  readonly errors: readonly IsoBmffParseError[];
  readonly anchors: readonly InitBoxAnchor[];
}

export interface InitBoxAnchor {
  readonly path: string;
  readonly type: string;
  readonly offset: number;
  readonly size: number;
}

/**
 * CMAF-Init-Validator (`ftyp`+`moov`+Brand-Allowlist `cmfc`/`cmf2`).
 * Akzeptiert generische ISO-BMFF-Brands wie `isom`/`iso6`/`mp41`/
 * `mp42` zusätzlich, aber nicht alleinig.
 */
export function validateCmafInit(data: Uint8Array): InitValidationResult {
  const parsed = parseTopLevelBoxes(data);
  const errors: IsoBmffParseError[] = [...parsed.errors];
  const anchors: InitBoxAnchor[] = [];

  const ftyp = findBox(parsed.boxes, "ftyp");
  if (ftyp === undefined) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: "Init-Segment enthält kein ftyp.",
      boxPath: "ftyp"
    });
  } else {
    anchors.push({ path: "ftyp", type: ftyp.type, offset: ftyp.offset, size: ftyp.size });
  }

  const moov = findBox(parsed.boxes, "moov");
  if (moov === undefined) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: "Init-Segment enthält kein moov.",
      boxPath: "moov"
    });
  } else {
    anchors.push({ path: "moov", type: moov.type, offset: moov.offset, size: moov.size });
  }

  const brands = ftyp !== undefined ? parseBrands(ftyp.payload) : null;
  if (ftyp !== undefined && brands === null) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: "Init-Segment ftyp ist zu kurz für Brand-Auswertung.",
      boxPath: "ftyp"
    });
  } else if (brands !== null && !hasAllowedInitBrand(brands)) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: `Init-Segment ftyp führt keine in 0.10.0 unterstützte CMAF-Header-Brand (${[...INIT_BRAND_ALLOWLIST].join(", ")}); gefunden: major=${brands.majorBrand}, compatible=[${brands.compatibleBrands.join(",")}].`,
      boxPath: "ftyp"
    });
  }

  return {
    ok: errors.length === 0,
    ...(brands !== null ? { brands } : {}),
    errors,
    anchors
  };
}

function hasAllowedInitBrand(brands: BrandsInfo): boolean {
  if (INIT_BRAND_ALLOWLIST.has(brands.majorBrand)) return true;
  return brands.compatibleBrands.some((b) => INIT_BRAND_ALLOWLIST.has(b));
}

export interface MediaValidationResult {
  readonly ok: boolean;
  readonly brands?: BrandsInfo;
  readonly errors: readonly IsoBmffParseError[];
  readonly anchors: readonly MediaBoxAnchor[];
  /** True, wenn ein optionales `sidx` erkannt wurde (informational). */
  readonly hasSidx: boolean;
}

export interface MediaBoxAnchor {
  readonly path: string;
  readonly type: string;
  readonly offset: number;
  readonly size: number;
}

/**
 * CMAF-Media-Fragment-Validator (`styp` + `moof/traf/tfdt` + `mdat`
 * + Brand-Allowlist `cmfs`/`cmff`/`cmfc`/`cmf2`). `sidx` wird
 * erkannt, ist aber kein Muss-Beweis. `cmfl`/Low-Latency bleibt
 * Folge-Scope.
 */
export function validateCmafMediaFragment(data: Uint8Array): MediaValidationResult {
  const parsed = parseTopLevelBoxes(data);
  const errors: IsoBmffParseError[] = [...parsed.errors];
  const anchors: MediaBoxAnchor[] = [];
  const styp = findBox(parsed.boxes, "styp");
  const moof = findBox(parsed.boxes, "moof");
  const mdat = findBox(parsed.boxes, "mdat");
  const sidx = findBox(parsed.boxes, "sidx");

  collectMediaTopLevelAnchors({ styp, moof, mdat, sidx }, anchors);
  collectMediaTopLevelErrors({ styp, moof, mdat }, errors);

  const brands = styp !== undefined ? parseBrands(styp.payload) : null;
  validateMediaBrands(styp, brands, errors);

  if (moof !== undefined) {
    validateMoofChildren(moof, anchors, errors);
  }

  return {
    ok: errors.length === 0,
    ...(brands !== null ? { brands } : {}),
    errors,
    anchors,
    hasSidx: sidx !== undefined
  };
}

function collectMediaTopLevelAnchors(
  boxes: { styp?: IsoBmffBox; moof?: IsoBmffBox; mdat?: IsoBmffBox; sidx?: IsoBmffBox },
  out: MediaBoxAnchor[]
): void {
  if (boxes.styp !== undefined) {
    out.push({ path: "styp", type: "styp", offset: boxes.styp.offset, size: boxes.styp.size });
  }
  if (boxes.sidx !== undefined) {
    out.push({ path: "sidx", type: "sidx", offset: boxes.sidx.offset, size: boxes.sidx.size });
  }
  if (boxes.moof !== undefined) {
    out.push({ path: "moof", type: "moof", offset: boxes.moof.offset, size: boxes.moof.size });
  }
  if (boxes.mdat !== undefined) {
    out.push({ path: "mdat", type: "mdat", offset: boxes.mdat.offset, size: boxes.mdat.size });
  }
}

function collectMediaTopLevelErrors(
  boxes: { styp?: IsoBmffBox; moof?: IsoBmffBox; mdat?: IsoBmffBox },
  out: IsoBmffParseError[]
): void {
  if (boxes.styp === undefined) {
    out.push({
      code: "cmaf_box_validation_failed",
      message: "Media-Segment enthält kein styp.",
      boxPath: "styp"
    });
  }
  if (boxes.moof === undefined) {
    out.push({
      code: "cmaf_box_validation_failed",
      message: "Media-Segment enthält kein moof.",
      boxPath: "moof"
    });
  }
  if (boxes.mdat === undefined) {
    out.push({
      code: "cmaf_box_validation_failed",
      message: "Media-Segment enthält kein mdat.",
      boxPath: "mdat"
    });
  }
}

function validateMediaBrands(
  styp: IsoBmffBox | undefined,
  brands: BrandsInfo | null,
  errors: IsoBmffParseError[]
): void {
  if (styp === undefined) return;
  if (brands === null) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: "Media-Segment styp ist zu kurz für Brand-Auswertung.",
      boxPath: "styp"
    });
    return;
  }
  if (!hasAllowedMediaBrand(brands)) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: `Media-Segment styp führt keine in 0.10.0 unterstützte CMAF-Media-Brand (${[...MEDIA_BRAND_ALLOWLIST].join(", ")}); gefunden: major=${brands.majorBrand}, compatible=[${brands.compatibleBrands.join(",")}].`,
      boxPath: "styp"
    });
  }
}

function hasAllowedMediaBrand(brands: BrandsInfo): boolean {
  if (MEDIA_BRAND_ALLOWLIST.has(brands.majorBrand)) return true;
  return brands.compatibleBrands.some((b) => MEDIA_BRAND_ALLOWLIST.has(b));
}

function validateMoofChildren(
  moof: IsoBmffBox,
  anchors: MediaBoxAnchor[],
  errors: IsoBmffParseError[]
): void {
  const children = parseChildBoxes(moof.payload);
  errors.push(...children.errors);
  const traf = findBox(children.boxes, "traf");
  if (traf === undefined) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: "moof enthält kein traf.",
      boxPath: "moof/traf"
    });
    return;
  }
  anchors.push({
    path: "moof/traf",
    type: "traf",
    offset: moof.offset + moof.headerSize + traf.offset,
    size: traf.size
  });
  const trafChildren = parseChildBoxes(traf.payload);
  errors.push(...trafChildren.errors);
  const tfdt = findBox(trafChildren.boxes, "tfdt");
  if (tfdt === undefined) {
    errors.push({
      code: "cmaf_box_validation_failed",
      message: "traf enthält kein tfdt.",
      boxPath: "moof/traf/tfdt"
    });
    return;
  }
  anchors.push({
    path: "moof/traf/tfdt",
    type: "tfdt",
    offset: moof.offset + moof.headerSize + traf.offset + traf.headerSize + tfdt.offset,
    size: tfdt.size
  });
}
