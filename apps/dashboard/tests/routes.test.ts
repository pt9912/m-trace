import { cleanup, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";

const sessions = [
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
];

const events = [
  {
    event_name: "playback_error",
    project_id: "demo",
    session_id: "session-1",
    client_timestamp: "2026-04-30T00:00:01.000Z",
    server_received_at: "2026-04-30T00:00:02.000Z",
    ingest_sequence: 2,
    sequence_number: 2,
    sdk: { name: "@npm9912/player-sdk", version: "0.2.0" }
  },
  {
    event_name: "rebuffer_started",
    project_id: "demo",
    session_id: "session-1",
    client_timestamp: "2026-04-30T00:00:00.000Z",
    server_received_at: "2026-04-30T00:00:01.000Z",
    ingest_sequence: 1,
    sequence_number: 1,
    sdk: { name: "@npm9912/player-sdk", version: "0.2.0" }
  }
];

vi.mock("$lib/api", () => ({
  formatTime: (value: string | undefined) => (value ? `time:${value}` : "n/a"),
  getHealth: vi.fn(async () => ({ ok: true, status: 200, text: "ok" })),
  getSession: vi.fn(async (sessionId: string) => ({
    session: sessions.find((session) => session.session_id === sessionId) ?? sessions[0],
    events
  })),
  isErrorEvent: vi.fn((event: { event_name: string }) => event.event_name.includes("error") || event.event_name.includes("warning")),
  listSessions: vi.fn(async () => ({ sessions }))
}));

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("dashboard route components", () => {
  it("renders the sessions table", async () => {
    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");

    render(SessionsPage);

    expect(screen.getByRole("heading", { name: "Sessions" })).toBeTruthy();
    expect(await screen.findByText("session-1")).toBeTruthy();
    expect(screen.getByText("stalled")).toBeTruthy();
  });

  it("renders the events table with event type filters", async () => {
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);

    expect(screen.getByRole("heading", { name: "Events" })).toBeTruthy();
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    expect(screen.getAllByText("rebuffer_started")).not.toHaveLength(0);
    expect(screen.getByLabelText("Event type filter")).toBeTruthy();
  });

  it("renders only error and warning events on the errors page", async () => {
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);

    expect(screen.getByRole("heading", { name: "Errors" })).toBeTruthy();
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    expect(screen.queryByText("rebuffer_started")).toBeNull();
  });

  it("renders API and observability status", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 204 })));
    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");

    render(StatusPage);

    expect(screen.getByRole("heading", { name: "System status" })).toBeTruthy();
    expect(await screen.findByText("OTel Collector")).toBeTruthy();
    expect(screen.getByText("Prometheus")).toBeTruthy();
    expect(screen.getByText("Grafana")).toBeTruthy();
  });
});
