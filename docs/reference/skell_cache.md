## skell cache

Manage the local registry cache

### Synopsis

Subcommands for managing the local clone of remote registry repositories.

Registries are cloned to ~/.skell/cache/<alias>/ on first use and updated
with 'skell cache refresh'.

### Examples

```
  skell cache status
  skell cache refresh
  skell cache clear
```

### Options

```
  -h, --help   help for cache
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.
* [skell cache clear](skell_cache_clear.md)	 - Delete all cached registry data
* [skell cache refresh](skell_cache_refresh.md)	 - Fetch latest from all configured registries
* [skell cache status](skell_cache_status.md)	 - Show cache contents and sizes

