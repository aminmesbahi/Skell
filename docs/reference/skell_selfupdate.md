## skell selfupdate

Upgrade skell to the latest release from GitHub

### Synopsis

Checks GitHub Releases for a newer version of skell.
If a newer version is found, downloads the platform-specific binary and
replaces the running executable in-place.

Use --check to only report whether an update is available without applying it.

```
skell selfupdate [flags]
```

### Examples

```
  # Check if a newer version exists (no download)
  skell selfupdate --check

  # Download and apply the latest release
  skell selfupdate
```

### Options

```
      --check   Only report if an update is available; do not download
  -h, --help    help for selfupdate
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

