package skell

import (
	"fmt"

	"github.com/aminmesbahi/skell/internal/engine"
	"github.com/aminmesbahi/skell/internal/output"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var f repoFlags

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check for manifest, lock file, and install problems",
		Long: `Audits the repository for common Skell problems:
  • missing or malformed skell.toml
  • missing skell.lock
  • skills listed in the manifest but not installed on disk
  • installed skills whose content hash no longer matches the lock file`,
		Example: `  # Run diagnostics on the current repo
  skell doctor

  # Diagnose a specific repo
  skell doctor --repo /path/to/repo

  # Output issues as JSON (useful for CI)
  skell doctor --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resolveRepos(f)
			if err != nil {
				return err
			}
			eng := engine.New(defaultCacheRoot())
			p := output.NewPrinterTo(cmd.OutOrStdout(), f.jsonOut)
			hasIssues := false
			for _, repo := range repos {
				issues, err := eng.Doctor(repo)
				if err != nil {
					return err
				}
				var entries []output.DiagnosticEntry
				for _, issue := range issues {
					entries = append(entries, output.DiagnosticEntry{
						Severity: string(issue.Severity),
						Code:     issue.Code,
						Message:  issue.Message,
						Hint:     issue.Hint,
					})
					hasIssues = true
				}
				if len(entries) == 0 {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  ok  %s — no issues found\n", repo)
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
	return cmd
}
