# Skell

Skell is a friendly cross-platform tool for managing Agent Skills.

It lets you install, update, sync, and keep your SKILL.md files organized across projects. Skills can come from GitHub repositories or local folders on your computer.

[![CI](https://github.com/aminmesbahi/skell/actions/workflows/ci.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/ci.yml)
[![CodeQL](https://github.com/aminmesbahi/skell/actions/workflows/codeql.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/codeql.yml)
[![Security](https://github.com/aminmesbahi/skell/actions/workflows/security.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/aminmesbahi/skell/branch/main/graph/badge.svg)](https://codecov.io/gh/aminmesbahi/skell)
[![Release](https://img.shields.io/github/v/release/aminmesbahi/skell)](https://github.com/aminmesbahi/skell/releases)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/aminmesbahi/skell)](https://goreportcard.com/report/github.com/aminmesbahi/skell)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](https://github.com/aminmesbahi/skell/releases)

---

## What's Included

| Component | Description |
|---|---|
| **CLI** (`skell`) | Cross-platform command-line tool for Windows, macOS, and Linux |
| **Desktop GUI** (`Skell.exe`) | Native desktop app built with Wails (Windows primary, macOS and Linux also supported) |

---

## Install

### CLI on Windows

```powershell
irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex
```

This installs both `skell.exe` and the desktop GUI (`skell-gui.exe`) into your user programs directory, so you can launch the app with:

```powershell
skell gui
```

Or with [winget](https://github.com/microsoft/winget-cli) (once the package is published):
```powershell
winget install aminmesbahi.skell
```

### CLI on macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh
```

Or with [Homebrew](https://brew.sh):
```sh
brew tap aminmesbahi/tap
brew install skell
```

### Manual Download for CLI

Grab the latest binary for your platform from [GitHub Releases](https://github.com/aminmesbahi/skell/releases):

| Platform | File |
|---|---|
| Windows (x64, CLI + GUI bundle) | `skell_0.x.x_windows_amd64_bundle.zip` |
| Windows (x64, CLI only) | `skell_0.x.x_windows_amd64.zip` |
| macOS (Apple Silicon) | `skell_0.x.x_darwin_arm64.tar.gz` |
| macOS (Intel) | `skell_0.x.x_darwin_amd64.tar.gz` |
| Linux (x64) | `skell_0.x.x_linux_amd64.tar.gz` |
| Linux (ARM64) | `skell_0.x.x_linux_arm64.tar.gz` |

Extract the archive and place the `skell` binary somewhere on your `PATH`. On Windows, the bundle archive also includes `skell-gui.exe`, which `skell gui` will launch when both files live in the same directory.

### Desktop GUI - Download

Download either the Windows bundle (`skell_0.x.x_windows_amd64_bundle.zip`) or the standalone GUI binary (`Skell-windows-amd64.exe`) from [GitHub Releases](https://github.com/aminmesbahi/skell/releases).
If you use the standalone GUI binary, keep the `skell` CLI installed and on `PATH`.

### Self-Update the CLI

```sh
skell selfupdate
```

### Launch the Desktop GUI

```sh
skell gui
```

On Windows, this looks for a GUI executable next to the `skell` binary.

---

## Uninstall

### CLI on Windows

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1))) -Uninstall
```

This removes the `skell.exe` binary and cleans the install directory from your user `PATH`.
If present, it also removes `skell-gui.exe`.

### CLI on macOS and Linux

```sh
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh -s -- uninstall
```

By default this removes `/usr/local/bin/skell`. Set `INSTALL_DIR` to match a custom location:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh -s -- uninstall
```

### Manual Removal of the CLI

Delete the `skell` (or `skell.exe`) binary from wherever you placed it and remove that directory from your `PATH` if it was added solely for skell.

---

## Desktop GUI

There's also a native desktop app built with Wails. It wraps the CLI so you can browse skills, install them, sync projects, and manage everything visually without typing commands all the time.

Key things you can do in the GUI:
- See all your projects and what skills they have
- Browse and search the skill registry
- Install, upgrade, pin, or remove skills with a couple of clicks
- Run sync and doctor checks with nice previews
- Contribute improvements back to skill authors

**Important:** The GUI needs the `skell` CLI on your PATH to work. Install the CLI first.

### Run GUI in development mode

```sh
cd gui
wails dev
```

### Build GUI for production

```sh
cd gui
wails build
# Output: gui/build/bin/Skell.exe (Windows)
```

---

## Quick Start

Here's the typical flow:

```sh
# Create a skell.toml for your project
skell init --repo /path/to/my-repo

# Install a skill (this can also register a new source on the fly)
skell install ilspy-decompile \
  --registry dotnet-skillz \
  --registry-url https://github.com/davidfowl/dotnet-skillz \
  --repo /path/to/my-repo

# See what you have
skell list --repo /path/to/my-repo

# Check if anything is outdated
skell status --repo /path/to/my-repo

# Upgrade everything that's not pinned
skell upgrade --repo /path/to/my-repo
```

---

## How It Works

Skell understands all the common ways different AI tools store skills. When you run `skell init`, it tries to detect which layout your project already uses (or you can pick one with `--target`).

Supported layouts:

- `claude` (default): `.claude/skills/`
- `codex`: `.codex/skills/`
- `copilot`: `.github/skills/`
- `cursor`: `.cursor/skills/`

The actual `SKILL.md` file inside each skill folder is the same no matter which layout you use.

Skell keeps two files in the target directory to stay organized:

- `skell.toml` - your source of truth. It lists which sources and skills you want.
- `skell.lock` - records exactly what is installed right now (with content hashes so we can detect local changes).

You can see all supported targets with `skell targets`.

A "registry" in Skell is just a Git repo (or a local folder) that contains one or more skill directories, each with its own `SKILL.md`. We support both remote Git sources and local folders as first-class citizens.

---

## Command Reference

### `init`
Create `skell.toml` for a repository (scans for already-installed skills).
If the repo already has a known agent folder (`.claude`, `.codex`, `.github`,
`.cursor`), Skell uses it automatically; otherwise pass `--target` to choose,
or run interactively to be prompted.
```sh
skell init                            # auto-detect or prompt
skell init --target copilot           # force VS Code / GitHub Copilot layout
skell init --target cursor --repo /path/to/repo
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
Apply `skell.toml` to a repository - installs missing skills, removes unlisted ones.
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

## Build from Source

### CLI

```sh
git clone https://github.com/aminmesbahi/skell
cd skell
go build -o skell .        # Linux/macOS
go build -o skell.exe .    # Windows
```

Or build all platforms at once:

```sh
./build-all.sh v0.1.0      # Linux/macOS
.\build-all.ps1 -Version v0.1.0  # Windows PowerShell
```

### Desktop GUI

Prerequisites: [Go 1.22+](https://go.dev), [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation), [Bun](https://bun.sh)

```sh
cd gui
wails build              # production build → gui/build/bin/Skell.exe
wails dev                # live-reload dev mode
```

---

## Further Reading

- System Design Document (docs/design.md) - architecture and data model
- Changelog
- Contributing guide

Thanks for using Skell. If you run into anything weird or have ideas, feel free to open an issue. We're always happy to improve it.
