## skell validate

Validate installed skills against the Agent Skills spec

### Synopsis

Validates one or all installed skills using the Agent Skills validator.

By default only offline structure checks run (required frontmatter, directory
layout, token budgets, code-fence integrity, internal links). Use --full to add
offline content-quality and contamination analysis, and --links to additionally
check external links over the network.

Exits non-zero if any skill has errors (or, with --strict, any warnings).

```
skell validate [skill-name] [flags]
```

### Examples

```
  # Validate every installed skill in the current repo
  skell validate

  # Validate a single skill
  skell validate pdf-processing

  # Full offline analysis, failing on warnings too
  skell validate --full --strict

  # Also verify external links resolve
  skell validate --links

  # JSON output for CI
  skell validate --json
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --full               Also run offline content-quality and contamination analysis
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for validate
      --json               Output results as JSON
      --links              Also validate external links (network access)
      --path string        Validate a skill directory (or every skill under a folder) without installing it
      --repo stringArray   Target repository path (repeatable)
      --strict             Treat warnings as failures (non-zero exit)
      --target string      Agent platform to validate
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

