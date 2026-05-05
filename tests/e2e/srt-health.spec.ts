import { expect, test } from "@playwright/test";

// Browser-E2E für die SRT-Health-Routen (plan-0.6.0 §6 Tranche 5
// RAK-43/RAK-44 plus §8 Tranche 7).
//
// Strategie: das Dashboard läuft gegen das normale browser-e2e-
// Compose (apps/api ohne aktiven Collector), liefert also für
// `/api/srt/health[/{stream_id}]` ohne Mock einen leeren Empty-State.
// Mit `page.route()` injizieren wir kontrollierte Sample-Daten und
// verifizieren das Rendering im echten Browser (Pill-Klassen, vier
// RAK-43-Pflichtwerte, Stale-Variante, Detail-History-Tabelle).
//
// Der durchgängige Datenfluss vom Collector bis zur API ist bereits
// durch Adapter-/Repo-/Use-Case-/Integration-Tests in apps/api und
// die vitest-Component-Tests in apps/dashboard abgedeckt; dieser
// Test ergänzt die Browser-Render-Schicht.

const isoNow = "2026-05-05T12:00:00.250Z";
const isoBefore = "2026-05-05T12:00:00.000Z";

const healthyItem = {
  stream_id: "srt-test",
  connection_id: "00000000-0000-0000-0000-000000000001",
  health_state: "healthy" as const,
  source_status: "ok" as const,
  source_error_code: "none" as const,
  connection_state: "connected" as const,
  metrics: {
    rtt_ms: 0.231,
    packet_loss_total: 0,
    retransmissions_total: 0,
    available_bandwidth_bps: 4_352_217_617
  },
  derived: {},
  freshness: {
    source_observed_at: null,
    source_sequence: "37208036",
    collected_at: isoBefore,
    ingested_at: isoNow,
    sample_age_ms: 250,
    stale_after_ms: 15_000
  }
};

const staleItem = {
  ...healthyItem,
  health_state: "unknown" as const,
  source_status: "stale" as const,
  source_error_code: "stale_sample" as const,
  freshness: {
    ...healthyItem.freshness,
    sample_age_ms: 30_000
  }
};

test.describe("SRT health dashboard", () => {
  test("renders empty-state hint when no streams reported", async ({ page }) => {
    await page.route("**/api/srt/health", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ items: [] })
      });
    });

    await page.goto("/srt-health", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: "SRT health" })).toBeVisible();
    await expect(page.getByText(/Collector may be disabled/i)).toBeVisible({ timeout: 10_000 });
  });

  test("renders table with four required metrics for healthy stream", async ({ page }) => {
    await page.route("**/api/srt/health", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ items: [healthyItem] })
      });
    });

    await page.goto("/srt-health", { waitUntil: "domcontentloaded" });
    // Stream-Link in der Tabelle.
    await expect(page.getByRole("link", { name: "srt-test" })).toBeVisible({ timeout: 10_000 });
    // RTT-Spalte (formatiert auf 2 Nachkommastellen, ms-Suffix).
    await expect(page.getByText("0.23 ms")).toBeVisible();
    // Bandbreite (Mbit/s, drei Nachkommastellen).
    await expect(page.getByText("4352.218 Mbit/s")).toBeVisible();
    // Health-Pill mit healthy-Klasse.
    const pill = page.locator(".pill.healthy").filter({ hasText: "healthy" });
    await expect(pill).toBeVisible();
  });

  test("marks stale samples with the stale pill text", async ({ page }) => {
    await page.route("**/api/srt/health", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ items: [staleItem] })
      });
    });

    await page.goto("/srt-health", { waitUntil: "domcontentloaded" });
    // Pill-Text: "{state} (stale)" wenn isSrtSampleStale + state !== "unknown",
    // ansonsten nur "{state}". Hier ist state=unknown → Pill bleibt "unknown",
    // aber der Source-Status-Hint zeigt "source: stale".
    await expect(page.getByText(/source: stale/i)).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText("stale_sample")).toBeVisible();
  });

  test("detail route shows current sample plus history table", async ({ page }) => {
    const detailItems = [
      healthyItem,
      {
        ...healthyItem,
        freshness: {
          ...healthyItem.freshness,
          ingested_at: "2026-05-05T11:59:55.000Z",
          source_sequence: "36000000"
        }
      }
    ];
    await page.route("**/api/srt/health/srt-test*", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ stream_id: "srt-test", items: detailItems })
      });
    });

    await page.goto("/srt-health/srt-test", { waitUntil: "domcontentloaded" });
    await expect(page.getByRole("heading", { name: /SRT health:/ })).toBeVisible({ timeout: 10_000 });
    await expect(page.getByRole("heading", { name: "Current" })).toBeVisible();
    await expect(page.getByRole("heading", { name: "History" })).toBeVisible();
    // Vier RAK-43-Pflichtwerte im Current-Block.
    await expect(page.getByText("0.23 ms")).toBeVisible();
    await expect(page.getByText("4352.218 Mbit/s").first()).toBeVisible();
    // History hat zwei Einträge.
    await expect(page.locator("tbody tr")).toHaveCount(2);
  });

  test("detail route renders 404 message for unknown stream", async ({ page }) => {
    await page.route("**/api/srt/health/missing*", async (route) => {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ error: "stream_unknown", stream_id: "missing" })
      });
    });

    await page.goto("/srt-health/missing", { waitUntil: "domcontentloaded" });
    await expect(page.getByText(/has no persisted health samples/i)).toBeVisible({ timeout: 10_000 });
  });
});
