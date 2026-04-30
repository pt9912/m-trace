import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const routeState = vi.hoisted<{ params: { id?: string } }>(() => ({
  params: { id: "session-1" }
}));

const apiState = vi.hoisted(() => ({
  sessions: [] as Array<{
    session_id: string;
    project_id: string;
    state: string;
    started_at: string;
    last_event_at: string;
    event_count: number;
  }>,
  events: [] as Array<{
    event_name: string;
    project_id: string;
    session_id: string;
    client_timestamp: string;
    server_received_at: string;
    ingest_sequence: number;
    sequence_number?: number;
    sdk: { name: string; version: string };
  }>,
  listSessionsError: undefined as Error | undefined,
  getSessionError: undefined as Error | undefined,
  health: { ok: true, status: 200, text: "ok" }
}));

vi.mock("$app/stores", () => ({
  page: {
    subscribe(run: (value: { params: { id?: string } }) => void) {
      run(routeState);
      return () => undefined;
    }
  }
}));

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
    sdk: { name: "@npm9912/player-sdk", version: "0.2.0" }
  }
];

vi.mock("$lib/api", () => ({
  formatTime: (value: string | undefined) => (value ? `time:${value}` : "n/a"),
  getHealth: vi.fn(async () => apiState.health),
  getSession: vi.fn(async (sessionId: string) => {
    if (apiState.getSessionError) {
      throw apiState.getSessionError;
    }
    return {
      session: apiState.sessions.find((session) => session.session_id === sessionId) ?? apiState.sessions[0],
      events: apiState.events
    };
  }),
  isErrorEvent: vi.fn((event: { event_name: string }) => event.event_name.includes("error") || event.event_name.includes("warning")),
  listSessions: vi.fn(async () => {
    if (apiState.listSessionsError) {
      throw apiState.listSessionsError;
    }
    return { sessions: apiState.sessions };
  })
}));

beforeEach(() => {
  routeState.params = { id: "session-1" };
  apiState.sessions = sessions;
  apiState.events = events;
  apiState.listSessionsError = undefined;
  apiState.getSessionError = undefined;
  apiState.health = { ok: true, status: 200, text: "ok" };
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("dashboard route components", () => {
  it("renders the app layout navigation", async () => {
    await import("../src/routes/+layout");
    const { default: Layout } = await import("../src/routes/+layout.svelte");

    render(Layout);

    expect(screen.getByText("m-trace")).toBeTruthy();
    expect(screen.getByRole("navigation", { name: "Main navigation" })).toBeTruthy();
    expect(screen.getByText("Demo player")).toBeTruthy();
  });

  it("renders the sessions table", async () => {
    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");

    render(SessionsPage);

    expect(screen.getByRole("heading", { name: "Sessions" })).toBeTruthy();
    expect(await screen.findByText("session-1")).toBeTruthy();
    expect(screen.getByText("stalled")).toBeTruthy();
  });

  it("filters sessions by state and renders the empty branch", async () => {
    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");

    render(SessionsPage);
    expect(await screen.findByText("session-1")).toBeTruthy();
    await fireEvent.change(screen.getByLabelText("State filter"), { target: { value: "ended" } });

    expect(screen.getByText("No matching sessions.")).toBeTruthy();
    expect(screen.queryByText("session-1")).toBeNull();
  });

  it("renders session loading errors", async () => {
    apiState.listSessionsError = new Error("session load failed");
    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");

    render(SessionsPage);

    expect(await screen.findByText("session load failed")).toBeTruthy();
    expect(screen.getByText("No matching sessions.")).toBeTruthy();
  });

  it("renders fallback text for unknown session loading errors", async () => {
    apiState.listSessionsError = "not an Error" as unknown as Error;
    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");

    render(SessionsPage);

    expect(await screen.findByText("Could not load sessions")).toBeTruthy();
  });

  it("renders the events table with event type filters", async () => {
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);

    expect(screen.getByRole("heading", { name: "Events" })).toBeTruthy();
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    expect(screen.getAllByText("rebuffer_started")).not.toHaveLength(0);
    expect(screen.getByLabelText("Event type filter")).toBeTruthy();
  });

  it("filters events by session and renders empty event results", async () => {
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    await fireEvent.change(screen.getByLabelText("Session filter"), { target: { value: "session-2" } });

    expect(screen.getByText("No matching events.")).toBeTruthy();
  });

  it("filters events by a matching session", async () => {
    apiState.events = [
      ...events,
      {
        event_name: "segment_loaded",
        project_id: "demo",
        session_id: "session-2",
        client_timestamp: "2026-04-30T00:00:04.000Z",
        server_received_at: "2026-04-30T00:00:04.000Z",
        ingest_sequence: 4,
        sdk: { name: "@npm9912/player-sdk", version: "0.2.0" }
      }
    ];
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);
    expect(await screen.findAllByText("segment_loaded")).not.toHaveLength(0);
    await fireEvent.change(screen.getByLabelText("Session filter"), { target: { value: "session-2" } });

    expect(screen.getAllByText("segment_loaded")).not.toHaveLength(0);
  });

  it("filters events by event type", async () => {
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    await fireEvent.change(screen.getByLabelText("Event type filter"), { target: { value: "rebuffer_started" } });

    expect(screen.getByText("2 of 4 loaded")).toBeTruthy();
    expect(screen.getAllByText("rebuffer_started")).not.toHaveLength(0);
  });

  it("renders event loading errors", async () => {
    apiState.getSessionError = new Error("events failed");
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);

    expect(await screen.findByText("events failed")).toBeTruthy();
    expect(screen.getByText("No matching events.")).toBeTruthy();
  });

  it("renders fallback text for unknown event loading errors", async () => {
    apiState.getSessionError = "not an Error" as unknown as Error;
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);

    expect(await screen.findByText("Could not load events")).toBeTruthy();
  });

  it("renders events with no recent sessions", async () => {
    apiState.sessions = [];
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);

    expect(await screen.findByText("No matching events.")).toBeTruthy();
  });

  it("refreshes event data from the toolbar", async () => {
    const { default: EventsPage } = await import("../src/routes/events/+page.svelte");

    render(EventsPage);
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    apiState.events = [
      {
        event_name: "segment_loaded",
        project_id: "demo",
        session_id: "session-2",
        client_timestamp: "2026-04-30T00:00:04.000Z",
        server_received_at: "2026-04-30T00:00:04.000Z",
        ingest_sequence: 4,
        sdk: { name: "@npm9912/player-sdk", version: "0.2.0" }
      }
    ];
    await fireEvent.click(screen.getByRole("button", { name: "Refresh" }));

    expect(await screen.findAllByText("segment_loaded")).not.toHaveLength(0);
  });

  it("renders only error and warning events on the errors page", async () => {
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);

    expect(screen.getByRole("heading", { name: "Errors" })).toBeTruthy();
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    expect(screen.queryByText("rebuffer_started")).toBeNull();
  });

  it("renders the no-errors branch", async () => {
    apiState.events = [events[1]];
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);

    expect(await screen.findByText("No playback errors found.")).toBeTruthy();
  });

  it("renders error loading failures", async () => {
    apiState.listSessionsError = new Error("errors failed");
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);

    expect(await screen.findByText("errors failed")).toBeTruthy();
    expect(screen.getByText("No playback errors found.")).toBeTruthy();
  });

  it("renders error loading failures from session details", async () => {
    apiState.getSessionError = new Error("error details failed");
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);

    expect(await screen.findByText("error details failed")).toBeTruthy();
  });

  it("refreshes error data from the toolbar", async () => {
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);
    expect(await screen.findAllByText("playback_error")).not.toHaveLength(0);
    apiState.events = [events[1]];
    await fireEvent.click(screen.getByRole("button", { name: "Refresh" }));

    expect(await screen.findByText("No playback errors found.")).toBeTruthy();
  });

  it("renders unknown error loading failures", async () => {
    apiState.listSessionsError = "not an Error" as unknown as Error;
    const { default: ErrorsPage } = await import("../src/routes/errors/+page.svelte");

    render(ErrorsPage);

    expect(await screen.findByText("Could not load errors")).toBeTruthy();
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

  it("renders inactive observability services", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new Error("offline");
      })
    );
    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");

    render(StatusPage);

    expect(await screen.findByText("OTel Collector")).toBeTruthy();
    expect(screen.getAllByText("inactive").length).toBeGreaterThan(0);
  });

  it("renders a session detail timeline", async () => {
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");

    render(SessionDetailPage);

    expect(screen.getByRole("heading", { name: "Session detail" })).toBeTruthy();
    expect(await screen.findByText("playback_error")).toBeTruthy();
    expect(screen.getByText("rebuffer_started")).toBeTruthy();
    expect(screen.getByText("2 loaded")).toBeTruthy();
  });

  it("renders session detail without a route id", async () => {
    routeState.params = {};
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");

    render(SessionDetailPage);

    expect(await screen.findByText("Missing session id")).toBeTruthy();
    expect(screen.getByText("No events for this session.")).toBeTruthy();
  });

  it("renders session detail loading errors", async () => {
    apiState.getSessionError = new Error("detail failed");
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");

    render(SessionDetailPage);

    expect(await screen.findByText("detail failed")).toBeTruthy();
    expect(screen.getByText("No events for this session.")).toBeTruthy();
  });

  it("renders unknown session detail loading errors", async () => {
    apiState.getSessionError = "not an Error" as unknown as Error;
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");

    render(SessionDetailPage);

    expect(await screen.findByText("Could not load session")).toBeTruthy();
  });
});
