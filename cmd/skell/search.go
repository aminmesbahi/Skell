package skell

import "github.com/spf13/cobra"

func newSearchCmd() *cobra.Command {
	var tag, lifecycle, owner string

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search available skills in configured registries",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Search(query, tag, lifecycle, owner)
			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&lifecycle, "lifecycle", "", "Filter by lifecycle state")
	cmd.Flags().StringVar(&owner, "owner", "", "Filter by owner")
	return cmd
}
