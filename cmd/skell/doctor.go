package skell

import "github.com/spf13/cobra"

func newDoctorCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check for manifest, lock file, and install problems",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: wire internal/engine.Doctor(f)
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	return cmd
}
