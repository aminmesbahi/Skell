package skell

import (
	"fmt"
	"io"
	"strings"
)

// tagline is the one-sentence answer to "what is this?".
const tagline = "Skell is a package manager for Agent Skills."

// whatIsSkell explains the idea in plain words, for someone who has never
// heard of it. It leads both 'skell --help' and the screen bare 'skell' shows.
const whatIsSkell = `Agent Skills are small folders with a SKILL.md file that teach an AI coding
agent (Claude Code, Codex, GitHub Copilot, Cursor and others) how to do one
job well: fill in PDF forms, write release notes, review a pull request.

Skell finds those skills, installs them into your project, keeps them up to
date, checks that they are safe, and lets a whole team share the same set —
the way npm or pip do for code.`

// whatYouCanDo lists the main jobs, each with the command that does it.
const whatYouCanDo = `  Find skills        skell catalog            well-known sources of skills
                     skell search <keyword>   search the skills you have added
  Install safely     skell install <skill>    shows what the skill contains first
  Stay up to date    skell status             what is outdated?
                     skell upgrade            take the updates
  Check trust        skell scan               look for risky content
                     skell verify             who published it, was it changed?
  Share with a team  commit .claude/skell.toml and skell.lock;
                     teammates run  skell sync
  Write your own     skell new my-skill       start a new skill`

// printIntro explains what Skell is and what it can do. Shown on a bare
// 'skell' before the project summary.
func printIntro(w io.Writer) {
	_, _ = fmt.Fprintf(w, "  %s\n\n", tagline)
	for _, line := range strings.Split(whatIsSkell, "\n") {
		if line == "" {
			_, _ = fmt.Fprintln(w)
			continue
		}
		_, _ = fmt.Fprintf(w, "  %s\n", line)
	}
	_, _ = fmt.Fprintf(w, "\n  What you can do\n%s\n\n", whatYouCanDo)
}
