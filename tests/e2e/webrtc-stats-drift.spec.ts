import { expect, test } from "@playwright/test";

// plan-0.9.0 §2 Tranche 1 (RAK-56) — Browser-Drift-Smoke gegen das
// `mtrace-webrtc`-Lab. Schließt R-12 als „automatisiert detektiert,
// Drift bricht den Drift-Smoke" und löst die manuelle Drift-Review-
// Pflicht aus `releasing.md` ab.
//
// Ablauf (PeerConnection im Page-Context, WHEP-HTTP-POST aus dem
// Playwright-Node-Context, ohne Adapter-Hook — Plan §0.5 gibt beide
// Pfade frei): eigene `RTCPeerConnection` mit recvonly-Transceivern
// gegen den WHEP-Endpoint (`http://localhost:8892/webrtc-test/whep`),
// nach erfolgreichem Handshake `pc.getStats()` aufrufen und gegen den
// Spec-§3.5.2-Muss-Felder-Korpus plus die §1.4-Enum-Allowlist
// validieren. Der Node-POST vermeidet Browser-CORS-Abhängigkeiten des
// MediaMTX-Lab-Endpoints; die WebRTC-Stats stammen weiterhin aus dem
// echten Browser.
//
// Validierungen:
//   1. `pc.connectionState` ∈ §1.4 `connection_state`-Allowlist und
//      ist im Endzustand nicht `failed`/`closed`. Der Snapshot-
//      Endzustand ist Soll-Verhalten — eine vollständig durchlaufene
//      Connection kann während der `sampleCollectMs`-Sample-Phase
//      legitim zwischen `connected` und `disconnected` flappen
//      (ICE-Reconfig/Path-Switch). Echte Verbindungsabbrüche
//      (`failed`/`closed`) brechen den Smoke weiter hart.
//      Soll-Erwartung `connected` wird als `[drift-soll]` geloggt.
//   2. `pc.iceConnectionState` ∈ §1.4 `ice_state`-Allowlist.
//   3. Für die stabilen `RTCStatsType`-Gruppen aus §3.5.2
//      (candidate-pair, inbound-rtp) sind die Muss-Felder vorhanden
//      (Drift = Fehler). `peer-connection.connectionState` wird über
//      die normative `pc.connectionState`-API geprüft, weil aktuelle
//      Browser das Feld nicht durchgängig im Stats-Report spiegeln.
//   4. `transport.dtlsState` ∈ §1.4 `dtls_state`-Allowlist, wenn
//      der Browser `RTCStatsType.transport` exponiert; fehlende
//      Transport-Reports werden als Engine-spezifischer Soll-Drift
//      geloggt.
//   5. Soll-Felder werden geloggt, aber nicht release-blockierend
//      geprüft (§3.5.2 Soll-Spalte).
//
// outbound-rtp ist legitim leer, weil der Smoke nur recvonly-
// Transceivers verhandelt; in einer echten Bidi-Session wäre der
// Report verpflichtend. Der Smoke skippt outbound-rtp deshalb.
//
// Opt-in via Env: setzt `MTRACE_WEBRTC_STATS_DRIFT=1` der Smoke-
// Skript (`scripts/smoke-webrtc-stats-drift.sh`). Ohne diese Env
// skippt der Test, damit `make browser-e2e` (anderer Stack, kein
// `mtrace-webrtc`-Lab) ihn nicht aufruft.

const driftActive = process.env.MTRACE_WEBRTC_STATS_DRIFT === "1";

const WHEP_URL =
  process.env.MTRACE_WEBRTC_DRIFT_WHEP_URL ??
  "http://localhost:8892/webrtc-test/whep";
const HANDSHAKE_TIMEOUT_MS = Number(
  process.env.MTRACE_WEBRTC_DRIFT_HANDSHAKE_TIMEOUT_MS ?? "20000"
);
const SAMPLE_COLLECT_MS = Number(
  process.env.MTRACE_WEBRTC_DRIFT_SAMPLE_COLLECT_MS ?? "1500"
);

const ALLOWED_CONNECTION_STATES = new Set<string>([
  "new",
  "connecting",
  "connected",
  "disconnected",
  "failed",
  "closed",
]);
const ALLOWED_ICE_STATES = new Set<string>([
  "new",
  "checking",
  "connected",
  "completed",
  "failed",
  "disconnected",
  "closed",
]);
const ALLOWED_DTLS_STATES = new Set<string>([
  "new",
  "connecting",
  "connected",
  "closed",
  "failed",
]);

const REQUIRED_FIELDS_BY_TYPE: Record<string, readonly string[]> = {
  "candidate-pair": ["state"],
  "inbound-rtp": ["packetsLost", "bytesReceived"],
  "outbound-rtp": ["bytesSent"],
};

const SOLL_FIELDS_BY_TYPE: Record<string, readonly string[]> = {
  "peer-connection": ["dataChannelsOpened", "dataChannelsClosed"],
  transport: ["selectedCandidatePairChanges"],
  "candidate-pair": ["roundTripTime", "availableOutgoingBitrate"],
  "inbound-rtp": ["jitter", "roundTripTime", "framesDecoded", "framesPerSecond"],
  "outbound-rtp": ["framesPerSecond"],
};

interface DriftReport {
  id: string;
  type: string;
  fields: Record<string, unknown>;
}

interface DriftPayload {
  connectionState: string;
  iceConnectionState: string;
  reports: DriftReport[];
}

interface DriftWindow extends Window {
  __mtraceDriftPc?: RTCPeerConnection;
}

test.describe("WebRTC getStats() drift smoke (RAK-56)", () => {
  test.skip(
    !driftActive,
    "drift-smoke is opt-in via make smoke-webrtc-stats-drift (sets MTRACE_WEBRTC_STATS_DRIFT=1)"
  );

  test("getStats() reports include §3.5.2 must-fields and §1.4 enum values", async ({
    page,
    browserName,
  }) => {
    await page.goto("about:blank");

    const localSdp = (await page.evaluate(async ({ includeVideo }) => {
      const driftWindow = window as DriftWindow;
      const pc = new RTCPeerConnection();
      driftWindow.__mtraceDriftPc = pc;

      if (includeVideo) {
        pc.addTransceiver("video", { direction: "recvonly" });
      }
      pc.addTransceiver("audio", { direction: "recvonly" });

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      // Wait for ICE-gathering to be complete (Plan-§0.5: WHEP-
      // POST mit dem vollständigen Local-SDP, kein Trickle).
      await new Promise<void>((resolve) => {
        if (pc.iceGatheringState === "complete") {
          resolve();
          return;
        }
        const onChange = () => {
          if (pc.iceGatheringState === "complete") {
            pc.removeEventListener("icegatheringstatechange", onChange);
            resolve();
          }
        };
        pc.addEventListener("icegatheringstatechange", onChange);
        // Safety timeout for browsers that never reach "complete".
        setTimeout(resolve, 3_000);
      });

      return pc.localDescription?.sdp ?? "";
    }, { includeVideo: browserName !== "firefox" })) as string;

    expect(localSdp, `Browser ${browserName} created an empty WHEP offer`).not.toBe("");

    const response = await fetch(WHEP_URL, {
      method: "POST",
      headers: {
        "Accept": "application/sdp",
        "Content-Type": "application/sdp"
      },
      body: localSdp,
    });
    if (!response.ok) {
      const body = await response.text();
      throw new Error(
        `WHEP handshake failed for browser ${browserName}: status=${response.status} body=${body.slice(0, 200)}`
      );
    }
    const answerSdp = await response.text();

    const payload = (await page.evaluate(
      async ({ answerSdp, handshakeTimeoutMs, sampleCollectMs }) => {
        const pc = (window as DriftWindow).__mtraceDriftPc;
        if (pc === undefined) {
          throw new Error("drift peer connection missing before WHEP answer");
        }
        try {
          await pc.setRemoteDescription({ type: "answer", sdp: answerSdp });

          const deadline = performance.now() + handshakeTimeoutMs;
          while (
            pc.connectionState !== "connected" &&
            pc.connectionState !== "failed" &&
            pc.connectionState !== "closed" &&
            performance.now() < deadline
          ) {
            await new Promise((r) => setTimeout(r, 100));
          }

          // Let inbound-rtp accumulate at least one frame/audio packet counter.
          await new Promise((r) => setTimeout(r, sampleCollectMs));

          const stats = await pc.getStats();
          const reports: Array<{
            id: string;
            type: string;
            fields: Record<string, unknown>;
          }> = [];
          stats.forEach((stat) => {
            const fields: Record<string, unknown> = {};
            for (const key of Object.keys(stat)) {
              fields[key] = (stat as Record<string, unknown>)[key];
            }
            reports.push({ id: String(stat.id), type: String(stat.type), fields });
          });

          return {
            connectionState: pc.connectionState,
            iceConnectionState: pc.iceConnectionState,
            reports,
          };
        } finally {
          try {
            pc.close();
          } catch {
            // ignore
          }
          delete (window as DriftWindow).__mtraceDriftPc;
        }
      },
      {
        answerSdp,
        handshakeTimeoutMs: HANDSHAKE_TIMEOUT_MS,
        sampleCollectMs: SAMPLE_COLLECT_MS,
      }
    )) as DriftPayload;

    expect(
      ALLOWED_CONNECTION_STATES.has(payload.connectionState),
      `Browser ${browserName} reports unknown connectionState=${payload.connectionState} (§1.4 connection_state allowlist)`
    ).toBe(true);
    expect(
      ALLOWED_ICE_STATES.has(payload.iceConnectionState),
      `Browser ${browserName} reports unknown iceConnectionState=${payload.iceConnectionState} (§1.4 ice_state allowlist)`
    ).toBe(true);
    // Harte Failure-Modi brechen den Smoke; ein Snapshot-Endzustand
    // `disconnected` ist hingegen legitim (Path-Switch/ICE-Reconfig
    // während der `sampleCollectMs`-Sample-Phase, §1.4-Allowlist).
    // Soll-Erwartung `connected` wird als `[drift-soll]` geloggt,
    // damit das Signal in Trend-Reviews sichtbar bleibt.
    // Hintergrund: Nightly-Run 26858728018 (2026-06-03) failte auf
    // Firefox unter CI-Last mit Snapshot=`disconnected`, lokal 3/3
    // grün. plan-0.22.3-webrtc-drift dokumentiert die Charakterisierung.
    expect(
      payload.connectionState !== "failed" && payload.connectionState !== "closed",
      `Browser ${browserName} reports release-blocking connectionState=${payload.connectionState}`
    ).toBe(true);
    if (payload.connectionState !== "connected") {
      console.log(
        `[drift-soll] Browser ${browserName} snapshot connectionState=${payload.connectionState} (≠ soll "connected"); §1.4 allowlist-konform, legitimer Pfad-Wechsel`
      );
    }

    const reportsByType = new Map<string, DriftReport[]>();
    for (const report of payload.reports) {
      const list = reportsByType.get(report.type) ?? [];
      list.push(report);
      reportsByType.set(report.type, list);
    }

    // Wenn der Snapshot-Endzustand ≠ `connected` ist (legitimer
    // Pfad-Wechsel während der `sampleCollectMs`-Sample-Phase),
    // sind unvollständige Stat-Reports erwartet — eine `disconnected`
    // PeerConnection liefert keine aktiven `inbound-rtp`-Counter.
    // In dem Fall werden die Required-Type-Drift-Checks zu
    // `[drift-soll]`-Logs herabgestuft; echte Schema-Drifts im
    // `connected`-Snapshot bleiben weiterhin release-blockierend.
    // Hintergrund: Folge-Symptom desselben CI-Flakes wie der
    // connectionState-Snapshot (Nightly-Run 26858728018, dispatch
    // Re-Run 26865980921).
    const snapshotConnected = payload.connectionState === "connected";
    const drift: string[] = [];
    for (const [type, requiredFields] of Object.entries(REQUIRED_FIELDS_BY_TYPE)) {
      const reports = reportsByType.get(type) ?? [];
      if (reports.length === 0) {
        if (type === "outbound-rtp") {
          // recvonly handshake → no outbound-rtp; legitimate skip.
          continue;
        }
        const note = `Browser ${browserName} dropped RTCStatsType.${type} entirely`;
        if (!snapshotConnected) {
          console.log(
            `[drift-soll] ${note} (snapshot connectionState=${payload.connectionState}; required-type-check herabgestuft)`
          );
        } else {
          drift.push(note);
        }
        continue;
      }
      for (const report of reports) {
        for (const field of requiredFields) {
          if (!(field in report.fields)) {
            const note = `Browser ${browserName} dropped field ${field} from RTCStatsType.${type} (id=${report.id})`;
            if (!snapshotConnected) {
              console.log(
                `[drift-soll] ${note} (snapshot connectionState=${payload.connectionState}; required-field-check herabgestuft)`
              );
            } else {
              drift.push(note);
            }
          }
        }
      }
    }

    const transportReports = reportsByType.get("transport") ?? [];
    if (transportReports.length === 0) {
      console.log(
        `[drift-soll] Browser ${browserName} drops RTCStatsType.transport entirely; dtlsState validation skipped`
      );
    }
    for (const transport of transportReports) {
      const dtls = transport.fields["dtlsState"];
      if (typeof dtls !== "string") {
        const note = `Browser ${browserName} dropped field dtlsState from RTCStatsType.transport (id=${transport.id})`;
        if (!snapshotConnected) {
          console.log(
            `[drift-soll] ${note} (snapshot connectionState=${payload.connectionState}; dtls-check herabgestuft)`
          );
        } else {
          drift.push(note);
        }
      } else if (!ALLOWED_DTLS_STATES.has(dtls)) {
        // Allowlist-Verletzung ist auch bei nicht-`connected` ein
        // echtes Schema-Drift-Signal (unbekannter Enum-Wert) — bleibt
        // hart.
        drift.push(
          `Browser ${browserName} reports unknown dtlsState=${dtls} (§1.4 dtls_state allowlist; transport id=${transport.id})`
        );
      }
    }

    const sollMissing: string[] = [];
    for (const [type, sollFields] of Object.entries(SOLL_FIELDS_BY_TYPE)) {
      const reports = reportsByType.get(type) ?? [];
      for (const report of reports) {
        for (const field of sollFields) {
          if (!(field in report.fields)) {
            sollMissing.push(
              `Browser ${browserName} drops soll-field ${field} from RTCStatsType.${type} (id=${report.id})`
            );
          }
        }
      }
    }

    for (const note of sollMissing) {
      // Soll-Felder sind opt-in pro Engine (§3.5.3) — Logs only,
      // nicht release-blockierend.
      console.log(`[drift-soll] ${note}`);
    }

    expect(drift, drift.join("\n")).toEqual([]);
  });
});
