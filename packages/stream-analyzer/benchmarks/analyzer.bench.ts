import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { bench, describe } from "vitest";

import { analyzeManifest } from "../src/index.js";
import { detectManifestKind } from "../src/internal/parsers/detect.js";
import { validateUrl } from "../src/internal/loader/ssrf.js";

// plan-0.9.5 §2 Tranche 1 (RAK-Wave-2 / extra-gates.md §3.2) —
// Stream-Analyzer-Hot-Path-Benchmarks für
// `make analyzer-benchmark-smoke`.
//
// Budgets aus `docs/perf/budgets.md` §4 (initial, Tranche-0-Stand;
// noch nicht mess-basiert, sondern Architektur-basierte Obergrenzen):
//
//   - HLS Master klein (5 Variants + 1 Rendition): ≤ 5 ms
//   - HLS Master groß (50 Variants + 20 Renditions): ≤ 25 ms
//   - HLS Media (1.000 Segmente):                    ≤ 50 ms
//   - DASH-MPD VOD (1 Period, 2 AdaptationSets):     ≤ 5 ms
//   - DASH-MPD Live (3 AdaptationSets):              ≤ 10 ms
//   - Detector über 256-KiB-Body:                    ≤ 500 µs
//   - SSRF-URL-Klassifizierung (100 Calls):          ≤ 5 ms / 100
//
// Ausführung: `vitest bench --run packages/stream-analyzer/benchmarks/`
// (eingebaute Vitest-Bench-API, keine externe Tinybench-Dependency).
// Beobachtungsphase laut Plan §2 DoD: erste N=3-5 grüne CI-Läufe
// non-blocking; danach landet `make benchmark-smoke` PR-blockierend
// in `make gates`.

const fixturesDir = join(
  dirname(fileURLToPath(import.meta.url)),
  "..",
  "tests",
  "fixtures"
);

const MASTER_SMALL = readFileSync(join(fixturesDir, "master.m3u8"), "utf8");
const MASTER_LARGE = generateLargeMaster(50, 20);
const MEDIA_1K = generate1kSegmentMedia();
const DASH_VOD = generateDashVod();
const DASH_LIVE = generateDashLive();
const DETECTOR_BODY_256K = generateDetectorBody();
const SSRF_URL_SAMPLES = generateSsrfUrlSamples();

describe("stream-analyzer / HLS hot-paths", () => {
  bench("HLS Master klein (5 Variants + 1 Rendition, ≤ 5 ms)", async () => {
    const r = await analyzeManifest({ kind: "text", text: MASTER_SMALL });
    if (r.status !== "ok") {
      throw new Error(`expected ok, got ${r.status}`);
    }
  });

  bench("HLS Master groß (50 Variants + 20 Renditions, ≤ 25 ms)", async () => {
    const r = await analyzeManifest({ kind: "text", text: MASTER_LARGE });
    if (r.status !== "ok") {
      throw new Error(`expected ok, got ${r.status}`);
    }
  });

  bench("HLS Media (1.000 Segmente, ≤ 50 ms)", async () => {
    const r = await analyzeManifest({ kind: "text", text: MEDIA_1K });
    if (r.status !== "ok") {
      throw new Error(`expected ok, got ${r.status}`);
    }
  });
});

describe("stream-analyzer / DASH hot-paths (plan-0.9.0 Tranche 3)", () => {
  bench("DASH-MPD VOD (1 Period / 2 AdaptationSets, ≤ 5 ms)", async () => {
    const r = await analyzeManifest({ kind: "text", text: DASH_VOD });
    if (r.status !== "ok") {
      throw new Error(`expected ok, got ${r.status}`);
    }
  });

  bench("DASH-MPD Live (3 AdaptationSets, ≤ 10 ms)", async () => {
    const r = await analyzeManifest({ kind: "text", text: DASH_LIVE });
    if (r.status !== "ok") {
      throw new Error(`expected ok, got ${r.status}`);
    }
  });
});

describe("stream-analyzer / Detector + SSRF", () => {
  bench("Detector über 256-KiB-Body (≤ 500 µs)", () => {
    const r = detectManifestKind(DETECTOR_BODY_256K);
    if (r.kind !== "unsupported") {
      throw new Error(`expected unsupported, got ${r.kind}`);
    }
  });

  bench("SSRF-URL-Klassifizierung (100 Calls, ≤ 5 ms)", () => {
    for (const url of SSRF_URL_SAMPLES) {
      validateUrl(url);
    }
  });
});

// --- Fixture generators -----------------------------------------------------
//
// Synthetische Inputs sind deterministisch und versionsstabil — Plan
// §2 DoD-Item 5 verlangt repo-lokale, netzwerkfreie Fixtures. Die
// kleinen Variants/Reps sind in der echten Lab-Welt repräsentativ;
// echte Production-MPDs oder CDN-Manifeste wären Operator-spezifisch
// und damit für Budget-Smokes ungeeignet.

function generateLargeMaster(variantCount: number, renditionCount: number): string {
  const lines: string[] = ["#EXTM3U", "#EXT-X-VERSION:6", ""];
  for (let i = 0; i < renditionCount; i++) {
    lines.push(
      `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aud-${i}",NAME="track ${i}",LANGUAGE="en",DEFAULT=${i === 0 ? "YES" : "NO"},AUTOSELECT=YES,URI="audio/${i}/main.m3u8"`
    );
  }
  lines.push("");
  for (let i = 0; i < variantCount; i++) {
    const bw = 800_000 + i * 250_000;
    const w = 640 + (i % 8) * 160;
    const h = 360 + (i % 8) * 90;
    lines.push(
      `#EXT-X-STREAM-INF:BANDWIDTH=${bw},RESOLUTION=${w}x${h},CODECS="avc1.4d401e,mp4a.40.2",AUDIO="aud-${i % renditionCount}"`
    );
    lines.push(`video/${i}/main.m3u8`);
  }
  return lines.join("\n");
}

function generate1kSegmentMedia(): string {
  const lines: string[] = [
    "#EXTM3U",
    "#EXT-X-VERSION:6",
    "#EXT-X-TARGETDURATION:6",
    "#EXT-X-MEDIA-SEQUENCE:0",
    "#EXT-X-PLAYLIST-TYPE:VOD",
  ];
  for (let i = 0; i < 1_000; i++) {
    lines.push(`#EXTINF:5.997,segment-${i}`);
    lines.push(`segment-${i}.ts`);
  }
  lines.push("#EXT-X-ENDLIST");
  return lines.join("\n");
}

function generateDashVod(): string {
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" mediaPresentationDuration="PT10M0S" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011">',
    "  <Period>",
    '    <AdaptationSet id="1" contentType="video" mimeType="video/mp4">',
    '      <Representation id="v1" bandwidth="1280000" width="1280" height="720" codecs="avc1.4d401e" frameRate="30"/>',
    '      <Representation id="v2" bandwidth="2560000" width="1920" height="1080" codecs="avc1.640028" frameRate="30"/>',
    "    </AdaptationSet>",
    '    <AdaptationSet id="2" contentType="audio" mimeType="audio/mp4" lang="en">',
    '      <Representation id="a1" bandwidth="128000" codecs="mp4a.40.2" audioSamplingRate="48000"/>',
    "    </AdaptationSet>",
    "  </Period>",
    "</MPD>",
    ""
  ].join("\n");
}

function generateDashLive(): string {
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="dynamic" minimumUpdatePeriod="PT2S" availabilityStartTime="2026-05-07T00:00:00Z" profiles="urn:mpeg:dash:profile:isoff-live:2011">',
    "  <Period>",
    '    <AdaptationSet id="0" contentType="video" mimeType="video/mp4">',
    '      <Representation id="v0" bandwidth="1500000" width="1280" height="720" codecs="avc1.4d401e" frameRate="30"/>',
    '      <Representation id="v1" bandwidth="3000000" width="1920" height="1080" codecs="avc1.640028" frameRate="30"/>',
    "    </AdaptationSet>",
    '    <AdaptationSet id="1" contentType="audio" mimeType="audio/mp4" lang="en">',
    '      <Representation id="a0" bandwidth="128000" codecs="mp4a.40.2" audioSamplingRate="48000"/>',
    "    </AdaptationSet>",
    '    <AdaptationSet id="2" contentType="text" mimeType="application/mp4" lang="en">',
    '      <Representation id="s0" bandwidth="1024" codecs="wvtt"/>',
    "    </AdaptationSet>",
    "  </Period>",
    "</MPD>",
    ""
  ].join("\n");
}

function generateDetectorBody(): string {
  // 256 KiB HTML-Body (typischer Negativ-Pfad: Detector klassifiziert
  // als "unsupported"). Pinnt, dass der Detector nicht über Body-
  // Größe skaliert — er sollte nur den Anfang sniffen.
  const filler = "<p>not a manifest</p>".repeat(11_000);
  return `<html><body>${filler}</body></html>`;
}

function generateSsrfUrlSamples(): URL[] {
  // Mix aus Allow- und Block-Pfaden: 50 öffentliche IPv4-/IPv6-/
  // Domain-Hosts (sollten validateUrl positiv durchlaufen) plus 50
  // private/loopback/credentials/scheme-Verstöße.
  const samples: URL[] = [];
  for (let i = 0; i < 25; i++) {
    samples.push(new URL(`https://cdn-${i}.example.test/manifest.m3u8`));
    samples.push(new URL(`https://203.0.113.${i + 1}/m.m3u8`));
  }
  for (let i = 0; i < 50; i++) {
    // Private/loopback-Pfade — validateUrl liefert ok=true (URL-Form
    // ist gültig), validateResolvedIp im Loader-Pfad würde sie
    // ablehnen. Für den Bench-Hot-Path ist nur die URL-Form-Prüfung
    // relevant.
    samples.push(new URL(`https://10.0.${i % 256}.${(i * 13) % 256}/m.m3u8`));
  }
  return samples;
}
