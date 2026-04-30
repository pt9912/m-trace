import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";

export default defineConfig(({ mode }) => {
  const testAliases =
    mode === "test"
      ? [
          {
            find: "@npm9912/player-sdk",
            replacement: new URL("./tests/mocks/player-sdk.ts", import.meta.url).pathname
          }
        ]
      : [];

  return {
    plugins: [sveltekit()],
    build: {
      chunkSizeWarningLimit: 600
    },
    server: {
      port: 5173,
      strictPort: true,
      proxy: {
        "/api": "http://localhost:8080"
      }
    },
    resolve: {
      alias: testAliases,
      conditions: ["browser"]
    },
    test: {
      environment: "jsdom",
      setupFiles: ["tests/setup.ts"],
      include: ["tests/**/*.test.ts"],
      coverage: {
        provider: "v8",
        reportsDirectory: "coverage",
        reporter: ["text", "json-summary", "lcov"],
        include: ["src/**/*.{ts,svelte}"],
        exclude: ["src/app.html"],
        thresholds: {
          branches: 90,
          functions: 90,
          lines: 90,
          statements: 90
        }
      }
    }
  };
});
