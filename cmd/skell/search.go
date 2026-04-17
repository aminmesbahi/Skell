package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var tag, lifecycle, owner, repo string

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search available skills in configured registries",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			m, err := manifest.Resolve(repoRoot)
			if err != nil {
				return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
			}

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			eng := engine.New(defaultCacheRoot())
			results, err := eng.Search(m, query, tag, lifecycle, owner)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  no skills found")
				return nil
			}
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %-12s  %s\n", "skill", "version", "lifecycle", "owner")
			_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %-12s  %s\n", "-----", "-------", "---------", "-----")
			for _, s := range results {
				_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %-12s  %s\n",
					s.Name, s.Metadata.Version, s.Metadata.Lifecycle, s.Metadata.Owner)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "Filter by lifecycle state")
	cmd.Flags().StringVar(&owner, "owner", "", "Filter by owner")
	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (for manifest resolution)")
	return cmd
}
