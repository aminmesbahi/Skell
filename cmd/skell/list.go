package skell

import "github.com/spf13/cobra"

func newListCmd() *cobra.Command {
	var f repoFlags
	var source string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed or registry skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.List
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&source, "source", "local", "Source to list from: local | registry")
	return cmd
}
