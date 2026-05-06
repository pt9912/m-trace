import { expect, test } from "@playwright/test";

// plan-0.8.0 §5 Tranche 4 (RAK-55, Kann) — Browser-E2E-Smoke gegen
// die /demo-webrtc-Route. Standard-Browser-E2E-Stack startet kein
// `mtrace-webrtc`-Lab-Compose; der Test verifiziert deshalb den
// **Fehlerpfad** des WebRTC-Adapters (whep_signaling_failed) als
// produktiv funktionierendes Tracking. Wenn der Operator den
// `mtrace-webrtc`-Stack zusätzlich startet und `MTRACE_WEBRTC_LAB=1`
// setzt, prüft der Test stattdessen den Happy-Path über
// playback_started.

const apiURL = process.env.API_URL ?? "http://localhost:8080";
const apiHeaders = { "X-MTrace-Token": process.env.API_TOKEN ?? "demo-token" };
const labActive = process.env.MTRACE_WEBRTC_LAB === "1";

test("demo-webrtc adapter emits webrtc.* events into the session timeline", async ({ browserName, page, request }) => {
  const sessionId = `playwright-webrtc-${browserName}-${Date.now()}`;

  await page.goto(`/demo-webrtc?autostart=1&session_id=${sessionId}`, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("heading", { name: "Demo player (WebRTC / WHEP)" })).toBeVisible();

  // Der Adapter sendet entweder playback_started (Lab aktiv) oder
  // playback_error (Lab nicht aktiv); beides taucht im Session-
  // Detail mit reservierten webrtc.*-Meta-Keys auf.
  await expect
    .poll(
      async () => {
        const response = await request.get(
          `${apiURL}/api/stream-sessions/${sessionId}?events_limit=20`,
          { headers: apiHeaders }
        );
        if (!response.ok()) {
          return false;
        }
        const payload = (await response.json()) as { events: Array<{ event_name: string; meta?: Record<string, unknown> }> };
        return payload.events.some((event) => {
          const meta = event.meta ?? {};
          return typeof meta["webrtc.peer_connection_run_id"] === "string";
        });
      },
      { timeout: 30_000 }
    )
    .toBe(true);

  const detail = await request.get(
    `${apiURL}/api/stream-sessions/${sessionId}?events_limit=20`,
    { headers: apiHeaders }
  );
  expect(detail.ok()).toBe(true);
  const payload = (await detail.json()) as {
    events: Array<{ event_name: string; meta?: Record<string, unknown> }>;
  };
  const webrtcEvents = payload.events.filter(
    (event) => typeof (event.meta ?? {})["webrtc.peer_connection_run_id"] === "string"
  );
  expect(webrtcEvents.length).toBeGreaterThan(0);

  const expectedEventName = labActive ? "playback_started" : "playback_error";
  const matched = webrtcEvents.some((event) => event.event_name === expectedEventName);
  expect(matched, `expected at least one ${expectedEventName} event with webrtc.* meta`).toBe(true);
});
