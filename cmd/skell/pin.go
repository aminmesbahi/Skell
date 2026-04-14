package skell

import "github.com/spf13/cobra"

func newPinCmd() *cobra.Command {
	var repo string
	var version string

	cmd := &cobra.Command{
		Use:   "pin <skill-name>",
		Short: "Pin a skill to its current version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Pin(args[0], version, repo)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	cmd.Flags().StringVar(&version, "version", "", "Pin to a specific version instead of installed")
	return cmd
}

func newUnpinCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "unpin <skill-name>",
		Short: "Remove pinning for a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Unpin(args[0], repo)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	return cmd
}
