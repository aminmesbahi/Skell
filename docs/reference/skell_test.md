## skell test

Run the eval cases a skill ships with (evals/evals.yaml) against your agent

### Synopsis

Skills can ship test cases in evals/evals.yaml: a prompt and what a good answer
looks like. 'skell test' installs the skill into a scratch workspace, gives each
prompt to your agent, and checks the answer. It tells you whether a skill —
yours or one you are about to upgrade to — still does its job.

The agent is any command that reads a prompt on standard input (or takes it
where you write {prompt}) and prints its answer, e.g. --agent "claude -p".
It defaults to $SKELL_TEST_AGENT. Every case is a real agent call, so there is
no built-in default. Without an agent only the suite itself is checked and cases
are reported as skipped, so this still works in plain CI.

An eval file looks like:

  cases:
    - name: summarises a report
      prompt: Summarise report.pdf in three bullets.
      timeout: 180
      expect:
        contains: ["summary"]
        not_contains: ["I cannot"]
        matches: ["(?m)^- "]

'skell new --evals' writes a starter file. 'skell upgrade --verify' runs the
tests of the new version before it replaces the old one.

```
skell test [skill-name...] [flags]
```

### Examples

```
  # A skill you are writing
  skell test --path ./release-notes --agent "claude -p"

  # Every installed skill that has evals
  skell test

  # One case, as JSON
  skell test pdf --case "extracts text" --json
```

### Options

```
      --agent string       Command that answers a prompt, e.g. "claude -p" (default: $SKELL_TEST_AGENT)
      --all-repos string   Scan all git repos under this root path
      --case stringArray   Run only this case (repeatable)
      --dry-run            Preview changes without applying them
      --global             Operate on the global manifest (~/.skell/.claude/skell.toml)
  -h, --help               help for test
      --json               Output results as JSON
      --path string        Test a skill folder (or a folder of skills) instead of installed skills
      --repo stringArray   Target repository path (repeatable)
      --target string      Agent platform the skill is installed for
      --timeout int        Seconds allowed per case that sets no timeout of its own (default 120)
      --user               Operate on your personal, user-level agent folders (e.g. ~/.claude/skills) available in every project
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

