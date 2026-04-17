package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	var f repoFlags
	var registry string

	cmd := &cobra.Command{
		Use:   "install <skill-name>",
		Short: "Install a skill into one or more repositories",
		Long:  "Fetches the skill from the configured registry and installs it into the target repository.\nUpdates skell.toml and skell.lock.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillName := args[0]

			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			installed := 0

			for _, repo := range repos {
				if err := eng.Install(repo, skillName, registry, f.dryRun); err != nil {
					return fmt.Errorf("%s: %w", repo, err)
				}
				if f.dryRun {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  dry-run  would install %s into %s\n",
						skillName, repo)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  install  %s → %s/.claude/skills/%s/\n",
						skillName, repo, skillName)
					installed++
				}
			}

			if !f.dryRun {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  done     %d skill installed\n", installed)
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&registry, "registry", "", "Registry alias to install from")
	return cmd
}
