/* c8 ignore start — bootstrap, ohne Test-Hook */

import { createAnalyzerServer } from "./server.js";

const PORT = parsePort(process.env.ANALYZER_LISTEN_PORT) ?? 7000;
const HOST = process.env.ANALYZER_LISTEN_HOST ?? "0.0.0.0";

const server = createAnalyzerServer();
server.listen(PORT, HOST, () => {
  console.log(`[analyzer-service] listening on http://${HOST}:${PORT}`);
});

const shutdown = (signal: string): void => {
  console.log(`[analyzer-service] received ${signal}, shutting down`);
  server.close((err) => {
    if (err) {
      console.error(`[analyzer-service] close failed: ${err.message}`);
      process.exit(1);
    }
    process.exit(0);
  });
};
process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("SIGINT", () => shutdown("SIGINT"));

function parsePort(value: string | undefined): number | null {
  if (value === undefined) return null;
  const n = Number(value);
  if (!Number.isFinite(n) || n <= 0 || n > 65_535) return null;
  return Math.floor(n);
}

/* c8 ignore end */
