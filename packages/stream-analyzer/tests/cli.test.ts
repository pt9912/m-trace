import { Writable } from "node:stream";
import { describe, expect, it } from "vitest";

import { runCli, EXIT_OK, EXIT_FAILURE, EXIT_USAGE } from "../src/cli/check.js";
import type { AnalysisErrorResult } from "../src/types/error.js";
import type { AnalysisResult } from "../src/types/result.js";
import type { AnalyzeOptions, ManifestInput } from "../src/types/input.js";

class StringStream extends Writable {
  data = "";
  _write(chunk: Buffer | string, _enc: BufferEncoding, cb: () => void): void {
    this.data += chunk.toString();
    cb();
  }
}

interface RunResult {
  exit: number;
  stdout: string;
  stderr: string;
}

async function run(
  argv: string[],
  options: {
    analyze?: (input: ManifestInput, opts?: AnalyzeOptions) => Promise<AnalysisResult | AnalysisErrorResult>;
    readFile?: (path: string) => Promise<string>;
    env?: Readonly<Record<string, string | undefined>>;
  } = {}
): Promise<RunResult> {
  const stdout = new StringStream();
  const stderr = new StringStream();
  const exit = await runCli({
    argv,
    stdout,
    stderr,
    analyze: options.analyze,
    readFile: options.readFile,
    env: options.env
  });
  return { exit, stdout: stdout.data, stderr: stderr.data };
}

const okMaster: AnalysisResult = {
  status: "ok",
  analyzerVersion: "0.13.0",
  analyzerKind: "hls",
  input: { source: "text" },
  playlistType: "master",
  summary: { itemCount: 0 },
  findings: [],
  details: { variants: [], renditions: [] }
};

const errNotHls: AnalysisErrorResult = {
  status: "error",
  analyzerVersion: "0.13.0",
  analyzerKind: "hls",
  code: "manifest_not_hls",
  message: "not HLS"
};

describe("m-trace CLI — usage", () => {
  it("prints usage on --help and exits 0", async () => {
    const r = await run(["--help"]);
    expect(r.exit).toBe(EXIT_OK);
    expect(r.stdout).toContain("Usage: m-trace check");
    expect(r.stderr).toBe("");
  });

  it("prints usage on -h and exits 0", async () => {
    const r = await run(["-h"]);
    expect(r.exit).toBe(EXIT_OK);
    expect(r.stdout).toContain("Usage: m-trace check");
  });

  it("prints version on --version and exits 0", async () => {
    const r = await run(["--version"]);
    expect(r.exit).toBe(EXIT_OK);
    expect(r.stdout.trim()).toMatch(/^\d+\.\d+\.\d+/);
  });

  it("rejects empty argv with usage error", async () => {
    const r = await run([]);
    expect(r.exit).toBe(EXIT_USAGE);
    expect(r.stderr).toContain("Usage: m-trace check");
  });

  it("rejects unknown command with usage error", async () => {
    const r = await run(["foo"]);
    expect(r.exit).toBe(EXIT_USAGE);
    expect(r.stderr).toContain("unbekanntes Kommando");
  });

  it("rejects missing target with usage error", async () => {
    const r = await run(["check"]);
    expect(r.exit).toBe(EXIT_USAGE);
    expect(r.stderr).toContain("fehlendes Argument");
  });

  it("rejects unexpected trailing args with usage error", async () => {
    const r = await run(["check", "a.m3u8", "extra"], {
      analyze: async () => okMaster,
      readFile: async () => "#EXTM3U\n"
    });
    expect(r.exit).toBe(EXIT_USAGE);
    expect(r.stderr).toContain("unerwartetes Argument");
  });
});

describe("m-trace CLI — file input", () => {
  it("reads the file, dispatches as text input, and prints pretty JSON", async () => {
    let observedInput: ManifestInput | null = null;
    const r = await run(["check", "/tmp/master.m3u8"], {
      analyze: async (input) => {
        observedInput = input;
        return okMaster;
      },
      readFile: async (p) => {
        expect(p).toBe("/tmp/master.m3u8");
        return "#EXTM3U\n";
      }
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observedInput).toMatchObject({ kind: "text", text: "#EXTM3U\n" });
    expect((observedInput as unknown as { baseUrl?: string }).baseUrl).toMatch(/^file:\/\//);
    const parsed = JSON.parse(r.stdout);
    expect(parsed.playlistType).toBe("master");
    // Pretty-printed (mind. ein Indent-Newline am Ende der ersten Zeile).
    expect(r.stdout.split("\n").length).toBeGreaterThan(2);
  });

  it("returns exit 1 with stderr message on file read failure", async () => {
    const r = await run(["check", "/missing.m3u8"], {
      readFile: async () => {
        throw new Error("ENOENT: no such file");
      }
    });
    expect(r.exit).toBe(EXIT_FAILURE);
    expect(r.stderr).toContain("konnte nicht gelesen werden");
    expect(r.stderr).toContain("/missing.m3u8");
    expect(r.stdout).toBe("");
  });

  it("returns exit 1 with JSON on analysis error from a file", async () => {
    const r = await run(["check", "/tmp/x.txt"], {
      analyze: async () => errNotHls,
      readFile: async () => "<html>"
    });
    expect(r.exit).toBe(EXIT_FAILURE);
    const parsed = JSON.parse(r.stdout);
    expect(parsed).toEqual(errNotHls);
  });
});

describe("m-trace CLI — URL input", () => {
  it("dispatches http:// as url input without reading the filesystem", async () => {
    let observedInput: ManifestInput | null = null;
    let readFileCalled = false;
    const r = await run(["check", "http://example.test/m.m3u8"], {
      analyze: async (input) => {
        observedInput = input;
        return okMaster;
      },
      readFile: async () => {
        readFileCalled = true;
        return "";
      }
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observedInput).toEqual({ kind: "url", url: "http://example.test/m.m3u8" });
    expect(readFileCalled).toBe(false);
  });

  it("dispatches https:// as url input", async () => {
    let observedInput: ManifestInput | null = null;
    const r = await run(["check", "https://cdn.test/m.m3u8"], {
      analyze: async (input) => {
        observedInput = input;
        return okMaster;
      }
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observedInput).toEqual({ kind: "url", url: "https://cdn.test/m.m3u8" });
  });

  it("returns exit 1 with JSON on URL analysis error", async () => {
    const errBlocked: AnalysisErrorResult = {
      status: "error",
      analyzerVersion: "0.13.0",
      analyzerKind: "hls",
      code: "fetch_blocked",
      message: "blocked",
      details: { host: "internal.test", address: "10.0.0.1", family: 4 }
    };
    const r = await run(["check", "http://internal.test/m.m3u8"], {
      analyze: async () => errBlocked
    });
    expect(r.exit).toBe(EXIT_FAILURE);
    const parsed = JSON.parse(r.stdout);
    expect(parsed.code).toBe("fetch_blocked");
    expect(parsed.details).toEqual(errBlocked.details);
  });

  it("returns exit 1 when analyze() throws", async () => {
    const r = await run(["check", "https://example.test/m.m3u8"], {
      analyze: async () => {
        throw new Error("network down");
      }
    });
    expect(r.exit).toBe(EXIT_FAILURE);
    expect(r.stderr).toContain("Analyse fehlgeschlagen");
    expect(r.stderr).toContain("network down");
  });

  it("describes non-Error analyze throws via String()", async () => {
    const r = await run(["check", "https://example.test/m.m3u8"], {
      analyze: async () => {
        throw "raw boom";
      }
    });
    expect(r.exit).toBe(EXIT_FAILURE);
    expect(r.stderr).toContain("Analyse fehlgeschlagen");
    expect(r.stderr).toContain("raw boom");
  });

  it("treats uppercase HTTPS as a URL (RFC 3986 §3.1)", async () => {
    let observedInput: ManifestInput | null = null;
    const r = await run(["check", "HTTPS://example.test/m.m3u8"], {
      analyze: async (input) => {
        observedInput = input;
        return okMaster;
      }
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observedInput).toEqual({ kind: "url", url: "HTTPS://example.test/m.m3u8" });
  });
});

// plan-0.9.0 §4 Tranche 3 (RAK-59) — DASH-Pfad: der CLI-Dispatcher
// liefert ein DASH-Result, sobald analyze() ein analyzerKind="dash"-
// Result zurückgibt. Der CLI-Code selbst entscheidet nichts — das
// macht der Detector in analyze.ts.
describe("m-trace CLI — DASH dispatch (RAK-59)", () => {
  const okDash: AnalysisResult = {
    status: "ok",
    analyzerVersion: "0.13.0",
    analyzerKind: "dash",
    input: { source: "text" },
    playlistType: "dash",
    summary: { itemCount: 1 },
    findings: [],
    details: {
      type: "static",
      live: false,
      periodCount: 1,
      adaptationSets: [
        {
          representations: [
            {
              id: "v0",
              bandwidth: 1500000,
              width: 1280,
              height: 720,
              codecs: "avc1.4d401e",
              mimeType: "video/mp4"
            }
          ]
        }
      ]
    }
  };

  it("prints the analyzer DASH result for an .mpd file path", async () => {
    let observedInput: ManifestInput | null = null;
    const r = await run(["check", "/abs/test.mpd"], {
      analyze: async (input) => {
        observedInput = input;
        return okDash;
      },
      readFile: async () =>
        '<?xml version="1.0"?><MPD type="static"><Period><AdaptationSet><Representation id="v0" bandwidth="1500000" width="1280" height="720" codecs="avc1.4d401e" mimeType="video/mp4"/></AdaptationSet></Period></MPD>'
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observedInput).toMatchObject({ kind: "text" });
    const parsed = JSON.parse(r.stdout);
    expect(parsed.analyzerKind).toBe("dash");
    expect(parsed.playlistType).toBe("dash");
    expect(parsed.details.adaptationSets[0].representations).toHaveLength(1);
  });

  it("dispatches an https://-MPD URL identically to HLS URLs", async () => {
    let observedInput: ManifestInput | null = null;
    const r = await run(["check", "https://cdn.example.test/manifest.mpd"], {
      analyze: async (input) => {
        observedInput = input;
        return okDash;
      }
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observedInput).toEqual({
      kind: "url",
      url: "https://cdn.example.test/manifest.mpd"
    });
  });

  it("returns exit 1 for an unsupported manifest body (e.g. HTML)", async () => {
    const errUnsupported: AnalysisErrorResult = {
      status: "error",
      analyzerVersion: "0.13.0",
      analyzerKind: "hls",
      code: "manifest_not_supported",
      message: "Manifest-Body wurde weder als HLS noch als DASH erkannt.",
      details: { firstLine: "<html>" }
    };
    const r = await run(["check", "/abs/index.html"], {
      analyze: async () => errUnsupported,
      readFile: async () => "<html><body>404</body></html>"
    });
    expect(r.exit).toBe(EXIT_FAILURE);
    const parsed = JSON.parse(r.stdout);
    expect(parsed.code).toBe("manifest_not_supported");
  });
});

/**
 * Plan `0.10.0` Tranche 5 (NF-13 / RAK-63): Opt-in-Env-Schalter
 * `MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS` reicht ausschließlich
 * `fetch.allowPrivateNetworks=true` an die Analyzer-Library durch.
 * Default bleibt unverändert; der vorhandene URL-SSRF-Smoke muss
 * ohne dieses Flag weiterhin `fetch_blocked` liefern.
 */
describe("m-trace CLI — MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS opt-in", () => {
  it("does not pass allowPrivateNetworks when env is unset", async () => {
    let observed: AnalyzeOptions | undefined;
    const r = await run(["check", "https://cdn.example.test/m.m3u8"], {
      analyze: async (_, opts) => {
        observed = opts;
        return okMaster;
      },
      env: {}
    });
    expect(r.exit).toBe(EXIT_OK);
    expect(observed).toBeUndefined();
  });

  it("passes fetch.allowPrivateNetworks=true when env is 'true'", async () => {
    let observed: AnalyzeOptions | undefined;
    await run(["check", "https://cdn.example.test/m.m3u8"], {
      analyze: async (_, opts) => {
        observed = opts;
        return okMaster;
      },
      env: { MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS: "true" }
    });
    expect(observed).toEqual({ fetch: { allowPrivateNetworks: true } });
  });

  it.each(["1", "TRUE", "yes", "on", "  true  "])(
    "treats %p as enabled",
    async (value) => {
      let observed: AnalyzeOptions | undefined;
      await run(["check", "https://cdn.example.test/m.m3u8"], {
        analyze: async (_, opts) => {
          observed = opts;
          return okMaster;
        },
        env: { MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS: value }
      });
      expect(observed).toEqual({ fetch: { allowPrivateNetworks: true } });
    }
  );

  it.each(["false", "0", "no", "off", "", "anything-else"])(
    "treats %p as not enabled",
    async (value) => {
      let observed: AnalyzeOptions | undefined;
      await run(["check", "https://cdn.example.test/m.m3u8"], {
        analyze: async (_, opts) => {
          observed = opts;
          return okMaster;
        },
        env: { MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS: value }
      });
      expect(observed).toBeUndefined();
    }
  );

  it("documents the opt-in switch in the help output", async () => {
    const r = await run(["--help"]);
    expect(r.stdout).toContain("MTRACE_CHECK_ALLOW_PRIVATE_NETWORKS");
    expect(r.stdout).toContain("fetch_blocked");
  });
});
