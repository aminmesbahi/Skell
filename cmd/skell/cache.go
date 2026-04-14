package skell

import "github.com/spf13/cobra"

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the local registry cache",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Show cache contents and sizes",
			RunE: func(cmd *cobra.Command, args []string) error {
				// TODO: wire internal/registry.CacheStatus
				return nil
			},
		},
		&cobra.Command{
			Use:   "refresh",
			Short: "Fetch latest from all configured registries",
			RunE: func(cmd *cobra.Command, args []string) error {
				// TODO: wire internal/registry.CacheRefresh
				return nil
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Delete all cached registry data",
			RunE: func(cmd *cobra.Command, args []string) error {
				// TODO: wire internal/registry.CacheClear
				return nil
			},
		},
	)

	return cmd
}
