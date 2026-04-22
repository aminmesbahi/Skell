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
	var check bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Apply skell.toml to the repository (install missing, remove unlisted)",
		Long: `Reconciles the repository's installed skills with skell.toml.

Skills listed in skell.toml but not installed are fetched and installed.
Skills installed but not listed in skell.toml are removed.
Use --check to detect drift without making any changes.`,
		Example: `  # Sync the current repo
  skell sync

  # Preview what would change without applying
  skell sync --dry-run

  # Only check for drift (exit non-zero if out of sync)
  skell sync --check

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
				report, err := eng.Sync(repo, check, f.dryRun)
				if err != nil {
					var diff *engine.SyncDiffError
					if errors.As(err, &diff) {
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
					}
					installed := report.Installed
					if installed == nil {
						installed = []string{}
					}
					removed := report.Removed
					if removed == nil {
						removed = []string{}
					}
					out, _ := json.Marshal(syncReportJSON{Installed: installed, Removed: removed})
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
	return cmd
}
