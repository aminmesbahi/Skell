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
	var check, prune, force, allowModified bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Apply skell.toml to the repository (install missing, remove unlisted)",
		Long: `Reconciles the repository's installed skills with skell.toml.

Skills listed in skell.toml but not installed are fetched and installed.
Skills that Skell previously installed (recorded in skell.lock) but are no
longer listed in skell.toml are removed.

Hand-authored skills that Skell never installed (not in skell.lock) are left
in place and reported as "untracked". Pass --prune to remove them too.
Installed skills are also verified against the content hash in skell.lock. A
skill edited by hand is reported as "modified" (and fails --check) but is left
alone; pass --force to discard the edits and restore the locked content.
Use --check to detect drift without making any changes; add --allow-modified if
you customise installed skills on purpose and don't want that to fail the check.`,
		Example: `  # Sync the current repo
  skell sync

  # Preview what would change without applying
  skell sync --dry-run

  # Only check for drift (exit non-zero if out of sync)
  skell sync --check

  # Also remove hand-authored skills not in the manifest
  skell sync --prune

  # Discard local edits to installed skills and restore the locked content
  skell sync --force

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
			var perRepo []syncReportJSON
			for _, repo := range repos {
				report, err := eng.SyncWith(repo, engine.SyncOptions{CheckOnly: check, DryRun: f.dryRun, Prune: prune, Force: force, AllowModified: allowModified})
				if err != nil {
					// Human-readable detail is skipped with --json so stdout stays a
					// single parseable document; the returned error still says why.
					if diff, ok := errors.AsType[*engine.SyncDiffError](err); ok && !f.jsonOut {
						_, _ = fmt.Fprintln(w, "  check    repo differs from manifest")
						for _, name := range diff.Missing {
							_, _ = fmt.Fprintf(w, "  missing  %s\n", name)
						}
						for _, name := range diff.Extra {
							_, _ = fmt.Fprintf(w, "  extra    %s\n", name)
						}
						for _, name := range diff.StaleMirrors {
							_, _ = fmt.Fprintf(w, "  stale    %s (mirror copy differs — run 'skell mirror refresh')\n", name)
						}
						for _, name := range diff.Modified {
							_, _ = fmt.Fprintf(w, "  modified %s (differs from skell.lock — 'skell diff %s' to see, 'skell sync --force' to restore)\n", name, name)
						}
					}
					return err
				}
				if f.jsonOut {
					perRepo = append(perRepo, syncReportJSON{
						Repo:      repo,
						Installed: orEmpty(report.Installed),
						Removed:   orEmpty(report.Removed),
						Untracked: orEmpty(report.Untracked),
						Modified:  orEmpty(report.Modified),
						Restored:  orEmpty(report.Restored),
					})
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
				for _, name := range report.Restored {
					p.PrintAction(output.ActionEvent{
						Action: "restore", Skill: name, Repo: repo, DryRun: f.dryRun,
					})
				}
				if !force {
					for _, name := range report.Modified {
						_, _ = fmt.Fprintf(w, "  modified %s (differs from skell.lock — kept; use --force to restore)\n", name)
					}
				}
				for _, name := range report.Untracked {
					_, _ = fmt.Fprintf(w, "  kept     %s (untracked local skill — use --prune to remove)\n", name)
				}
				if len(report.Installed)+len(report.Removed)+len(report.Restored) == 0 {
					if len(report.Modified) == 0 {
						_, _ = fmt.Fprintln(w, "  done     already in sync")
					}
				} else if !f.dryRun {
					p.Success(fmt.Sprintf("%d installed, %d removed, %d restored", len(report.Installed), len(report.Removed), len(report.Restored)))
				}
			}

			if f.jsonOut {
				// A single repo keeps the original flat-object shape (no
				// "repo" key) for backward compatibility. With more than one
				// repo, printing one json.Marshal per repo produced several
				// concatenated top-level JSON values on separate lines — not
				// parseable as a single JSON document — so those are wrapped
				// in one array instead, each tagged with its repo.
				var out []byte
				var marshalErr error
				if len(perRepo) == 1 {
					single := perRepo[0]
					single.Repo = ""
					out, marshalErr = json.Marshal(single)
				} else {
					out, marshalErr = json.Marshal(perRepo)
				}
				if marshalErr != nil {
					return marshalErr
				}
				_, _ = fmt.Fprintf(w, "%s\n", out)
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&check, "check", false, "Exit non-zero if state differs from manifest (CI use)")
	cmd.Flags().BoolVar(&force, "force", false, "Discard local edits to installed skills and restore them from skell.lock")
	cmd.Flags().BoolVar(&allowModified, "allow-modified", false, "With --check, do not fail for skills edited since install")
	cmd.Flags().BoolVar(&prune, "prune", false, "Also remove hand-authored skills not in the manifest or lock file")
	return cmd
}

// syncReportJSON is the --json shape for one repo's sync report. Repo is
// omitted when only one repo was targeted, to keep that (by far the most
// common) case's output identical to before multi-repo support was added.
type syncReportJSON struct {
	Repo      string   `json:"repo,omitempty"`
	Installed []string `json:"installed"`
	Removed   []string `json:"removed"`
	Untracked []string `json:"untracked"`
	Modified  []string `json:"modified"`
	Restored  []string `json:"restored"`
}

// orEmpty returns a non-nil slice so JSON output renders [] rather than null.
func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
