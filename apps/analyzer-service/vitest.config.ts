import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      provider: "v8",
      reportsDirectory: "coverage",
      reporter: ["text", "json-summary", "lcov"],
      include: ["src/**/*.ts"],
      // main.ts ist der reine Bootstrap (Port-Lookup, listen, Signal-
      // Handler) und greift auf process.env / process.on zu. Tests
      // injizieren stattdessen den createServer-Factory direkt.
      exclude: ["src/main.ts"],
      thresholds: {
        branches: 90,
        functions: 90,
        lines: 90,
        statements: 90
      }
    }
  }
});
