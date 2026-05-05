<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { env } from "$env/dynamic/public";
  import { formatTime, listSessions, type StreamSession } from "$lib/api";
  import { startSseClient, type SseClient } from "$lib/sse-client";

  let sessions: StreamSession[] = [];
  let filter = "all";
  let loading = true;
  let error = "";
  let sseClient: SseClient | undefined;

  $: visibleSessions = filter === "all" ? sessions : sessions.filter((session) => session.state === filter);

  async function refresh() {
    loading = true;
    error = "";
    try {
      sessions = (await listSessions(250)).sessions;
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not load sessions";
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void refresh();
    // §5 H5: Live-Updates via SSE; bei Frame triggert ein listSessions-
    // Refresh, weil SSE-Frames nur Mindest-Payload liefern (REST-Read
    // ist die Source-of-Truth, Spec §10a). Der Client fällt automatisch
    // auf Polling zurück, wenn SSE persistent fehlschlägt (z. B. wenn
    // kein Token gesetzt ist und der Server 401 liefert) — das schließt
    // den ADR-0003-Vertrag.
    const apiBaseUrl = (env.PUBLIC_API_BASE_URL ?? "").replace(/\/$/, "");
    sseClient = startSseClient({
      url: `${apiBaseUrl}/api/stream-sessions/stream`,
      token: env.PUBLIC_API_TOKEN ?? "",
      onAppended: () => {
        void refresh();
      },
      // backend-api-contract.md §10a: bei Reconnect-Lücke > 1000 Events
      // sendet der Server `backfill_truncated`; der Konsument muss dann
      // den Snapshot neu laden, weil keine Live-`event_appended`-Frames
      // die Lücke schließen.
      onTruncated: () => {
        void refresh();
      },
      onPollingTick: () => refresh()
    });
  });

  onDestroy(() => {
    sseClient?.disconnect();
  });
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Sessions</h1>
    <p>Stream lifecycle state and event volume.</p>
  </div>
  <div class="toolbar">
    <select bind:value={filter} aria-label="State filter">
      <option value="all">All states</option>
      <option value="active">Active</option>
      <option value="stalled">Stalled</option>
      <option value="ended">Ended</option>
    </select>
    <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
  </div>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

<section class="panel">
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Session</th>
          <th>State</th>
          <th>Project</th>
          <th>Events</th>
          <th>Started</th>
          <th>Last event</th>
        </tr>
      </thead>
      <tbody>
        {#each visibleSessions as session (session.session_id)}
          <tr>
            <td><a href={`/sessions/${session.session_id}`}>{session.session_id}</a></td>
            <td>
              <span class={`pill ${session.state}`}>{session.state}</span>
              {#if session.end_source}
                <span
                  class="muted end-source"
                  title={session.end_source === "client"
                    ? "Ended by explicit session_ended event"
                    : "Ended by lifecycle sweeper after timeout"}
                >
                  via {session.end_source}
                </span>
              {/if}
            </td>
            <td>{session.project_id}</td>
            <td>{session.event_count}</td>
            <td>{formatTime(session.started_at)}</td>
            <td>{formatTime(session.last_event_at)}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="6" class="muted">No matching sessions.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>

<style>
  .end-source {
    margin-left: 6px;
    font-size: 0.85em;
  }
</style>
