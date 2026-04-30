<script lang="ts">
  import { onMount } from "svelte";
  import { formatTime, getHealth, isErrorEvent, listSessions, type HealthStatus, type StreamSession } from "$lib/api";

  let sessions: StreamSession[] = [];
  let health: HealthStatus = { ok: false, status: 0, text: "not checked" };
  let loading = true;
  let error = "";

  $: activeCount = sessions.filter((session) => session.state === "active").length;
  $: stalledCount = sessions.filter((session) => session.state === "stalled").length;
  $: eventCount = sessions.reduce((sum, session) => sum + session.event_count, 0);

  async function refresh() {
    loading = true;
    error = "";
    try {
      const [sessionRes, healthRes] = await Promise.all([listSessions(), getHealth()]);
      sessions = sessionRes.sessions;
      health = healthRes;
    } catch (err) {
      error = err instanceof Error ? err.message : "Dashboard refresh failed";
    } finally {
      loading = false;
    }
  }

  onMount(refresh);
</script>

<section class="page-head">
  <div class="page-title">
    <h1>Live overview</h1>
    <p>Current ingest state from the local m-trace API.</p>
  </div>
  <button class="button" on:click={refresh} disabled={loading}>Refresh</button>
</section>

{#if error}
  <p class="error">{error}</p>
{/if}

<section class="stats">
  <div class="stat">
    <span>API</span>
    <strong>{health.ok ? "up" : "down"}</strong>
  </div>
  <div class="stat">
    <span>Active sessions</span>
    <strong>{activeCount}</strong>
  </div>
  <div class="stat">
    <span>Stalled sessions</span>
    <strong>{stalledCount}</strong>
  </div>
  <div class="stat">
    <span>Events</span>
    <strong>{eventCount}</strong>
  </div>
</section>

<section class="panel" style="margin-top: 18px;">
  <div class="panel-head">
    <h2>Recent sessions</h2>
    <a class="button secondary" href="/sessions">Open sessions</a>
  </div>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Session</th>
          <th>State</th>
          <th>Project</th>
          <th>Events</th>
          <th>Last event</th>
        </tr>
      </thead>
      <tbody>
        {#each sessions.slice(0, 8) as session}
          <tr>
            <td><a href={`/sessions/${session.session_id}`}>{session.session_id}</a></td>
            <td><span class={`pill ${session.state}`}>{session.state}</span></td>
            <td>{session.project_id}</td>
            <td>{session.event_count}</td>
            <td>{formatTime(session.last_event_at)}</td>
          </tr>
        {:else}
          <tr>
            <td colspan="5" class="muted">No sessions yet.</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</section>
