package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/target"
	"github.com/spf13/cobra"
)

func newTargetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "targets",
		Short: "List supported AI-agent platform layouts",
		Long: `Lists the AI-agent platforms (targets) Skell can manage skills for.

Each target maps to a different on-disk convention. The skill content format
(SKILL.md + YAML frontmatter) is identical across platforms; only the directory
where Skell places the files differs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "  %-10s %-32s %s\n", "id", "platform", "directory")
			for _, t := range target.All() {
				_, _ = fmt.Fprintf(out, "  %-10s %-32s %s/skills/\n", t.ID, t.DisplayName, t.Dir)
			}
			return nil
		},
	}
	return cmd
}
