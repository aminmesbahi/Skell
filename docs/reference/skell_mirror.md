## skell mirror

Share one manifest's skills with other agents in the same repo

### Synopsis

Many repositories are used with several AI agents at once (for example Claude
Code, GitHub Copilot and Cursor). Mirrors let one skell.toml serve them all:
the manifest's own agent folder holds the real skills, and every mirror
target's skills/ folder receives an identical copy, kept up to date by
install, upgrade, remove and sync.

Mirror copies are derived files — 'skell sync' rebuilds them and
'skell doctor' reports when they drift.

### Examples

```
  # Also provide the skills to Copilot and Cursor
  skell mirror add copilot cursor

  # See what is mirrored
  skell mirror list

  # Stop mirroring to Cursor (removes the copies Skell placed there)
  skell mirror remove cursor
```

### Options

```
  -h, --help          help for mirror
      --json          Output as JSON
      --repo string   Target repository path (defaults to current directory)
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.
* [skell mirror add](skell_mirror_add.md)	 - Mirror skills into more agent folders
* [skell mirror list](skell_mirror_list.md)	 - Show mirror targets
* [skell mirror refresh](skell_mirror_refresh.md)	 - Rebuild every mirror from the primary install
* [skell mirror remove](skell_mirror_remove.md)	 - Stop mirroring into agent folders (removes mirrored copies)

