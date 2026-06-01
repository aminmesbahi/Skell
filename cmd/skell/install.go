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
	var validate, noValidate bool

	cmd := &cobra.Command{
		Use:   "install <skill-name>",
		Short: "Install a skill into one or more repositories",
		Long: `Fetches the skill from the configured registry and installs it into the target repository.
Updates skell.toml and skell.lock.

If the registry alias is not yet in skell.toml, supply --registry-url to auto-add it.`,
		Example: `  # Install a skill from a registry already in skell.toml
  skell install pdf-processing --registry my-registry

  # Bootstrap a new registry and install in one step
  skell install ilspy-decompile \
    --registry dotnet-skillz \
    --registry-url https://github.com/davidfowl/dotnet-skillz

  # Preview the install without writing files
  skell install pdf-processing --registry my-registry --dry-run

  # Install into a specific repo
  skell install pdf-processing --registry my-registry --repo /path/to/repo

  # Install into all repos under a directory
  skell install pdf-processing --registry my-registry --all-repos /home/user/projects`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			applyValidateFlags(eng, validate, noValidate)
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			installed := 0

			for _, repo := range repos {
				if err := eng.Install(repo, skillName, registry, registryURL, f.dryRun); err != nil {
					return fmt.Errorf("%s: %w", repo, err)
				}
				p.PrintAction(output.ActionEvent{
					Action: "install", Skill: skillName, Repo: repo, DryRun: f.dryRun,
				})
				if !f.dryRun {
					installed++
				}
			}

			if !f.dryRun {
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
	cmd.Flags().StringVar(&registry, "registry", "", "Registry alias to install from (must exist in skell.toml, or supply --registry-url)")
	cmd.Flags().StringVar(&registryURL, "registry-url", "", "URL for the registry alias (auto-adds it to skell.toml if not present)")
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
