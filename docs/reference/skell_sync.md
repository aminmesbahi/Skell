## skell sync

Apply skell.toml to the repository (install missing, remove unlisted)

### Synopsis

Reconciles the repository's installed skills with skell.toml.

Skills listed in skell.toml but not installed are fetched and installed.
Skills that Skell previously installed (recorded in skell.lock) but are no
longer listed in skell.toml are removed.

Hand-authored skills that Skell never installed (not in skell.lock) are left
in place and reported as "untracked". Pass --prune to remove them too.
Installed skills are also verified against the content hash in skell.lock. A
skill edited by hand is reported as "modified" (and fails --check) but is left
alone; pass --force to discard the edits and restore the locked content.
Use --check to detect drift without making any changes; add --allow-modified if
you customise installed skills on purpose and don't want that to fail the check.

```
skell sync [flags]
```

### Examples

```
  # Sync the current repo
  skell sync

  # Preview what would change without applying
  skell sync --dry-run

  # Only check for drift (exit non-zero if out of sync)
  skell sync --check

  # Also remove hand-authored skills not in the manifest
  skell sync --prune

  # Discard local edits to installed skills and restore the locked content
  skell sync --force

  # Sync multiple repos
  skell sync --repo ./api --repo ./worker
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --allow-modified     With --check, do not fail for skills edited since install
      --check              Exit non-zero if state differs from manifest (CI use)
      --dry-run            Preview changes without applying them
      --force              Discard local edits to installed skills and restore them from skell.lock
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for sync
      --json               Output results as JSON
      --prune              Also remove hand-authored skills not in the manifest or lock file
      --repo stringArray   Target repository path (repeatable)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

