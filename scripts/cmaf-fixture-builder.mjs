#!/usr/bin/env node
/**
 * Erzeugt deterministische Minimal-CMAF-Bytes-Fixtures für den
 * CMAF-Smoke aus `make smoke-cli` (`0.10.0`  NF-13 / RAK-63).
 *
 * Output (relativ zum übergebenen Verzeichnis):
 *   init.mp4   — minimaler CMAF-Init mit ftyp(cmfc)+moov
 *   seg-0.m4s  — minimaler CMAF-Media-Fragment mit styp(cmfs)+moof/traf/tfdt+mdat
 *
 * Box-Layout exakt analog `tests/iso-bmff.test.ts`-Helper, damit die
 * Brand-Allowlist und Pflicht-Boxen aus T4 (cmfc/cmf2 Init,
 * cmfs/cmff/cmfc/cmf2 Media; ftyp+moov bzw. styp+moof+traf+tfdt+mdat)
 * von der Live-Library als `binary.status:"passed"` validiert werden.
 *
 * Aufruf: node scripts/cmaf-fixture-builder.mjs <out-dir>
 */
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

function makeBox(type, payload = []) {
  const body = payload instanceof Uint8Array ? payload : new Uint8Array(payload);
  const total = 8 + body.byteLength;
  const buf = new Uint8Array(total);
  const dv = new DataView(buf.buffer);
  dv.setUint32(0, total, false);
  for (let i = 0; i < 4; i++) buf[4 + i] = type.charCodeAt(i);
  buf.set(body, 8);
  return buf;
}

function brandPayload(major, compatible = []) {
  const out = new Uint8Array(8 + compatible.length * 4);
  for (let i = 0; i < 4; i++) out[i] = major.charCodeAt(i);
  for (let i = 0; i < compatible.length; i++) {
    for (let j = 0; j < 4; j++) {
      out[8 + i * 4 + j] = compatible[i].charCodeAt(j);
    }
  }
  return out;
}

function concat(...parts) {
  const total = parts.reduce((s, p) => s + p.byteLength, 0);
  const out = new Uint8Array(total);
  let offset = 0;
  for (const p of parts) {
    out.set(p, offset);
    offset += p.byteLength;
  }
  return out;
}

function buildInit() {
  // ftyp(cmfc) + moov(mvhd 20 zero-bytes). Brand-Allowlist erfordert
  // cmfc oder cmf2 als Init-Header-Brand.
  return concat(
    makeBox("ftyp", brandPayload("cmfc")),
    makeBox("moov", makeBox("mvhd", new Uint8Array(20)))
  );
}

function buildMedia() {
  // styp(cmfs) + moof(traf(tfdt)) + mdat(16 zero-bytes). Brand-
  // Allowlist erfordert cmfs/cmff/cmfc/cmf2 als Media-Brand.
  const traf = makeBox("traf", makeBox("tfdt", new Uint8Array(8)));
  return concat(
    makeBox("styp", brandPayload("cmfs")),
    makeBox("moof", traf),
    makeBox("mdat", new Uint8Array(16))
  );
}

function main() {
  const outDir = process.argv[2];
  if (!outDir) {
    console.error("usage: cmaf-fixture-builder.mjs <out-dir>");
    process.exit(2);
  }
  mkdirSync(outDir, { recursive: true });
  writeFileSync(path.join(outDir, "init.mp4"), buildInit());
  writeFileSync(path.join(outDir, "seg-0.m4s"), buildMedia());
  console.log(`[cmaf-fixture-builder] wrote init.mp4 (${buildInit().byteLength} bytes), seg-0.m4s (${buildMedia().byteLength} bytes) into ${outDir}`);
}

main();
