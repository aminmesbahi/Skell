/**
 * Helpers for the e2e specs. The app runs in a plain browser against the Vite
 * dev server started with `--mode e2e`, where `@wailsio/runtime` is replaced by
 * ./wails-runtime-mock.ts. `injectWails()` tells that mock how to answer.
 */
import type { Page } from "@playwright/test";

export interface SkellAnswer {
  stdout?: string;
  stderr?: string;
  success?: boolean;
}

export interface WailsConfig {
  /** Answers for `skell ...` calls, keyed by "cmd", "cmd sub" or the full argument list. */
  skell?: Record<string, SkellAnswer | string>;
  /** Return values for other service methods (e.g. { IsRepoInitialized: false }). */
  values?: Record<string, unknown>;
  /** Projects to pre-register; the first is selected unless `selected` is given. */
  repos?: string[];
  selected?: string;
  colorScheme?: "light" | "dark";
  /** The product tour starts by itself for users who haven't seen it; specs opt in with false. */
  tourSeen?: boolean;
}

/** Same hash as src/lib/navigation.ts createProjectId. */
export function createProjectId(path: string): string {
  const normalized = path.replace(/\\/g, "/").toLowerCase();
  let hash = 0;
  for (let i = 0; i < normalized.length; i += 1) {
    hash = (hash << 5) - hash + normalized.charCodeAt(i);
    hash |= 0;
  }
  return `project-${Math.abs(hash).toString(36)}`;
}

export const projectRoute = (path: string, sub = "") => `/projects/${createProjectId(path)}${sub}`;

export async function injectWails(page: Page, config: WailsConfig = {}): Promise<void> {
  const repos = config.repos ?? [];
  await page.addInitScript(
    ({ skell, values, repos, selected, tourSeen }) => {
      (window as unknown as { __e2e: unknown }).__e2e = { skell, values, calls: [] };
      try {
        localStorage.setItem(
          "skell-gui-repos",
          JSON.stringify({ version: 2, state: { repos, selectedRepo: selected, recentProjects: [], sidebarCollapsed: false, tourSeen } })
        );
      } catch {
        /* storage unavailable */
      }
    },
    { skell: config.skell ?? {}, values: config.values ?? {}, repos, selected: config.selected ?? repos[0] ?? "global", tourSeen: config.tourSeen ?? true }
  );
  if (config.colorScheme) await page.emulateMedia({ colorScheme: config.colorScheme });
}

/** The service calls the page has made so far (method name + arguments). */
export function serviceCalls(page: Page): Promise<{ method: string; args: unknown[] }[]> {
  return page.evaluate(() => (window as unknown as { __e2e: { calls: { method: string; args: unknown[] }[] } }).__e2e.calls);
}

export const installedSkill = (name: string, extra: Record<string, unknown> = {}) => ({
  name,
  version: "",
  registry: "team",
  source_repo: "https://example.com/team",
  installed_path: `.claude/skills/${name}`,
  installed_at: "2026-01-01T00:00:00Z",
  pinned: false,
  content_hash: "sha256:abc",
  ...extra,
});

export const statusEntry = (name: string, status: string, extra: Record<string, unknown> = {}) => ({
  name,
  installed: "",
  latest: "",
  status,
  target: "claude",
  ...extra,
});

export const registrySkill = (name: string, description: string, extra: Record<string, unknown> = {}) => ({
  name,
  description,
  registry_alias: "team",
  registry_url: "https://example.com/team",
  registry_source: "global",
  metadata: { version: "1.0.0", lifecycle: "stable", tags: "review", owner: "Engineering" },
  ...extra,
});
