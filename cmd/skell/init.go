package skell

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create skell.toml from currently installed skills",
		Long:  "Scans the repository for installed skills and generates a skell.toml manifest.\nUseful for migrating existing repositories to Skell.",
		Example: `  # Initialise the current directory
  skell init

  # Initialise a specific repository path
  skell init --repo /path/to/repo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				var err error
				repo, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			eng := engine.New(defaultCacheRoot())
			if err := eng.Init(repo); err != nil {
				return err
			}

			manifestPath := filepath.Join(repo, ".claude", "skell.toml")
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  done  skell.toml created at %s\n", manifestPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	return cmd
}
