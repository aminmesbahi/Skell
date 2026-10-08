## skell link

Use a skill you are developing without copying it (live edits)

### Synopsis

Links a local skill folder into the repository's skills directory, so the
agent sees your edits immediately — like 'npm link'. Skell uses a symlink, or
a directory junction on Windows when symlinks need admin rights.

Linked skills are recorded in skell.lock as linked, are never pruned by sync
or touched by upgrade, and are added to .git/info/exclude so the
machine-specific link is not committed. Unlink with 'skell remove <name>'
(your folder is left untouched).

```
skell link <skill-folder> [flags]
```

### Examples

```
  # Work on a skill in another folder
  skell link ~/src/my-skills/release-notes

  # Stop using it
  skell remove release-notes
```

### Options

```
  -h, --help            help for link
      --json            Output as JSON
      --repo string     Target repository path (defaults to current directory)
      --target string   Agent platform to link for
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

