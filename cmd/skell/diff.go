package skell

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/manifest"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	var repo, targetID string
	var refresh, jsonOut, noColor bool

	cmd := &cobra.Command{
		Use:   "diff <skill-name>",
		Short: "Show what changed upstream for an installed skill",
		Long: `Shows a unified diff from the installed copy of a skill to the current
version in its source, so you can review an upgrade before running
'skell upgrade'. Local edits you made to the installed copy show up too.

Use --refresh to fetch the latest from the source first.`,
		Example: `  # Review upstream changes before upgrading
  skell diff pdf --refresh
  skell upgrade pdf

  # Machine-readable
  skell diff pdf --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeInstalledSkills,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepo(repo)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			if refresh {
				m, err := manifest.Resolve(root)
				if err != nil {
					return err
				}
				if err := eng.CacheRefresh(m); err != nil {
					return err
				}
			}
			color := !noColor && !jsonOut && isInteractive() && os.Getenv("NO_COLOR") == ""
			res, err := eng.Diff(root, args[0], targetID, color)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if jsonOut {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(res)
			}
			from := dash(res.InstalledVer)
			if res.InstalledAt != "" {
				from += " @ " + shortSHA(res.InstalledAt)
			}
			to := dash(res.LatestVer)
			if res.LatestCommit != "" {
				to += " @ " + shortSHA(res.LatestCommit)
			}
			_, _ = fmt.Fprintf(w, "  %s  (%s)  installed %s → latest %s\n", res.Name, res.Registry, from, to)
			if res.LocallyChanged {
				_, _ = fmt.Fprintln(w, "  note: the installed copy has local edits; they appear as differences below")
			}
			if res.Patch == "" {
				_, _ = fmt.Fprintln(w, "  ✓  no differences — installed copy matches the source")
				return nil
			}
			_, _ = fmt.Fprintln(w)
			_, _ = fmt.Fprint(w, res.Patch)
			_, _ = fmt.Fprintf(w, "\n  apply with: skell upgrade %s\n", res.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&repo, "repo", "", "Target repository path (defaults to current directory)")
	cmd.Flags().StringVar(&targetID, "target", "", "Agent platform the skill is installed for")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "Fetch the latest from the source before comparing")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	return cmd
}
