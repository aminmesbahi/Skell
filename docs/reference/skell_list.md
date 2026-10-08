## skell list

List installed or registry skills

### Synopsis

Lists skills for one or more repositories.

By default shows skills installed locally (from skell.lock).
Use --source registry to browse all skills available in the configured registries.

```
skell list [flags]
```

### Examples

```
  # List skills installed in the current repo
  skell list

  # List all skills available in configured registries
  skell list --source registry

  # List installed skills as JSON
  skell list --json

  # List skills across every git repo under a root directory
  skell list --all-repos /home/user/projects

  # List skills for a specific agent platform
  skell list --target cursor
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for list
      --json               Output results as JSON
      --repo stringArray   Target repository path (repeatable)
      --source string      Source to list from: local | registry (default "local")
      --target string      Agent platform to list skills for: claude | codex | copilot | cursor | windsurf | opencode | cline | grok
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

