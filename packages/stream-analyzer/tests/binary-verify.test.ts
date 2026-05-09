import { describe, expect, it } from "vitest";

import { analyzeWithRuntime } from "../src/analyze.js";
import type {
  LoaderResponse,
  LoaderRuntime,
  ResolvedAddress
} from "../src/internal/loader/runtime.js";

/**
 * Tranche-4-Tests für `0.10.0` (NF-13 / RAK-64). Decken den Verifier-
 * Pfad mit injizierter LoaderRuntime ab — alle 13 Failure-Codes,
 * Status-Aggregation (`failed > skipped > passed`) und
 * `maxBinarySegments`-Cap.
 */

function makeBox(type: string, payload: Uint8Array | number[] = []): Uint8Array {
  const body = payload instanceof Uint8Array ? payload : new Uint8Array(payload);
  const total = 8 + body.byteLength;
  const buf = new Uint8Array(total);
  const dv = new DataView(buf.buffer);
  dv.setUint32(0, total, false);
  for (let i = 0; i < 4; i++) buf[4 + i] = type.charCodeAt(i);
  buf.set(body, 8);
  return buf;
}

function brandPayload(major: string, compatible: string[] = []): Uint8Array {
  const out = new Uint8Array(8 + compatible.length * 4);
  for (let i = 0; i < 4; i++) out[i] = major.charCodeAt(i);
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

function validInitBytes(): Uint8Array {
  return concat(
    makeBox("ftyp", brandPayload("cmfc")),
    makeBox("moov", makeBox("mvhd", new Uint8Array(20)))
  );
}

function validMediaBytes(): Uint8Array {
  const traf = makeBox("traf", makeBox("tfdt", new Uint8Array(8)));
  return concat(
    makeBox("styp", brandPayload("cmfs")),
    makeBox("moof", traf),
    makeBox("mdat", new Uint8Array(16))
  );
}

interface BinaryResponseSpec {
  status?: number;
  headers?: Record<string, string>;
  bytes?: Uint8Array;
  delayMs?: number;
}

function asyncIterableOf(bytes: Uint8Array): AsyncIterable<Uint8Array> {
  return (async function* () {
    yield bytes;
  })();
}

function makeBinaryRuntime(map: Record<string, BinaryResponseSpec>): LoaderRuntime {
  return {
    async resolveHost(hostname): Promise<ResolvedAddress[]> {
      void hostname;
      return [{ address: "93.184.216.34", family: 4 }];
    },
    async fetch(url, init) {
      const spec = map[url];
      if (!spec) throw new Error(`unexpected fetch ${url}`);
      if (spec.delayMs !== undefined) {
        await new Promise<void>((resolve, reject) => {
          const t = setTimeout(resolve, spec.delayMs);
          init.signal.addEventListener("abort", () => {
            clearTimeout(t);
            reject(new DOMException("aborted", "AbortError"));
          });
        });
      }
      const status = spec.status ?? 200;
      const headersMap = new Map<string, string>(
        Object.entries(spec.headers ?? {}).map(([k, v]) => [k.toLowerCase(), v])
      );
      const response: LoaderResponse = {
        status,
        headers: { get: (n) => headersMap.get(n.toLowerCase()) ?? null },
        body: spec.bytes !== undefined ? asyncIterableOf(spec.bytes) : null
      };
      return response;
    }
  };
}

const HLS_MAP_TEMPLATE = (mapAttr: string, segmentLine: string) =>
  [
    "#EXTM3U",
    "#EXT-X-VERSION:7",
    "#EXT-X-TARGETDURATION:4",
    "#EXT-X-MEDIA-SEQUENCE:0",
    `#EXT-X-MAP:${mapAttr}`,
    "#EXTINF:4.0,",
    segmentLine,
    "#EXT-X-ENDLIST"
  ].join("\n");

const DASH_TEMPLATE = (segmentTemplate: string) =>
  [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<MPD type="static">',
    "  <Period>",
    '    <AdaptationSet id="v" mimeType="video/mp4">',
    `      <SegmentTemplate ${segmentTemplate} startNumber="1"/>`,
    '      <Representation id="v1" bandwidth="1280000"/>',
    "    </AdaptationSet>",
    "  </Period>",
    "</MPD>"
  ].join("\n");

describe("binary-verify — HLS happy path (passed)", () => {
  it("verifies a CMAF Media-Playlist end-to-end", async () => {
    const initUrl = "https://cdn.example.test/dash/init.mp4";
    const mediaUrl = "https://cdn.example.test/dash/seg-0.m4s";
    const runtime = makeBinaryRuntime({
      [initUrl]: { headers: { "content-type": "video/mp4" }, bytes: validInitBytes() },
      [mediaUrl]: { headers: { "content-type": "video/mp4" }, bytes: validMediaBytes() }
    });
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") {
      throw new Error(`unexpected result: ${JSON.stringify(result)}`);
    }
    const cmaf = result.details.cmaf!;
    expect(cmaf.confidence).toBe("binary");
    expect(cmaf.binary?.status).toBe("passed");
    expect(cmaf.binary?.segmentsChecked.every((c) => c.status === "passed")).toBe(true);
    expect(cmaf.binary?.boxes.length).toBeGreaterThan(0);
  });
});

describe("binary-verify — HLS edge cases", () => {
  it("emits reference_missing for init when fMP4 segments exist without EXT-X-MAP", async () => {
    const runtime = makeBinaryRuntime({
      "https://cdn.example.test/dash/seg-0.m4s": {
        headers: { "content-type": "video/mp4" },
        bytes: validMediaBytes()
      }
    });
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:4",
      "#EXTINF:4.0,",
      "seg-0.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    const initCheck = result.details.cmaf!.binary!.segmentsChecked.find(
      (c) => c.kind === "init"
    );
    expect(initCheck?.failureCode).toBe("segment_reference_missing");
  });

  it("emits reference_missing for media when EXT-X-MAP exists but segments are .ts", async () => {
    const runtime = makeBinaryRuntime({
      "https://cdn.example.test/dash/init.mp4": {
        headers: { "content-type": "video/mp4" },
        bytes: validInitBytes()
      }
    });
    const text = [
      "#EXTM3U",
      "#EXT-X-TARGETDURATION:6",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:6.0,",
      "seg-0.ts",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    const mediaCheck = result.details.cmaf!.binary!.segmentsChecked.find(
      (c) => c.kind === "media"
    );
    expect(mediaCheck?.failureCode).toBe("segment_reference_missing");
  });
});

describe("binary-verify — HLS failure modes", () => {
  it("skipped + hls_map_byterange_unsupported when EXT-X-MAP has BYTERANGE", async () => {
    const runtime = makeBinaryRuntime({});
    const text = HLS_MAP_TEMPLATE('URI="init.mp4",BYTERANGE="1024@0"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(result.details.cmaf!.binary!.status).toBe("skipped");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init")?.failureCode
    ).toBe("hls_map_byterange_unsupported");
  });

  it("skipped + hls_media_byterange_unsupported when first segment has #EXT-X-BYTERANGE", async () => {
    const runtime = makeBinaryRuntime({});
    const text = [
      "#EXTM3U",
      "#EXT-X-VERSION:7",
      "#EXT-X-TARGETDURATION:4",
      '#EXT-X-MAP:URI="init.mp4"',
      "#EXTINF:4.0,",
      "#EXT-X-BYTERANGE:1024@0",
      "seg-0.m4s",
      "#EXT-X-ENDLIST"
    ].join("\n");
    const runtime2 = makeBinaryRuntime({
      "https://cdn.example.test/dash/init.mp4": {
        headers: { "content-type": "video/mp4" },
        bytes: validInitBytes()
      }
    });
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime2
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "media")?.failureCode
    ).toBe("hls_media_byterange_unsupported");
    void runtime;
  });

  it("failed + cmaf_box_validation_failed when init has wrong brand", async () => {
    const initUrl = "https://cdn.example.test/dash/init.mp4";
    const mediaUrl = "https://cdn.example.test/dash/seg-0.m4s";
    const badInit = concat(
      makeBox("ftyp", brandPayload("isom", ["mp42"])),
      makeBox("moov", makeBox("mvhd", new Uint8Array(20)))
    );
    const runtime = makeBinaryRuntime({
      [initUrl]: { headers: { "content-type": "video/mp4" }, bytes: badInit },
      [mediaUrl]: { headers: { "content-type": "video/mp4" }, bytes: validMediaBytes() }
    });
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(result.details.cmaf!.binary!.status).toBe("failed");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init")?.failureCode
    ).toBe("cmaf_box_validation_failed");
  });

  it("failed + invalid_box_structure when bytes are truncated", async () => {
    const initUrl = "https://cdn.example.test/dash/init.mp4";
    const mediaUrl = "https://cdn.example.test/dash/seg-0.m4s";
    const truncated = new Uint8Array([0, 0, 0, 0xff, 0x66, 0x74, 0x79, 0x70]);
    const runtime = makeBinaryRuntime({
      [initUrl]: { headers: { "content-type": "video/mp4" }, bytes: truncated },
      [mediaUrl]: { headers: { "content-type": "video/mp4" }, bytes: validMediaBytes() }
    });
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    const initCheck = result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init");
    expect(initCheck?.failureCode).toBe("invalid_box_structure");
    expect(result.details.cmaf!.binary!.status).toBe("failed");
  });

  it("skipped + segment_too_large when bytes exceed maxSegmentBytes", async () => {
    const initUrl = "https://cdn.example.test/dash/init.mp4";
    const mediaUrl = "https://cdn.example.test/dash/seg-0.m4s";
    // initBytes deutlich größer als maxSegmentBytes=64.
    const runtime = makeBinaryRuntime({
      [initUrl]: {
        headers: { "content-type": "video/mp4" },
        bytes: new Uint8Array(1024)
      },
      [mediaUrl]: { headers: { "content-type": "video/mp4" }, bytes: validMediaBytes() }
    });
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      { cmaf: { binary: { maxSegmentBytes: 64 } } },
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init")?.failureCode
    ).toBe("segment_too_large");
  });

  it("skipped + segment_content_type_unsupported when content-type is text/html", async () => {
    const initUrl = "https://cdn.example.test/dash/init.mp4";
    const mediaUrl = "https://cdn.example.test/dash/seg-0.m4s";
    const runtime = makeBinaryRuntime({
      [initUrl]: { headers: { "content-type": "text/html" }, bytes: validInitBytes() },
      [mediaUrl]: { headers: { "content-type": "video/mp4" }, bytes: validMediaBytes() }
    });
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init")?.failureCode
    ).toBe("segment_content_type_unsupported");
  });

  it("skipped + segment_fetch_failed on HTTP error status", async () => {
    const initUrl = "https://cdn.example.test/dash/init.mp4";
    const mediaUrl = "https://cdn.example.test/dash/seg-0.m4s";
    const runtime = makeBinaryRuntime({
      [initUrl]: { status: 500 },
      [mediaUrl]: { headers: { "content-type": "video/mp4" }, bytes: validMediaBytes() }
    });
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init")?.failureCode
    ).toBe("segment_fetch_failed");
  });

  it("skipped + segment_base_url_missing when text input has no baseUrl", async () => {
    const runtime = makeBinaryRuntime({});
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.find((c) => c.kind === "init")?.failureCode
    ).toBe("segment_base_url_missing");
  });

  it("emits binary_disabled when cmaf.binary.enabled=false", async () => {
    const runtime = makeBinaryRuntime({});
    const text = HLS_MAP_TEMPLATE('URI="init.mp4"', "seg-0.m4s");
    const result = await analyzeWithRuntime(
      { kind: "text", text, baseUrl: "https://cdn.example.test/dash/" },
      { cmaf: { binary: { enabled: false } } },
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "media") throw new Error("?");
    expect(result.details.cmaf!.binary!.status).toBe("skipped");
    expect(
      result.details.cmaf!.binary!.segmentsChecked.every((c) => c.failureCode === "binary_disabled")
    ).toBe(true);
  });
});

describe("binary-verify — DASH failure modes", () => {
  it("skipped + dash_template_unresolved for $Time$ in media template", async () => {
    const runtime = makeBinaryRuntime({});
    const mpd = DASH_TEMPLATE('initialization="init.mp4" media="seg-$Time$.m4s"');
    const result = await analyzeWithRuntime(
      { kind: "text", text: mpd, baseUrl: "https://cdn.example.test/dash/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "dash") throw new Error("?");
    const mediaCheck = result.details.cmaf!.binary!.segmentsChecked.find(
      (c) => c.kind === "media"
    );
    expect(mediaCheck?.failureCode).toBe("dash_template_unresolved");
  });

  it("skipped + segment_uri_blocked when AdaptationSet BaseURL chain is blocked", async () => {
    const runtime = makeBinaryRuntime({});
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      '    <AdaptationSet id="v" mimeType="video/mp4">',
      "      <BaseURL>file:///bad/</BaseURL>",
      '      <SegmentTemplate initialization="init.mp4" media="seg-$Number$.m4s" startNumber="1"/>',
      '      <Representation id="v1" bandwidth="1280000"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = await analyzeWithRuntime(
      { kind: "text", text: mpd, baseUrl: "https://cdn.example.test/" },
      {},
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "dash") throw new Error("?");
    const initCheck = result.details.cmaf!.binary!.segmentsChecked.find(
      (c) => c.kind === "init"
    );
    expect(initCheck?.failureCode).toBe("segment_uri_blocked");
  });

  it("skipped + not_planned_due_to_limit when adaptation sets exceed maxBinarySegments", async () => {
    const runtime = makeBinaryRuntime({});
    // 4 AdaptationSets × init+media = 8 Pflichtchecks, Cap=2.
    const sets = Array.from({ length: 4 }, (_, i) =>
      [
        `    <AdaptationSet id="a${i}" mimeType="video/mp4">`,
        '      <SegmentTemplate initialization="init.mp4" media="seg-$Number$.m4s" startNumber="1"/>',
        `      <Representation id="r${i}" bandwidth="${1000 * (i + 1)}"/>`,
        "    </AdaptationSet>"
      ].join("\n")
    );
    const mpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD type="static">',
      "  <Period>",
      ...sets,
      "  </Period>",
      "</MPD>"
    ].join("\n");
    const result = await analyzeWithRuntime(
      { kind: "text", text: mpd, baseUrl: "https://cdn.example.test/" },
      { cmaf: { binary: { maxBinarySegments: 2 } } },
      runtime
    );
    if (result.status !== "ok" || result.playlistType !== "dash") throw new Error("?");
    const codes = result.details.cmaf!.binary!.segmentsChecked.map((c) => c.failureCode);
    expect(codes).toContain("not_planned_due_to_limit");
    expect(result.details.cmaf!.binary!.limits.requiredSegmentChecks).toBe(8);
  });
});
