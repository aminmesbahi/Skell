## skell report

See the skills of many repositories at once: versions, drift, risk, signatures, activity

### Synopsis

Builds a fleet report over several repositories: which skills each one runs
and in which versions, who is outdated or hand-edited, the security risk and
signature status of every install, repositories that need attention, and
recent install/upgrade activity from the audit log.

By default it reads only local files (fast, works offline). Add
--check-updates to ask each source whether newer versions exist.

Formats: text (default), json, md, csv, html. The html report is a single
self-contained page you can open, attach to an email or archive.

--webhook posts the JSON report to an address you choose (https only, or
localhost). Nothing is sent anywhere unless you pass it.

```
skell report [flags]
```

### Examples

```
  # Every repo under ~/work, as a page you can share
  skell report --all-repos ~/work --format html --out fleet.html

  # Spreadsheet-friendly
  skell report --all-repos ~/work --format csv --out fleet.csv

  # Include update checks, ship to a dashboard
  skell report --all-repos ~/work --check-updates --webhook https://hooks.example.com/skell
```

### Options

```
      --all-repos string   Scan all git repos under this root path
      --check-updates      Ask each source whether newer versions exist (uses the network)
      --dry-run            Preview changes without applying them
      --events int         Recent audit-log entries to include (0 for none) (default 20)
      --format string      Output format: text | json | md | csv | html
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for report
      --json               Output results as JSON
  -o, --out string         Write the report to a file instead of the terminal
      --repo stringArray   Target repository path (repeatable)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
      --webhook string     POST the JSON report to this https URL
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

