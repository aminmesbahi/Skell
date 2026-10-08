## skell doctor

Check for manifest, lock file, and install problems

### Synopsis

Audits the repository for common Skell problems:
  • missing or malformed skell.toml
  • missing skell.lock
  • skills listed in the lock file but missing on disk
  • installed skills whose content hash no longer matches the lock file
  • skills that fail Agent Skills spec validation
  • mirror copies (see 'skell mirror') that are missing or out of date

With --fix, Skell repairs what it safely can (reinstalls missing skills at
their locked commit and refreshes mirrors) and then
re-checks. Local edits are never overwritten.

```
skell doctor [flags]
```

### Examples

```
  # Run diagnostics on the current repo
  skell doctor

  # Repair what can be repaired automatically
  skell doctor --fix

  # Output issues as JSON (useful for CI)
  skell doctor --json
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --fix                Reinstall missing skills and refresh mirrors, then re-check
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for doctor
      --json               Output results as JSON
      --repo stringArray   Target repository path (repeatable)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

