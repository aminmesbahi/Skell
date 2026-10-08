## skell init

Create skell.toml from currently installed skills

### Synopsis

Scans the repository for installed skills and generates a skell.toml manifest.

Skell supports multiple AI-agent layouts:

  claude    .claude/skills/   (Anthropic Claude Code, agentskills.io)
  codex     .codex/skills/    (OpenAI Codex CLI)
  copilot   .github/skills/   (VS Code & GitHub Copilot)
  cursor    .cursor/skills/   (Cursor)
  windsurf  .windsurf/skills/ (Windsurf / Cascade)
  opencode  .opencode/skills/ (OpenCode)
  cline     .cline/skills/    (Cline)
  grok      .grok/skills/     (xAI Grok)

If a known agent folder already exists in the repo it is used automatically.
Otherwise pass --target to choose, or run interactively to be prompted.

```
skell init [flags]
```

### Examples

```
  # Initialise the current directory (auto-detect or prompt)
  skell init

  # Initialise for a specific platform
  skell init --target copilot

  # Initialise a specific repository path
  skell init --repo /path/to/repo --target cursor
```

### Options

```
  -h, --help             help for init
      --mirror strings   Also provide skills to these agents, e.g. --mirror copilot,cursor (see 'skell mirror')
      --repo string      Target repository path (defaults to current directory)
      --target string    Agent platform layout: claude | codex | copilot | cursor | windsurf | opencode | cline | grok
      --user             Set up your personal, user-level skills (e.g. ~/.claude/skills) instead of a project
      --yes              Do not prompt; use --target or the default (claude)
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

