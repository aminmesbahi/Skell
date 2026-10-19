import { test, expect } from "@playwright/test";
import { injectWails, installedSkill, projectRoute, statusEntry } from "./mock-wails";

const PROJECT = "/work/alpha";

test.beforeEach(async ({ page }) => {
  await injectWails(page, {
    repos: [PROJECT],
    skell: {
      list: JSON.stringify([installedSkill("review-code"), installedSkill("write-docs")]),
      status: JSON.stringify([statusEntry("review-code", "outdated"), statusEntry("write-docs", "up-to-date")]),
      doctor: JSON.stringify([{ severity: "warning", code: "stale-mirror", message: "mirror copy cursor/review-code is missing or out of date" }]),
    },
  });
});

test("a project's root lands on its Skills tab, with the old overview stats as a summary line", async ({ page }) => {
  await page.goto(projectRoute(PROJECT));
  await expect(page).toHaveURL(/\/projects\/project-[a-z0-9]+\/skills$/);
  await expect(page.getByRole("heading", { name: "Skills for alpha" })).toBeVisible();
  await expect(page.getByText("2 skills · 1 outdated")).toBeVisible();
});

test("tabs switch between Skills, Agents, Health and Activity", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/skills"));
  const tabs = page.getByRole("navigation", { name: "Project sections" });
  await expect(tabs.getByRole("link", { name: "Skills" })).toHaveAttribute("aria-current", "page");

  await tabs.getByRole("link", { name: "Health" }).click();
  await expect(page).toHaveURL(/\/health$/);
  await expect(page.getByRole("heading", { name: "Health for alpha" })).toBeVisible();
  await expect(tabs.getByRole("link", { name: "Health" })).toHaveAttribute("aria-current", "page");

  await tabs.getByRole("link", { name: "Agents" }).click();
  await expect(page.getByRole("heading", { name: "Agents for alpha" })).toBeVisible();
  await tabs.getByRole("link", { name: "Activity" }).click();
  await expect(page.getByRole("heading", { name: "Activity for alpha" })).toBeVisible();
});

test("Health speaks plainly and tucks the technical detail away", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/health"));
  await expect(page.getByText("Another agent's copy is out of date")).toBeVisible();
  await expect(page.getByText("stale-mirror: mirror copy")).toBeHidden();
  await page.getByText("Technical details").click();
  await expect(page.getByText("stale-mirror: mirror copy")).toBeVisible();
});
