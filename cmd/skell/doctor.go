package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

// fixableCodes are doctor issues that 'skell sync' repairs.
var fixableCodes = map[string]bool{"missing-dir": true, "stale-mirror": true}

func newDoctorCmd() *cobra.Command {
	var f repoFlags
	var fix bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check for manifest, lock file, and install problems",
		Long: `Audits the repository for common Skell problems:
  • missing or malformed skell.toml
  • missing skell.lock
  • skills listed in the lock file but missing on disk
  • installed skills whose content hash no longer matches the lock file
  • skills that fail Agent Skills spec validation
  • mirror copies (see 'skell mirror') that are missing or out of date

With --fix, Skell repairs what it safely can (reinstalls missing skills at
their locked commit and refreshes mirrors) and then
re-checks. Local edits are never overwritten.`,
		Example: `  # Run diagnostics on the current repo
  skell doctor

  # Repair what can be repaired automatically
  skell doctor --fix

  # Output issues as JSON (useful for CI)
  skell doctor --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			w := cmd.OutOrStdout()
			hasIssues := false
			for _, repo := range repos {
				issues, err := eng.Doctor(repo)
				if err != nil {
					return err
				}
				if fix && hasFixable(issues) {
					report, err := eng.Sync(repo, false, false, false)
					if err != nil {
						return fmt.Errorf("%s: fix failed: %w", repo, err)
					}
					if !f.jsonOut {
						for _, name := range report.Installed {
							_, _ = fmt.Fprintf(w, "  fixed    reinstalled %s\n", name)
						}
						if len(report.Mirrored) > 0 {
							_, _ = fmt.Fprintf(w, "  fixed    refreshed mirrors (%d skills)\n", len(report.Mirrored))
						}
					}
					if issues, err = eng.Doctor(repo); err != nil {
						return err
					}
				}
				var entries []output.DiagnosticEntry
				for _, issue := range issues {
					hint := issue.Hint
					if !fix && fixableCodes[issue.Code] {
						hint += " (or 'skell doctor --fix')"
					}
					entries = append(entries, output.DiagnosticEntry{
						Severity: string(issue.Severity),
						Code:     issue.Code,
						Message:  issue.Message,
						Hint:     hint,
					})
					hasIssues = true
				}
				if len(entries) == 0 {
					_, _ = fmt.Fprintf(w, "  ok  %s — no issues found\n", repo)
					continue
				}
				p.PrintDiagnostics(entries)
			}
			if hasIssues {
				return fmt.Errorf("doctor found issues — see above")
			}
			return nil
		},
	}

	bindRepoFlags(cmd, &f)
	cmd.Flags().BoolVar(&fix, "fix", false, "Reinstall missing skills and refresh mirrors, then re-check")
	return cmd
}

func hasFixable(issues []engine.DiagnosticIssue) bool {
	for _, i := range issues {
		if fixableCodes[i.Code] {
			return true
		}
	}
	return false
}
