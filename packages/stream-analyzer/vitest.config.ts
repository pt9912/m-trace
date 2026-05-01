import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    coverage: {
      provider: "v8",
      reportsDirectory: "coverage",
      reporter: ["text", "json-summary", "lcov"],
      include: ["src/**/*.ts"],
      // runtime.ts ist der reine Node-Stdlib-Adapter (dns.lookup +
      // globales fetch). Tests injizieren stattdessen einen Stub-
      // Runtime; die Default-Implementierung würde echte Netzwerk-
      // bzw. DNS-Calls erzwingen, was plan-0.3.0 §3 ausschließt
      // ("Der Parser arbeitet deterministisch ohne echte
      // Netzwerkabhängigkeit"). Sie bleibt außerhalb des
      // Coverage-Scopes.
      exclude: ["src/**/*.d.ts", "src/internal/loader/runtime.ts"],
      thresholds: {
        branches: 90,
        functions: 90,
        lines: 90,
        statements: 90
      }
    }
  }
});
