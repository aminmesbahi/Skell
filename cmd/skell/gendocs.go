package skell

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// newGenDocsCmd generates the Markdown command reference (docs/reference)
// from the same cobra definitions the CLI uses, so docs can't drift.
func newGenDocsCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "gen-docs <dir>",
		Short:  "Generate the Markdown command reference",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(args[0], 0o755); err != nil {
				return err
			}
			root := cmd.Root()
			root.DisableAutoGenTag = true
			if err := doc.GenMarkdownTree(root, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  ✓  reference written to %s\n", args[0])
			return nil
		},
	}
}
