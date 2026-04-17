package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
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
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			for _, repo := range repos {
				report, err := eng.Upgrade(repo, skillName, force, f.dryRun)
				if err != nil {
					return err
				}
				for _, u := range report.Upgraded {
					p.PrintAction(output.ActionEvent{
						Action: "upgrade", Skill: u, Repo: repo, DryRun: f.dryRun,
					})
				}
				for _, s := range report.Skipped {
					p.PrintAction(output.ActionEvent{
						Action: "skip", Skill: s, Repo: repo,
					})
				}
				if len(report.Upgraded) == 0 && len(report.Skipped) == 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  done     nothing to upgrade")
				} else if !f.dryRun {
					p.Success(fmt.Sprintf("%d skill(s) upgraded", len(report.Upgraded)))
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite locally-modified skills")
	return cmd
}
