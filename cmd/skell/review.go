package skell

import (
	"encoding/json"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newReviewCmd() *cobra.Command {
	var repo, registry string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "review <skill-name>",
		Short: "Inspect a skill before installing it (files, scripts, tools, links)",
		Long: `Shows what installing a skill would add to the project — its files, any
scripts or executables, the tools it pre-approves (allowed-tools) and external
links — with warnings for anything that deserves a closer look. Nothing is
installed.`,
		Example: `  skell review pdf
  skell review pdf --registry anthropic --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeRegistrySkills,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			rev, err := engine.New(defaultCacheRoot()).ReviewSkill(root, args[0], registry, "")
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(rev)
			}
			printReview(cmd.OutOrStdout(), rev)
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Project whose sources to use (defaults to current directory)")
	cmd.Flags().StringVar(&registry, "registry", "", "Source alias (found automatically when omitted)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
