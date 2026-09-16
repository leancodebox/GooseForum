import { defineConfig } from "vitest/config";
import { fileURLToPath } from "node:url";

export default defineConfig({
  root: fileURLToPath(new URL(".", import.meta.url)),
  test: {
    environment: "jsdom",
    include: [
      "apps/web/src/**/*.test.{ts,tsx}",
      "packages/theme-default/test/**/*.test.{ts,tsx}",
    ],
    setupFiles: [
      "./packages/theme-default/test/setup.ts",
      "./apps/web/src/test/setup.ts",
    ],
    coverage: {
      provider: "v8",
      include: [
        "apps/web/src/**/*.{ts,tsx}",
        "packages/{theme-default,runtime,ui}/src/**/*.{ts,tsx}",
      ],
      exclude: [
        "apps/web/src/**/*.test.{ts,tsx}",
        "apps/web/src/**/messages/**",
        "apps/web/src/test/**",
      ],
      reporter: ["text-summary", "json-summary", "lcov"],
      reportsDirectory: "./coverage/react",
      thresholds: {
        statements: 41,
        branches: 35,
        functions: 41,
        lines: 46,
      },
    },
  },
});
