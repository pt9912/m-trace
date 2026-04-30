import { cleanup, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("$lib/api", () => ({
  formatTime: (value: string | undefined) => (value ? `time:${value}` : "n/a"),
  getHealth: vi.fn(async () => ({ ok: true, status: 200, text: "ok" })),
  isErrorEvent: vi.fn(() => false),
  listSessions: vi.fn(async () => ({
    sessions: [
      {
        session_id: "session-1",
        project_id: "demo",
        state: "active",
        started_at: "2026-04-30T00:00:00.000Z",
        last_event_at: "2026-04-30T00:00:02.000Z",
        event_count: 3
      },
      {
        session_id: "session-2",
        project_id: "demo",
        state: "stalled",
        started_at: "2026-04-30T00:00:00.000Z",
        last_event_at: "2026-04-30T00:00:03.000Z",
        event_count: 2
      }
    ]
  }))
}));

afterEach(() => {
  cleanup();
});

describe("overview page", () => {
  it("renders health and session aggregates from the API", async () => {
    const { default: OverviewPage } = await import("../src/routes/+page.svelte");

    render(OverviewPage);

    expect(screen.getByRole("heading", { name: "Live overview" })).toBeTruthy();
    expect(await screen.findByText("session-1")).toBeTruthy();
    expect(screen.getByText("session-2")).toBeTruthy();
    expect(screen.getByText("up")).toBeTruthy();
    expect(screen.getByText("time:2026-04-30T00:00:02.000Z")).toBeTruthy();
  });
});
