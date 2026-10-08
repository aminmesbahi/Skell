## skell

Install, update and share Agent Skills (SKILL.md) across projects and agents.

### Synopsis

Skell is a package manager for Agent Skills (SKILL.md) — for Claude Code,
Codex, GitHub Copilot, Cursor and other agents.

Get started:
  skell init                  # set up this project (pick your agents)
  skell catalog               # browse well-known skill sources
  skell add anthropic         # add one (catalog id, owner/repo, URL or folder)
  skell search pdf            # find skills
  skell install pdf           # install (shows a review if it ships scripts)
  skell status                # anything outdated?  then: skell diff / upgrade

Teams commit skell.toml and skell.lock; everyone else runs 'skell sync' to
get exactly the same skills. Run 'skell <command> --help' for details.

```
skell [flags]
```

### Options

```
  -h, --help   help for skell
```

### SEE ALSO

* [skell add](skell_add.md)	 - Add a skill source (or a single skill) by catalog id, owner/repo, URL or path
* [skell apply](skell_apply.md)	 - Adopt a shared team baseline (sources + skills) in one step
* [skell cache](skell_cache.md)	 - Manage the local registry cache
* [skell catalog](skell_catalog.md)	 - Browse well-known skill sources you can add
* [skell completion](skell_completion.md)	 - Generate the autocompletion script for the specified shell
* [skell diff](skell_diff.md)	 - Show what changed upstream for an installed skill
* [skell doctor](skell_doctor.md)	 - Check for manifest, lock file, and install problems
* [skell gui](skell_gui.md)	 - Launch the desktop GUI
* [skell info](skell_info.md)	 - Show full metadata for a skill
* [skell init](skell_init.md)	 - Create skell.toml from currently installed skills
* [skell install](skell_install.md)	 - Install one or more skills into one or more repositories
* [skell link](skell_link.md)	 - Use a skill you are developing without copying it (live edits)
* [skell list](skell_list.md)	 - List installed or registry skills
* [skell mcp](skell_mcp.md)	 - Run Skell as an MCP server so AI agents can find and install skills
* [skell mirror](skell_mirror.md)	 - Share one manifest's skills with other agents in the same repo
* [skell new](skell_new.md)	 - Create a new skill (or a whole skill-source repository)
* [skell pin](skell_pin.md)	 - Pin a skill to its current version
* [skell remove](skell_remove.md)	 - Remove a skill from one or more repositories
* [skell review](skell_review.md)	 - Inspect a skill before installing it (files, scripts, tools, links)
* [skell search](skell_search.md)	 - Search available skills in configured registries
* [skell selfupdate](skell_selfupdate.md)	 - Upgrade skell to the latest release from GitHub
* [skell status](skell_status.md)	 - Show comparison between registry and local installs
* [skell sync](skell_sync.md)	 - Apply skell.toml to the repository (install missing, remove unlisted)
* [skell targets](skell_targets.md)	 - List supported AI-agent platform layouts
* [skell unpin](skell_unpin.md)	 - Remove pinning for a skill
* [skell upgrade](skell_upgrade.md)	 - Upgrade one or all skills
* [skell validate](skell_validate.md)	 - Validate installed skills against the Agent Skills spec

