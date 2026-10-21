# Recipes

## Share skills with your team

```sh
skell init
skell add acme/skills            # your team's source
skell install code-review release-notes
git add .claude/skell.toml .claude/skell.lock .claude/skills
git commit -m "Add team skills"
```

Teammates pull and run `skell sync` — they get exactly the locked commits, verified
by hash. Whether you also commit `.claude/skills/` is up to you: committing it
means agents work right after `git clone`; leaving it out (gitignored) keeps the
repo small and `skell sync` restores it.

## Use one set of skills with Claude, Copilot and Cursor

```sh
skell init --target claude --mirror copilot,cursor
# or, in an existing project:
skell mirror add copilot cursor
```

## Keep skills up to date

```sh
skell status --refresh           # what's outdated?
skell diff code-review           # what exactly changed?
skell upgrade code-review        # take it
skell pin release-notes          # hold one back
```

### Weekly upgrade pull requests

```yaml
# .github/workflows/skills-upgrade.yml
name: Upgrade skills
on:
  schedule: [{ cron: "0 6 * * 1" }]
  workflow_dispatch:
permissions:
  contents: write
  pull-requests: write
jobs:
  upgrade:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aminmesbahi/skell@v0.13.0   # pin a release tag
        with:
          command: upgrade
      - uses: peter-evans/create-pull-request@v7
        with:
          title: "chore: upgrade agent skills"
          branch: skell/upgrade-skills
          commit-message: "chore: upgrade agent skills"
```

## Check skills in CI

```yaml
# .github/workflows/skills.yml
name: Skills
on: [pull_request]
jobs:
  skills:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aminmesbahi/skell@v0.13.0   # pin a release tag
        with:
          command: sync --check          # fails if installed skills drift from skell.toml or skell.lock (including hand edits)
      - uses: aminmesbahi/skell@v0.13.0
        with:
          command: validate              # fails on spec errors (add --strict to fail on warnings)
```

The action accepts `command`, `version` and `working-directory` inputs. If your team
customises installed skills on purpose, use `sync --check --allow-modified`.

## Roll out a team baseline

Publish a `baseline.toml` (an ordinary `skell.toml`) in a team repo, then:

```sh
skell apply https://raw.githubusercontent.com/acme/skills/main/baseline.toml --all-repos ~/work
```

Missing sources, skills and mirrors are added and installed; each project's own
choices are kept. A baseline that points an existing alias at a different URL is
rejected.

## Lock down sources for your organisation

Distribute `~/.skell/config.toml` (or set `SKELL_HOME` to a managed folder):

```toml
[policy]
allowed-registries = ["https://github.com/acme/skills", "https://github.com/anthropics/skills"]
block-unlisted     = true
require-validation = true
```

Point `SKELL_CATALOG_URL` at your own catalog so `skell catalog` lists only
approved sources.

## Private sources

Any URL your `git` can clone works — Skell runs your git with your credentials
(SSH keys, credential helper). It never prompts for passwords.

```sh
skell add git@github.com:acme/private-skills.git --alias acme
```

## Personal skills for every project

```sh
skell init --user                # ~/.claude/skell.toml
skell add anthropic --user
skell install pdf --user         # → ~/.claude/skills/pdf
```

## Write a skill

```sh
skell new release-notes --description "Write release notes from merged PRs. Use when preparing a release." --link
# edit ./release-notes/SKILL.md — the agent sees changes immediately
skell validate --path release-notes --full
```

## Start a team skill repository

```sh
skell new team-skills --registry
cd team-skills && git init && git add . && git commit -m "Skill repository"
```

You get a README, an example skill and a CI workflow that validates every skill
on pull requests. Teammates add it with `skell add <owner>/team-skills`.

## Let your agent manage skills

```sh
claude mcp add skell -- skell mcp            # Claude Code
# or install the plugin (MCP server + usage skill):
claude plugin marketplace add aminmesbahi/skell
claude plugin install skell@skell
```

Ask: *"Find a skill for filling PDF forms and install it."* The agent searches,
shows you the review, and installs after you confirm.

## Check what a skill does before you trust it

```sh
skell review pdf             # files, scripts, tools, links — and the security scan
skell scan                   # re-scan everything installed
skell scan --path ./skills   # scan a folder of skills, e.g. before you publish it
skell scan --fail-on medium  # for CI: exit non-zero at medium risk or worse
```

## Publish signed skills, trust signed skills

```sh
skell keygen acme
skell sign ./team-skills --key ~/.skell/keys/acme.key --publisher acme
git add . && git commit -m "Sign skills" && git push

# Everyone who consumes them
skell trust add acme acme.pub        # the public key you published in your README
skell verify                         # trusted · untrusted-key · tampered · invalid · unsigned
```

To only allow signed skills on a managed machine, put this in `config.toml` and
distribute `trusted-keys.json` next to it:

```toml
[policy]
require-signed = true
max-risk       = "medium"
```

## Starter packs for a role

```sh
skell bundle create frontend react css-review a11y-check --description "UI developer pack"
git add .claude/skell.toml && git commit -m "Add frontend bundle"

skell bundle install frontend        # or: skell install @frontend
```

A skill that needs another one says so in its frontmatter
(`metadata: requires: "pdf, docx>=1.2.0"`) and `skell install` brings it along.

## A weekly view of every repository

```sh
skell report --all-repos ~/work --format html --out fleet.html
skell report --all-repos ~/work --format csv  --out fleet.csv     # for a spreadsheet
skell report --all-repos ~/work --check-updates                    # also ask the sources
```

```yaml
# .github/workflows/skills-report.yml — publish the page from a scheduled job
on:
  schedule: [{ cron: "0 6 * * 1" }]
jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aminmesbahi/skell@v0.13.0
        with:
          command: report --repo . --format md --out skills-report.md
      - uses: actions/upload-artifact@v4
        with: { name: skills-report, path: skills-report.md }
```

`--webhook https://…` also posts the JSON report to a dashboard (https only).

## Test a skill, and upgrade only when it still works

```sh
skell new release-notes --evals                      # writes evals/evals.yaml
skell test --path release-notes --agent "claude -p"  # or set SKELL_TEST_AGENT
skell upgrade --verify --agent "claude -p"           # keeps a skill whose new version breaks its tests
```
