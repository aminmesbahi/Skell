## skell apply

Adopt a shared team baseline (sources + skills) in one step

### Synopsis

Merges a shared skell.toml — your team's or organisation's standard set of
skill sources, skills and mirror targets — into each project, then syncs.

<baseline> is a path to a skell.toml or an https:// URL to one (for example a
raw file in a team repository). Entries the project already has are kept; a
source alias that points somewhere else in the baseline is an error, so a
baseline can never silently swap where a project's skills come from.

```
skell apply <baseline> [flags]
```

### Examples

```
  # Adopt the team baseline
  skell apply https://raw.githubusercontent.com/acme/skills/main/baseline.toml

  # Preview, then apply to every repo under ~/work
  skell apply ./baseline.toml --dry-run
  skell apply ./baseline.toml --all-repos ~/work
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for apply
      --json               Output results as JSON
      --no-sync            Only update skell.toml; don't install
      --repo stringArray   Target repository path (repeatable)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

