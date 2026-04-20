// Wails v2 generates TypeScript bindings at build time into wailsjs/go/main/App.
// Import from the generated paths — these are created by `wails dev` / `wails build`.
import {
  RunSkell,
  ReadFileContent,
  ListDirectory,
  SkellVersion,
  AuditLogPath,
  GlobalRootDir,
} from "../../wailsjs/go/main/App";
import type {
  InstalledSkill,
  RegistrySkill,
  StatusEntry,
  InfoResult,
  DiagnosticEntry,
  SyncReport,
  FileEntry,
  AuditEntry,
  SkellResult,
  AddResult,
} from "./types";

// ---------------------------------------------------------------------------
// Low-level Wails bridge
// ---------------------------------------------------------------------------

function run(args: string[]): Promise<SkellResult> {
  return RunSkell(args);
}

async function runJSON<T>(args: string[]): Promise<T> {
  const result = await run([...args, "--json"]);
  if (!result.success) throw new Error(result.stderr || "skell command failed");
  const text = result.stdout.trim();
  try {
    return JSON.parse(text) as T;
  } catch {
    // Binary returned non-JSON (e.g. plaintext "no skills found") — treat as empty list.
    if (!text || text.includes("no skills found")) return [] as unknown as T;
    throw new SyntaxError(`Unexpected skell output: ${text.slice(0, 120)}`);
  }
}

// ---------------------------------------------------------------------------
// skell commands
// ---------------------------------------------------------------------------

export async function listInstalled(repo: string): Promise<InstalledSkill[]> {
  return runJSON<InstalledSkill[]>(["list", "--repo", repo]);
}

export async function listInstalledGlobal(): Promise<InstalledSkill[]> {
  return runJSON<InstalledSkill[]>(["list", "--global"]);
}

export async function listRegistry(): Promise<RegistrySkill[]> {
  return runJSON<RegistrySkill[]>(["list", "--source", "registry"]);
}

export async function getStatus(repo: string): Promise<StatusEntry[]> {
  return runJSON<StatusEntry[]>(["status", "--repo", repo]);
}

export async function getInfo(
  skillName: string,
  repo?: string
): Promise<InfoResult> {
  const args = ["info", skillName];
  if (repo) args.push("--repo", repo);
  return runJSON<InfoResult>(args);
}

export async function installSkill(opts: {
  skillName: string;
  repo: string;
  registry?: string;
  registryURL?: string;
  dryRun?: boolean;
}): Promise<SkellResult> {
  const args = ["install", opts.skillName, "--repo", opts.repo];
  if (opts.registry) args.push("--registry", opts.registry);
  if (opts.registryURL) args.push("--registry-url", opts.registryURL);
  if (opts.dryRun) args.push("--dry-run");
  return run(args);
}

export async function upgradeSkill(opts: {
  skillName?: string;
  repo: string;
  force?: boolean;
  dryRun?: boolean;
}): Promise<SkellResult> {
  const args = ["upgrade"];
  if (opts.skillName) args.push(opts.skillName);
  args.push("--repo", opts.repo);
  if (opts.force) args.push("--force");
  if (opts.dryRun) args.push("--dry-run");
  return run(args);
}

export async function removeSkill(opts: {
  skillName: string;
  repo: string;
  dryRun?: boolean;
}): Promise<SkellResult> {
  const args = ["remove", opts.skillName, "--repo", opts.repo];
  if (opts.dryRun) args.push("--dry-run");
  return run(args);
}

export async function pinSkill(opts: {
  skillName: string;
  repo: string;
  version?: string;
}): Promise<SkellResult> {
  const args = ["pin", opts.skillName, "--repo", opts.repo];
  if (opts.version) args.push("--version", opts.version);
  return run(args);
}

export async function unpinSkill(opts: {
  skillName: string;
  repo: string;
}): Promise<SkellResult> {
  return run(["unpin", opts.skillName, "--repo", opts.repo]);
}

export async function syncRepo(opts: {
  repo: string;
  check?: boolean;
  dryRun?: boolean;
}): Promise<SyncReport> {
  const args = ["sync", "--repo", opts.repo];
  if (opts.check) args.push("--check");
  if (opts.dryRun) args.push("--dry-run");
  return runJSON<SyncReport>(args);
}

export async function initRepo(repo: string): Promise<SkellResult> {
  return run(["init", "--repo", repo]);
}

export async function searchSkills(opts: {
  query?: string;
  lifecycle?: string;
  owner?: string;
  tag?: string;
  repo?: string;
}): Promise<RegistrySkill[]> {
  const args = ["search"];
  if (opts.query) args.push(opts.query);
  if (opts.lifecycle) args.push("--lifecycle", opts.lifecycle);
  if (opts.owner) args.push("--owner", opts.owner);
  if (opts.tag) args.push("--tag", opts.tag);
  if (opts.repo) args.push("--repo", opts.repo);
  return runJSON<RegistrySkill[]>(args);
}

export async function doctorCheck(repo: string): Promise<DiagnosticEntry[]> {
  return runJSON<DiagnosticEntry[]>(["doctor", "--repo", repo]);
}

export async function cacheStatus(): Promise<SkellResult> {
  return run(["cache", "status"]);
}

export async function cacheRefresh(repo?: string): Promise<SkellResult> {
  const args = ["cache", "refresh"];
  if (repo) args.push("--repo", repo);
  return run(args);
}

export async function cacheClear(): Promise<SkellResult> {
  return run(["cache", "clear"]);
}

export async function selfUpdateCheck(): Promise<SkellResult> {
  return run(["selfupdate", "--check"]);
}

export async function selfUpdate(): Promise<SkellResult> {
  return run(["selfupdate"]);
}

// ---------------------------------------------------------------------------
// File system helpers (Wails native)
// ---------------------------------------------------------------------------

export function readFileContent(path: string): Promise<string> {
  return ReadFileContent(path);
}

export function listDirectory(path: string): Promise<FileEntry[]> {
  return ListDirectory(path);
}

export function getSkellVersion(): Promise<string> {
  return SkellVersion();
}

// ---------------------------------------------------------------------------
// Audit log (reads ~/.skell/audit.log via Go backend)
// ---------------------------------------------------------------------------

export async function readAuditLog(): Promise<AuditEntry[]> {
  try {
    const auditPath = await AuditLogPath();
    if (!auditPath) return [];
    const content = await ReadFileContent(auditPath).catch(() => "");
    if (!content) return [];

    return content
      .trim()
      .split("\n")
      .filter(Boolean)
      .map((line) => {
        try {
          return JSON.parse(line) as AuditEntry;
        } catch {
          return null;
        }
      })
      .filter((e): e is AuditEntry => e !== null)
      .reverse(); // newest first
  } catch {
    return [];
  }
}

export function getGlobalRootDir(): Promise<string> {
  return GlobalRootDir();
}

export async function addSkillFromURL(opts: {
  url: string;
  repo?: string;
  dryRun?: boolean;
}): Promise<AddResult[]> {
  const args = ["add", opts.url];
  if (opts.repo) args.push("--repo", opts.repo);
  if (opts.dryRun) args.push("--dry-run");
  return runJSON<AddResult[]>(args);
}
