import { test, expect } from "@playwright/test";
import { injectWailsMock } from "./mock-wails";

/**
 * Smoke tests — every route must render without a JS crash.
 */

const routes = [
  { path: "/", label: "Dashboard" },
  { path: "/repositories", label: "Repositories" },
  { path: "/skills", label: "Installed Skills" },
  { path: "/registry", label: "Registry" },
  { path: "/sync", label: "Sync" },
  { path: "/doctor", label: "Doctor" },
  { path: "/cache", label: "Cache" },
  { path: "/audit", label: "Audit Log" },
  { path: "/settings", label: "Settings" },
];

test.beforeEach(async ({ page }) => {
  await injectWailsMock(page);
});

for (const route of routes) {
  test(`${route.label} — renders without crashing`, async ({ page }) => {
    // Mock RunSkell to return sensible empty results for all commands
    await page.addInitScript(() => {
      (window as Record<string, unknown>)["__wailsSetOverride"]?.("RunSkell", (args: string[]) => {
        const cmd = args[0];
        if (cmd === "version") return Promise.resolve({ stdout: "v0.1.0", stderr: "", success: true });
        if (cmd === "list" || cmd === "status" || cmd === "doctor" || cmd === "search")
          return Promise.resolve({ stdout: "[]", stderr: "", success: true });
        if (cmd === "cache") return Promise.resolve({ stdout: "Cache: OK", stderr: "", success: true });
        if (cmd === "selfupdate") return Promise.resolve({ stdout: "up to date", stderr: "", success: true });
        return Promise.resolve({ stdout: "", stderr: "", success: true });
      });
    });

    await page.goto(route.path);

    // No fatal "Unexpected Application Error!" overlay
    await expect(page.getByText("Unexpected Application Error!")).toBeHidden();

    // Route-specific heading appears
    await expect(page.getByText(route.label, { exact: false })).toBeVisible({ timeout: 5000 });
  });
}

test("unknown route redirects to Dashboard", async ({ page }) => {
  await injectWailsMock(page);
  await page.goto("/this-does-not-exist");
  await expect(page.getByText("Dashboard")).toBeVisible();
});
