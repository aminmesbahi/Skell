package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var f repoFlags
	var registry string

	cmd := &cobra.Command{
		Use:   "install <skill-name>",
		Short: "Install a skill into one or more repositories",
		Long:  "Fetches the skill from the configured registry and installs it into the target repository.\nUpdates skell.toml and skell.lock.",
		Example: `  # Install a skill from the default registry
  skell install pdf-processing

  # Install from a specific registry alias
  skell install ilspy-decompile --registry dotnet-skillz

  # Preview the install without writing files
  skell install pdf-processing --dry-run

  # Install into a specific repo
  skell install pdf-processing --repo /path/to/repo

  # Install into all repos under a directory
  skell install pdf-processing --all-repos /home/user/projects`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			installed := 0

			for _, repo := range repos {
				if err := eng.Install(repo, skillName, registry, f.dryRun); err != nil {
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
				p.Success(fmt.Sprintf("%d skill installed", installed))
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&registry, "registry", "", "Registry alias to install from")
	return cmd
}
