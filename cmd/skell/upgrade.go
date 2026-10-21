package skell

import (
	"fmt"
	"strings"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	var targetID string
	var f repoFlags
	var force, assumeYes bool
	var validate, noValidate bool
	var verify bool
	var verifyAgent string

	cmd := &cobra.Command{
		Use:   "upgrade [skill-name]",
		Short: "Upgrade one or all skills",
		Long: `Fetches the latest version of skills from the registry and replaces the local
copies. Pinned skills are skipped unless --force is specified.

A skill whose update adds scripts or pre-approves broad tool access that the
installed copy did not have is shown to you and needs confirmation. Without a
terminal (CI, scripts, --json) it is skipped and the command fails unless you
pass --yes. --force also restores a locally edited skill whose
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
			// Like 'install', an upgrade that adds scripts or broad tool access is
			// never applied without approval: with nobody to ask (no terminal, or
			// --json) and no --yes it is skipped and the command fails.
			if verify {
				// Run the new version evals first and refuse an upgrade that breaks them.
				root, _ := resolveRepo("")
				if len(repos) > 0 {
					root = repos[0]
				}
				eng.SetUpgradeVerifier(upgradeVerifier(verifyAgent, targetID, root, 120))
			}
			var declined []string
			eng.SetUpgradeGate(func(r engine.UpgradeRisk) bool {
				if !f.jsonOut {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n  %s: this update adds more than the installed copy has\n", r.Skill)
					printReview(cmd.OutOrStdout(), r.Review)
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  new: %s\n", r.Summary())
				}
				if assumeYes {
					return true
				}
				if f.jsonOut || !isInteractive() {
					declined = append(declined, r.Skill)
					return false
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
				held := 0
				for _, s := range report.Skipped {
					if strings.Contains(s, "(not upgraded:") {
						held++
					}
				}
				if verify && held > 0 {
					return fmt.Errorf("%d upgrade(s) held back because their evals failed — see 'skell test'", held)
				}
				if len(declined) > 0 {
					return fmt.Errorf("%s: the update adds scripts or broad tool access — review it with 'skell diff <skill>' and re-run with --yes to upgrade without a prompt",
						strings.Join(declined, ", "))
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
	cmd.Flags().BoolVar(&verify, "verify", false, "Run each skill evals (evals/evals.yaml) on the new version first and skip upgrades that break them")
	cmd.Flags().StringVar(&verifyAgent, "agent", "", "Agent command for --verify, e.g. \"claude -p\" (default: $SKELL_TEST_AGENT)")
	return cmd
}
