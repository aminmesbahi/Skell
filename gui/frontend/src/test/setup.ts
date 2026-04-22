import "@testing-library/jest-dom";
import { vi } from "vitest";

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
      },
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
