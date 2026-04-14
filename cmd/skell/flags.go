package skell

import "github.com/spf13/cobra"

// repoFlags holds the common repository-targeting flags shared across commands.
type repoFlags struct {
	repo     []string
	allRepos string
	global   bool
	dryRun   bool
	jsonOut  bool
}

// bindRepoFlags attaches the standard targeting flags to a command.
func bindRepoFlags(cmd *cobra.Command, f *repoFlags) {
	cmd.Flags().StringArrayVar(&f.repo, "repo", nil, "Target repository path (repeatable)")
	cmd.Flags().StringVar(&f.allRepos, "all-repos", "", "Scan all git repos under this root path")
	cmd.Flags().BoolVar(&f.global, "global", false, "Operate on the global manifest (~/.skell/skell.toml)")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "Preview changes without applying them")
	cmd.Flags().BoolVar(&f.jsonOut, "json", false, "Output results as JSON")
}
