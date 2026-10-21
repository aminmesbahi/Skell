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
  moves a skill to a newer commit. Sources that are plain local folders have no
  commit, so they are copied as they are on disk and are not locked to a
  revision (`skell review` says so).
- **Drift detection.** `skell sync` also compares every installed skill with
  the locked `content_hash`. A skill that was edited by hand is reported as
  *modified*: `sync --check` fails on it (so CI catches tampering) and a plain
  `sync` leaves it alone. `skell sync --force` discards the edits and restores
  the locked content.
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
| `locally-modified` | you edited the installed copy | keep it, or `skell upgrade <skill> --force` to restore the source version |
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
max-risk           = "medium"  # refuse skills whose security scan is above this level
require-signed     = true      # only install skills signed by a trusted publisher
```

URLs are compared ignoring case, trailing `/` and `.git`. If the file exists but
can't be parsed, **every** source is refused (fail closed) so a typo can't
silently disable the policy.

Every install, upgrade, remove, pin and sync is appended to
`~/.skell/audit.log` (JSON lines, with the acting user).

## Security scan

`skell scan` reads every file of a skill and looks for what makes a skill
dangerous. Each finding has a severity (info, low, medium, high, critical); the
skill gets a score from 0 to 100 and a **risk level**: `clean`, `low`,
`medium`, `high` or `critical`.

| Looks for | Example |
|---|---|
| Prompt injection and concealment | "ignore all previous instructions", "do not tell the user" |
| Remote code execution | `curl … \| sh`, `iwr … \| iex`, download-then-run |
| Obfuscation | base64 decoded into a shell, long encoded blobs, `eval` |
| Credentials and exfiltration | `~/.ssh`, `.aws/credentials`, posting secrets with `curl` |
| Hidden text | zero-width and right-to-left control characters |
| Persistence and destruction | cron, shell profiles, `rm -rf /`, disabling security tools |
| Binaries and broad tools | `.exe`/`.so`, `allowed-tools: "*"` |

It is a **heuristic**: it shows what deserves a human look and cannot prove a
skill safe. A skill that documents these patterns (a security guide) will be
flagged too. The scan runs automatically in `skell review`, `skell install` and
`skell upgrade`; a medium-or-worse result needs the same confirmation as a skill
that ships scripts.

## Signatures and trusted publishers

A publisher signs a skill folder with an Ed25519 key; consumers choose which
publishers they trust.

```sh
# Publisher
skell keygen acme                                   # acme.key (secret) + acme.pub (share)
skell sign ./team-skills --key ~/.skell/keys/acme.key --publisher acme

# Consumer
skell trust add acme acme.pub
skell verify                                        # trusted | untrusted-key | tampered | invalid | unsigned
```

The signature (`SKILL.sig`) covers every file in the skill, so any later edit
makes `verify` report `tampered`. Re-sign after each change. The list of trusted
publishers is `trusted-keys.json` in the Skell home, so a company can ship it
together with `config.toml`.

## Dependencies and bundles

A skill can require other skills in its frontmatter:

```yaml
metadata:
  requires: "pdf, docx>=1.2.0"
```

Constraints are `>=`, `>`, `<=`, `<`, `=`, `^` (same major) and `~` (same minor).
`skell install` installs the requirements too, refuses a combination no source
can satisfy, and `skell doctor` reports requirements that are missing or too
old, and loops.

A **bundle** is a named list of skills in `skell.toml`:

```toml
[bundles.frontend]
description = "Everything a UI developer needs"
skills = ["react", "css-review", "a11y-check"]
```

`skell bundle install frontend` (or `skell install @frontend`) installs what is
missing. Bundles travel with a team baseline (`skell apply`).

## Fleet report

`skell report` summarises many repositories in one view: which skills run in
which versions (and which are on mixed versions), hand-edited and outdated
installs, the risk and signature of every install, repositories that need
attention, and recent activity from the audit log. It reads local files only
unless you pass `--check-updates`. Output formats: text, `json`, `md`, `csv` and
a self-contained `html` page.

## Skill tests

A skill can ship example prompts in `evals/evals.yaml`:

```yaml
cases:
  - name: summarises a report
    prompt: Summarise report.pdf in three bullets.
    timeout: 180
    expect:
      contains: ["summary"]
      not_contains: ["I cannot"]
      matches: ["(?m)^- "]
```

`skell test` installs the skill into a scratch workspace, gives each prompt to
your agent (`--agent "claude -p"` or `$SKELL_TEST_AGENT`; any command that reads
a prompt on stdin, or takes it where you write `{prompt}`), and checks the
answer. Every case is a real agent call, so there is no default agent. Without
one only the file is checked. `skell upgrade --verify` runs the new version's
tests first and keeps the old version when it breaks cases that used to pass.
`skell new my-skill --evals` writes a starter file.

## Environment variables

| Variable | Effect |
|---|---|
| `SKELL_HOME` | Skell's data folder (default `~/.skell`) |
| `SKELL_OFFLINE` | never download the catalog |
| `SKELL_CATALOG_URL` | use your own catalog file (same format as `internal/catalog/index.json`) |
| `NO_COLOR` | disable colored diffs |
| `SKELL_TEST_AGENT` | the command `skell test` uses to answer prompts, e.g. `claude -p` |
