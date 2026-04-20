# Skell

**A cross-platform skill package manager for [Agent Skills](https://agentskills.io).**

Install, upgrade, sync, and govern SKILL.md files across one or many repositories from versioned GitHub registries.

[![CI](https://github.com/aminmesbahi/skell/actions/workflows/ci.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/ci.yml)
[![CodeQL](https://github.com/aminmesbahi/skell/actions/workflows/codeql.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/codeql.yml)
[![Security](https://github.com/aminmesbahi/skell/actions/workflows/security.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/aminmesbahi/skell/branch/main/graph/badge.svg)](https://codecov.io/gh/aminmesbahi/skell)
[![Release](https://img.shields.io/github/v/release/aminmesbahi/skell)](https://github.com/aminmesbahi/skell/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/aminmesbahi/skell)](https://goreportcard.com/report/github.com/aminmesbahi/skell)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](https://github.com/aminmesbahi/skell/releases)

---

## Install

### Download binary (recommended)

Grab the latest release for your platform from [GitHub Releases](https://github.com/aminmesbahi/skell/releases):

| Platform | File |
|---|---|
| Windows (x64) | `skell_windows_amd64.exe` → rename to `skell.exe` |
| macOS (Apple Silicon) | `skell_darwin_arm64` → `chmod +x` and move to PATH |
| macOS (Intel) | `skell_darwin_amd64` |
| Linux (x64) | `skell_linux_amd64` |
| Linux (ARM64) | `skell_linux_arm64` |

### Build from source

```sh
git clone https://github.com/aminmesbahi/skell
cd skell
go build -o skell .        # Linux/macOS
go build -o skell.exe .    # Windows
```

### Self-update

```sh
skell selfupdate
```

---

## Quick Start

```sh
# 1. Initialise a repository
skell init --repo /path/to/my-repo

# 2. Install a skill (bootstraps the registry on first use)
skell install ilspy-decompile \
  --registry dotnet-skillz \
  --registry-url https://github.com/davidfowl/dotnet-skillz \
  --repo /path/to/my-repo

# 3. See what's installed
skell list --repo /path/to/my-repo

# 4. Check for updates
skell status --repo /path/to/my-repo

# 5. Upgrade all non-pinned skills
skell upgrade --repo /path/to/my-repo
```

---

## How It Works

Skills live in `.claude/skills/<name>/` inside each repository. Skell tracks them via two files:

| File | Purpose |
|---|---|
| `.claude/skell.toml` | Declares which registries and skills a repo uses |
| `.claude/skell.lock` | Records exact install state (hash, timestamp, source URL) |

Registries are plain GitHub (or any Git) repositories that contain folders with `SKILL.md` files.

---

## Command Reference

### `init`
Create `skell.toml` for a repository (scans for already-installed skills).
```sh
skell init
skell init --repo /path/to/repo
```

### `install`
Install a skill from a registry.
```sh
# Registry already in skell.toml
skell install run-tests --registry dotnet-skills

# Bootstrap a new registry in one step
skell install ilspy-decompile \
  --registry dotnet-skillz \
  --registry-url https://github.com/davidfowl/dotnet-skillz

# Dry-run (no files written)
skell install run-tests --registry dotnet-skills --dry-run

# Target a specific repo or all repos under a directory
skell install run-tests --registry dotnet-skills --repo ./my-project
skell install run-tests --registry dotnet-skills --all-repos ~/projects
```

### `list`
Show installed or available skills.
```sh
skell list                          # installed in current repo
skell list --repo /path/to/repo
skell list --all-repos ~/projects   # all managed repos under a root
skell list --source registry        # browse available skills from configured registries
skell list --json                   # JSON output (for CI)
```

### `status`
Compare local installs against the registry.
```sh
skell status
skell status --repo /path/to/repo
```

Status values: `up-to-date` · `outdated` · `pinned` · `locally-modified` · `deprecated` · `missing-metadata`

### `upgrade`
Upgrade all (or a single) non-pinned skill.
```sh
skell upgrade                       # upgrade all
skell upgrade run-tests             # upgrade one
skell upgrade --dry-run             # preview only
skell upgrade --force               # also overwrite locally-modified skills
```

### `remove`
Remove a skill from a repository.
```sh
skell remove run-tests
skell remove run-tests --repo /path/to/repo
```

### `pin` / `unpin`
Lock a skill to its current version (skips `upgrade`).
```sh
skell pin ilspy-decompile --repo /path/to/repo
skell unpin ilspy-decompile --repo /path/to/repo
```

### `sync`
Apply `skell.toml` to a repository — installs missing skills, removes unlisted ones.
```sh
skell sync
skell sync --repo /path/to/repo
skell sync --dry-run
```

### `search`
Search available skills across all configured registries.
```sh
skell search                              # list all
skell search maui                         # filter by name/description/tags
skell search --lifecycle stable
skell search --owner microsoft
skell search dotnet --lifecycle stable --owner microsoft
```

### `info`
Show full metadata for a skill.
```sh
skell info ilspy-decompile              # local install info
skell info ilspy-decompile --repo /path/to/repo
skell info ilspy-decompile --source registry   # registry lookup
skell info ilspy-decompile --json
```

### `doctor`
Check a repository for manifest/lock/install problems.
```sh
skell doctor
skell doctor --repo /path/to/repo
skell doctor --all-repos ~/projects
skell doctor --json
```

### `cache`
Manage the local registry cache (`~/.skell/cache`).
```sh
skell cache status                 # show cached registries
skell cache refresh <alias>        # re-fetch a registry
skell cache clear                  # delete the entire cache
```

### `selfupdate`
Upgrade skell itself to the latest GitHub release.
```sh
skell selfupdate
skell selfupdate --check            # check only, don't download
```

---

## Working with Multiple Repos

All commands that change files accept `--repo` (repeatable) and `--all-repos`:

```sh
# Install into two specific repos
skell install run-tests --registry dotnet-skills \
  --repo ./api \
  --repo ./worker

# Sync all repos found under ~/projects (git repos OR skell-managed folders)
skell sync --all-repos ~/projects

# Doctor check across an entire workspace
skell doctor --all-repos ~/projects
```

---

## Global Skills

Use `--global` to install skills into `~/.skell/` (available everywhere):

```sh
skell install ilspy-decompile --registry dotnet-skillz \
  --registry-url https://github.com/davidfowl/dotnet-skillz \
  --global
skell list --global
```

---

## Useful Registries

| Alias | URL | Contents |
|---|---|---|
| `dotnet-skillz` | `https://github.com/davidfowl/dotnet-skillz` | .NET decompile, C# scripts, MCP tools |
| `dotnet-skills` | `https://github.com/dotnet/skills` | Testing, MAUI, MSBuild, migrations, diagnostics |

Add them to your `skell.toml`:

```toml
[registries]
dotnet-skillz = "https://github.com/davidfowl/dotnet-skillz"
dotnet-skills  = "https://github.com/dotnet/skills"
```

Or let `skell install --registry-url` add them automatically on first use.

---

## JSON Output

Every command supports `--json` for scripting and CI:

```sh
skell list --json
skell status --json
skell doctor --json
skell search maui --json
```

---

## File Layout

```
<repo>/
└── .claude/
    ├── skell.toml       ← manifest (commit this)
    ├── skell.lock       ← lock file (commit this)
    └── skills/
        ├── ilspy-decompile/
        │   └── SKILL.md
        └── run-tests/
            └── SKILL.md
```

---

## Further Reading

- [System Design Document](docs/design.md) — architecture, product vision, data model
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
