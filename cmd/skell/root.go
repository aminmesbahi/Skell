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
		Long: tagline + "\n\n" + whatIsSkell + "\n\nWhat you can do:\n" + whatYouCanDo + `

Get started:
  skell init                  # set up this project (pick your agents)
  skell catalog               # browse well-known skill sources
  skell add anthropic         # add one (catalog id, owner/repo, URL or folder)
  skell search <keyword>      # find skills, e.g. skell search pdf
  skell install <skill>       # install a skill by name (shows a review first)
  skell status                # anything outdated?  then: skell diff / upgrade

Run 'skell <command> --help' for details on any command, or 'skell gui' for
the desktop app. Documentation: https://github.com/aminmesbahi/skell`,
		// Print just the error for runtime failures; a full usage dump buries
		// the actual message. Flag/argument mistakes still show a hint.
		SilenceUsage: true,
	}
	root.AddGroup(
		&cobra.Group{ID: "start", Title: "Getting started:"},
		&cobra.Group{ID: "manage", Title: "Managing skills:"},
		&cobra.Group{ID: "team", Title: "Teams, CI and agents:"},
		&cobra.Group{ID: "trust", Title: "Security, trust and reporting:"},
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
		newPinCmd(), newUnpinCmd(), newInfoCmd(), newDoctorCmd(), newBundleCmd())
	add("team", newSyncCmd(), newApplyCmd(), newMirrorCmd(), newTargetsCmd(), newMCPCmd(), newGUICmd())
	add("trust", newScanCmd(), newVerifyCmd(), newTrustCmd(), newReportCmd())
	add("author", newNewCmd(), newLinkCmd(), newValidateCmd(), newTestCmd(), newKeygenCmd(), newSignCmd())
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
			// Someone running skell in a folder that is not set up yet is most
			// likely new here: say what it is before saying what to type.
			if _, _, rerr := manifest.ResolveWithTarget(cwd); rerr != nil {
				printIntro(w)
			}
			printDashboard(w, cwd)
		}
		_, _ = fmt.Fprintln(w, "\n  See every command: skell --help   ·   Help for one: skell <command> --help")
		_, _ = fmt.Fprintln(w, "  Documentation: https://github.com/aminmesbahi/skell")
		return nil
	}

	return root
}

// printDashboard summarises the project in dir (no network access), or shows
// how to get started when dir has no manifest yet.
func printDashboard(w io.Writer, dir string) {
	m, t, err := manifest.ResolveWithTarget(dir)
	if err != nil {
		_, _ = fmt.Fprintf(w, "  Start here — %s is not set up for Skell yet\n\n", dir)
		_, _ = fmt.Fprintln(w, "    1. skell init               choose the agents you use (Claude, Copilot, Cursor…)")
		_, _ = fmt.Fprintln(w, "    2. skell add anthropic      add a source of skills (try: skell catalog to see more)")
		_, _ = fmt.Fprintln(w, "    3. skell search <keyword>   find a skill (for example: skell search pdf)")
		_, _ = fmt.Fprintln(w, "    4. skell install <skill>    install it by name — Skell shows what it contains first")
		_, _ = fmt.Fprintln(w, "\n  Want skills in every project, not just this folder?   skell init --user")
		_, _ = fmt.Fprintln(w, "  Prefer a window?   skell gui   (the desktop app is a separate download, see docs/install.md)")
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
	// Double-clicked from a file manager: help instead of printing an error.
	if runFirstRunIfDoubleClicked() {
		return
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
