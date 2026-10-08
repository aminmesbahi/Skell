# Skell

**A package manager for Agent Skills.** Find, install, update and share `SKILL.md`
skills for Claude Code, Codex, GitHub Copilot, Cursor and other agents — per
project, per user, or across a whole team.

[![CI](https://github.com/aminmesbahi/skell/actions/workflows/ci.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/ci.yml)
[![CodeQL](https://github.com/aminmesbahi/skell/actions/workflows/codeql.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/codeql.yml)
[![Security](https://github.com/aminmesbahi/skell/actions/workflows/security.yml/badge.svg)](https://github.com/aminmesbahi/skell/actions/workflows/security.yml)
[![codecov](https://codecov.io/gh/aminmesbahi/skell/branch/main/graph/badge.svg)](https://codecov.io/gh/aminmesbahi/skell)
[![Release](https://img.shields.io/github/v/release/aminmesbahi/skell)](https://github.com/aminmesbahi/skell/releases)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](https://github.com/aminmesbahi/skell/releases)

---

## Install

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/aminmesbahi/skell/main/install.sh | sh
```

```powershell
# Windows (also installs the desktop app)
irm https://raw.githubusercontent.com/aminmesbahi/skell/main/install.ps1 | iex
```

Both installers verify the download against the release's SHA-256 checksums.
Other options (Homebrew, `go install`, manual download, uninstall): see
[docs/install.md](docs/install.md).

## Get started in two minutes

```sh
cd my-project
skell init                 # pick the agents you use (Claude, Copilot, Cursor…)
skell catalog              # well-known skill sources
skell add anthropic        # add one: catalog id, owner/repo, URL or a local folder
skell search pdf           # find skills
skell install pdf          # install — you'll see a review if it ships scripts
```

Then, any time:

```sh
skell status --refresh     # anything outdated?
skell diff pdf             # what changed upstream?
skell upgrade              # take the updates
```

Run plain `skell` for a summary of the current project and what to do next.

## What you can do

| I want to… | Command |
|---|---|
| Add a source of skills | `skell add anthropic` · `skell add owner/repo` · `skell add ./my-skills` |
| Install one skill straight from GitHub | `skell add anthropics/skills/skills/pdf` |
| Install several skills | `skell install pdf docx xlsx` |
| See what a skill contains before installing | `skell review pdf` |
| See what an update changes | `skell diff <skill> --refresh` |
| Give my teammates the same skills | commit `.claude/skell.toml` + `skell.lock`; they run `skell sync` |
| Use the same skills in Claude, Copilot **and** Cursor | `skell mirror add copilot cursor` |
| Keep personal skills for every project | `skell init --user`, then `skell install <skill> --user` |
| Roll out a team baseline | `skell apply https://…/baseline.toml --all-repos ~/work` |
| Check skills in CI | `uses: aminmesbahi/skell@main` (runs `skell sync --check`) |
| Let my agent find and install skills | `claude mcp add skell -- skell mcp` |
| Write my own skill | `skell new my-skill --link` · `skell validate --path my-skill` |
| Fix a broken setup | `skell doctor --fix` |

## How it works, briefly

- A **source** is a git repository (or local folder) containing skill folders,
  each with a `SKILL.md`.
- `skell.toml` (the **manifest**) lists your sources and the skills you want.
  `skell.lock` records exactly what was installed — the source **commit** and a
  content hash — so `skell sync` reproduces byte-identical skills on every
  machine and in CI, and Skell can tell when you or the source changed a skill.
- Skills are installed into your agent's folder (`.claude/skills/`,
  `.github/skills/`, `.cursor/skills/`, …), optionally **mirrored** to several
  agents at once.

More in [docs/concepts.md](docs/concepts.md).

## Trust and safety

Skills are instructions your agent follows, and some ship scripts. Skell:

- shows a **review** (files, scripts, pre-approved tools, external links) and
  asks before installing anything that ships scripts or broad tool access;
- pins installs to a **commit** and verifies content hashes on `sync`;
- rejects symlinks that escape a skill folder;
- supports an organisation **policy** (allowed sources, required validation)
  and an audit log — see [docs/concepts.md](docs/concepts.md#policy-and-audit).

## Desktop app

`skell gui` opens the desktop app (Windows and macOS): browse and install skills,
see every project's health, review upgrades, and contribute metadata fixes back
to skill authors. It uses the `skell` CLI under the hood.

## Use Skell from your agent

```sh
# Claude Code: MCP server + a skill that teaches Claude how to use Skell
claude plugin marketplace add aminmesbahi/skell
claude plugin install skell@skell

# Any MCP client
command: skell   args: ["mcp"]
```

The agent can then search, review, install and upgrade skills itself. Installs of
skills with scripts or broad tool access require your confirmation.

## Documentation

- [Install & uninstall](docs/install.md)
- [Concepts](docs/concepts.md) — sources, manifest & lock, targets & mirrors, policy
- [Recipes](docs/recipes.md) — teams, CI, private sources, personal skills, authoring
- [Command reference](docs/reference/skell.md) — generated from `--help`
- [Design document](docs/design.md)

## Build from source

```sh
git clone https://github.com/aminmesbahi/skell && cd skell
go build -o skell .            # CLI
cd gui && wails3 build         # desktop app (needs Wails v3 + bun, 64-bit Go)
```

Regenerate the command reference after changing commands:
`go run . gen-docs docs/reference`.

Issues and ideas are welcome — please open an issue.
