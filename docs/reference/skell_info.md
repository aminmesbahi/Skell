## skell info

Show full metadata for a skill

### Synopsis

Displays the full metadata for a skill, including frontmatter fields and lock file state.

```
skell info <skill-name> [flags]
```

### Examples

```
  # Show info for an installed skill
  skell info pdf-processing

  # Show info for a skill in a specific repo
  skell info pdf-processing --repo /path/to/repo

  # Show info for a specific agent platform
  skell info pdf-processing --target cursor

  # Look up a skill in the registry (not yet installed)
  skell info ilspy-decompile --source registry

  # Output as JSON
  skell info pdf-processing --json
```

### Options

```
  -h, --help            help for info
      --json            Output results as JSON
      --repo string     Target repository path (defaults to current directory)
      --source string   Show only: registry | local
      --target string   Agent platform to look up: claude | codex | copilot | cursor | windsurf | opencode | cline | grok
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

