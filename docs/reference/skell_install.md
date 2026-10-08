## skell install

Install one or more skills into one or more repositories

### Synopsis

Fetches skills from your configured sources and installs them into the
repository. Updates skell.toml and skell.lock (which records the exact source
commit, so 'skell sync' reproduces the same files everywhere).

Without --registry, Skell finds the source that provides each skill; if
several do, you are asked to pick one with --registry. Use --registry-url to
add and use a new source in one step.

Before installing a skill that ships scripts or pre-approves broad tool access,
Skell shows what it contains and asks for confirmation (skip with --yes; CI and
other non-interactive runs proceed automatically). --dry-run always shows the
review without installing.

```
skell install <skill-name>... [flags]
```

### Examples

```
  # Install from whichever configured source has it
  skell install pdf

  # Several at once
  skell install pdf docx xlsx

  # From a specific source
  skell install run-tests --registry dotnet-skills

  # Add a new source and install in one step
  skell install ilspy-decompile \
    --registry dotnet-skillz \
    --registry-url https://github.com/davidfowl/dotnet-skillz

  # Preview (shows files, scripts and tool permissions)
  skell install pdf --dry-run

  # Install for a specific agent platform (auto-inits if needed)
  skell install pdf --target copilot
```

### Options

```
      --all-repos string      Scan all git repos under this root path
      --dry-run               Preview changes without applying them
      --global                Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help                  help for install
      --json                  Output results as JSON
      --no-validate           Skip spec validation even if policy requires it
      --registry string       Source (registry alias) to install from; found automatically when omitted
      --registry-url string   URL for the registry alias (auto-adds it to skell.toml if not present)
      --repo stringArray      Target repository path (repeatable)
      --target string         Agent platform to install for: claude | codex | copilot | cursor | windsurf | opencode | cline | grok
      --user                  Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
      --validate              Validate the skill against the spec before writing (fail on errors)
  -y, --yes                   Don't ask for confirmation before installing
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

