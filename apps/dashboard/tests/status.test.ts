import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";
import {
  clearLastReadError,
  configuredServiceUrl,
  lastReadError,
  recordReadError,
  sseConnection
} from "../src/lib/status";

beforeEach(() => {
  clearLastReadError();
  sseConnection.set({ state: "not_yet_connected", changedAt: null, detail: null });
});

afterEach(() => {
  clearLastReadError();
});

describe("dashboard status module (§5 H3)", () => {
  it("starts with no last-read-error", () => {
    expect(get(lastReadError)).toBeNull();
  });

  it("records Error instances with their message", () => {
    recordReadError("/api/stream-sessions", new Error("boom"));
    const rec = get(lastReadError);
    expect(rec?.source).toBe("/api/stream-sessions");
    expect(rec?.message).toBe("boom");
    expect(typeof rec?.occurredAt).toBe("string");
  });

  it("records non-Error throws as a stringified message", () => {
    recordReadError("/api/health", "offline");
    expect(get(lastReadError)?.message).toBe("offline");
  });

  it("clearLastReadError resets the store", () => {
    recordReadError("/api/health", new Error("x"));
    clearLastReadError();
    expect(get(lastReadError)).toBeNull();
  });

  it("starts SSE connection in `not_yet_connected`", () => {
    expect(get(sseConnection).state).toBe("not_yet_connected");
  });

  it("returns undefined for unset service-link env keys", () => {
    // Test environment liefert keine PUBLIC_*_URL-Variablen.
    expect(configuredServiceUrl("PUBLIC_GRAFANA_URL")).toBeUndefined();
    expect(configuredServiceUrl("PUBLIC_PROMETHEUS_URL")).toBeUndefined();
    expect(configuredServiceUrl("PUBLIC_MEDIAMTX_URL")).toBeUndefined();
    expect(configuredServiceUrl("PUBLIC_OTEL_HEALTH_URL")).toBeUndefined();
  });
});

describe("buildServiceLinks (§5 F-40 binary)", () => {
  it("marks services as not_configured without env URLs", async () => {
    const { buildServiceLinks } = await import("../src/lib/status");
    const links = buildServiceLinks({});
    expect(links).toHaveLength(4);
    for (const link of links) {
      expect(link.status).toBe("not_configured");
      expect(link.openUrl).toBeUndefined();
      expect(link.probeUrl).toBeUndefined();
    }
  });

  it("derives probe URLs and open URLs from configured env", async () => {
    const { buildServiceLinks } = await import("../src/lib/status");
    const links = buildServiceLinks({
      PUBLIC_GRAFANA_URL: "https://grafana.test/",
      PUBLIC_PROMETHEUS_URL: "https://prom.test",
      PUBLIC_OTEL_HEALTH_URL: "https://otel.test/health",
      PUBLIC_MEDIAMTX_URL: "https://mediamtx.test"
    });
    const grafana = links.find((l) => l.name === "Grafana");
    expect(grafana?.status).toBe("inactive");
    // openUrl behält den Trailing-Slash der Env-Variable bei;
    // probeUrl normalisiert ihn weg (siehe probePath in
    // serviceDefinitions).
    expect(grafana?.openUrl).toBe("https://grafana.test/");
    expect(grafana?.probeUrl).toBe("https://grafana.test/api/health");

    const prom = links.find((l) => l.name === "Prometheus");
    expect(prom?.probeUrl).toBe("https://prom.test/-/ready");

    const otel = links.find((l) => l.name === "OTel Collector");
    // hasOpenUrl=false → kein Open-Link, aber probeUrl bleibt.
    expect(otel?.openUrl).toBeUndefined();
    expect(otel?.probeUrl).toBe("https://otel.test/health");

    const media = links.find((l) => l.name === "MediaMTX");
    // Default hasOpenUrl=true und kein probePath → openUrl=origin,
    // probeUrl=origin.
    expect(media?.openUrl).toBe("https://mediamtx.test");
    expect(media?.probeUrl).toBe("https://mediamtx.test");
  });

  it("trims whitespace-only env values to undefined", async () => {
    const { configuredServiceUrl, buildServiceLinks } = await import("../src/lib/status");
    expect(configuredServiceUrl("PUBLIC_GRAFANA_URL", { PUBLIC_GRAFANA_URL: "   " })).toBeUndefined();
    const links = buildServiceLinks({ PUBLIC_GRAFANA_URL: "   " });
    const grafana = links.find((l) => l.name === "Grafana");
    expect(grafana?.status).toBe("not_configured");
  });
});

describe("probeServices (§5 H3 reachability)", () => {
  it("leaves not_configured services untouched", async () => {
    const { buildServiceLinks, probeServices } = await import("../src/lib/status");
    const links = buildServiceLinks();
    const after = await probeServices(links);
    expect(after.every((l) => l.status === "not_configured")).toBe(true);
  });

  it("marks reachable services as connected and unreachable as inactive", async () => {
    const { probeServices } = await import("../src/lib/status");
    const fetchSeq: Array<() => Promise<Response>> = [
      async () => new Response(null, { status: 200 }),
      async () => {
        throw new Error("offline");
      }
    ];
    let i = 0;
    globalThis.fetch = async () => {
      const next = fetchSeq[i++] ?? fetchSeq[0];
      return next();
    };

    const after = await probeServices([
      {
        name: "A",
        envKey: "PUBLIC_GRAFANA_URL",
        configHint: "PUBLIC_GRAFANA_URL",
        probeUrl: "https://a.test/probe",
        openUrl: "https://a.test",
        status: "inactive"
      },
      {
        name: "B",
        envKey: "PUBLIC_PROMETHEUS_URL",
        configHint: "PUBLIC_PROMETHEUS_URL",
        probeUrl: "https://b.test/probe",
        openUrl: "https://b.test",
        status: "inactive"
      }
    ]);
    expect(after[0].status).toBe("connected");
    expect(after[1].status).toBe("inactive");
  });
});
