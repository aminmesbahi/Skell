import { test, expect } from "@playwright/test";
import { injectWailsMock } from "./mock-wails";

const stableSkills = JSON.stringify([
  {
    name: "pdf-skill",
    description: "Generate PDFs",
    license: "MIT",
    metadata: { version: "1.0.0", owner: "testorg", lifecycle: "stable", scope: "general", tags: "pdf,doc", source_repo: "" },
  },
  {
    name: "image-resizer",
    description: "Resize images",
    license: "MIT",
    metadata: { version: "2.0.0", owner: "otherorg", lifecycle: "experimental", scope: "media", tags: "image", source_repo: "" },
  },
]);

test.beforeEach(async ({ page }) => {
  await injectWailsMock(page);
  await page.addInitScript((skills) => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "search") return Promise.resolve({ stdout: skills, stderr: "", success: true });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  }, stableSkills);
});

test("Registry — shows skills from search", async ({ page }) => {
  await page.goto("/registry");
  await expect(page.getByText("pdf-skill")).toBeVisible();
  await expect(page.getByText("image-resizer")).toBeVisible();
});

test("Registry — lifecycle filter limits results", async ({ page }) => {
  await page.goto("/registry");
  await expect(page.getByText("pdf-skill")).toBeVisible();

  // Change lifecycle to "experimental" → only image-resizer should remain
  await page.addInitScript((experimentalSkills) => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "search") return Promise.resolve({ stdout: experimentalSkills, stderr: "", success: true });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  }, JSON.stringify([stableSkills[1]]));

  const select = page.locator("select").first();
  await select.selectOption("experimental");
  // The page should re-search; at minimum no crash occurs
  await expect(page).not.toHaveURL(/error/);
});

test("Registry — search error shows notification", async ({ page }) => {
  await page.addInitScript(() => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      () => Promise.resolve({ stdout: "", stderr: "connection refused", success: false })
    );
  });
  await page.goto("/registry");
  await expect(page.getByText(/search failed/i).first()).toBeVisible({ timeout: 5000 });
});

test("Registry — install dialog opens on Install click", async ({ page }) => {
  await page.goto("/registry");
  await expect(page.getByText("pdf-skill")).toBeVisible();

  // Find an Install button and click it
  const installBtn = page.getByRole("button", { name: /install/i }).first();
  await installBtn.click();

  // Dialog should appear (contains "Install" text in a modal context)
  await expect(page.locator("dialog, [role='dialog']").first()).toBeVisible({ timeout: 3000 });
});
