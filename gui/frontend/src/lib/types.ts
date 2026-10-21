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
  | "unversioned"
  | "linked";

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
  compatibility?: string;
  name: string;
  description: string;
  license: string;
  metadata: SkillMetadata;
  registry_alias?: string;
  registry_url?: string;
  /** "global" when from the global manifest, "local" when from a repo-local manifest. */
  registry_source?: "global" | "local" | string;
}

export interface InstalledSkill {
  name: string;
  version: string;
  registry: string;
  source_repo: string;
  installed_path: string;
  installed_at: string;
  pinned: boolean;
  content_hash: string;
  /** Source commit the skill was installed from (git sources). */
  commit?: string;
  source_path?: string;
  /** True for skills linked from a local folder with `skell link`. */
  linked?: boolean;
}

export interface StatusEntry {
  name: string;
  installed: string;
  latest: string;
  status: SkillStatus;
  /** Source content differs from the installed copy (even without a version bump). */
  changed?: boolean;
  target?: string;
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
  untracked?: string[];
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

// ----- Wails bridge (generated bindings) -----------------------------------

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

// SkillSource - persistent global sources managed in Settings (git or local folder)
export interface SkillSource {
  alias: string;
  url: string;
  is_local: boolean;
}

// ----- Validation (skell validate) -----------------------------------------

export interface SkillValidationFinding {
  severity: "error" | "warning" | "info" | string;
  category: string;
  message: string;
  file?: string;
  line?: number;
}

export interface SkillAnalysis {
  word_count: number;
  code_block_ratio: number;
  imperative_ratio: number;
  information_density: number;
  instruction_specificity: number;
  section_count: number;
  list_item_count: number;
  has_content: boolean;
  contamination_score: number;
  contamination_level: string;
  code_languages?: string[];
  mismatched_categories?: string[];
  language_mismatch: boolean;
  has_contamination: boolean;
  total_tokens: number;
  skill_tokens: number;
}

export interface SkillValidationResult {
  target?: string;
  name: string;
  errors: number;
  warnings: number;
  findings: SkillValidationFinding[];
  analysis?: SkillAnalysis;
}

// SkillPreview mirrors gui.SkillPreview — returned by PreviewRegistrySkill for
// the Discover Skills preview modal.
export interface SkillPreview {
  readme_content: string;
  source_path: string;
  source_url: string;
  source_type: "git" | "local" | string;
  found: boolean;
}

// ----- Review (skell review --json) ----------------------------------------

/** What installing a skill would bring in; mirrors engine.Review. */
export interface SkillReview {
  skill: RegistrySkill;
  registry: string;
  registry_url: string;
  commit?: string;
  files: string[];
  scripts?: string[];
  allowed_tools?: string[];
  links?: string[];
  total_bytes: number;
  warnings?: string[];
  /** Static security scan of the skill's files. */
  risk?: ScanReport;
  /** Publisher-signature check against the trusted publishers. */
  signature?: Omit<VerifyEntry, "skill" | "repo">;
}

// ----- Diff (skell diff --json) --------------------------------------------

/** Installed → latest comparison; mirrors engine.DiffResult. */
export interface SkillDiff {
  name: string;
  registry: string;
  installed_version?: string;
  latest_version?: string;
  installed_commit?: string;
  latest_commit?: string;
  patch: string;
  locally_modified: boolean;
}

// ----- Catalog (skell catalog --json) --------------------------------------

export interface CatalogSource {
  id: string;
  name: string;
  url: string;
  description: string;
  tags?: string[];
}

// ----- Mirrors (skell mirror … --json) -------------------------------------

export interface MirrorReport {
  targets: string[] | null;
  copied: string[] | null;
  removed: string[] | null;
}

// ----- Security, trust, bundles, fleet, tests ------------------------------

export type RiskLevel = "clean" | "low" | "medium" | "high" | "critical";
export type FindingSeverity = "info" | "low" | "medium" | "high" | "critical";

/** One thing the security scanner noticed; mirrors skillscan.Finding. */
export interface ScanFinding {
  rule: string;
  severity: FindingSeverity;
  message: string;
  file: string;
  line?: number;
  evidence?: string;
}

/** Result of `skell scan --json` for one skill; mirrors skillscan.Report. */
export interface ScanReport {
  repo?: string;
  skill: string;
  path?: string;
  score: number;
  level: RiskLevel;
  files: number;
  findings: ScanFinding[];
}

export type SignatureStatus = "trusted" | "untrusted-key" | "tampered" | "invalid" | "unsigned";

/** Result of `skell verify --json` for one skill; mirrors signing.Result. */
export interface VerifyEntry {
  repo?: string;
  skill: string;
  status: SignatureStatus;
  publisher?: string;
  key_id?: string;
  signed_at?: string;
  detail?: string;
}

export interface TrustedKey {
  publisher: string;
  key_id: string;
  public_key: string;
  added_at: string;
}

export interface BundleInfo {
  name: string;
  description?: string;
  skills: string[];
  /** How many of the bundle's skills are already installed. */
  installed: number;
}

export interface FleetSkill {
  name: string;
  version: string;
  commit?: string;
  registry: string;
  status: string;
  pinned?: boolean;
  linked?: boolean;
  modified?: boolean;
  risk: RiskLevel;
  risk_score: number;
  signature: SignatureStatus;
  publisher?: string;
}

export interface FleetRepo {
  path: string;
  name: string;
  target?: string;
  mirrors?: string[];
  skills: FleetSkill[];
  errors: number;
  warnings: number;
  health: "healthy" | "attention" | "problem";
  note?: string;
}

export interface FleetVersionCount {
  version: string;
  repos: number;
  paths?: string[];
}

export interface FleetMatrixRow {
  name: string;
  repos: number;
  versions: FleetVersionCount[];
  spread: boolean;
}

export interface FleetSummary {
  repos: number;
  installs: number;
  unique_skills: number;
  outdated: number;
  modified: number;
  pinned: number;
  high_risk: number;
  unsigned: number;
  version_spread: number;
  unhealthy_repos: number;
}

/** Result of `skell report --json`; mirrors engine.FleetReport. */
export interface FleetReport {
  generated_at: string;
  root?: string;
  summary: FleetSummary;
  repos: FleetRepo[];
  matrix: FleetMatrixRow[];
  events?: AuditEntry[];
}

export type TestCaseStatus = "pass" | "fail" | "error" | "skipped";

export interface TestCaseResult {
  name: string;
  status: TestCaseStatus;
  failures?: string[];
  duration_ms: number;
  output?: string;
}

/** Result of running one skill's eval suite; mirrors skilltest.Result. */
export interface SkillTestResult {
  skill: string;
  suite?: string;
  agent?: string;
  static_problems?: string[];
  cases: TestCaseResult[];
  passed: number;
  failed: number;
  errored: number;
  skipped: number;
}

export interface SkillTestRun {
  results: SkillTestResult[];
  without_evals: string[];
}
