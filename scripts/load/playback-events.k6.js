// k6-Feasibility-Lastskript (erste Stufe des Load-Smoke).
//
// POSTet Playback-Event-Batches gegen /api/playback-events über den
// Legacy-Project-Token-Pfad (X-MTrace-Token). Jeder k6-VU ist eine
// eigene Session mit fortlaufender sequence_number — Grundlage fuer die
// Readback-Reconciliation (persistierte vs. gesendete Events). Session-
// Token-Issuance + Kapazitaets-Modus (angehobenes Rate-Limit) sind
// spätere Schritte; dieses Skript belegt nur die Machbarkeit + Baseline.
//
// ENV (k6 -e KEY=VALUE):
//   BASE_URL (default http://localhost:8080)
//   PROJECT_TOKEN (default demo-token)
//   PROJECT_ID (default demo)
//   ORIGIN (default http://localhost:5173)
//   BATCH_SIZE (default 20)
//   SESSION_PREFIX (default load-vu)
// VUs/Dauer kommen ueber k6-CLI-Flags (--vus / --duration).
//
// Lauf (Feasibility; Core-Lab via `make dev-detached`):
//   docker run --rm --network host -v "$PWD/scripts/load:/scripts:ro" \
//     grafana/k6 run --vus 20 --duration 30s \
//     -e BASE_URL=http://localhost:8080 /scripts/playback-events.k6.js
// Readback-Reconciliation danach: GET /api/stream-sessions?limit=50
//   (Header X-MTrace-Token) und Summe der event_count der load-vu-*-
//   Sessions gegen mtrace_events_accepted abgleichen (muss deckungs-
//   gleich sein; Counts sind Vielfache von BATCH_SIZE -> Ganz-Batch-
//   Persist, kein stiller Verlust). Automatisierung folgt als nächster
//   Schritt.

import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const PROJECT_TOKEN = __ENV.PROJECT_TOKEN || "demo-token";
const PROJECT_ID = __ENV.PROJECT_ID || "demo";
const ORIGIN = __ENV.ORIGIN || "http://localhost:5173";
const BATCH_SIZE = parseInt(__ENV.BATCH_SIZE || "20", 10);
const SESSION_PREFIX = __ENV.SESSION_PREFIX || "load-vu";

// Custom-Counter fuer die Reconciliation-Bilanz (Batch-granular).
export const eventsSent = new Counter("mtrace_events_sent");
export const eventsAccepted = new Counter("mtrace_events_accepted");
export const eventsRateLimited = new Counter("mtrace_events_rate_limited");
export const eventsRejected = new Counter("mtrace_events_rejected");

// Modul-Scope = pro VU isoliert in k6: fortlaufende Sequenz je VU.
let seq = 0;

export default function () {
  const sessionId = `${SESSION_PREFIX}-${__VU}`;
  const events = [];
  for (let i = 0; i < BATCH_SIZE; i++) {
    seq += 1;
    events.push({
      event_name: "rebuffer_started",
      project_id: PROJECT_ID,
      session_id: sessionId,
      client_timestamp: new Date().toISOString(),
      sequence_number: seq,
      sdk: { name: "@pt9912/player-sdk", version: "loadtest" },
    });
  }
  const payload = JSON.stringify({ schema_version: "1.0", events });
  const res = http.post(`${BASE_URL}/api/playback-events`, payload, {
    headers: {
      "Content-Type": "application/json",
      "X-MTrace-Token": PROJECT_TOKEN,
      Origin: ORIGIN,
    },
  });

  eventsSent.add(BATCH_SIZE);
  if (res.status === 202) {
    eventsAccepted.add(BATCH_SIZE);
  } else if (res.status === 429) {
    eventsRateLimited.add(BATCH_SIZE);
  } else {
    eventsRejected.add(BATCH_SIZE);
  }

  // Contract: unter dem Default-Limit ist 202 erwartet; ueber dem Limit
  // ist 429 der korrekte (nicht-stille) Pfad. Beides ist "ok"; alles
  // andere (5xx, 4xx != 429) ist ein echter Fehler.
  check(res, {
    "status 202 oder 429 (kein stiller Fehler)": (r) =>
      r.status === 202 || r.status === 429,
  });
}

// handleSummary schreibt die volle Metrik-Struktur als JSON fuer die
// Auswertung im Smoke (zuverlaessiger als das deprecatete
// --summary-export) plus eine knappe stdout-Zusammenfassung ohne externe
// jslib-Abhaengigkeit.
export function handleSummary(data) {
  const m = data.metrics || {};
  const val = (name, key) =>
    (m[name] && m[name].values && m[name].values[key]) || 0;
  const text =
    `\n  http_reqs: ${val("http_reqs", "count")} (${val("http_reqs", "rate").toFixed(1)}/s)\n` +
    `  events accepted: ${val("mtrace_events_accepted", "count")} (${val("mtrace_events_accepted", "rate").toFixed(1)}/s)\n` +
    `  events rate_limited: ${val("mtrace_events_rate_limited", "count")}\n` +
    `  events rejected: ${val("mtrace_events_rejected", "count")}\n` +
    `  http_req_duration: p90=${val("http_req_duration", "p(90)").toFixed(1)}ms ` +
    `p95=${val("http_req_duration", "p(95)").toFixed(1)}ms ` +
    `max=${val("http_req_duration", "max").toFixed(1)}ms\n`;
  return {
    stdout: text,
    "/work/summary.json": JSON.stringify(data),
  };
}
