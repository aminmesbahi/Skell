# Concepts

## Skills and sources

A **skill** is a folder with a `SKILL.md` file (YAML frontmatter + instructions)
and optional supporting files, following the [Agent Skills](https://agentskills.io)
format. Agents such as Claude Code, Codex, Copilot and Cursor load skills from a
folder in your project or home directory.

A **source** (called a *registry* in some flags and files) is a git repository or
a local folder that contains skill folders anywhere inside it. Add one with
`skell add`:

| You type | Skell uses |
|---|---|
| `anthropic` | the catalog entry with that id (`skell catalog`) |
| `owner/repo` | `https://github.com/owner/repo` |
| `owner/repo/path/to/skill` | that one skill, installed directly |
| any git URL | as given (`https://…`, `git@…`) |
| `./folder`, `/abs/path`, `~/x` | a local folder, used in place |

Remote sources are cloned once into `~/.skell/cache/<alias>` and refreshed with
`skell cache refresh` (or `--refresh` on `status` / `diff`).

## Manifest and lock file

Each project has two files in its agent folder (e.g. `.claude/`). Commit both.

**`skell.toml`** — what you want:

```toml
target  = "claude"
mirrors = ["copilot", "cursor"]      # optional, see below

[registries]
anthropic = "https://github.com/anthropics/skills"
team      = "https://github.com/acme/skills"

[skills]
pdf          = { registry = "anthropic" }
release-notes = { registry = "team", pinned = true, version = "1.2.0" }
```

**`skell.lock`** — what is installed, exactly:

```json
{
  "name": "pdf",
  "registry": "anthropic",
  "commit": "683bc88e56f3e09ba94f7055977f3d3aa499f202",
  "source_path": "skills/pdf",
  "content_hash": "sha256:6a01…",
  "installed_path": ".claude/skills/pdf"
}
```

The lock makes three things possible:

- **Reproducible installs.** `skell sync` reinstalls a missing skill at its locked
  `commit` and checks the result against `content_hash`, so every machine and
  CI run gets byte-identical files. `skell upgrade` is the only command that
  moves a skill to a newer commit.
- **Change detection.** If the source's files differ from the locked hash, the
  skill is *outdated* — even when the author didn't bump a version (most skills
  have none). If your installed copy differs, it is *locally modified*, and
  `upgrade` won't overwrite it without `--force`.
- **Reviewable updates.** `skell diff <skill>` shows installed → latest.

## Status values

| Status | Meaning | What to do |
|---|---|---|
| `up-to-date` | matches the source | nothing |
| `outdated` | newer version, or the source content changed | `skell diff`, then `skell upgrade` |
| `locally-modified` | you edited the installed copy | keep it, or `skell upgrade --force` to discard |
| `pinned` | held at a version on purpose | `skell unpin` to resume updates |
| `linked` | a live link to a folder you're developing | nothing — edits are live |
| `deprecated` / `archived` | the author retired it | look for a replacement |
| `unknown` | source unreachable or not configured | `skell doctor` |

## Targets and mirrors

A **target** is an agent's folder layout:

| Target | Folder |
|---|---|
| `claude` | `.claude/skills/` |
| `codex` | `.codex/skills/` |
| `copilot` | `.github/skills/` |
| `cursor` | `.cursor/skills/` |
| `windsurf`, `opencode`, `cline`, `grok` | `.<name>/skills/` |

The manifest lives in its target's folder. If you use several agents in one repo,
add **mirrors** (`skell mirror add copilot cursor`): every skill is copied into
each mirror's folder, and install / upgrade / remove / sync keep the copies in
step. Mirrors are derived files; `skell doctor` reports drift and
`skell doctor --fix` (or `sync`) repairs it.

## Scopes

| Scope | Where skills go | How |
|---|---|---|
| Project | `<repo>/.claude/skills/` | default (current directory or `--repo`) |
| Many projects | each repo | `--repo a --repo b` or `--all-repos ~/work` |
| Personal (user) | `~/.claude/skills/`, `~/.codex/skills/`, … | `--user` |
| Global manifest | `~/.skell/.claude/` (search-only sources) | `--global` |

## Linked skills

`skell link <folder>` (or `skell new <name> --link`) symlinks a skill you are
writing into the project — a directory junction on Windows when symlinks need
admin rights. The link is recorded as `linked` in the lock, never pruned or
upgraded, and added to `.git/info/exclude` so it isn't committed.

## Policy and audit

Organisation-wide settings live in `~/.skell/config.toml` (or
`$SKELL_HOME/config.toml`):

```toml
[sources]                       # available to every project
shared = "https://github.com/acme/skills"

[policy]
allowed-registries = ["https://github.com/acme/skills"]
block-unlisted     = true       # refuse any other source (install, search, sync…)
require-validation = true       # validate against the spec before install/upgrade
```

URLs are compared ignoring case, trailing `/` and `.git`. If the file exists but
can't be parsed, **every** source is refused (fail closed) so a typo can't
silently disable the policy.

Every install, upgrade, remove, pin and sync is appended to
`~/.skell/audit.log` (JSON lines, with the acting user).

## Environment variables

| Variable | Effect |
|---|---|
| `SKELL_HOME` | Skell's data folder (default `~/.skell`) |
| `SKELL_OFFLINE` | never download the catalog |
| `SKELL_CATALOG_URL` | use your own catalog file (same format as `internal/catalog/index.json`) |
| `NO_COLOR` | disable colored diffs |
