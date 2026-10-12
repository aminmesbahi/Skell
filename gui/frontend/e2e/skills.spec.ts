import { test, expect } from "@playwright/test";
import { injectWails, installedSkill, projectRoute, serviceCalls, statusEntry } from "./mock-wails";

const PROJECT = "/work/alpha";

const installed = [installedSkill("review-code"), installedSkill("write-docs"), installedSkill("pinned-skill", { pinned: true })];
const statuses = [statusEntry("review-code", "outdated"), statusEntry("write-docs", "up-to-date"), statusEntry("pinned-skill", "pinned")];

test.beforeEach(async ({ page }) => {
  await injectWails(page, {
    repos: [PROJECT],
    skell: { list: JSON.stringify(installed), status: JSON.stringify(statuses) },
  });
});

test("lists installed skills with their status", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/skills"));
  await expect(page.getByRole("link", { name: "review-code" })).toBeVisible();
  await expect(page.getByRole("link", { name: "write-docs" })).toBeVisible();
  await expect(page.getByRole("row", { name: /review-code/ }).getByText("Outdated")).toBeVisible();
  await expect(page.getByRole("row", { name: /pinned-skill/ }).getByText(/pinned/i).first()).toBeVisible();
});

test("search filters the table and Clear filters restores it", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/skills"));
  await expect(page.getByRole("link", { name: "write-docs" })).toBeVisible();

  await page.keyboard.press("/");
  await expect(page.getByRole("textbox", { name: "Search installed skills" })).toBeFocused();
  await page.getByRole("textbox", { name: "Search installed skills" }).fill("review");
  await expect(page.getByRole("link", { name: "write-docs" })).toBeHidden();

  await page.getByRole("textbox", { name: "Search installed skills" }).fill("nothing-matches");
  await expect(page.getByText("No skills match your filters.")).toBeVisible();
  await page.getByRole("button", { name: "Clear filters" }).click();
  await expect(page.getByRole("link", { name: "write-docs" })).toBeVisible();
});

test("the page header shows the breadcrumb once, without a duplicate back button", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/skills"));
  await expect(page.getByRole("navigation", { name: "Breadcrumb" })).toBeVisible();
  await expect(page.getByText("Back to Projects")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /add project/i })).toHaveCount(1);
});

test("upgrading everything outdated asks for confirmation and sends the right command", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/skills"));
  await page.getByRole("button", { name: "Upgrade all outdated (1)" }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);

  await page.getByRole("button", { name: "Upgrade all outdated (1)" }).click();
  await page.getByRole("dialog").getByRole("button", { name: "Upgrade", exact: true }).click();
  await expect(page.getByText(/review-code \(claude\): Upgrade finished/)).toBeVisible();
  const upgrades = (await serviceCalls(page)).filter((c) => c.method === "RunSkell" && (c.args[0] as string[])[0] === "upgrade");
  expect(upgrades).toHaveLength(1);
  expect(upgrades[0].args[0]).toContain("review-code");
});

test("a narrow window collapses the sidebar and keeps row actions reachable", async ({ page }) => {
  await page.setViewportSize({ width: 760, height: 900 });
  await page.goto(projectRoute(PROJECT, "/skills"));
  await expect(page.getByRole("link", { name: "review-code" })).toBeVisible();

  // Icon rail: labels are hidden and there is no collapse toggle to fight the media query.
  await expect(page.locator("aside").getByText("Catalog", { exact: true })).toHaveCount(0);
  await expect(page.getByRole("link", { name: "Catalog" })).toBeVisible(); // still reachable as an icon (title)
  await expect(page.getByTitle("Collapse sidebar")).toHaveCount(0);

  // The row actions stick to the right edge of the scrolling table, inside the viewport.
  const remove = page.getByRole("row", { name: /review-code/ }).getByRole("button", { name: "Remove" });
  await expect(remove).toBeVisible();
  const box = await remove.boundingBox();
  expect(box).not.toBeNull();
  expect(box!.x + box!.width).toBeLessThanOrEqual(760);
});
