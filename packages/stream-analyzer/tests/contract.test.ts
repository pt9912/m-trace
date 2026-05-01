import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

import { analyzeHlsManifest } from "../src/index.js";

const here = dirname(fileURLToPath(import.meta.url));
const fixturesRoot = join(here, "..", "..", "..", "spec", "contract-fixtures", "analyzer");

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
  it.each([
    ["success-master.json", "contract-success-master.json"],
    ["error-fetch-blocked.json", "contract-error-fetch-blocked.json"]
  ])("Go testdata copy of %s is byte-equal to the spec source", (specName, goName) => {
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
});
