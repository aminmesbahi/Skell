import { test, expect } from "@playwright/test";
import { injectWailsMock } from "./mock-wails";

const skills = JSON.stringify([
  { name: "alpha-skill", version: "1.0.0", registry: "default", source_repo: "", installed_path: "/p/alpha", installed_at: "2024-01-01T00:00:00Z", pinned: false, content_hash: "abc" },
  { name: "beta-skill",  version: "2.0.0", registry: "default", source_repo: "", installed_path: "/p/beta",  installed_at: "2024-01-02T00:00:00Z", pinned: true,  content_hash: "def" },
]);
const statuses = JSON.stringify([
  { name: "alpha-skill", installed: "1.0.0", latest: "1.0.0", status: "up-to-date" },
  { name: "beta-skill",  installed: "2.0.0", latest: "2.0.0", status: "pinned" },
]);

test.beforeEach(async ({ page }) => {
  await injectWailsMock(page);
  await page.addInitScript(
    ({ skills, statuses }) => {
      (window as Record<string, unknown>)["__wailsSetOverride"]?.(
        "RunSkell",
        (args: string[]) => {
          if (args[0] === "list" && args.includes("--global"))
            return Promise.resolve({ stdout: skills, stderr: "", success: true });
          if (args[0] === "list")
            return Promise.resolve({ stdout: skills, stderr: "", success: true });
          if (args[0] === "status")
            return Promise.resolve({ stdout: statuses, stderr: "", success: true });
          return Promise.resolve({ stdout: "[]", stderr: "", success: true });
        }
      );
    },
    { skills, statuses }
  );
});

test("Dashboard — renders stat cards", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: "Dashboard" })).toBeVisible();
  // At least one stat card is rendered
  await expect(page.locator(".card").first()).toBeVisible({ timeout: 5000 });
});

test("Dashboard — shows refresh button", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("button").first()).toBeVisible();
});

test("Installed Skills — renders skill rows", async ({ page }) => {
  await page.goto("/skills");
  await expect(page.getByText("alpha-skill")).toBeVisible({ timeout: 5000 });
  await expect(page.getByText("beta-skill")).toBeVisible();
});

test("Installed Skills — search filters rows", async ({ page }) => {
  await page.goto("/skills");
  await expect(page.getByText("alpha-skill")).toBeVisible({ timeout: 5000 });

  await page.getByPlaceholder(/search/i).fill("alpha");
  await expect(page.getByText("beta-skill")).toBeHidden();
  await expect(page.getByText("alpha-skill")).toBeVisible();
});

test("Skill Detail — renders without crash when metadata absent", async ({ page }) => {
  // Simulate a skill where metadata is missing from the registry entry
  const infoWithoutMetadata = JSON.stringify({
    name: "no-meta-skill",
    skill: { name: "no-meta-skill", description: "", license: "MIT", metadata: null },
    lock: { name: "no-meta-skill", version: "1.0.0", registry: "default", source_repo: "", installed_path: "/p", installed_at: "2024-01-01T00:00:00Z", pinned: false, content_hash: "x" },
    status: "up-to-date",
  });
  await page.addInitScript((info) => {
    (window as Record<string, unknown>)["__wailsSetOverride"]?.(
      "RunSkell",
      (args: string[]) => {
        if (args[0] === "info") return Promise.resolve({ stdout: info, stderr: "", success: true });
        return Promise.resolve({ stdout: "[]", stderr: "", success: true });
      }
    );
  }, infoWithoutMetadata);

  await page.goto("/skills/no-meta-skill");
  // The page must not show the error overlay (regression for the lifecycle crash)
  await expect(page.getByText("Unexpected Application Error!")).toBeHidden({ timeout: 5000 });
  await expect(page.getByText("no-meta-skill")).toBeVisible();
});
