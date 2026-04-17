package skell

import (
	"encoding/json"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var f repoFlags
	var source string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed or registry skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())

			if source == "registry" {
				return listRegistry(cmd, eng, repos)
			}
			return listLocal(cmd, eng, repos, f.jsonOut)
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().StringVar(&source, "source", "local", "Source to list from: local | registry")
	return cmd
}

func listLocal(cmd *cobra.Command, eng *engine.Engine, repos []string, jsonOut bool) error {
	for _, repo := range repos {
		skills, err := eng.List(repo)
		if err != nil {
			return err
		}
		if jsonOut {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(skills)
		}
		if len(skills) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  no skills installed")
			continue
		}
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %s\n", "skill", "version", "registry")
		_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %s\n", "-----", "-------", "--------")
		for _, s := range skills {
			pinned := ""
			if s.Pinned {
				pinned = " [pinned]"
			}
			_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %s%s\n", s.Name, s.Version, s.Registry, pinned)
		}
	}
	return nil
}

func listRegistry(cmd *cobra.Command, eng *engine.Engine, repos []string) error {
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
		w := cmd.OutOrStdout()
		_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %s\n", "skill", "version", "lifecycle")
		_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %s\n", "-----", "-------", "---------")
		for _, s := range skills {
			_, _ = fmt.Fprintf(w, "  %-30s  %-12s  %s\n", s.Name, s.Metadata.Version, s.Metadata.Lifecycle)
		}
	}
	return nil
}
