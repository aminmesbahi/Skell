import { test, expect } from "@playwright/test";
import { injectWails } from "./mock-wails";

// The standalone Sync page was folded into the per-project screens; old links
// must still land somewhere useful.
test("the old /sync URL lands on Projects", async ({ page }) => {
  await injectWails(page, { repos: ["/work/alpha"] });
  await page.goto("/sync");
  await expect(page).toHaveURL(/\/projects$/);
  await expect(page.getByRole("heading", { name: "Projects", exact: true })).toBeVisible();
});
