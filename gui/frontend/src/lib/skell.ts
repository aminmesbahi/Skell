import {
  RunSkell,
  ReadFileContent,
  ListDirectory,
  SkellVersion,
  AuditLogPath,
  GlobalRootDir,
  IsRepoInitialized,
  SupportedTargets,
  DetectTargets,
  ActiveTarget,
  ListSkillSources,
  AddSkillSource,
  RemoveSkillSource,
  PreviewRegistrySkill,
  ValidateSkill,
  SkellPresent as _SkellPresent,
} from "../../bindings/skell-gui/app";
import type * as wailsModels from "../../bindings/skell-gui/models";
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
  SkillPreview,
  SkillValidationResult,
} from "./types";

function run(args: string[]): Promise<SkellResult> {
  return RunSkell(args);
}

async function runJSON<T>(args: string[]): Promise<T> {
  const result = await run([...args, "--json"]);
  const text = result.stdout.trim();

  // Always try to parse stdout as JSON first — some commands (e.g. doctor,
  // validate) exit non-zero when they find issues but still write valid JSON
  // to stdout. Only fall back to error/plain-text handling when parse fails.
  if (text) {
    try {
      return JSON.parse(text) as T;
    } catch {
      // fall through to plain-text / error handling below
    }
  }

  // Non-zero exit with no parseable stdout → surface the error message.
  if (!result.success) throw new Error(result.stderr || "skell command failed");

  // Plain-text "nothing to report" responses — treat as empty list.
  if (
    !text ||
    text.includes("no skills found") ||
    text.includes("no issues found") ||
    text.includes("already up to date") ||
    text.includes("ok ") // doctor "  ok  <path> — no issues found"
  ) {
    return [] as unknown as T;
  }

  throw new SyntaxError(`Unexpected skell output: ${text.slice(0, 120)}`);
}

export async function listInstalled(repo: string, target?: string): Promise<InstalledSkill[]> {
  const args = ["list", "--repo", repo];
  if (target) args.push("--target", target);
  return runJSON<InstalledSkill[]>(args);
}

export async function listInstalledGlobal(): Promise<InstalledSkill[]> {
  return runJSON<InstalledSkill[]>(["list", "--global"]);
}

export async function listRegistry(): Promise<RegistrySkill[]> {
  return runJSON<RegistrySkill[]>(["list", "--source", "registry"]);
}

export async function getStatus(repo: string, target?: string): Promise<StatusEntry[]> {
  const args = repo === "global" ? ["status", "--global"] : ["status", "--repo", repo];
  if (target) args.push("--target", target);
  return runJSON<StatusEntry[]>(args);
}

export async function getInfo(
  skillName: string,
  repo?: string,
  target?: string
): Promise<InfoResult> {
  const args = ["info", skillName];
  if (repo === "global") {
    // `skell info` has no --global flag; resolve the actual root path instead.
    const globalRoot = await getGlobalRootDir();
    args.push("--repo", globalRoot);
  } else if (repo) {
    args.push("--repo", repo);
  }
  if (target) args.push("--target", target);
  return runJSON<InfoResult>(args);
}

export async function installSkill(opts: {
  skillName: string;
  repo: string;
  registry?: string;
  registryURL?: string;
  dryRun?: boolean;
  noValidate?: boolean;
  target?: string;
}): Promise<SkellResult> {
  const args = ["install", opts.skillName];
  if (opts.repo === "global") {
    args.push("--global");
  } else {
    args.push("--repo", opts.repo);
  }
  if (opts.registry) args.push("--registry", opts.registry);
  if (opts.registryURL) args.push("--registry-url", opts.registryURL);
  if (opts.dryRun) args.push("--dry-run");
  if (opts.noValidate) args.push("--no-validate");
  if (opts.target) args.push("--target", opts.target);
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
  if (opts.repo === "global") {
    args.push("--global");
  } else {
    args.push("--repo", opts.repo);
  }
  if (opts.force) args.push("--force");
  if (opts.dryRun) args.push("--dry-run");
  return run(args);
}

export async function removeSkill(opts: {
  skillName: string;
  repo: string;
  dryRun?: boolean;
}): Promise<SkellResult> {
  const args = ["remove", opts.skillName];
  if (opts.repo === "global") {
    args.push("--global");
  } else {
    args.push("--repo", opts.repo);
  }
  if (opts.dryRun) args.push("--dry-run");
  return run(args);
}

export async function pinSkill(opts: {
  skillName: string;
  repo: string;
  version?: string;
}): Promise<SkellResult> {
  const args = ["pin", opts.skillName];
  if (opts.repo === "global") {
    args.push("--global");
  } else {
    args.push("--repo", opts.repo);
  }
  if (opts.version) args.push("--version", opts.version);
  return run(args);
}

export async function unpinSkill(opts: {
  skillName: string;
  repo: string;
}): Promise<SkellResult> {
  const args = ["unpin", opts.skillName];
  if (opts.repo === "global") {
    args.push("--global");
  } else {
    args.push("--repo", opts.repo);
  }
  return run(args);
}

export async function syncRepo(opts: {
  repo: string;
  check?: boolean;
  dryRun?: boolean;
}): Promise<SyncReport> {
  const args = ["sync"];
  if (opts.repo === "global") {
    args.push("--global");
  } else {
    args.push("--repo", opts.repo);
  }
  if (opts.check) args.push("--check");
  if (opts.dryRun) args.push("--dry-run");
  return runJSON<SyncReport>(args);
}

export async function initRepo(repo: string, target?: string): Promise<SkellResult> {
  const args = ["init", "--repo", repo];
  if (target) args.push("--target", target);
  return run(args);
}

export type AgentTarget = wailsModels.AgentTarget;

/** Extract the target ID from an installed_path like ".cursor/skills/blueprint". */
export function targetFromInstalledPath(installedPath: string): string {
  if (!installedPath) return "";
  const seg = installedPath.split(/[/\\]/)[0];
  if (!seg) return "";
  // Strip leading dot: ".cursor" → "cursor"
  return seg.startsWith(".") ? seg.slice(1) : seg;
}

export async function listSupportedTargets(): Promise<AgentTarget[]> {
  const result = await SupportedTargets();
  return result ?? [];
}

export async function detectRepoTargets(repo: string): Promise<AgentTarget[]> {
  const result = await DetectTargets(repo);
  return result ?? [];
}

export async function activeRepoTarget(repo: string): Promise<string> {
  return await ActiveTarget(repo);
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
  if (repo === "global") return runJSON<DiagnosticEntry[]>(["doctor", "--global"]);
  return runJSON<DiagnosticEntry[]>(["doctor", "--repo", repo]);
}

/**
 * Validate installed skills in a repo against the Agent Skills spec.
 * Pass skillName to validate a single skill, or "" for all. When full is true,
 * offline content & contamination analysis is included.
 */
export async function validateSkills(
  repo: string,
  skillName = "",
  full = false
): Promise<SkillValidationResult[]> {
  return ValidateSkill(repo, skillName, full) as Promise<SkillValidationResult[]>;
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

// ─────────────────────────────────────────────────────────────────────────────
// Global Skill Sources (managed in Settings)
// ─────────────────────────────────────────────────────────────────────────────

export async function listSkillSources(): Promise<any[]> {
  const result = await ListSkillSources();
  return result ?? [];
}

export async function addSkillSource(alias: string, url: string): Promise<void> {
  await AddSkillSource(alias, url);
}

export async function removeSkillSource(alias: string): Promise<void> {
  await RemoveSkillSource(alias);
}

// ---------------------------------------------------------------------------
// File system helpers (Wails native)
// ---------------------------------------------------------------------------

export function readFileContent(path: string): Promise<string> {
  return ReadFileContent(path);
}

export async function listDirectory(path: string): Promise<FileEntry[]> {
  const result = await ListDirectory(path);
  return result ?? [];
}

export function getSkellVersion(): Promise<string> {
  return SkellVersion();
}

// skellPresent is the preferred way for UI to detect a usable CLI without
// triggering a full command (and its error toast side effects).
export function skellPresent(): Promise<boolean> {
  if (typeof _SkellPresent === "function") {
    return _SkellPresent();
  }
  // Fallback for dev snapshots before bindings are regenerated: assume present
  // (the actual RunSkell paths will surface the friendly not-found message).
  return Promise.resolve(true);
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
      .map((line: string) => {
        try {
          return JSON.parse(line) as AuditEntry;
        } catch {
          return null;
        }
      })
      .filter((e: AuditEntry | null): e is AuditEntry => e !== null)
      .reverse(); // newest first
  } catch {
    return [];
  }
}

export function getGlobalRootDir(): Promise<string> {
  return GlobalRootDir();
}

/** Returns the platform path to ~/.skell/audit.log (for display/diagnostics). */
export function getAuditLogPath(): Promise<string> {
  return AuditLogPath();
}

/** Returns true when the given repo path contains a Skell manifest (.claude/skell.toml). */
export function isRepoInitialized(repoPath: string): Promise<boolean> {
  if (!repoPath || repoPath === "global") return Promise.resolve(true);
  return IsRepoInitialized(repoPath);
}

export async function previewRegistrySkill(
  registryAlias: string,
  registryURL: string,
  skillName: string
): Promise<SkillPreview> {
  return PreviewRegistrySkill(registryAlias, registryURL, skillName) as Promise<SkillPreview>;
}

export async function addSkillFromURL(opts: {
  url: string;
  repo?: string;
  dryRun?: boolean;
}): Promise<AddResult[]> {
  const args = ["add", opts.url];
  if (opts.repo === "global") {
    args.push("--global");
  } else if (opts.repo) {
    args.push("--repo", opts.repo);
  }
  if (opts.dryRun) args.push("--dry-run");
  return runJSON<AddResult[]>(args);
}
