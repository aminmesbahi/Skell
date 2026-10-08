## skell diff

Show what changed upstream for an installed skill

### Synopsis

Shows a unified diff from the installed copy of a skill to the current
version in its source, so you can review an upgrade before running
'skell upgrade'. Local edits you made to the installed copy show up too.

Use --refresh to fetch the latest from the source first.

```
skell diff <skill-name> [flags]
```

### Examples

```
  # Review upstream changes before upgrading
  skell diff pdf --refresh
  skell upgrade pdf

  # Machine-readable
  skell diff pdf --json
```

### Options

```
  -h, --help            help for diff
      --json            Output as JSON
      --no-color        Disable colored output
      --refresh         Fetch the latest from the source before comparing
      --repo string     Target repository path (defaults to current directory)
      --target string   Agent platform the skill is installed for
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

