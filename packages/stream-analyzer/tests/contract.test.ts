import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

import { analyzeHlsManifest, analyzeManifest } from "../src/index.js";

const here = dirname(fileURLToPath(import.meta.url));
const fixturesRoot = join(here, "..", "..", "..", "spec", "contract-fixtures", "analyzer");
const analyzerFixturePairs = [
  ["success-master.json", "contract-success-master.json"],
  ["success-dash-vod.json", "contract-success-dash-vod.json"],
  ["success-dash-live.json", "contract-success-dash-live.json"],
  ["error-fetch-blocked.json", "contract-error-fetch-blocked.json"],
  ["success-hls-cmaf-vod.json", "contract-success-hls-cmaf-vod.json"],
  ["success-hls-ts-negative.json", "contract-success-hls-ts-negative.json"],
  ["success-hls-master-codecs-only.json", "contract-success-hls-master-codecs-only.json"],
  ["success-hls-map-byterange.json", "contract-success-hls-map-byterange.json"],
  ["success-hls-media-byterange.json", "contract-success-hls-media-byterange.json"],
  ["success-dash-mp4-mime-only.json", "contract-success-dash-mp4-mime-only.json"],
  ["success-dash-cmaf-vod.json", "contract-success-dash-cmaf-vod.json"],
  ["success-dash-no-cmaf-signals.json", "contract-success-dash-no-cmaf-signals.json"],
  ["success-dash-baseurl-inheritance.json", "contract-success-dash-baseurl-inheritance.json"],
  ["success-dash-segmentlist.json", "contract-success-dash-segmentlist.json"],
  ["error-cmaf-binary-validation.json", "contract-error-cmaf-binary-validation.json"],
  ["error-cmaf-invalid-box-structure.json", "contract-error-cmaf-invalid-box-structure.json"],
  ["success-cmaf-skipped-too-large.json", "contract-success-cmaf-skipped-too-large.json"],
  ["success-cmaf-skipped-content-type.json", "contract-success-cmaf-skipped-content-type.json"],
  ["success-cmaf-skipped-binary-disabled.json", "contract-success-cmaf-skipped-binary-disabled.json"],
  ["success-cmaf-skipped-not-planned.json", "contract-success-cmaf-skipped-not-planned.json"]
] as const;

function readContractFixture(name: string): unknown {
  const path = join(fixturesRoot, name);
  return JSON.parse(readFileSync(path, "utf8"));
}

describe("contract fixture parity (TS-Producer vs spec/contract-fixtures/analyzer)", () => {
  it("success-master.json matches the live analyzer output for the canonical master input", async () => {
    const masterManifest = [
      "#EXTM3U",
      "#EXT-X-VERSION:6",
      '#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aud-en",NAME="English",LANGUAGE="en",DEFAULT=YES,AUTOSELECT=YES,URI="audio/en/main.m3u8"',
      '#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=1280x720,CODECS="avc1.4d401e,mp4a.40.2",AUDIO="aud-en"',
      "video/720p/main.m3u8",
      ""
    ].join("\n");
    const result = await analyzeHlsManifest({
      kind: "text",
      text: masterManifest,
      baseUrl: "https://cdn.example.test/"
    });
    const expected = readContractFixture("success-master.json");
    expect(result).toEqual(expected);
  });

  it("success-dash-vod.json matches the live analyzer output for the canonical VOD MPD input", async () => {
    const vodMpd = [
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
    const result = await analyzeManifest({
      kind: "text",
      text: vodMpd,
      baseUrl: "https://cdn.example.test/"
    });
    const expected = readContractFixture("success-dash-vod.json");
    expect(result).toEqual(expected);
  });

  it("success-dash-live.json matches the live analyzer output for a dynamic MPD input", async () => {
    const liveMpd = [
      '<?xml version="1.0" encoding="UTF-8"?>',
      '<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="dynamic" minimumUpdatePeriod="PT2S" availabilityStartTime="2026-05-07T00:00:00Z" profiles="urn:mpeg:dash:profile:isoff-live:2011">',
      "  <Period>",
      '    <AdaptationSet id="0" contentType="video" mimeType="video/mp4">',
      '      <Representation id="v0" bandwidth="1500000" width="1280" height="720" codecs="avc1.4d401e" frameRate="30"/>',
      "    </AdaptationSet>",
      "  </Period>",
      "</MPD>",
      ""
    ].join("\n");
    const result = await analyzeManifest({
      kind: "text",
      text: liveMpd,
      baseUrl: "https://live.example.test/"
    });
    const expected = readContractFixture("success-dash-live.json");
    expect(result).toEqual(expected);
  });

  it("error-fetch-blocked.json shape is structurally what the producer emits", () => {
    // Hand-konstruiertes Fehler-Result entspricht dem Output, den
    // analyze.ts aktuell für SSRF-Blocks generiert (Form gepinnt;
    // konkrete Message und Detail-Werte sind beispielhaft, das
    // Schema-Skelett ist verbindlich).
    const fixture = readContractFixture("error-fetch-blocked.json") as Record<string, unknown>;
    expect(fixture).toMatchObject({
      status: "error",
      analyzerVersion: expect.any(String),
      analyzerKind: "hls",
      code: "fetch_blocked",
      message: expect.any(String),
      details: expect.objectContaining({
        host: expect.any(String),
        address: expect.any(String),
        family: expect.any(Number)
      })
    });
  });

  // Go-Tests können nicht aus `spec/` heraus go:embed nutzen, weil
  // der Docker-Build-Context auf `apps/api/` beschränkt ist. Wir
  // committen deshalb Kopien in `apps/api/.../streamanalyzer/testdata/`
  // und prüfen hier, dass die Kopien byte-gleich mit der Spec-Quelle
  // sind. Drift fällt damit beim ersten workspace-test auf — bevor
  // Go-Tests gegen eine veraltete Kopie grün laufen.
  it.each(analyzerFixturePairs)("Go testdata copy of %s is byte-equal to the spec source", (specName, goName) => {
    const specPath = join(fixturesRoot, specName);
    const goPath = join(
      here,
      "..",
      "..",
      "..",
      "apps",
      "api",
      "adapters",
      "driven",
      "streamanalyzer",
      "testdata",
      goName
    );
    const spec = readFileSync(specPath, "utf8");
    const goCopy = readFileSync(goPath, "utf8");
    expect(goCopy).toBe(spec);
  });

  it("pins the complete CMAF contract fixture inventory", () => {
    expect(analyzerFixturePairs.map(([specName]) => specName).sort()).toEqual([
      "error-cmaf-binary-validation.json",
      "error-cmaf-invalid-box-structure.json",
      "error-fetch-blocked.json",
      "success-cmaf-skipped-binary-disabled.json",
      "success-cmaf-skipped-content-type.json",
      "success-cmaf-skipped-not-planned.json",
      "success-cmaf-skipped-too-large.json",
      "success-dash-baseurl-inheritance.json",
      "success-dash-cmaf-vod.json",
      "success-dash-live.json",
      "success-dash-mp4-mime-only.json",
      "success-dash-no-cmaf-signals.json",
      "success-dash-segmentlist.json",
      "success-dash-vod.json",
      "success-hls-cmaf-vod.json",
      "success-hls-map-byterange.json",
      "success-hls-master-codecs-only.json",
      "success-hls-media-byterange.json",
      "success-hls-ts-negative.json",
      "success-master.json"
    ]);
  });
});
