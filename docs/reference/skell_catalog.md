## skell catalog

Browse well-known skill sources you can add

### Synopsis

Lists curated, well-known skill sources (git repositories of SKILL.md skills).
Add one to the current project with 'skell add <id>', then browse its skills
with 'skell search'.

The catalog ships with Skell and is refreshed from the Skell repository at most
once a day. Set SKELL_OFFLINE=1 to never contact the network, or
SKELL_CATALOG_URL to use your organisation's own catalog.

```
skell catalog [query] [flags]
```

### Examples

```
  # Show every catalog source
  skell catalog

  # Filter by keyword
  skell catalog dotnet

  # Add a source by id, then search it
  skell add anthropic
  skell search pdf
```

### Options

```
  -h, --help      help for catalog
      --json      Output as JSON
      --refresh   Download the latest catalog now
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

