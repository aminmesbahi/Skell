// Shared TypeScript types mirroring internal/model/model.go and output structures.

export type Lifecycle =
  | "draft"
  | "experimental"
  | "stable"
  | "deprecated"
  | "archived";

export type SkillStatus =
  | "up-to-date"
  | "outdated"
  | "pinned"
  | "deprecated"
  | "archived"
  | "locally-modified"
  | "unknown"
  | "missing-metadata"
  | "unversioned";

export type DiagnosticSeverity = "error" | "warning" | "info";

// ----- Domain models -------------------------------------------------------

export interface SkillMetadata {
  version: string;
  owner: string;
  lifecycle: Lifecycle;
  scope: string;
  tags: string;
  source_repo: string;
}

export interface RegistrySkill {
  name: string;
  description: string;
  license: string;
  metadata: SkillMetadata;
  registry_alias?: string;
  registry_url?: string;
}

export interface InstalledSkill {
  name: string;
  version: string;
  registry: string;
  source_repo: string;
  source_ref: string;
  installed_path: string;
  installed_at: string;
  pinned: boolean;
  content_hash: string;
}

export interface StatusEntry {
  name: string;
  installed: string;
  latest: string;
  status: SkillStatus;
}

export interface InfoResult {
  name: string;
  skill?: RegistrySkill;
  lock?: InstalledSkill;
  status?: SkillStatus;
}

export interface DiagnosticEntry {
  severity: DiagnosticSeverity;
  code: string;
  message: string;
  hint?: string;
}

export interface ActionEvent {
  action: string;
  skill: string;
  repo?: string;
  dry_run?: boolean;
}

export interface SyncReport {
  installed: string[];
  removed: string[];
}

// ----- File system ---------------------------------------------------------

export interface FileEntry {
  name: string;
  path: string;
  is_dir: boolean;
}

// ----- Audit log -----------------------------------------------------------

export interface AuditEntry {
  timestamp: string;
  action: string;
  skill: string;
  version?: string;
  registry?: string;
  repo?: string;
  user?: string;
}

// ----- Tauri bridge --------------------------------------------------------

export interface SkellResult {
  stdout: string;
  stderr: string;
  success: boolean;
}

// ----- add command ---------------------------------------------------------

export interface AddResult {
  repo: string;
  alias: string;
  skill_name: string;
  registered: boolean;
  installed: boolean;
  dry_run: boolean;
}
