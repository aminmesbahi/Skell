package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	var f repoFlags
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade [skill-name]",
		Short: "Upgrade one or all skills",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			skillName := ""
			if len(args) > 0 {
				skillName = args[0]
			}
			eng := engine.New(defaultCacheRoot())
			w := cmd.OutOrStdout()
			for _, repo := range repos {
				report, err := eng.Upgrade(repo, skillName, force, f.dryRun)
				if err != nil {
					return err
				}
				if f.dryRun {
					_, _ = fmt.Fprintln(w, "  dry-run  no changes applied")
				}
				for _, u := range report.Upgraded {
					_, _ = fmt.Fprintf(w, "  upgrade  %s\n", u)
				}
				for _, s := range report.Skipped {
					_, _ = fmt.Fprintf(w, "  skip     %s\n", s)
				}
				if len(report.Upgraded) == 0 && len(report.Skipped) == 0 {
					_, _ = fmt.Fprintln(w, "  done     nothing to upgrade")
				} else if !f.dryRun {
					_, _ = fmt.Fprintf(w, "  done     %d skill(s) upgraded\n", len(report.Upgraded))
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite locally-modified skills")
	return cmd
}
