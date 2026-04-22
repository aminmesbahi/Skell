package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var tag, lifecycle, owner, repo string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search available skills in configured registries",
		Long: `Searches all configured registries for skills matching the optional query string.
Results can be filtered by tag, lifecycle stage, or owner.`,
		Example: `  # Search all skills
  skell search

  # Full-text search by name/description
  skell search pdf

  # Filter by lifecycle
  skell search --lifecycle stable

  # Filter by owner
  skell search --owner dotnet

  # Combine filters
  skell search dotnet --lifecycle stable --owner microsoft

  # Output as JSON
  skell search --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			// Ensure global manifest exists so SearchMerged can always fall back to it.
			_ = manifest.EnsureGlobal()

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			eng := engine.New(defaultCacheRoot())
			results, err := eng.SearchMerged(repoRoot, query, tag, lifecycle, owner)
			if err != nil {
				return err
			}

			p := output.NewPrinterTo(cmd.OutOrStdout(), jsonOut)
			if len(results) == 0 {
				if jsonOut {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[]")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  no skills found")
				}
				return nil
			}
			p.PrintRegistrySkillList(results)
			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "Filter by lifecycle state")
	cmd.Flags().StringVar(&owner, "owner", "", "Filter by owner")
	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (for manifest resolution)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output results as JSON")
	return cmd
}
