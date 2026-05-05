import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { startSseClient } from "../src/lib/sse-client";
import { listSessions } from "../src/lib/api";

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
    network_signal_absent: Array<{ kind: string; adapter: string; reason: string }>;
    end_source: "client" | "sweeper" | null;
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
    delivery_status?: "accepted" | "duplicate_suspected" | "replayed";
  }>,
  nextCursor: undefined as string | undefined,
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
    event_count: 3,
    network_signal_absent: [] as Array<{ kind: string; adapter: string; reason: string }>,
    end_source: null as "client" | "sweeper" | null
  },
  {
    session_id: "session-2",
    project_id: "demo",
    state: "stalled",
    started_at: "2026-04-30T00:00:00.000Z",
    last_event_at: "2026-04-30T00:00:03.000Z",
    event_count: 2,
    network_signal_absent: [],
    end_source: null
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
    sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
  },
  {
    event_name: "rebuffer_started",
    project_id: "demo",
    session_id: "session-1",
    client_timestamp: "2026-04-30T00:00:00.000Z",
    server_received_at: "2026-04-30T00:00:01.000Z",
    ingest_sequence: 1,
    sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
  }
];

vi.mock("$lib/sse-client", () => ({
  // Tests rendern Sessions-Page in JSDOM ohne echten SSE-Server;
  // wir stuben den Client als `vi.fn()`-Spy, damit Tests an die
  // übergebenen Optionen (z. B. `onTruncated`) kommen, ohne einen
  // Reconnect-Loop oder den shared `sseConnection`-Store anzufassen.
  startSseClient: vi.fn(() => ({ disconnect: () => undefined }))
}));

vi.mock("$lib/api", () => ({
  formatTime: (value: string | undefined) => (value ? `time:${value}` : "n/a"),
  getHealth: vi.fn(async () => apiState.health),
  getSession: vi.fn(async (sessionId: string, _eventsLimit?: number, eventsCursor?: string) => {
    if (apiState.getSessionError) {
      throw apiState.getSessionError;
    }
    return {
      session: apiState.sessions.find((session) => session.session_id === sessionId) ?? apiState.sessions[0],
      // Production-faithful: jede Session liefert nur ihre eigenen
      // Events. Vor diesem Fix lieferte der Mock dasselbe Array für
      // jede sessionId, was die events/errors-Pages mit künstlichen
      // Cross-Session-Duplikaten konfrontierte.
      events: apiState.events.filter((event) => event.session_id === sessionId),
      // §5 H2: Pagination-Roundtrip — der Test setzt apiState.nextCursor
      // beim ersten Call und erwartet beim zweiten Call mit cursor !== ""
      // einen leeren Slice (drained).
      next_cursor: eventsCursor ? undefined : apiState.nextCursor
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

beforeEach(async () => {
  routeState.params = { id: "session-1" };
  apiState.sessions = sessions;
  apiState.events = events;
  apiState.nextCursor = undefined;
  apiState.listSessionsError = undefined;
  apiState.getSessionError = undefined;
  apiState.health = { ok: true, status: 200, text: "ok" };
  // §5 H5: Tests können den Store mutieren; Default für den nächsten
  // Test wieder zurücksetzen.
  const { sseConnection } = await import("../src/lib/status");
  sseConnection.set({ state: "not_yet_connected", changedAt: null, detail: null });
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

  // Spec backend-api-contract.md §10a verlangt, dass der Konsument
  // bei `backfill_truncated` den Snapshot neu lädt — sonst bleiben
  // Sessions stale, wenn die Reconnect-Lücke > sseBackfillLimit ist.
  // Vor diesem Fix übergab `+page.svelte` keinen `onTruncated`-Handler,
  // also ist der Vertragsbruch im Test nicht detektiert worden.
  it("triggers a sessions refresh on backfill_truncated", async () => {
    const sseSpy = vi.mocked(startSseClient);
    sseSpy.mockClear();

    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");
    render(SessionsPage);
    await screen.findByText("session-1");

    expect(sseSpy).toHaveBeenCalledTimes(1);
    const options = sseSpy.mock.calls[0]?.[0];
    expect(options?.onTruncated).toBeTypeOf("function");

    const listSpy = vi.mocked(listSessions);
    const callsBefore = listSpy.mock.calls.length;
    options?.onTruncated?.({ oldest_ingest_sequence: 100 });
    // refresh ist async; auf den nächsten Microtask warten.
    await Promise.resolve();
    await Promise.resolve();
    expect(listSpy.mock.calls.length).toBeGreaterThan(callsBefore);
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
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
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

    expect(screen.getByText("1 of 2 loaded")).toBeTruthy();
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
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
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
    // §5 H3 F-40: ohne PUBLIC_*_URL-Env-Variablen sind die Service-
    // Einträge `not configured`, nicht `inactive`.
    expect(screen.getAllByText("not configured").length).toBeGreaterThan(0);
  });

  it("renders the SSE panel with not_yet_connected default", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 204 })));
    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");

    render(StatusPage);

    // Header + Default-Pill für den SSE-Block.
    expect(await screen.findByRole("heading", { name: "SSE" })).toBeTruthy();
    expect(screen.getByText("not yet connected")).toBeTruthy();
    expect(
      screen.getByText("SSE-Live-Updates werden in Tranche 4 §5 H5 verdrahtet.")
    ).toBeTruthy();
  });

  it("renders SSE panel detail message when set", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 204 })));
    const { sseConnection } = await import("../src/lib/status");
    sseConnection.set({
      state: "polling_fallback",
      detail: "network error — falling back to polling",
      changedAt: "2026-05-04T12:00:00.000Z"
    });

    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");
    render(StatusPage);

    expect(await screen.findByText("polling fallback")).toBeTruthy();
    expect(screen.getByText("network error — falling back to polling")).toBeTruthy();
  });

  it("renders SSE last-change timestamp when no detail is set", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 204 })));
    const { sseConnection } = await import("../src/lib/status");
    sseConnection.set({
      state: "connected",
      detail: null,
      changedAt: "2026-05-04T12:34:56.000Z"
    });

    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");
    render(StatusPage);

    expect(await screen.findByText("connected")).toBeTruthy();
    expect(screen.getByText("Last change: 2026-05-04T12:34:56.000Z")).toBeTruthy();
  });

  it("renders open-link buttons for services with configured URLs", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 204 })));
    const { observabilityServices } = await import("../src/lib/status");
    observabilityServices.set([
      {
        name: "Grafana",
        envKey: "PUBLIC_GRAFANA_URL",
        configHint: "PUBLIC_GRAFANA_URL",
        openUrl: "https://grafana.test",
        probeUrl: "https://grafana.test/api/health",
        status: "connected"
      }
    ]);
    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");
    render(StatusPage);
    expect(await screen.findByRole("link", { name: "Open" })).toBeTruthy();
    // Connected-Service triggert die "connected"-Pill auf dem
    // Observability-Header.
    expect(screen.getAllByText("connected").length).toBeGreaterThan(0);
    // Reset so other tests start fresh.
    const { buildServiceLinks } = await import("../src/lib/status");
    observabilityServices.set(buildServiceLinks({}));
  });

  it("renders the last-read-error panel and clears it on demand", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => new Response(null, { status: 204 })));
    const { recordReadError, clearLastReadError } = await import("../src/lib/status");
    clearLastReadError();
    recordReadError("/api/stream-sessions/sess-1", new Error("boom"));

    const { default: StatusPage } = await import("../src/routes/status/+page.svelte");
    render(StatusPage);

    expect(await screen.findByText("/api/stream-sessions/sess-1")).toBeTruthy();
    expect(screen.getByText("boom")).toBeTruthy();
    const clearBtn = screen.getByRole("button", { name: /Clear/ });
    await fireEvent.click(clearBtn);
    expect(
      await screen.findByText("No session-read error since dashboard load.")
    ).toBeTruthy();
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

  // §5 H2 — neue Read-Shape-Felder

  it("shows end_source pill in session list when set", async () => {
    apiState.sessions = [
      {
        session_id: "session-end-1",
        project_id: "demo",
        state: "ended",
        started_at: "2026-04-30T00:00:00.000Z",
        last_event_at: "2026-04-30T00:00:05.000Z",
        event_count: 4,
        network_signal_absent: [],
        end_source: "client"
      },
      {
        session_id: "session-end-2",
        project_id: "demo",
        state: "ended",
        started_at: "2026-04-30T00:00:00.000Z",
        last_event_at: "2026-04-30T00:00:06.000Z",
        event_count: 5,
        network_signal_absent: [],
        end_source: "sweeper"
      }
    ];
    const { default: SessionsPage } = await import("../src/routes/sessions/+page.svelte");
    render(SessionsPage);
    expect(await screen.findByText("via client")).toBeTruthy();
    expect(screen.getByText("via sweeper")).toBeTruthy();
  });

  it("renders network_signal_absent section when entries are present", async () => {
    apiState.sessions = [
      {
        session_id: "session-1",
        project_id: "demo",
        state: "active",
        started_at: "2026-04-30T00:00:00.000Z",
        last_event_at: "2026-04-30T00:00:02.000Z",
        event_count: 1,
        network_signal_absent: [
          { kind: "segment", adapter: "native_hls", reason: "native_hls_unavailable" }
        ],
        end_source: null
      }
    ];
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    expect(await screen.findByRole("heading", { name: "Network signal absent" })).toBeTruthy();
    expect(screen.getByText("native_hls")).toBeTruthy();
    expect(screen.getByText("native_hls_unavailable")).toBeTruthy();
  });

  it("hides the network_signal_absent section when empty", async () => {
    apiState.sessions = sessions; // both have network_signal_absent: []
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    await screen.findByText("playback_error");
    expect(screen.queryByRole("heading", { name: "Network signal absent" })).toBeNull();
  });

  it("highlights non-accepted delivery_status events", async () => {
    apiState.sessions = sessions;
    apiState.events = [
      {
        event_name: "manifest_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:00.000Z",
        server_received_at: "2026-04-30T00:00:01.000Z",
        ingest_sequence: 1,
        sequence_number: 1,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" },
        delivery_status: "duplicate_suspected"
      },
      {
        event_name: "segment_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:02.000Z",
        server_received_at: "2026-04-30T00:00:03.000Z",
        ingest_sequence: 2,
        sequence_number: 2,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" },
        delivery_status: "replayed"
      }
    ];
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    expect(await screen.findByText("duplicate suspected")).toBeTruthy();
    expect(screen.getByText("replayed")).toBeTruthy();
  });

  it("shows a load-more button when next_cursor is present", async () => {
    apiState.sessions = sessions;
    apiState.nextCursor = "opaque-cursor";
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    expect(await screen.findByRole("button", { name: /Load more events/ })).toBeTruthy();
  });

  it("shows end_source on detail stats panel", async () => {
    apiState.sessions = [
      {
        session_id: "session-end",
        project_id: "demo",
        state: "ended",
        started_at: "2026-04-30T00:00:00.000Z",
        last_event_at: "2026-04-30T00:00:05.000Z",
        event_count: 4,
        network_signal_absent: [],
        end_source: "sweeper"
      }
    ];
    routeState.params.id = "session-end";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    expect(await screen.findByText("via sweeper")).toBeTruthy();
  });

  it("renders an error when load-more fails", async () => {
    apiState.sessions = sessions;
    apiState.events = [
      {
        event_name: "manifest_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:00.000Z",
        server_received_at: "2026-04-30T00:00:01.000Z",
        ingest_sequence: 1,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      }
    ];
    apiState.nextCursor = "opaque-cursor";
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    await screen.findByText("manifest_loaded");
    apiState.getSessionError = new Error("page 2 failed");
    const loadMoreBtn = screen.getByRole("button", { name: /Load more events/ });
    await fireEvent.click(loadMoreBtn);
    expect(await screen.findByText("page 2 failed")).toBeTruthy();
  });

  it("renders unknown load-more errors with fallback text", async () => {
    apiState.sessions = sessions;
    apiState.events = [
      {
        event_name: "manifest_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:00.000Z",
        server_received_at: "2026-04-30T00:00:01.000Z",
        ingest_sequence: 1,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      }
    ];
    apiState.nextCursor = "opaque-cursor";
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    await screen.findByText("manifest_loaded");
    apiState.getSessionError = "not an Error" as unknown as Error;
    const loadMoreBtn = screen.getByRole("button", { name: /Load more events/ });
    await fireEvent.click(loadMoreBtn);
    expect(await screen.findByText("Could not load more events")).toBeTruthy();
  });

  it("categorizes manifest, segment and lifecycle events with category tags", async () => {
    apiState.sessions = sessions;
    apiState.events = [
      {
        event_name: "manifest_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:00.000Z",
        server_received_at: "2026-04-30T00:00:01.000Z",
        ingest_sequence: 1,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      },
      {
        event_name: "segment_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:02.000Z",
        server_received_at: "2026-04-30T00:00:03.000Z",
        ingest_sequence: 2,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      },
      {
        event_name: "playback_started",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:04.000Z",
        server_received_at: "2026-04-30T00:00:05.000Z",
        ingest_sequence: 3,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      }
    ];
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    await screen.findByText("manifest_loaded");
    // Drei Kategorien sind sichtbar: manifest, segment, lifecycle.
    expect(screen.getAllByText("manifest").length).toBeGreaterThan(0);
    expect(screen.getAllByText("segment").length).toBeGreaterThan(0);
    expect(screen.getAllByText("lifecycle").length).toBeGreaterThan(0);
  });

  it("appends events on load-more click", async () => {
    apiState.sessions = sessions;
    apiState.events = [
      {
        event_name: "manifest_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:00.000Z",
        server_received_at: "2026-04-30T00:00:01.000Z",
        ingest_sequence: 1,
        sequence_number: 1,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      },
      {
        event_name: "segment_loaded",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:02.000Z",
        server_received_at: "2026-04-30T00:00:03.000Z",
        ingest_sequence: 2,
        sequence_number: 2,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      }
    ];
    apiState.nextCursor = "opaque-cursor";
    routeState.params.id = "session-1";
    const { default: SessionDetailPage } = await import("../src/routes/sessions/[id]/+page.svelte");
    render(SessionDetailPage);
    await screen.findByText("manifest_loaded");
    expect(screen.getByText("2 loaded")).toBeTruthy();
    // Beim Klick muss der Mock disjoint events liefern, sonst
    // dupliziert Svelte-each-Block die ingest_sequence-Keys.
    apiState.events = [
      {
        event_name: "rebuffer_started",
        project_id: "demo",
        session_id: "session-1",
        client_timestamp: "2026-04-30T00:00:04.000Z",
        server_received_at: "2026-04-30T00:00:05.000Z",
        ingest_sequence: 3,
        sequence_number: 3,
        sdk: { name: "@npm9912/player-sdk", version: "0.4.0" }
      }
    ];
    const loadMoreBtn = screen.getByRole("button", { name: /Load more events/ });
    await fireEvent.click(loadMoreBtn);
    await screen.findByText("3 loaded");
    expect(screen.getByText("rebuffer_started")).toBeTruthy();
  });
});
