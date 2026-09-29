// Run against a Vite dev server: node e2e/ux-smoke.mjs http://127.0.0.1:34117
// Native services are mocked; this test never installs or removes real skills.
import { chromium } from "playwright";
import { mkdir } from "node:fs/promises";
import assert from "node:assert/strict";
const browser = await chromium.launch({ channel: "chrome", headless: true });
try {
  const page = await browser.newPage({ viewport: { width: 1440, height: 1000 }, colorScheme: "dark" });
  const errors = [];
  page.on("pageerror", (error) => errors.push(String(error)));
  await page.route(/@wailsio_runtime\.js/, (route) => route.fulfill({ contentType: "text/javascript", body: `export const CancellablePromise = Promise; export const Call = { ByID: (id, ...args) => window.__uxCall(id, args), ByName: (name, ...args) => window.__uxCall(name, args) };` }));
  await page.addInitScript(() => {
    localStorage.setItem("skell-gui-repos", JSON.stringify({ version: 2, state: { repos: ["/work/alpha", "/work/beta"], selectedRepo: "/work/beta", recentProjects: [], sidebarCollapsed: false } }));
    const targets = [{ id: "claude", displayName: "Claude Code", dir: ".claude", detected: true }, { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: true }];
    const skill = { name: "review-code", description: "Review code changes for correctness and maintainability.", registry_alias: "team", registry_url: "https://example.com/team", registry_source: "global", metadata: { version: "2.0.0", lifecycle: "stable", tags: "review, quality", owner: "Engineering" } };
    window.__uxCall = async (id, args) => {
      if (id === 2394374327) {
        const cmd = args[0];
        let data = [];
        if (cmd[0] === "list") data = cmd.includes("registry") ? [skill] : [{ name: skill.name, version: "1.0.0", registry: "team", installed_path: ".cursor/skills/review-code", pinned: false }];
        if (cmd[0] === "status") data = [{ name: skill.name, installed: "1.0.0", latest: "2.0.0", status: "outdated" }];
        return { stdout: JSON.stringify(data), stderr: "", success: true };
      }
      if ([1734000213, 2920689880].includes(id)) return targets;
      if ([3848862957, 2004525469].includes(id)) return true;
      if (id === 2394438090) return "cursor";
      if (id === 567263264) return [{ name: "review-code", target: "cursor", errors: 0, warnings: 1, findings: [{ severity: "warning", category: "metadata", message: "Review compatibility with the selected agent." }] }];
      if (id === 3669316229) return { found: false, source_path: "", readme_content: "" };
      if (id === 3500197102) return "test";
      if (id === 3882022436) return [];
      if (id === 2693605881) return "/global";
      if (id === 3339277332) return [];
      if (id === 3629702033) return "";
      throw new Error("Unmocked native service: " + id);
    };
  });
  let hash = 0;
  for (const char of "/work/alpha") hash = ((hash << 5) - hash + char.charCodeAt(0)) | 0;
  const path = `/projects/project-${Math.abs(hash).toString(36)}/skills`;
  const base = process.argv[2] || "http://127.0.0.1:34117";
  await page.goto(base + path);
  await page.getByRole("link", { name: "review-code" }).waitFor();
  assert.match(await page.getByRole("link", { name: "review-code" }).getAttribute("href"), /repo=%2Fwork%2Falpha&target=cursor/);
  await mkdir("test-results/ux", { recursive: true });
  await page.screenshot({ path: "test-results/ux/skills-dark.png", fullPage: true });
  await page.getByRole("combobox", { name: "Theme" }).selectOption("light");
  assert.equal(await page.locator("html").getAttribute("data-theme"), "light");
  await page.screenshot({ path: "test-results/ux/skills-light.png", fullPage: true, animations: "disabled" });
  await page.keyboard.press("/");
  assert.equal(await page.getByRole("textbox", { name: "Search installed skills" }).evaluate((el) => el === document.activeElement), true);
  await page.getByRole("textbox", { name: "Search installed skills" }).fill("missing");
  await page.getByText("No skills match your filters.").waitFor();
  await page.getByRole("button", { name: "Clear filters" }).click();
  await page.getByRole("button", { name: "Upgrade all outdated (1)" }).click();
  await page.getByRole("dialog").waitFor();
  await page.keyboard.press("Escape");
  assert.equal(await page.getByRole("dialog").count(), 0);
  await page.goto(base + "/catalog");
  await page.getByRole("button", { name: "Preview" }).waitFor();
  await page.getByRole("checkbox", { name: "Claude Code" }).check();
  await page.screenshot({ path: "test-results/ux/catalog-light.png", fullPage: true });
  await page.getByRole("button", { name: "Preview" }).click();
  await page.getByRole("dialog").waitFor();
  await page.keyboard.press("Escape");
  assert.equal(await page.getByRole("dialog").count(), 0);
  await page.goto(base + "/");
  await page.getByRole("combobox", { name: "Theme" }).waitFor();
  await page.keyboard.press("Control+k");
  await page.getByRole("textbox", { name: "Search skills" }).waitFor();
  await page.waitForFunction(() => document.activeElement?.getAttribute("aria-label") === "Search skills");
  assert.deepEqual(errors, []);
  console.log("Browser smoke passed: route identity, dark/light themes, search shortcut, empty filters, dialogs, multi-agent selection.");
} finally { await browser.close(); }
