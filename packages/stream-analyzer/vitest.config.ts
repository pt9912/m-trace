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
        // Branches strikt auf 91. Die Edge-Case-Tests aus dem
        // 0.22.3-Coverage-Härtungs-Patch heben den Median-Wert von
        // ~90.0% auf ~91.1% (3 lokale Läufe: 90.82/91.09/91.26%
        // — Vitest v8-Coverage hat eine Run-to-Run-Variance von
        // ~0.4pp). Bei diesem strikten Threshold können sporadische
        // CI-Runs unter den Grenzwert rutschen; bis weitere
        // Härtungs-Tests die Marge auf ~92%+ schieben, ist
        // `gh run rerun <run-id>` der dokumentierte Re-Run-Pfad
        // (analog zum webrtc-drift.yml-Flake-Handling).
        // Folge-Item: Coverage-Stabilisierung auf 92.5% Median oder
        // Wechsel auf deterministischen istanbul-Provider.
        branches: 91,
        functions: 90,
        lines: 90,
        statements: 90
      }
    }
  }
});
