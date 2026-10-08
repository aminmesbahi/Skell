## skell cache refresh

Fetch latest from all configured registries

### Synopsis

Refreshes the local registry cache: configured registries (global sources
plus the selected repo's manifest, if any) and every other registry already
cached on disk. A repo manifest is optional, so this works from any directory or
against the global Shared Library.

```
skell cache refresh [flags]
```

### Examples

```
  # Refresh all configured + cached registries
  skell cache refresh

  # Also include a specific repo's manifest registries
  skell cache refresh --repo /path/to/repo
```

### Options

```
  -h, --help          help for refresh
      --repo string   Optional repository path to include its manifest registries
```

### SEE ALSO

* [skell cache](skell_cache.md)	 - Manage the local registry cache

