package skell

import (
	"os"
	"path/filepath"

	"github.com/aminmesbahi/skell/internal/scanner"
	"github.com/spf13/cobra"
)

// repoFlags holds the common repository-targeting flags shared across commands.
type repoFlags struct {
	repo     []string
	allRepos string
	global   bool
	dryRun   bool
	jsonOut  bool
}

// bindRepoFlags attaches the standard targeting flags to a command.
func bindRepoFlags(cmd *cobra.Command, f *repoFlags) {
	cmd.Flags().StringArrayVar(&f.repo, "repo", nil, "Target repository path (repeatable)")
	cmd.Flags().StringVar(&f.allRepos, "all-repos", "", "Scan all git repos under this root path")
	cmd.Flags().BoolVar(&f.global, "global", false, "Operate on the global manifest (~/.skell/skell.toml)")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "Preview changes without applying them")
	cmd.Flags().BoolVar(&f.jsonOut, "json", false, "Output results as JSON")
}

// resolveRepos returns the list of repository roots to operate on based on the flags.
// When no flags are set, the current working directory is used.
func resolveRepos(f repoFlags) ([]string, error) {
	if len(f.repo) > 0 {
		return f.repo, nil
	}
	if f.allRepos != "" {
		results, err := scanner.ScanAll(f.allRepos)
		if err != nil {
			return nil, err
		}
		repos := make([]string, len(results))
		for i, r := range results {
			repos[i] = r.RepoRoot
		}
		return repos, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return []string{cwd}, nil
}

// defaultCacheRoot returns the path to Skell's local registry cache (~/.skell/cache).
func defaultCacheRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".skell", "cache")
	}
	return filepath.Join(home, ".skell", "cache")
}
