// Package skell wires all cobra commands and exposes Execute.
package skell

import (
	"fmt"
	"os"

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
		Short:   "Govern, install, and sync engineering skills at scale.",
		Long: `Skell is a cross-platform skill package manager for Agent Skills (SKILL.md).

It lets you install, upgrade, pin, and sync Claude/Copilot skill files across
one or many repositories from versioned GitHub registries.

Quick start:
  skell init                         # create skell.toml in the current repo
  skell list --source registry       # browse available skills
  skell install <skill>              # install a skill
  skell status                       # check for outdated skills
  skell upgrade                      # upgrade all non-pinned skills

Run 'skell <command> --help' for detailed help and examples on each command.`,
	}
	root.AddCommand(
		newGUICmd(),
		newListCmd(),
		newStatusCmd(),
		newInfoCmd(),
		newInstallCmd(),
		newUpgradeCmd(),
		newRemoveCmd(),
		newPinCmd(),
		newUnpinCmd(),
		newSyncCmd(),
		newInitCmd(),
		newSearchCmd(),
		newDoctorCmd(),
		newCacheCmd(),
		newSelfUpdateCmd(),
		newAddCmd(),
		newTargetsCmd(),
	)

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {}
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "skell version %s (commit %s, built %s)\n\n",
			version.Version, version.Commit, version.Date); err != nil {
			return err
		}
		return cmd.Help()
	}

	return root
}

var rootCmd = newRootCmd()

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
