## skell add

Add a skill source (or a single skill) by catalog id, owner/repo, URL or path

### Synopsis

Adds a skill source to skell.toml, or installs a single skill from it.

<source> can be:

  • a catalog id                 skell add anthropic          (see 'skell catalog')
  • a GitHub owner/repo          skell add davidfowl/dotnet-skillz
  • one skill in a GitHub repo   skell add anthropics/skills/skills/pdf
  • a full git or GitHub URL     skell add https://github.com/owner/repo/tree/main/skills/my-skill
  • a local folder               skell add ./my-skills

When the source points at a single skill it is installed (after a short review
if it ships scripts or asks for broad tool access); otherwise the source is
registered so 'skell search' and 'skell install' can use it.

The registry alias is the repository name (or the catalog id); if that name is
already taken by another source, the owner is prefixed (e.g. "openai-skills").
Use --alias to choose it yourself.

```
skell add <source>... [flags]
```

### Examples

```
  # Add a well-known source and browse it
  skell add anthropic
  skell search pdf

  # Add a GitHub repo by owner/repo
  skell add davidfowl/dotnet-skillz

  # Install one skill directly
  skell add anthropics/skills/skills/pdf

  # Full URL, explicit alias, preview only
  skell add https://github.com/owner/repo --alias team --dry-run
```

### Options

```
      --alias string       Registry alias to use instead of the derived one
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for add
      --json               Output results as JSON
      --repo stringArray   Target repository path (repeatable)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
  -y, --yes                Don't ask for confirmation before installing
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

