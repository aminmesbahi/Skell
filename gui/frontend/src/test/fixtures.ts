import type {
  InstalledSkill,
  RegistrySkill,
  StatusEntry,
  InfoResult,
  DiagnosticEntry,
  SyncReport,
  AuditEntry,
  FileEntry,
  SkellResult,
} from "@/lib/types";

// ---------------------------------------------------------------------------
// Reusable typed fixtures for tests
// ---------------------------------------------------------------------------

export const mockOkResult = (stdout = ""): SkellResult => ({
  stdout,
  stderr: "",
  success: true,
});

export const mockErrResult = (stderr = "error"): SkellResult => ({
  stdout: "",
  stderr,
  success: false,
});

export const mockInstalledSkill = (
  overrides: Partial<InstalledSkill> = {}
): InstalledSkill => ({
  name: "test-skill",
  version: "1.0.0",
  registry: "default",
  source_repo: "https://github.com/org/repo",
  installed_path: "/home/user/.skell/skills/test-skill",
  installed_at: "2024-01-01T00:00:00Z",
  pinned: false,
  content_hash: "abc123",
  ...overrides,
});

export const mockRegistrySkill = (
  overrides: Partial<RegistrySkill> = {}
): RegistrySkill => ({
  name: "registry-skill",
  description: "A test registry skill",
  license: "MIT",
  metadata: {
    version: "2.0.0",
    owner: "testowner",
    lifecycle: "stable",
    scope: "general",
    tags: "test,demo",
    source_repo: "https://github.com/org/registry-skill",
  },
  ...overrides,
});

export const mockStatusEntry = (
  overrides: Partial<StatusEntry> = {}
): StatusEntry => ({
  name: "test-skill",
  installed: "1.0.0",
  latest: "1.0.0",
  status: "up-to-date",
  ...overrides,
});

export const mockInfoResult = (
  overrides: Partial<InfoResult> = {}
): InfoResult => ({
  name: "test-skill",
  skill: mockRegistrySkill(),
  lock: mockInstalledSkill(),
  status: "up-to-date",
  ...overrides,
});

export const mockDiagnosticEntry = (
  overrides: Partial<DiagnosticEntry> = {}
): DiagnosticEntry => ({
  severity: "error",
  code: "E001",
  message: "Missing version in metadata",
  hint: "Add version field",
  ...overrides,
});

export const mockSyncReport = (
  overrides: Partial<SyncReport> = {}
): SyncReport => ({
  installed: ["skill-a"],
  removed: [],
  ...overrides,
});

export const mockAuditEntry = (
  overrides: Partial<AuditEntry> = {}
): AuditEntry => ({
  timestamp: "2024-01-01T12:00:00Z",
  action: "install",
  skill: "test-skill",
  version: "1.0.0",
  registry: "default",
  repo: "/home/user/project",
  ...overrides,
});

export const mockFileEntry = (overrides: Partial<FileEntry> = {}): FileEntry => ({
  name: "SKILL.md",
  path: "/home/user/.skell/skills/test-skill/SKILL.md",
  is_dir: false,
  ...overrides,
});
