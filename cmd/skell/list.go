package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var f repoFlags
	var source string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed or registry skills",
		Long: `Lists skills for one or more repositories.

By default shows skills installed locally (from skell.lock).
Use --source registry to browse all skills available in the configured registries.`,
		Example: `  # List skills installed in the current repo
  skell list

  # List all skills available in configured registries
  skell list --source registry

  # List installed skills as JSON
  skell list --json

  # List skills across every git repo under a root directory
  skell list --all-repos /home/user/projects`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)

			if source == "registry" {
				return listRegistry(cmd, eng, repos, p)
			}
			return listLocal(cmd, eng, repos, p)
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&source, "source", "local", "Source to list from: local | registry")
	return cmd
}

func listLocal(cmd *cobra.Command, eng *engine.Engine, repos []string, p *output.Printer) error {
	for _, repo := range repos {
		skills, err := eng.List(repo)
		if err != nil {
			return err
		}
		if len(skills) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  no skills installed")
			continue
		}
		p.PrintSkillList(skills)
	}
	return nil
}

func listRegistry(cmd *cobra.Command, eng *engine.Engine, repos []string, p *output.Printer) error {
	for _, repo := range repos {
		m, err := manifest.Resolve(repo)
		if err != nil {
			return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repo, err)
		}
		skills, err := eng.ListRegistry(m)
		if err != nil {
			return err
		}
		if len(skills) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  no skills found in registry")
			continue
		}
		p.PrintRegistrySkillList(skills)
	}
	return nil
}
