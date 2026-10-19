import { test, expect } from "@playwright/test";
import { injectWails, installedSkill, registrySkill } from "./mock-wails";

/**
 * First-run guidance: the Getting started checklist follows what actually
 * exists (project, set-up, source, installed skill) and glossary hints explain
 * the vocabulary.
 */

const PROJECT = "/work/alpha";

test("a first-time user is walked through the first step, with a hint for each term", async ({ page }) => {
  await injectWails(page);
  await page.goto("/");

  const guide = page.getByRole("list", { name: "Getting started" });
  await expect(guide.getByRole("button", { name: "Add Project" })).toBeVisible();

  await guide.getByRole("button", { name: "What is set up?" }).click();
  await expect(page.getByRole("tooltip")).toContainText("skell.toml");
  await page.keyboard.press("Escape");
  await expect(page.getByRole("tooltip")).toHaveCount(0);
});

test("with a project but no source, the next step is to choose a source", async ({ page }) => {
  await injectWails(page, { repos: [PROJECT] });
  await page.goto("/");

  const guide = page.getByRole("region", { name: "Getting started" });
  await expect(guide.getByText("3 of 5 steps done")).toBeVisible();
  await guide.getByRole("link", { name: "Choose a source" }).click();
  await expect(page).toHaveURL(/\/catalog$/);
  await expect(page.getByLabel("Install into")).toHaveValue(PROJECT);
});

test("the guide can be hidden and brought back", async ({ page }) => {
  await injectWails(page, { repos: [PROJECT] });
  await page.goto("/");

  await page.getByRole("button", { name: "Hide the getting started guide" }).click();
  await expect(page.getByRole("region", { name: "Getting started" })).toHaveCount(0);

  // Persisted for next launch (injectWails reseeds storage on reload, so check it directly).
  expect(await page.evaluate(() => localStorage.getItem("skell-gui-repos"))).toContain('"gettingStartedDismissed":true');
  await page.getByRole("button", { name: "Show the setup guide" }).click();
  await expect(page.getByRole("region", { name: "Getting started" })).toBeVisible();
});

test("once a skill is installed the guide says you're all set", async ({ page }) => {
  await injectWails(page, {
    repos: [PROJECT],
    skell: {
      list: JSON.stringify([installedSkill("review-code")]),
      search: JSON.stringify([registrySkill("review-code", "Reviews code")]),
    },
  });
  await page.goto("/");
  await expect(page.getByText("You're all set.")).toBeVisible();
});

test.describe("product tour", () => {
  test("starts by itself for a new user on Home, and Skip tour dismisses it for good", async ({ page }) => {
    await injectWails(page, { repos: [PROJECT], tourSeen: false });
    await page.goto("/");

    const tour = page.getByRole("dialog", { name: "Product tour" });
    await expect(tour).toBeVisible();
    await expect(tour).toContainText("Find your way around");
    await tour.getByRole("button", { name: "Next" }).click();
    await expect(tour).toContainText("Add a project");

    await tour.getByRole("button", { name: "Skip tour" }).click();
    await expect(tour).toHaveCount(0);
    expect(await page.evaluate(() => localStorage.getItem("skell-gui-repos"))).toContain('"tourSeen":true');
  });

  test("walks every step with the keyboard and finishes", async ({ page }) => {
    await injectWails(page, { repos: [PROJECT], tourSeen: false });
    await page.goto("/");

    const tour = page.getByRole("dialog", { name: "Product tour" });
    await expect(tour).toBeVisible();
    const total = Number((await tour.getByText(/^Step 1 of \d+$/).textContent())!.match(/of (\d+)/)![1]);
    expect(total).toBeGreaterThanOrEqual(5);
    for (let i = 1; i < total; i += 1) await page.keyboard.press("ArrowRight");
    await expect(tour).toContainText(`Step ${total} of ${total}`);
    await tour.getByRole("button", { name: "Finish" }).click();
    await expect(tour).toHaveCount(0);
  });

  test("does not start once seen, and can be replayed from the Tour button", async ({ page }) => {
    await injectWails(page, { repos: [PROJECT] });
    await page.goto("/catalog");
    await expect(page.getByRole("dialog", { name: "Product tour" })).toHaveCount(0);

    await page.getByRole("button", { name: "Tour" }).click();
    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole("dialog", { name: "Product tour" })).toBeVisible();
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog", { name: "Product tour" })).toHaveCount(0);
  });

  test("a brand-new user with no project isn't interrupted", async ({ page }) => {
    await injectWails(page, { tourSeen: false });
    await page.goto("/");
    await expect(page.getByRole("button", { name: "Add Project" }).first()).toBeVisible();
    await expect(page.getByRole("dialog", { name: "Product tour" })).toHaveCount(0);
  });
});
