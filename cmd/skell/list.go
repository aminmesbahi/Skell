package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var f repoFlags
	var source string
	var targetID string

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
  skell list --all-repos /home/user/projects

  # List skills for a specific agent platform
  skell list --target cursor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)

			if source == "registry" {
				// Fall back to the global manifest only when --repo was not given.
				fallback := len(f.repo) == 0 && !f.global && f.allRepos == ""
				return listRegistry(cmd, eng, repos, p, fallback)
			}
			return listLocal(cmd, eng, repos, p, targetID)
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&source, "source", "local", "Source to list from: local | registry")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to list skills for: claude | codex | copilot | cursor | windsurf | opencode | cline | grok")
	return cmd
}

func listLocal(cmd *cobra.Command, eng *engine.Engine, repos []string, p *output.Printer, targetID string) error {
	for _, repo := range repos {
		skills, err := eng.ListFor(repo, targetID)
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

func listRegistry(cmd *cobra.Command, eng *engine.Engine, repos []string, p *output.Printer, fallback bool) error {
	for _, repo := range repos {
		m, err := resolveManifest(repo, fallback)
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
