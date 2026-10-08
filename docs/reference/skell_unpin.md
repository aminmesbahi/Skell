## skell unpin

Remove pinning for a skill

### Synopsis

Clears the pinned flag so the skill will be included in future 'skell upgrade' runs.

```
skell unpin <skill-name> [flags]
```

### Examples

```
  # Unpin a skill
  skell unpin pdf-processing

  # Unpin in a specific repo
  skell unpin pdf-processing --repo /path/to/repo
```

### Options

```
  -h, --help            help for unpin
      --json            Output as JSON
      --repo string     Target repository path
      --target string   Agent platform to manage
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

