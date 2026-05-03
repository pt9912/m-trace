import { expect, test } from "@playwright/test";

const apiURL = process.env.API_URL ?? "http://localhost:8080";
// Read-Endpunkte sind ab plan-0.4.0 §4.3 tokenpflichtig; Playwright
// schickt den Default-Lab-Token, der zum Project "demo" auflöst.
const apiHeaders = { "X-MTrace-Token": process.env.API_TOKEN ?? "demo-token" };

test("demo player emits events and dashboard renders the session", async ({ browserName, page, request }) => {
  const sessionId = `playwright-${browserName}-${Date.now()}`;

  await page.goto(`/demo?autostart=1&session_id=${sessionId}`, { waitUntil: "domcontentloaded" });
  await expect(page.getByRole("heading", { name: "Demo player" })).toBeVisible();

  await expect
    .poll(
      async () => {
        const response = await request.get(`${apiURL}/api/stream-sessions`, { headers: apiHeaders });
        const body = await response.text();
        return response.ok() && body.includes(sessionId);
      },
      { timeout: 30_000 }
    )
    .toBe(true);

  const detail = await request.get(`${apiURL}/api/stream-sessions/${sessionId}?events_limit=100`, { headers: apiHeaders });
  expect(detail.ok()).toBe(true);
  const payload = (await detail.json()) as { events: Array<{ event_name: string }> };
  expect(payload.events.length).toBeGreaterThan(0);
  const eventName = payload.events[0]?.event_name ?? "";
  expect(eventName).not.toBe("");

  await page.goto(`/sessions/${sessionId}`, { waitUntil: "domcontentloaded" });
  await expect(page.getByText(sessionId)).toBeVisible();
  await expect(page.getByText(/manifest_loaded|segment_loaded|playback_started|startup_time_measured/).first()).toBeVisible();

  await page.goto("/events", { waitUntil: "domcontentloaded" });
  await page.getByLabel("Session filter").selectOption(sessionId);
  await page.getByLabel("Event type filter").selectOption(eventName);
  await expect(page.getByRole("heading", { name: "Events" })).toBeVisible();
  const eventRow = page.getByRole("row", { name: new RegExp(`${eventName} ${sessionId}`) });
  await expect(eventRow).toBeVisible();
});
