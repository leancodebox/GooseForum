import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./apps/next/e2e",
  fullyParallel: true,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL: "http://127.0.0.1:3012",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile-chromium", use: { ...devices["Pixel 7"] } },
  ],
  webServer: {
    command: "pnpm next:dev",
    url: "http://127.0.0.1:3012",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
