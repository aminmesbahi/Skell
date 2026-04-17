package skell

import (
	"errors"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var f repoFlags
	var check bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Apply skell.toml to the repository (install missing, remove unlisted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			w := cmd.OutOrStdout()
			for _, repo := range repos {
				report, err := eng.Sync(repo, check, f.dryRun)
				if err != nil {
					var diff *engine.SyncDiffError
					if errors.As(err, &diff) {
						_, _ = fmt.Fprintln(w, "  check    repo differs from manifest")
						for _, name := range diff.Missing {
							_, _ = fmt.Fprintf(w, "  missing  %s\n", name)
						}
						for _, name := range diff.Extra {
							_, _ = fmt.Fprintf(w, "  extra    %s\n", name)
						}
					}
					return err
				}
				for _, name := range report.Installed {
					prefix := "install"
					if f.dryRun {
						prefix = "would install"
					}
					_, _ = fmt.Fprintf(w, "  %-14s %s\n", prefix, name)
				}
				for _, name := range report.Removed {
					prefix := "remove"
					if f.dryRun {
						prefix = "would remove"
					}
					_, _ = fmt.Fprintf(w, "  %-14s %s\n", prefix, name)
				}
				if len(report.Installed) == 0 && len(report.Removed) == 0 {
					_, _ = fmt.Fprintln(w, "  done     already in sync")
				} else if !f.dryRun {
					_, _ = fmt.Fprintf(w, "  done     %d installed, %d removed\n",
						len(report.Installed), len(report.Removed))
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&check, "check", false, "Exit non-zero if state differs from manifest (CI use)")
	return cmd
}
