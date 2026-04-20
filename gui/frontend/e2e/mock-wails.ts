/**
 * Wails runtime mock injected by Playwright before each test page load.
 *
 * The Wails frontend bridge calls `window.go.main.App.METHOD()`.
 * This script installs a mock implementation so the React app boots
 * normally in the browser without a native Wails host.
 *
 * Import this in E2E specs:
 *   import { injectWailsMock, type WailsMock } from "./mock-wails";
 */

import type { Page } from "@playwright/test";

export interface WailsMock {
  runSkell: (result: { stdout: string; stderr: string; success: boolean }) => void;
  selectDirectory: (path: string) => void;
  skellVersion: (version: string) => void;
  auditLogPath: (path: string) => void;
  readFileContent: (content: string) => void;
  listDirectory: (entries: { name: string; path: string; is_dir: boolean }[]) => void;
}

/**
 * Inject a controllable window.go mock into the page before navigation.
 * Returns helper functions to set return values for specific calls.
 */
export async function injectWailsMock(page: Page): Promise<void> {
  await page.addInitScript(() => {
    // Default responses — can be overridden per-test by exposing window.__wailsMockConfig
    const defaults = {
      RunSkell: () => Promise.resolve({ stdout: "[]", stderr: "", success: true }),
      ReadFileContent: () => Promise.resolve(""),
      ListDirectory: () => Promise.resolve([]),
      SkellVersion: () => Promise.resolve("v0.1.0-test"),
      SelectDirectory: () => Promise.resolve(""),
      AuditLogPath: () => Promise.resolve(""),
    };

    // Allow tests to override via window.__wailsOverrides
    const overrides: Record<string, () => Promise<unknown>> = {};
    Object.defineProperty(window, "__wailsOverride", {
      set(key: string) {
        // not used directly; overrides set via __wailsSetOverride
        void key;
      },
    });
    (window as Record<string, unknown>)["__wailsSetOverride"] = (
      method: string,
      fn: () => Promise<unknown>
    ) => {
      overrides[method] = fn;
    };

    const handler = (method: keyof typeof defaults) => (...args: unknown[]) => {
      if (overrides[method]) return overrides[method](...args);
      return defaults[method](...args);
    };

    (window as Record<string, unknown>)["go"] = {
      main: {
        App: {
          RunSkell: handler("RunSkell"),
          ReadFileContent: handler("ReadFileContent"),
          ListDirectory: handler("ListDirectory"),
          SkellVersion: handler("SkellVersion"),
          SelectDirectory: handler("SelectDirectory"),
          AuditLogPath: handler("AuditLogPath"),
        },
      },
    };
  });
}
