import { test, expect } from "@playwright/test";
import { injectWails, projectRoute } from "./mock-wails";

/**
 * Smoke tests: every route renders without a crash, legacy URLs redirect, and
 * the theme switch works.
 */

const PROJECT = "/work/alpha";

const routes = [
  { path: "/", heading: "Home" },
  { path: "/projects", heading: "Projects" },
  { path: "/catalog", heading: "Catalog" },
  { path: "/activity", heading: "Activity" },
  { path: "/settings", heading: "Settings" },
  { path: projectRoute(PROJECT), heading: "alpha" },
  { path: projectRoute(PROJECT, "/skills"), heading: "Skills for alpha" },
  { path: projectRoute(PROJECT, "/agents"), heading: "Agents for alpha" },
  { path: projectRoute(PROJECT, "/health"), heading: "Health for alpha" },
  { path: projectRoute(PROJECT, "/activity"), heading: "Activity for alpha" },
];

test.beforeEach(async ({ page }) => {
  await injectWails(page, { repos: [PROJECT] });
});

for (const route of routes) {
  test(`${route.path} renders without crashing`, async ({ page }) => {
    const errors: string[] = [];
    page.on("pageerror", (error) => errors.push(String(error)));
    await page.goto(route.path);
    await expect(page.getByRole("heading", { name: route.heading, exact: true }).first()).toBeVisible();
    await expect(page.getByText("Unexpected Application Error!")).toBeHidden();
    expect(errors).toEqual([]);
  });
}

test("unknown routes fall back to Home", async ({ page }) => {
  await page.goto("/definitely-not-a-page");
  await expect(page.getByRole("heading", { name: "Home", exact: true })).toBeVisible();
});

for (const [legacy, destination] of [["/registry", "/catalog"], ["/audit", "/activity"]] as const) {
  test(`legacy ${legacy} redirects to ${destination}`, async ({ page }) => {
    await page.goto(legacy);
    await expect(page).toHaveURL(new RegExp(`${destination}$`));
  });
}

test("theme switch updates the document and survives navigation", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("combobox", { name: "Theme" }).selectOption("light");
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
  await page.getByRole("link", { name: "Settings" }).click();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
});
