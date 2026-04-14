package skell

import "github.com/spf13/cobra"

func newSyncCmd() *cobra.Command {
	var f repoFlags
	var check bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Apply skell.toml to the repository (install missing, remove unlisted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Sync(check, f)
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&check, "check", false, "Exit non-zero if state differs from manifest (CI use)")
	return cmd
}
