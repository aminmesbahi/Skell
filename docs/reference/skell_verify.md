## skell verify

Check who published installed skills and that nobody has changed them

### Synopsis

Checks each skill's signature against the publishers you trust
('skell trust list'). Results:

  trusted        signed by a publisher you trust, and unchanged
  untrusted-key  validly signed, but by someone you have not trusted
  tampered       files changed after signing
  invalid        the signature file is damaged or forged
  unsigned       no signature (most skills today)

Exits non-zero when a skill is tampered or invalid. With --require-signed it
also fails on unsigned and untrusted skills.

```
skell verify [skill-name...] [flags]
```

### Examples

```
  skell verify
  skell verify pdf --json
  skell verify --path ./my-skill
  skell verify --require-signed      # for CI in a locked-down team
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for verify
      --json               Output results as JSON
      --path string        Verify a skill folder (or a folder of skills) instead of installed skills
      --repo stringArray   Target repository path (repeatable)
      --require-signed     Also fail on unsigned skills and untrusted publishers
      --target string      Agent platform to check
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

