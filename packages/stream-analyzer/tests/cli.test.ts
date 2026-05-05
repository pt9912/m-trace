import { Writable } from "node:stream";
import { describe, expect, it } from "vitest";

import { runCli, EXIT_OK, EXIT_FAILURE, EXIT_USAGE } from "../src/cli/check.js";
import type { AnalysisErrorResult } from "../src/types/error.js";
import type { AnalysisResult } from "../src/types/result.js";
import type { ManifestInput } from "../src/types/input.js";

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
    analyze?: (input: ManifestInput) => Promise<AnalysisResult | AnalysisErrorResult>;
    readFile?: (path: string) => Promise<string>;
  } = {}
): Promise<RunResult> {
  const stdout = new StringStream();
  const stderr = new StringStream();
  const exit = await runCli({
    argv,
    stdout,
    stderr,
    analyze: options.analyze,
    readFile: options.readFile
  });
  return { exit, stdout: stdout.data, stderr: stderr.data };
}

const okMaster: AnalysisResult = {
  status: "ok",
  analyzerVersion: "0.5.0",
  analyzerKind: "hls",
  input: { source: "text" },
  playlistType: "master",
  summary: { itemCount: 0 },
  findings: [],
  details: { variants: [], renditions: [] }
};

const errNotHls: AnalysisErrorResult = {
  status: "error",
  analyzerVersion: "0.5.0",
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
      analyzerVersion: "0.5.0",
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
