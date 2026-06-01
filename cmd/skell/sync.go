package skell

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var f repoFlags
	var check, prune bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Apply skell.toml to the repository (install missing, remove unlisted)",
		Long: `Reconciles the repository's installed skills with skell.toml.

Skills listed in skell.toml but not installed are fetched and installed.
Skills that Skell previously installed (recorded in skell.lock) but are no
longer listed in skell.toml are removed.

Hand-authored skills that Skell never installed (not in skell.lock) are left
in place and reported as "untracked". Pass --prune to remove them too.
Use --check to detect drift without making any changes.`,
		Example: `  # Sync the current repo
  skell sync

  # Preview what would change without applying
  skell sync --dry-run

  # Only check for drift (exit non-zero if out of sync)
  skell sync --check

  # Also remove hand-authored skills not in the manifest
  skell sync --prune

  # Sync multiple repos
  skell sync --repo ./api --repo ./worker`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			w := cmd.OutOrStdout()
			for _, repo := range repos {
				report, err := eng.Sync(repo, check, f.dryRun, prune)
				if err != nil {
					if diff, ok := errors.AsType[*engine.SyncDiffError](err); ok {
						_, _ = fmt.Fprintln(w, "  check    repo differs from manifest")
						for _, name := range diff.Missing {
							_, _ = fmt.Fprintf(w, "  missing  %s\n", name)
						}
						for _, name := range diff.Extra {
							_, _ = fmt.Fprintf(w, "  extra    %s\n", name)
						}
					}
					return err
				}
				if f.jsonOut {
					type syncReportJSON struct {
						Installed []string `json:"installed"`
						Removed   []string `json:"removed"`
						Untracked []string `json:"untracked"`
					}
					out, _ := json.Marshal(syncReportJSON{
						Installed: orEmpty(report.Installed),
						Removed:   orEmpty(report.Removed),
						Untracked: orEmpty(report.Untracked),
					})
					_, _ = fmt.Fprintf(w, "%s\n", out)
					continue
				}
				for _, name := range report.Installed {
					p.PrintAction(output.ActionEvent{
						Action: "install", Skill: name, Repo: repo, DryRun: f.dryRun,
					})
				}
				for _, name := range report.Removed {
					p.PrintAction(output.ActionEvent{
						Action: "remove", Skill: name, Repo: repo, DryRun: f.dryRun,
					})
				}
				for _, name := range report.Untracked {
					_, _ = fmt.Fprintf(w, "  kept     %s (untracked local skill — use --prune to remove)\n", name)
				}
				if len(report.Installed) == 0 && len(report.Removed) == 0 {
					_, _ = fmt.Fprintln(w, "  done     already in sync")
				} else if !f.dryRun {
					p.Success(fmt.Sprintf("%d installed, %d removed", len(report.Installed), len(report.Removed)))
				}
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&check, "check", false, "Exit non-zero if state differs from manifest (CI use)")
	cmd.Flags().BoolVar(&prune, "prune", false, "Also remove hand-authored skills not in the manifest or lock file")
	return cmd
}

// orEmpty returns a non-nil slice so JSON output renders [] rather than null.
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
