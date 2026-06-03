import { defineConfig } from "vitest/config";

// b — separater Vitest-Config-Pfad für die
// Bench-Suite, damit `pnpm test` (Unit-Tests) und `pnpm bench`
// (Benchmark-Smoke) sich nicht gegenseitig in die Quere kommen.
//
// `vitest bench --run --config vitest.bench.config.ts` läuft genau
// die Files unter `benchmarks/**/*.bench.ts` aus. Default-Includes
// von `vitest run` (`tests/**`) bleiben unverändert.

export default defineConfig({
  test: {
    include: ["benchmarks/**/*.bench.ts"],
    benchmark: {
      include: ["benchmarks/**/*.bench.ts"]
    }
  }
});
