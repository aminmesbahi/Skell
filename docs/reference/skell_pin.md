## skell pin

Pin a skill to its current version

### Synopsis

Marks a skill as pinned so that 'skell upgrade' will not update it.
The installed version is recorded in skell.toml and skell.lock.

```
skell pin <skill-name> [flags]
```

### Examples

```
  # Pin a skill to its currently-installed version
  skell pin pdf-processing

  # Pin to a specific version
  skell pin pdf-processing --version 1.2.0

  # Pin in a specific repo
  skell pin pdf-processing --repo /path/to/repo
```

### Options

```
  -h, --help             help for pin
      --json             Output as JSON
      --repo string      Target repository path
      --target string    Agent platform to manage
      --version string   Pin to a specific version instead of installed
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

