---
name: skell
description: Find, review, install and update Agent Skills (SKILL.md) for this project with Skell. Use when the user asks for a skill for some task, wants to add or update skills, share skills with their team, or check whether installed skills are outdated.
---

# Managing skills with Skell

Skell is a package manager for Agent Skills. Prefer the `skell_*` MCP tools when
they are available; otherwise run the `skell` CLI.

## Finding and installing a skill

1. Search configured sources: `skell_search` (CLI: `skell search <keyword>`).
2. If nothing is configured or nothing matches, list well-known sources with
   `skell_catalog` (CLI: `skell catalog`) and add one with `skell_add_source`
   (CLI: `skell add <id-or-owner/repo>`), then search again.
3. Inspect the skill with `skell_review` before installing. Tell the user what it
   contains — especially scripts and pre-approved tools.
4. Install with `skell_install`. If it returns `needs_confirmation`, show the
   review to the user and only retry with `confirm=true` after they agree.
5. Tell the user the new skill is available after the agent reloads skills.

## Keeping skills current

- `skell_status` (CLI: `skell status --refresh`) shows outdated or locally
  modified skills.
- `skell_diff` (CLI: `skell diff <name>`) shows what would change. Summarise the
  diff for the user before running `skell_upgrade`.
- Never pass `--force` to upgrade without the user's agreement: it discards
  their local edits.

## Teams

- `skell.toml` and `skell.lock` belong in version control. Teammates run
  `skell sync` to get exactly the same skills (pinned to the locked commits).
- To serve several agents in one repo, suggest `skell mirror add copilot cursor`.
- For a team-wide baseline, suggest `skell apply <url-to-baseline-skell.toml>`.

## Writing skills

- `skell new <name> --link` scaffolds a skill and links it into the project so
  edits are live; `skell validate --path <dir>` checks it against the spec.
