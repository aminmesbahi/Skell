/**
 * Stand-in for `@wailsio/runtime`, used only by the e2e dev server
 * (`vite --mode e2e`, see vite.config.ts). The real runtime talks to the native
 * Wails host; this one answers the generated service bindings from
 * `window.__e2e`, which each spec configures through `injectWails()` in
 * ./mock-wails.ts.
 *
 * Calls are matched by service-method NAME: the numeric IDs in the generated
 * bindings are parsed out of bindings/skell-gui/app.ts, so regenerating the
 * bindings never breaks the tests.
 */
import appBindings from "../bindings/skell-gui/app.ts?raw";

interface SkellResult {
  stdout?: string;
  stderr?: string;
  success?: boolean;
}

interface E2EState {
  /** `skell` CLI answers keyed by "cmd", "cmd sub" or the full argument list. */
  skell?: Record<string, SkellResult | string>;
  /** Return values for other service methods, keyed by method name. */
  values?: Record<string, unknown>;
  calls: { method: string; args: unknown[] }[];
}

const methodById = new Map<number, string>();
for (const match of appBindings.matchAll(/export function (\w+)\([^)]*\)[^{]*\{\s*return \$Call\.ByID\((\d+)/g)) {
  methodById.set(Number(match[2]), match[1]);
}

const defaults: Record<string, unknown> = {
  SkellVersion: "v0.0.0-e2e",
  SkellPresent: true,
  IsRepoInitialized: true,
  GlobalRootDir: "/tmp/skell-e2e-global",
  AuditLogPath: "",
  ActiveTarget: "claude",
  SelectDirectory: "",
  ReadFileContent: "",
  ListDirectory: [],
  ListSkillSources: [],
  DetectTargets: [],
  SupportedTargets: [
    { id: "claude", displayName: "Claude Code", dir: ".claude", detected: true },
    { id: "cursor", displayName: "Cursor", dir: ".cursor", detected: false },
  ],
  ValidateSkill: [],
  PreviewRegistrySkill: { found: false, source_path: "", readme_content: "" },
  ReadSkillMetadata: {},
  ResolveSkillSourceRepoURL: "",
};

function state(): E2EState {
  const w = window as unknown as { __e2e?: E2EState };
  w.__e2e ??= { calls: [] };
  w.__e2e.calls ??= [];
  return w.__e2e;
}

function runSkell(args: string[]): SkellResult {
  const answers = state().skell ?? {};
  for (const key of [args.join(" "), args.slice(0, 2).join(" "), args[0]]) {
    const answer = answers[key];
    if (answer === undefined) continue;
    return typeof answer === "string" ? { stdout: answer, stderr: "", success: true } : { stdout: "", stderr: "", success: true, ...answer };
  }
  return { stdout: "[]", stderr: "", success: true };
}

function dispatch(method: string, args: unknown[]): unknown {
  const e2e = state();
  e2e.calls.push({ method, args });
  if (method === "RunSkell") return runSkell(args[0] as string[]);
  if (e2e.values && method in e2e.values) return e2e.values[method];
  if (method in defaults) return defaults[method];
  throw new Error(`e2e mock: no answer for service method ${method}`);
}

const answer = (method: string, args: unknown[]) => new Promise((resolve, reject) => {
  try { resolve(dispatch(method, args)); } catch (error) { reject(error); }
});

export const Call = {
  ByID: (id: number, ...args: unknown[]) => {
    const method = methodById.get(id);
    return method ? answer(method, args) : Promise.reject(new Error(`e2e mock: unknown binding id ${id}`));
  },
  ByName: (name: string, ...args: unknown[]) => answer(name.split(".").at(-1) ?? name, args),
};

export const CancellablePromise = Promise;

// The generated models only use Create for decoding; identity is enough here.
const identity = (value: unknown) => value;
export const Create = new Proxy({}, { get: () => identity });
