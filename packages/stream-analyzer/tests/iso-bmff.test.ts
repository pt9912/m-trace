import { describe, expect, it } from "vitest";

import {
  findBox,
  parseBrands,
  parseTopLevelBoxes,
  validateCmafInit,
  validateCmafMediaFragment
} from "../src/internal/cmaf/iso-bmff.js";

/**
 * ISO-BMFF-Box-Parser-Tests für (NF-13 / RAK-64).
 * Bauen Bytes-Fixtures programmatisch — keine externe Datei-
 * Dependency, keine Netzwerk-Abhängigkeit. Brand-Allowlist-Cases
 * decken die Plan-DoD-Pflichtfälle ab.
 */

function box(type: string, payload: Uint8Array | number[] = []): Uint8Array {
  const body = payload instanceof Uint8Array ? payload : new Uint8Array(payload);
  const total = 8 + body.byteLength;
  const buf = new Uint8Array(total);
  const dv = new DataView(buf.buffer);
  dv.setUint32(0, total, false);
  for (let i = 0; i < 4; i++) buf[4 + i] = type.charCodeAt(i);
  buf.set(body, 8);
  return buf;
}

function brandsPayload(major: string, minor: number, compatible: string[]): Uint8Array {
  const out = new Uint8Array(8 + compatible.length * 4);
  for (let i = 0; i < 4; i++) out[i] = major.charCodeAt(i);
  new DataView(out.buffer).setUint32(4, minor, false);
  for (let i = 0; i < compatible.length; i++) {
    for (let j = 0; j < 4; j++) {
      out[8 + i * 4 + j] = compatible[i].charCodeAt(j);
    }
  }
  return out;
}

function concat(...parts: Uint8Array[]): Uint8Array {
  const total = parts.reduce((s, p) => s + p.byteLength, 0);
  const out = new Uint8Array(total);
  let offset = 0;
  for (const p of parts) {
    out.set(p, offset);
    offset += p.byteLength;
  }
  return out;
}

function ftyp(major: string, compatible: string[] = []): Uint8Array {
  return box("ftyp", brandsPayload(major, 0, compatible));
}

function styp(major: string, compatible: string[] = []): Uint8Array {
  return box("styp", brandsPayload(major, 0, compatible));
}

function moov(): Uint8Array {
  // Minimaler moov mit leerem mvhd-Container reicht — Tranche 4
  // prüft moov nur auf Anwesenheit.
  return box("moov", box("mvhd", new Uint8Array(20)));
}

function tfdt(): Uint8Array {
  return box("tfdt", new Uint8Array(8));
}

function moof(): Uint8Array {
  const traf = box("traf", tfdt());
  return box("moof", traf);
}

function mdat(size = 16): Uint8Array {
  return box("mdat", new Uint8Array(size));
}

describe("iso-bmff — parseTopLevelBoxes", () => {
  it("parses three sequential boxes", () => {
    const data = concat(ftyp("cmfc"), moov(), mdat(8));
    const result = parseTopLevelBoxes(data);
    expect(result.errors).toEqual([]);
    expect(result.boxes.map((b) => b.type)).toEqual(["ftyp", "moov", "mdat"]);
    expect(result.boxes[0].offset).toBe(0);
    expect(result.boxes[1].offset).toBeGreaterThan(0);
  });

  it("returns invalid_box_structure when a box header is truncated", () => {
    const data = new Uint8Array([0, 0, 0]);
    const result = parseTopLevelBoxes(data);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].code).toBe("invalid_box_structure");
  });

  it("flags non-ASCII box types as invalid_box_structure", () => {
    const data = new Uint8Array(8);
    new DataView(data.buffer).setUint32(0, 8, false);
    data[4] = 0x01; // non-printable
    const result = parseTopLevelBoxes(data);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].code).toBe("invalid_box_structure");
  });

  it("rejects size=0 (extends-to-end-of-file) in 0.10.0 scope", () => {
    const data = new Uint8Array(8);
    // size=0, type=mdat
    data.set([0, 0, 0, 0, 0x6d, 0x64, 0x61, 0x74]);
    const result = parseTopLevelBoxes(data);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].code).toBe("invalid_box_structure");
    expect(result.errors[0].message).toContain("extends-to-end-of-file");
  });

  it("rejects sizes smaller than the header", () => {
    const data = new Uint8Array(8);
    new DataView(data.buffer).setUint32(0, 4, false); // too small
    data.set([0x6d, 0x64, 0x61, 0x74], 4);
    const result = parseTopLevelBoxes(data);
    expect(result.errors[0].code).toBe("invalid_box_structure");
  });

  it("rejects boxes that overrun the buffer", () => {
    const buf = new Uint8Array(8);
    new DataView(buf.buffer).setUint32(0, 32, false); // claims 32, only 8 available
    buf.set([0x6d, 0x6f, 0x6f, 0x76], 4);
    const result = parseTopLevelBoxes(buf);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].code).toBe("invalid_box_structure");
  });

  it("supports largesize (size=1, 64-bit) headers", () => {
    const total = 16 + 4; // 16-byte header + 4-byte payload
    const data = new Uint8Array(total);
    const dv = new DataView(data.buffer);
    dv.setUint32(0, 1, false); // size=1 → largesize
    data.set([0x6d, 0x64, 0x61, 0x74], 4);
    dv.setUint32(8, 0, false);
    dv.setUint32(12, total, false);
    const result = parseTopLevelBoxes(data);
    expect(result.errors).toEqual([]);
    expect(result.boxes[0].type).toBe("mdat");
    expect(result.boxes[0].size).toBe(total);
    expect(result.boxes[0].headerSize).toBe(16);
  });

  it("rejects largesize boxes that overflow 32-bit (high != 0)", () => {
    const data = new Uint8Array(16);
    const dv = new DataView(data.buffer);
    dv.setUint32(0, 1, false);
    data.set([0x6d, 0x64, 0x61, 0x74], 4);
    dv.setUint32(8, 1, false); // high=1 → > 4 GiB
    dv.setUint32(12, 0, false);
    const result = parseTopLevelBoxes(data);
    expect(result.errors[0].code).toBe("invalid_box_structure");
  });
});

describe("iso-bmff — parseBrands", () => {
  it("decodes major brand, minor version and compatible brands", () => {
    const payload = brandsPayload("cmfc", 0x00010002, ["isom", "iso6"]);
    const brands = parseBrands(payload);
    expect(brands).toEqual({
      majorBrand: "cmfc",
      minorVersion: 0x00010002,
      compatibleBrands: ["isom", "iso6"]
    });
  });

  it("returns null for under-sized payload", () => {
    expect(parseBrands(new Uint8Array(4))).toBeNull();
  });

  it("rejects non-ASCII major brand", () => {
    const payload = new Uint8Array(8);
    payload[0] = 0x01;
    expect(parseBrands(payload)).toBeNull();
  });
});

describe("iso-bmff — validateCmafInit", () => {
  it("passes for ftyp with major_brand cmfc + moov", () => {
    const data = concat(ftyp("cmfc"), moov());
    const result = validateCmafInit(data);
    expect(result.ok).toBe(true);
    expect(result.brands?.majorBrand).toBe("cmfc");
    expect(result.anchors.map((a) => a.type)).toEqual(["ftyp", "moov"]);
  });

  it("passes when cmf2 is in compatible_brands and major is generic", () => {
    const data = concat(ftyp("isom", ["cmf2"]), moov());
    expect(validateCmafInit(data).ok).toBe(true);
  });

  it("fails when only cmfs is in compatible_brands (cmfs is a media brand, not header)", () => {
    const data = concat(ftyp("isom", ["cmfs"]), moov());
    const result = validateCmafInit(data);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.code === "cmaf_box_validation_failed")).toBe(true);
  });

  it("fails when only generic brands are present", () => {
    const data = concat(ftyp("isom", ["iso6", "mp42"]), moov());
    expect(validateCmafInit(data).ok).toBe(false);
  });

  it("fails for cmf1 (Folge-Scope, not in 0.10.0 allowlist)", () => {
    const data = concat(ftyp("cmf1"), moov());
    expect(validateCmafInit(data).ok).toBe(false);
  });

  it("fails when ftyp is missing", () => {
    const data = moov();
    const result = validateCmafInit(data);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.boxPath === "ftyp")).toBe(true);
  });

  it("fails when moov is missing", () => {
    const data = ftyp("cmfc");
    const result = validateCmafInit(data);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.boxPath === "moov")).toBe(true);
  });
});

describe("iso-bmff — validateCmafMediaFragment", () => {
  it("passes for styp(cmfs) + moof/traf/tfdt + mdat", () => {
    const data = concat(styp("cmfs"), moof(), mdat());
    const result = validateCmafMediaFragment(data);
    expect(result.ok).toBe(true);
    expect(result.brands?.majorBrand).toBe("cmfs");
    const paths = result.anchors.map((a) => a.path);
    expect(paths).toContain("styp");
    expect(paths).toContain("moof");
    expect(paths).toContain("mdat");
    expect(paths).toContain("moof/traf");
    expect(paths).toContain("moof/traf/tfdt");
  });

  it("recognizes optional sidx without flagging it as required", () => {
    const sidx = box("sidx", new Uint8Array(12));
    const data = concat(styp("cmff"), sidx, moof(), mdat());
    const result = validateCmafMediaFragment(data);
    expect(result.ok).toBe(true);
    expect(result.hasSidx).toBe(true);
  });

  it("fails when styp is missing", () => {
    const data = concat(moof(), mdat());
    expect(validateCmafMediaFragment(data).ok).toBe(false);
  });

  it("fails when mdat is missing (status passed must not rely on metadata only)", () => {
    const data = concat(styp("cmfs"), moof());
    const result = validateCmafMediaFragment(data);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.boxPath === "mdat")).toBe(true);
  });

  it("fails when moof is missing", () => {
    const data = concat(styp("cmfs"), mdat());
    expect(validateCmafMediaFragment(data).ok).toBe(false);
  });

  it("fails when traf is missing inside moof", () => {
    const moofWithoutTraf = box("moof", new Uint8Array(0));
    const data = concat(styp("cmfs"), moofWithoutTraf, mdat());
    const result = validateCmafMediaFragment(data);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.boxPath === "moof/traf")).toBe(true);
  });

  it("fails when tfdt is missing inside traf", () => {
    const trafWithoutTfdt = box("traf", new Uint8Array(0));
    const moofMissing = box("moof", trafWithoutTfdt);
    const data = concat(styp("cmfs"), moofMissing, mdat());
    const result = validateCmafMediaFragment(data);
    expect(result.ok).toBe(false);
    expect(result.errors.some((e) => e.boxPath === "moof/traf/tfdt")).toBe(true);
  });

  it("fails for media styp brand outside the allowlist (mp42)", () => {
    const data = concat(styp("mp42"), moof(), mdat());
    expect(validateCmafMediaFragment(data).ok).toBe(false);
  });

  it("accepts cmf2 in styp (allowed for both header and media in 0.10.0)", () => {
    const data = concat(styp("isom", ["cmf2"]), moof(), mdat());
    expect(validateCmafMediaFragment(data).ok).toBe(true);
  });
});

describe("iso-bmff — findBox", () => {
  it("returns undefined when not found", () => {
    const result = parseTopLevelBoxes(concat(ftyp("cmfc")));
    expect(findBox(result.boxes, "moov")).toBeUndefined();
  });
});
