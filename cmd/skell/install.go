package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var f repoFlags
	var registry, registryURL string
	var validate, noValidate, assumeYes bool
	var targetID string

	cmd := &cobra.Command{
		Use:   "install <skill-name>...",
		Short: "Install one or more skills into one or more repositories",
		Long: `Fetches skills from your configured sources and installs them into the
repository. Updates skell.toml and skell.lock (which records the exact source
commit, so 'skell sync' reproduces the same files everywhere).

Without --registry, Skell finds the source that provides each skill; if
several do, you are asked to pick one with --registry. Use --registry-url to
add and use a new source in one step.

Before installing a skill that ships scripts or pre-approves broad tool access,
Skell shows what it contains and asks for confirmation (skip with --yes; CI and
other non-interactive runs proceed automatically). --dry-run always shows the
review without installing.`,
		Example: `  # Install from whichever configured source has it
  skell install pdf

  # Several at once
  skell install pdf docx xlsx

  # From a specific source
  skell install run-tests --registry dotnet-skills

  # Add a new source and install in one step
  skell install ilspy-decompile \
    --registry dotnet-skillz \
    --registry-url https://github.com/davidfowl/dotnet-skillz

  # Preview (shows files, scripts and tool permissions)
  skell install pdf --dry-run

  # Install for a specific agent platform (auto-inits if needed)
  skell install pdf --target copilot`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: completeRegistrySkills,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			applyValidateFlags(eng, validate, noValidate)
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			installed, skipped := 0, 0

			for _, repo := range repos {
				for _, skillName := range args {
					if !f.jsonOut {
						// A failed review (e.g. no manifest yet) falls through to
						// InstallTo, which reports the authoritative error or
						// auto-inits for --target.
						rev, err := eng.ReviewSkill(repo, skillName, registry, registryURL)
						if err == nil && !reviewAndConfirm(cmd, rev, assumeYes || f.dryRun, f.dryRun) {
							p.PrintAction(output.ActionEvent{Action: "skipped", Skill: skillName, Repo: repo})
							skipped++
							continue
						}
					}
					if err := eng.InstallTo(repo, skillName, registry, registryURL, targetID, f.dryRun); err != nil {
						return fmt.Errorf("%s: %w", repo, err)
					}
					p.PrintAction(output.ActionEvent{
						Action: "install", Skill: skillName, Repo: repo, DryRun: f.dryRun,
					})
					if !f.dryRun {
						installed++
					}
				}
			}

			if !f.dryRun && installed+skipped > 0 {
				noun := "skill"
				if installed != 1 {
					noun = "skills"
				}
				p.Success(fmt.Sprintf("%d %s installed", installed, noun))
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&registry, "registry", "", "Source (registry alias) to install from; found automatically when omitted")
	cmd.Flags().StringVar(&registryURL, "registry-url", "", "URL for the registry alias (auto-adds it to skell.toml if not present)")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to install for: claude | codex | copilot | cursor | windsurf | opencode | cline | grok")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Don't ask for confirmation before installing")
	bindValidateFlags(cmd, &validate, &noValidate)
	return cmd
}

// bindValidateFlags registers the --validate / --no-validate pair shared by
// install and upgrade.
func bindValidateFlags(cmd *cobra.Command, validate, noValidate *bool) {
	cmd.Flags().BoolVar(validate, "validate", false, "Validate the skill against the spec before writing (fail on errors)")
	cmd.Flags().BoolVar(noValidate, "no-validate", false, "Skip spec validation even if policy requires it")
}

// applyValidateFlags lets a single invocation override the policy-derived
// validation gate. --no-validate wins over --validate.
func applyValidateFlags(eng *engine.Engine, validate, noValidate bool) {
	switch {
	case noValidate:
		eng.SetRequireValidation(false)
	case validate:
		eng.SetRequireValidation(true)
	}
}
