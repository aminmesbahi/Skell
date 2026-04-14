package skell

import "github.com/spf13/cobra"

func newStatusCmd() *cobra.Command {
	var f repoFlags
	var only string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show comparison between registry and local installs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Status
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&only, "only", "", "Filter by status (e.g. outdated, locally-modified)")
	return cmd
}
