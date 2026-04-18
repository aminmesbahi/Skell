package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/model"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var f repoFlags
	var only string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show comparison between registry and local installs",
		Long: `Compares every installed skill against the registry to show whether each is
up-to-date, outdated, pinned, locally-modified, or has missing metadata.`,
		Example: `  # Check status of all skills in the current repo
  skell status

  # Show only outdated skills
  skell status --only outdated

  # Status across multiple repos
  skell status --repo ./service-a --repo ./service-b

  # Machine-readable JSON output
  skell status --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
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
				p.PrintStatusTable(filtered)
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
