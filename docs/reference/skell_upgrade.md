## skell upgrade

Upgrade one or all skills

### Synopsis

Fetches the latest version of skills from the registry and replaces the local
copies. Pinned skills are skipped unless --force is specified.

A skill whose update adds scripts or pre-approves broad tool access that the
installed copy did not have is shown to you and needs confirmation (skip with
--yes; CI and other non-interactive runs proceed automatically, so review the
resulting pull request). --force also restores a locally edited skill whose
source has not changed.

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

  # Upgrade even pinned skills, and discard local edits
  skell upgrade --force

  # Don't ask before taking updates that add scripts or tool permissions
  skell upgrade --yes

  # Upgrade across multiple repos
  skell upgrade --repo ./api --repo ./worker
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --force              Overwrite locally-modified skills (and upgrade pinned ones)
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for upgrade
      --json               Output results as JSON
      --no-validate        Skip spec validation even if policy requires it
      --repo stringArray   Target repository path (repeatable)
      --target string      Agent platform to manage
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
      --validate           Validate the skill against the spec before writing (fail on errors)
  -y, --yes                Don't ask before upgrades that add scripts or broad tool access
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

