package skell

import "github.com/spf13/cobra"

func newUpgradeCmd() *cobra.Command {
	var f repoFlags
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade [skill-name]",
		Short: "Upgrade one or all skills",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Upgrade(args, force, f)
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite locally-modified skills")
	return cmd
}
