package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	var targetID string
	var f repoFlags
	var force, assumeYes bool
	var validate, noValidate bool

	cmd := &cobra.Command{
		Use:   "upgrade [skill-name]",
		Short: "Upgrade one or all skills",
		Long: `Fetches the latest version of skills from the registry and replaces the local
copies. Pinned skills are skipped unless --force is specified.

A skill whose update adds scripts or pre-approves broad tool access that the
installed copy did not have is shown to you and needs confirmation (skip with
--yes; CI and other non-interactive runs proceed automatically, so review the
resulting pull request). --force also restores a locally edited skill whose
source has not changed.`,
		Example: `  # Upgrade all skills in the current repo
  skell upgrade

  # Upgrade a single skill
  skell upgrade pdf-processing

  # Preview upgrades without applying
  skell upgrade --dry-run

  # Upgrade even pinned skills, and discard local edits
  skell upgrade --force

  # Don't ask before taking updates that add scripts or tool permissions
  skell upgrade --yes

  # Upgrade across multiple repos
  skell upgrade --repo ./api --repo ./worker`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeInstalledSkills,
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
			applyValidateFlags(eng, validate, noValidate)
			eng.SetUpgradeGate(func(r engine.UpgradeRisk) bool {
				if !f.jsonOut {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n  %s: this update adds more than the installed copy has\n", r.Skill)
					printReview(cmd.OutOrStdout(), r.Review)
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  new: %s\n", r.Summary())
				}
				if assumeYes || !isInteractive() {
					return true
				}
				return confirm(cmd, fmt.Sprintf("Upgrade %s?", r.Skill), false)
			})
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			for _, repo := range repos {
				report, err := eng.UpgradeFor(repo, skillName, targetID, force, f.dryRun)
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
				if f.dryRun && !f.jsonOut {
					for _, r := range report.NeedsReview {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  !  %s: this update %s\n", r.Skill, r.Summary())
					}
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
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite locally-modified skills (and upgrade pinned ones)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Don't ask before upgrades that add scripts or broad tool access")
	bindValidateFlags(cmd, &validate, &noValidate)
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to manage")
	return cmd
}
