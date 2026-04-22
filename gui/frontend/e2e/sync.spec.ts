import { test, expect } from "@playwright/test";
import { injectWailsMock } from "./mock-wails";

const inSyncReport = JSON.stringify({ installed: [], removed: [] });
const pendingReport = JSON.stringify({ installed: ["skill-a"], removed: ["old-skill"] });

test.beforeEach(async ({ page }) => {
  await injectWailsMock(page);
});

test("Sync — shows 'Not checked' badge before any run", async ({ page }) => {
  await page.goto("/sync");
  await expect(page.getByText("Not checked")).toBeVisible({ timeout: 5000 });
});

test("Sync — shows 'Already up to date' after preview when in sync", async ({ page }) => {
  await page.addInitScript((report) => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "sync") return Promise.resolve({ stdout: report, stderr: "", success: true });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  }, inSyncReport);

  await page.goto("/sync");
  await page.getByRole("button", { name: /preview sync/i }).click();
  await expect(page.getByText("Already up to date")).toBeVisible({ timeout: 5000 });
  await expect(page.getByText("Up to date")).toBeVisible();
});

test("Sync — shows pending changes and 'Apply now' button on dry-run", async ({ page }) => {
  await page.addInitScript((report) => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "sync") return Promise.resolve({ stdout: report, stderr: "", success: true });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  }, pendingReport);

  await page.goto("/sync");
  await page.getByRole("button", { name: /preview sync/i }).click();

  await expect(page.getByText("skill-a")).toBeVisible({ timeout: 5000 });
  await expect(page.getByText("old-skill")).toBeVisible();
  await expect(page.getByRole("button", { name: /apply now/i })).toBeVisible();
  await expect(page.getByText(/2 changes pending/i)).toBeVisible();
});

test("Sync — per-repo Preview button triggers individual sync", async ({ page }) => {
  await page.addInitScript((report) => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "sync") return Promise.resolve({ stdout: report, stderr: "", success: true });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  }, inSyncReport);

  await page.goto("/sync");
  await page.getByRole("button", { name: /^preview$/i }).first().click();
  await expect(page.getByText("Already up to date")).toBeVisible({ timeout: 5000 });
});

test("Sync — shows error badge on sync failure", async ({ page }) => {
  await page.addInitScript(() => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "sync")
          return Promise.resolve({ stdout: "", stderr: "connection refused", success: false });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  });

  await page.goto("/sync");
  await page.getByRole("button", { name: /preview sync/i }).click();
  await expect(page.getByText("Error")).toBeVisible({ timeout: 5000 });
  await expect(page.getByText(/connection refused/i)).toBeVisible();
});
