import { test, expect } from "@playwright/test";
import { injectWails, installedSkill, statusEntry } from "./mock-wails";

/**
 * Home's "Needs attention" list: what to do across all projects, with links
 * into each project's skills page.
 */

const alpha = "/work/alpha";
const beta = "/work/beta";

test("lists only the projects that need something and links to their skills", async ({ page }) => {
  await injectWails(page, {
    repos: [alpha, beta],
    skell: {
      list: JSON.stringify([installedSkill("review-code")]),
      status: {
        stdout: JSON.stringify([statusEntry("review-code", "outdated"), statusEntry("write-docs", "locally-modified")]),
      },
    },
  });
  await page.goto("/");

  const attention = page.getByRole("region", { name: "Needs attention" });
  await expect(attention.getByText("1 outdated").first()).toBeVisible();
  await expect(attention.getByText("1 edited locally").first()).toBeVisible();

  await attention.getByRole("link", { name: /alpha/ }).click();
  await expect(page).toHaveURL(/\/projects\/project-[a-z0-9]+\/skills$/);
  await expect(page.getByRole("heading", { name: "Skills for alpha" })).toBeVisible();
});

test("says so when every project is up to date", async ({ page }) => {
  await injectWails(page, {
    repos: [alpha],
    skell: { list: JSON.stringify([installedSkill("review-code")]), status: JSON.stringify([statusEntry("review-code", "up-to-date")]) },
  });
  await page.goto("/");
  await expect(page.getByText("Your project is up to date.")).toBeVisible();
});

test("a project that cannot be checked is flagged, not hidden", async ({ page }) => {
  await injectWails(page, {
    repos: [alpha],
    skell: { list: { success: false, stderr: "no manifest found" }, status: { success: false, stderr: "no manifest found" } },
  });
  await page.goto("/");
  await expect(page.getByText("Could not check")).toBeVisible();
});
