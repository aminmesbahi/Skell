package skell

import "github.com/spf13/cobra"

func newInfoCmd() *cobra.Command {
	var source string

	cmd := &cobra.Command{
		Use:   "info <skill-name>",
		Short: "Show full metadata for a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Info(args[0], source)
			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Show only: registry | local")
	return cmd
}
