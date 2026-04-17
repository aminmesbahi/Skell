package skell

import (
	"fmt"
	"os"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newPinCmd() *cobra.Command {
	var repo string
	var version string

	cmd := &cobra.Command{
		Use:   "pin <skill-name>",
		Short: "Pin a skill to its current version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			if err := eng.Pin(repoRoot, args[0], version); err != nil {
				return err
			}
			pinned := args[0]
			if version != "" {
				pinned += "@" + version
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  pinned   %s\n", pinned)
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	cmd.Flags().StringVar(&version, "version", "", "Pin to a specific version instead of installed")
	return cmd
}

func newUnpinCmd() *cobra.Command {
	var repo string

	cmd := &cobra.Command{
		Use:   "unpin <skill-name>",
		Short: "Remove pinning for a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			if err := eng.Unpin(repoRoot, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  unpinned %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	return cmd
}

// resolveRepo returns the given path or the current working directory when empty.
func resolveRepo(repo string) (string, error) {
	if repo != "" {
		return repo, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}
