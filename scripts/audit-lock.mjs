#!/usr/bin/env node
// Ersatz-Verifikation zu `make audit-ts`: fragt denselben npm-Bulk-
// Advisory-Endpoint ab, den `pnpm audit` benutzt, aber gegen die
// `packages:`-Sektion von pnpm-lock.yaml und mit eigener gzip-Behandlung.
//
// Hintergrund: Hinter einem transparenten Proxy kommt die Antwort von
// `/-/npm/v1/security/advisories/bulk` gzip-komprimiert an, ohne
// `Content-Encoding`-Header. pnpm bricht dann mit
// ERR_PNPM_AUDIT_BAD_RESPONSE ab — unabhaengig vom Lockfile-Inhalt, also
// ohne jede Aussage ueber die Dependency-Closure. Dieses Skript erkennt
// die gzip-Magic-Bytes `1f 8b` selbst und dekomprimiert.
//
// DIAGNOSE, KEIN GATE. Verbindlich bleibt `make audit-ts`
// (`pnpm audit --audit-level high`) im Security-CI-Job. Dieses Skript
// zaehlt Advisories pro Paket, pnpm zaehlt pro Dependency-Pfad — die
// Gesamtzahlen weichen deshalb ab. Deckungsgleich ist das
// Abbruchkriterium: die Anzahl critical+high.
//
// Usage:
//   node scripts/audit-lock.mjs [pfad/zu/pnpm-lock.yaml]
//
// Exit-Codes: 0 = keine critical/high, 1 = critical/high gefunden,
// 2 = Werkzeugfehler (Lockfile unlesbar, Endpoint nicht erreichbar).

import { readFileSync } from 'node:fs';
import { gunzipSync } from 'node:zlib';

const BULK_ENDPOINT =
  'https://registry.npmjs.org/-/npm/v1/security/advisories/bulk';
const RANK = { critical: 0, high: 1, moderate: 2, low: 3, info: 4 };

const lockPath = process.argv[2] ?? 'pnpm-lock.yaml';

let lock;
try {
  lock = readFileSync(lockPath, 'utf8');
} catch (err) {
  console.error(`[audit-lock] ${lockPath} nicht lesbar: ${err.message}`);
  process.exit(2);
}

// `packages:`-Sektion isolieren (bis zur naechsten Top-Level-Sektion).
// Nur sie fuehrt reine `name@version`-Keys; `snapshots:` haengt
// Peer-Suffixe an, die der Endpoint nicht als Version akzeptiert.
const section = /^packages:\n(.*?)(?=^\S)/ms.exec(lock);
if (!section) {
  console.error(`[audit-lock] ${lockPath}: packages:-Sektion nicht gefunden`);
  process.exit(2);
}

const versions = new Map();
for (const line of section[1].split('\n')) {
  const entry = /^ {2}((?:@[^/]+\/)?[^@\s]+)@([^:\s]+):\s*$/.exec(line);
  if (!entry) continue;
  const [, name, version] = entry;
  if (!versions.has(name)) versions.set(name, new Set());
  versions.get(name).add(version);
}

const payload = Object.fromEntries(
  [...versions]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([name, set]) => [name, [...set].sort()]),
);
const packageCount = Object.keys(payload).length;
const versionCount = Object.values(payload).reduce((n, v) => n + v.length, 0);
console.log(
  `[audit-lock] ${lockPath}: ${packageCount} Pakete, ${versionCount} Versionen`,
);

let body;
try {
  const response = await fetch(BULK_ENDPOINT, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    console.error(
      `[audit-lock] Bulk-Endpoint antwortete ${response.status} ${response.statusText}`,
    );
    process.exit(2);
  }
  body = Buffer.from(await response.arrayBuffer());
} catch (err) {
  console.error(`[audit-lock] Bulk-Endpoint nicht erreichbar: ${err.message}`);
  process.exit(2);
}

// Genau der Punkt, an dem pnpm scheitert: gzip ohne Content-Encoding.
if (body[0] === 0x1f && body[1] === 0x8b) {
  console.log('[audit-lock] Antwort war gzip ohne Content-Encoding-Header');
  body = gunzipSync(body);
}

let advisories;
try {
  advisories = JSON.parse(body.toString('utf8'));
} catch (err) {
  console.error(`[audit-lock] Antwort ist kein JSON: ${err.message}`);
  process.exit(2);
}

const rows = [];
const counts = new Map();
for (const [name, entries] of Object.entries(advisories)) {
  for (const advisory of entries) {
    const severity = advisory.severity ?? 'info';
    counts.set(severity, (counts.get(severity) ?? 0) + 1);
    rows.push({
      severity,
      name,
      range: advisory.vulnerable_versions,
      title: advisory.title,
      url: advisory.url,
    });
  }
}

rows.sort(
  (a, b) =>
    (RANK[a.severity] ?? 9) - (RANK[b.severity] ?? 9) ||
    a.name.localeCompare(b.name),
);
for (const row of rows) {
  console.log(
    `${row.severity.padEnd(9)} ${row.name.padEnd(24)} ${String(row.range).padEnd(18)} ${row.title}  ${row.url}`,
  );
}

const summary = [...counts]
  .sort(([a], [b]) => (RANK[a] ?? 9) - (RANK[b] ?? 9))
  .map(([severity, n]) => `${n} ${severity}`)
  .join(', ');
console.log(`\n[audit-lock] Summe: ${summary || 'keine Advisories'}`);

const blocking = (counts.get('critical') ?? 0) + (counts.get('high') ?? 0);
console.log(
  `[audit-lock] audit-level high => ${blocking ? 'FAIL' : 'PASS'} (${blocking} critical/high)`,
);
process.exit(blocking ? 1 : 0);
