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
// MT_PROJECTS >= 2 aktiviert das Multi-Tenant-Szenario (R-26 b):
// Projekt lab-1 ist der Noisy Neighbor (MT_NOISY_EVENT_RATE ev/s,
// deutlich ueber der Limiter-Capacity), lab-2..lab-N sind Victims
// (MT_VICTIM_EVENT_RATE ev/s je Projekt, unter der Capacity).
// ACHTUNG Rundung: Raten werden auf ganze Batches AUFgerundet — die
// effektiv angebotene Rate ist ceil(rate/BATCH_SIZE)*BATCH_SIZE ev/s
// (Default 50 -> effektiv 60 bei Batch 20); die Victim-Rate muss auch
// NACH Rundung unter der Limiter-Capacity liegen, sonst meldet das
// 0x429-Gate einen Rundungs-Artefakt als Isolation-Failure. Nicht mit
// LOAD_PROFILE=open kombinierbar (harter Abbruch statt Silent-Win). Die
// Lab-Projekte muessen API-seitig via MTRACE_LAB_PROJECTS geseedet
// sein. Jedes Projekt sendet eine eigene synthetische Client-IP als
// X-Forwarded-For (10.99.0.<i>; das Lab setzt
// MTRACE_TRUST_FORWARDED_FOR=1) und KEINEN Origin-Header (curl-Pfad)
// — sonst wuerden die geteilte Quell-IP bzw. der geteilte Origin die
// client_ip-/origin-Buckets ueber Projekte hinweg konfundieren und
// der Test misst nicht die Per-Projekt-Isolation.
// Gates (k6-Thresholds, Schwellen: R-26):
//   - Victims sehen KEIN 429 (mtrace_mt_victim_rate_limited count<1)
//   - Victim-p95 < P95_BUDGET_MS (http_req_duration{role:victim})
//   - der Noisy WIRD gedrosselt (mtrace_mt_noisy_rate_limited count>0)
//
// Lauf (Feasibility; Core-Lab via `make dev-detached`):
//   docker run --rm --network host -v "$PWD/scripts/load:/scripts:ro" \
//     grafana/k6 run --vus 20 --duration 30s \
//     -e BASE_URL=http://localhost:8080 /scripts/playback-events.k6.js
// Die Readback-Reconciliation ist in scripts/smoke-load.sh automatisiert:
//   sie zählt die TATSÄCHLICH persistierten Events (Länge des
//   `events[]`-Arrays von GET /api/stream-sessions/{id}, paginiert; kommt
//   aus playback_events) je load-vu-*-Session und prüft persisted >=
//   accepted (kein stiller Verlust). Bewusst NICHT der Session-event_count
//   (wird vor dem Append getickt) und NICHT die paginierte Listen-API.

import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const PROJECT_TOKEN = __ENV.PROJECT_TOKEN || "demo-token";
const PROJECT_ID = __ENV.PROJECT_ID || "demo";
const ORIGIN = __ENV.ORIGIN || "http://localhost:5173";
const BATCH_SIZE = parseInt(__ENV.BATCH_SIZE || "20", 10);
const SESSION_PREFIX = __ENV.SESSION_PREFIX || "load-vu";

// LOAD_PROFILE=open: open-loop SLO-Szenario (constant-arrival-rate). Die
// Last (TARGET_EVENT_RATE Events/s) wird vorgegeben; gemessen wird, ob
// das System mitkommt (p95-Budget + dropped_iterations). Das entkoppelt
// die Last von der Maschinen-Geschwindigkeit -> stabile Nightly-Schwelle
// über Runner hinweg (Review-Empfehlung). Ohne LOAD_PROFILE=open bleibt
// das Skript closed-loop (--vus/--duration), wie für die
// Korrektheits-Gates.
const MT_PROJECTS = parseInt(__ENV.MT_PROJECTS || "0", 10);

// arrivalScenario baut die geteilte constant-arrival-rate-Form: die
// EVENT-Rate wird auf ganze Batches AUFgerundet — die effektiv
// angebotene Rate ist also `ceil(rate/BATCH_SIZE)*BATCH_SIZE` ev/s
// (z. B. 50 -> 60 bei Batch 20). Aufrufer, die gegen eine Capacity
// messen, muessen mit der EFFEKTIVEN Rate rechnen (s. offeredRate).
function offeredRate(eventRate) {
  return Math.max(1, Math.ceil(eventRate / BATCH_SIZE)) * BATCH_SIZE;
}
function arrivalScenario(eventRate, duration, prealloc, maxVUs, extra) {
  return Object.assign(
    {
      executor: "constant-arrival-rate",
      rate: Math.max(1, Math.ceil(eventRate / BATCH_SIZE)),
      timeUnit: "1s",
      duration: duration,
      preAllocatedVUs: prealloc,
      maxVUs: maxVUs,
    },
    extra || {},
  );
}

export const options = (function () {
  const o = { thresholds: {} };
  if (MT_PROJECTS >= 2) {
    if ((__ENV.LOAD_PROFILE || "closed") === "open") {
      // Die Achsen sind nicht kombinierbar: MT definiert eigene
      // Szenarien/Gates. Still zu gewinnen waere ein False-Green-Trap
      // (Operator erwartet das Open-Loop-SLO, gefahren wuerde MT).
      throw new Error(
        "MT_PROJECTS>=2 ist mit LOAD_PROFILE=open nicht kombinierbar (Multi-Tenant definiert eigene Szenarien und Gates)",
      );
    }
    const noisyRate = parseInt(__ENV.MT_NOISY_EVENT_RATE || "400", 10);
    const victimRate = parseInt(__ENV.MT_VICTIM_EVENT_RATE || "50", 10);
    const duration = __ENV.DURATION || "30s";
    o.scenarios = {
      noisy: arrivalScenario(noisyRate, duration, 10, 40, {
        exec: "mtProject",
        env: { MT_INDEX: "1", MT_ROLE: "noisy" },
        tags: { role: "noisy" },
      }),
    };
    for (let i = 2; i <= MT_PROJECTS; i++) {
      o.scenarios[`victim${i}`] = arrivalScenario(victimRate, duration, 5, 15, {
        exec: "mtProject",
        env: { MT_INDEX: String(i), MT_ROLE: "victim" },
        tags: { role: "victim" },
      });
    }
    // Noisy-Neighbor-Gates (Schwellen: R-26): Victims voll
    // isoliert (kein einziges 429, p95 im Budget), Noisy nachweislich
    // gedrosselt (sonst misst der Lauf gar keine Limiter-Kontention).
    // dropped_iterations: unterdimensionierte maxVUs wuerden die
    // angebotene Last sonst STILL unter das Ziel druecken und die
    // Fairness-Messung entwerten (Gates gruen ohne echte Kontention).
    o.thresholds["mtrace_mt_victim_rate_limited"] = ["count<1"];
    o.thresholds["mtrace_mt_noisy_rate_limited"] = ["count>0"];
    o.thresholds["http_req_duration{role:victim}"] = [
      `p(95)<${parseInt(__ENV.P95_BUDGET_MS || "1000", 10)}`,
    ];
    o.thresholds["dropped_iterations"] = ["rate<0.01"];
    return o;
  }
  if ((__ENV.LOAD_PROFILE || "closed") === "open") {
    const eventRate = parseInt(__ENV.TARGET_EVENT_RATE || "400", 10);
    o.scenarios = {
      slo: arrivalScenario(
        eventRate,
        __ENV.DURATION || "30s",
        parseInt(__ENV.OPEN_PREALLOC_VUS || "50", 10),
        parseInt(__ENV.OPEN_MAX_VUS || "100", 10),
      ),
    };
    o.thresholds["http_req_duration"] = [
      `p(95)<${parseInt(__ENV.P95_BUDGET_MS || "1000", 10)}`,
    ];
    // dropped_iterations: k6 konnte die vorgegebene Rate nicht halten
    // (System zu langsam / zu wenige VUs) -> SLO verfehlt.
    o.thresholds["dropped_iterations"] = ["rate<0.01"];
  }
  return o;
})();

// Custom-Counter fuer die Reconciliation-Bilanz (Batch-granular).
export const eventsSent = new Counter("mtrace_events_sent");
export const eventsAccepted = new Counter("mtrace_events_accepted");
export const eventsRateLimited = new Counter("mtrace_events_rate_limited");
export const eventsRejected = new Counter("mtrace_events_rejected");

// Multi-Tenant-Zaehler (Rollen-granular): tragen die Noisy-Neighbor-
// Gates (Thresholds oben). Die globalen Zaehler laufen zusaetzlich
// weiter, damit die Reconciliation-/Fehlerquoten-Auswertung des Smoke
// unveraendert funktioniert.
export const mtVictimSent = new Counter("mtrace_mt_victim_sent");
export const mtVictimAccepted = new Counter("mtrace_mt_victim_accepted");
export const mtVictimRateLimited = new Counter("mtrace_mt_victim_rate_limited");
export const mtNoisySent = new Counter("mtrace_mt_noisy_sent");
export const mtNoisyAccepted = new Counter("mtrace_mt_noisy_accepted");
export const mtNoisyRateLimited = new Counter("mtrace_mt_noisy_rate_limited");
const mtVictim = { sent: mtVictimSent, accepted: mtVictimAccepted, rateLimited: mtVictimRateLimited };
const mtNoisy = { sent: mtNoisySent, accepted: mtNoisyAccepted, rateLimited: mtNoisyRateLimited };

// Modul-Scope = pro VU isoliert in k6: fortlaufende Sequenz je VU.
// (Ein VU gehoert genau einem Szenario/Projekt -> Sequenz je Session.)
let seq = 0;

// postBatch kapselt Batch-Bau + POST + Zaehler — geteilt zwischen dem
// Single-Projekt-Default und den Multi-Tenant-Szenarien.
function postBatch(cfg) {
  const events = [];
  for (let i = 0; i < BATCH_SIZE; i++) {
    seq += 1;
    events.push({
      event_name: "rebuffer_started",
      project_id: cfg.projectId,
      session_id: cfg.sessionId,
      client_timestamp: new Date().toISOString(),
      sequence_number: seq,
      sdk: { name: "@pt9912/player-sdk", version: "loadtest" },
    });
  }
  const headers = {
    "Content-Type": "application/json",
    "X-MTrace-Token": cfg.token,
  };
  if (cfg.origin) {
    headers.Origin = cfg.origin;
  }
  if (cfg.xff) {
    headers["X-Forwarded-For"] = cfg.xff;
  }
  const payload = JSON.stringify({ schema_version: "1.0", events });
  const res = http.post(`${BASE_URL}/api/playback-events`, payload, {
    headers: headers,
  });

  eventsSent.add(BATCH_SIZE);
  if (cfg.counters) cfg.counters.sent.add(BATCH_SIZE);
  if (res.status === 202) {
    eventsAccepted.add(BATCH_SIZE);
    if (cfg.counters) cfg.counters.accepted.add(BATCH_SIZE);
  } else if (res.status === 429) {
    eventsRateLimited.add(BATCH_SIZE);
    if (cfg.counters) cfg.counters.rateLimited.add(BATCH_SIZE);
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

export default function () {
  postBatch({
    token: PROJECT_TOKEN,
    projectId: PROJECT_ID,
    sessionId: `${SESSION_PREFIX}-${__VU}`,
    origin: ORIGIN,
    counters: null,
  });
}

// mtProject ist die exec-Funktion aller Multi-Tenant-Szenarien; das
// Projekt kommt aus dem Szenario-env (MT_INDEX/MT_ROLE). Eigene
// synthetische Client-IP je Projekt (XFF), kein Origin-Header — s.
// Header-Kommentar (Konfundierungs-Vermeidung).
export function mtProject() {
  const idx = __ENV.MT_INDEX;
  postBatch({
    token: `lab-token-${idx}`,
    projectId: `lab-${idx}`,
    sessionId: `${SESSION_PREFIX}-p${idx}-${__VU}`,
    xff: `10.99.0.${idx}`,
    origin: null,
    counters: __ENV.MT_ROLE === "noisy" ? mtNoisy : mtVictim,
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
  let text =
    `\n  http_reqs: ${val("http_reqs", "count")} (${val("http_reqs", "rate").toFixed(1)}/s)\n` +
    `  events accepted: ${val("mtrace_events_accepted", "count")} (${val("mtrace_events_accepted", "rate").toFixed(1)}/s)\n` +
    `  events rate_limited: ${val("mtrace_events_rate_limited", "count")}\n` +
    `  events rejected: ${val("mtrace_events_rejected", "count")}\n` +
    `  http_req_duration: p90=${val("http_req_duration", "p(90)").toFixed(1)}ms ` +
    `p95=${val("http_req_duration", "p(95)").toFixed(1)}ms ` +
    `max=${val("http_req_duration", "max").toFixed(1)}ms\n`;
  if (MT_PROJECTS >= 2) {
    const noisyOffered = offeredRate(parseInt(__ENV.MT_NOISY_EVENT_RATE || "400", 10));
    const victimOffered = offeredRate(parseInt(__ENV.MT_VICTIM_EVENT_RATE || "50", 10));
    text +=
      `  MT offered (batch-aufgerundet): noisy=${noisyOffered} ev/s, victim=${victimOffered} ev/s je Projekt\n` +
      `  MT noisy (lab-1): sent=${val("mtrace_mt_noisy_sent", "count")} ` +
      `accepted=${val("mtrace_mt_noisy_accepted", "count")} ` +
      `rate_limited=${val("mtrace_mt_noisy_rate_limited", "count")}\n` +
      `  MT victims (lab-2..lab-${MT_PROJECTS}): sent=${val("mtrace_mt_victim_sent", "count")} ` +
      `accepted=${val("mtrace_mt_victim_accepted", "count")} ` +
      `rate_limited=${val("mtrace_mt_victim_rate_limited", "count")}\n` +
      `  victim p95: ${val("http_req_duration{role:victim}", "p(95)").toFixed(1)}ms\n`;
  }
  return {
    stdout: text,
    "/work/summary.json": JSON.stringify(data),
  };
}
