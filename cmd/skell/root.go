// Package skell wires all cobra commands and exposes Execute.
package skell

import (
	"os"

	"github.com/spf13/cobra"
)

// newRootCmd builds a fresh root command tree. Calling this for every test
// run ensures that flag state (e.g. StringArray accumulation) does not leak
// between test cases.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "skell",
		Short: "Govern, install, and sync engineering skills at scale.",
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
	)
	return root
}

var rootCmd = newRootCmd()

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
