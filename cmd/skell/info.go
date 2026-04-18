package skell

import (
	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newInfoCmd() *cobra.Command {
	var source, repo string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "info <skill-name>",
		Short: "Show full metadata for a skill",
		Long:  "Displays the full metadata for a skill, including frontmatter fields and lock file state.",
		Example: `  # Show info for an installed skill
  skell info pdf-processing

  # Show info for a skill in a specific repo
  skell info pdf-processing --repo /path/to/repo

  # Look up a skill in the registry (not yet installed)
  skell info ilspy-decompile --source registry

  # Output as JSON
  skell info pdf-processing --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := resolveRepo(repo)
			if err != nil {
				return err
			}

			eng := engine.New(defaultCacheRoot())
			result, err := eng.Info(repoRoot, args[0], source)
			if err != nil {
				return err
			}

			p := output.NewPrinterTo(cmd.OutOrStdout(), jsonOut)
			p.PrintInfoResult(args[0], result)
			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Show only: registry | local")
	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output results as JSON")
	return cmd
}
