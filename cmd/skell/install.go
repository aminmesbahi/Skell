package skell

import "github.com/spf13/cobra"

func newInstallCmd() *cobra.Command {
	var f repoFlags
	var registry string

	cmd := &cobra.Command{
		Use:   "install <skill-name>",
		Short: "Install a skill into one or more repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Install(args[0], registry, f)
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&registry, "registry", "", "Registry alias to install from")
	return cmd
}
