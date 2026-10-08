## skell search

Search available skills in configured registries

### Synopsis

Searches all configured registries for skills matching the optional query string.
Results can be filtered by tag, lifecycle stage, or owner.

```
skell search [query] [flags]
```

### Examples

```
  # Search all skills
  skell search

  # Full-text search by name/description
  skell search pdf

  # Filter by lifecycle
  skell search --lifecycle stable

  # Filter by owner
  skell search --owner dotnet

  # Combine filters
  skell search dotnet --lifecycle stable --owner microsoft

  # Output as JSON
  skell search --json
```

### Options

```
  -h, --help               help for search
      --json               Output results as JSON
      --lifecycle string   Filter by lifecycle state
      --owner string       Filter by owner
      --repo string        Target repository path (for manifest resolution)
      --tag string         Filter by tag
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

