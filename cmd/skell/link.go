package skell

import (
	"encoding/json"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/spf13/cobra"
)

func newLinkCmd() *cobra.Command {
	var repo, targetID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "link <skill-folder>",
		Short: "Use a skill you are developing without copying it (live edits)",
		Long: `Links a local skill folder into the repository's skills directory, so the
agent sees your edits immediately — like 'npm link'. Skell uses a symlink, or
a directory junction on Windows when symlinks need admin rights.

Linked skills are recorded in skell.lock as linked, are never pruned by sync
or touched by upgrade, and are added to .git/info/exclude so the
machine-specific link is not committed. Unlink with 'skell remove <name>'
(your folder is left untouched).`,
		Example: `  # Work on a skill in another folder
  skell link ~/src/my-skills/release-notes

  # Stop using it
  skell remove release-notes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			res, err := engine.New(defaultCacheRoot()).Link(root, args[0], targetID)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			_, _ = fmt.Fprintf(w, "  ✓  linked %s → %s\n", res.Path, res.Source)
			_, _ = fmt.Fprintf(w, "     edits are live; unlink with 'skell remove %s'\n", res.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform to link for")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	return cmd
}
