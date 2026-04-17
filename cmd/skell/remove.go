package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "remove <skill-name>",
		Short: "Remove a skill from one or more repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			w := cmd.OutOrStdout()
			for _, repo := range repos {
				if err := eng.Remove(repo, args[0], f.dryRun); err != nil {
					return err
				}
				if f.dryRun {
					_, _ = fmt.Fprintf(w, "  dry-run  would remove %s\n", args[0])
				} else {
					_, _ = fmt.Fprintf(w, "  removed  %s\n", args[0])
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	return cmd
}
