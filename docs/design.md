# Skell

**Govern, install, and sync engineering skills at scale.**

## Design Document
**Version:** 0.2 (Final Draft)
**Status:** Ready for implementation

---

# 1. Product Overview

## What Skell is

Skell is a cross-platform skill package manager for [Agent Skills](https://agentskills.io), the open format for giving AI coding agents new capabilities.

Skell solves a gap: Agent Skills has a specification and a validator (`skills-ref`), but no tooling for installing, updating, removing, or distributing skills across teams and repositories. Skell fills that gap.

It is designed for organizations that use AI-assisted development and need a controlled, repeatable way to manage skill distribution, local setup consistency, and skill lifecycle across developers and repositories. It starts as a CLI-first product and later adds a minimal desktop UI for discovery, review, and bulk actions.

## What Skell is not

- Not only a viewer.
- Not a thin wrapper around git clone or copy-paste.
- Not a cloud-first governance platform in its first version.
- Not a skills runtime. Skell does not execute skills or interact with Claude Code's agent loop. It only manages files on disk.
- Not a git wrapper. Skell does not commit, push, or manage branches in project repositories.
- Not a secrets manager. Registry authentication is delegated to standard git credential helpers.

---

# 2. Problem Statement

Teams using engineering skills across many repositories face recurring operational problems:

1. New or updated skills are difficult to discover.
2. Local skill folders become cluttered over time with stale, sprint-specific, or obsolete skills, increasing token consumption in Claude Code.
3. Registry and local installations drift apart silently.
4. Teams cannot easily tell which skills are installed where.
5. Ownership, maturity, and intended usage of skills are often unclear.
6. Onboarding and staff turnover increase inconsistency in local skill setups.
7. Enterprises need consistency and governance, but developers still need local flexibility.

Skell exists to reduce this drift, clutter, and inconsistency while keeping daily usage simple.

---

# 3. Vision

Provide a reliable operational layer for managing Agent Skills across repositories and teams.

Skell should let organizations:
- Keep a central source of truth for skills.
- Safely distribute skills into many repositories.
- Detect drift between registry and local copies.
- Remove obsolete or irrelevant skills.
- Preserve local control where justified.
- Support enterprise governance without making small-team usage painful.

---

# 4. Target Users

## 4.1 Platform Administrator
**Responsibilities:** maintain the central skill registry; classify skill lifecycle and ownership; guide recommended usage across teams.

**Needs:** visibility into available skills; ability to mark stable, experimental, or deprecated skills; confidence that local installs can be compared against source-of-truth.

## 4.2 Team Lead / Staff Engineer
**Responsibilities:** keep repositories aligned with team workflows; roll out useful skills; remove no-longer-relevant skills; reduce setup inconsistency during team changes.

**Needs:** install/update/remove across selected repositories; safe bulk actions; clarity on what is outdated, modified, or pinned.

## 4.3 Developer
**Responsibilities:** maintain a clean and useful local working environment; adopt the right skills for active work; avoid stale skill clutter.

**Needs:** see what is installed locally; install useful skills quickly; update safely; remove irrelevant skills without fear of breaking local workflows.

---

# 5. Product Goals

## 5.1 Primary Goals
- Make skills discoverable.
- Make install/update/remove easy and safe.
- Detect and surface drift between registry and local repos.
- Reduce local clutter and outdated skills.
- Support enterprise governance metadata and lifecycle states.
- Remain lightweight enough for smaller teams.

## 5.2 Secondary Goals
- Make onboarding repeatable.
- Reduce local configuration entropy.
- Support future auditability and policy enforcement.

## 5.3 Non-Goals for MVP
- No required cloud backend.
- No mandatory user authentication.
- No role-based access control.
- No IDE plugin.
- No complex policy engine.
- No automatic destructive cleanup without explicit user action.

---

# 6. Core Concepts

## 6.1 Registry Skill
A skill defined in a central registry source, following the Agent Skills specification: a directory containing a `SKILL.md` with YAML frontmatter plus optional subdirectories (`scripts/`, `references/`, `assets/`).

## 6.2 Installed Skill
A local copy of a registry skill, installed into a specific repository under `.claude/skills/<skill-name>/`.

## 6.3 Local Override
An installed skill whose files have been modified locally after installation. Skell detects this via hash comparison against the lock file baseline.

## 6.4 Unknown Local Skill
A skill present in `.claude/skills/` in a repository but not recognized in any configured registry. These are surfaced, not silently ignored.

## 6.5 Pinned Skill
An installed skill intentionally fixed to a specific version. `skell upgrade` skips pinned skills unless `--force` is passed.

## 6.6 Lifecycle States
Skills in the registry carry a lifecycle state:

| State | Meaning |
|---|---|
| `draft` | Work in progress, not for general use |
| `experimental` | Available but not yet stable |
| `stable` | Recommended for general use |
| `deprecated` | Superseded or no longer recommended |
| `archived` | Frozen, no further updates |

---

# 7. Product Principles

1. **CLI-first core.** The core must be scriptable, testable, and cross-platform before any GUI is built.
2. **Safe by default.** Destructive actions support dry-run and require explicit confirmation.
3. **Registry-aware, not registry-dependent.** Skell works best with a registry, but local workflows remain understandable and manageable without one.
4. **Enterprise-aware without enterprise bloat.** Enterprise metadata and policy support are designed in from the start but do not force complexity onto smaller teams.
5. **Drift visibility matters.** Skell must clearly show when local installs differ from the registry or when local overrides exist.
6. **Never overwrite user work silently.** Locally modified skills must never be silently overwritten.

---

# 8. Manifest Specification

## 8.1 Alignment with Agent Skills Spec

Skell does **not** introduce a separate JSON manifest file. Doing so would create two sources of truth and inevitable drift. Instead, Skell reads metadata directly from the `SKILL.md` frontmatter defined by the Agent Skills specification.

The `metadata` block in frontmatter is where Skell-specific fields live:

```yaml
---
name: pdf-processing
description: Extract PDF text, fill forms, merge files. Use when handling PDFs.
license: Apache-2.0
compatibility: Requires Python 3.10+
metadata:
  version: "1.2.0"
  owner: platform-team
  lifecycle: stable
  scope: shared
  tags: documents, extraction
  source_repo: https://github.com/mycompany/skills-registry
---
```

## 8.2 Skell Metadata Fields

| Field | Required by Skell | Description |
|---|---|---|
| `version` | Yes | Semantic version string (e.g. `"1.2.0"`) |
| `owner` | Recommended | Team or individual responsible for this skill |
| `lifecycle` | Recommended | One of: draft, experimental, stable, deprecated, archived |
| `scope` | No | Intended scope: `shared`, `team`, `project`, `personal` |
| `tags` | No | Comma-separated tags for search and filtering |
| `source_repo` | Yes (when published) | URL of the registry this skill originates from |

**Note:** `name` is already required by the Agent Skills spec and must match the directory name. Skell uses `name` as the canonical skill identifier, there is no separate `id` field.

## 8.3 Version Policy

Skell recommends but does not enforce semantic versioning. If `metadata.version` is absent, the skill is treated as `unversioned` and will always be replaced on install or upgrade (with a warning). Pinning an unversioned skill is not supported.

---

# 9. Skell Manifest, `skell.toml`

`skell.toml` records which skills are declared for a given scope. It is the source of truth for what should be installed, not what currently is installed (that is the lock file's job).

## 9.1 Two-Level Hierarchy

**Global manifest** (`~/.skell/skell.toml`): The user's personal baseline. Applied to all repositories that do not have a local manifest.

**Local manifest** (`.claude/skell.toml` inside a repository): Applies exclusively to that repository. When a local manifest is present, the global manifest is not consulted. There is no merging, local wins entirely.

This makes behavior predictable. A repository with a local manifest is fully self-contained.

## 9.2 Format

```toml
# .claude/skell.toml

[registries]
default = "https://github.com/mycompany/skills-registry"
public  = "https://github.com/agentskills/agentskills"

[skills]
# name = { version = "...", registry = "...", pinned = false }
pdf-processing      = { version = "1.2.0", registry = "default" }
code-review         = { version = "2.0.0", registry = "default", pinned = true }
typescript-patterns = { version = "latest", registry = "public" }
```

## 9.3 Field Reference

| Field | Required | Description |
|---|---|---|
| `version` | Yes | Exact version string or `"latest"` |
| `registry` | Yes | Registry alias from `[registries]` |
| `pinned` | No | If `true`, `skell upgrade` skips this skill. Default: `false` |

---

# 10. Lock File, `skell.lock`

## 10.1 Why from v1

Without a lock file, Skell cannot reliably:
- Detect local modifications (no baseline to compare against).
- Know the installed version without re-fetching the registry.
- Operate offline.
- Guarantee reproducible installs across machines.

The lock file ships in v1. It is not optional.

## 10.2 Location

`.claude/skell.lock` alongside `skell.toml`.

## 10.3 Format

```json
{
  "skell_version": "0.1.0",
  "locked_at": "2026-04-12T10:00:00Z",
  "skills": [
    {
      "name": "pdf-processing",
      "version": "1.2.0",
      "registry": "default",
      "source_repo": "https://github.com/mycompany/skills-registry",
      "source_ref": "v1.2.0",
      "installed_path": ".claude/skills/pdf-processing",
      "installed_at": "2026-04-12T10:00:00Z",
      "pinned": false,
      "content_hash": "sha256:a1b2c3d4..."
    }
  ]
}
```

## 10.4 Lock File Rules

- `skell.lock` is generated and updated by Skell. It should not be hand-edited.
- `skell.lock` **should** be committed to version control. This enables reproducible installs and drift detection across team members.
- `content_hash` is a SHA-256 hash of the installed skill's file tree. Skell compares this hash on every status check to detect local modifications.
- If `skell.lock` is absent but `skell.toml` exists, `skell sync` creates it.

---

# 11. Registry

## 11.1 Structure

A registry is a git repository where each skill occupies a top-level subdirectory:

```
skills-registry/
├── pdf-processing/
│   ├── SKILL.md
│   └── scripts/
│       └── extract.py
├── code-review/
│   └── SKILL.md
└── typescript-patterns/
    ├── SKILL.md
    └── references/
        └── REFERENCE.md
```

Any git remote following this layout is a valid Skell registry. No server-side component is required.

## 11.2 Cache

Skell caches registry contents in `~/.skell/cache/<registry-alias>/`. Before any install or upgrade, Skell performs a `git fetch` on the cached registry. This means:

- No repeated full cloning.
- Offline mode works if cache is warm.
- Cache can be cleared with `skell cache clear`.

Skell copies skill files from cache into the target repository. It does **not** use git submodules. This avoids git complexity inside project repositories.

## 11.3 Multiple Registries

Multiple registries can be configured in `skell.toml`. When the same skill name appears in two registries, the first registry listed in `[registries]` takes precedence. To override, specify `registry` explicitly in the skill entry.

---

# 12. Functional Requirements

## 12.1 Registry Operations
- Fetch and cache registry sources.
- Parse `SKILL.md` frontmatter for metadata.
- Build a local index of available skills, versions, and lifecycle states.

## 12.2 Local Repository Scanning
- Scan one or more local repository roots.
- Detect git repositories.
- Locate skill installation folders (`.claude/skills/`).
- Read `skell.toml` and `skell.lock` where present.
- Map installed skills to registry skills.
- Detect unknown local skills not present in any configured registry.

## 12.3 Status Classification
Classify each installed skill as one of:

| Status | Meaning |
|---|---|
| `up-to-date` | Matches registry version, no local modifications |
| `outdated` | Registry has a newer version |
| `pinned` | Intentionally fixed, upgrade skipped |
| `deprecated` | Registry lifecycle is `deprecated` |
| `archived` | Registry lifecycle is `archived` |
| `locally-modified` | Content hash differs from lock file baseline |
| `unknown` | Not found in any configured registry |
| `missing-metadata` | Installed skill has no `metadata.version` in frontmatter |
| `unversioned` | Registry skill has no version field |

## 12.4 Installation Actions
- Install skill into one or more target repositories.
- Update installed skill, with protection for locally-modified skills.
- Remove installed skill from one or more repositories.
- Pin and unpin installed skills.
- Preview any action before applying (`--dry-run`).

## 12.5 Diagnostics (`skell doctor`)
Detect and report:
- Malformed or invalid `SKILL.md` frontmatter.
- Skills in `skell.toml` not installed on disk.
- Skills on disk not in `skell.toml` (drift).
- Locally modified skills (hash mismatch).
- Duplicate skill directories.
- Missing or corrupt `skell.lock`.
- Path or permission issues.
- Skills referencing an unrecognized registry alias.

---

# 13. CLI Specification

## 13.1 CLI Name
`skell`

## 13.2 Command Set

### `skell list`
Lists skills in registry or local mode.

```bash
skell list                         # installed skills in current repo
skell list --source registry       # all skills in configured registry
skell list --repo ~/repos/proj-a   # installed skills in a specific repo
skell list --scan ~/repos          # installed skills across all repos under path
```

### `skell status`
Shows comparison between registry and local installs.

```bash
skell status                       # current repo
skell status --repo ~/repos/proj-a
skell status --scan ~/repos
skell status --only outdated
skell status --only locally-modified
```

### `skell info <skill-name>`
Shows full metadata for a skill: registry definition, installed state, version diff, and modification status.

```bash
skell info pdf-processing
skell info pdf-processing --source registry   # registry only
skell info pdf-processing --source local      # installed state only
```

### `skell install <skill-name>`
Installs a skill into one or more repositories.

```bash
skell install pdf-processing
skell install pdf-processing --repo ~/repos/proj-a
skell install pdf-processing --repo ~/repos/proj-a --repo ~/repos/proj-b
skell install pdf-processing --global
skell install pdf-processing --registry public
skell install pdf-processing --dry-run
```

### `skell upgrade [skill-name]`
Upgrades one or all skills where installed. Respects pinned skills.

```bash
skell upgrade                      # all upgradeable skills in current repo
skell upgrade pdf-processing
skell upgrade --repo ~/repos/proj-a
skell upgrade --all-repos ~/repos
skell upgrade pdf-processing --force   # upgrade even if locally modified
skell upgrade --dry-run
```

### `skell remove <skill-name>`
Removes a skill from one or more repositories.

```bash
skell remove sprint-legacy-helper
skell remove sprint-legacy-helper --repo ~/repos/proj-a
skell remove sprint-legacy-helper --repo ~/repos/proj-a --repo ~/repos/proj-b
skell remove sprint-legacy-helper --all-repos ~/repos
skell remove sprint-legacy-helper --dry-run
```

### `skell pin <skill-name>`
Pins a skill to its current installed version.

```bash
skell pin logging-standards
skell pin logging-standards --version 1.2.0
skell pin logging-standards --repo ~/repos/proj-a
```

### `skell unpin <skill-name>`
Removes pinning for a skill.

```bash
skell unpin logging-standards
skell unpin logging-standards --repo ~/repos/proj-a
```

### `skell sync`
Applies `skell.toml` to the current repository. Installs missing skills, removes skills no longer in the manifest. Does not upgrade pinned skills.

```bash
skell sync
skell sync --repo ~/repos/proj-a
skell sync --all-repos ~/repos
skell sync --dry-run
```

### `skell init`
Creates a `skell.toml` from the skills currently installed in a repository. Useful for migrating existing repositories.

```bash
skell init
skell init --repo ~/repos/proj-a
```

### `skell search <query>`
Searches available skills in configured registries by name, description, or tag.

```bash
skell search pdf
skell search --tag backend
skell search --lifecycle stable
skell search --owner platform-team
```

### `skell doctor`
Checks for manifest, lock file, install, and mapping problems.

```bash
skell doctor
skell doctor --repo ~/repos/proj-a
skell doctor --scan ~/repos
```

### `skell cache`
Manages the local registry cache.

```bash
skell cache status
skell cache refresh
skell cache clear
```

## 13.3 Flag Consistency Rules

All commands that operate on repositories follow the same targeting flags:

| Flag | Meaning |
|---|---|
| *(no flag)* | Current working directory |
| `--repo <path>` | Specific repository. Repeatable for multiple repos |
| `--all-repos <root>` | All git repositories found under `<root>` |
| `--global` | Operates on global manifest `~/.skell/skell.toml` |
| `--dry-run` | Preview changes without applying them |

## 13.4 CLI Behavior Rules

- Scanning of registry and local state happens implicitly as needed. There is no standalone scan command.
- All destructive commands (`remove`, `upgrade` over locally-modified) support `--dry-run`.
- `upgrade` and `install` warn and halt if a locally-modified skill would be overwritten, unless `--force` is passed.
- Output default is human-readable. Machine-readable JSON output available via `--json` flag (v0.2+).
- Error messages include the corrective action wherever possible.

## 13.5 Output Style

```
$ skell install pdf-processing
  registry  default (github.com/mycompany/skills-registry)
  resolved  pdf-processing@1.2.0
  install   .claude/skills/pdf-processing/
  lock      .claude/skell.lock updated
  done      1 skill installed

$ skell status
  skill                  installed   latest   status
  pdf-processing         1.1.0       1.2.0    outdated
  code-review            2.0.0       2.1.0    pinned
  typescript-patterns    1.3.0       1.3.0    up-to-date
  sprint-helper         ,          ,        unknown (not in registry)
  logging-standards      1.0.0       1.0.0    locally modified
```

---

# 14. File System Layout

```
my-project/                          ← project repository
├── .claude/
│   ├── skell.toml                   ← local manifest (declares what should be installed)
│   ├── skell.lock                   ← lock file (records what is installed, with hashes)
│   └── skills/
│       ├── pdf-processing/
│       │   └── SKILL.md
│       └── code-review/
│           ├── SKILL.md
│           └── references/
│               └── REFERENCE.md
└── src/

~/.skell/
├── skell.toml                       ← global manifest
├── config.toml                      ← global configuration (policy, etc.)
├── audit.log                        ← audit log (JSONL)
└── cache/
    ├── default/                     ← cached clone of default registry
    └── public/                      ← cached clone of public registry
```

---

# 15. Safety Model

Skell must be conservative around destructive operations.

| Rule | Detail |
|---|---|
| `remove` requires dry-run support | Always |
| `upgrade` warns on local modification | Halts unless `--force` is passed |
| `install` fails clearly | If target paths are invalid or conflicting |
| `sync --check` for CI | Exits non-zero if installed state differs from manifest |
| `doctor` surfaces problems early | Before user reaches destructive commands |

**Modified skill protection:** When `upgrade` or `install` would overwrite a skill whose `content_hash` differs from `skell.lock`, Skell halts with:

```
  error   pdf-processing has local modifications
          installed hash: sha256:a1b2...
          expected hash:  sha256:a1b2...
  hint    use --force to overwrite, or commit your local changes first
```

---

# 16. Enterprise Features

## 16.1 Private Registries
Any git remote is a valid registry. Authentication is handled by standard git credential helpers.

## 16.2 Registry Policy
Organizations can restrict which registries are allowed via `~/.skell/config.toml` (distributed via MDM or dotfiles):

```toml
[policy]
allowed-registries = [
  "https://github.com/mycompany/skills-registry",
  "https://github.com/agentskills/agentskills"
]
block-unlisted = true
```

When `block-unlisted = true`, Skell refuses to install from any unlisted registry.

## 16.3 Audit Log
Every install, upgrade, remove, and pin is appended to `~/.skell/audit.log` in JSONL format:

```jsonl
{"timestamp":"2026-04-12T09:14:22Z","action":"install","skill":"pdf-processing","version":"1.2.0","registry":"default","repo":"my-project","user":"amin"}
{"timestamp":"2026-04-12T09:15:01Z","action":"remove","skill":"sprint-14-helper","repo":"my-project","user":"amin"}
```

## 16.4 CI Integration

```bash
# Fails if installed skills do not match skell.toml (drift detection)
skell sync --check

# Fails if any installed skill has an available update
skell status --only outdated --fail
```

## 16.5 Onboarding
When a new engineer joins and clones a repository with a `skell.toml`:

```bash
skell sync   # installs all declared skills, writes skell.lock
```

One command. No manual file copying, no searching the registry.

---

# 17. Architecture

## 17.1 Layers

```
┌─────────────────────────────────────────┐
│  CLI interface                          │  user entry point
├─────────────────────────────────────────┤
│  Action executor                        │  install/upgrade/remove/pin + dry-run
├─────────────────────────────────────────┤
│  Comparison engine                      │  maps registry ↔ local, computes status
├─────────────────────────────────────────┤
│  Local repo scanner                     │  finds repos, reads toml/lock/skills
├─────────────────────────────────────────┤
│  Registry adapter                       │  fetches, caches, parses registry sources
├─────────────────────────────────────────┤
│  Minimal desktop UI  (v0.3+)            │  built on top of CLI core
└─────────────────────────────────────────┘
```

## 17.2 Implementation

**Core + CLI:** TypeScript.
- Fast development, strong cross-platform story, good filesystem support.
- Reusable directly with a future Tauri UI.
- Distributed as a single binary via `pkg` or `bun build`.

**Alternative:** Go, if single-binary delivery is the primary constraint.

**Desktop UI:** Tauri + React.
- Smaller footprint than Electron.
- Good fit if TypeScript is chosen for the core.

---

# 18. Minimal Desktop UI

The UI is built after the CLI is stable. It does not replace the CLI for power users. It makes discovery, review, and cleanup easier and less error-prone.

## 18.1 Screens

**Overview:** total registry skills; installed; outdated; deprecated; locally modified; unknown local.

**Registry:** searchable list of all skills with filters by owner, lifecycle, tag, scope, and updated date.

**Repositories:** all detected repositories with installed skill counts and drift warnings.

**Updates:** all outdated skills with source vs. local diff before applying.

**Cleanup:** deprecated installs; unknown installs; orphaned installs.

---

# 19. Roadmap

## Milestone 1, Core Read Operations
- Registry adapter with cache.
- Local repo scanner.
- `skell.toml` and `skell.lock` read/write.
- `skell list`, `skell status`, `skell info`, `skell search`.
- `skell doctor`.

## Milestone 2, Write Operations
- `skell install`, `skell upgrade`, `skell remove`.
- `skell pin`, `skell unpin`.
- `skell sync` and `skell init`.
- Dry-run and safety checks.
- Audit log.

## Milestone 3, Enterprise
- Policy config (`allowed-registries`, `block-unlisted`).
- Multi-registry support with conflict resolution.
- CI integration (`--check`, `--fail`).
- `--all-repos` flag across all write commands.

## Milestone 4, Desktop UI
- Tauri desktop app (Mac, Windows, Linux).
- Overview, Registry, Repositories, Updates, Cleanup screens.

---

# 20. Open Decisions

| # | Decision | Current Position |
|---|---|---|
| 1 | Lock file versioning format | JSONL per skill entry, SHA-256 hash |
| 2 | Renamed skills in registry | Warn on status; no auto-rename |
| 3 | Split/merged skills in registry | Treat as separate skills; user removes old manually |
| 4 | Skill removed from registry | Warn as `unknown`; do not auto-remove locally |
| 5 | Project-specific skills not in registry | Visible as `unknown`; supported but not managed |
| 6 | Team-level manifest above repository | Deferred to v0.4 |
| 7 | Windows path handling for `.claude/skills/` | Needs verification before Milestone 2 ships |

---

# 21. Immediate Next Steps

1. Finalize open decisions 1–4.
2. Write CLI behavior contract with edge cases (modified skill on upgrade, missing lock file, conflicting registries).
3. Create project skeleton: monorepo with `packages/core`, `packages/cli`, `packages/ui`.
4. Implement Milestone 1.
