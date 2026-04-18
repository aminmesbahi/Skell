package skell

import (
	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newPinCmd() *cobra.Command {
	var repo string
	var version string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "pin <skill-name>",
		Short: "Pin a skill to its current version",
		Long: `Marks a skill as pinned so that 'skell upgrade' will not update it.
The installed version is recorded in skell.toml and skell.lock.`,
		Example: `  # Pin a skill to its currently-installed version
  skell pin pdf-processing

  # Pin to a specific version
  skell pin pdf-processing --version 1.2.0

  # Pin in a specific repo
  skell pin pdf-processing --repo /path/to/repo`,
		Args: cobra.ExactArgs(1),
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
			p := output.NewPrinterTo(cmd.OutOrStdout(), jsonOut)
			p.PrintAction(output.ActionEvent{Action: "pin", Skill: pinned, Repo: repoRoot})
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	cmd.Flags().StringVar(&version, "version", "", "Pin to a specific version instead of installed")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}

func newUnpinCmd() *cobra.Command {
	var repo string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "unpin <skill-name>",
		Short: "Remove pinning for a skill",
		Long:  "Clears the pinned flag so the skill will be included in future 'skell upgrade' runs.",
		Example: `  # Unpin a skill
  skell unpin pdf-processing

  # Unpin in a specific repo
  skell unpin pdf-processing --repo /path/to/repo`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			if err := eng.Unpin(repoRoot, args[0]); err != nil {
				return err
			}
			p := output.NewPrinterTo(cmd.OutOrStdout(), jsonOut)
			p.PrintAction(output.ActionEvent{Action: "unpin", Skill: args[0], Repo: repoRoot})
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
