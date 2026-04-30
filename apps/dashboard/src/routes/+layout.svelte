<script lang="ts">
  import "../app.css";
</script>

<div class="shell">
  <aside class="sidebar">
    <div class="brand">
      <strong>m-trace</strong>
      <span>Media telemetry lab</span>
    </div>
    <nav class="nav" aria-label="Main navigation">
      <a href="/">Overview</a>
      <a href="/sessions">Sessions</a>
      <a href="/events">Events</a>
      <a href="/errors">Errors</a>
      <a href="/status">Status</a>
      <a href="/demo">Demo player</a>
    </nav>
  </aside>
  <main class="content">
    <slot />
  </main>
</div>
