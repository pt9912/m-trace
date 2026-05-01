#!/usr/bin/env node
/* c8 ignore start — bootstrap, ohne Test-Hook */

import { runCli } from "./check.js";

runCli({
  argv: process.argv.slice(2),
  stdout: process.stdout,
  stderr: process.stderr
})
  .then((code) => {
    process.exitCode = code;
  })
  .catch((error) => {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`m-trace: unerwarteter Fehler: ${message}\n`);
    process.exitCode = 1;
  });

/* c8 ignore end */
