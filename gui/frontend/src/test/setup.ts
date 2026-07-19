import "@testing-library/jest-dom";
import { vi } from "vitest";

// Wails v3 generated bindings call @wailsio/runtime directly instead of using
// the old window.go bridge. Route those calls back through the bridge below so
// existing tests can replace individual service methods with vi.fn().
vi.mock("@wailsio/runtime", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@wailsio/runtime")>();
  const methodByID: Record<number, string> = {
    2394438090: "ActiveTarget",
    2803609340: "AddSkillSource",
    3629702033: "AuditLogPath",
    1008264935: "ContributeMetadata",
    2920689880: "DetectTargets",
    2693605881: "GlobalRootDir",
    3848862957: "IsRepoInitialized",
    3882022436: "ListDirectory",
    3339277332: "ListSkillSources",
    3669316229: "PreviewRegistrySkill",
    2440131916: "ReadFileContent",
    2621935323: "ReadSkillMetadata",
    2166783643: "RemoveSkillSource",
    1846080558: "ResolveSkillSourceRepoURL",
    2394374327: "RunSkell",
    1735672136: "SelectDirectory",
    2004525469: "SkellPresent",
    3500197102: "SkellVersion",
    1734000213: "SupportedTargets",
    567263264: "ValidateSkill",
  };

  return {
    ...actual,
    Call: {
      ...actual.Call,
      ByID: vi.fn((id: number, ...args: unknown[]) => {
        const method = methodByID[id];
        const app = window.go?.main?.App as unknown as Record<string, (...callArgs: unknown[]) => unknown>;
        if (!method || typeof app?.[method] !== "function") {
          return Promise.reject(new Error(`Unmocked Wails method ID: ${id}`));
        }
        return app[method](...args);
      }),
    },
  };
});

// ---------------------------------------------------------------------------
// Mock the Wails runtime bridge (window.go) so imports of wailsjs bindings
// don't throw "Cannot read properties of undefined" in jsdom.
// ---------------------------------------------------------------------------
const noop = () => Promise.resolve(undefined);

Object.defineProperty(window, "go", {
  value: {
    main: {
      App: {
        RunSkell: vi.fn().mockResolvedValue({ stdout: "", stderr: "", success: true }),
        ReadFileContent: vi.fn().mockResolvedValue(""),
        ListDirectory: vi.fn().mockResolvedValue([]),
        SkellVersion: vi.fn().mockResolvedValue("0.1.0"),
        SelectDirectory: vi.fn().mockResolvedValue(""),
        AuditLogPath: vi.fn().mockResolvedValue(""),
        GlobalRootDir: vi.fn().mockResolvedValue(""),
        ActiveTarget: vi.fn().mockResolvedValue(""),
        AddSkillSource: vi.fn().mockResolvedValue(undefined),
        ContributeMetadata: vi.fn().mockResolvedValue({}),
        DetectTargets: vi.fn().mockResolvedValue([]),
        IsRepoInitialized: vi.fn().mockResolvedValue(false),
        ListSkillSources: vi.fn().mockResolvedValue([]),
        PreviewRegistrySkill: vi.fn().mockResolvedValue({ found: false }),
        ReadSkillMetadata: vi.fn().mockResolvedValue({}),
        RemoveSkillSource: vi.fn().mockResolvedValue(undefined),
        ResolveSkillSourceRepoURL: vi.fn().mockResolvedValue(""),
        SkellPresent: vi.fn().mockResolvedValue(true),
        SupportedTargets: vi.fn().mockResolvedValue([]),
        ValidateSkill: vi.fn().mockResolvedValue([]),
      },
    },
  },
  writable: true,
});

Object.defineProperty(window, "runtime", {
  value: {
    Call: {
      ByID: vi.fn().mockImplementation((_id: number, ..._args: unknown[]) => Promise.resolve(undefined)),
    },
  },
  writable: true,
});

// Suppress noisy console.error output from React act() warnings in tests
const originalError = console.error.bind(console);
console.error = (...args: unknown[]) => {
  const msg = String(args[0]);
  if (
    msg.includes("Warning: An update to") ||
    msg.includes("Warning: ReactDOM.render") ||
    msg.includes("act(")
  ) {
    return;
  }
  originalError(...args);
};

export { noop };
