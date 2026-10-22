## skell

Install, update and share Agent Skills (SKILL.md) across projects and agents.

### Synopsis

Skell is a package manager for Agent Skills.

Agent Skills are small folders with a SKILL.md file that teach an AI coding
agent (Claude Code, Codex, GitHub Copilot, Cursor and others) how to do one
job well: fill in PDF forms, write release notes, review a pull request.

Skell finds those skills, installs them into your project, keeps them up to
date, checks that they are safe, and lets a whole team share the same set —
the way npm or pip do for code.

What you can do:
  Find skills        skell catalog            well-known sources of skills
                     skell search <keyword>   search the skills you have added
  Install safely     skell install <skill>    shows what the skill contains first
  Stay up to date    skell status             what is outdated?
                     skell upgrade            take the updates
  Check trust        skell scan               look for risky content
                     skell verify             who published it, was it changed?
  Share with a team  commit .claude/skell.toml and skell.lock;
                     teammates run  skell sync
  Write your own     skell new my-skill       start a new skill

Get started:
  skell init                  # set up this project (pick your agents)
  skell catalog               # browse well-known skill sources
  skell add anthropic         # add one (catalog id, owner/repo, URL or folder)
  skell search <keyword>      # find skills, e.g. skell search pdf
  skell install <skill>       # install a skill by name (shows a review first)
  skell status                # anything outdated?  then: skell diff / upgrade

Run 'skell <command> --help' for details on any command, or 'skell gui' for
the desktop app. Documentation: https://github.com/aminmesbahi/skell

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
* [skell bundle](skell_bundle.md)	 - Named sets of skills (a role or team starter pack) you install in one step
* [skell cache](skell_cache.md)	 - Manage the local registry cache
* [skell catalog](skell_catalog.md)	 - Browse well-known skill sources you can add
* [skell completion](skell_completion.md)	 - Generate the autocompletion script for the specified shell
* [skell diff](skell_diff.md)	 - Show what changed upstream for an installed skill
* [skell doctor](skell_doctor.md)	 - Check for manifest, lock file, and install problems
* [skell gui](skell_gui.md)	 - Open the desktop app (offers to download it if you do not have it)
* [skell info](skell_info.md)	 - Show full metadata for a skill
* [skell init](skell_init.md)	 - Create skell.toml from currently installed skills
* [skell install](skell_install.md)	 - Install one or more skills into one or more repositories
* [skell keygen](skell_keygen.md)	 - Create a signing key for publishing skills
* [skell link](skell_link.md)	 - Use a skill you are developing without copying it (live edits)
* [skell list](skell_list.md)	 - List installed or registry skills
* [skell mcp](skell_mcp.md)	 - Run Skell as an MCP server so AI agents can find and install skills
* [skell mirror](skell_mirror.md)	 - Share one manifest's skills with other agents in the same repo
* [skell new](skell_new.md)	 - Create a new skill (or a whole skill-source repository)
* [skell pin](skell_pin.md)	 - Pin a skill to its current version
* [skell remove](skell_remove.md)	 - Remove a skill from one or more repositories
* [skell report](skell_report.md)	 - See the skills of many repositories at once: versions, drift, risk, signatures, activity
* [skell review](skell_review.md)	 - Inspect a skill before installing it (files, scripts, tools, links)
* [skell scan](skell_scan.md)	 - Scan skills for risky content (prompt injection, hidden text, pipe-to-shell, secrets…)
* [skell search](skell_search.md)	 - Search available skills in configured registries
* [skell selfupdate](skell_selfupdate.md)	 - Upgrade skell to the latest release from GitHub
* [skell sign](skell_sign.md)	 - Sign a skill (or every skill in a folder) so others can verify who published it
* [skell status](skell_status.md)	 - Show comparison between registry and local installs
* [skell sync](skell_sync.md)	 - Apply skell.toml to the repository (install missing, remove unlisted)
* [skell targets](skell_targets.md)	 - List supported AI-agent platform layouts
* [skell test](skell_test.md)	 - Run the eval cases a skill ships with (evals/evals.yaml) against your agent
* [skell trust](skell_trust.md)	 - Choose which skill publishers you trust
* [skell unpin](skell_unpin.md)	 - Remove pinning for a skill
* [skell upgrade](skell_upgrade.md)	 - Upgrade one or all skills
* [skell validate](skell_validate.md)	 - Validate installed skills against the Agent Skills spec
* [skell verify](skell_verify.md)	 - Check who published installed skills and that nobody has changed them

