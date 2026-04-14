// Package skell wires all cobra commands and exposes Execute.
package skell

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skell",
	Short: "Govern, install, and sync engineering skills at scale.",
	Long:  `Skell is a cross-platform skill package manager for Agent Skills.`,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
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
	)
}
