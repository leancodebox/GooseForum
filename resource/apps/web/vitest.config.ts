import { defineConfig } from "vitest/config";
export default defineConfig({
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.{ts,tsx}", "src/**/messages/**", "src/test/**"],
      reporter: ["text-summary", "json-summary", "lcov"],
      reportsDirectory: "../../coverage/react-app",
      thresholds: {
        statements: 16,
        branches: 11,
        functions: 15,
        lines: 20,
      },
    },
  },
});
