## skell scan

Scan skills for risky content (prompt injection, hidden text, pipe-to-shell, secrets…)

### Synopsis

Statically inspects skills for the patterns that make an agent skill dangerous:
instructions that try to override the agent or hide things from you, commands
that pipe a download into a shell, credential access and exfiltration,
obfuscated code, persistence, destructive commands, binaries and broad tool
permissions. Each skill gets a risk level (clean, low, medium, high, critical)
and a score from 0 to 100.

The scan is heuristic: it points at what deserves a human look, it does not
prove a skill safe or unsafe. Exits non-zero when a skill reaches --fail-on.

Skell also runs this scan during 'skell review' and 'skell install', and an
organisation can refuse risky skills with 'max-risk' in ~/.skell/config.toml.

```
skell scan [skill-name...] [flags]
```

### Examples

```
  # Every installed skill in the current project
  skell scan

  # One skill, with every finding listed
  skell scan pdf --verbose

  # A skill you are writing, or a whole folder of skills
  skell scan --path ./my-skill
  skell scan --path ~/src/team-skills --fail-on medium

  # Machine-readable
  skell scan --json
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --fail-on string     Exit non-zero when a skill reaches this level: low | medium | high | critical (empty: never) (default "high")
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for scan
      --json               Output results as JSON
      --path string        Scan a skill folder (or a folder of skills) instead of installed skills
      --repo stringArray   Target repository path (repeatable)
      --target string      Agent platform to scan
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
  -v, --verbose            List every finding, including notes
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

