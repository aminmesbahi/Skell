import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright E2E configuration.
 *
 * Tests run against the Vite dev server (same port as `wails dev`).
 * The Wails runtime bridge (window.go) is injected as a mock via
 * the global setup / addInitScript in each spec.
 *
 * Run:  bunx playwright test
 * UI:   bunx playwright test --ui
 */
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? "github" : [["html", { outputFolder: "playwright-report" }]],
  use: {
    baseURL: "http://localhost:34115",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: "bun run dev",
    url: "http://localhost:34115",
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
  },
});
