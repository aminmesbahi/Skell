import { test, expect } from "@playwright/test";
import { injectWails, registrySkill } from "./mock-wails";

const PROJECT = "/work/alpha";

const skills = [
  registrySkill("pdf-skill", "Generate PDFs"),
  registrySkill("image-resizer", "Resize images", { metadata: { version: "2.0.0", lifecycle: "experimental", tags: "image", owner: "Media" } }),
];

test.beforeEach(async ({ page }) => {
  await injectWails(page, {
    repos: [PROJECT],
    skell: { search: JSON.stringify(skills), list: "[]" },
  });
});

test("catalog lists skills from the configured sources", async ({ page }) => {
  await page.goto("/catalog");
  await expect(page.getByText("pdf-skill")).toBeVisible();
  await expect(page.getByText("image-resizer")).toBeVisible();
});

test("search narrows the catalog", async ({ page }) => {
  await page.goto("/catalog");
  await expect(page.getByText("pdf-skill")).toBeVisible();
  await page.getByRole("textbox", { name: "Search skills" }).fill("resize");
  await expect(page.getByText("pdf-skill")).toBeHidden();
  await expect(page.getByText("image-resizer")).toBeVisible();
});

test("the destination and agent pickers sit together under one label each", async ({ page }) => {
  await page.goto("/catalog");
  await expect(page.getByLabel("Install into")).toBeVisible();
  await expect(page.getByRole("group", { name: "For these agents" })).toBeVisible();
  await expect(page.getByRole("checkbox", { name: "Claude Code" })).toBeChecked();
});

test("Preview opens a dialog that Escape closes", async ({ page }) => {
  await page.goto("/catalog");
  await page.getByRole("button", { name: "Preview" }).first().click();
  await expect(page.getByRole("dialog")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog")).toHaveCount(0);
});

test("a catalog skill that is not installed yet explains its validation state", async ({ page }) => {
  await page.goto("/catalog");
  await expect(page.getByText("Not validated yet").first()).toBeVisible();
});
