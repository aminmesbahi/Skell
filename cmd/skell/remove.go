package skell

import "github.com/spf13/cobra"

func newRemoveCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "remove <skill-name>",
		Short: "Remove a skill from one or more repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Remove(args[0], f)
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	return cmd
}
