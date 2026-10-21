## skell bundle install

Install every skill in one or more bundles

### Synopsis

Same as 'skell install @name', with all of install's checks: dependencies are added, and skills that ship scripts or broad tool access are reviewed first.

```
skell bundle install <name>... [flags]
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for install
      --json               Output results as JSON
      --repo stringArray   Target repository path (repeatable)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
  -y, --yes                Don't ask for confirmation before installing
```

### SEE ALSO

* [skell bundle](skell_bundle.md)	 - Named sets of skills (a role or team starter pack) you install in one step

