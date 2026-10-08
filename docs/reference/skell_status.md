## skell status

Show comparison between registry and local installs

### Synopsis

Compares every installed skill against its source to show whether each is
up-to-date, outdated, pinned, linked, locally-modified, or has missing metadata.

A skill is "outdated" when its source has a newer version — or when the source
files changed even though the version did not (many skills have no version).
Status compares against the locally cached copy of each source; pass
--refresh to fetch the latest first.

```
skell status [flags]
```

### Examples

```
  # Check status of all skills in the current repo
  skell status

  # Fetch the latest from every source first
  skell status --refresh

  # Show only outdated skills
  skell status --only outdated

  # Status across multiple repos
  skell status --repo ./service-a --repo ./service-b

  # Machine-readable JSON output
  skell status --json
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for status
      --json               Output results as JSON
      --only string        Filter by status (e.g. outdated, locally-modified)
      --refresh            Fetch the latest from every source before comparing
      --repo stringArray   Target repository path (repeatable)
      --target string      Agent platform to check status for: claude | codex | copilot | cursor | windsurf | opencode | cline | grok
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

