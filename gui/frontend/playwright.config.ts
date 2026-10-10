import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright E2E configuration.
 *
 * The specs drive the real UI in Chromium against a Vite dev server started in
 * `e2e` mode, which replaces the native Wails runtime with a mock (see
 * e2e/wails-runtime-mock.ts and e2e/mock-wails.ts). No desktop host or `skell`
 * binary is needed.
 *
 * Run:  bunx playwright test
 * UI:   bunx playwright test --ui
 *
 * PW_CHROMIUM_PATH points Playwright at an already-installed Chromium instead
 * of its own download (useful in containers).
 */
const chromiumPath = process.env.PW_CHROMIUM_PATH;

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? "github" : [["html", { outputFolder: "playwright-report" }]],
  use: {
    baseURL: "http://127.0.0.1:34115",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        ...(chromiumPath ? { launchOptions: { executablePath: chromiumPath, args: ["--no-sandbox"] } } : {}),
      },
    },
  ],
  webServer: {
    command: "bunx vite --mode e2e --port 34115 --strictPort",
    url: "http://127.0.0.1:34115",
    reuseExistingServer: !process.env.CI,
    timeout: 60_000,
  },
});
