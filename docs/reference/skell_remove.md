## skell remove

Remove a skill from one or more repositories

### Synopsis

Removes a skill's files and its entries from skell.toml and skell.lock.

```
skell remove <skill-name> [flags]
```

### Examples

```
  # Remove a skill from the current repo
  skell remove pdf-processing

  # Preview the removal without deleting files
  skell remove pdf-processing --dry-run

  # Remove from a specific repo
  skell remove pdf-processing --repo /path/to/repo
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for remove
      --json               Output results as JSON
      --repo stringArray   Target repository path (repeatable)
      --target string      Agent platform to manage
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

