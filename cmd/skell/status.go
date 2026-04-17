package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var f repoFlags
	var only string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show comparison between registry and local installs",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			w := cmd.OutOrStdout()
			for _, repo := range repos {
				entries, err := eng.Status(repo)
				if err != nil {
					return err
				}
				if len(repos) > 1 {
					_, _ = fmt.Fprintf(w, "\n  %s\n", repo)
				}
				filtered := filterStatus(entries, only)
				if len(filtered) == 0 {
					_, _ = fmt.Fprintln(w, "  no skills to show")
					continue
				}
				_, _ = fmt.Fprintf(w, "  %-28s  %-12s  %-12s  %s\n", "skill", "installed", "latest", "status")
				_, _ = fmt.Fprintf(w, "  %-28s  %-12s  %-12s  %s\n", "-----", "---------", "------", "------")
				for _, e := range filtered {
					_, _ = fmt.Fprintf(w, "  %-28s  %-12s  %-12s  %s\n", e.Name, e.Installed, e.Latest, e.Status)
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&only, "only", "", "Filter by status (e.g. outdated, locally-modified)")
	return cmd
}

func filterStatus(entries []model.StatusEntry, only string) []model.StatusEntry {
	if only == "" {
		return entries
	}
	var out []model.StatusEntry
	for _, e := range entries {
		if string(e.Status) == only {
			out = append(out, e)
		}
	}
	return out
}
