// Package skell wires all cobra commands and exposes Execute.
package skell

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/version"
	"github.com/spf13/cobra"
)

// newRootCmd builds a fresh root command tree. Calling this for every test
// run ensures that flag state (e.g. StringArray accumulation) does not leak
// between test cases.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "skell",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version.Version, version.Commit, version.Date),
		Short:   "Install, update and share Agent Skills (SKILL.md) across projects and agents.",
		Long: `Skell is a package manager for Agent Skills (SKILL.md) — for Claude Code,
Codex, GitHub Copilot, Cursor and other agents.

Get started:
  skell init                  # set up this project (pick your agents)
  skell catalog               # browse well-known skill sources
  skell add anthropic         # add one (catalog id, owner/repo, URL or folder)
  skell search pdf            # find skills
  skell install pdf           # install (shows a review if it ships scripts)
  skell status                # anything outdated?  then: skell diff / upgrade

Teams commit skell.toml and skell.lock; everyone else runs 'skell sync' to
get exactly the same skills. Run 'skell <command> --help' for details.`,
		// Print just the error for runtime failures; a full usage dump buries
		// the actual message. Flag/argument mistakes still show a hint.
		SilenceUsage: true,
	}
	root.AddGroup(
		&cobra.Group{ID: "start", Title: "Getting started:"},
		&cobra.Group{ID: "manage", Title: "Managing skills:"},
		&cobra.Group{ID: "team", Title: "Teams, CI and agents:"},
		&cobra.Group{ID: "author", Title: "Writing skills:"},
	)
	add := func(group string, cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.GroupID = group
			root.AddCommand(c)
		}
	}
	add("start", newInitCmd(), newCatalogCmd(), newAddCmd(), newSearchCmd(), newReviewCmd(), newInstallCmd())
	add("manage", newListCmd(), newStatusCmd(), newDiffCmd(), newUpgradeCmd(), newRemoveCmd(),
		newPinCmd(), newUnpinCmd(), newInfoCmd(), newDoctorCmd())
	add("team", newSyncCmd(), newApplyCmd(), newMirrorCmd(), newTargetsCmd(), newMCPCmd(), newGUICmd())
	add("author", newNewCmd(), newLinkCmd(), newValidateCmd())
	root.AddCommand(newCacheCmd(), newSelfUpdateCmd(), newGenDocsCmd())

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {}
	root.RunE = func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		if _, err := fmt.Fprintf(w, "skell version %s (commit %s, built %s)\n\n",
			version.Version, version.Commit, version.Date); err != nil {
			return err
		}
		cwd, err := os.Getwd()
		if err == nil {
			printDashboard(w, cwd)
		}
		_, _ = fmt.Fprintln(w, "\n  Run 'skell --help' for all commands.")
		return nil
	}

	return root
}

// printDashboard summarises the project in dir (no network access), or shows
// how to get started when dir has no manifest yet.
func printDashboard(w io.Writer, dir string) {
	m, t, err := manifest.ResolveWithTarget(dir)
	if err != nil {
		_, _ = fmt.Fprintf(w, "  %s is not set up for Skell yet.\n\n", dir)
		_, _ = fmt.Fprintln(w, "  Get started:")
		_, _ = fmt.Fprintln(w, "    skell init              set up this project")
		_, _ = fmt.Fprintln(w, "    skell catalog           browse well-known skill sources")
		_, _ = fmt.Fprintln(w, "    skell add anthropic     add a source, then 'skell search' and 'skell install'")
		_, _ = fmt.Fprintln(w, "    skell init --user       manage your personal skills (all projects)")
		return
	}
	eng := engine.New(defaultCacheRoot())
	skills, _ := eng.ListFor(dir, t.ID)
	agents := t.ID
	if len(m.Mirrors) > 0 {
		agents += " (+ " + strings.Join(m.Mirrors, ", ") + ")"
	}
	_, _ = fmt.Fprintf(w, "  project  %s\n", dir)
	_, _ = fmt.Fprintf(w, "  agents   %s\n", agents)
	_, _ = fmt.Fprintf(w, "  sources  %d   skills %d\n", len(m.Registries), len(skills))

	_, _ = fmt.Fprintln(w, "\n  Next:")
	switch {
	case len(m.Registries) == 0:
		_, _ = fmt.Fprintln(w, "    skell catalog           find a skill source")
		_, _ = fmt.Fprintln(w, "    skell add <source>      add it (catalog id, owner/repo or URL)")
	case len(skills) == 0:
		_, _ = fmt.Fprintln(w, "    skell search            browse skills from your sources")
		_, _ = fmt.Fprintln(w, "    skell install <skill>   install one")
	default:
		_, _ = fmt.Fprintln(w, "    skell status --refresh  check for updates")
		_, _ = fmt.Fprintln(w, "    skell search            find more skills")
		_, _ = fmt.Fprintln(w, "    skell doctor            check everything is healthy")
	}
}

var rootCmd = newRootCmd()

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
