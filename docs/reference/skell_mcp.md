## skell mcp

Run Skell as an MCP server so AI agents can find and install skills

### Synopsis

Starts a Model Context Protocol server on stdio. Add it to your agent so it
can browse, review, install and update skills itself ("find me a skill for PDF
forms and install it").

Claude Code:   claude mcp add skell -- skell mcp
Other clients: command "skell", args ["mcp"]

The server works on the project in the current directory (or --repo). Installs
of skills that ship scripts or pre-approve broad tool access are not performed
until the agent calls again with confirm=true, so you see the review first.

```
skell mcp [flags]
```

### Options

```
  -h, --help          help for mcp
      --repo string   Project the server operates on (defaults to current directory)
```

### SEE ALSO

* [skell](skell.md)	 - Install, update and share Agent Skills (SKILL.md) across projects and agents.

