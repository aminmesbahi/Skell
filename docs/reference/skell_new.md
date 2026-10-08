## skell new

Create a new skill (or a whole skill-source repository)

### Synopsis

Scaffolds a spec-compliant skill folder with a starter SKILL.md.

By default the skill is created in a standalone folder (--path, default the
current directory). Pass --link to also link it into the current project so
your agent can use it while you write it.

With --registry, creates a skill-source repository instead: a README, an
example skill under skills/, and a GitHub Actions workflow that validates
every skill on pull requests.

```
skell new <name> [flags]
```

### Examples

```
  # A new skill in ./release-notes
  skell new release-notes --description "Write release notes from merged PRs. Use when preparing a release."

  # Create it in a skills folder and use it in this project right away
  skell new release-notes --path ~/src/my-skills --link

  # Start a team skill repository
  skell new team-skills --registry
```

### Options

```
      --description string   What the skill does and when to use it
  -h, --help                 help for new
      --link                 Link the new skill into the current project (see 'skell link')
      --owner string         Owning team or person (metadata.owner)
      --path string          Parent folder to create the skill (or repository) in (default ".")
      --registry             Create a skill-source repository instead of a single skill
      --repo string          Project to link into with --link (defaults to current directory)
      --scripts              Also create scripts/ and references/ folders
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

