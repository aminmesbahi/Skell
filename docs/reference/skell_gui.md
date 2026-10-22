## skell gui

Open the desktop app (offers to download it if you do not have it)

### Synopsis

Opens the Skell desktop app: browse and install skills, see what needs
attention across your projects, review upgrades and run security checks in a
window instead of the terminal.

The app is found next to the skell program or where its installer puts it. If
you do not have it yet, skell offers to download the latest version from
GitHub, checks it against the SHA-256 published with the release, installs it
for your user (no administrator rights) and opens it. Use --install to do that
without being asked, for example in a script.

```
skell gui [flags]
```

### Examples

```
  skell gui
  skell gui --install
```

### Options

```
  -h, --help      help for gui
      --install   Download and install the desktop app without asking when it is missing
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

