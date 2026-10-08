## skell review

Inspect a skill before installing it (files, scripts, tools, links)

### Synopsis

Shows what installing a skill would add to the project — its files, any
scripts or executables, the tools it pre-approves (allowed-tools) and external
links — with warnings for anything that deserves a closer look. Nothing is
installed.

```
skell review <skill-name> [flags]
```

### Examples

```
  skell review pdf
  skell review pdf --registry anthropic --json
```

### Options

```
  -h, --help              help for review
      --json              Output as JSON
      --registry string   Source alias (found automatically when omitted)
      --repo string       Project whose sources to use (defaults to current directory)
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

