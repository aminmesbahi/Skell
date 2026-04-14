package skell

import "github.com/spf13/cobra"

func newInitCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create skell.toml from currently installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Init(repo)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	return cmd
}
