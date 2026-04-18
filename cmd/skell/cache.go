package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the local registry cache",
		Long: `Subcommands for managing the local clone of remote registry repositories.

Registries are cloned to ~/.skell/cache/<alias>/ on first use and updated
with 'skell cache refresh'.`,
		Example: `  skell cache status
  skell cache refresh
  skell cache clear`,
	}

	cmd.AddCommand(
		newCacheStatusCmd(),
		newCacheRefreshCmd(),
		newCacheClearCmd(),
	)

	return cmd
}

func newCacheStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show cache contents and sizes",
		Long:    "Prints a summary of each cached registry: alias, skill count, and last-fetched time.",
		Example: `  skell cache status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng := engine.New(defaultCacheRoot())
			summary, err := eng.CacheStatus()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), summary)
			return nil
		},
	}
}

func newCacheRefreshCmd() *cobra.Command {
	var repo string

	c := &cobra.Command{
		Use:   "refresh",
		Short: "Fetch latest from all configured registries",
		Long:  "Runs 'git pull' on every registry clone to bring the local cache up to date.",
		Example: `  # Refresh using the current directory's manifest
  skell cache refresh

  # Refresh using a specific repo's manifest
  skell cache refresh --repo /path/to/repo`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			m, err := manifest.Resolve(repoRoot)
			if err != nil {
				return fmt.Errorf("no manifest found in %s — run 'skell init' first: %w", repoRoot, err)
			}
			eng := engine.New(defaultCacheRoot())
			if err := eng.CacheRefresh(m); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  done  cache refreshed")
			return nil
		},
	}
	c.Flags().StringVar(&repo, "repo", "", "Target repository path (for manifest resolution)")
	return c
}

func newCacheClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "clear",
		Short:   "Delete all cached registry data",
		Long:    "Removes the entire local registry cache (~/.skell/cache). Registries will be re-cloned on next use.",
		Example: `  skell cache clear`,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng := engine.New(defaultCacheRoot())
			if err := eng.CacheClear(); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  done  cache cleared")
			return nil
		},
	}
}
