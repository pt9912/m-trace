// plan-0.9.5 §5 Tranche 4 (RAK-Wave-2 / extra-gates.md §3.6) —
// Stryker-Konfiguration für das TS-Pilot-Modul
// `src/adapters/webrtc/sampling.ts`.
//
// Scope: nur das Sampling-File mutieren — die übrigen Adapter
// (HLS-Adapter, Tracker, Redact) sind nicht Teil des Tranche-4-
// Pilot-Scopes und würden den Lauf unnötig verlängern. Test-Runner
// ist Vitest 4.1 (selbe Version wie `make ts-test`).
//
// Initial nicht-blockierend (Plan-DoD §5). PR-Blockierung erst, wenn
// der Score drei Beobachtungsläufe in Folge > 70 % erreicht; siehe
// `docs/dev/mutation-testing.md` §3 Score-Schwelle.
//
// Reports:
//   - `html`: visueller Report unter `reports/mutation/mutation.html`,
//     wird vom Nightly-Workflow als Artefakt hochgeladen.
//   - `json`: Maschinen-lesbarer Report unter
//     `reports/mutation/mutation.json` für etwaige Downstream-
//     Konsumenten (Trend-Tracking, Folge-Backlog).
//   - `clear-text`: stdout-Summary, damit der Workflow-Run-Log den
//     Score zeigt ohne den HTML-Report zu öffnen.

/** @type {import('@stryker-mutator/api/core').PartialStrykerOptions} */
module.exports = {
  packageManager: "pnpm",
  testRunner: "vitest",
  // Plugins explizit deklarieren statt über den Default-Glob
  // `@stryker-mutator/*`: unter pnpms isoliertem node_modules liegen
  // `core` und `vitest-runner` in getrennten `.pnpm`-Verzeichnissen,
  // sodass die Glob-Discovery den Runner nicht als Sibling findet. Der
  // explizite Name lässt Stryker `require.resolve` nutzen, das pnpm
  // auflöst.
  plugins: ["@stryker-mutator/vitest-runner"],
  reporters: ["html", "json", "clear-text", "progress"],
  htmlReporter: { fileName: "reports/mutation/mutation.html" },
  jsonReporter: { fileName: "reports/mutation/mutation.json" },
  mutate: ["src/adapters/webrtc/sampling.ts"],
  // Stryker scant standardmäßig alle Test-Files; wir lassen das so,
  // damit auch Tests aus `tests/tracker.test.ts` und `tests/webrtc-
  // adapter.test.ts` zur Mutant-Killing-Quote beitragen (beide
  // exercieren `sampling.ts`-Pfade indirekt).
  thresholds: {
    // Plan-DoD §5-4: > 70 % Wunsch-Ziel; PR-Blockierung erst nach
    // Beobachtungsphase. `break: 0` → Stryker exitet immer mit
    // 0 (Workflow-Failure würde an `break`-Schwelle scheitern).
    high: 80,
    low: 60,
    break: 0,
  },
  timeoutMS: 30_000,
  concurrency: 2,
  cleanTempDir: "always",
  disableTypeChecks: "{src,tests}/**/*.{ts,tsx}",
};
