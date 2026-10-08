## skell upgrade

Upgrade one or all skills

### Synopsis

Fetches the latest version of skills from the registry and replaces the local
copies. Pinned skills are skipped unless --force is specified.

```
skell upgrade [skill-name] [flags]
```

### Examples

```
  # Upgrade all skills in the current repo
  skell upgrade

  # Upgrade a single skill
  skell upgrade pdf-processing

  # Preview upgrades without applying
  skell upgrade --dry-run

  # Upgrade even pinned skills
  skell upgrade --force

  # Upgrade across multiple repos
  skell upgrade --repo ./api --repo ./worker
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --force              Overwrite locally-modified skills
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for upgrade
      --json               Output results as JSON
      --no-validate        Skip spec validation even if policy requires it
      --repo stringArray   Target repository path (repeatable)
      --target string      Agent platform to manage
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
      --validate           Validate the skill against the spec before writing (fail on errors)
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

